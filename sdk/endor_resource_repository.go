package sdk

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewEndorResourceRepository(microServiceId string, internalEndorResources []EndorResource, client *mongo.Client, context context.Context, databaseName string) *EndorResourceRepository {
	database := client.Database(databaseName)
	collection := database.Collection(COLLECTION_RESOURCES)
	return &EndorResourceRepository{
		microServiceId:         microServiceId,
		internalEndorResources: internalEndorResources,
		client:                 client,
		collection:             collection,
		context:                context,
	}
}

type EndorResourceRepository struct {
	microServiceId         string
	internalEndorResources []EndorResource
	context                context.Context
	client                 *mongo.Client
	collection             *mongo.Collection
}

type EndorResourceDictionary struct {
	endorResource EndorResource
	resource      Resource
}

func (h *EndorResourceRepository) Map() (map[string]EndorResourceDictionary, error) {
	resources := map[string]EndorResourceDictionary{}
	// internal
	for _, internalEndorResource := range h.internalEndorResources {
		resource := Resource{
			ID:          internalEndorResource.Resource,
			Description: internalEndorResource.Description,
			Service:     h.microServiceId,
		}
		for methodName, method := range internalEndorResource.Methods {
			payload, _ := resolvePayloadType(method)
			requestSchema := NewSchemaByType(payload)
			if methodName == "create" {
				definition := ResourceDefinition{
					Schema: *requestSchema,
				}
				stringDefinition, _ := definition.ToYAML()
				resource.Definition = stringDefinition
			}
		}
		resources[internalEndorResource.Resource] = EndorResourceDictionary{
			endorResource: internalEndorResource,
			resource:      resource,
		}
	}
	config := LoadConfiguration()
	if config.EndorResourceServiceEnabled {
		// dynamic
		dynamicResources, err := h.DynamiResourceList()
		if err != nil {
			return map[string]EndorResourceDictionary{}, nil
		}
		// create endor resource
		for _, resource := range dynamicResources {
			defintion, err := resource.UnmarshalDefinition()
			if err == nil {
				resources[resource.ID] = EndorResourceDictionary{
					endorResource: NewAbstractResourceService(resource.ID, resource.Description, *defintion, h.client, h.microServiceId, h.context),
					resource:      resource,
				}
			} else {
				// TODO: non blocked log
			}
		}
	}
	return resources, nil
}

func (h *EndorResourceRepository) ResourceList() ([]Resource, error) {
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

func (h *EndorResourceRepository) EndorResourceList() ([]EndorResource, error) {
	resources, err := h.Map()
	if err != nil {
		return []EndorResource{}, err
	}
	resourceList := make([]EndorResource, 0, len(resources))
	for _, service := range resources {
		resourceList = append(resourceList, service.endorResource)
	}
	return resourceList, nil
}

func (h *EndorResourceRepository) DynamiResourceList() ([]Resource, error) {
	cursor, err := h.collection.Find(h.context, bson.M{})
	if err != nil {
		return nil, ErrInternalServerError
	}
	var storedResources []Resource
	if err := cursor.All(h.context, &storedResources); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []Resource{}, nil
		} else {
			return nil, ErrInternalServerError
		}
	}
	return storedResources, nil
}

func (h *EndorResourceRepository) Instance(dto ReadInstanceDTO) (*EndorResourceDictionary, error) {
	// search from internal services
	for _, service := range h.internalEndorResources {
		if service.Resource == dto.Id {
			resource := Resource{
				ID:          service.Resource,
				Description: service.Description,
			}
			for methodName, method := range service.Methods {
				payload, _ := resolvePayloadType(method)
				requestSchema := NewSchemaByType(payload)
				if methodName == "create" {
					definition := ResourceDefinition{
						Schema: *requestSchema,
					}
					stringDefinition, _ := definition.ToYAML()
					resource.Definition = stringDefinition
				}
			}
			return &EndorResourceDictionary{
				endorResource: service,
				resource:      resource,
			}, nil
		}
	}
	// search from database
	resource := Resource{}
	filter := bson.M{"_id": dto.Id}
	err := h.collection.FindOne(h.context, filter).Decode(&resource)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		} else {
			return nil, ErrInternalServerError
		}
	}
	defintion, err := resource.UnmarshalDefinition()
	if err != nil {
		return nil, err
	}
	return &EndorResourceDictionary{
		endorResource: NewAbstractResourceService(resource.ID, resource.Description, *defintion, h.client, h.microServiceId, h.context),
		resource:      resource,
	}, nil
}

func (h *EndorResourceRepository) Create(dto CreateDTO[Resource]) error {
	dto.Data.Service = h.microServiceId
	_, err := h.Instance(ReadInstanceDTO{
		Id: dto.Data.ID,
	})
	if errors.Is(err, ErrNotFound) {
		_, err := h.collection.InsertOne(h.context, dto.Data)
		if err != nil {
			return ErrInternalServerError
		}
		h.reloadRouteConfiguration(h.microServiceId)
		return nil
	} else {
		return ErrAlreadyExists
	}
}

func (h *EndorResourceRepository) UpdateOne(dto UpdateByIdDTO[Resource]) (*Resource, error) {
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
		return nil, ErrInternalServerError
	}

	h.reloadRouteConfiguration(h.microServiceId)

	return &dto.Data, nil
}

func (h *EndorResourceRepository) DeleteOne(dto DeleteByIdDTO) error {
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

func (h *EndorResourceRepository) reloadRouteConfiguration(microserviceId string) error {
	config := LoadConfiguration()
	resources, err := h.EndorResourceList()
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
