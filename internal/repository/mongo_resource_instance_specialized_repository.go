package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/internal/configuration"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoResourceInstanceSpecializedRepository with dependency injection
type MongoResourceInstanceSpecializedRepository[T sdk.ResourceInstanceSpecializedInterface, C any] struct {
	collectionName string
	idConverter    IDConverter
	docConverter   *SpecializedDocumentConverter[T, C]
	autoGenerateID bool
}

// NewMongoResourceInstanceSpecializedRepository creates specialized repository with injected dependencies
func NewMongoResourceInstanceSpecializedRepository[T sdk.ResourceInstanceSpecializedInterface, C any](
	collectionName string,
	options ResourceInstanceRepositoryOptions,
) *MongoResourceInstanceSpecializedRepository[T, C] {
	var idConverter IDConverter
	if *options.AutoGenerateID {
		idConverter = &ObjectIDConverter{}
	} else {
		idConverter = &StringIDConverter{}
	}

	return &MongoResourceInstanceSpecializedRepository[T, C]{
		collectionName: collectionName,
		idConverter:    idConverter,
		docConverter:   &SpecializedDocumentConverter[T, C]{},
		autoGenerateID: *options.AutoGenerateID,
	}
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.ResourceInstanceSpecialized[T, C], error) {
	var rawResult bson.M

	filter, err := r.idConverter.ToFilter(dto.Id)
	if err != nil {
		return nil, sdk.NewBadRequestError(err)
	}

	err = r.getCollection().FindOne(ctx, filter).Decode(&rawResult)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sdk.NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
		}
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to find resource instance: %w", err))
	}

	return r.docConverter.ToSpecialized(rawResult, r.idConverter)
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) List(ctx context.Context, dto sdk.ReadDTO) ([]sdk.ResourceInstanceSpecialized[T, C], error) {
	mongoFilter, err := prepareFilter[T](dto.Filter)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetProjection(prepareProjection[T](dto.Projection))
	cursor, err := r.getCollection().Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to find resource instances: %w", err))
	}
	defer cursor.Close(ctx)

	var rawResults []bson.M
	if err = cursor.All(ctx, &rawResults); err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to decode resource instances: %w", err))
	}

	specializedResults := make([]sdk.ResourceInstanceSpecialized[T, C], 0, len(rawResults))
	for _, rawResult := range rawResults {
		converted, err := r.docConverter.ToSpecialized(rawResult, r.idConverter)
		if err != nil {
			return nil, sdk.NewInternalServerError(err)
		}
		specializedResults = append(specializedResults, *converted)
	}

	return specializedResults, nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Create(ctx context.Context, dto sdk.CreateDTO[sdk.ResourceInstanceSpecialized[T, C]]) (*sdk.ResourceInstanceSpecialized[T, C], error) {
	idPtr := dto.Data.This.GetID()
	var idStr string

	if r.autoGenerateID {
		idStr = r.idConverter.GenerateNewID()
		dto.Data.This.SetID(idStr)
	} else {
		if idPtr == nil || *idPtr == "" {
			return nil, sdk.NewBadRequestError(fmt.Errorf("ID is required when auto-generation is disabled"))
		}
		idStr = *idPtr

		// Check if exists
		_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: idStr})
		if err == nil {
			return nil, sdk.NewConflictError(fmt.Errorf("resource instance with id %v already exists", idStr))
		}
		var endorErr *sdk.EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			return nil, err
		}
	}

	doc, err := r.docConverter.ToDocument(dto.Data, r.idConverter)
	if err != nil {
		return nil, sdk.NewInternalServerError(err)
	}

	_, err = r.getCollection().InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, sdk.NewConflictError(fmt.Errorf("resource instance already exists: %w", err))
		}
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to create resource instance: %w", err))
	}

	return &dto.Data, nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.ResourceInstanceSpecialized[T, C]]) (*sdk.ResourceInstanceSpecialized[T, C], error) {
	_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return nil, err
	}

	filter, err := r.idConverter.ToFilter(dto.Id)
	if err != nil {
		return nil, sdk.NewBadRequestError(err)
	}

	doc, err := r.docConverter.ToDocument(dto.Data, r.idConverter)
	if err != nil {
		return nil, sdk.NewInternalServerError(err)
	}

	result, err := r.getCollection().ReplaceOne(ctx, filter, doc)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to update resource instance: %w", err))
	}

	if result.MatchedCount == 0 {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	dto.Data.This.SetID(dto.Id)
	return &dto.Data, nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return err
	}

	filter, err := r.idConverter.ToFilter(dto.Id)
	if err != nil {
		return sdk.NewBadRequestError(err)
	}

	result, err := r.getCollection().DeleteOne(ctx, filter)
	if err != nil {
		return sdk.NewInternalServerError(fmt.Errorf("failed to delete resource instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return sdk.NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	return nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) getCollection() *mongo.Collection {
	client, _ := sdk.GetMongoClient()
	return client.Database(configuration.GetConfig().DynamicResourceDocumentDBName).Collection(r.collectionName)
}
