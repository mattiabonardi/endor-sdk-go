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

// Singleton instance and initialization sync
var (
	registryCoreInstance *RegistryCore
	registryCoreOnce     sync.Once
)

// GetRegistryCore returns the singleton RegistryCore engine instance.
func GetRegistryCore() *RegistryCore {
	return registryCoreInstance
}

// InitRegistryCore initializes the singleton RegistryCore engine.
// This should be called once during application startup. Subsequent calls are no-ops.
//
// The engine always scans ./prod/ui/entities/<microServiceId>/ for production DSL
// entity definitions and installs an fsnotify watcher on ./prod/ with a 1-second debounce.
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

// RegistryCore is the internal engine that manages building and caching the handler registry.
// It resolves the correct handler dictionary based on the given session: the production
// dictionary is returned for normal requests, while a per-user ephemeral overlay is
// returned when session.Development == true.
type RegistryCore struct {
	microServiceId        string
	internalEndorHandlers *[]sdk.EndorHandlerInterface
	// prodDAO is scoped to ./prod/ and used to read production entity DSL files.
	prodDAO          *sdk.DSLDAO
	logger           *sdk.Logger
	mu               *sync.RWMutex
	cachedDictionary map[string]EndorDIContainer
	cacheInitialized bool
	// ephemeralCache holds per-user ephemeral (debug) overlay registries.
	ephemeralCache *EphemeralCacheManager
}

// EndorDIContainer is the per-entity dependency injection container produced by RegistryCore.
// It holds the entity's compiled handler, its entity metadata, and a snapshot of all
// registered repositories available at container build time.
// Use the generic helpers GetHandlerFromContainer and GetRepositoryFromContainer for
// type-safe access from within action handlers.
type EndorDIContainer struct {
	OriginalInstance *sdk.EndorHandlerInterface
	EndorHandler     sdk.EndorHandler
	entity           sdk.EntityInterface
	repositories     map[string]sdk.EndorRepositoryInterface
}

// GetRepositoryByName returns the repository registered under name.
// Implements sdk.EndorDIContainerInterface.
func (c *EndorDIContainer) GetRepositoryByName(name string) (sdk.EndorRepositoryInterface, bool) {
	repo, ok := c.repositories[name]
	return repo, ok
}

// GetHandlerByName returns the original handler instance for the given entity name.
// Implements sdk.EndorDIContainerInterface.
func (c *EndorDIContainer) GetHandlerByName(name string) (sdk.EndorHandlerInterface, bool) {
	if c.OriginalInstance != nil {
		h := *c.OriginalInstance
		if h.GetEntity() == name {
			return h, true
		}
	}
	return nil, false
}

type EndorHandlerActionDictionary struct {
	EndorHandlerAction sdk.EndorHandlerActionInterface
	entityAction       sdk.EntityAction
	// Container is the DI container for the entity that owns this action.
	Container EndorDIContainer
}

// entityDSLFile is the YAML structure used to deserialize entity definition files from the DSL.
type entityDSLFile struct {
	Description          string                `yaml:"description"`
	Type                 string                `yaml:"type"`
	AdditionalSchema     string                `yaml:"additionalSchema"`
	Categories           []sdk.HybridCategory  `yaml:"categories"`
	AdditionalCategories []sdk.DynamicCategory `yaml:"additionalCategories"`
}

// #region Public API

// Dictionary returns the handler dictionary appropriate for the given session.
//
//   - session.Development == false → production dictionary (cached)
//   - session.Development == true  → per-user ephemeral overlay (dev DSL on top of prod)
//
// On any error building the ephemeral overlay the method falls back to the production
// dictionary and logs a warning.
func (c *RegistryCore) Dictionary(session sdk.Session) (map[string]EndorDIContainer, error) {
	if !session.Development {
		return c.dictionaryMap()
	}
	userID := session.Username
	if userID == "" {
		// Debug mode without a user ID is not allowed; fall back to production.
		return c.dictionaryMap()
	}

	// Fast path: cached ephemeral registry.
	if cached := c.ephemeralCache.Get(userID); cached != nil {
		return cached, nil
	}

	// Slow path: build the overlay and populate the cache.
	devDict, err := c.buildDevDictionary(userID)
	if err != nil {
		c.logger.Warn(fmt.Sprintf("failed to build ephemeral registry, falling back to prod: %s", err.Error()))
		return c.dictionaryMap()
	}
	c.ephemeralCache.Set(userID, devDict)
	return devDict, nil
}

