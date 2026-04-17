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

// idToString converts an ID of any type (string, ObjectID) to a string
func idToString(id any) string {
	if id == nil {
		return ""
	}
	switch v := id.(type) {
	case string:
		return v
	case sdk.ObjectID:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// entityDSLBasePath returns the production filesystem path where entity DSL YAML files
// for the given microservice are stored, relative to the service's working directory.
// Structure: ./prod/ui/entities/<ms-id>/
func entityDSLBasePath(msId string) string {
	return filepath.Join("prod", "entities", msId)
}

// devEntityDSLBasePath returns the per-user debug overlay path, relative to the
// service's working directory.
// Structure: ./dev/<user-id>/ui/entities/<ms-id>/
func devEntityDSLBasePath(userID, msId string) string {
	return filepath.Join("dev", userID, "entities", msId)
}

// Singleton instance and initialization sync
var (
	endorServiceRepositoryInstance *EndorHandlerRepository
	endorServiceRepositoryOnce     sync.Once
)

// GetEndorHandlerRepository returns the singleton instance of EndorHandlerRepository.
// It must be initialized first by calling InitEndorHandlerRepository.
func GetEndorHandlerRepository() *EndorHandlerRepository {
	return endorServiceRepositoryInstance
}

// InitEndorHandlerRepository initializes the singleton EndorHandlerRepository instance.
// This should be called once during application startup.
// Subsequent calls will return the existing instance without reinitializing.
//
// The repository always scans ./prod/ui/entities/<microServiceId>/ for production DSL
// entity definitions and installs an fsnotify watcher on ./prod/ with a 1-second debounce.
func InitEndorHandlerRepository(microServiceId string, internalEndorHandlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger) *EndorHandlerRepository {
	endorServiceRepositoryOnce.Do(func() {
		relDSLPath := entityDSLBasePath(microServiceId)
		absDSLPath, _ := filepath.Abs(relDSLPath)
		// prodDSLRoot is 3 levels above entityDSLPath: prod/ui/entities/<ms> → prod
		absProdRoot, _ := filepath.Abs(filepath.Join(relDSLPath, "..", "..", ".."))
		endorServiceRepositoryInstance = &EndorHandlerRepository{
			microServiceId:        microServiceId,
			internalEndorHandlers: internalEndorHandlers,
			logger:                logger,
			mu:                    &sync.RWMutex{},
			entityDSLPath:         absDSLPath,
			prodDSLRoot:           absProdRoot,
			ephemeralCache:        newEphemeralCacheManager(),
		}
		if err := endorServiceRepositoryInstance.startEntityWatcher(); err != nil {
			logger.Warn(fmt.Sprintf("unable to start entity DSL watcher: %s", err.Error()))
		}
	})
	return endorServiceRepositoryInstance
}

// NewEndorHandlerRepository returns the singleton instance of EndorHandlerRepository.
// If the singleton hasn't been initialized yet, it initializes it with the provided parameters.
// Deprecated: Use InitEndorHandlerRepository for explicit initialization or GetEndorHandlerRepository to get the instance.
func NewEndorHandlerRepository(microServiceId string, internalEndorHandlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger) *EndorHandlerRepository {
	return InitEndorHandlerRepository(microServiceId, internalEndorHandlers, logger)
}

type EndorHandlerRepository struct {
	microServiceId        string
	internalEndorHandlers *[]sdk.EndorHandlerInterface
	// entityDSLPath is the absolute path to ./prod/ui/entities/<ms-id>/
	entityDSLPath string
	// prodDSLRoot is the absolute path to ./prod/ – the root watched by fsnotify.
	prodDSLRoot      string
	logger           *sdk.Logger
	mu               *sync.RWMutex
	cachedDictionary map[string]EndorHandlerDictionary
	cacheInitialized bool
	// ephemeralCache holds per-user ephemeral (debug) overlay registries.
	ephemeralCache *EphemeralCacheManager
}

type EndorHandlerDictionary struct {
	OriginalInstance *sdk.EndorHandlerInterface
	EndorHandler     sdk.EndorHandler
	entity           sdk.EntityInterface
}

type EndorHandlerActionDictionary struct {
	EndorHandlerAction sdk.EndorHandlerActionInterface
	entityAction       sdk.EntityAction
}

// #region Framework CRUD

func (h *EndorHandlerRepository) DictionaryMap() (map[string]EndorHandlerDictionary, error) {
	// Check cache first with read lock
	h.mu.RLock()
	if h.cacheInitialized {
		// Return a copy of the cached dictionary to prevent external modifications
		result := make(map[string]EndorHandlerDictionary, len(h.cachedDictionary))
		for k, v := range h.cachedDictionary {
			result[k] = v
		}
		h.mu.RUnlock()
		return result, nil
	}
	h.mu.RUnlock()

	// Acquire write lock to build the dictionary
	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check after acquiring write lock
	if h.cacheInitialized {
		// Return a copy of the cached dictionary
		result := make(map[string]EndorHandlerDictionary, len(h.cachedDictionary))
		for k, v := range h.cachedDictionary {
			result[k] = v
		}
		return result, nil
	}

	entities := map[string]EndorHandlerDictionary{}

	// internal EndorHandlers
	if h.internalEndorHandlers != nil {
		for _, internalEndorHandler := range *h.internalEndorHandlers {
			schema, err := internalEndorHandler.GetSchema().ToYAML()
			if err != nil {
				h.logger.Warn(fmt.Sprintf("unable to read entity schema from %s", internalEndorHandler.GetEntity()))
			}
			baseEntity := sdk.Entity{
				ID:          internalEndorHandler.GetEntity(),
				Description: internalEndorHandler.GetEntityDescription(),
				Service:     h.microServiceId,
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
							h.logger.Warn(fmt.Sprintf("unable to create entity %s from service", internalEndorHandler.GetEntity()))
						}
					}
				}
			}

			entities[internalEndorHandler.GetEntity()] = EndorHandlerDictionary{
				OriginalInstance: &internalEndorHandler,
				EndorHandler:     endorService,
				entity:           entity,
			}
		}
	}

	// dynamic EndorHandlers
	if sdk_configuration.GetConfig().HybridEntitiesEnabled || sdk_configuration.GetConfig().DynamicEntitiesEnabled {
		dynamicEntities, err := h.DynamicEntityList()
		if err != nil {
			return map[string]EndorHandlerDictionary{}, nil
		}

		for _, entity := range dynamicEntities {
			// check if service is already defined
			// search service
			entityID := idToString(entity.GetID())
			if v, ok := entities[entityID]; ok {
				// check entity hybrid
				if entityHybrid, ok := entity.(*sdk.EntityHybrid); ok {
					if hybridInstance, ok := (*v.OriginalInstance).(sdk.EndorHybridHandlerInterface); ok {
						defintion, err := entityHybrid.UnmarshalAdditionalAttributes()
						if err != nil {
							h.logger.Warn(fmt.Sprintf("unable to unmarshal definition for hybrid entity %s: %s", entityHybrid.ID, err.Error()))
						}
						// inject dynamic schema
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
							h.logger.Warn(fmt.Sprintf("unable to unmarshal definition for hybrid specialized entity %s: %s", entitySpecialized.ID, err.Error()))
						}
						// inject categories and schema
						categories := []sdk.EndorHybridSpecializedHandlerCategoryInterface{}
						categoriesAdditionalSchema := map[string]sdk.RootSchema{}
						for _, c := range entitySpecialized.Categories {
							categories = append(categories, NewEndorHybridSpecializedHandlerCategory[*sdk.DynamicEntitySpecialized](c.ID, c.Description))
							categoryAdditionalSchema, err := c.UnmarshalAdditionalAttributes()
							if err != nil {
								h.logger.Warn(fmt.Sprintf("unable to unmarshal category definition %s for hybrid entity %s: %s", c.ID, entitySpecialized.ID, err.Error()))
							}
							categoriesAdditionalSchema[c.ID] = *categoryAdditionalSchema
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
						h.logger.Warn(fmt.Sprintf("unable to unmarshal definition for dynamic entity %s: %s", entityHybrid.ID, err.Error()))
					}
					// create a new hybrid service
					hybridService := NewEndorHybridHandler[*sdk.DynamicEntity](entityHybrid.ID, entityHybrid.Description)
					schema, _ := sdk.NewSchema(sdk.DynamicEntity{}).ToYAML()
					entityHybrid.Schema = schema
					entities[entityHybrid.ID] = EndorHandlerDictionary{
						EndorHandler: hybridService.ToEndorHandler(*defintion),
						entity:       entityHybrid,
					}
				}
				if entitySpecialized, ok := entity.(*sdk.EntityHybridSpecialized); ok {
					defintion, err := entitySpecialized.UnmarshalAdditionalAttributes()
					if err != nil {
						h.logger.Warn(fmt.Sprintf("unable to unmarshal definition for dynamic specialized entity %s: %s", entitySpecialized.ID, err.Error()))
					}
					// create categories
					categories := []sdk.EndorHybridSpecializedHandlerCategoryInterface{}
					categoriesAdditionalSchema := map[string]sdk.RootSchema{}
					for _, c := range entitySpecialized.Categories {
						categories = append(categories, NewEndorHybridSpecializedHandlerCategory[*sdk.DynamicEntitySpecialized](c.ID, c.Description))
						categoryAdditionalSchema, err := c.UnmarshalAdditionalAttributes()
						if err != nil {
							h.logger.Warn(fmt.Sprintf("unable to unmarshal category definition %s for dynamic specialized entity %s: %s", c.ID, entitySpecialized.ID, err.Error()))
						}
						categoriesAdditionalSchema[c.ID] = *categoryAdditionalSchema
					}
					// create a new specilized service
					hybridService := NewEndorHybridSpecializedHandler[*sdk.DynamicEntitySpecialized](entitySpecialized.ID, entitySpecialized.Description).WithHybridCategories(categories)
					schema, _ := sdk.NewSchema(sdk.DynamicEntitySpecialized{}).ToYAML()
					entitySpecialized.Schema = schema
					entities[entitySpecialized.ID] = EndorHandlerDictionary{
						EndorHandler: hybridService.ToEndorHandler(*defintion, categoriesAdditionalSchema, entitySpecialized.AdditionalCategories),
						entity:       entitySpecialized,
					}
				}
			}
		}
	}

	// Cache the result
	h.cachedDictionary = entities
	h.cacheInitialized = true

	return entities, nil
}

