package sdk

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewResourceRepository(microServiceId string, services []EndorService, client *mongo.Client, context context.Context, databaseName string) *ResourceRepository {
	database := client.Database(databaseName)
	collection := database.Collection(COLLECTION_RESOURCES)
	return &ResourceRepository{
		microServiceId: microServiceId,
		services:       services,
		collection:     collection,
		context:        context,
	}
}

type ResourceRepository struct {
	microServiceId string
	services       []EndorService
	context        context.Context
	collection     *mongo.Collection
}

func (h *ResourceRepository) List() ([]Resource, error) {
	resources := []Resource{}
	// search from internal services
	for _, service := range h.services {
		resource := Resource{
			ID:          service.Resource,
			Description: service.Description,
			Service:     h.microServiceId,
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
		resources = append(resources, resource)
	}
	storedResources, err := h.DynamiResourceList()
	if err != nil {
		return []Resource{}, nil
	}
	resources = append(resources, storedResources...)
	return resources, nil
}

func (h *ResourceRepository) DynamiResourceList() ([]Resource, error) {
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

func (h *ResourceRepository) Instance(dto ReadInstanceDTO) (*Resource, error) {
	// search from internal services
	for _, service := range h.services {
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
			return &resource, nil
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
	return &resource, nil
}

func (h *ResourceRepository) Create(dto CreateDTO[Resource]) error {
	dto.Data.Service = h.microServiceId
	_, err := h.Instance(ReadInstanceDTO{
		Id: dto.Data.ID,
	})
	if errors.Is(err, ErrNotFound) {
		_, err := h.collection.InsertOne(h.context, dto.Data)
		if err != nil {
			return ErrInternalServerError
		}
		return nil
	} else {
		return ErrAlreadyExists
	}
}

func (h *ResourceRepository) UpdateOne(dto UpdateByIdDTO[Resource]) (*Resource, error) {
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

	return &dto.Data, nil
}

func (h *ResourceRepository) DeleteOne(dto DeleteByIdDTO) error {
	// check if resources already exist
	_, err := h.Instance(ReadInstanceDTO(dto))
	if err != nil {
		return err
	}
	_, err = h.collection.DeleteOne(h.context, bson.M{"_id": dto.Id})
	return err
}
