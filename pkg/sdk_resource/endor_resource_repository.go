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

func NewEndorServiceRepository(microServiceId string, internalEndorServices *[]sdk.EndorService, internalHybridServices *[]sdk.EndorHybridService) *EndorServiceRepository {
	serviceRepository := &EndorServiceRepository{
		microServiceId:         microServiceId,
		internalEndorServices:  internalEndorServices,
		internalHybridServices: internalHybridServices,
		context:                context.TODO(),
	}
	if configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled {
		client, _ := sdk.GetMongoClient()
		database := client.Database(configuration.GetConfig().DynamicResourceDocumentDBName)
		serviceRepository.collection = database.Collection(COLLECTION_RESOURCES)
	}

	return serviceRepository
}

type EndorServiceRepository struct {
	microServiceId         string
	internalEndorServices  *[]sdk.EndorService
	internalHybridServices *[]sdk.EndorHybridService
	collection             *mongo.Collection
	context                context.Context
}

type EndorServiceDictionary struct {
	EndorService sdk.EndorService
	resource     sdk.Resource
}

type EndorServiceActionDictionary struct {
	EndorServiceAction sdk.EndorServiceAction
	resourceAction     sdk.ResourceAction
}

// ensureHybridServiceDocument ensures that a MongoDB document exists for the hybrid service
func (h *EndorServiceRepository) ensureHybridServiceDocument(hybridService sdk.EndorHybridService) {
	if (configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled) && h.collection != nil {
		// Check if document exists in MongoDB
		var existingDoc sdk.Resource
		filter := bson.M{"_id": hybridService.GetResource()}
		err := h.collection.FindOne(h.context, filter).Decode(&existingDoc)

		// If document doesn't exist, create it
		if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
			newResource := sdk.Resource{
				ID:                   hybridService.GetResource(),
				Description:          hybridService.GetResourceDescription(),
				Service:              h.microServiceId,
				AdditionalAttributes: "{}", // Empty JSON for additional attributes
			}

			_, insertErr := h.collection.InsertOne(h.context, newResource)
			if insertErr != nil {
				// Log error but don't fail the initialization
				// TODO: Add proper logging
			}
		}
	}
}

