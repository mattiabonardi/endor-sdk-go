package sdk

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoResourceInstanceRepository[T ResourceInstanceInterface] struct {
	collection *mongo.Collection
}

func NewMongoResourceInstanceRepository[T ResourceInstanceInterface](resourceId string) *MongoResourceInstanceRepository[T] {
	client, _ := GetMongoClient()

	collection := client.Database(GetConfig().EndorDynamicResourceDBName).Collection(resourceId)
	return &MongoResourceInstanceRepository[T]{
		collection: collection,
	}
}

func (r *MongoResourceInstanceRepository[T]) Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstance[T], error) {
	var result ResourceInstance[T]
	err := r.collection.FindOne(ctx, bson.M{"_id": dto.Id}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
		}
		return nil, NewInternalServerError(fmt.Errorf("failed to find resource instance: %w", err))
	}
	return &result, nil
}

func (r *MongoResourceInstanceRepository[T]) List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error) {
	var opts *options.FindOptions = options.Find().SetProjection(dto.Projection)
	cursor, err := r.collection.Find(ctx, dto.Filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []ResourceInstance[T]
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *MongoResourceInstanceRepository[T]) Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	// Auto-generate string ID if empty
	if idPtr := dto.Data.This.GetID(); idPtr != nil && *idPtr == "" {
		newID := primitive.NewObjectID().Hex()

		// Use reflection to set the ID field generically
		thisValue := reflect.ValueOf(&dto.Data.This).Elem()
		if thisValue.Kind() == reflect.Struct {
			// Look for a field named "Id" of type string
			idField := thisValue.FieldByName("Id")
			if idField.IsValid() && idField.CanSet() && idField.Type() == reflect.TypeOf("") {
				idField.Set(reflect.ValueOf(newID))
			}
		}
	}

	// Check if resource instance already exists (only if ID is not empty)
	if idPtr := dto.Data.This.GetID(); idPtr != nil && *idPtr != "" {
		_, err := r.Instance(ctx, ReadInstanceDTO{Id: *idPtr})
		if err == nil {
			return nil, NewConflictError(fmt.Errorf("resource instance with id %v already exists", *idPtr))
		}
		// If it's a NotFoundError, that's good - we can proceed with creation
		var endorErr *EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			// If it's any other error besides NotFound, return it
			return nil, err
		}
	}

	_, err := r.collection.InsertOne(ctx, dto.Data)
	if err != nil {
		// Handle MongoDB duplicate key error
		if mongo.IsDuplicateKeyError(err) {
			return nil, NewConflictError(fmt.Errorf("resource instance already exists: %w", err))
		}
		return nil, NewInternalServerError(fmt.Errorf("failed to create resource instance: %w", err))
	}
	return &dto.Data, nil
}

func (r *MongoResourceInstanceRepository[T]) Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	// First, check if the resource instance exists
	_, err := r.Instance(ctx, ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return nil, err // This will be a NotFoundError if the instance doesn't exist
	}

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": dto.Id}, dto.Data)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to update resource instance: %w", err))
	}

	if result.ModifiedCount == 0 && result.MatchedCount == 0 {
		return nil, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	return &dto.Data, nil
}

func (r *MongoResourceInstanceRepository[T]) Delete(ctx context.Context, dto DeleteByIdDTO) error {
	// First, check if the resource instance exists
	_, err := r.Instance(ctx, ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return err // This will be a NotFoundError if the instance doesn't exist
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": dto.Id})
	if err != nil {
		return NewInternalServerError(fmt.Errorf("failed to delete resource instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	return nil
}