func (h *EndorHandlerRepository) DictionaryActionMap() (map[string]EndorHandlerActionDictionary, error) {
	actions := map[string]EndorHandlerActionDictionary{}
	entities, err := h.DictionaryMap()
	if err != nil {
		return actions, err
	}
	for entityName, entity := range entities {
		for actionName, EndorHandlerAction := range entity.EndorHandler.Actions {
			action, err := h.createAction(entityName, entity.EndorHandler.Version, actionName, EndorHandlerAction)
			if err == nil {
				actions[action.entityAction.ID] = *action
			}
		}
	}
	return actions, nil
}

func (h *EndorHandlerRepository) DictionaryInstance(dto sdk.ReadInstanceDTO) (*EndorHandlerDictionary, error) {
	// get all service
	entities, err := h.DictionaryMap()
	if err != nil {
		return nil, err
	}
	if entity, ok := entities[dto.Id]; ok {
		return &entity, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id)).WithTranslation("entities.entity.not_found", map[string]any{"id": dto.Id})
}

func (h *EndorHandlerRepository) DictionaryActionInstance(dto sdk.ReadInstanceDTO) (*EndorHandlerActionDictionary, error) {
	actions, err := h.DictionaryActionMap()
	if err != nil {
		return nil, err
	}
	if action, ok := actions[dto.Id]; ok {
		return &action, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity action not found")).WithTranslation("entities.entity.action_not_found", nil)
}

func (h *EndorHandlerRepository) EntityActionList() ([]sdk.EntityAction, error) {
	actions, err := h.DictionaryActionMap()
	if err != nil {
		return []sdk.EntityAction{}, err
	}
	actionList := make([]sdk.EntityAction, 0, len(actions))
	for _, action := range actions {
		actionList = append(actionList, action.entityAction)
	}
	return actionList, nil
}

func (h *EndorHandlerRepository) EndorHandlerList() ([]sdk.EndorHandler, error) {
	entities, err := h.DictionaryMap()
	if err != nil {
		return []sdk.EndorHandler{}, err
	}
	entityList := make([]sdk.EndorHandler, 0, len(entities))
	for _, service := range entities {
		entityList = append(entityList, service.EndorHandler)
	}
	return entityList, nil
}

// entityDSLFile is the YAML structure used to deserialize entity definition files from the DSL.
type entityDSLFile struct {
	Description          string                `yaml:"description"`
	Type                 string                `yaml:"type"`
	AdditionalSchema     string                `yaml:"additionalSchema"`
	Categories           []sdk.HybridCategory  `yaml:"categories"`
	AdditionalCategories []sdk.DynamicCategory `yaml:"additionalCategories"`
}

// DynamicEntityList reads entity definitions from the DSL filesystem path and returns them
// as EntityInterface instances. Each file in the entity DSL path represents one entity;
// the filename (without extension) is used as the entity ID.
func (h *EndorHandlerRepository) DynamicEntityList() ([]sdk.EntityInterface, error) {
	if h.entityDSLPath == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(h.entityDSLPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entities []sdk.EntityInterface

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		entityID := strings.TrimSuffix(name, filepath.Ext(name))

		data, err := os.ReadFile(filepath.Join(h.entityDSLPath, name))
		if err != nil {
			h.logger.Warn(fmt.Sprintf("unable to read entity DSL file %s: %s", name, err.Error()))
			continue
		}

		var def entityDSLFile
		if err := yaml.Unmarshal(data, &def); err != nil {
			h.logger.Warn(fmt.Sprintf("unable to parse entity DSL file %s: %s", name, err.Error()))
			continue
		}

		baseEntity := sdk.Entity{
			ID:          entityID,
			Description: def.Description,
			Type:        def.Type,
			Service:     h.microServiceId,
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
			h.logger.Warn(fmt.Sprintf("unknown entity type %q in DSL file %s", def.Type, name))
		}
	}

	return entities, nil
}

// #endregion

// #region Entity CRUD

func (h *EndorHandlerRepository) List(entityType *sdk.EntityType) ([]sdk.EntityInterface, error) {
	entities, err := h.DictionaryMap()
	if err != nil {
		return []sdk.EntityInterface{}, err
	}
	entityList := make([]sdk.EntityInterface, 0, len(entities))
	for _, service := range entities {
		entityList = append(entityList, service.entity)
	}
	// filter by entity type
	filtered := make([]sdk.EntityInterface, 0, len(entities))
	for _, r := range entityList {
		if r.GetID() != "entity" && r.GetID() != "entity-action" {
			if r.GetCategoryType() == string(*entityType) {
				filtered = append(filtered, r)
			} else {
				if entityType == nil || *entityType == "" {
					filtered = append(filtered, r)
				}
			}
		}
	}
	return filtered, nil
}

func (h *EndorHandlerRepository) Instance(entityType *sdk.EntityType, dto sdk.ReadInstanceDTO) (*sdk.EntityInterface, error) {
	entity, err := h.DictionaryInstance(dto)
	if err != nil {
		return nil, err
	}
	if entityType == nil || *entityType == "" {
		return &entity.entity, nil
	}
	if entity.entity.GetCategoryType() == string(*entityType) {
		return &entity.entity, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id)).WithTranslation("entities.entity.not_found", map[string]any{"id": dto.Id})
}

