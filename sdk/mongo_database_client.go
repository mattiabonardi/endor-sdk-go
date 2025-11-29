// Package sdk provides the MongoDB implementation of database client interfaces.
// This implementation wraps the MongoDB driver to satisfy the DatabaseClientInterface
// while maintaining all existing functionality and performance characteristics.
package sdk

import (
	"context"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDatabaseClient implements DatabaseClientInterface using MongoDB driver.
// This provides the concrete implementation for dependency injection while maintaining
// the same performance and functionality as direct MongoDB client usage.
//
// Acceptance Criteria 3: All MongoDB operations use injected DatabaseClientInterface.
type MongoDatabaseClient struct {
	client *mongo.Client
}

// NewMongoDatabaseClient creates a new MongoDB database client implementation.
// This enables dependency injection of database clients in repository constructors.
//
// Acceptance Criteria 5: Repository Factory Patterns - support direct construction
// with NewRepositoryWithClient() pattern.
func NewMongoDatabaseClient(client *mongo.Client) interfaces.DatabaseClientInterface {
	return &MongoDatabaseClient{
		client: client,
	}
}

// Collection returns a MongoDB collection interface for the specified collection name.
func (m *MongoDatabaseClient) Collection(name string) interfaces.CollectionInterface {
	// Get default database from config
	config := GetConfig()
	database := m.client.Database(config.DynamicResourceDocumentDBName)
	collection := database.Collection(name)
	return &MongoCollectionAdapter{collection: collection}
}

// Database returns a MongoDB database interface for the specified database name.
func (m *MongoDatabaseClient) Database(name string) interfaces.DatabaseInterface {
	database := m.client.Database(name)
	return &MongoDatabaseAdapter{database: database}
}

// StartTransaction begins a new MongoDB transaction.
//
// Acceptance Criteria 7: Transaction handling works with injected database client interfaces.
func (m *MongoDatabaseClient) StartTransaction(ctx context.Context) (interfaces.TransactionInterface, error) {
	session, err := m.client.StartSession()
	if err != nil {
		return nil, err
	}

	return &MongoTransactionAdapter{session: session}, nil
}

// Close closes the MongoDB client connection and cleans up resources.
//
// Acceptance Criteria 4: Connection lifecycle management is managed by injected dependencies.
func (m *MongoDatabaseClient) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// Ping verifies connectivity to the MongoDB server.
func (m *MongoDatabaseClient) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

// MongoCollectionAdapter adapts MongoDB collection to CollectionInterface.
type MongoCollectionAdapter struct {
	collection *mongo.Collection
}

// FindOne adapts MongoDB FindOne to the interface.
func (m *MongoCollectionAdapter) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) interfaces.SingleResultInterface {
	result := m.collection.FindOne(ctx, filter, opts...)
	return &MongoSingleResultAdapter{result: result}
}

// Find adapts MongoDB Find to the interface.
func (m *MongoCollectionAdapter) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (interfaces.CursorInterface, error) {
	cursor, err := m.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return &MongoCursorAdapter{cursor: cursor}, nil
}

// InsertOne adapts MongoDB InsertOne to the interface.
func (m *MongoCollectionAdapter) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return m.collection.InsertOne(ctx, document, opts...)
}

// UpdateOne adapts MongoDB UpdateOne to the interface.
func (m *MongoCollectionAdapter) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.collection.UpdateOne(ctx, filter, update, opts...)
}

// DeleteOne adapts MongoDB DeleteOne to the interface.
func (m *MongoCollectionAdapter) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return m.collection.DeleteOne(ctx, filter, opts...)
}

// CountDocuments adapts MongoDB CountDocuments to the interface.
func (m *MongoCollectionAdapter) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.collection.CountDocuments(ctx, filter, opts...)
}

// Aggregate adapts MongoDB Aggregate to the interface.
func (m *MongoCollectionAdapter) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (interfaces.CursorInterface, error) {
	cursor, err := m.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, err
	}
	return &MongoCursorAdapter{cursor: cursor}, nil
}

// MongoDatabaseAdapter adapts MongoDB database to DatabaseInterface.
type MongoDatabaseAdapter struct {
	database *mongo.Database
}

// Collection returns a collection interface from the database.
func (m *MongoDatabaseAdapter) Collection(name string) interfaces.CollectionInterface {
	collection := m.database.Collection(name)
	return &MongoCollectionAdapter{collection: collection}
}

// Name returns the database name.
func (m *MongoDatabaseAdapter) Name() string {
	return m.database.Name()
}

// RunCommand runs a database command.
func (m *MongoDatabaseAdapter) RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) interfaces.SingleResultInterface {
	result := m.database.RunCommand(ctx, runCommand, opts...)
	return &MongoSingleResultAdapter{result: result}
}

// MongoSingleResultAdapter adapts MongoDB SingleResult to SingleResultInterface.
type MongoSingleResultAdapter struct {
	result *mongo.SingleResult
}

// Decode decodes the result into the provided value.
func (m *MongoSingleResultAdapter) Decode(v interface{}) error {
	return m.result.Decode(v)
}

// Err returns any error from the operation.
func (m *MongoSingleResultAdapter) Err() error {
	return m.result.Err()
}

// MongoCursorAdapter adapts MongoDB cursor to CursorInterface.
type MongoCursorAdapter struct {
	cursor *mongo.Cursor
}

// Next advances the cursor to the next document.
func (m *MongoCursorAdapter) Next(ctx context.Context) bool {
	return m.cursor.Next(ctx)
}

// Decode decodes the current document into the provided value.
func (m *MongoCursorAdapter) Decode(val interface{}) error {
	return m.cursor.Decode(val)
}

// Close closes the cursor and releases resources.
func (m *MongoCursorAdapter) Close(ctx context.Context) error {
	return m.cursor.Close(ctx)
}

// Err returns any error from cursor operations.
func (m *MongoCursorAdapter) Err() error {
	return m.cursor.Err()
}

// All decodes all documents from the cursor into the provided slice.
func (m *MongoCursorAdapter) All(ctx context.Context, results interface{}) error {
	return m.cursor.All(ctx, results)
}

// MongoTransactionAdapter adapts MongoDB session to TransactionInterface.
type MongoTransactionAdapter struct {
	session mongo.Session
}

// Commit commits the transaction.
func (m *MongoTransactionAdapter) Commit(ctx context.Context) error {
	return m.session.CommitTransaction(ctx)
}

// Abort aborts the transaction.
func (m *MongoTransactionAdapter) Abort(ctx context.Context) error {
	return m.session.AbortTransaction(ctx)
}

// WithTransaction executes a function within the transaction context.
func (m *MongoTransactionAdapter) WithTransaction(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	return m.session.WithTransaction(ctx, func(sessionCtx mongo.SessionContext) (interface{}, error) {
		return fn(sessionCtx)
	})
}

// DefaultDatabaseClient creates a default MongoDB database client using the global singleton.
// This provides backward compatibility for existing repository creation patterns.
//
// Acceptance Criteria 6: Backward compatibility with convenience constructors using default implementations.
func DefaultDatabaseClient() (interfaces.DatabaseClientInterface, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, interfaces.NewDatabaseRepositoryError(
			interfaces.RepositoryErrorCodeDatabaseConnection,
			"Failed to create default database client",
			err,
		)
	}
	return NewMongoDatabaseClient(client), nil
}
