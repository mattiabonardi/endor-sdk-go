package sdk_resource

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

// COLLECTION_RESOURCES is the MongoDB collection name for resources
const COLLECTION_RESOURCES = "resources"

func NewEndorServiceRepository(microServiceId string, internalEndorServices *[]sdk.EndorServiceInterface) *EndorServiceRepository {
	serviceRepository := &EndorServiceRepository{
		microServiceId:        microServiceId,
		internalEndorServices: internalEndorServices,
		context:               context.TODO(),
	}
	if configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled {
		client, _ := sdk.GetMongoClient()
		database := client.Database(configuration.GetConfig().DynamicResourceDocumentDBName)
		serviceRepository.collection = database.Collection(COLLECTION_RESOURCES)
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
	resource         sdk.ResourceInterface
}

type EndorServiceActionDictionary struct {
	EndorServiceAction sdk.EndorServiceActionInterface
	resourceAction     sdk.ResourceAction
}

func (h *EndorServiceRepository) ensureResourceDocumentOfInternalService(resource sdk.ResourceInterface) {
	if (configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled) && h.collection != nil {
		// Check if document exists in MongoDB
		var existingDoc sdk.Resource
		filter := bson.M{"_id": resource.GetID()}
		err := h.collection.FindOne(h.context, filter).Decode(&existingDoc)

		// If document doesn't exist, create it
		if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
			_, insertErr := h.collection.InsertOne(h.context, resource)
			if insertErr != nil {
				// Log error but don't fail the initialization
				// TODO: Add proper logging
			}
		}
	}
}

