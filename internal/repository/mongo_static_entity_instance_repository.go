package repository

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"go.mongodb.org/mongo-driver/bson"
)

// MongoStaticEntityInstanceRepository provides MongoDB implementation for StaticEntityInstanceRepositoryInterface.
type MongoStaticEntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	base *mongoBaseRepository[T]
}

func NewMongoStaticEntityInstanceRepository[T sdk.EntityInstanceInterface](entityId string, options sdk.StaticEntityInstanceRepositoryOptions) *MongoStaticEntityInstanceRepository[T] {
	client, _ := sdk.GetMongoClient()
	collection := client.Database(sdk_configuration.GetConfig().DynamicEntityDocumentDBName).Collection(entityId)

	return &MongoStaticEntityInstanceRepository[T]{
		base: newMongoBaseRepository[T](collection, *options.AutoGenerateID),
	}
}

func (r *MongoStaticEntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (T, error) {
	var zero T

	rawResult, err := r.base.findOne(ctx, dto.Id)
	if err != nil {
		return zero, err
	}

	// Decode directly into the model - ObjectID fields are automatically converted via UnmarshalBSONValue
	var result T
	entityBytes, err := bson.Marshal(rawResult)
	if err != nil {
		return zero, sdk.NewInternalServerError(fmt.Errorf("failed to marshal entity: %w", err))
	}

	if err := bson.Unmarshal(entityBytes, &result); err != nil {
		return zero, sdk.NewInternalServerError(fmt.Errorf("failed to unmarshal entity: %w", err))
	}

	return result, nil
}

func (r *MongoStaticEntityInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]T, error) {
	rawResults, err := r.base.find(ctx, dto.Filter, dto.Projection)
	if err != nil {
		return nil, err
	}

	// Decode directly into models - ObjectID fields are automatically converted via UnmarshalBSONValue
	results := make([]T, 0, len(rawResults))
	for _, raw := range rawResults {
		var result T
		entityBytes, err := bson.Marshal(raw)
		if err != nil {
			return nil, sdk.NewInternalServerError(fmt.Errorf("failed to marshal entity: %w", err))
		}

		if err := bson.Unmarshal(entityBytes, &result); err != nil {
			return nil, sdk.NewInternalServerError(fmt.Errorf("failed to unmarshal entity: %w", err))
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *MongoStaticEntityInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[T]) (T, error) {
	var zero T

	// Serialize the struct - ObjectID fields are automatically converted via MarshalBSONValue
	docBytes, err := bson.Marshal(dto.Data)
	if err != nil {
		return zero, sdk.NewInternalServerError(fmt.Errorf("failed to marshal entity: %w", err))
	}
	var doc bson.M
	if err := bson.Unmarshal(docBytes, &doc); err != nil {
		return zero, sdk.NewInternalServerError(fmt.Errorf("failed to unmarshal entity: %w", err))
	}

	// Insert using base repository
	idStr, err := r.base.insertOne(ctx, doc, dto.Data.GetID())
	if err != nil {
		return zero, err
	}

	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: idStr})
}

func (r *MongoStaticEntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.base.deleteOne(ctx, dto.Id)
}

func (r *MongoStaticEntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[map[string]interface{}]) (T, error) {
	var zero T

	// Update using base repository
	if err := r.base.updateOne(ctx, dto.Id, dto.Data); err != nil {
		return zero, err
	}

	// Retrieve and return the updated document
	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
}
