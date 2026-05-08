package sdk_entity

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mattiabonardi/endor-sdk-go/internal/api_gateway"
	"github.com/mattiabonardi/endor-sdk-go/internal/swagger"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_i18n"
	"gopkg.in/yaml.v3"
)

var (
	registryCoreInstance *RegistryCore
	registryCoreOnce     sync.Once
)

func GetRegistryCore() *RegistryCore {
	return registryCoreInstance
}

func InitRegistryCore(microServiceId string, module string, internalEndorHandlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger) *RegistryCore {
	registryCoreOnce.Do(func() {
		absProdRoot, _ := filepath.Abs("prod")
		registryCoreInstance = &RegistryCore{
			microServiceId:        microServiceId,
			module:                module,
			internalEndorHandlers: internalEndorHandlers,
			logger:                logger,
			mu:                    &sync.RWMutex{},
			prodDAO:               &sdk.DSLDAO{BasePath: absProdRoot},
			ephemeralCache:        newEphemeralCacheManager(),
		}
		if err := registryCoreInstance.startDslProdWatcher(); err != nil {
			logger.Warn(fmt.Sprintf("unable to start entity DSL watcher: %s", err.Error()))
		}
	})
	return registryCoreInstance
}

// RegistryCore manages building and caching the handler registry.
// Production requests use a cached dictionary; development requests (x-development: true)
// get a per-user ephemeral overlay built on top of the production dictionary.
type RegistryCore struct {
	module                string
	microServiceId        string
	internalEndorHandlers *[]sdk.EndorHandlerInterface
	prodDAO               *sdk.DSLDAO
	logger                *sdk.Logger
	mu                    *sync.RWMutex
	cachedDictionary      map[string]EndorEntityDictionary
	cachedDIContainer     *EndorDIContainer
	cacheInitialized      bool
	ephemeralCache        *EphemeralCacheManager
}

// EndorEntityDictionary is the per-entity descriptor: compiled handler and entity metadata.
type EndorEntityDictionary struct {
	OriginalInstance *sdk.EndorHandlerInterface
	EndorHandler     sdk.EndorHandler
	entity           sdk.EntityInterface
}

type EndorHandlerActionDictionary struct {
	EndorHandlerAction sdk.EndorHandlerActionInterface
	entityAction       sdk.EntityAction
	Container          *EndorDIContainer
}

// dslCategory is the YAML structure for category entries in entityDSLFile.
// AdditionalSchema is parsed directly as a sdk.RootSchema, unlike sdk.DynamicCategory
// which stores it as a raw YAML string.
type dslCategory struct {
	ID               string         `yaml:"id"`
	Title            string         `yaml:"title"`
	Description      string         `yaml:"description"`
	AdditionalSchema sdk.RootSchema `yaml:"additionalSchema"`
}

// entityDSLFile is the YAML structure for entity definition files.
// The entity type is inferred: no categories → dynamic, with categories → dynamic-specialized.
type entityDSLFile struct {
	Title            string         `yaml:"title"`
	Description      string         `yaml:"description"`
	AdditionalSchema sdk.RootSchema `yaml:"additionalSchema"`
	Categories       []dslCategory  `yaml:"categories"`
}

// #region Public API

// Dictionary returns the handler dictionary for the given session.
// Production sessions get the cached dictionary; development sessions get a per-user
// ephemeral overlay (falls back to production on error).
func (c *RegistryCore) Dictionary(session sdk.Session) (map[string]EndorEntityDictionary, error) {
	if !session.Development || session.Username == "" {
		return c.dictionaryMap()
	}
	if cached, _ := c.ephemeralCache.Get(session.Username); cached != nil {
		return cached, nil
	}
	devDict, devContainer, err := c.buildDevDictionary(session)
	if err != nil {
		c.logger.Warn(fmt.Sprintf("failed to build ephemeral registry, falling back to prod: %s", err.Error()))
		return c.dictionaryMap()
	}
	c.ephemeralCache.Set(session.Username, devDict, devContainer)
	return devDict, nil
}

