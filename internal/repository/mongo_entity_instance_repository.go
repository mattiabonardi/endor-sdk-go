package repository

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoEntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	base         *mongoBaseRepository[T]
	docConverter *DocumentConverter[T]
}

func NewMongoEntityInstanceRepository[T sdk.EntityInstanceInterface](
	collectionName string,
	options sdk.EntityInstanceRepositoryOptions,
) *MongoEntityInstanceRepository[T] {
	client, _ := sdk.GetMongoClient()
	collection := client.Database(sdk_configuration.GetConfig().DynamicEntityDocumentDBName).Collection(collectionName)

	return &MongoEntityInstanceRepository[T]{
		base:         newMongoBaseRepository[T](collection, *options.AutoGenerateID),
		docConverter: &DocumentConverter[T]{},
	}
}

func (r *MongoEntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.EntityInstance[T], error) {
	rawResult, err := r.base.findOne(ctx, dto.Id)
	if err != nil {
		return nil, err
	}

	metadata, err := r.docConverter.ExtractMetadata(rawResult)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to extract metadata: %w", err))
	}

	model, err := r.docConverter.ToModel(rawResult, r.base.idConverter)
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

	mongoProjection := prepareProjection[T](dto.Projection)

	rawResults, err := r.base.find(ctx, mongoFilter, mongoProjection)
	if err != nil {
		return nil, err
	}

	results := make([]sdk.EntityInstance[T], 0, len(rawResults))
	for _, raw := range rawResults {
		metadata, err := r.docConverter.ExtractMetadata(raw)
		if err != nil {
			return nil, sdk.NewInternalServerError(fmt.Errorf("failed to extract metadata: %w", err))
		}

		model, err := r.docConverter.ToModel(raw, r.base.idConverter)
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
	doc, err := r.docConverter.ToDocument(dto.Data.This, dto.Data.Metadata, r.base.idConverter)
	if err != nil {
		return nil, sdk.NewInternalServerError(err)
	}

	// Get provided ID (if any)
	idPtr := dto.Data.This.GetID()
	var providedID string
	if !isIDEmpty(idPtr) {
		providedID = idToString(idPtr)
	}

	// Insert using base repository
	idStr, err := r.base.insertOne(ctx, doc, providedID)
	if err != nil {
		return nil, err
	}

	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: idStr})
}

func (r *MongoEntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.base.deleteOne(ctx, dto.Id)
}

func (r *MongoEntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.PartialEntityInstance[T]]) (*sdk.EntityInstance[T], error) {
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

	// Update using base repository
	if err := r.base.updateOne(ctx, dto.Id, setDoc); err != nil {
		return nil, err
	}

	// Retrieve and return the updated document
	return r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
}

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
