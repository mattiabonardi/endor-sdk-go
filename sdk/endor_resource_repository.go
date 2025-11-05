package sdk

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewEndorServiceRepository(microServiceId string, internalEndorServices *[]EndorService, databaseName string) *EndorServiceRepository {
	client, _ := GetMongoClient()
	serviceRepository := &EndorServiceRepository{
		microServiceId:        microServiceId,
		internalEndorServices: internalEndorServices,
		context:               context.TODO(),
	}
	if GetConfig().EndorDynamicResourcesEnabled {
		database := client.Database(databaseName)
		serviceRepository.collection = database.Collection(COLLECTION_RESOURCES)
	}
	return serviceRepository
}

type EndorServiceRepository struct {
	microServiceId        string
	internalEndorServices *[]EndorService
	collection            *mongo.Collection
	context               context.Context
}

type EndorServiceDictionary struct {
	EndorService EndorService
	resource     Resource
}

type EndorServiceActionDictionary struct {
	EndorServiceAction EndorServiceAction
	resourceAction     ResourceAction
}

func (h *EndorServiceRepository) Map() (map[string]EndorServiceDictionary, error) {
	resources := map[string]EndorServiceDictionary{}
	// internal
	for _, internalEndorService := range *h.internalEndorServices {
		resource := Resource{
			ID:          internalEndorService.Resource,
			Description: internalEndorService.Description,
			Service:     h.microServiceId,
		}
		for methodName, method := range internalEndorService.Methods {
			if methodName == "create" && method.GetOptions().InputSchema != nil {
				definition := ResourceDefinition{
					Schema: *method.GetOptions().InputSchema,
				}
				stringDefinition, _ := definition.ToYAML()
				resource.Definition = stringDefinition
			}
		}
		resources[internalEndorService.Resource] = EndorServiceDictionary{
			EndorService: internalEndorService,
			resource:     resource,
		}
	}
	if GetConfig().EndorDynamicResourcesEnabled {
		// dynamic
		dynamicResources, err := h.DynamiResourceList()
		if err != nil {
			return map[string]EndorServiceDictionary{}, nil
		}
		// create endor resource
		for _, resource := range dynamicResources {
			defintion, err := resource.UnmarshalDefinition()
			if err == nil {
				resources[resource.ID] = EndorServiceDictionary{
					EndorService: NewAbstractResourceService(resource.ID, resource.Description, *defintion),
					resource:     resource,
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
			for methodName, method := range service.Methods {
				if methodName == "create" && method.GetOptions().InputSchema != nil {
					definition := ResourceDefinition{
						Schema: *method.GetOptions().InputSchema,
					}
					stringDefinition, _ := definition.ToYAML()
					resource.Definition = stringDefinition
				}
			}
			return &EndorServiceDictionary{
				EndorService: service,
				resource:     resource,
			}, nil
		}
	}
	if GetConfig().EndorDynamicResourcesEnabled {
		// search from database
		resource := Resource{}
		filter := bson.M{"_id": dto.Id}
		err := h.collection.FindOne(h.context, filter).Decode(&resource)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, NewNotFoundError(fmt.Errorf("resourse not found"))
			} else {
				return nil, err
			}
		}
		defintion, err := resource.UnmarshalDefinition()
		if err != nil {
			return nil, err
		}
		return &EndorServiceDictionary{
			EndorService: NewAbstractResourceService(resource.ID, resource.Description, *defintion),
			resource:     resource,
		}, nil
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
		Id: dto.Data.ID,
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

func (h *EndorServiceRepository) DeleteOne(dto DeleteByIdDTO) error {
	// check if resources already exist
	_, err := h.Instance(ReadInstanceDTO(dto))
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