// Container returns the unified EndorDIContainer for the given session.
// Production sessions get the cached container; development sessions get the per-user one.
func (c *RegistryCore) Container(session sdk.Session) (*EndorDIContainer, error) {
	if !session.Development || session.Username == "" {
		_, err := c.dictionaryMap()
		if err != nil {
			return nil, err
		}
		c.mu.RLock()
		container := c.cachedDIContainer
		c.mu.RUnlock()
		return container, nil
	}
	if _, container := c.ephemeralCache.Get(session.Username); container != nil {
		return container, nil
	}
	// Trigger rebuild which also caches the container
	_, err := c.Dictionary(session)
	if err != nil {
		return nil, err
	}
	_, container := c.ephemeralCache.Get(session.Username)
	if container != nil {
		return container, nil
	}
	// Fallback to prod container
	return c.Container(sdk.Session{})
}

// DictionaryInstance resolves a single entry by entity ID for the given session.
func (c *RegistryCore) DictionaryInstance(session sdk.Session, dto sdk.ReadInstanceDTO) (*EndorEntityDictionary, error) {
	dict, err := c.Dictionary(session)
	if err != nil {
		return nil, err
	}
	if entry, ok := dict[dto.Id]; ok {
		return &entry, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id)).WithTranslation("sdk.entity.messages.not_found", map[string]any{"id": dto.Id})
}

// Sync invalidates the production cache and all per-user ephemeral overlays,
// forcing a full rebuild on the next access.
func (c *RegistryCore) Sync() {
	c.mu.Lock()
	c.cacheInitialized = false
	c.cachedDictionary = nil
	c.cachedDIContainer = nil
	c.mu.Unlock()
	c.ephemeralCache.InvalidateAll()
}

// #endregion

// #region Internal machinery

// toDynamicCategories converts DSL categories (with sdk.RootSchema) to sdk.DynamicCategory (with YAML string).
func toDynamicCategories(dslCats []dslCategory) ([]sdk.DynamicCategory, error) {
	result := make([]sdk.DynamicCategory, 0, len(dslCats))
	for _, c := range dslCats {
		s, err := c.AdditionalSchema.ToYAML()
		if err != nil {
			return nil, fmt.Errorf("marshal additional category %q additionalSchema: %w", c.ID, err)
		}
		result = append(result, sdk.DynamicCategory{
			ID:               c.ID,
			Title:            c.Title,
			Description:      c.Description,
			AdditionalSchema: s,
		})
	}
	return result, nil
}

// buildCategorySchemas converts the HybridCategory slice of a specialized entity into the
// handler category list and per-category schema map.
func (c *RegistryCore) buildCategorySchemas(entityID string, cats []sdk.HybridCategory) ([]sdk.EndorHybridSpecializedHandlerCategoryInterface, map[string]sdk.RootSchema) {
	handlerCats := make([]sdk.EndorHybridSpecializedHandlerCategoryInterface, 0, len(cats))
	schemas := make(map[string]sdk.RootSchema, len(cats))
	for _, cat := range cats {
		handlerCats = append(handlerCats, NewEndorHybridSpecializedHandlerCategory[*sdk.DynamicEntitySpecialized](cat.ID, cat.Description))
		catSchema, err := cat.UnmarshalAdditionalAttributes()
		if err != nil {
			c.logger.Warn(fmt.Sprintf("unable to unmarshal category schema %s/%s: %s", entityID, cat.ID, err.Error()))
			continue
		}
		schemas[cat.ID] = *catSchema
	}
	return handlerCats, schemas
}

