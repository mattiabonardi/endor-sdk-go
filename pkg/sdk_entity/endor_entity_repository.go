package sdk_entity

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

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

// entityDSLBasePath returns the filesystem path where entity DSL files for the given microservice are stored.
// Structure: <dsl-root>/entities/<ms-id>/
func entityDSLBasePath(msId string, development bool) string {
	homeDir, _ := os.UserHomeDir()
	var base string
	if development {
		base = filepath.Join(homeDir, "etc", "endor", "dsl", "dev")
	} else {
		base = filepath.Join(homeDir, "etc", "endor", "dsl", "prod")
	}
	return filepath.Join(base, "entities", msId)
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
func InitEndorHandlerRepository(microServiceId string, internalEndorHandlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger) *EndorHandlerRepository {
	endorServiceRepositoryOnce.Do(func() {
		config := sdk_configuration.GetConfig()
		endorServiceRepositoryInstance = &EndorHandlerRepository{
			microServiceId:        microServiceId,
			internalEndorHandlers: internalEndorHandlers,
			logger:                logger,
			mu:                    &sync.RWMutex{},
		}
		if config.HybridEntitiesEnabled || config.DynamicEntitiesEnabled {
			endorServiceRepositoryInstance.entityDSLPath = entityDSLBasePath(microServiceId, config.Development)
			if err := endorServiceRepositoryInstance.startEntityWatcher(); err != nil {
				logger.Warn(fmt.Sprintf("unable to start entity DSL watcher: %s", err.Error()))
			}
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
	entityDSLPath         string
	logger                *sdk.Logger
	mu                    *sync.RWMutex
	cachedDictionary      map[string]EndorHandlerDictionary
	cacheInitialized      bool
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
			action, err := h.createAction(entityName, actionName, EndorHandlerAction)
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
	idSegments := strings.Split(dto.Id, "/")
	if len(idSegments) == 2 {
		entityInstance, err := h.DictionaryInstance(sdk.ReadInstanceDTO{
			Id: idSegments[0],
		})
		if err != nil {
			return nil, err
		}
		if entityAction, ok := entityInstance.EndorHandler.Actions[idSegments[1]]; ok {
			return h.createAction(idSegments[0], idSegments[1], entityAction)
		} else {
			return nil, sdk.NewNotFoundError(fmt.Errorf("entity action not found")).WithTranslation("entities.entity.action_not_found", nil)
		}
	} else {
		return nil, sdk.NewBadRequestError(fmt.Errorf("invalid entity action id")).WithTranslation("entities.entity.invalid_action_id", nil)
	}
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

// invalidateCache clears the cached dictionary so it will be rebuilt on next access
func (h *EndorHandlerRepository) invalidateCache() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cacheInitialized = false
	h.cachedDictionary = nil
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

func (h *EndorHandlerRepository) createAction(entityName string, actionName string, endorServiceAction sdk.EndorHandlerActionInterface) (*EndorHandlerActionDictionary, error) {
	actionId := path.Join(entityName, actionName)
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

// startEntityWatcher watches the entity DSL directory for file changes.
// It tolerates a completely absent DSL tree: it walks up the path until it finds
// an existing ancestor, watches that, and progressively descends into newly-created
// subdirectories until entityDSLPath itself is watched.
func (h *EndorHandlerRepository) startEntityWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Find the deepest existing ancestor of entityDSLPath.
	watchRoot := h.entityDSLPath
	for {
		if _, err := os.Stat(watchRoot); err == nil {
			break
		}
		parent := filepath.Dir(watchRoot)
		if parent == watchRoot {
			// Reached filesystem root without finding an existing dir; nothing to watch.
			watcher.Close()
			return fmt.Errorf("no existing ancestor directory found for %s", h.entityDSLPath)
		}
		watchRoot = parent
	}

	if err := watcher.Add(watchRoot); err != nil {
		watcher.Close()
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) {
					// If a new directory was created that is an ancestor of (or equal to)
					// entityDSLPath, start watching it so we can keep descending.
					if isAncestorOrEqual(event.Name, h.entityDSLPath) {
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							_ = watcher.Add(event.Name)
						}
					}
				}
				// React to file changes directly inside entityDSLPath.
				if filepath.Dir(event.Name) == h.entityDSLPath {
					if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) ||
						event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
						h.logger.Info(fmt.Sprintf("entity DSL change detected (%s), reloading", event.Name))
						h.invalidateCache()
						if err := h.reloadRouteConfiguration(h.microServiceId); err != nil {
							h.logger.Warn(fmt.Sprintf("unable to reload route configuration after DSL change: %s", err.Error()))
						}
					}
				}
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

// #endregion
