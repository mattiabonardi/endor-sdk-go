package sdk_entity

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/mattiabonardi/endor-sdk-go/internal/api_gateway"
	"github.com/mattiabonardi/endor-sdk-go/internal/swagger"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// idToString converts an ID of any type (string, ObjectID) to a string
// This helper is used to handle GetID() which now returns any
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

// COLLECTION_ENTITIES is the MongoDB collection name for entities
const COLLECTION_ENTITIES = "entities"

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
		endorServiceRepositoryInstance = &EndorHandlerRepository{
			microServiceId:        microServiceId,
			internalEndorHandlers: internalEndorHandlers,
			context:               context.TODO(),
			logger:                logger,
			mu:                    &sync.RWMutex{},
		}
		if sdk_configuration.GetConfig().HybridEntitiesEnabled || sdk_configuration.GetConfig().DynamicEntitiesEnabled {
			client, err := sdk.GetMongoClient()
			if client != nil && err == nil {
				database := client.Database(sdk_configuration.GetConfig().DynamicEntityDocumentDBName)
				endorServiceRepositoryInstance.collection = database.Collection(COLLECTION_ENTITIES)
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
	collection            *mongo.Collection
	context               context.Context
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
				h.ensureEntityDocumentOfInternalService(&hybridSpecializedEntity)
				entity = &hybridSpecializedEntity
				endorService = hybridSpecializedService.ToEndorHandler(sdk.RootSchema{}, map[string]sdk.RootSchema{}, []sdk.DynamicCategory{})
			} else {
				// hybrid
				if hybridService, ok := internalEndorHandler.(sdk.EndorHybridHandlerInterface); ok {
					baseEntity.Type = string(sdk.EntityTypeHybrid)
					h.ensureEntityDocumentOfInternalService(&baseEntity)
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
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id))
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
			return nil, sdk.NewNotFoundError(fmt.Errorf("entity action not found"))
		}
	} else {
		return nil, sdk.NewBadRequestError(fmt.Errorf("invalid entity action id"))
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

func (h *EndorHandlerRepository) DynamicEntityList() ([]sdk.EntityInterface, error) {
	cursor, err := h.collection.Find(h.context, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(h.context)

	var entities []sdk.EntityInterface

	for cursor.Next(h.context) {
		raw := cursor.Current

		// First decode to Entity base
		var discr sdk.Entity
		if err := bson.Unmarshal(raw, &discr); err != nil {
			h.logger.Warn(fmt.Sprintf("unable to decode generic entity details from datasource: %s", err.Error()))
		}

		switch discr.GetCategoryType() {
		case string(sdk.EntityTypeHybrid):
			var r sdk.EntityHybrid
			if err := bson.Unmarshal(raw, &r); err != nil {
				h.logger.Warn(fmt.Sprintf("unable to decode hybrid entity %s details from datasource: %s", discr.ID, err.Error()))
			} else {
				entities = append(entities, &r)
			}

		case string(sdk.EntityTypeHybridSpecialized):
			var r sdk.EntityHybridSpecialized
			if err := bson.Unmarshal(raw, &r); err != nil {
				h.logger.Warn(fmt.Sprintf("unable to decode hybrid specialized entity %s details from datasource: %s", discr.ID, err.Error()))
			} else {
				entities = append(entities, &r)
			}

		case string(sdk.EntityTypeDynamic):
			var r sdk.EntityHybrid
			if err := bson.Unmarshal(raw, &r); err != nil {
				h.logger.Warn(fmt.Sprintf("unable to decode dynamic entity %s details from datasource: %s", discr.ID, err.Error()))
			} else {
				entities = append(entities, &r)
			}

		case string(sdk.EntityTypeDynamicSpecialized):
			var r sdk.EntityHybridSpecialized
			if err := bson.Unmarshal(raw, &r); err != nil {
				h.logger.Warn(fmt.Sprintf("unable to decode dynamic specialized entity %s details from datasource: %s", discr.ID, err.Error()))
			} else {
				entities = append(entities, &r)
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
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
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id))
}

func (h *EndorHandlerRepository) Create(entityType *sdk.EntityType, dto sdk.CreateDTO[sdk.EntityInterface]) (*sdk.EntityInterface, error) {
	if *entityType == sdk.EntityTypeDynamic || *entityType == sdk.EntityTypeDynamicSpecialized {
		dto.Data.SetCategoryType(string(*entityType))
		dto.Data.SetService(h.microServiceId)
		_, err := h.DictionaryInstance(sdk.ReadInstanceDTO{
			Id: idToString(dto.Data.GetID()),
		})
		var endorError *sdk.EndorError
		if errors.As(err, &endorError) && endorError.StatusCode == 404 {
			_, err := h.collection.InsertOne(h.context, dto.Data)
			if err != nil {
				return nil, err
			}
			h.invalidateCache()
			h.reloadRouteConfiguration(h.microServiceId)
			return &dto.Data, nil
		} else {
			return nil, sdk.NewConflictError(fmt.Errorf("entity already exist"))
		}
	} else {
		return nil, sdk.NewForbiddenError(fmt.Errorf("create entity not permitted"))
	}
}

func (h *EndorHandlerRepository) Update(entityType *sdk.EntityType, dto sdk.UpdateByIdDTO[map[string]interface{}]) (*sdk.EntityInterface, error) {
	if *entityType == sdk.EntityTypeDynamic || *entityType == sdk.EntityTypeDynamicSpecialized ||
		*entityType == sdk.EntityTypeHybrid || *entityType == sdk.EntityTypeHybridSpecialized {
		var instance *sdk.EntityInterface
		_, err := h.DictionaryInstance(sdk.ReadInstanceDTO{
			Id: dto.Id,
		})
		if err != nil {
			return instance, err
		}
		updateBson, err := bson.Marshal(dto.Data)
		if err != nil {
			return nil, err
		}
		update := bson.M{"$set": bson.Raw(updateBson)}
		filter := bson.M{"_id": dto.Id}
		_, err = h.collection.UpdateOne(h.context, filter, update)
		if err != nil {
			return nil, err
		}

		h.invalidateCache()
		h.reloadRouteConfiguration(h.microServiceId)

		instance, err = h.Instance(entityType, sdk.ReadInstanceDTO{
			Id: dto.Id,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update entity %s", dto.Id)
		}

		return instance, nil
	} else {
		return nil, sdk.NewForbiddenError(fmt.Errorf("update entity not permitted"))
	}
}

func (h *EndorHandlerRepository) Delete(entityType *sdk.EntityType, dto sdk.ReadInstanceDTO) error {
	if *entityType == sdk.EntityTypeDynamic || *entityType == sdk.EntityTypeDynamicSpecialized {
		// check if entities already exist
		_, err := h.DictionaryInstance(dto)
		if err != nil {
			return err
		}
		_, err = h.collection.DeleteOne(h.context, bson.M{"_id": dto.Id})
		if err == nil {
			h.invalidateCache()
			h.reloadRouteConfiguration(h.microServiceId)
		}
		return err
	} else {
		return sdk.NewForbiddenError(fmt.Errorf("delete entity not permitted"))
	}
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

func (h *EndorHandlerRepository) ensureEntityDocumentOfInternalService(entity sdk.EntityInterface) {
	if (sdk_configuration.GetConfig().HybridEntitiesEnabled || sdk_configuration.GetConfig().DynamicEntitiesEnabled) && h.collection != nil {
		// Check if document exists in MongoDB
		var existingDoc sdk.Entity
		filter := bson.M{"_id": entity.GetID()}
		err := h.collection.FindOne(h.context, filter).Decode(&existingDoc)

		// If document doesn't exist, create it
		if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
			_, insertErr := h.collection.InsertOne(h.context, entity)
			if insertErr != nil {
				h.logger.Warn(fmt.Sprintf("unable to create entity %s initilialization: %s", entity.GetID(), err.Error()))
			}
		}
	}
}

// #endregion