// buildHybridDSLEntry extends an existing static handler entry with DSL-defined additional schema.
// For hybrid-specialized handlers, DSL categories are assigned as AdditionalCategories.
// For plain hybrid handlers, declaring categories in the DSL is forbidden.
func (c *RegistryCore) buildHybridDSLEntry(entityID, fullID string, def entityDSLFile, existing EndorEntityDictionary) (EndorEntityDictionary, error) {
	additionalSchema, err := def.AdditionalSchema.ToYAML()
	if err != nil {
		return EndorEntityDictionary{}, fmt.Errorf("marshal additionalSchema: %w", err)
	}
	entry := existing
	if specInst, ok := (*existing.OriginalInstance).(sdk.EndorHybridSpecializedHandlerInterface); ok {
		origEntity, ok := existing.entity.(*sdk.EntityHybridSpecialized)
		if !ok {
			return EndorEntityDictionary{}, fmt.Errorf("entity type mismatch for hybrid-specialized %q", entityID)
		}
		addCats, err := toDynamicCategories(def.Categories)
		if err != nil {
			return EndorEntityDictionary{}, err
		}
		cats, catsSchema := c.buildCategorySchemas(fullID, origEntity.Categories)
		entry.EndorHandler = specInst.WithHybridCategories(cats).ToEndorHandler(def.AdditionalSchema, catsSchema, addCats)
		cloned := *origEntity
		cloned.AdditionalSchema = additionalSchema
		cloned.AdditionalCategories = addCats
		entry.entity = &cloned
	} else if hybridInst, ok := (*existing.OriginalInstance).(sdk.EndorHybridHandlerInterface); ok {
		if len(def.Categories) > 0 {
			return EndorEntityDictionary{}, fmt.Errorf("hybrid DSL entity %q must not declare categories: plain hybrid handlers have no categories", entityID)
		}
		origEntity, ok := existing.entity.(*sdk.EntityHybrid)
		if !ok {
			return EndorEntityDictionary{}, fmt.Errorf("entity type mismatch for hybrid %q", entityID)
		}
		entry.EndorHandler = hybridInst.ToEndorHandler(def.AdditionalSchema)
		cloned := *origEntity
		cloned.AdditionalSchema = additionalSchema
		entry.entity = &cloned
	} else {
		return EndorEntityDictionary{}, fmt.Errorf("static handler for %q does not support hybrid DSL extension", entityID)
	}
	return entry, nil
}

// buildDynamicDSLEntry creates a new dictionary entry for a pure DSL entity with no static counterpart.
// If categories are declared, the entity becomes dynamic-specialized; otherwise it is plain dynamic.
func (c *RegistryCore) buildDynamicDSLEntry(fullID string, def entityDSLFile) (EndorEntityDictionary, error) {
	additionalSchema, err := def.AdditionalSchema.ToYAML()
	if err != nil {
		return EndorEntityDictionary{}, fmt.Errorf("marshal additionalSchema: %w", err)
	}
	var entry EndorEntityDictionary
	if len(def.Categories) == 0 {
		schema, _ := sdk.NewSchema(sdk.DynamicEntity{}).ToYAML()
		base := sdk.Entity{
			ID:          fullID,
			Title:       def.Title,
			Description: def.Description,
			Type:        string(sdk.EntityTypeDynamic),
			Module:      c.module,
			Schema:      schema,
		}
		e := &sdk.EntityHybrid{Entity: base, AdditionalSchema: additionalSchema}
		handler := NewEndorHybridHandler[*sdk.DynamicEntity](fullID, def.Title).
			WithExtendedDescription(def.Description).ToEndorHandler(def.AdditionalSchema)
		entry = EndorEntityDictionary{EndorHandler: handler, entity: e}
	} else {
		addCats, err := toDynamicCategories(def.Categories)
		if err != nil {
			return EndorEntityDictionary{}, err
		}
		schema, _ := sdk.NewSchema(sdk.DynamicEntitySpecialized{}).ToYAML()
		base := sdk.Entity{
			ID:          fullID,
			Title:       def.Title,
			Description: def.Description,
			Type:        string(sdk.EntityTypeDynamicSpecialized),
			Module:      c.module,
			Schema:      schema,
		}
		cats, catsSchema := c.buildCategorySchemas(fullID, nil)
		e := &sdk.EntityHybridSpecialized{
			EntityHybrid:         sdk.EntityHybrid{Entity: base, AdditionalSchema: additionalSchema},
			AdditionalCategories: addCats,
		}
		handler := NewEndorHybridSpecializedHandler[*sdk.DynamicEntitySpecialized](fullID, def.Title).
			WithExtendedDescription(def.Description).WithHybridCategories(cats).
			ToEndorHandler(def.AdditionalSchema, catsSchema, addCats)
		entry = EndorEntityDictionary{EndorHandler: handler, entity: e}
	}
	return entry, nil
}

