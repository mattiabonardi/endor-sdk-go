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

// COLLECTION_ENTITIES is the MongoDB collection name for entities
const COLLECTION_ENTITIES = "entities"

func NewEndorServiceRepository(microServiceId string, internalEndorServices *[]sdk.EndorServiceInterface, logger *sdk.Logger) *EndorServiceRepository {
	serviceRepository := &EndorServiceRepository{
		microServiceId:        microServiceId,
		internalEndorServices: internalEndorServices,
		context:               context.TODO(),
		logger:                logger,
		mu:                    &sync.RWMutex{},
	}
	if sdk_configuration.GetConfig().HybridEntitiesEnabled || sdk_configuration.GetConfig().DynamicEntitiesEnabled {
		client, _ := sdk.GetMongoClient()
		database := client.Database(sdk_configuration.GetConfig().DynamicEntityDocumentDBName)
		serviceRepository.collection = database.Collection(COLLECTION_ENTITIES)
	}

	return serviceRepository
}

type EndorServiceRepository struct {
	microServiceId        string
	internalEndorServices *[]sdk.EndorServiceInterface
	collection            *mongo.Collection
	context               context.Context
	logger                *sdk.Logger
	mu                    *sync.RWMutex
	cachedDictionary      map[string]EndorServiceDictionary
	cacheInitialized      bool
}

type EndorServiceDictionary struct {
	OriginalInstance *sdk.EndorServiceInterface
	EndorService     sdk.EndorService
	entity           sdk.EntityInterface
}

type EndorServiceActionDictionary struct {
	EndorServiceAction sdk.EndorServiceActionInterface
	entityAction       sdk.EntityAction
}

// #region Framework CRUD