// DictionaryInstance resolves a single registry entry by ID for the given session.
func (c *RegistryCore) DictionaryInstance(session sdk.Session, dto sdk.ReadInstanceDTO) (*EndorDIContainer, error) {
	dict, err := c.Dictionary(session)
	if err != nil {
		return nil, err
	}
	if entry, ok := dict[dto.Id]; ok {
		return &entry, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id)).WithTranslation("entities.entity.not_found", map[string]any{"id": dto.Id})
}

// DynamicEntityList reads entity definitions from the DSL filesystem path and returns them
// as EntityInterface instances. Each file in the entity DSL path represents one entity;
// the filename (without extension) is used as the entity ID.
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

		var def entityDSLFile
		if err := yaml.Unmarshal([]byte(content), &def); err != nil {
			c.logger.Warn(fmt.Sprintf("unable to parse entity DSL file %s: %s", entityID, err.Error()))
			continue
		}

		baseEntity := sdk.Entity{
			ID:          entityID,
			Description: def.Description,
			Type:        def.Type,
			Service:     c.microServiceId,
		}

		switch sdk.EntityType(def.Type) {
		case sdk.EntityTypeHybrid, sdk.EntityTypeDynamic:
			entities = append(entities, &sdk.EntityHybrid{
				Entity:           baseEntity,
				AdditionalSchema: def.AdditionalSchema,
			})
		case sdk.EntityTypeHybridSpecialized, sdk.EntityTypeDynamicSpecialized:
			entities = append(entities, &sdk.EntityHybridSpecialized{
				EntityHybrid: sdk.EntityHybrid{
					Entity:           baseEntity,
					AdditionalSchema: def.AdditionalSchema,
				},
				Categories:           def.Categories,
				AdditionalCategories: def.AdditionalCategories,
			})
		default:
			c.logger.Warn(fmt.Sprintf("unknown entity type %q in DSL file %s", def.Type, entityID))
		}
	}

	return entities, nil
}

// Sync invalidates the production registry cache, forcing a full rebuild from the
// DSL filesystem (./prod/ui/entities/<ms-id>/) on the next access. It also invalidates
// all per-user ephemeral (debug) registries, since they are overlays of the production data.
func (c *RegistryCore) Sync() {
	c.mu.Lock()
	c.cacheInitialized = false
	c.cachedDictionary = nil
	c.mu.Unlock()
	c.ephemeralCache.InvalidateAll()
}

// #endregion

// #region DI Container helpers

// GetHandlerFromContainer returns the entity's compiled handler instance cast to T.
// Use this from within action handlers to access the strongly-typed handler stored in
// the DI container attached to EndorContext.
//
// Example:
//
//	h, ok := sdk_entity.GetHandlerFromContainer[*MyHandler](ec.DIContainer)
func GetHandlerFromContainer[T sdk.EndorHandlerInterface](c sdk.EndorDIContainerInterface) (T, bool) {
	container, ok := c.(*EndorDIContainer)
	if !ok {
		var zero T
		return zero, false
	}
	if container.OriginalInstance == nil {
		var zero T
		return zero, false
	}
	typed, ok := (*container.OriginalInstance).(T)
	return typed, ok
}