// applyDSLOverlay reads entity DSL files from dao and merges them into dict in-place.
// Returns the Translator built from the DAO's locales path (always non-nil).
func (c *RegistryCore) applyDSLOverlay(dict map[string]EndorEntityDictionary, dao *sdk.DSLDAO) *sdk_i18n.Translator {
	translator := sdk_i18n.NewTranslator(dao.LocalesPath())
	entityIDs, err := dao.ListAllEntities(c.module)
	if err != nil {
		if !os.IsNotExist(err) {
			c.logger.Warn(fmt.Sprintf("unable to list DSL entities: %s", err.Error()))
		}
		return translator
	}
	for _, entityID := range entityIDs {
		content, err := dao.ReadEntity(c.module, entityID)
		if err != nil {
			c.logger.Warn(fmt.Sprintf("unable to read DSL entity %s: %s", entityID, err.Error()))
			continue
		}
		var def entityDSLFile
		if err := yaml.Unmarshal([]byte(content), &def); err != nil {
			c.logger.Warn(fmt.Sprintf("invalid DSL entity %s: %s", entityID, err.Error()))
			continue
		}
		fullID := path.Join(c.module, entityID)
		var entry EndorEntityDictionary
		if existing, ok := dict[fullID]; ok && existing.OriginalInstance != nil {
			entry, err = c.buildHybridDSLEntry(entityID, fullID, def, existing)
		} else {
			entry, err = c.buildDynamicDSLEntry(fullID, def)
		}
		if err != nil {
			c.logger.Warn(fmt.Sprintf("unable to build entry for DSL entity %s: %s", entityID, err.Error()))
			continue
		}
		dict[fullID] = entry
	}
	return translator
}

// buildStaticEntry builds an EndorEntityDictionary from a compiled EndorHandlerInterface.
func (c *RegistryCore) buildStaticEntry(h sdk.EndorHandlerInterface) (EndorEntityDictionary, error) {
	schema, err := h.GetSchema().ToYAML()
	if err != nil {
		c.logger.Warn(fmt.Sprintf("unable to read entity schema from %s", h.GetEntity()))
	}
	base := sdk.Entity{
		ID:          path.Join(c.module, h.GetEntity()),
		Title:       h.GetEntityTitle(),
		Description: h.GetEntityDescription(),
		Module:      c.module,
		Type:        string(sdk.EntityTypeBase),
		Schema:      schema,
	}

	var handler sdk.EndorHandler
	var entity sdk.EntityInterface = &base

	if hs, ok := h.(sdk.EndorHybridSpecializedHandlerInterface); ok {
		base.Type = string(sdk.EntityTypeHybridSpecialized)
		entity = &sdk.EntityHybridSpecialized{
			EntityHybrid: sdk.EntityHybrid{Entity: base},
			Categories:   hs.GetHybridCategories(),
		}
		handler = hs.ToEndorHandler(sdk.RootSchema{}, map[string]sdk.RootSchema{}, []sdk.DynamicCategory{})
	} else if hh, ok := h.(sdk.EndorHybridHandlerInterface); ok {
		base.Type = string(sdk.EntityTypeHybrid)
		entity = &sdk.EntityHybrid{Entity: base}
		handler = hh.ToEndorHandler(sdk.RootSchema{})
	} else if bs, ok := h.(sdk.EndorBaseSpecializedHandlerInterface); ok {
		base.Type = string(sdk.EntityTypeBaseSpecialized)
		entity = &sdk.EntitySpecialized{Entity: base, Categories: bs.GetCategories()}
		handler = bs.ToEndorHandler()
	} else if b, ok := h.(sdk.EndorBaseHandlerInterface); ok {
		handler = b.ToEndorHandler()
	} else {
		return EndorEntityDictionary{}, fmt.Errorf("unknown handler type for entity %s", h.GetEntity())
	}

	hCopy := h
	return EndorEntityDictionary{
		OriginalInstance: &hCopy,
		EndorHandler:     handler,
		entity:           entity,
	}, nil
}

