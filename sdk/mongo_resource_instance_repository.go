package sdk

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoResourceInstanceRepository[ID comparable, T ResourceInstanceInterface[ID]] struct {
	collection *mongo.Collection
}

func NewMongoResourceInstanceRepository[ID comparable, T ResourceInstanceInterface[ID]](resourceId string) *MongoResourceInstanceRepository[ID, T] {
	client, _ := GetMongoClient()

	collection := client.Database(GetConfig().EndorDynamicResourceDBName).Collection(resourceId)
	return &MongoResourceInstanceRepository[ID, T]{
		collection: collection,
	}
}

func (r *MongoResourceInstanceRepository[ID, T]) Instance(ctx context.Context, dto ReadInstanceDTO[ID]) (*ResourceInstance[ID, T], error) {
	var result ResourceInstance[ID, T]
	err := r.collection.FindOne(ctx, bson.M{"_id": dto.Id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *MongoResourceInstanceRepository[ID, T]) List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[ID, T], error) {
	var opts *options.FindOptions = options.Find().SetProjection(dto.Projection)
	cursor, err := r.collection.Find(ctx, dto.Filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []ResourceInstance[ID, T]
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *MongoResourceInstanceRepository[ID, T]) Create(ctx context.Context, dto CreateDTO[ResourceInstance[ID, T]]) (*ResourceInstance[ID, T], error) {
	// Auto-generate ObjectID if the ID type is primitive.ObjectID and the ID is empty
	if idPtr := dto.Data.This.GetID(); idPtr != nil {
		if oid, ok := any(*idPtr).(primitive.ObjectID); ok && oid.IsZero() {
			newOid := primitive.NewObjectID()

			// Use reflection to set the ID field generically
			thisValue := reflect.ValueOf(&dto.Data.This).Elem()
			if thisValue.Kind() == reflect.Struct {
				// Look for a field named "Id" of type primitive.ObjectID
				idField := thisValue.FieldByName("Id")
				if idField.IsValid() && idField.CanSet() && idField.Type() == reflect.TypeOf(primitive.ObjectID{}) {
					idField.Set(reflect.ValueOf(newOid))
				}
			}
		}
	}

	_, err := r.collection.InsertOne(ctx, dto.Data)
	if err != nil {
		return nil, err
	}
	return &dto.Data, nil
}

func (r *MongoResourceInstanceRepository[ID, T]) Update(ctx context.Context, dto UpdateByIdDTO[ID, ResourceInstance[ID, T]]) (*ResourceInstance[ID, T], error) {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": dto.Id}, dto.Data)
	if err != nil {
		return nil, err
	}
	return &dto.Data, nil
}

func (r *MongoResourceInstanceRepository[ID, T]) Delete(ctx context.Context, dto DeleteByIdDTO[ID]) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": dto.Id})
	return err
}
