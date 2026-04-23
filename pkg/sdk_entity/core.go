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
	"gopkg.in/yaml.v3"
)

var (
	registryCoreInstance *RegistryCore
	registryCoreOnce     sync.Once
)

func GetRegistryCore() *RegistryCore {
	return registryCoreInstance
}

func InitRegistryCore(microServiceId string, internalEndorHandlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger) *RegistryCore {
	registryCoreOnce.Do(func() {
		absProdRoot, _ := filepath.Abs("prod")
		registryCoreInstance = &RegistryCore{
			microServiceId:        microServiceId,
			internalEndorHandlers: internalEndorHandlers,
			logger:                logger,
			mu:                    &sync.RWMutex{},
			prodDAO:               &sdk.DSLDAO{BasePath: absProdRoot},
			ephemeralCache:        newEphemeralCacheManager(),
		}
		if err := registryCoreInstance.startEntityWatcher(); err != nil {
			logger.Warn(fmt.Sprintf("unable to start entity DSL watcher: %s", err.Error()))
		}
	})
	return registryCoreInstance
}

// RegistryCore manages building and caching the handler registry.
// Production requests use a cached dictionary; development requests (x-development: true)
// get a per-user ephemeral overlay built on top of the production dictionary.
type RegistryCore struct {
	microServiceId        string
	internalEndorHandlers *[]sdk.EndorHandlerInterface
	prodDAO               *sdk.DSLDAO
	logger                *sdk.Logger
	mu                    *sync.RWMutex
	cachedDictionary      map[string]EndorEntityDictionary
	cacheInitialized      bool
	ephemeralCache        *EphemeralCacheManager
}

// EndorEntityDictionary is the per-entity DI container: compiled handler, entity metadata,
// and instantiated repositories.
type EndorEntityDictionary struct {
	OriginalInstance *sdk.EndorHandlerInterface
	EndorHandler     sdk.EndorHandler
	entity           sdk.EntityInterface
	repositories     map[string]sdk.EndorRepositoryInterface
}

func (c *EndorEntityDictionary) GetRepositories() map[string]sdk.EndorRepositoryInterface {
	return c.repositories
}

type EndorHandlerActionDictionary struct {
	EndorHandlerAction sdk.EndorHandlerActionInterface
	entityAction       sdk.EntityAction
	Container          EndorEntityDictionary
}

// entityDSLFile is the YAML structure for entity definition files.
type entityDSLFile struct {
	Description          string                `yaml:"description"`
	Type                 string                `yaml:"type"`
	AdditionalSchema     string                `yaml:"additionalSchema"`
	Categories           []sdk.HybridCategory  `yaml:"categories"`
	AdditionalCategories []sdk.DynamicCategory `yaml:"additionalCategories"`
}

// #region Public API

// Dictionary returns the handler dictionary for the given session.
// Production sessions get the cached dictionary; development sessions get a per-user
// ephemeral overlay (falls back to production on error).
func (c *RegistryCore) Dictionary(session sdk.Session) (map[string]EndorEntityDictionary, error) {
	if !session.Development || session.Username == "" {
		return c.dictionaryMap()
	}
	if cached := c.ephemeralCache.Get(session.Username); cached != nil {
		return cached, nil
	}
	devDict, err := c.buildDevDictionary(session)
	if err != nil {
		c.logger.Warn(fmt.Sprintf("failed to build ephemeral registry, falling back to prod: %s", err.Error()))
		return c.dictionaryMap()
	}
	c.ephemeralCache.Set(session.Username, devDict)
	return devDict, nil
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
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id)).WithTranslation("entities.entity.not_found", map[string]any{"id": dto.Id})
}

