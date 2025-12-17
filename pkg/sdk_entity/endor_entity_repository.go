package sdk_entity

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/internal/api_gateway"
	"github.com/mattiabonardi/endor-sdk-go/internal/configuration"
	"github.com/mattiabonardi/endor-sdk-go/internal/swagger"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// COLLECTION_ENTITIES is the MongoDB collection name for entities
const COLLECTION_ENTITIES = "entities"

func NewEndorServiceRepository(microServiceId string, internalEndorServices *[]sdk.EndorServiceInterface) *EndorServiceRepository {
	serviceRepository := &EndorServiceRepository{
		microServiceId:        microServiceId,
		internalEndorServices: internalEndorServices,
		context:               context.TODO(),
	}
	if configuration.GetConfig().HybridEntitiesEnabled || configuration.GetConfig().DynamicEntitiesEnabled {
		client, _ := sdk.GetMongoClient()
		database := client.Database(configuration.GetConfig().DynamicEntityDocumentDBName)
		serviceRepository.collection = database.Collection(COLLECTION_ENTITIES)
	}

	return serviceRepository
}

type EndorServiceRepository struct {
	microServiceId        string
	internalEndorServices *[]sdk.EndorServiceInterface
	collection            *mongo.Collection
	context               context.Context
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

func (h *EndorServiceRepository) ensureEntityDocumentOfInternalService(entity sdk.EntityInterface) {
	if (configuration.GetConfig().HybridEntitiesEnabled || configuration.GetConfig().DynamicEntitiesEnabled) && h.collection != nil {
		// Check if document exists in MongoDB
		var existingDoc sdk.Entity
		filter := bson.M{"_id": entity.GetID()}
		err := h.collection.FindOne(h.context, filter).Decode(&existingDoc)

		// If document doesn't exist, create it
		if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
			_, insertErr := h.collection.InsertOne(h.context, entity)
			if insertErr != nil {
				// Log error but don't fail the initialization
				// TODO: Add proper logging
			}
		}
	}
}

