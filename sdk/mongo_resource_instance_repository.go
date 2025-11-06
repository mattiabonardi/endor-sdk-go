package sdk

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoResourceInstanceRepository[T any, ID comparable] struct {
	collection *mongo.Collection
}

func NewMongoResourceInstanceRepository[T any, ID comparable](resourceId string) *MongoResourceInstanceRepository[T, ID] {
	client, _ := GetMongoClient()

	collection := client.Database(GetConfig().EndorDynamicResourceDBName).Collection(resourceId)
	return &MongoResourceInstanceRepository[T, ID]{
		collection: collection,
	}
}

func (r *MongoResourceInstanceRepository[T, ID]) Instance(ctx context.Context, dto ReadInstanceDTO[ID]) (*ResourceInstance[T], error) {
	var result ResourceInstance[T]
	err := r.collection.FindOne(ctx, bson.M{"_id": dto.Id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *MongoResourceInstanceRepository[T, ID]) List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error) {
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

func (r *MongoResourceInstanceRepository[T, ID]) Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	_, err := r.collection.InsertOne(ctx, dto.Data)
	if err != nil {
		return nil, err
	}
	return &dto.Data, nil
}

func (r *MongoResourceInstanceRepository[T, ID]) Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T], ID]) (*ResourceInstance[T], error) {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": dto.Id}, dto.Data)
	if err != nil {
		return nil, err
	}
	return &dto.Data, nil
}

func (r *MongoResourceInstanceRepository[T, ID]) Delete(ctx context.Context, dto DeleteByIdDTO[ID]) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": dto.Id})
	return err
}