// DynamicEntityList returns all entity definitions read from the production DSL filesystem.
func (c *RegistryCore) DynamicEntityList() ([]sdk.EntityInterface, error) {
	entityIDs, err := c.prodDAO.ListAllEntities(c.microServiceId)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entities []sdk.EntityInterface
	for _, entityID := range entityIDs {
		content, err := c.prodDAO.ReadEntity(c.microServiceId, entityID)
		if err != nil {
			c.logger.Warn(fmt.Sprintf("unable to read entity DSL file %s: %s", entityID, err.Error()))
			continue
		}
		entity, err := c.parseEntityDSL(entityID, content)
		if err != nil {
			c.logger.Warn(fmt.Sprintf("unable to parse entity DSL file %s: %s", entityID, err.Error()))
			continue
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

// Sync invalidates the production cache and all per-user ephemeral overlays,
// forcing a full rebuild on the next access.
func (c *RegistryCore) Sync() {
	c.mu.Lock()
	c.cacheInitialized = false
	c.cachedDictionary = nil
	c.mu.Unlock()
	c.ephemeralCache.InvalidateAll()
}

// #endregion

// #region Internal machinery

// parseEntityDSL parses a raw YAML DSL string into an EntityInterface.
func (c *RegistryCore) parseEntityDSL(entityID, content string) (sdk.EntityInterface, error) {
	var def entityDSLFile
	if err := yaml.Unmarshal([]byte(content), &def); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	base := sdk.Entity{
		ID:          entityID,
		Description: def.Description,
		Type:        def.Type,
		Service:     c.microServiceId,
	}
	switch sdk.EntityType(def.Type) {
	case sdk.EntityTypeHybrid, sdk.EntityTypeDynamic:
		return &sdk.EntityHybrid{Entity: base, AdditionalSchema: def.AdditionalSchema}, nil
	case sdk.EntityTypeHybridSpecialized, sdk.EntityTypeDynamicSpecialized:
		return &sdk.EntityHybridSpecialized{
			EntityHybrid:         sdk.EntityHybrid{Entity: base, AdditionalSchema: def.AdditionalSchema},
			Categories:           def.Categories,
			AdditionalCategories: def.AdditionalCategories,
		}, nil
	default:
		return nil, fmt.Errorf("unknown entity type %q", def.Type)
	}
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

// buildDSLEntry creates or updates an EndorEntityDictionary from a parsed DSL entity.
// If existing is non-nil and has a compiled OriginalInstance, the handler is re-configured
// using the DSL definition. Otherwise a pure-dynamic handler is created.
// Repositories are always rebuilt to match the resulting handler and session.
func (c *RegistryCore) buildDSLEntry(session sdk.Session, entity sdk.EntityInterface, existing *EndorEntityDictionary) (EndorEntityDictionary, error) {
	switch e := entity.(type) {
	case *sdk.EntityHybrid:
		return c.buildHybridEntry(session, e, existing)
	case *sdk.EntityHybridSpecialized:
		return c.buildHybridSpecializedEntry(session, e, existing)
	default:
		return EndorEntityDictionary{}, fmt.Errorf("unsupported entity type %T", entity)
	}
}

func (c *RegistryCore) buildHybridEntry(session sdk.Session, e *sdk.EntityHybrid, existing *EndorEntityDictionary) (EndorEntityDictionary, error) {
	definition, err := e.UnmarshalAdditionalAttributes()
	if err != nil {
		return EndorEntityDictionary{}, fmt.Errorf("unmarshal error: %w", err)
	}

	var entry EndorEntityDictionary
	if existing != nil && existing.OriginalInstance != nil {
		if hybridInst, ok := (*existing.OriginalInstance).(sdk.EndorHybridHandlerInterface); ok {
			entry = *existing
			entry.EndorHandler = hybridInst.ToEndorHandler(*definition)
			if orig, ok := entry.entity.(*sdk.EntityHybrid); ok {
				cloned := *orig
				cloned.AdditionalSchema = e.AdditionalSchema
				entry.entity = &cloned
			}
		}
	} else {
		schema, _ := sdk.NewSchema(sdk.DynamicEntity{}).ToYAML()
		e.Schema = schema
		handler := NewEndorHybridHandler[*sdk.DynamicEntity](e.ID, e.Description).ToEndorHandler(*definition)
		entry = EndorEntityDictionary{EndorHandler: handler, entity: e}
	}
	entry.repositories = buildRepositoriesFromFactories(session, &entry, entry.EndorHandler.RepositoryFactories)
	return entry, nil
}

func (c *RegistryCore) buildHybridSpecializedEntry(session sdk.Session, e *sdk.EntityHybridSpecialized, existing *EndorEntityDictionary) (EndorEntityDictionary, error) {
	definition, err := e.UnmarshalAdditionalAttributes()
	if err != nil {
		return EndorEntityDictionary{}, fmt.Errorf("unmarshal error: %w", err)
	}
	cats, catsSchema := c.buildCategorySchemas(e.ID, e.Categories)

	var entry EndorEntityDictionary
	if existing != nil && existing.OriginalInstance != nil {
		if specInst, ok := (*existing.OriginalInstance).(sdk.EndorHybridSpecializedHandlerInterface); ok {
			entry = *existing
			entry.EndorHandler = specInst.WithHybridCategories(cats).ToEndorHandler(*definition, catsSchema, e.AdditionalCategories)
			if orig, ok := entry.entity.(*sdk.EntityHybridSpecialized); ok {
				cloned := *orig
				cloned.AdditionalSchema = e.AdditionalSchema
				cloned.AdditionalCategories = e.AdditionalCategories
				entry.entity = &cloned
			}
		}
	} else {
		schema, _ := sdk.NewSchema(sdk.DynamicEntitySpecialized{}).ToYAML()
		e.Schema = schema
		handler := NewEndorHybridSpecializedHandler[*sdk.DynamicEntitySpecialized](e.ID, e.Description).
			WithHybridCategories(cats).ToEndorHandler(*definition, catsSchema, e.AdditionalCategories)
		entry = EndorEntityDictionary{EndorHandler: handler, entity: e}
	}
	entry.repositories = buildRepositoriesFromFactories(session, &entry, entry.EndorHandler.RepositoryFactories)
	return entry, nil
}

// applyDSLOverlay reads entity DSL files from dao and merges them into dict in-place.
func (c *RegistryCore) applyDSLOverlay(session sdk.Session, dict map[string]EndorEntityDictionary, dao *sdk.DSLDAO) {
	entityIDs, err := dao.ListAllEntities(c.microServiceId)
	if err != nil {
		if !os.IsNotExist(err) {
			c.logger.Warn(fmt.Sprintf("unable to list DSL entities: %s", err.Error()))
		}
		return
	}
	for _, entityID := range entityIDs {
		content, err := dao.ReadEntity(c.microServiceId, entityID)
		if err != nil {
			c.logger.Warn(fmt.Sprintf("unable to read DSL entity %s: %s", entityID, err.Error()))
			continue
		}
		entity, err := c.parseEntityDSL(entityID, content)
		if err != nil {
			c.logger.Warn(fmt.Sprintf("invalid DSL entity %s: %s", entityID, err.Error()))
			continue
		}
		var existingPtr *EndorEntityDictionary
		if existing, ok := dict[entityID]; ok {
			existingPtr = &existing
		}
		entry, err := c.buildDSLEntry(session, entity, existingPtr)
		if err != nil {
			c.logger.Warn(fmt.Sprintf("unable to build entry for DSL entity %s: %s", entityID, err.Error()))
			continue
		}
		dict[entityID] = entry
	}
}

// buildStaticEntry builds an EndorEntityDictionary from a compiled EndorHandlerInterface.
func (c *RegistryCore) buildStaticEntry(h sdk.EndorHandlerInterface) (EndorEntityDictionary, error) {
	schema, err := h.GetSchema().ToYAML()
	if err != nil {
		c.logger.Warn(fmt.Sprintf("unable to read entity schema from %s", h.GetEntity()))
	}
	base := sdk.Entity{
		ID:          h.GetEntity(),
		Description: h.GetEntityDescription(),
		Service:     c.microServiceId,
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
	entry := EndorEntityDictionary{
		OriginalInstance: &hCopy,
		EndorHandler:     handler,
		entity:           entity,
	}
	entry.repositories = buildRepositoriesFromFactories(sdk.Session{}, &entry, handler.RepositoryFactories)
	return entry, nil
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
			dict[h.GetEntity()] = entry
		}
	}

	if sdk_configuration.GetConfig().HybridEntitiesEnabled || sdk_configuration.GetConfig().DynamicEntitiesEnabled {
		c.applyDSLOverlay(sdk.Session{}, dict, c.prodDAO)
	}

	injectAllRepositories(dict)

	c.cachedDictionary = dict
	c.cacheInitialized = true
	return dict, nil
}

// buildDevDictionary constructs a development overlay dictionary for the given session.
// It shallow-clones the production dictionary and applies DSL entities from the user's
// dev path on top, rebuilding repositories with the dev session for any modified entry.
func (c *RegistryCore) buildDevDictionary(session sdk.Session) (map[string]EndorEntityDictionary, error) {
	mainDict, err := c.dictionaryMap()
	if err != nil {
		return nil, err
	}
	devDict := make(map[string]EndorEntityDictionary, len(mainDict))
	for k, v := range mainDict {
		devDict[k] = v
	}
	c.applyDSLOverlay(session, devDict, sdk.NewDSLDAO(session.Username, true))
	injectAllRepositories(devDict)
	return devDict, nil
}

// injectAllRepositories collects every repository from every entry in the dictionary
// and sets the merged map as the repositories of each container, so that any action
// can reach any registered repository via GetDynamicRepository / GetStaticRepository.
func injectAllRepositories(dict map[string]EndorEntityDictionary) {
	allRepos := make(map[string]sdk.EndorRepositoryInterface)
	for _, entry := range dict {
		for k, v := range entry.repositories {
			allRepos[k] = v
		}
	}
	for entityID, entry := range dict {
		entry.repositories = allRepos
		dict[entityID] = entry
	}
}

// buildRepositoriesFromFactories instantiates repositories from handler factories.
func buildRepositoriesFromFactories(session sdk.Session, container sdk.EndorDIContainerInterface, factories map[string]sdk.RepositoryFactory) map[string]sdk.EndorRepositoryInterface {
	repos := make(map[string]sdk.EndorRepositoryInterface, len(factories))
	for key, factory := range factories {
		if factory != nil {
			repos[key] = factory(session, container)
		}
	}
	return repos
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

func (c *RegistryCore) reloadRouteConfiguration(microserviceId string) error {
	config := sdk_configuration.GetConfig()
	entities, err := c.endorHandlerList()
	if err != nil {
		return err
	}
	if err = api_gateway.InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), entities); err != nil {
		return err
	}
	_, err = swagger.CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), entities, "/api")
	return err
}

// createAction builds an EndorHandlerActionDictionary from a handler action.
func (c *RegistryCore) createAction(entityName string, version string, actionName string, endorServiceAction sdk.EndorHandlerActionInterface) (*EndorHandlerActionDictionary, error) {
	if version == "" {
		version = "v1"
	}
	actionId := path.Join(c.microServiceId, version, entityName, actionName)
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

// startEntityWatcher installs an fsnotify watcher on ./prod/ and calls
// Sync() + reloadRouteConfiguration() with a 1-second debounce on any file-system event.
func (c *RegistryCore) startEntityWatcher() error {
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
					if isAncestorOrEqual(c.prodDAO.BasePath, event.Name) {
						if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
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
					if reloadErr := c.reloadRouteConfiguration(c.microServiceId); reloadErr != nil {
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