func (h *EndorServiceRepository) DictionaryMap() (map[string]EndorServiceDictionary, error) {
	// Check cache first with read lock
	h.mu.RLock()
	if h.cacheInitialized {
		// Return a copy of the cached dictionary to prevent external modifications
		result := make(map[string]EndorServiceDictionary, len(h.cachedDictionary))
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
		result := make(map[string]EndorServiceDictionary, len(h.cachedDictionary))
		for k, v := range h.cachedDictionary {
			result[k] = v
		}
		return result, nil
	}

	entities := map[string]EndorServiceDictionary{}

	// internal EndorServices
	if h.internalEndorServices != nil {
		for _, internalEndorService := range *h.internalEndorServices {
			schema, err := internalEndorService.GetSchema().ToYAML()
			if err != nil {
				h.logger.Warn(fmt.Sprintf("unable to read entity schema from %s", internalEndorService.GetEntity()))
			}
			baseEntity := sdk.Entity{
				ID:          internalEndorService.GetEntity(),
				Description: internalEndorService.GetEntityDescription(),
				Service:     h.microServiceId,
				Type:        string(sdk.EntityTypeBase),
				Schema:      schema,
			}

			var endorService sdk.EndorService
			var entity sdk.EntityInterface = &baseEntity

			// hybrid specialized
			if hybridSpecializedService, ok := internalEndorService.(sdk.EndorHybridSpecializedServiceInterface); ok {
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
				endorService = hybridSpecializedService.ToEndorService(sdk.RootSchema{}, map[string]sdk.RootSchema{}, []sdk.DynamicCategory{})
			} else {
				// hybrid
				if hybridService, ok := internalEndorService.(sdk.EndorHybridServiceInterface); ok {
					baseEntity.Type = string(sdk.EntityTypeHybrid)
					h.ensureEntityDocumentOfInternalService(&baseEntity)
					endorService = hybridService.ToEndorService(sdk.RootSchema{})
				} else {
					// base specialized
					if baseSpecializedService, ok := internalEndorService.(sdk.EndorBaseSpecializedServiceInterface); ok {
						baseEntity.Type = string(sdk.EntityTypeBaseSpecialized)
						baseSpecializedEntity := sdk.EntitySpecialized{
							Entity:     baseEntity,
							Categories: baseSpecializedService.GetCategories(),
						}
						entity = &baseSpecializedEntity
						endorService = baseSpecializedService.ToEndorService()
					} else {
						// base
						if baseService, ok := internalEndorService.(sdk.EndorBaseServiceInterface); ok {
							endorService = baseService.ToEndorService()
						} else {
							h.logger.Warn(fmt.Sprintf("unable to create entity %s from service", internalEndorService.GetEntity()))
						}
					}
				}
			}

			entities[internalEndorService.GetEntity()] = EndorServiceDictionary{
				OriginalInstance: &internalEndorService,
				EndorService:     endorService,
				entity:           entity,
			}
		}
	}

	// dynamic EndorServices
	if sdk_configuration.GetConfig().HybridEntitiesEnabled || sdk_configuration.GetConfig().DynamicEntitiesEnabled {
		dynamicEntities, err := h.DynamicEntityList()
		if err != nil {
			return map[string]EndorServiceDictionary{}, nil
		}

		for _, entity := range dynamicEntities {
			// check if service is already defined
			// search service
			if v, ok := entities[entity.GetID()]; ok {
				// check entity hybrid
				if entityHybrid, ok := entity.(*sdk.EntityHybrid); ok {
					if hybridInstance, ok := (*v.OriginalInstance).(sdk.EndorHybridServiceInterface); ok {
						defintion, err := entityHybrid.UnmarshalAdditionalAttributes()
						if err != nil {
							h.logger.Warn(fmt.Sprintf("unable to unmarshal definition for hybrid entity %s: %s", entityHybrid.ID, err.Error()))
						}
						// inject dynamic schema
						v.EndorService = hybridInstance.ToEndorService(*defintion)
						entities[entity.GetID()] = v
					}
				}

				// check entity specialized
				if entitySpecialized, ok := entity.(*sdk.EntityHybridSpecialized); ok {
					if specializedInstance, ok := (*v.OriginalInstance).(sdk.EndorHybridSpecializedServiceInterface); ok {
						defintion, err := entitySpecialized.UnmarshalAdditionalAttributes()
						if err != nil {
							h.logger.Warn(fmt.Sprintf("unable to unmarshal definition for hybrid specialized entity %s: %s", entitySpecialized.ID, err.Error()))
						}
						// inject categories and schema
						categories := []sdk.EndorHybridSpecializedServiceCategoryInterface{}
						categoriesAdditionalSchema := map[string]sdk.RootSchema{}
						for _, c := range entitySpecialized.Categories {
							categories = append(categories, NewEndorHybridSpecializedServiceCategory[*sdk.DynamicEntitySpecialized](c.ID, c.Description))
							categoryAdditionalSchema, err := c.UnmarshalAdditionalAttributes()
							if err != nil {
								h.logger.Warn(fmt.Sprintf("unable to unmarshal category definition %s for hybrid entity %s: %s", c.ID, entitySpecialized.ID, err.Error()))
							}
							categoriesAdditionalSchema[c.ID] = *categoryAdditionalSchema
						}
						v.EndorService = specializedInstance.WithHybridCategories(categories).ToEndorService(*defintion, categoriesAdditionalSchema, entitySpecialized.AdditionalCategories)
						entities[entity.GetID()] = v
					}
				}
			} else {
				if entityHybrid, ok := entity.(*sdk.EntityHybrid); ok {
					defintion, err := entityHybrid.UnmarshalAdditionalAttributes()
					if err != nil {
						h.logger.Warn(fmt.Sprintf("unable to unmarshal definition for dynamic entity %s: %s", entityHybrid.ID, err.Error()))
					}
					// create a new hybrid service
					hybridService := NewEndorHybridService[*sdk.DynamicEntity](entityHybrid.ID, entityHybrid.Description)
					schema, _ := sdk.NewSchema(sdk.DynamicEntity{}).ToYAML()
					entityHybrid.Schema = schema
					entities[entityHybrid.ID] = EndorServiceDictionary{
						EndorService: hybridService.ToEndorService(*defintion),
						entity:       entityHybrid,
					}
				}
				if entitySpecialized, ok := entity.(*sdk.EntityHybridSpecialized); ok {
					defintion, err := entitySpecialized.UnmarshalAdditionalAttributes()
					if err != nil {
						h.logger.Warn(fmt.Sprintf("unable to unmarshal definition for dynamic specialized entity %s: %s", entitySpecialized.ID, err.Error()))
					}
					// create categories
					categories := []sdk.EndorHybridSpecializedServiceCategoryInterface{}
					categoriesAdditionalSchema := map[string]sdk.RootSchema{}
					for _, c := range entitySpecialized.Categories {
						categories = append(categories, NewEndorHybridSpecializedServiceCategory[*sdk.DynamicEntitySpecialized](c.ID, c.Description))
						categoryAdditionalSchema, err := c.UnmarshalAdditionalAttributes()
						if err != nil {
							h.logger.Warn(fmt.Sprintf("unable to unmarshal category definition %s for dynamic specialized entity %s: %s", c.ID, entitySpecialized.ID, err.Error()))
						}
						categoriesAdditionalSchema[c.ID] = *categoryAdditionalSchema
					}
					// create a new specilized service
					hybridService := NewHybridSpecializedService[*sdk.DynamicEntitySpecialized](entitySpecialized.ID, entitySpecialized.Description).WithHybridCategories(categories)
					schema, _ := sdk.NewSchema(sdk.DynamicEntitySpecialized{}).ToYAML()
					entitySpecialized.Schema = schema
					entities[entitySpecialized.ID] = EndorServiceDictionary{
						EndorService: hybridService.ToEndorService(*defintion, categoriesAdditionalSchema, entitySpecialized.AdditionalCategories),
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

func (h *EndorServiceRepository) DictionaryActionMap() (map[string]EndorServiceActionDictionary, error) {
	actions := map[string]EndorServiceActionDictionary{}
	entities, err := h.DictionaryMap()
	if err != nil {
		return actions, err
	}
	for entityName, entity := range entities {
		for actionName, EndorServiceAction := range entity.EndorService.Actions {
			action, err := h.createAction(entityName, actionName, EndorServiceAction)
			if err == nil {
				actions[action.entityAction.ID] = *action
			}
		}
	}
	return actions, nil
}

func (h *EndorServiceRepository) DictionaryInstance(dto sdk.ReadInstanceDTO) (*EndorServiceDictionary, error) {
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

func (h *EndorServiceRepository) DictionaryActionInstance(dto sdk.ReadInstanceDTO) (*EndorServiceActionDictionary, error) {
	idSegments := strings.Split(dto.Id, "/")
	if len(idSegments) == 2 {
		entityInstance, err := h.DictionaryInstance(sdk.ReadInstanceDTO{
			Id: idSegments[0],
		})
		if err != nil {
			return nil, err
		}
		if entityAction, ok := entityInstance.EndorService.Actions[idSegments[1]]; ok {
			return h.createAction(idSegments[0], idSegments[1], entityAction)
		} else {
			return nil, sdk.NewNotFoundError(fmt.Errorf("entity action not found"))
		}
	} else {
		return nil, sdk.NewBadRequestError(fmt.Errorf("invalid entity action id"))
	}
}

func (h *EndorServiceRepository) EntityActionList() ([]sdk.EntityAction, error) {
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

func (h *EndorServiceRepository) EndorServiceList() ([]sdk.EndorService, error) {
	entities, err := h.DictionaryMap()
	if err != nil {
		return []sdk.EndorService{}, err
	}
	entityList := make([]sdk.EndorService, 0, len(entities))
	for _, service := range entities {
		entityList = append(entityList, service.EndorService)
	}
	return entityList, nil
}

func (h *EndorServiceRepository) DynamicEntityList() ([]sdk.EntityInterface, error) {
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

func (h *EndorServiceRepository) List(entityType *sdk.EntityType) ([]sdk.EntityInterface, error) {
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

func (h *EndorServiceRepository) Instance(entityType *sdk.EntityType, dto sdk.ReadInstanceDTO) (*sdk.EntityInterface, error) {
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

func (h *EndorServiceRepository) Create(entityType *sdk.EntityType, dto sdk.CreateDTO[sdk.EntityInterface]) (*sdk.EntityInterface, error) {
	if *entityType == sdk.EntityTypeDynamic || *entityType == sdk.EntityTypeDynamicSpecialized {
		dto.Data.SetCategoryType(string(*entityType))
		dto.Data.SetService(h.microServiceId)
		_, err := h.DictionaryInstance(sdk.ReadInstanceDTO{
			Id: dto.Data.GetID(),
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

func (h *EndorServiceRepository) Update(entityType *sdk.EntityType, dto sdk.UpdateByIdDTO[sdk.EntityInterface]) (*sdk.EntityInterface, error) {
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
			return &dto.Data, err
		}
		update := bson.M{"$set": bson.Raw(updateBson)}
		filter := bson.M{"_id": dto.Id}
		_, err = h.collection.UpdateOne(h.context, filter, update)
		if err != nil {
			return nil, err
		}

		h.invalidateCache()
		h.reloadRouteConfiguration(h.microServiceId)

		return &dto.Data, nil
	} else {
		return nil, sdk.NewForbiddenError(fmt.Errorf("update entity not permitted"))
	}
}

func (h *EndorServiceRepository) Delete(entityType *sdk.EntityType, dto sdk.ReadInstanceDTO) error {
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
func (h *EndorServiceRepository) invalidateCache() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cacheInitialized = false
	h.cachedDictionary = nil
}

func (h *EndorServiceRepository) reloadRouteConfiguration(microserviceId string) error {
	config := sdk_configuration.GetConfig()
	entities, err := h.EndorServiceList()
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

func (h *EndorServiceRepository) createAction(entityName string, actionName string, endorServiceAction sdk.EndorServiceActionInterface) (*EndorServiceActionDictionary, error) {
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
	return &EndorServiceActionDictionary{
		EndorServiceAction: endorServiceAction,
		entityAction:       action,
	}, nil
}

func (h *EndorServiceRepository) ensureEntityDocumentOfInternalService(entity sdk.EntityInterface) {
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
