package repository

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"go.mongodb.org/mongo-driver/bson"
)

// MongoEntityInstanceRepository handles entities with compile-time fields + runtime metadata.
//
// Document structure in MongoDB:
//
//	{
//	    "_id": "...",
//	    "field1": "...",      // Model fields at root level
//	    "field2": "...",
//	    "metadata": {         // Runtime metadata in nested object
//	        "metaField1": "...",
//	        "metaField2": "..."
//	    }
//	}
//
// The repository automatically:
// - Separates model fields from metadata when reading/writing
// - Converts sdk.ObjectID fields to primitive.ObjectID in MongoDB
// - Handles embedded structs with bson:",inline" tags
type MongoEntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	base        *mongoBaseRepository[T]
	modelFields *ModelFieldRegistry
}

// NewMongoEntityInstanceRepository creates a new repository for the given collection.
func NewMongoEntityInstanceRepository[T sdk.EntityInstanceInterface](
	collectionName string,
	options sdk.EntityInstanceRepositoryOptions,
) *MongoEntityInstanceRepository[T] {
	client, err := sdk.GetMongoClient()
	if client == nil || err != nil {
		return &MongoEntityInstanceRepository[T]{
			base:        nil,
			modelFields: NewModelFieldRegistry[T](),
		}
	}
	collection := client.Database(sdk_configuration.GetConfig().DynamicEntityDocumentDBName).Collection(collectionName)

	return &MongoEntityInstanceRepository[T]{
		base:        newMongoBaseRepository[T](collection, *options.AutoGenerateID),
		modelFields: NewModelFieldRegistry[T](),
	}
}

// Instance retrieves a single entity by ID.
func (r *MongoEntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.EntityInstance[T], error) {
	rawDoc, err := r.base.FindByID(ctx, dto.Id)
	if err != nil {
		return nil, err
	}

	return r.toEntityInstance(rawDoc)
}

// List retrieves entities matching the filter.
func (r *MongoEntityInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]sdk.EntityInstance[T], error) {
	// Prepare filter and projection - prefix metadata fields appropriately
	filter := r.modelFields.PrepareFilter(dto.Filter)
	projection := r.modelFields.PrepareProjection(dto.Projection)

	rawDocs, err := r.base.Find(ctx, filter, projection)
	if err != nil {
		return nil, err
	}

	results := make([]sdk.EntityInstance[T], 0, len(rawDocs))
	for _, rawDoc := range rawDocs {
		instance, err := r.toEntityInstance(rawDoc)
		if err != nil {
			return nil, err
		}
		results = append(results, *instance)
	}

	return results, nil
}

// Create inserts a new entity.
func (r *MongoEntityInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[sdk.EntityInstance[T]]) (*sdk.EntityInstance[T], error) {
	mapper := r.base.GetDocumentMapper()

	doc, err := mapper.ToDocument(dto.Data.This, dto.Data.Metadata, r.base.GetIDStrategy())
	if err != nil {
		return nil, sdk.NewInternalServerError(err)
	}

	// Get provided ID (if any)
	var providedID any
	idPtr := dto.Data.This.GetID()
	if !isIDEmpty(idPtr) {
		providedID = idToString(idPtr)
	}

	idStr, err := r.base.Insert(ctx, doc, providedID)
	if err != nil {
		return nil, err
	}

	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: idStr})
}

// Update modifies an existing entity by ID.
func (r *MongoEntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.PartialEntityInstance[T]]) (*sdk.EntityInstance[T], error) {
	// Build the $set document with model fields at root and metadata fields prefixed
	setDoc := bson.M{}

	// Model fields go at root level
	for k, v := range dto.Data.This {
		setDoc[k] = v
	}

	// Metadata fields are prefixed with "metadata."
	for k, v := range dto.Data.Metadata {
		setDoc["metadata."+k] = v
	}

	if err := r.base.Update(ctx, dto.Id, setDoc); err != nil {
		return nil, err
	}

	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
}

// Delete removes an entity by ID.
func (r *MongoEntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.base.Delete(ctx, dto.Id)
}

// toEntityInstance converts a raw MongoDB document to EntityInstance[T].
func (r *MongoEntityInstanceRepository[T]) toEntityInstance(rawDoc bson.M) (*sdk.EntityInstance[T], error) {
	mapper := r.base.GetDocumentMapper()

	metadata, err := mapper.ExtractMetadata(rawDoc)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to extract metadata: %w", err))
	}

	model, err := mapper.ToModel(rawDoc)
	if err != nil {
		return nil, sdk.NewInternalServerError(err)
	}

	return &sdk.EntityInstance[T]{
		This:     model,
		Metadata: metadata,
	}, nil
}
