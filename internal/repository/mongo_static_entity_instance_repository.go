package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStaticEntityInstanceRepository provides MongoDB implementation for StaticEntityInstanceRepositoryInterface.
// This implementation works directly with the model type T without the EntityInstance[T] wrapper,
// offering a simpler interface for cases where metadata functionality is not required.
//
// ObjectID Handling Strategy:
// The ID storage type and all sdk.ObjectID fields are automatically handled:
//
//  1. Detection: detectIDType inspects the model's _id field type at initialization
//  2. Storage: All sdk.ObjectID fields are stored as primitive.ObjectID in MongoDB
//  3. Conversion: convertObjectIDsToStorage ensures proper conversion in filters and updates
//  4. BSON Marshaling: sdk.ObjectID implements MarshalBSONValue/UnmarshalBSONValue for
//     automatic conversion during struct serialization
//
// This provides transparent ObjectID handling across all repository operations.
type MongoStaticEntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	collection *mongo.Collection
	options    sdk.StaticEntityInstanceRepositoryOptions
	idType     string // "string" or "objectid"
}

// NewMongoStaticEntityInstanceRepository creates a new MongoDB-based static repository
// The ID storage type is automatically detected from the model's ID field type
func NewMongoStaticEntityInstanceRepository[T sdk.EntityInstanceInterface](entityId string, options sdk.StaticEntityInstanceRepositoryOptions) *MongoStaticEntityInstanceRepository[T] {
	client, _ := sdk.GetMongoClient()

	collection := client.Database(sdk_configuration.GetConfig().DynamicEntityDocumentDBName).Collection(entityId)

	// Detect ID type from the model structure
	idType := detectIDType[T]()

	return &MongoStaticEntityInstanceRepository[T]{
		collection: collection,
		options:    options,
		idType:     idType,
	}
}

func (r *MongoStaticEntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (T, error) {
	var zero T

	// Prepare filter based on detected ID type
	var filter bson.M
	if r.idType == "objectid" {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return zero, sdk.NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	// Decode directly into the model - ObjectID fields are automatically converted via UnmarshalBSONValue
	var result T
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return zero, sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
		}
		return zero, sdk.NewInternalServerError(fmt.Errorf("failed to find entity instance: %w", err))
	}

	return result, nil
}

func (r *MongoStaticEntityInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]T, error) {
	// Use filter and projection directly from DTO
	mongoFilter := dto.Filter
	if mongoFilter == nil {
		mongoFilter = bson.M{}
	} else {
		// Clone the filter to avoid modifying the input
		mongoFilter = cloneBsonM(mongoFilter)
	}

	// Convert ObjectID fields in filter to primitive.ObjectID
	objectIDFields := getObjectIDFields[T]()
	if err := convertObjectIDsToStorage(mongoFilter, objectIDFields); err != nil {
		return nil, sdk.NewBadRequestError(err)
	}

	var opts *options.FindOptions
	if dto.Projection != nil {
		opts = options.Find().SetProjection(dto.Projection)
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to list entities: %w", err))
	}
	defer cursor.Close(ctx)

	// Decode directly into models - ObjectID fields are automatically converted via UnmarshalBSONValue
	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to decode entities: %w", err))
	}

	return results, nil
}

func (r *MongoStaticEntityInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[T]) (T, error) {
	var zero T
	idPtr := dto.Data.GetID()
	var idStr string

	if *r.options.AutoGenerateID {
		// Generate ID based on detected type
		oid := primitive.NewObjectID()
		idStr = oid.Hex()

		// Serialize the struct - ObjectID fields are automatically converted via MarshalBSONValue
		docBytes, err := bson.Marshal(dto.Data)
		if err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to marshal entity: %w", err))
		}
		var doc bson.M
		if err := bson.Unmarshal(docBytes, &doc); err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to unmarshal entity: %w", err))
		}

		// Set _id based on detected type
		if r.idType == "objectid" {
			// For ObjectID type, store as primitive.ObjectID
			doc["_id"] = oid
		} else {
			// For string type, store as hex string
			doc["_id"] = idStr
		}

		_, err = r.collection.InsertOne(ctx, doc)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return zero, sdk.NewConflictError(fmt.Errorf("entity instance already exists: %w", err))
			}
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to create entity instance: %w", err))
		}
	} else {
		if isIDEmpty(idPtr) {
			return zero, sdk.NewBadRequestError(fmt.Errorf("ID is required when auto-generation is disabled"))
		}
		idStr = idToString(idPtr)

		// Verify that the ID doesn't already exist
		_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: idStr})
		if err == nil {
			return zero, sdk.NewConflictError(fmt.Errorf("entity instance with id %v already exists", idStr))
		}
		var endorErr *sdk.EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			return zero, err
		}

		// Serialize the struct - ObjectID fields are automatically converted via MarshalBSONValue
		docBytes, err := bson.Marshal(dto.Data)
		if err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to marshal entity: %w", err))
		}
		var doc bson.M
		if err := bson.Unmarshal(docBytes, &doc); err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to unmarshal entity: %w", err))
		}

		// Set _id based on detected type and ensure all ObjectID fields are properly stored
		if r.idType == "objectid" {
			oid, err := primitive.ObjectIDFromHex(idStr)
			if err != nil {
				return zero, sdk.NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
			}
			doc["_id"] = oid
		} else {
			doc["_id"] = idStr
		}

		// Insert the document
		_, err = r.collection.InsertOne(ctx, doc)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return zero, sdk.NewConflictError(fmt.Errorf("entity instance already exists: %w", err))
			}
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to create entity instance: %w", err))
		}
	}

	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: idStr})
}

func (r *MongoStaticEntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	// Verify that the instance exists
	_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return err
	}

	// Prepare filter based on detected ID type
	var filter bson.M
	if r.idType == "objectid" {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return sdk.NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	// Delete the document
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return sdk.NewInternalServerError(fmt.Errorf("failed to delete entity instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
	}

	return nil
}

func (r *MongoStaticEntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[map[string]interface{}]) (T, error) {
	var zero T

	// Verify the instance exists
	_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return zero, err
	}

	// Prepare filter based on detected ID type
	var filter bson.M
	if r.idType == "objectid" {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return zero, sdk.NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	// If no fields to update, return error
	if len(dto.Data) == 0 {
		return zero, sdk.NewBadRequestError(fmt.Errorf("no fields to update"))
	}

	// Clone update data to avoid modifying input and convert ObjectID fields
	updateData := cloneBsonM(dto.Data)
	objectIDFields := getObjectIDFields[T]()
	if err := convertObjectIDsToStorage(updateData, objectIDFields); err != nil {
		return zero, sdk.NewBadRequestError(err)
	}

	// Perform the update with $set
	update := bson.M{"$set": updateData}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return zero, sdk.NewInternalServerError(fmt.Errorf("failed to update entity instance: %w", err))
	}
	if result.MatchedCount == 0 {
		return zero, sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
	}

	// Retrieve and return the updated document
	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
}
