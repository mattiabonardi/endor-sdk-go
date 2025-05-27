package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoResourceRepository[T any] struct {
	collection *mongo.Collection
}

func NewMongoResourceRepository[T any](client *mongo.Client, resource sdk.Resource) *MongoResourceRepository[T] {
	dbName := /*resource.Persistence.Options["database"]*/ ""
	collName := /*resource.Persistence.Options["collection"]*/ ""

	collection := client.Database(dbName).Collection(collName)
	return &MongoResourceRepository[T]{
		collection: collection,
	}
}

// Instance retrieves a document by ID
func (r *MongoResourceRepository[T]) Instance(id string, _ sdk.IntanceOptions) (T, error) {
	var result T

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return result, nil // not found, return zero T and no error (optional behavior)
		}
		return result, err
	}

	return result, nil
}

// List retrieves all documents in the collection
func (r *MongoResourceRepository[T]) List(_ sdk.ListOptions) ([]T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []T
	for cursor.Next(ctx) {
		var elem T
		if err := cursor.Decode(&elem); err != nil {
			return nil, err
		}
		results = append(results, elem)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Create inserts a new document
func (r *MongoResourceRepository[T]) Create(resource T) (T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, resource)
	if err != nil {
		var zero T
		return zero, err
	}

	return resource, nil
}

// Update replaces a document by ID
func (r *MongoResourceRepository[T]) Update(id string, resource T) (T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": id}, resource)
	if err != nil {
		var zero T
		return zero, err
	}
	if result.MatchedCount == 0 {
		return resource, mongo.ErrNoDocuments
	}

	return resource, nil
}

// Delete removes a document by ID
func (r *MongoResourceRepository[T]) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
