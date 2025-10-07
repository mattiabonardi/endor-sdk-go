package sdk

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance    *mongo.Client
	clientInstanceErr error
	mongoOnce         sync.Once
	connectionTimeout = 10 * time.Second
)

// GetMongoClient returns a singleton MongoDB client
func GetMongoClient() (*mongo.Client, error) {
	configuration := LoadConfiguration()
	mongoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()

		clientOptions := options.Client().ApplyURI(configuration.EndorServiceDBUri)
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			clientInstanceErr = err
			return
		}

		err = client.Ping(ctx, nil)
		if err != nil {
			clientInstanceErr = err
			return
		}

		clientInstance = client
		log.Println("MongoDB connected successfully")
	})

	return clientInstance, clientInstanceErr
}