func (h *EndorServiceRepository) Map() (map[string]EndorServiceDictionary, error) {
	resources := map[string]EndorServiceDictionary{}

	// internal EndorServices
	if h.internalEndorServices != nil {
		for _, internalEndorService := range *h.internalEndorServices {
			resource := sdk.Resource{
				ID:          internalEndorService.GetResource(),
				Description: internalEndorService.GetResourceDescription(),
				Service:     h.microServiceId,
				Type:        string(sdk.ResourceTypeBase),
			}

			var endorService sdk.EndorService

			// hybrid specialized
			if hybridSpecializedService, ok := internalEndorService.(sdk.EndorHybridSpecializedServiceInterface); ok {
				resource.Type = string(sdk.ResourceTypeHybridSpecialized)
				h.ensureResourceDocumentOfInternalService(&resource)
				endorService = hybridSpecializedService.ToEndorService(sdk.Schema{}, map[string]sdk.Schema{})
			} else {
				// hybrid
				if hybridService, ok := internalEndorService.(sdk.EndorHybridServiceInterface); ok {
					resource.Type = string(sdk.ResourceTypeHybrid)
					h.ensureResourceDocumentOfInternalService(&resource)
					endorService = hybridService.ToEndorService(sdk.Schema{})
				} else {
					// base specialized
					if baseSpecializedService, ok := internalEndorService.(sdk.EndorBaseSpecializedServiceInterface); ok {
						resource.Type = string(sdk.ResourceTypeBaseSpecialized)
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

			resources[internalEndorService.GetResource()] = EndorServiceDictionary{
				OriginalInstance: &internalEndorService,
				EndorService:     endorService,
				resource:         &resource,
			}
		}
	}

	// dynamic EndorServices
	if configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled {
		dynamicResources, err := h.DynamicResourceList()
		if err != nil {
			return map[string]EndorServiceDictionary{}, nil
		}

		for _, resource := range dynamicResources {
			// check if service is already defined
			// search service
			if v, ok := resources[*resource.GetID()]; ok {
				// check resource hybrid
				if resourceHybrid, ok := resource.(*sdk.ResourceHybrid); ok {
					if hybridInstance, ok := (*v.OriginalInstance).(sdk.EndorHybridServiceInterface); ok {
						defintion, err := resourceHybrid.UnmarshalAdditionalAttributes()
						if err != nil {
							// TODO: log
						}
						// inject dynamic schema
						v.EndorService = hybridInstance.ToEndorService(defintion.Schema)
						resources[*resource.GetID()] = v
					} else {
						//TODO: log that only hybrid service supports additional attributes
					}
				}

				// check resource specialized
				if resourceSpecialized, ok := resource.(*sdk.ResourceHybridSpecialized); ok {
					if specializedInstance, ok := (*v.OriginalInstance).(sdk.EndorHybridSpecializedServiceInterface); ok {
						defintion, err := resourceSpecialized.UnmarshalAdditionalAttributes()
						if err != nil {
							// TODO: log
						}
						// inject categories and schema
						categories := []sdk.EndorHybridSpecializedServiceCategoryInterface{}
						categoriesAdditionalSchema := map[string]sdk.Schema{}
						for _, c := range resourceSpecialized.Categories {
							categories = append(categories, NewEndorHybridSpecializedServiceCategory[*sdk.DynamicResourceSpecialized](c.ID, c.Description))
							categoryAdditionalSchema, err := c.UnmarshalAdditionalAttributes()
							if err != nil {
								// TODO: log
							}
							categoriesAdditionalSchema[c.ID] = categoryAdditionalSchema.Schema
						}
						v.EndorService = specializedInstance.WithCategories(categories).ToEndorService(defintion.Schema, categoriesAdditionalSchema)
						resources[*resource.GetID()] = v
					} else {
						//TODO: log that only hybrid service supports additional attributes
					}
				}
			} else {
				if resourceHybrid, ok := resource.(*sdk.ResourceHybrid); ok {
					defintion, err := resourceHybrid.UnmarshalAdditionalAttributes()
					if err != nil {
						// TODO: log
					}
					// create a new hybrid service
					hybridService := NewEndorHybridService[*sdk.DynamicResource](resourceHybrid.ID, resourceHybrid.Description)
					resources[resourceHybrid.ID] = EndorServiceDictionary{
						EndorService: hybridService.ToEndorService(defintion.Schema),
						resource:     resourceHybrid,
					}
				}
				if resourceSpecialized, ok := resource.(*sdk.ResourceHybridSpecialized); ok {
					defintion, err := resourceSpecialized.UnmarshalAdditionalAttributes()
					if err != nil {
						// TODO: log
					}
					// create categories
					categories := []sdk.EndorHybridSpecializedServiceCategoryInterface{}
					categoriesAdditionalSchema := map[string]sdk.Schema{}
					for _, c := range resourceSpecialized.Categories {
						categories = append(categories, NewEndorHybridSpecializedServiceCategory[*sdk.DynamicResourceSpecialized](c.ID, c.Description))
						categoryAdditionalSchema, err := c.UnmarshalAdditionalAttributes()
						if err != nil {
							// TODO: log
						}
						categoriesAdditionalSchema[c.ID] = categoryAdditionalSchema.Schema
					}
					// create a new specilized service
					hybridService := NewHybridSpecializedService[*sdk.DynamicResourceSpecialized](resourceSpecialized.ID, resourceSpecialized.Description).WithCategories(categories)
					resources[resourceSpecialized.ID] = EndorServiceDictionary{
						EndorService: hybridService.ToEndorService(defintion.Schema, categoriesAdditionalSchema),
						resource:     resourceSpecialized,
					}
				}

			}
		}
	}
	return resources, nil
}

func (h *EndorServiceRepository) ActionMap() (map[string]EndorServiceActionDictionary, error) {
	actions := map[string]EndorServiceActionDictionary{}
	resources, err := h.Map()
	if err != nil {
		return actions, err
	}
	for resourceName, resource := range resources {
		for actionName, EndorServiceAction := range resource.EndorService.Actions {
			action, err := h.createAction(resourceName, actionName, EndorServiceAction)
			if err == nil {
				actions[action.resourceAction.ID] = *action
			}
		}
	}
	return actions, nil
}

func (h *EndorServiceRepository) ResourceActionList() ([]sdk.ResourceAction, error) {
	actions, err := h.ActionMap()
	if err != nil {
		return []sdk.ResourceAction{}, err
	}
	actionList := make([]sdk.ResourceAction, 0, len(actions))
	for _, action := range actions {
		actionList = append(actionList, action.resourceAction)
	}
	return actionList, nil
}

func (h *EndorServiceRepository) ResourceList() ([]sdk.ResourceInterface, error) {
	resources, err := h.Map()
	if err != nil {
		return []sdk.ResourceInterface{}, err
	}
	resourceList := make([]sdk.ResourceInterface, 0, len(resources))
	for _, service := range resources {
		resourceList = append(resourceList, service.resource)
	}
	return resourceList, nil
}

func (h *EndorServiceRepository) EndorServiceList() ([]sdk.EndorService, error) {
	resources, err := h.Map()
	if err != nil {
		return []sdk.EndorService{}, err
	}
	resourceList := make([]sdk.EndorService, 0, len(resources))
	for _, service := range resources {
		resourceList = append(resourceList, service.EndorService)
	}
	return resourceList, nil
}

func (h *EndorServiceRepository) DynamicResourceList() ([]sdk.ResourceInterface, error) {
	cursor, err := h.collection.Find(h.context, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(h.context)

	var resources []sdk.ResourceInterface

	for cursor.Next(h.context) {
		raw := cursor.Current

		// First decode to Resource base
		var discr sdk.Resource
		if err := bson.Unmarshal(raw, &discr); err != nil {
			return nil, fmt.Errorf("failed to read resource type: %w", err)
		}

		switch *discr.GetCategoryType() {
		case string(sdk.ResourceTypeBase):
			resources = append(resources, &discr)

		case string(sdk.ResourceTypeBaseSpecialized):
			var r sdk.ResourceSpecialized
			if err := bson.Unmarshal(raw, &r); err != nil {
				return nil, fmt.Errorf("failed to decode base specialized resource: %w", err)
			}
			resources = append(resources, &r)

		case string(sdk.ResourceTypeHybrid):
			var r sdk.ResourceHybrid
			if err := bson.Unmarshal(raw, &r); err != nil {
				return nil, fmt.Errorf("failed to decode hybrid resource: %w", err)
			}
			resources = append(resources, &r)

		case string(sdk.ResourceTypeHybridSpecialized):
			var r sdk.ResourceHybridSpecialized
			if err := bson.Unmarshal(raw, &r); err != nil {
				return nil, fmt.Errorf("failed to decode hybrid-specialized resource: %w", err)
			}
			resources = append(resources, &r)

		default:
			return nil, fmt.Errorf("unknown resource type: %s", discr.Type)
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}

func (h *EndorServiceRepository) Instance(dto sdk.ReadInstanceDTO) (*EndorServiceDictionary, error) {
	// get all service
	resources, err := h.Map()
	if err != nil {
		return nil, err
	}
	if resource, ok := resources[dto.Id]; ok {
		return &resource, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s not found", dto.Id))
}

func (h *EndorServiceRepository) ActionInstance(dto sdk.ReadInstanceDTO) (*EndorServiceActionDictionary, error) {
	idSegments := strings.Split(dto.Id, "/")
	if len(idSegments) == 2 {
		resourceInstance, err := h.Instance(sdk.ReadInstanceDTO{
			Id: idSegments[0],
		})
		if err != nil {
			return nil, err
		}
		if resourceAction, ok := resourceInstance.EndorService.Actions[idSegments[1]]; ok {
			return h.createAction(idSegments[0], idSegments[1], resourceAction)
		} else {
			return nil, sdk.NewNotFoundError(fmt.Errorf("resource action not found"))
		}
	} else {
		return nil, sdk.NewBadRequestError(fmt.Errorf("invalid resource action id"))
	}
}

func (h *EndorServiceRepository) Create(dto sdk.CreateDTO[sdk.ResourceInterface]) error {
	dto.Data.SetService(h.microServiceId)
	_, err := h.Instance(sdk.ReadInstanceDTO{
		Id: *dto.Data.GetID(),
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
		return sdk.NewConflictError(fmt.Errorf("resource already exist"))
	}
}

func (h *EndorServiceRepository) Update(id string, data sdk.ResourceInterface) (*sdk.ResourceInterface, error) {
	var instance *sdk.ResourceInterface
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
	// check if resources already exist
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
	resources, err := h.EndorServiceList()
	if err != nil {
		return err
	}
	err = api_gateway.InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), resources)
	if err != nil {
		return err
	}
	_, err = swagger.CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), resources, "/api")
	if err != nil {
		return err
	}
	return nil
}

func (h *EndorServiceRepository) createAction(resourceName string, actionName string, endorServiceAction sdk.EndorServiceActionInterface) (*EndorServiceActionDictionary, error) {
	actionId := path.Join(resourceName, actionName)
	action := sdk.ResourceAction{
		ID:          actionId,
		Resource:    resourceName,
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
		resourceAction:     action,
	}, nil
}
