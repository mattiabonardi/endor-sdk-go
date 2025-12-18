package sdk

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
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
	configuration := sdk_configuration.GetConfig()
	mongoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()

		clientOptions := options.Client().ApplyURI(configuration.DocumentDBUri)
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
