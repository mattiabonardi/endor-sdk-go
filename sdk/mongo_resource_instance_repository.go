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
	options    ResourceInstanceRepositoryOptions
}

func NewMongoResourceInstanceRepository[T ResourceInstanceInterface](resourceId string, options ResourceInstanceRepositoryOptions) *MongoResourceInstanceRepository[T] {
	client, _ := GetMongoClient()

	collection := client.Database(GetConfig().EndorDynamicResourceDBName).Collection(resourceId)
	return &MongoResourceInstanceRepository[T]{
		collection: collection,
		options:    options,
	}
}

func (r *MongoResourceInstanceRepository[T]) Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstance[T], error) {
	var result ResourceInstance[T]

	// Convert string ID to ObjectID for MongoDB query if auto-generation is enabled
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return nil, NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
		}
		return nil, NewInternalServerError(fmt.Errorf("failed to find resource instance: %w", err))
	}

	// Convert ObjectID back to string in the result if using auto-generation
	if *r.options.AutoGenerateID {
		thisValue := reflect.ValueOf(&result.This).Elem()
		if thisValue.Kind() == reflect.Struct {
			idField := thisValue.FieldByName("Id")
			if idField.IsValid() && idField.CanSet() && idField.Type() == reflect.TypeOf("") {
				// Set the string representation of the ID
				idField.Set(reflect.ValueOf(dto.Id))
			}
		}
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

	// Convert ObjectID back to string in each result if using auto-generation
	if *r.options.AutoGenerateID {
		for i := range results {
			thisValue := reflect.ValueOf(&results[i].This).Elem()
			if thisValue.Kind() == reflect.Struct {
				idField := thisValue.FieldByName("Id")
				if idField.IsValid() && idField.CanSet() && idField.Type() == reflect.TypeOf("") {
					// Extract ObjectID from the database result and convert to string
					if objectIDPtr := results[i].This.GetID(); objectIDPtr != nil && *objectIDPtr != "" {
						// The ID should already be converted, but ensure it's a string
						idField.Set(reflect.ValueOf(*objectIDPtr))
					}
				}
			}
		}
	}

	return results, nil
}

func (r *MongoResourceInstanceRepository[T]) Create(
	ctx context.Context,
	dto CreateDTO[ResourceInstance[T]],
) (*ResourceInstance[T], error) {

	idPtr := dto.Data.This.GetID()
	var idStr string

	if *r.options.AutoGenerateID {
		// Generazione automatica dell’ObjectID
		oid := primitive.NewObjectID()
		idStr = oid.Hex()
		dto.Data.This.SetID(idStr)
	} else {
		// L'ID deve essere fornito
		if idPtr == nil || *idPtr == "" {
			return nil, NewBadRequestError(fmt.Errorf("ID is required when auto-generation is disabled"))
		}
		idStr = *idPtr

		// Verifica che non esista già
		_, err := r.Instance(ctx, ReadInstanceDTO{Id: idStr})
		if err == nil {
			return nil, NewConflictError(fmt.Errorf("resource instance with id %v already exists", idStr))
		}
		var endorErr *EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			return nil, err
		}
	}

	// Marshal del modello T in mappa BSON
	resourceBytes, err := bson.Marshal(dto.Data.This)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to marshal resource: %w", err))
	}
	var resourceMap bson.M
	if err := bson.Unmarshal(resourceBytes, &resourceMap); err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to unmarshal resource: %w", err))
	}

	// Imposta l’_id corretto
	if *r.options.AutoGenerateID {
		oid, _ := primitive.ObjectIDFromHex(idStr)
		resourceMap["_id"] = oid
	} else {
		resourceMap["_id"] = idStr
	}

	// Aggiunge i metadata al documento
	for k, v := range dto.Data.Metadata {
		resourceMap[k] = v
	}

	// Inserisci nel DB
	_, err = r.collection.InsertOne(ctx, resourceMap)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, NewConflictError(fmt.Errorf("resource instance already exists: %w", err))
		}
		return nil, NewInternalServerError(fmt.Errorf("failed to create resource instance: %w", err))
	}

	// Restituisce il modello con l’ID string impostato
	return &dto.Data, nil
}

func (r *MongoResourceInstanceRepository[T]) Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	// First, check if the resource instance exists
	_, err := r.Instance(ctx, ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return nil, err // This will be a NotFoundError if the instance doesn't exist
	}

	// Prepare filter and update document based on auto-generation setting
	var filter bson.M
	var updateDoc interface{}

	if *r.options.AutoGenerateID {
		// Convert string ID to ObjectID for MongoDB query
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return nil, NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}

		// Prepare update document with ObjectID
		resourceBytes, err := bson.Marshal(dto.Data.This)
		if err != nil {
			return nil, NewInternalServerError(fmt.Errorf("failed to marshal resource: %w", err))
		}

		var resourceMap bson.M
		if err := bson.Unmarshal(resourceBytes, &resourceMap); err != nil {
			return nil, NewInternalServerError(fmt.Errorf("failed to unmarshal resource: %w", err))
		}

		// Replace _id with ObjectID
		resourceMap["_id"] = objectID

		// Add metadata to the document
		mongoDoc := bson.M{}
		for k, v := range resourceMap {
			mongoDoc[k] = v
		}
		for k, v := range dto.Data.Metadata {
			mongoDoc[k] = v
		}

		updateDoc = mongoDoc
	} else {
		// Use string ID as-is
		filter = bson.M{"_id": dto.Id}
		updateDoc = dto.Data
	}

	result, err := r.collection.ReplaceOne(ctx, filter, updateDoc)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to update resource instance: %w", err))
	}

	if result.ModifiedCount == 0 && result.MatchedCount == 0 {
		return nil, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	// Return the updated data with correct string ID
	updatedData := dto.Data
	if idPtr := updatedData.This.GetID(); idPtr != nil {
		// Ensure the returned object has the correct string ID
		thisValue := reflect.ValueOf(&updatedData.This).Elem()
		if thisValue.Kind() == reflect.Struct {
			idField := thisValue.FieldByName("Id")
			if idField.IsValid() && idField.CanSet() && idField.Type() == reflect.TypeOf("") {
				idField.Set(reflect.ValueOf(dto.Id))
			}
		}
	}

	return &updatedData, nil
}

func (r *MongoResourceInstanceRepository[T]) Delete(ctx context.Context, dto ReadInstanceDTO) error {
	// First, check if the resource instance exists
	_, err := r.Instance(ctx, ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return err // This will be a NotFoundError if the instance doesn't exist
	}

	// Prepare filter based on auto-generation setting
	var filter bson.M
	if *r.options.AutoGenerateID {
		// Convert string ID to ObjectID for MongoDB query
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		// Use string ID as-is
		filter = bson.M{"_id": dto.Id}
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return NewInternalServerError(fmt.Errorf("failed to delete resource instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	return nil
}