func (h *EndorServiceRepository) Map() (map[string]EndorServiceDictionary, error) {
	resources := map[string]EndorServiceDictionary{}

	// 1. Internal EndorServices (highest priority)
	if h.internalEndorServices != nil {
		for _, internalEndorService := range *h.internalEndorServices {
			resource := sdk.Resource{
				ID:          internalEndorService.Resource,
				Description: internalEndorService.Description,
				Service:     h.microServiceId,
			}
			resources[internalEndorService.Resource] = EndorServiceDictionary{
				EndorService: internalEndorService,
				resource:     resource,
			}
		}
	}

	// 2. Internal EndorHybridServices (medium priority - with empty schema initially)
	if h.internalHybridServices != nil {
		for _, hybridService := range *h.internalHybridServices {
			// Skip if already handled by EndorService
			if _, exists := resources[hybridService.GetResource()]; exists {
				continue
			}

			// Ensure MongoDB document exists for this hybrid service
			h.ensureHybridServiceDocument(hybridService)

			resource := sdk.Resource{
				ID:          hybridService.GetResource(),
				Description: hybridService.GetResourceDescription(),
				Service:     h.microServiceId,
			}
			// Convert hybrid service with empty schema (will be updated if MongoDB document exists)
			resources[hybridService.GetResource()] = EndorServiceDictionary{
				EndorService: hybridService.ToEndorService(sdk.Schema{}),
				resource:     resource,
			}
		}
	}

	// 3. Dynamic resources from MongoDB (lowest priority + schema injection)
	if configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled {
		dynamicResources, err := h.DynamiResourceList()
		if err != nil {
			return map[string]EndorServiceDictionary{}, nil
		}

		for _, resource := range dynamicResources {
			defintion, err := resource.UnmarshalAdditionalAttributes()

			categories := []sdk.EndorHybridServiceCategory{}
			for _, c := range resource.Categories {
				categories = append(categories, &EndorHybridServiceCategoryImpl[*sdk.DynamicResource, *sdk.DynamicResourceSpecialized]{
					Category: c,
				})
			}

			if err == nil {
				// Check if there's a corresponding hybrid service for this resource
				var foundHybridService sdk.EndorHybridService
				if h.internalHybridServices != nil {
					for i, hybridService := range *h.internalHybridServices {
						if hybridService.GetResource() == resource.ID {
							foundHybridService = (*h.internalHybridServices)[i]
							break
						}
					}
				}

				// If this resource already exists (EndorService or HybridService)
				if _, exists := resources[resource.ID]; exists {
					// Update ONLY hybrid services with MongoDB schema
					if foundHybridService != nil {
						resources[resource.ID] = EndorServiceDictionary{
							EndorService: foundHybridService.WithCategories(categories).ToEndorService(defintion.Schema),
							resource:     resource,
						}
					}
					// Skip EndorServices (they have absolute priority)
					continue
				}

				// Create new service for resources not handled internally
				if foundHybridService != nil {
					// Use the existing hybrid service with the dynamic schema
					resources[resource.ID] = EndorServiceDictionary{
						EndorService: foundHybridService.WithCategories(categories).ToEndorService(defintion.Schema),
						resource:     resource,
					}
				} else {
					// Create default hybrid service with all 6 actions
					hybridService := NewHybridService[*sdk.DynamicResource](resource.ID, resource.Description)
					resources[resource.ID] = EndorServiceDictionary{
						EndorService: hybridService.WithCategories(categories).ToEndorService(defintion.Schema),
						resource:     resource,
					}
				}
			} else {
				// TODO: non blocked log
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
		for actionName, EndorServiceAction := range resource.EndorService.Methods {
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

func (h *EndorServiceRepository) ResourceList() ([]sdk.Resource, error) {
	resources, err := h.Map()
	if err != nil {
		return []sdk.Resource{}, err
	}
	resourceList := make([]sdk.Resource, 0, len(resources))
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

func (h *EndorServiceRepository) DynamiResourceList() ([]sdk.Resource, error) {
	cursor, err := h.collection.Find(h.context, bson.M{})
	if err != nil {
		return nil, err
	}
	var storedResources []sdk.Resource
	if err := cursor.All(h.context, &storedResources); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []sdk.Resource{}, nil
		} else {
			return nil, err
		}
	}
	return storedResources, nil
}

func (h *EndorServiceRepository) Instance(dto sdk.ReadInstanceDTO) (*EndorServiceDictionary, error) {
	// search from internal services
	for _, service := range *h.internalEndorServices {
		if service.Resource == dto.Id {
			resource := sdk.Resource{
				ID:          service.Resource,
				Description: service.Description,
				Service:     h.microServiceId,
			}
			return &EndorServiceDictionary{
				EndorService: service,
				resource:     resource,
			}, nil
		}
	}
	if configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled {
		// search from database
		resource := sdk.Resource{}
		filter := bson.M{"_id": dto.Id}
		err := h.collection.FindOne(h.context, filter).Decode(&resource)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, sdk.NewNotFoundError(fmt.Errorf("resource not found"))
			} else {
				return nil, err
			}
		}
		additionalAttributesDefinition, err := resource.UnmarshalAdditionalAttributes()
		if err != nil {
			return nil, err
		}

		// Check if there's a corresponding hybrid service for this resource
		var foundHybridService sdk.EndorHybridService
		if h.internalHybridServices != nil {
			for i, hybridService := range *h.internalHybridServices {
				if hybridService.GetResource() == resource.ID {
					foundHybridService = (*h.internalHybridServices)[i]
					break
				}
			}
		}

		categories := []sdk.EndorHybridServiceCategory{}
		for _, c := range resource.Categories {
			categories = append(categories, &EndorHybridServiceCategoryImpl[*sdk.DynamicResource, *sdk.DynamicResourceSpecialized]{
				Category: c,
			})
		} // Use existing hybrid service or create abstract one
		if foundHybridService != nil {
			// Use the existing hybrid service with the dynamic schema
			return &EndorServiceDictionary{
				EndorService: foundHybridService.WithCategories(categories).ToEndorService(additionalAttributesDefinition.Schema),
				resource:     resource,
			}, nil
		} else {
			// Create default hybrid service with all 6 actions
			hybridService := NewHybridService[*sdk.DynamicResource](resource.ID, resource.Description)
			return &EndorServiceDictionary{
				EndorService: hybridService.WithCategories(categories).ToEndorService(additionalAttributesDefinition.Schema),
				resource:     resource,
			}, nil
		}
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
		if resourceAction, ok := resourceInstance.EndorService.Methods[idSegments[1]]; ok {
			return h.createAction(idSegments[0], idSegments[1], resourceAction)
		} else {
			return nil, sdk.NewNotFoundError(fmt.Errorf("resource action not found"))
		}
	} else {
		return nil, sdk.NewBadRequestError(fmt.Errorf("invalid resource action id"))
	}
}

func (h *EndorServiceRepository) Create(dto sdk.CreateDTO[sdk.Resource]) error {
	dto.Data.Service = h.microServiceId
	_, err := h.Instance(sdk.ReadInstanceDTO{
		Id: dto.Data.ID,
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

func (h *EndorServiceRepository) UpdateOne(dto sdk.UpdateByIdDTO[sdk.Resource]) (*sdk.Resource, error) {
	var instance *sdk.Resource
	_, err := h.Instance(sdk.ReadInstanceDTO{
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

	h.reloadRouteConfiguration(h.microServiceId)

	return &dto.Data, nil
}

func (h *EndorServiceRepository) DeleteOne(dto sdk.ReadInstanceDTO) error {
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

func (h *EndorServiceRepository) createAction(resourceName string, actionName string, endorServiceAction sdk.EndorServiceAction) (*EndorServiceActionDictionary, error) {
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