// GetRepositoryFromContainer returns the repository registered under name, cast to T.
// Use this from within action handlers to obtain a strongly-typed repository from
// the DI container attached to EndorContext.
//
// Example:
//
//	repo, ok := sdk_entity.GetRepositoryFromContainer[MyRepositoryInterface](ec.DIContainer, "myEntity")
func GetRepositoryFromContainer[T sdk.EndorRepositoryInterface](c sdk.EndorDIContainerInterface, name string) (T, bool) {
	repo, ok := c.GetRepositoryByName(name)
	if !ok {
		var zero T
		return zero, false
	}
	typed, ok := repo.(T)
	return typed, ok
}

// #endregion

// #region Internal machinery

// dictionaryMap builds (and caches) the production handler dictionary from compiled handlers
// and DSL-defined dynamic/hybrid entities.
func (c *RegistryCore) dictionaryMap() (map[string]EndorDIContainer, error) {
	// Check cache first with read lock
	c.mu.RLock()
	if c.cacheInitialized {
		// Return a copy of the cached dictionary to prevent external modifications
		result := make(map[string]EndorDIContainer, len(c.cachedDictionary))
		for k, v := range c.cachedDictionary {
			result[k] = v
		}
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	// Acquire write lock to build the dictionary
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.cacheInitialized {
		// Return a copy of the cached dictionary
		result := make(map[string]EndorDIContainer, len(c.cachedDictionary))
		for k, v := range c.cachedDictionary {
			result[k] = v
		}
		return result, nil
	}

	entities := map[string]EndorDIContainer{}

	// internal EndorHandlers
	if c.internalEndorHandlers != nil {
		for _, internalEndorHandler := range *c.internalEndorHandlers {
			schema, err := internalEndorHandler.GetSchema().ToYAML()
			if err != nil {
				c.logger.Warn(fmt.Sprintf("unable to read entity schema from %s", internalEndorHandler.GetEntity()))
			}
			baseEntity := sdk.Entity{
				ID:          internalEndorHandler.GetEntity(),
				Description: internalEndorHandler.GetEntityDescription(),
				Service:     c.microServiceId,
				Type:        string(sdk.EntityTypeBase),
				Schema:      schema,
			}

			var endorService sdk.EndorHandler
			var entity sdk.EntityInterface = &baseEntity

			// hybrid specialized
			if hybridSpecializedService, ok := internalEndorHandler.(sdk.EndorHybridSpecializedHandlerInterface); ok {
				baseEntity.Type = string(sdk.EntityTypeHybridSpecialized)
				hybridSpecializedEntity := sdk.EntityHybridSpecialized{
					EntityHybrid: sdk.EntityHybrid{
						Entity:           baseEntity,
						AdditionalSchema: "",
					},
					Categories: hybridSpecializedService.GetHybridCategories(),
				}
				entity = &hybridSpecializedEntity
				endorService = hybridSpecializedService.ToEndorHandler(sdk.RootSchema{}, map[string]sdk.RootSchema{}, []sdk.DynamicCategory{})
			} else {
				// hybrid
				if hybridService, ok := internalEndorHandler.(sdk.EndorHybridHandlerInterface); ok {
					baseEntity.Type = string(sdk.EntityTypeHybrid)
					endorService = hybridService.ToEndorHandler(sdk.RootSchema{})
					hybridEntity := sdk.EntityHybrid{
						Entity: baseEntity,
					}
					entity = &hybridEntity
				} else {
					// base specialized
					if baseSpecializedService, ok := internalEndorHandler.(sdk.EndorBaseSpecializedHandlerInterface); ok {
						baseEntity.Type = string(sdk.EntityTypeBaseSpecialized)
						baseSpecializedEntity := sdk.EntitySpecialized{
							Entity:     baseEntity,
							Categories: baseSpecializedService.GetCategories(),
						}
						entity = &baseSpecializedEntity
						endorService = baseSpecializedService.ToEndorHandler()
					} else {
						// base
						if baseService, ok := internalEndorHandler.(sdk.EndorBaseHandlerInterface); ok {
							endorService = baseService.ToEndorHandler()
						} else {
							c.logger.Warn(fmt.Sprintf("unable to create entity %s from service", internalEndorHandler.GetEntity()))
						}
					}
				}
			}

			entities[internalEndorHandler.GetEntity()] = EndorDIContainer{
				OriginalInstance: &internalEndorHandler,
				EndorHandler:     endorService,
				entity:           entity,
			}
		}
	}

	// dynamic EndorHandlers
	if sdk_configuration.GetConfig().HybridEntitiesEnabled || sdk_configuration.GetConfig().DynamicEntitiesEnabled {
		dynamicEntities, err := c.DynamicEntityList()
		if err != nil {
			return map[string]EndorDIContainer{}, nil
		}

		for _, entity := range dynamicEntities {
			entityID := entity.GetID().(string)
			if v, ok := entities[entityID]; ok {
				// check entity hybrid
				if entityHybrid, ok := entity.(*sdk.EntityHybrid); ok {
					if hybridInstance, ok := (*v.OriginalInstance).(sdk.EndorHybridHandlerInterface); ok {
						defintion, err := entityHybrid.UnmarshalAdditionalAttributes()
						if err != nil {
							c.logger.Warn(fmt.Sprintf("unable to unmarshal definition for hybrid entity %s: %s", entityHybrid.ID, err.Error()))
						}
						v.EndorHandler = hybridInstance.ToEndorHandler(*defintion)
						if originalEntity, ok := v.entity.(*sdk.EntityHybrid); ok {
							originalEntity.AdditionalSchema = entityHybrid.AdditionalSchema
						}
						entities[entityID] = v
					}
				}

				// check entity specialized
				if entitySpecialized, ok := entity.(*sdk.EntityHybridSpecialized); ok {
					if specializedInstance, ok := (*v.OriginalInstance).(sdk.EndorHybridSpecializedHandlerInterface); ok {
						defintion, err := entitySpecialized.UnmarshalAdditionalAttributes()
						if err != nil {
							c.logger.Warn(fmt.Sprintf("unable to unmarshal definition for hybrid specialized entity %s: %s", entitySpecialized.ID, err.Error()))
						}
						categories := []sdk.EndorHybridSpecializedHandlerCategoryInterface{}
						categoriesAdditionalSchema := map[string]sdk.RootSchema{}
						for _, cat := range entitySpecialized.Categories {
							categories = append(categories, NewEndorHybridSpecializedHandlerCategory[*sdk.DynamicEntitySpecialized](cat.ID, cat.Description))
							categoryAdditionalSchema, err := cat.UnmarshalAdditionalAttributes()
							if err != nil {
								c.logger.Warn(fmt.Sprintf("unable to unmarshal category definition %s for hybrid entity %s: %s", cat.ID, entitySpecialized.ID, err.Error()))
							}
							categoriesAdditionalSchema[cat.ID] = *categoryAdditionalSchema
						}
						v.EndorHandler = specializedInstance.WithHybridCategories(categories).ToEndorHandler(*defintion, categoriesAdditionalSchema, entitySpecialized.AdditionalCategories)
						if originalEntity, ok := v.entity.(*sdk.EntityHybridSpecialized); ok {
							originalEntity.AdditionalCategories = entitySpecialized.AdditionalCategories
							originalEntity.AdditionalSchema = entitySpecialized.AdditionalSchema
						}
						entities[entityID] = v
					}
				}
			} else {
				if entityHybrid, ok := entity.(*sdk.EntityHybrid); ok {
					defintion, err := entityHybrid.UnmarshalAdditionalAttributes()
					if err != nil {
						c.logger.Warn(fmt.Sprintf("unable to unmarshal definition for dynamic entity %s: %s", entityHybrid.ID, err.Error()))
					}
					hybridService := NewEndorHybridHandler[*sdk.DynamicEntity](entityHybrid.ID, entityHybrid.Description)
					schema, _ := sdk.NewSchema(sdk.DynamicEntity{}).ToYAML()
					entityHybrid.Schema = schema
					entities[entityHybrid.ID] = EndorDIContainer{
						EndorHandler: hybridService.ToEndorHandler(*defintion),
						entity:       entityHybrid,
					}
				}
				if entitySpecialized, ok := entity.(*sdk.EntityHybridSpecialized); ok {
					defintion, err := entitySpecialized.UnmarshalAdditionalAttributes()
					if err != nil {
						c.logger.Warn(fmt.Sprintf("unable to unmarshal definition for dynamic specialized entity %s: %s", entitySpecialized.ID, err.Error()))
					}
					categories := []sdk.EndorHybridSpecializedHandlerCategoryInterface{}
					categoriesAdditionalSchema := map[string]sdk.RootSchema{}
					for _, cat := range entitySpecialized.Categories {
						categories = append(categories, NewEndorHybridSpecializedHandlerCategory[*sdk.DynamicEntitySpecialized](cat.ID, cat.Description))
						categoryAdditionalSchema, err := cat.UnmarshalAdditionalAttributes()
						if err != nil {
							c.logger.Warn(fmt.Sprintf("unable to unmarshal category definition %s for dynamic specialized entity %s: %s", cat.ID, entitySpecialized.ID, err.Error()))
						}
						categoriesAdditionalSchema[cat.ID] = *categoryAdditionalSchema
					}
					hybridService := NewEndorHybridSpecializedHandler[*sdk.DynamicEntitySpecialized](entitySpecialized.ID, entitySpecialized.Description).WithHybridCategories(categories)
					schema, _ := sdk.NewSchema(sdk.DynamicEntitySpecialized{}).ToYAML()
					entitySpecialized.Schema = schema
					entities[entitySpecialized.ID] = EndorDIContainer{
						EndorHandler: hybridService.ToEndorHandler(*defintion, categoriesAdditionalSchema, entitySpecialized.AdditionalCategories),
						entity:       entitySpecialized,
					}
				}
			}
		}
	}

	// Snapshot repositories from the global registry and inject them into each container.
	repoSnapshot := sdk.GetRepositoryRegistry().Snapshot()
	for entityID, entry := range entities {
		entry.repositories = repoSnapshot
		entities[entityID] = entry
	}

	// Cache the result
	c.cachedDictionary = entities
	c.cacheInitialized = true

	return entities, nil
}

// endorHandlerList returns all registered EndorHandlers; used internally for route config reload.
func (c *RegistryCore) endorHandlerList() ([]sdk.EndorHandler, error) {
	entities, err := c.dictionaryMap()
	if err != nil {
		return []sdk.EndorHandler{}, err
	}
	entityList := make([]sdk.EndorHandler, 0, len(entities))
	for _, service := range entities {
		entityList = append(entityList, service.EndorHandler)
	}
	return entityList, nil
}

func (c *RegistryCore) reloadRouteConfiguration(microserviceId string) error {
	config := sdk_configuration.GetConfig()
	entities, err := c.endorHandlerList()
	if err != nil {
		return err
	}
	err = api_gateway.InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), entities)
	if err != nil {
		return err
	}
	_, err = swagger.CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), entities, "/api")
	if err != nil {
		return err
	}
	return nil
}