// dictionaryMap builds (and caches) the production handler dictionary.
// Step 1: build entries for all compiled (static) handlers.
// Step 2: apply DSL overlay for hybrid/dynamic entities if enabled.
func (c *RegistryCore) dictionaryMap() (map[string]EndorEntityDictionary, error) {
	c.mu.RLock()
	if c.cacheInitialized {
		result := make(map[string]EndorEntityDictionary, len(c.cachedDictionary))
		for k, v := range c.cachedDictionary {
			result[k] = v
		}
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cacheInitialized {
		result := make(map[string]EndorEntityDictionary, len(c.cachedDictionary))
		for k, v := range c.cachedDictionary {
			result[k] = v
		}
		return result, nil
	}

	dict := map[string]EndorEntityDictionary{}

	if c.internalEndorHandlers != nil {
		for _, h := range *c.internalEndorHandlers {
			entry, err := c.buildStaticEntry(h)
			if err != nil {
				c.logger.Warn(fmt.Sprintf("unable to build static entry for %s: %s", h.GetEntity(), err.Error()))
				continue
			}
			dict[path.Join(c.module, h.GetEntity())] = entry
		}
	}

	c.applyDSLOverlay(dict, c.prodDAO)

	allRepos := collectAllRepositories(sdk.Session{}, dict)
	prodTranslator := sdk_i18n.NewTranslator(c.prodDAO.LocalesPath())

	c.cachedDictionary = dict
	c.cachedDIContainer = &EndorDIContainer{repositories: allRepos, translator: prodTranslator}
	c.cacheInitialized = true
	return dict, nil
}

// buildDevDictionary constructs a development overlay dictionary and DI container for the given session.
func (c *RegistryCore) buildDevDictionary(session sdk.Session) (map[string]EndorEntityDictionary, *EndorDIContainer, error) {
	mainDict, err := c.dictionaryMap()
	if err != nil {
		return nil, nil, err
	}
	devDict := make(map[string]EndorEntityDictionary, len(mainDict))
	for k, v := range mainDict {
		devDict[k] = v
	}
	devDAO := sdk.NewDSLDAO(session.Username, true)
	devTranslator := c.applyDSLOverlay(devDict, devDAO)
	allRepos := collectAllRepositories(session, devDict)
	devContainer := &EndorDIContainer{repositories: allRepos, translator: devTranslator}
	return devDict, devContainer, nil
}

// collectAllRepositories instantiates all repository factories from the dictionary entries.
// An empty container is passed during instantiation; the resulting map is used to
// build the session EndorDIContainer after the dictionary is fully assembled.
func collectAllRepositories(session sdk.Session, dict map[string]EndorEntityDictionary) map[string]sdk.EndorRepositoryInterface {
	allRepos := make(map[string]sdk.EndorRepositoryInterface)
	emptyContainer := &EndorDIContainer{}
	for _, entry := range dict {
		for key, factory := range entry.EndorHandler.RepositoryFactories {
			if factory != nil {
				allRepos[key] = factory(session, emptyContainer)
			}
		}
	}
	return allRepos
}

// endorHandlerList returns all registered EndorHandlers; used internally for route config reload.
func (c *RegistryCore) endorHandlerList() ([]sdk.EndorHandler, error) {
	entities, err := c.dictionaryMap()
	if err != nil {
		return []sdk.EndorHandler{}, err
	}
	list := make([]sdk.EndorHandler, 0, len(entities))
	for _, svc := range entities {
		list = append(list, svc.EndorHandler)
	}
	return list, nil
}

func (c *RegistryCore) reloadRouteConfiguration() error {
	config := sdk_configuration.GetConfig()
	entities, err := c.endorHandlerList()
	if err != nil {
		return err
	}
	if err = api_gateway.InitializeApiGatewayConfiguration(c.microServiceId, c.module, fmt.Sprintf("http://%s:%s", c.microServiceId, config.ServerPort), entities); err != nil {
		return err
	}
	_, err = swagger.CreateSwaggerConfiguration(c.microServiceId, c.module, fmt.Sprintf("http://localhost:%s", config.ServerPort), entities, "/api")
	return err
}

// createAction builds an EndorHandlerActionDictionary from a handler action.
// entityName is the full entity ID (module/version/entity).
func (c *RegistryCore) createAction(entityName string, actionName string, endorServiceAction sdk.EndorHandlerActionInterface) (*EndorHandlerActionDictionary, error) {
	actionId := path.Join(entityName, actionName)
	action := sdk.EntityAction{
		ID:          actionId,
		Entity:      entityName,
		Description: endorServiceAction.GetOptions().Description,
	}
	if endorServiceAction.GetOptions().InputSchema != nil {
		if inputSchema, err := endorServiceAction.GetOptions().InputSchema.ToYAML(); err == nil {
			action.InputSchema = inputSchema
		}
	}
	return &EndorHandlerActionDictionary{
		EndorHandlerAction: endorServiceAction,
		entityAction:       action,
	}, nil
}

// startDslProdWatcher installs an fsnotify watcher on ./prod/ and ./prod/locales/ and calls
// Sync() + reloadRouteConfiguration() with a 1-second debounce on any file-system event.
// If prod/ or prod/locales/ do not exist yet, they are registered as soon as they are created.
func (c *RegistryCore) startDslProdWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	watchRoot := c.prodDAO.BasePath
	for {
		if _, err := os.Stat(watchRoot); err == nil {
			break
		}
		parent := filepath.Dir(watchRoot)
		if parent == watchRoot {
			watcher.Close()
			return fmt.Errorf("no existing ancestor directory found for %s", c.prodDAO.BasePath)
		}
		watchRoot = parent
	}

	if err := watcher.Add(watchRoot); err != nil {
		watcher.Close()
		return err
	}

	// If prod/ already exists, also register locales/ immediately if present.
	localesPath := c.prodDAO.LocalesPath()
	if info, statErr := os.Stat(localesPath); statErr == nil && info.IsDir() {
		_ = watcher.Add(localesPath)
	}

	var (
		debounceMu    sync.Mutex
		debounceTimer *time.Timer
	)

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) {
					if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
						// Register any new directory created under prod/ (covers locales/ created later).
						if isAncestorOrEqual(c.prodDAO.BasePath, event.Name) {
							_ = watcher.Add(event.Name)
						}
					}
				}
				if !isAncestorOrEqual(c.prodDAO.BasePath, event.Name) {
					continue
				}
				capturedName := event.Name
				debounceMu.Lock()
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(1*time.Second, func() {
					c.logger.Info(fmt.Sprintf("prod DSL change detected (%s), syncing registry", capturedName))
					c.Sync()
					if reloadErr := c.reloadRouteConfiguration(); reloadErr != nil {
						c.logger.Warn(fmt.Sprintf("unable to reload route configuration after DSL change: %s", reloadErr.Error()))
					}
				})
				debounceMu.Unlock()

			case watchErr, ok := <-watcher.Errors:
				if !ok {
					return
				}
				c.logger.Warn(fmt.Sprintf("entity DSL watcher error: %s", watchErr.Error()))
			}
		}
	}()
	return nil
}

// isAncestorOrEqual reports whether candidate is a path prefix of (or equal to) target.
func isAncestorOrEqual(candidate, target string) bool {
	rel, err := filepath.Rel(candidate, target)
	return err == nil && !strings.HasPrefix(rel, "..")
}

// #endregion
