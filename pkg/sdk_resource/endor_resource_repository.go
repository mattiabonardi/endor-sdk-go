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
	resource         sdk.Resource
}

type EndorServiceActionDictionary struct {
	EndorServiceAction sdk.EndorServiceAction
	resourceAction     sdk.ResourceAction
}

// check if resource document exist in resource collection for EndorHybridServiceInterface
func (h *EndorServiceRepository) ensureHybridServiceDocument(hybridService sdk.EndorHybridServiceInterface) {
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

	// internal EndorServices
	if h.internalEndorServices != nil {
		for _, internalEndorService := range *h.internalEndorServices {
			resource := sdk.Resource{
				ID:          internalEndorService.GetResource(),
				Description: internalEndorService.GetResourceDescription(),
				Service:     h.microServiceId,
			}

			// create document in resource collection for EndorHybridServiceInterface and EndorHybridSpecilizedServiceInterface
			if hybrid, ok := internalEndorService.(sdk.EndorHybridServiceInterface); ok {
				h.ensureHybridServiceDocument(hybrid)
			}

			resources[internalEndorService.GetResource()] = EndorServiceDictionary{
				OriginalInstance: &internalEndorService,
				EndorService:     internalEndorService.ToEndorService(sdk.Schema{}),
				resource:         resource,
			}
		}
	}

	// dynamic EndorServices
	if configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled {
		dynamicResources, err := h.DynamiResourceList()
		if err != nil {
			return map[string]EndorServiceDictionary{}, nil
		}

		for _, resource := range dynamicResources {
			defintion, err := resource.UnmarshalAdditionalAttributes()
			if err != nil {
				// TODO: log the error but continue
				continue
			}

			// check if service is already defined
			// search service
			if v, ok := resources[resource.ID]; ok {
				//TODO: supports dynamic specialized services

				// if found check if has EndorHybridServiceInterface
				if _, ok := (*v.OriginalInstance).(sdk.EndorHybridServiceInterface); ok {
					// inject dynamic schema
					v.EndorService = (*v.OriginalInstance).ToEndorService(defintion.Schema)
					resources[resource.ID] = v
				} else {
					//TODO: log that only hybrid service supports additional attributes
				}
			} else {
				// create a new hybrid service
				hybridService := NewHybridService[*sdk.DynamicResource](resource.ID, resource.Description)
				resources[resource.ID] = EndorServiceDictionary{
					EndorService: hybridService.ToEndorService(defintion.Schema),
					resource:     resource,
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