func (h *EndorServiceRepository) Map() (map[string]EndorServiceDictionary, error) {
	entities := map[string]EndorServiceDictionary{}

	// internal EndorServices
	if h.internalEndorServices != nil {
		for _, internalEndorService := range *h.internalEndorServices {
			entity := sdk.Entity{
				ID:          internalEndorService.GetEntity(),
				Description: internalEndorService.GetEntityDescription(),
				Service:     h.microServiceId,
				Type:        string(sdk.EntityTypeBase),
			}

			var endorService sdk.EndorService

			// hybrid specialized
			if hybridSpecializedService, ok := internalEndorService.(sdk.EndorHybridSpecializedServiceInterface); ok {
				entity.Type = string(sdk.EntityTypeHybridSpecialized)
				h.ensureEntityDocumentOfInternalService(&entity)
				endorService = hybridSpecializedService.ToEndorService(sdk.Schema{}, map[string]sdk.Schema{})
			} else {
				// hybrid
				if hybridService, ok := internalEndorService.(sdk.EndorHybridServiceInterface); ok {
					entity.Type = string(sdk.EntityTypeHybrid)
					h.ensureEntityDocumentOfInternalService(&entity)
					endorService = hybridService.ToEndorService(sdk.Schema{})
				} else {
					// base specialized
					if baseSpecializedService, ok := internalEndorService.(sdk.EndorBaseSpecializedServiceInterface); ok {
						entity.Type = string(sdk.EntityTypeBaseSpecialized)
						endorService = baseSpecializedService.ToEndorService()
					} else {
						// base
						if baseService, ok := internalEndorService.(sdk.EndorBaseServiceInterface); ok {
							endorService = baseService.ToEndorService()
						} else {
							// TODO: log
						}
					}
				}
			}

			entities[internalEndorService.GetEntity()] = EndorServiceDictionary{
				OriginalInstance: &internalEndorService,
				EndorService:     endorService,
				entity:           &entity,
			}
		}
	}

	// dynamic EndorServices
	if configuration.GetConfig().HybridEntitiesEnabled || configuration.GetConfig().DynamicEntitiesEnabled {
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
							// TODO: log
						}
						// inject dynamic schema
						v.EndorService = hybridInstance.ToEndorService(defintion.Schema)
						entities[entity.GetID()] = v
					} else {
						//TODO: log that only hybrid service supports additional attributes
					}
				}

				// check entity specialized
				if entitySpecialized, ok := entity.(*sdk.EntityHybridSpecialized); ok {
					if specializedInstance, ok := (*v.OriginalInstance).(sdk.EndorHybridSpecializedServiceInterface); ok {
						defintion, err := entitySpecialized.UnmarshalAdditionalAttributes()
						if err != nil {
							// TODO: log
						}
						// inject categories and schema
						categories := []sdk.EndorHybridSpecializedServiceCategoryInterface{}
						categoriesAdditionalSchema := map[string]sdk.Schema{}
						for _, c := range entitySpecialized.Categories {
							categories = append(categories, NewEndorHybridSpecializedServiceCategory[*sdk.DynamicEntitySpecialized](c.ID, c.Description))
							categoryAdditionalSchema, err := c.UnmarshalAdditionalAttributes()
							if err != nil {
								// TODO: log
							}
							categoriesAdditionalSchema[c.ID] = categoryAdditionalSchema.Schema
						}
						v.EndorService = specializedInstance.WithCategories(categories).ToEndorService(defintion.Schema, categoriesAdditionalSchema)
						entities[entity.GetID()] = v
					} else {
						//TODO: log that only hybrid service supports additional attributes
					}
				}
			} else {
				if entityHybrid, ok := entity.(*sdk.EntityHybrid); ok {
					defintion, err := entityHybrid.UnmarshalAdditionalAttributes()
					if err != nil {
						// TODO: log
					}
					// create a new hybrid service
					hybridService := NewEndorHybridService[*sdk.DynamicEntity](entityHybrid.ID, entityHybrid.Description)
					entities[entityHybrid.ID] = EndorServiceDictionary{
						EndorService: hybridService.ToEndorService(defintion.Schema),
						entity:       entityHybrid,
					}
				}
				if entitySpecialized, ok := entity.(*sdk.EntityHybridSpecialized); ok {
					defintion, err := entitySpecialized.UnmarshalAdditionalAttributes()
					if err != nil {
						// TODO: log
					}
					// create categories
					categories := []sdk.EndorHybridSpecializedServiceCategoryInterface{}
					categoriesAdditionalSchema := map[string]sdk.Schema{}
					for _, c := range entitySpecialized.Categories {
						categories = append(categories, NewEndorHybridSpecializedServiceCategory[*sdk.DynamicEntitySpecialized](c.ID, c.Description))
						categoryAdditionalSchema, err := c.UnmarshalAdditionalAttributes()
						if err != nil {
							// TODO: log
						}
						categoriesAdditionalSchema[c.ID] = categoryAdditionalSchema.Schema
					}
					// create a new specilized service
					hybridService := NewHybridSpecializedService[*sdk.DynamicEntitySpecialized](entitySpecialized.ID, entitySpecialized.Description).WithCategories(categories)
					entities[entitySpecialized.ID] = EndorServiceDictionary{
						EndorService: hybridService.ToEndorService(defintion.Schema, categoriesAdditionalSchema),
						entity:       entitySpecialized,
					}
				}

			}
		}
	}
	return entities, nil
}