// #endregion

// #region Utility

// Sync invalidates the production registry cache, forcing a full rebuild from the
// DSL filesystem (./prod/ui/entities/<ms-id>/) on the next access. It also invalidates
// all per-user ephemeral (debug) registries, since they are overlays of the production data.
func (h *EndorHandlerRepository) Sync() {
	h.mu.Lock()
	h.cacheInitialized = false
	h.cachedDictionary = nil
	h.mu.Unlock()
	h.ephemeralCache.InvalidateAll()
}

func (h *EndorHandlerRepository) reloadRouteConfiguration(microserviceId string) error {
	config := sdk_configuration.GetConfig()
	entities, err := h.EndorHandlerList()
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

func (h *EndorHandlerRepository) createAction(entityName string, version string, actionName string, endorServiceAction sdk.EndorHandlerActionInterface) (*EndorHandlerActionDictionary, error) {
	if version == "" {
		version = "v1"
	}
	actionId := path.Join(h.microServiceId, version, entityName, actionName)
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

// startEntityWatcher installs an fsnotify watcher on ./prod/ (h.prodDSLRoot) and calls
// Sync() + reloadRouteConfiguration() with a 1-second debounce on any file-system change
// within that tree. It tolerates an absent prod directory: it walks up the path until an
// existing ancestor is found, watches that, and progressively adds newly-created
// subdirectories as they appear.
//
// Gateway & Swagger configuration are always regenerated from the production (MainRegistry)
// data – debug overlay directories (./dev/) are never watched here.
func (h *EndorHandlerRepository) startEntityWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Start watching from prodDSLRoot; walk up to find the deepest existing ancestor.
	watchRoot := h.prodDSLRoot
	for {
		if _, err := os.Stat(watchRoot); err == nil {
			break
		}
		parent := filepath.Dir(watchRoot)
		if parent == watchRoot {
			watcher.Close()
			return fmt.Errorf("no existing ancestor directory found for %s", h.prodDSLRoot)
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
					if isAncestorOrEqual(h.prodDSLRoot, event.Name) {
						if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
							_ = watcher.Add(event.Name)
						}
					}
				}

				// Only react to events inside the prod tree.
				if !isAncestorOrEqual(h.prodDSLRoot, event.Name) {
					continue
				}

				// Debounce: collapse rapid bursts of changes into a single sync call.
				capturedName := event.Name
				debounceMu.Lock()
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(1*time.Second, func() {
					h.logger.Info(fmt.Sprintf("prod DSL change detected (%s), syncing registry", capturedName))
					h.Sync()
					if reloadErr := h.reloadRouteConfiguration(h.microServiceId); reloadErr != nil {
						h.logger.Warn(fmt.Sprintf("unable to reload route configuration after DSL change: %s", reloadErr.Error()))
					}
				})
				debounceMu.Unlock()

			case watchErr, ok := <-watcher.Errors:
				if !ok {
					return
				}
				h.logger.Warn(fmt.Sprintf("entity DSL watcher error: %s", watchErr.Error()))
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

// #region Debug / Ephemeral Registry

// Resolve returns the entity dictionary appropriate for the current request.
//
//   - Normal request            → production MainRegistry (DictionaryMap)
//   - session.Development=true → per-user EphemeralRegistry (shallow clone of prod + dev overlay)
//
// If the ephemeral registry for the requesting user is cached and not yet expired it is
// returned immediately; otherwise it is built from ./dev/<userID>/ui/entities/<ms-id>/
// and stored in the cache with a 30-second TTL.
//
// On any error building the ephemeral registry the method falls back to the production
// dictionary and logs a warning.
func (h *EndorHandlerRepository) Resolve(session sdk.Session) (map[string]EndorHandlerDictionary, error) {
	if !session.Development {
		return h.DictionaryMap()
	}
	userID := session.Username
	if userID == "" {
		// Debug mode without a user ID is not allowed; fall back to production.
		return h.DictionaryMap()
	}

	// Fast path: cached ephemeral registry.
	if cached := h.ephemeralCache.Get(userID); cached != nil {
		return cached, nil
	}

	// Slow path: build the overlay and populate the cache.
	devDict, err := h.buildDevDictionary(userID)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[debug][user=%s] failed to build ephemeral registry, falling back to prod: %s", userID, err.Error()))
		return h.DictionaryMap()
	}
	h.ephemeralCache.Set(userID, devDict)
	return devDict, nil
}

// DictionaryActionInstanceForSession resolves the entity dictionary for the given session
// (main or ephemeral) and looks up the action by its composite ID
// (ms-id/version/entity/action[/category]).
func (h *EndorHandlerRepository) DictionaryActionInstanceForSession(session sdk.Session, dto sdk.ReadInstanceDTO) (*EndorHandlerActionDictionary, error) {
	dict, err := h.Resolve(session)
	if err != nil {
		return nil, err
	}

	// Rebuild the action index from the resolved dictionary.
	actions := make(map[string]EndorHandlerActionDictionary)
	for entityName, entity := range dict {
		for actionName, endorHandlerAction := range entity.EndorHandler.Actions {
			action, err := h.createAction(entityName, entity.EndorHandler.Version, actionName, endorHandlerAction)
			if err == nil {
				actions[action.entityAction.ID] = *action
			}
		}
	}

	if action, ok := actions[dto.Id]; ok {
		return &action, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity action not found")).WithTranslation("entities.entity.action_not_found", nil)
}

// buildDevDictionary constructs a debug overlay dictionary for the given user.
//
// Algorithm:
//  1. Shallow-clone the production MainRegistry (each EndorHandlerDictionary is copied by
//     value, so overwriting a key in the clone never mutates the production dict).
//  2. Scan ./dev/<userID>/ui/entities/<ms-id>/ for YAML files.
//  3. For each valid file, override (or add) the corresponding entity in the clone:
//     - If a hardcoded hybrid handler exists in prod → re-run the hybrid merge with the
//     dev DSL so the compiled handler logic is retained but the schema comes from dev.
//     - Otherwise (purely dynamic entity or brand-new dev entity) → create a fresh
//     dynamic handler from the dev DSL.
//  4. Corrupt/unreadable dev YAML files are skipped with a warning; the production entry
//     (if any) is preserved, providing automatic fallback.
//
// Security: db isolation is provided by the sdk.DebugDBPrefix() helper; repository
// implementations should call it when constructing their database name.
func (h *EndorHandlerRepository) buildDevDictionary(userID string) (map[string]EndorHandlerDictionary, error) {
	// Build (and cache) the production dict first.
	mainDict, err := h.DictionaryMap()
	if err != nil {
		return nil, err
	}

	// Shallow clone: struct-value copy, no production map is mutated.
	devDict := make(map[string]EndorHandlerDictionary, len(mainDict))
	for k, v := range mainDict {
		devDict[k] = v
	}

	devPath := devEntityDSLBasePath(userID, h.microServiceId)
	entries, err := os.ReadDir(devPath)
	if err != nil {
		if os.IsNotExist(err) {
			return devDict, nil
		}
		// Non-fatal I/O error: log and return the production clone as-is.
		h.logger.Warn(fmt.Sprintf("[debug][user=%s] unable to read dev entity dir %s: %s", userID, devPath, err.Error()))
		return devDict, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		entityID := strings.TrimSuffix(name, filepath.Ext(name))

		data, err := os.ReadFile(filepath.Join(devPath, name))
		if err != nil {
			h.logger.Warn(fmt.Sprintf("[debug][user=%s] unable to read dev entity DSL %s: %s", userID, name, err.Error()))
			continue
		}

		var def entityDSLFile
		if err := yaml.Unmarshal(data, &def); err != nil {
			h.logger.Warn(fmt.Sprintf("[debug][user=%s] corrupt dev entity DSL %s, keeping prod version: %s", userID, name, err.Error()))
			continue // resilient fallback: production entry (if any) remains
		}

		baseEntity := sdk.Entity{
			ID:          entityID,
			Description: def.Description,
			Type:        def.Type,
			Service:     h.microServiceId,
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
			h.logger.Warn(fmt.Sprintf("[debug][user=%s] unknown entity type %q in dev DSL %s", userID, def.Type, name))
			continue
		}

		existing, existsInProd := devDict[entityID]

		var handlerDict EndorHandlerDictionary

		if existsInProd && existing.OriginalInstance != nil {
			// Override: re-merge dev DSL with the original compiled handler.
			handlerDict = existing

			if entityHybrid, ok := devEntity.(*sdk.EntityHybrid); ok {
				if hybridInst, ok := (*existing.OriginalInstance).(sdk.EndorHybridHandlerInterface); ok {
					definition, err := entityHybrid.UnmarshalAdditionalAttributes()
					if err != nil {
						h.logger.Warn(fmt.Sprintf("[debug][user=%s] unmarshal error for hybrid %s: %s", userID, entityID, err.Error()))
						continue
					}
					handlerDict.EndorHandler = hybridInst.ToEndorHandler(*definition)
					// Clone the entity to avoid mutating the production object.
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
						h.logger.Warn(fmt.Sprintf("[debug][user=%s] unmarshal error for specialized %s: %s", userID, entityID, err.Error()))
						continue
					}
					cats := []sdk.EndorHybridSpecializedHandlerCategoryInterface{}
					catsSchema := map[string]sdk.RootSchema{}
					for _, c := range entitySpec.Categories {
						cats = append(cats, NewEndorHybridSpecializedHandlerCategory[*sdk.DynamicEntitySpecialized](c.ID, c.Description))
						catSchema, err := c.UnmarshalAdditionalAttributes()
						if err != nil {
							h.logger.Warn(fmt.Sprintf("[debug][user=%s] unmarshal error for category %s/%s: %s", userID, entityID, c.ID, err.Error()))
						} else {
							catsSchema[c.ID] = *catSchema
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
			// New dev entity (not in prod) or prod entry has no compiled handler:
			// create a fresh dynamic handler from the dev DSL.
			if entityHybrid, ok := devEntity.(*sdk.EntityHybrid); ok {
				definition, err := entityHybrid.UnmarshalAdditionalAttributes()
				if err != nil {
					h.logger.Warn(fmt.Sprintf("[debug][user=%s] unmarshal error for new hybrid %s: %s", userID, entityID, err.Error()))
					continue
				}
				schema, _ := sdk.NewSchema(sdk.DynamicEntity{}).ToYAML()
				entityHybrid.Schema = schema
				hybridSvc := NewEndorHybridHandler[*sdk.DynamicEntity](entityHybrid.ID, entityHybrid.Description)
				handlerDict = EndorHandlerDictionary{
					EndorHandler: hybridSvc.ToEndorHandler(*definition),
					entity:       entityHybrid,
				}
			} else if entitySpec, ok := devEntity.(*sdk.EntityHybridSpecialized); ok {
				definition, err := entitySpec.UnmarshalAdditionalAttributes()
				if err != nil {
					h.logger.Warn(fmt.Sprintf("[debug][user=%s] unmarshal error for new specialized %s: %s", userID, entityID, err.Error()))
					continue
				}
				cats := []sdk.EndorHybridSpecializedHandlerCategoryInterface{}
				catsSchema := map[string]sdk.RootSchema{}
				for _, c := range entitySpec.Categories {
					cats = append(cats, NewEndorHybridSpecializedHandlerCategory[*sdk.DynamicEntitySpecialized](c.ID, c.Description))
					catSchema, err := c.UnmarshalAdditionalAttributes()
					if err != nil {
						h.logger.Warn(fmt.Sprintf("[debug][user=%s] unmarshal error for category %s/%s: %s", userID, entityID, c.ID, err.Error()))
					} else {
						catsSchema[c.ID] = *catSchema
					}
				}
				schema, _ := sdk.NewSchema(sdk.DynamicEntitySpecialized{}).ToYAML()
				entitySpec.Schema = schema
				hybridSvc := NewEndorHybridSpecializedHandler[*sdk.DynamicEntitySpecialized](entitySpec.ID, entitySpec.Description).WithHybridCategories(cats)
				handlerDict = EndorHandlerDictionary{
					EndorHandler: hybridSvc.ToEndorHandler(*definition, catsSchema, entitySpec.AdditionalCategories),
					entity:       entitySpec,
				}
			} else {
				continue
			}
		}

		devDict[entityID] = handlerDict
	}

	return devDict, nil
}

// #endregion Debug / Ephemeral Registry

// #endregion
