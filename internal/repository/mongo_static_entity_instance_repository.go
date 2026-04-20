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
	options  sdk.StaticEntityInstanceRepositoryOptions[T]
	entityId string

	_base *mongoBaseRepository[T]
}

// NewMongoStaticEntityInstanceRepository creates a new repository for the given entity.
func NewMongoStaticEntityInstanceRepository[T sdk.EntityInstanceInterface](
	entityId string,
	options sdk.StaticEntityInstanceRepositoryOptions[T],
) *MongoStaticEntityInstanceRepository[T] {
	return &MongoStaticEntityInstanceRepository[T]{
		options:  options,
		entityId: entityId,
	}
}

func (r *MongoStaticEntityInstanceRepository[T]) GetEntity() string {
	return r.entityId
}

// Instance retrieves a single entity by ID.
func (r *MongoStaticEntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (T, error) {
	var zero T

	rawDoc, err := r.getBaseRepository().FindByID(ctx, dto.Id)
	if err != nil {
		return zero, err
	}

	instance, err := r.toModel(rawDoc)
	if err != nil {
		return zero, err
	}
	if r.options.Hooks.AfterFind != nil {
		err = r.options.Hooks.AfterFind(instance)
		if err != nil {
			return zero, err
		}
	}
	return instance, err
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

	rawDocs, err := r.getBaseRepository().Find(ctx, filter, projection)
	if err != nil {
		return nil, err
	}

	results := make([]T, 0, len(rawDocs))
	for _, rawDoc := range rawDocs {
		model, err := r.toModel(rawDoc)
		if err != nil {
			return nil, err
		}
		// call hook
		if r.options.Hooks.AfterFind != nil {
			err := r.options.Hooks.AfterFind(model)
			if err != nil {
				return nil, err
			}
		}
		results = append(results, model)
	}

	return results, nil
}

// Create inserts a new entity.
func (r *MongoStaticEntityInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[T]) (T, error) {
	var zero T

	mapper := r.getBaseRepository().GetDocumentMapper()
	doc, err := mapper.ToDocumentWithoutMetadata(dto.Data, r.getBaseRepository().GetIDStrategy())
	if err != nil {
		return zero, sdk.NewInternalServerError(err)
	}

	// Get provided ID
	providedID := dto.Data.GetID()

	idStr, err := r.getBaseRepository().Insert(ctx, doc, providedID)
	if err != nil {
		return zero, err
	}

	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: idStr})
}

// Update modifies an existing entity by ID.
func (r *MongoStaticEntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[map[string]interface{}]) (T, error) {
	var zero T

	if err := r.getBaseRepository().Update(ctx, dto.Id, dto.Data); err != nil {
		return zero, err
	}

	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
}

// Delete removes an entity by ID.
func (r *MongoStaticEntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.getBaseRepository().Delete(ctx, dto.Id)
}

func (r *MongoStaticEntityInstanceRepository[T]) FindReferences(ctx context.Context, dto sdk.ReadInstancesDTO) (sdk.EntityReferenceGroupDescriptions, error) {
	var zero T
	descriptionAttributeKey := sdk.NewSchema(zero).UISchema.EntityDescriptionKey
	if descriptionAttributeKey == nil {
		return make(sdk.EntityReferenceGroupDescriptions), nil
	}
	return r.getBaseRepository().FindReferences(ctx, dto, *descriptionAttributeKey)
}

func (r *MongoStaticEntityInstanceRepository[T]) InstanceWithReferences(ctx context.Context, dto sdk.ReadInstanceDTO) (T, sdk.EntityRefererenceGroup, error) {
	var zero T

	rawDoc, err := r.getBaseRepository().FindByID(ctx, dto.Id)
	if err != nil {
		return zero, nil, err
	}

	entityIDs := extractEntityReferenceIDsFromDoc(sdk.NewSchema(zero), rawDoc)
	references, err := resolveEntityReferences(ctx, entityIDs)
	if err != nil {
		return zero, nil, err
	}

	instance, err := r.toModel(rawDoc)
	if err != nil {
		return zero, nil, err
	}
	if r.options.Hooks.AfterFind != nil {
		err = r.options.Hooks.AfterFind(instance)
		if err != nil {
			return zero, nil, err
		}
	}

	return instance, references, nil
}

func (r *MongoStaticEntityInstanceRepository[T]) ListWithReferences(ctx context.Context, dto sdk.ReadDTO) ([]T, sdk.EntityRefererenceGroup, error) {
	var filter bson.M
	if dto.Filter != nil {
		filter = cloneBsonM(dto.Filter)
	}

	var projection bson.M
	if dto.Projection != nil {
		projection = cloneBsonM(dto.Projection)
	}

	rawDocs, err := r.getBaseRepository().Find(ctx, filter, projection)
	if err != nil {
		return nil, nil, err
	}

	var zero T
	schema := sdk.NewSchema(zero)
	allEntityIDs := make(map[string][]string)
	for _, rawDoc := range rawDocs {
		for entityName, ids := range extractEntityReferenceIDsFromDoc(schema, rawDoc) {
			allEntityIDs[entityName] = append(allEntityIDs[entityName], ids...)
		}
	}

	references, err := resolveEntityReferences(ctx, allEntityIDs)
	if err != nil {
		return nil, nil, err
	}

	instances := make([]T, 0, len(rawDocs))
	for _, rawDoc := range rawDocs {
		instance, err := r.toModel(rawDoc)
		if err != nil {
			return nil, nil, err
		}
		// call hook
		if r.options.Hooks.AfterFind != nil {
			err := r.options.Hooks.AfterFind(instance)
			if err != nil {
				return nil, nil, err
			}
		}
		instances = append(instances, instance)
	}

	return instances, references, nil
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

func (r *MongoStaticEntityInstanceRepository[T]) getBaseRepository() *mongoBaseRepository[T] {
	if r._base != nil {
		return r._base
	}
	client, _ := sdk.GetMongoClient()
	dbName := sdk_configuration.GetConfig().DynamicEntityDocumentDBName
	if *r.options.Development == true && *r.options.UserId != "" {
		dbName = *r.options.UserId + "-" + dbName
	}
	collection := client.Database(dbName).Collection(r.entityId)
	return newMongoBaseRepository[T](collection, *r.options.AutoGenerateID)
}
