package sdk

import (
	"context"
	"errors"

	"github.com/mattiabonardi/endor-sdk-go/sdk/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewResourceRepository(services []EndorService, client *mongo.Client, context context.Context, databaseName string) *ResourceRepository {
	database := client.Database(databaseName)
	collection := database.Collection(COLLECTION_RESOURCES)
	return &ResourceRepository{
		services:   services,
		collection: collection,
		context:    context,
	}
}

type ResourceRepository struct {
	services   []EndorService
	context    context.Context
	collection *mongo.Collection
}

func (h *ResourceRepository) List(options ResourceListDTO) ([]Resource, error) {
	resources := []Resource{}
	// search from internal services
	for _, service := range h.services {
		if len(service.Apps) == 0 || utils.StringElemMatch(service.Apps, options.App) {
			resource := Resource{
				ID:          service.Resource,
				Description: service.Description,
			}
			for methodName, method := range service.Methods {
				payload, _ := resolvePayloadType(method)
				requestSchema := NewSchemaByType(payload)
				if methodName == "create" {
					stringSchema, _ := requestSchema.ToYAML()
					resource.Schema = stringSchema
				}
			}
			resources = append(resources, resource)
		}
	}
	// search from database
	filter := bson.M{"apps": options.App}
	cursor, err := h.collection.Find(h.context, filter)
	if err != nil {
		return nil, ErrInternalServerError
	}
	var storedResources []Resource
	if err := cursor.All(h.context, &storedResources); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return resources, nil
		} else {
			return nil, ErrInternalServerError
		}
	}
	resources = append(resources, storedResources...)
	return resources, nil
}

func (h *ResourceRepository) Instance(options ResourceInstanceDTO) (*Resource, error) {
	// search from internal services
	for _, service := range h.services {
		if service.Resource == options.Id {
			if len(service.Apps) == 0 || utils.StringElemMatch(service.Apps, options.App) {
				resource := Resource{
					ID:          service.Resource,
					Description: service.Description,
				}
				for methodName, method := range service.Methods {
					payload, _ := resolvePayloadType(method)
					requestSchema := NewSchemaByType(payload)
					if methodName == "create" {
						stringSchema, _ := requestSchema.ToYAML()
						resource.Schema = stringSchema
					}
				}
				return &resource, nil
			}
		}
	}
	// search from database
	filter := bson.M{"_id": options.Id, "apps": options.App}
	resource := Resource{}
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
	// check if resources already exist
	for _, app := range dto.Data.Apps {
		instanceOptions := ResourceInstanceDTO{
			App: app,
			Id:  dto.Data.ID,
		}
		instance, err := h.Instance(instanceOptions)
		if err != nil && errors.Is(err, ErrInternalServerError) {
			return ErrInternalServerError
		}
		if instance != nil {
			return ErrAlreadyExists
		}
	}
	_, err := h.collection.InsertOne(h.context, dto.Data)
	if err != nil {
		return ErrInternalServerError
	}
	return nil
}

func (h *ResourceRepository) UpdateByID(dto ResourceUpdateByIdDTO) (*Resource, error) {
	// check if resources already exist
	for _, app := range dto.Data.Apps {
		instanceOptions := ResourceInstanceDTO{
			App: app,
			Id:  dto.Data.ID,
		}
		_, err := h.Instance(instanceOptions)
		if err != nil {
			return &dto.Data, err
		}
	}

	updateBson, err := bson.Marshal(dto.Data)
	if err != nil {
		return &dto.Data, err
	}
	update := bson.M{"$set": bson.Raw(updateBson)}
	filter := bson.M{"_id": dto.Id, "apps": dto.App}
	_, err = h.collection.UpdateOne(h.context, filter, update)
	if err != nil {
		return nil, ErrInternalServerError
	}

	return &dto.Data, nil
}