// createAction builds an EndorHandlerActionDictionary from a handler action.
// Accessible within the package for use by EndorHandlerActionRepository.
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
		inputSchema, err := endorServiceAction.GetOptions().InputSchema.ToYAML()
		if err == nil {
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

	// Start watching from prodDAO.BasePath; walk up to find the deepest existing ancestor.
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

				// Progressively watch newly-created sub-directories within the prod tree.
				if event.Has(fsnotify.Create) {
					if isAncestorOrEqual(c.prodDAO.BasePath, event.Name) {
						if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
							_ = watcher.Add(event.Name)
						}
					}
				}

				// Only react to events inside the prod tree.
				if !isAncestorOrEqual(c.prodDAO.BasePath, event.Name) {
					continue
				}

				// Debounce: collapse rapid bursts of changes into a single sync call.
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

// buildDevDictionary constructs a debug overlay dictionary for the given user.
//
// Algorithm:
//  1. Shallow-clone the production dictionary.
//  2. Scan ./dev/<userID>/ui/entities/<ms-id>/ for YAML files.
//  3. For each valid file, override (or add) the corresponding entity in the clone.
//  4. Corrupt/unreadable dev YAML files are skipped; the production entry (if any) is preserved.
func (c *RegistryCore) buildDevDictionary(userID string) (map[string]EndorDIContainer, error) {
	// Build (and cache) the production dict first.
	mainDict, err := c.dictionaryMap()
	if err != nil {
		return nil, err
	}

	// Shallow clone: struct-value copy, no production map is mutated.
	devDict := make(map[string]EndorDIContainer, len(mainDict))
	for k, v := range mainDict {
		devDict[k] = v
	}

	devDAO := sdk.NewDSLDAO(userID, true)
	entityIDs, err := devDAO.ListAllEntities(c.microServiceId)
	if err != nil {
		if os.IsNotExist(err) {
			return devDict, nil
		}
		c.logger.Warn(fmt.Sprintf("unable to read dev entity dir: %s", err.Error()))
		return devDict, nil
	}

	for _, entityID := range entityIDs {
		content, err := devDAO.ReadEntity(c.microServiceId, entityID)
		if err != nil {
			c.logger.Warn(fmt.Sprintf("unable to read dev entity DSL %s: %s", entityID, err.Error()))
			continue
		}

		var def entityDSLFile
		if err := yaml.Unmarshal([]byte(content), &def); err != nil {
			c.logger.Warn(fmt.Sprintf("corrupt dev entity DSL %s, keeping prod version: %s", entityID, err.Error()))
			continue
		}

		baseEntity := sdk.Entity{
			ID:          entityID,
			Description: def.Description,
			Type:        def.Type,
			Service:     c.microServiceId,
		}

		var devEntity sdk.EntityInterface
		switch sdk.EntityType(def.Type) {
		case sdk.EntityTypeHybrid, sdk.EntityTypeDynamic:
			devEntity = &sdk.EntityHybrid{
				Entity:           baseEntity,
				AdditionalSchema: def.AdditionalSchema,
			}
		case sdk.EntityTypeHybridSpecialized, sdk.EntityTypeDynamicSpecialized:
			devEntity = &sdk.EntityHybridSpecialized{
				EntityHybrid: sdk.EntityHybrid{
					Entity:           baseEntity,
					AdditionalSchema: def.AdditionalSchema,
				},
				Categories:           def.Categories,
				AdditionalCategories: def.AdditionalCategories,
			}
		default:
			c.logger.Warn(fmt.Sprintf("unknown entity type %q in dev DSL %s", def.Type, entityID))
			continue
		}

		existing, existsInProd := devDict[entityID]

		var handlerDict EndorDIContainer

		if existsInProd && existing.OriginalInstance != nil {
			// Override: re-merge dev DSL with the original compiled handler.
			handlerDict = existing

			if entityHybrid, ok := devEntity.(*sdk.EntityHybrid); ok {
				if hybridInst, ok := (*existing.OriginalInstance).(sdk.EndorHybridHandlerInterface); ok {
					definition, err := entityHybrid.UnmarshalAdditionalAttributes()
					if err != nil {
						c.logger.Warn(fmt.Sprintf("unmarshal error for hybrid %s: %s", entityID, err.Error()))
						continue
					}
					handlerDict.EndorHandler = hybridInst.ToEndorHandler(*definition)
					if orig, ok := handlerDict.entity.(*sdk.EntityHybrid); ok {
						cloned := *orig
						cloned.AdditionalSchema = entityHybrid.AdditionalSchema
						handlerDict.entity = &cloned
					}
				}
			}

			if entitySpec, ok := devEntity.(*sdk.EntityHybridSpecialized); ok {
				if specInst, ok := (*existing.OriginalInstance).(sdk.EndorHybridSpecializedHandlerInterface); ok {
					definition, err := entitySpec.UnmarshalAdditionalAttributes()
					if err != nil {
						c.logger.Warn(fmt.Sprintf("unmarshal error for specialized %s: %s", entityID, err.Error()))
						continue
					}
					cats := []sdk.EndorHybridSpecializedHandlerCategoryInterface{}
					catsSchema := map[string]sdk.RootSchema{}
					for _, cat := range entitySpec.Categories {
						cats = append(cats, NewEndorHybridSpecializedHandlerCategory[*sdk.DynamicEntitySpecialized](cat.ID, cat.Description))
						catSchema, err := cat.UnmarshalAdditionalAttributes()
						if err != nil {
							c.logger.Warn(fmt.Sprintf("unmarshal error for category %s/%s: %s", entityID, cat.ID, err.Error()))
						} else {
							catsSchema[cat.ID] = *catSchema
						}
					}
					handlerDict.EndorHandler = specInst.WithHybridCategories(cats).ToEndorHandler(*definition, catsSchema, entitySpec.AdditionalCategories)
					if orig, ok := handlerDict.entity.(*sdk.EntityHybridSpecialized); ok {
						cloned := *orig
						cloned.AdditionalSchema = entitySpec.AdditionalSchema
						cloned.AdditionalCategories = entitySpec.AdditionalCategories
						handlerDict.entity = &cloned
					}
				}
			}
		} else {
			// New dev entity (not in prod) or prod entry has no compiled handler.
			repoSnapshot := sdk.GetRepositoryRegistry().Snapshot()
			if entityHybrid, ok := devEntity.(*sdk.EntityHybrid); ok {
				definition, err := entityHybrid.UnmarshalAdditionalAttributes()
				if err != nil {
					c.logger.Warn(fmt.Sprintf("unmarshal error for new hybrid %s: %s", entityID, err.Error()))
					continue
				}
				schema, _ := sdk.NewSchema(sdk.DynamicEntity{}).ToYAML()
				entityHybrid.Schema = schema
				hybridSvc := NewEndorHybridHandler[*sdk.DynamicEntity](entityHybrid.ID, entityHybrid.Description)
				handlerDict = EndorDIContainer{
					EndorHandler: hybridSvc.ToEndorHandler(*definition),
					entity:       entityHybrid,
					repositories: repoSnapshot,
				}
			} else if entitySpec, ok := devEntity.(*sdk.EntityHybridSpecialized); ok {
				definition, err := entitySpec.UnmarshalAdditionalAttributes()
				if err != nil {
					c.logger.Warn(fmt.Sprintf("unmarshal error for new specialized %s: %s", entityID, err.Error()))
					continue
				}
				cats := []sdk.EndorHybridSpecializedHandlerCategoryInterface{}
				catsSchema := map[string]sdk.RootSchema{}
				for _, cat := range entitySpec.Categories {
					cats = append(cats, NewEndorHybridSpecializedHandlerCategory[*sdk.DynamicEntitySpecialized](cat.ID, cat.Description))
					catSchema, err := cat.UnmarshalAdditionalAttributes()
					if err != nil {
						c.logger.Warn(fmt.Sprintf("unmarshal error for category %s/%s: %s", entityID, cat.ID, err.Error()))
					} else {
						catsSchema[cat.ID] = *catSchema
					}
				}
				schema, _ := sdk.NewSchema(sdk.DynamicEntitySpecialized{}).ToYAML()
				entitySpec.Schema = schema
				hybridSvc := NewEndorHybridSpecializedHandler[*sdk.DynamicEntitySpecialized](entitySpec.ID, entitySpec.Description).WithHybridCategories(cats)
				handlerDict = EndorDIContainer{
					EndorHandler: hybridSvc.ToEndorHandler(*definition, catsSchema, entitySpec.AdditionalCategories),
					entity:       entitySpec,
					repositories: repoSnapshot,
				}
			} else {
				continue
			}
		}

		devDict[entityID] = handlerDict
	}

	return devDict, nil
}

// #endregion