func (h *EndorServiceRepository) ActionMap() (map[string]EndorServiceActionDictionary, error) {
	actions := map[string]EndorServiceActionDictionary{}
	entities, err := h.Map()
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

func (h *EndorServiceRepository) EntityActionList() ([]sdk.EntityAction, error) {
	actions, err := h.ActionMap()
	if err != nil {
		return []sdk.EntityAction{}, err
	}
	actionList := make([]sdk.EntityAction, 0, len(actions))
	for _, action := range actions {
		actionList = append(actionList, action.entityAction)
	}
	return actionList, nil
}

func (h *EndorServiceRepository) EntityList() ([]sdk.EntityInterface, error) {
	entities, err := h.Map()
	if err != nil {
		return []sdk.EntityInterface{}, err
	}
	entityList := make([]sdk.EntityInterface, 0, len(entities))
	for _, service := range entities {
		entityList = append(entityList, service.entity)
	}
	return entityList, nil
}

func (h *EndorServiceRepository) EndorServiceList() ([]sdk.EndorService, error) {
	entities, err := h.Map()
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
			return nil, fmt.Errorf("failed to read entity type: %w", err)
		}

		switch discr.GetCategoryType() {
		case string(sdk.EntityTypeBase):
			entities = append(entities, &discr)

		case string(sdk.EntityTypeBaseSpecialized):
			var r sdk.EntitySpecialized
			if err := bson.Unmarshal(raw, &r); err != nil {
				return nil, fmt.Errorf("failed to decode base specialized entity: %w", err)
			}
			entities = append(entities, &r)

		case string(sdk.EntityTypeHybrid):
			var r sdk.EntityHybrid
			if err := bson.Unmarshal(raw, &r); err != nil {
				return nil, fmt.Errorf("failed to decode hybrid entity: %w", err)
			}
			entities = append(entities, &r)

		case string(sdk.EntityTypeHybridSpecialized):
			var r sdk.EntityHybridSpecialized
			if err := bson.Unmarshal(raw, &r); err != nil {
				return nil, fmt.Errorf("failed to decode hybrid-specialized entity: %w", err)
			}
			entities = append(entities, &r)

		default:
			return nil, fmt.Errorf("unknown entity type: %s", discr.Type)
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return entities, nil
}

func (h *EndorServiceRepository) Instance(dto sdk.ReadInstanceDTO) (*EndorServiceDictionary, error) {
	// get all service
	entities, err := h.Map()
	if err != nil {
		return nil, err
	}
	if entity, ok := entities[dto.Id]; ok {
		return &entity, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id))
}

func (h *EndorServiceRepository) ActionInstance(dto sdk.ReadInstanceDTO) (*EndorServiceActionDictionary, error) {
	idSegments := strings.Split(dto.Id, "/")
	if len(idSegments) == 2 {
		entityInstance, err := h.Instance(sdk.ReadInstanceDTO{
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

func (h *EndorServiceRepository) Create(dto sdk.CreateDTO[sdk.EntityInterface]) error {
	dto.Data.SetService(h.microServiceId)
	_, err := h.Instance(sdk.ReadInstanceDTO{
		Id: dto.Data.GetID(),
	})
	var endorError *sdk.EndorError
	if errors.As(err, &endorError) && endorError.StatusCode == 404 {
		_, err := h.collection.InsertOne(h.context, dto.Data)
		if err != nil {
			return err
		}
		h.reloadRouteConfiguration(h.microServiceId)
		return nil
	} else {
		return sdk.NewConflictError(fmt.Errorf("entity already exist"))
	}
}

func (h *EndorServiceRepository) Update(id string, data sdk.EntityInterface) (*sdk.EntityInterface, error) {
	var instance *sdk.EntityInterface
	_, err := h.Instance(sdk.ReadInstanceDTO{
		Id: id,
	})
	if err != nil {
		return instance, err
	}
	updateBson, err := bson.Marshal(data)
	if err != nil {
		return &data, err
	}
	update := bson.M{"$set": bson.Raw(updateBson)}
	filter := bson.M{"_id": id}
	_, err = h.collection.UpdateOne(h.context, filter, update)
	if err != nil {
		return nil, err
	}

	h.reloadRouteConfiguration(h.microServiceId)

	return &data, nil
}

func (h *EndorServiceRepository) Delete(dto sdk.ReadInstanceDTO) error {
	// check if entities already exist
	_, err := h.Instance(dto)
	if err != nil {
		return err
	}
	_, err = h.collection.DeleteOne(h.context, bson.M{"_id": dto.Id})
	if err != nil {
		h.reloadRouteConfiguration(h.microServiceId)
	}
	return err
}

func (h *EndorServiceRepository) reloadRouteConfiguration(microserviceId string) error {
	config := configuration.GetConfig()
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
