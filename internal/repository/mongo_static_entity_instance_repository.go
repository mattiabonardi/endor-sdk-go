package repository

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"go.mongodb.org/mongo-driver/bson"
)

// MongoStaticEntityInstanceRepository handles entities defined entirely at compile time.
//
// Document structure in MongoDB:
//
//	{
//	    "_id": "...",
//	    "field1": "...",      // All model fields at root level
//	    "field2": "...",
//	}
//
// Unlike MongoEntityInstanceRepository, this repository:
// - Does NOT have a metadata field
// - Stores entities directly as they are defined in the struct
// - Still automatically converts sdk.ObjectID fields to primitive.ObjectID
type MongoStaticEntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	base *mongoBaseRepository[T]
}

// NewMongoStaticEntityInstanceRepository creates a new repository for the given entity.
func NewMongoStaticEntityInstanceRepository[T sdk.EntityInstanceInterface](
	entityId string,
	options sdk.StaticEntityInstanceRepositoryOptions,
) *MongoStaticEntityInstanceRepository[T] {
	client, err := sdk.GetMongoClient()
	if client == nil || err != nil {
		// Return a repository with nil base - operations will fail at runtime
		// This allows the service to be constructed without a DB connection (useful for tests)
		return &MongoStaticEntityInstanceRepository[T]{
			base: nil,
		}
	}
	collection := client.Database(sdk_configuration.GetConfig().DynamicEntityDocumentDBName).Collection(entityId)

	return &MongoStaticEntityInstanceRepository[T]{
		base: newMongoBaseRepository[T](collection, *options.AutoGenerateID),
	}
}

// Instance retrieves a single entity by ID.
func (r *MongoStaticEntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (T, error) {
	var zero T

	rawDoc, err := r.base.FindByID(ctx, dto.Id)
	if err != nil {
		return zero, err
	}

	return r.toModel(rawDoc)
}

// List retrieves entities matching the filter.
func (r *MongoStaticEntityInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]T, error) {
	// For static entities, filter and projection are used directly (no metadata separation)
	var filter bson.M
	if dto.Filter != nil {
		filter = cloneBsonM(dto.Filter)
	}

	var projection bson.M
	if dto.Projection != nil {
		projection = cloneBsonM(dto.Projection)
	}

	rawDocs, err := r.base.Find(ctx, filter, projection)
	if err != nil {
		return nil, err
	}

	results := make([]T, 0, len(rawDocs))
	for _, rawDoc := range rawDocs {
		model, err := r.toModel(rawDoc)
		if err != nil {
			return nil, err
		}
		results = append(results, model)
	}

	return results, nil
}

// Create inserts a new entity.
func (r *MongoStaticEntityInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[T]) (T, error) {
	var zero T

	mapper := r.base.GetDocumentMapper()
	doc, err := mapper.ToDocumentWithoutMetadata(dto.Data, r.base.GetIDStrategy())
	if err != nil {
		return zero, sdk.NewInternalServerError(err)
	}

	// Get provided ID
	providedID := dto.Data.GetID()

	idStr, err := r.base.Insert(ctx, doc, providedID)
	if err != nil {
		return zero, err
	}

	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: idStr})
}

// Update modifies an existing entity by ID.
func (r *MongoStaticEntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[map[string]interface{}]) (T, error) {
	var zero T

	if err := r.base.Update(ctx, dto.Id, dto.Data); err != nil {
		return zero, err
	}

	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
}

// Delete removes an entity by ID.
func (r *MongoStaticEntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.base.Delete(ctx, dto.Id)
}

// toModel converts a raw MongoDB document to the model type T.
func (r *MongoStaticEntityInstanceRepository[T]) toModel(rawDoc bson.M) (T, error) {
	var zero T

	entityBytes, err := bson.Marshal(rawDoc)
	if err != nil {
		return zero, sdk.NewInternalServerError(fmt.Errorf("failed to marshal entity: %w", err))
	}

	var result T
	if err := bson.Unmarshal(entityBytes, &result); err != nil {
		return zero, sdk.NewInternalServerError(fmt.Errorf("failed to unmarshal entity: %w", err))
	}

	return result, nil
}
