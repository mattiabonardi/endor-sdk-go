// Package interfaces provides database client abstractions for endor-sdk-go framework.
// These interfaces enable dependency injection of database operations, allowing for
// alternative implementations and comprehensive testing with mock database clients.
//
// The database abstraction follows the pattern established in Stories 2.1-2.3 where
// dependencies are injected as interface parameters to constructors.
package interfaces

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseClientInterface abstracts database client operations for dependency injection.
// This enables alternative implementations (testing, different databases) and eliminates
// hard-coded calls to GetMongoClient() throughout the repository layer.
//
// Acceptance Criteria 3: All MongoDB operations use injected DatabaseClientInterface,
// eliminating hard-coded GetMongoClient() calls.
type DatabaseClientInterface interface {
	// Collection returns a collection interface for the specified collection name
	Collection(name string) CollectionInterface
	// Database returns a database interface for the specified database name
	Database(name string) DatabaseInterface
	// StartTransaction begins a new transaction, returning a transaction interface
	StartTransaction(ctx context.Context) (TransactionInterface, error)
	// Close closes the database connection and cleans up resources
	Close(ctx context.Context) error
	// Ping verifies connectivity to the database server
	Ping(ctx context.Context) error
}

// CollectionInterface abstracts MongoDB collection operations for repository implementations.
// This enables mocking of collection operations in tests while maintaining the same
// method signatures as the actual MongoDB collection.
//
// Acceptance Criteria 2: Repository implementations satisfy RepositoryInterface from
// interfaces package and can be mocked in tests.
type CollectionInterface interface {
	// FindOne finds a single document matching the filter
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResultInterface
	// Find finds multiple documents matching the filter
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (CursorInterface, error)
	// InsertOne inserts a single document
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	// UpdateOne updates a single document matching the filter
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	// DeleteOne deletes a single document matching the filter
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	// CountDocuments counts documents matching the filter
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
	// Aggregate runs an aggregation pipeline
	Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (CursorInterface, error)
}

// SingleResultInterface abstracts MongoDB single result operations.
// This enables mocking of single result decoding and error handling in tests.
type SingleResultInterface interface {
	// Decode decodes the result into the provided value
	Decode(v interface{}) error
	// Err returns any error from the operation
	Err() error
}

// CursorInterface abstracts MongoDB cursor operations for iteration over results.
// This enables mocking of cursor operations in repository tests.
type CursorInterface interface {
	// Next advances the cursor to the next document
	Next(ctx context.Context) bool
	// Decode decodes the current document into the provided value
	Decode(val interface{}) error
	// Close closes the cursor and releases resources
	Close(ctx context.Context) error
	// Err returns any error from cursor operations
	Err() error
	// All decodes all documents from the cursor into the provided slice
	All(ctx context.Context, results interface{}) error
}

// DatabaseInterface abstracts MongoDB database operations.
// This enables dependency injection of database instances and testing with mock databases.
type DatabaseInterface interface {
	// Collection returns a collection interface for the specified collection name
	Collection(name string) CollectionInterface
	// Name returns the database name
	Name() string
	// RunCommand runs a database command
	RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) SingleResultInterface
}

// TransactionInterface abstracts MongoDB transaction operations for dependency injection.
// This enables transaction support to work through injected client interfaces.
//
// Acceptance Criteria 7: Transaction handling works with injected database client interfaces.
type TransactionInterface interface {
	// Commit commits the transaction
	Commit(ctx context.Context) error
	// Abort aborts the transaction
	Abort(ctx context.Context) error
	// WithTransaction executes a function within the transaction context
	WithTransaction(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error)
}

// RepositoryDependencies defines the complete set of dependencies required for
// repository construction with dependency injection.
//
// Acceptance Criteria 1: Repository constructors accept DatabaseClientInterface
// and ConfigInterface instead of using global singletons.
type RepositoryDependencies struct {
	// DatabaseClient provides database access operations
	DatabaseClient DatabaseClientInterface
	// Config provides repository configuration settings
	Config ConfigProviderInterface
	// Logger provides logging capabilities (optional)
	Logger LoggerInterface
	// MicroServiceID identifies the microservice for multi-tenant scenarios
	MicroServiceID string
}

// DatabaseRepositoryError defines structured error types for repository operations
// following the pattern established in previous stories.
type DatabaseRepositoryError struct {
	Code    string
	Message string
	Cause   error
}

func (e DatabaseRepositoryError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e DatabaseRepositoryError) ErrorCode() string {
	return e.Code
}

func (e DatabaseRepositoryError) ErrorMessage() string {
	return e.Message
}

// Repository error codes for standardized error handling
const (
	RepositoryErrorCodeInvalidDependencies = "REPOSITORY_INVALID_DEPENDENCIES"
	RepositoryErrorCodeDatabaseConnection  = "REPOSITORY_DATABASE_CONNECTION"
	RepositoryErrorCodeNotFound            = "REPOSITORY_NOT_FOUND"
	RepositoryErrorCodeOperationFailed     = "REPOSITORY_OPERATION_FAILED"
)

// NewDatabaseRepositoryError creates a new repository error with the specified code and message
func NewDatabaseRepositoryError(code, message string, cause error) DatabaseRepositoryError {
	return DatabaseRepositoryError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
