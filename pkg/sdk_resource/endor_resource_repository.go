package sdk_resource

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewEndorServiceRepository(microServiceId string, internalEndorServices *[]EndorService, internalHybridServices *[]EndorHybridService) *EndorServiceRepository {
	serviceRepository := &EndorServiceRepository{
		microServiceId:         microServiceId,
		internalEndorServices:  internalEndorServices,
		internalHybridServices: internalHybridServices,
		context:                context.TODO(),
	}
	if GetConfig().HybridResourcesEnabled || GetConfig().DynamicResourcesEnabled {
		client, _ := GetMongoClient()
		database := client.Database(GetConfig().DynamicResourceDocumentDBName)
		serviceRepository.collection = database.Collection(COLLECTION_RESOURCES)
	}

	return serviceRepository
}

type EndorServiceRepository struct {
	microServiceId         string
	internalEndorServices  *[]EndorService
	internalHybridServices *[]EndorHybridService
	collection             *mongo.Collection
	context                context.Context
}

type EndorServiceDictionary struct {
	EndorService EndorService
	resource     Resource
}

type EndorServiceActionDictionary struct {
	EndorServiceAction EndorServiceAction
	resourceAction     ResourceAction
}

// ensureHybridServiceDocument ensures that a MongoDB document exists for the hybrid service
func (h *EndorServiceRepository) ensureHybridServiceDocument(hybridService EndorHybridService) {
	if (GetConfig().HybridResourcesEnabled || GetConfig().DynamicResourcesEnabled) && h.collection != nil {
		// Check if document exists in MongoDB
		var existingDoc Resource
		filter := bson.M{"_id": hybridService.GetResource()}
		err := h.collection.FindOne(h.context, filter).Decode(&existingDoc)

		// If document doesn't exist, create it
		if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
			newResource := Resource{
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
			resource := Resource{
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

			resource := Resource{
				ID:          hybridService.GetResource(),
				Description: hybridService.GetResourceDescription(),
				Service:     h.microServiceId,
			}
			// Convert hybrid service with empty schema (will be updated if MongoDB document exists)
			resources[hybridService.GetResource()] = EndorServiceDictionary{
				EndorService: hybridService.ToEndorService(Schema{}),
				resource:     resource,
			}
		}
	}

	// 3. Dynamic resources from MongoDB (lowest priority + schema injection)
	if GetConfig().HybridResourcesEnabled || GetConfig().DynamicResourcesEnabled {
		dynamicResources, err := h.DynamiResourceList()
		if err != nil {
			return map[string]EndorServiceDictionary{}, nil
		}

		for _, resource := range dynamicResources {
			defintion, err := resource.UnmarshalAdditionalAttributes()

			categories := []EndorHybridServiceCategory{}
			for _, c := range resource.Categories {
				categories = append(categories, &EndorHybridServiceCategoryImpl[*DynamicResource, *DynamicResourceSpecialized]{
					Category: c,
				})
			}

			if err == nil {
				// Check if there's a corresponding hybrid service for this resource
				var foundHybridService EndorHybridService
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
					hybridService := NewHybridService[*DynamicResource](resource.ID, resource.Description)
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

func (h *EndorServiceRepository) ResourceActionList() ([]ResourceAction, error) {
	actions, err := h.ActionMap()
	if err != nil {
		return []ResourceAction{}, err
	}
	actionList := make([]ResourceAction, 0, len(actions))
	for _, action := range actions {
		actionList = append(actionList, action.resourceAction)
	}
	return actionList, nil
}

func (h *EndorServiceRepository) ResourceList() ([]Resource, error) {
	resources, err := h.Map()
	if err != nil {
		return []Resource{}, err
	}
	resourceList := make([]Resource, 0, len(resources))
	for _, service := range resources {
		resourceList = append(resourceList, service.resource)
	}
	return resourceList, nil
}

func (h *EndorServiceRepository) EndorServiceList() ([]EndorService, error) {
	resources, err := h.Map()
	if err != nil {
		return []EndorService{}, err
	}
	resourceList := make([]EndorService, 0, len(resources))
	for _, service := range resources {
		resourceList = append(resourceList, service.EndorService)
	}
	return resourceList, nil
}

func (h *EndorServiceRepository) DynamiResourceList() ([]Resource, error) {
	cursor, err := h.collection.Find(h.context, bson.M{})
	if err != nil {
		return nil, err
	}
	var storedResources []Resource
	if err := cursor.All(h.context, &storedResources); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []Resource{}, nil
		} else {
			return nil, err
		}
	}
	return storedResources, nil
}

func (h *EndorServiceRepository) Instance(dto ReadInstanceDTO) (*EndorServiceDictionary, error) {
	// search from internal services
	for _, service := range *h.internalEndorServices {
		if service.Resource == dto.Id {
			resource := Resource{
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
	if GetConfig().HybridResourcesEnabled || GetConfig().DynamicResourcesEnabled {
		// search from database
		resource := Resource{}
		filter := bson.M{"_id": dto.Id}
		err := h.collection.FindOne(h.context, filter).Decode(&resource)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, NewNotFoundError(fmt.Errorf("resource not found"))
			} else {
				return nil, err
			}
		}
		additionalAttributesDefinition, err := resource.UnmarshalAdditionalAttributes()
		if err != nil {
			return nil, err
		}

		// Check if there's a corresponding hybrid service for this resource
		var foundHybridService EndorHybridService
		if h.internalHybridServices != nil {
			for i, hybridService := range *h.internalHybridServices {
				if hybridService.GetResource() == resource.ID {
					foundHybridService = (*h.internalHybridServices)[i]
					break
				}
			}
		}

		categories := []EndorHybridServiceCategory{}
		for _, c := range resource.Categories {
			categories = append(categories, &EndorHybridServiceCategoryImpl[*DynamicResource, *DynamicResourceSpecialized]{
				Category: c,
			})
		}

		// Use existing hybrid service or create abstract one
		if foundHybridService != nil {
			// Use the existing hybrid service with the dynamic schema
			return &EndorServiceDictionary{
				EndorService: foundHybridService.WithCategories(categories).ToEndorService(additionalAttributesDefinition.Schema),
				resource:     resource,
			}, nil
		} else {
			// Create default hybrid service with all 6 actions
			hybridService := NewHybridService[*DynamicResource](resource.ID, resource.Description)
			return &EndorServiceDictionary{
				EndorService: hybridService.WithCategories(categories).ToEndorService(additionalAttributesDefinition.Schema),
				resource:     resource,
			}, nil
		}
	}
	return nil, NewNotFoundError(fmt.Errorf("resource %s not found", dto.Id))
}

func (h *EndorServiceRepository) ActionInstance(dto ReadInstanceDTO) (*EndorServiceActionDictionary, error) {
	idSegments := strings.Split(dto.Id, "/")
	if len(idSegments) == 2 {
		resourceInstance, err := h.Instance(ReadInstanceDTO{
			Id: idSegments[0],
		})
		if err != nil {
			return nil, err
		}
		if resourceAction, ok := resourceInstance.EndorService.Methods[idSegments[1]]; ok {
			return h.createAction(idSegments[0], idSegments[1], resourceAction)
		} else {
			return nil, NewNotFoundError(fmt.Errorf("resource action not found"))
		}
	} else {
		return nil, NewBadRequestError(fmt.Errorf("invalid resource action id"))
	}
}

func (h *EndorServiceRepository) Create(dto CreateDTO[Resource]) error {
	dto.Data.Service = h.microServiceId
	_, err := h.Instance(ReadInstanceDTO{
		Id: dto.Data.ID,
	})
	var endorError *EndorError
	if errors.As(err, &endorError) && endorError.StatusCode == 404 {
		_, err := h.collection.InsertOne(h.context, dto.Data)
		if err != nil {
			return err
		}
		h.reloadRouteConfiguration(h.microServiceId)
		return nil
	} else {
		return NewConflictError(fmt.Errorf("resource already exist"))
	}
}

func (h *EndorServiceRepository) UpdateOne(dto UpdateByIdDTO[Resource]) (*Resource, error) {
	var instance *Resource
	_, err := h.Instance(ReadInstanceDTO{
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

func (h *EndorServiceRepository) DeleteOne(dto ReadInstanceDTO) error {
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
	config := GetConfig()
	resources, err := h.EndorServiceList()
	if err != nil {
		return err
	}
	err = InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), resources)
	if err != nil {
		return err
	}
	_, err = CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), resources, "/api")
	if err != nil {
		return err
	}
	return nil
}

func (h *EndorServiceRepository) createAction(resourceName string, actionName string, endorServiceAction EndorServiceAction) (*EndorServiceActionDictionary, error) {
	actionId := path.Join(resourceName, actionName)
	action := ResourceAction{
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
