package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoEntityInstanceRepository with dependency injection
type MongoEntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	collectionName string
	idConverter    IDConverter
	docConverter   *DocumentConverter[T]
	autoGenerateID bool
}

// NewMongoEntityInstanceRepository creates repository with injected dependencies
func NewMongoEntityInstanceRepository[T sdk.EntityInstanceInterface](
	collectionName string,
	options sdk.EntityInstanceRepositoryOptions,
) *MongoEntityInstanceRepository[T] {
	var idConverter IDConverter
	if *options.AutoGenerateID {
		idConverter = &ObjectIDConverter{}
	} else {
		idConverter = &StringIDConverter{}
	}

	return &MongoEntityInstanceRepository[T]{
		collectionName: collectionName,
		idConverter:    idConverter,
		docConverter:   &DocumentConverter[T]{},
		autoGenerateID: *options.AutoGenerateID,
	}
}

func (r *MongoEntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.EntityInstance[T], error) {
	var rawResult bson.M

	filter, err := r.idConverter.ToFilter(dto.Id)
	if err != nil {
		return nil, sdk.NewBadRequestError(err)
	}

	err = r.getCollection().FindOne(ctx, filter).Decode(&rawResult)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
		}
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to find entity instance: %w", err))
	}

	metadata, err := r.docConverter.ExtractMetadata(rawResult)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to extract metadata: %w", err))
	}

	model, err := r.docConverter.ToModel(rawResult, r.idConverter)
	if err != nil {
		return nil, sdk.NewInternalServerError(err)
	}

	return &sdk.EntityInstance[T]{
		This:     model,
		Metadata: metadata,
	}, nil
}

func (r *MongoEntityInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]sdk.EntityInstance[T], error) {
	mongoFilter, err := prepareFilter[T](dto.Filter)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetProjection(prepareProjection[T](dto.Projection))
	cursor, err := r.getCollection().Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to list entities: %w", err))
	}
	defer cursor.Close(ctx)

	var rawResults []bson.M
	if err := cursor.All(ctx, &rawResults); err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to decode entities: %w", err))
	}

	results := make([]sdk.EntityInstance[T], 0, len(rawResults))
	for _, raw := range rawResults {
		metadata, err := r.docConverter.ExtractMetadata(raw)
		if err != nil {
			return nil, sdk.NewInternalServerError(fmt.Errorf("failed to extract metadata: %w", err))
		}

		model, err := r.docConverter.ToModel(raw, r.idConverter)
		if err != nil {
			return nil, sdk.NewInternalServerError(err)
		}

		results = append(results, sdk.EntityInstance[T]{
			This:     model,
			Metadata: metadata,
		})
	}

	return results, nil
}

func (r *MongoEntityInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[sdk.EntityInstance[T]]) (*sdk.EntityInstance[T], error) {
	idPtr := dto.Data.This.GetID()
	var idStr string

	if r.autoGenerateID {
		idStr = r.idConverter.GenerateNewID()
		dto.Data.This.SetID(idStr)
	} else {
		if idPtr == "" {
			return nil, sdk.NewBadRequestError(fmt.Errorf("ID is required when auto-generation is disabled"))
		}
		idStr = idPtr

		// Check if exists
		_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: idStr})
		if err == nil {
			return nil, sdk.NewConflictError(fmt.Errorf("entity instance with id %v already exists", idStr))
		}
		var endorErr *sdk.EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			return nil, err
		}
	}

	doc, err := r.docConverter.ToDocument(dto.Data.This, dto.Data.Metadata, r.idConverter)
	if err != nil {
		return nil, sdk.NewInternalServerError(err)
	}

	_, err = r.getCollection().InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, sdk.NewConflictError(fmt.Errorf("entity instance already exists: %w", err))
		}
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to create entity instance: %w", err))
	}

	return &dto.Data, nil
}

func (r *MongoEntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
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
		return sdk.NewInternalServerError(fmt.Errorf("failed to delete entity instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
	}

	return nil
}

func (r *MongoEntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.PartialEntityInstance[T]]) (*sdk.EntityInstance[T], error) {
	// Verify the instance exists
	_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return nil, err
	}

	filter, err := r.idConverter.ToFilter(dto.Id)
	if err != nil {
		return nil, sdk.NewBadRequestError(err)
	}

	// Prepare the $set document with both entity data and metadata
	setDoc := bson.M{}

	// Add entity data fields (at root level) from This map
	for k, v := range dto.Data.This {
		setDoc[k] = v
	}

	// Add metadata fields (under metadata prefix) from Metadata map
	for k, v := range dto.Data.Metadata {
		setDoc["metadata."+k] = v
	}

	// If no fields to update, return error
	if len(setDoc) == 0 {
		return nil, sdk.NewBadRequestError(fmt.Errorf("no fields to update"))
	}

	// Perform the update with $set
	update := bson.M{"$set": setDoc}
	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to update entity instance: %w", err))
	}
	if result.MatchedCount == 0 {
		return nil, sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
	}

	// Retrieve and return the updated document
	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
}

// Helper functions remain the same
func prepareProjection[T sdk.EntityInstanceInterface](projection map[string]interface{}) bson.M {
	result := bson.M{}
	var thisModel T
	modelFields := map[string]struct{}{}
	b, _ := bson.Marshal(thisModel)
	_ = bson.Unmarshal(b, &modelFields)

	for k, v := range projection {
		if _, ok := modelFields[k]; ok {
			result[k] = v
		} else {
			result["metadata."+k] = v
		}
	}
	return result
}

func prepareFilter[T sdk.EntityInstanceInterface](filter map[string]interface{}) (bson.M, error) {
	result := bson.M{}
	var thisModel T
	modelFields := map[string]struct{}{}

	// Use reflection to get the actual field names from bson tags
	t := reflect.TypeOf(thisModel)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		bsonTag := field.Tag.Get("bson")
		if bsonTag != "" && bsonTag != "-" {
			// Handle bson tags like "type" or "type,omitempty"
			tagName := strings.Split(bsonTag, ",")[0]
			modelFields[tagName] = struct{}{}
		}
	}

	for k, v := range filter {
		if _, ok := modelFields[k]; ok {
			result[k] = v
		} else {
			result["metadata."+k] = v
		}
	}
	return result, nil
}

func (r *MongoEntityInstanceRepository[T]) getCollection() *mongo.Collection {
	client, _ := sdk.GetMongoClient()
	return client.Database(sdk_configuration.GetConfig().DynamicEntityDocumentDBName).Collection(r.collectionName)
}
