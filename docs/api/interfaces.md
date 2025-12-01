# Interfaces

> Package documentation for Interfaces

**Import Path:** `github.com/mattiabonardi/endor-sdk-go/sdk/interfaces`
**Generated:** 2025-12-01 10:07:50 UTC

---

const RepositoryErrorCodeInvalidDependencies = "REPOSITORY_INVALID_DEPENDENCIES" ...
type CollectionInterface interface{ ... }
type ConfigProviderInterface interface{ ... }
type ConflictError interface{ ... }
type CursorInterface interface{ ... }
type DIContainerInterface interface{ ... }
type DatabaseClientInterface interface{ ... }
type DatabaseInterface interface{ ... }
type DatabaseRepositoryError struct{ ... }
    func NewDatabaseRepositoryError(code, message string, cause error) DatabaseRepositoryError
type EndorContext[T any] struct{ ... }
type EndorContextInterface[T any] interface{ ... }
type EndorHybridServiceCategory interface{ ... }
type EndorHybridServiceInterface interface{ ... }
type EndorServiceAction interface{ ... }
type EndorServiceActionOptions struct{ ... }
type EndorServiceInterface interface{ ... }
type InternalError interface{ ... }
type LogLevel int
    const DebugLevel LogLevel = iota ...
type LoggerInterface interface{ ... }
type Message struct{ ... }
type MessageGravity string
    const Info MessageGravity = "info" ...
type MockableRepository interface{ ... }
type NoPayload struct{}
type NotFoundError interface{ ... }
type RepositoryContainerFactoryFunc func(container DIContainerInterface) (RepositoryInterface, error)
type RepositoryDependencies struct{ ... }
type RepositoryError interface{ ... }
type RepositoryFactoryFunc func(deps RepositoryDependencies) (RepositoryInterface, error)
type RepositoryInterface interface{ ... }
type RepositoryOptions interface{ ... }
type RepositoryPattern interface{}
type ResourceInstanceInterface interface{ ... }
type ResourceInstanceSpecializedInterface interface{ ... }
type Response[T any] struct{ ... }
type RootSchema struct{ ... }
type Schema struct{ ... }
type SchemaType string
    const StringType SchemaType = "string" ...
type Session struct{ ... }
type SingleResultInterface interface{ ... }
type StructuredLoggerInterface interface{ ... }
type TestableRepository interface{ ... }
type TransactionInterface interface{ ... }
type ValidationError interface{ ... }

## Package Overview

package interfaces // import "github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"

Package interfaces provides configuration and context interfaces for
endor-sdk-go framework. These interfaces enable dependency injection,
testing with mocks, and flexible configuration management while maintaining the
framework's configuration patterns and context propagation.

Package interfaces provides database client abstractions for endor-sdk-go
framework. These interfaces enable dependency injection of database operations,
allowing for alternative implementations and comprehensive testing with mock
database clients.

The database abstraction follows the pattern established in Stories 2.1-2.3
where dependencies are injected as interface parameters to constructors.

Package interfaces provides logger interfaces for endor-sdk-go framework.
These interfaces enable dependency injection, testing with mocks, and flexible
logging implementations while maintaining consistent logging patterns across the
framework.


## Exported Types

### CollectionInterface

```go
type CollectionInterface interface{ ... }
```


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
    CollectionInterface abstracts MongoDB collection operations for repository
    implementations. This enables mocking of collection operations in tests
    while maintaining the same method signatures as the actual MongoDB
    collection.

    Acceptance Criteria 2: Repository implementations satisfy
    RepositoryInterface from interfaces package and can be mocked in tests.


### ConfigProviderInterface

```go
type ConfigProviderInterface interface{ ... }
```


type ConfigProviderInterface interface {
	// GetServerPort returns the HTTP server port configuration.
	// Typically read from PORT environment variable or defaults to "8080".
	GetServerPort() string

	// GetDocumentDBUri returns the MongoDB connection URI.
	// Typically read from DOCUMENT_DB_URI environment variable.
	GetDocumentDBUri() string

	// IsHybridResourcesEnabled returns whether hybrid resource functionality is enabled.
	// Typically read from HYBRID_RESOURCES_ENABLED environment variable.
	IsHybridResourcesEnabled() bool

	// IsDynamicResourcesEnabled returns whether dynamic resource functionality is enabled.
	// Typically read from DYNAMIC_RESOURCES_ENABLED environment variable.
	IsDynamicResourcesEnabled() bool

	// GetDynamicResourceDocumentDBName returns the database name for dynamic resources.
	// This is typically set at runtime based on the microservice ID.
	GetDynamicResourceDocumentDBName() string

	// Reload forces configuration reload from sources (environment variables, files).
	// Useful for testing scenarios where configuration changes need to be applied.
	// Returns error if configuration reload fails.
	Reload() error

	// Validate performs configuration validation to ensure all required values are present
	// and properly formatted. Returns error if configuration is invalid.
	Validate() error
}
    ConfigProviderInterface defines the contract for configuration access
    in the framework. This interface abstracts configuration loading and
    access patterns, enabling testing with different configuration values and
    supporting both environment-based and custom configuration sources.

    The interface follows the framework's existing configuration access patterns
    while making them testable and mockable for unit testing scenarios.

    Example usage:

        // Production usage with concrete implementation
        config := sdk.GetConfig() // implements ConfigProviderInterface
        port := config.GetServerPort()
        dbUri := config.GetDocumentDBUri()

        // Testing usage with mock configuration
        mockConfig := &MockConfigProvider{}
        mockConfig.On("GetServerPort").Return("8080")
        mockConfig.On("GetDocumentDBUri").Return("mongodb://test:27017")
        mockConfig.On("IsHybridResourcesEnabled").Return(true)

        // Custom configuration for testing
        testConfig := &TestConfigProvider{
        	ServerPort: "9999",
        	DocumentDBUri: "mongodb://localhost:27017/test",
        	HybridResourcesEnabled: false,
        }


### ConflictError

```go
type ConflictError interface{ ... }
```


type ConflictError interface {
	RepositoryError
	// ConflictType returns the type of conflict (e.g., "duplicate_id", "concurrent_modification")
	ConflictType() string
}
    ConflictError indicates that a create or update operation failed due
    to conflicts. This abstracts database unique constraint violations and
    concurrent modification issues.

    Instead of MongoDB duplicate key errors or optimistic concurrency failures,
    repositories should return implementations of this interface.


### CursorInterface

```go
type CursorInterface interface{ ... }
```


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
    CursorInterface abstracts MongoDB cursor operations for iteration over
    results. This enables mocking of cursor operations in repository tests.


### DIContainerInterface

```go
type DIContainerInterface interface{ ... }
```


type DIContainerInterface interface {
	// ResolveType resolves a dependency by interface type
	ResolveType(interfaceType interface{}) (interface{}, error)
}
    DIContainerInterface defines the minimum interface required for dependency
    resolution. This is a subset of the full DI container interface to avoid
    import cycles.


### DatabaseClientInterface

```go
type DatabaseClientInterface interface{ ... }
```


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
    DatabaseClientInterface abstracts database client operations for dependency
    injection. This enables alternative implementations (testing, different
    databases) and eliminates hard-coded calls to GetMongoClient() throughout
    the repository layer.

    Acceptance Criteria 3: All MongoDB operations use injected
    DatabaseClientInterface, eliminating hard-coded GetMongoClient() calls.


### DatabaseInterface

```go
type DatabaseInterface interface{ ... }
```


type DatabaseInterface interface {
	// Collection returns a collection interface for the specified collection name
	Collection(name string) CollectionInterface
	// Name returns the database name
	Name() string
	// RunCommand runs a database command
	RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) SingleResultInterface
}
    DatabaseInterface abstracts MongoDB database operations. This enables
    dependency injection of database instances and testing with mock databases.


### DatabaseRepositoryError

```go
type DatabaseRepositoryError struct{ ... }
```


type DatabaseRepositoryError struct {
	Code    string
	Message string
	Cause   error
}
    DatabaseRepositoryError defines structured error types for repository
    operations following the pattern established in previous stories.

func NewDatabaseRepositoryError(code, message string, cause error) DatabaseRepositoryError
func (e DatabaseRepositoryError) Error() string
func (e DatabaseRepositoryError) ErrorCode() string
func (e DatabaseRepositoryError) ErrorMessage() string

### EndorContext[T

```go
type EndorContext[T any] struct{ ... }
```


### EndorContextInterface[T

```go
type EndorContextInterface[T any] interface{ ... }
```


### EndorHybridServiceCategory

```go
type EndorHybridServiceCategory interface{ ... }
```


type EndorHybridServiceCategory interface {
	// GetID returns the unique identifier for this category.
	GetID() string

	// CreateDefaultActions generates the default CRUD actions for this category.
	// Uses the provided resource information and metadata schema to create
	// specialized endpoints like "categoryId/create", "categoryId/list", etc.
	CreateDefaultActions(resource string, resourceDescription string, metadataSchema Schema) map[string]EndorServiceAction
}
    EndorHybridServiceCategory represents a category-based specialization.
    Categories enable automatic CRUD operations with specialized schemas and
    behaviors.


### EndorHybridServiceInterface

```go
type EndorHybridServiceInterface interface{ ... }
```


type EndorHybridServiceInterface interface {
	// GetResource returns the resource name that this hybrid service manages.
	GetResource() string

	// GetResourceDescription returns a human-readable description of the resource.
	GetResourceDescription() string

	// GetPriority returns the service priority for conflict resolution.
	// Returns nil if no specific priority is set.
	GetPriority() *int

	// WithCategories configures the hybrid service with category-based specializations.
	// Each category provides specialized CRUD operations and schema extensions.
	// Returns a new hybrid service instance with the categories applied.
	WithCategories(categories []EndorHybridServiceCategory) EndorHybridServiceInterface

	// WithActions configures the hybrid service with custom actions beyond CRUD.
	// The provided function receives a schema getter and returns custom actions.
	// This enables dynamic action creation based on the resolved schema.
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceAction) EndorHybridServiceInterface

	// ToEndorService converts the hybrid service to a concrete EndorService.
	// This performs the final resolution of schema, categories, and actions
	// into a complete service ready for registration and routing.
	ToEndorService(metadataSchema Schema) EndorServiceInterface

	// EmbedService embeds an existing EndorService within this hybrid service with prefix-based namespacing.
	// AC 1: Service Embedding Interface - provides method for embedding existing services
	// AC 2: Method Delegation - embedded service methods accessible with configurable prefix namespacing
	// AC 7: Multiple Service Support - multiple services can be embedded with namespace isolation
	EmbedService(prefix string, service EndorServiceInterface) error

	// GetEmbeddedServices returns all embedded services for service discovery and introspection.
	// AC 1, 7: Service discovery method enabling multiple embedded services with clear method resolution
	GetEmbeddedServices() map[string]EndorServiceInterface

	// Validate performs hybrid service configuration validation.
	// Should verify that categories and actions are properly configured.
	Validate() error
}
    EndorHybridServiceInterface defines the contract for EndorHybridService
    components. This interface enables testing of hybrid services that provide
    automated CRUD operations with category-based specialization and dynamic
    schema generation.

    Hybrid services combine the automation of CRUD operations with the
    flexibility of custom actions, supporting the framework's type-safe generic
    architecture.

    Example usage:

        // Production usage
        hybridService := sdk.NewHybridService[User]("users", "User management")
        hybridService = hybridService.WithCategories([]sdk.EndorHybridServiceCategory{
        	sdk.NewEndorHybridServiceCategory[User, AdminUser](adminCategory),
        }).WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
        	return map[string]sdk.EndorServiceAction{
        		"custom": sdk.NewAction(customHandler, "Custom user action"),
        	}
        })

        // Testing usage with mock
        mockHybridService := &MockEndorHybridService{}
        mockHybridService.On("GetResource").Return("users")
        mockHybridService.On("ToEndorService", mock.Any).Return(mockEndorService)


### EndorServiceAction

```go
type EndorServiceAction interface{ ... }
```


type EndorServiceAction interface {
	// CreateHTTPCallback creates a Gin HTTP handler for this action.
	// The microserviceId is used for request tracking and service identification.
	CreateHTTPCallback(microserviceId string) func(c *gin.Context)

	// GetOptions returns the configuration options for this action.
	// Includes description, visibility, validation settings, and input schema.
	GetOptions() EndorServiceActionOptions
}
    EndorServiceAction represents an action that can be performed on a service.
    This interface abstracts the HTTP callback creation and configuration
    options.


### EndorServiceActionOptions

```go
type EndorServiceActionOptions struct{ ... }
```


type EndorServiceActionOptions struct {
	Description     string      // Human-readable description for documentation
	Public          bool        // Whether the action is publicly accessible
	ValidatePayload bool        // Whether to validate request payload
	InputSchema     *RootSchema // Schema for request validation
}
    EndorServiceActionOptions contains configuration for service actions.


### EndorServiceInterface

```go
type EndorServiceInterface interface{ ... }
```


type EndorServiceInterface interface {
	// GetResource returns the resource name that this service manages.
	// This is used for API routing and service identification.
	GetResource() string

	// GetDescription returns a human-readable description of the service.
	// Used in API documentation and service discovery.
	GetDescription() string

	// GetMethods returns the map of available actions for this service.
	// Each action corresponds to an API endpoint and handler function.
	GetMethods() map[string]EndorServiceAction

	// GetPriority returns the service priority for conflict resolution.
	// Services with higher priority take precedence during registration.
	// Returns nil if no specific priority is set.
	GetPriority() *int

	// GetVersion returns the API version for this service.
	// Used for API versioning and backward compatibility.
	// Returns empty string if not specified.
	GetVersion() string

	// Validate performs service configuration validation.
	// Should check that all required fields are properly configured
	// and that the service can be safely registered.
	Validate() error
}
    EndorServiceInterface defines the contract for EndorService components.
    This interface enables mocking and testing of services without requiring
    concrete implementations or external dependencies like databases.

    The interface follows Go's "accept interfaces, return structs" philosophy
    and supports the framework's manual control service pattern.

    Example usage:

        // Production usage with concrete implementation
        service := sdk.EndorService{
        	Resource: "users",
        	Description: "User management service",
        	Methods: map[string]sdk.EndorServiceAction{
        		"create": sdk.NewAction(createUserHandler, "Create a new user"),
        	},
        }

        // Testing usage with mock
        mockService := &MockEndorService{}
        mockService.On("GetResource").Return("users")
        mockService.On("GetDescription").Return("Mock user service")


### InternalError

```go
type InternalError interface{ ... }
```


type InternalError interface {
	RepositoryError
	// ShouldRetry indicates whether the operation can be safely retried
	ShouldRetry() bool
}
    InternalError indicates an internal repository error that should not
    be exposed to clients. This abstracts database connection errors, query
    failures, and other infrastructure issues.

    Instead of MongoDB connection errors or query syntax errors, repositories
    should return implementations of this interface.


### LogLevel

```go
type LogLevel int
```


type LogLevel int
    LogLevel represents the severity level of a log entry.

const DebugLevel LogLevel = iota ...
func (l LogLevel) String() string

### LoggerInterface

```go
type LoggerInterface interface{ ... }
```


type LoggerInterface interface {
	// Debug logs a debug-level message with optional key-value pairs.
	// Debug messages are typically used for detailed diagnostic information
	// that is only of interest when diagnosing problems.
	Debug(msg string, keysAndValues ...interface{})

	// Info logs an info-level message with optional key-value pairs.
	// Info messages are typically used for general operational entries
	// about what's happening inside the application.
	Info(msg string, keysAndValues ...interface{})

	// Warn logs a warning-level message with optional key-value pairs.
	// Warning messages are typically used for events that should be looked into
	// but don't necessarily represent errors.
	Warn(msg string, keysAndValues ...interface{})

	// Error logs an error-level message with optional key-value pairs.
	// Error messages are typically used for events that indicate something
	// went wrong but the application can continue to operate.
	Error(msg string, keysAndValues ...interface{})

	// Fatal logs a fatal-level message with optional key-value pairs.
	// Fatal messages indicate severe error conditions that require
	// immediate attention and may cause application termination.
	Fatal(msg string, keysAndValues ...interface{})

	// With creates a new logger instance with additional key-value pairs
	// that will be included in all subsequent log entries from that logger.
	// This is useful for adding context like request IDs, user IDs, etc.
	With(keysAndValues ...interface{}) LoggerInterface

	// WithName creates a new logger instance with a specific name/component identifier.
	// This helps organize log entries by component or service area.
	WithName(name string) LoggerInterface
}
    LoggerInterface defines the contract for logging in the endor-sdk-go
    framework. This interface abstracts logging operations, enabling testing
    with different log levels and outputs while supporting both structured and
    simple logging patterns.

    The interface follows standard logging practices while making logging
    testable and mockable for unit testing scenarios where log verification is
    needed.

    Example usage:

        // Production usage with concrete implementation
        logger := sdk.NewLogger() // implements LoggerInterface
        logger.Info("Service started", "port", "8080", "service", "endor-api")
        logger.Error("Database connection failed", "error", err, "database", "mongodb")

        // Testing usage with mock logger
        mockLogger := &MockLogger{}
        mockLogger.On("Info").Return()
        mockLogger.On("Error", "message", mock.Any, "error", mock.Any).Return()

        // Test logger for verification
        testLogger := &TestLogger{}
        service := NewEndorServiceWithDeps(repo, config, testLogger, ctx)
        // ... perform operations
        testLogger.AssertInfoCalled(t, "Expected log message")


### Message

```go
type Message struct{ ... }
```


type Message struct {
	Gravity MessageGravity `json:"gravity"`
	Value   string         `json:"value"`
}
    Message represents a response message with severity level.


### MessageGravity

```go
type MessageGravity string
```


type MessageGravity string
    MessageGravity indicates the severity level of a message.

const Info MessageGravity = "info" ...

### MockableRepository

```go
type MockableRepository interface{ ... }
```


type MockableRepository interface {
	TestableRepository
}
    MockableRepository defines the interface for repositories that can be
    easily mocked. This enables comprehensive unit testing without database
    dependencies.


### NoPayload

```go
type NoPayload struct{}
```


type NoPayload struct{}
    NoPayload is used for actions that don't require input payload.


### NotFoundError

```go
type NotFoundError interface{ ... }
```


type NotFoundError interface {
	RepositoryError
	// ResourceType returns the type of resource that was not found (e.g., "User", "Product")
	ResourceType() string
	// ResourceID returns the ID that was not found
	ResourceID() string
}
    NotFoundError indicates that a requested resource was not found. This
    abstracts database-specific "no documents" errors into domain language.

    Instead of mongo.ErrNoDocuments, repositories should return implementations
    of this interface to enable proper testing and error handling.


### RepositoryContainerFactoryFunc

```go
type RepositoryContainerFactoryFunc func(container DIContainerInterface) (RepositoryInterface, error)
```


type RepositoryContainerFactoryFunc func(container DIContainerInterface) (RepositoryInterface, error)
    RepositoryContainerFactoryFunc defines the signature for container-based
    repository factories. This enables automatic dependency resolution from DI
    containers.

    Acceptance Criteria 8: Container integration enables repository resolution
    by interface type.


### RepositoryDependencies

```go
type RepositoryDependencies struct{ ... }
```


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
    RepositoryDependencies defines the complete set of dependencies required for
    repository construction with dependency injection.

    Acceptance Criteria 1: Repository constructors accept
    DatabaseClientInterface and ConfigInterface instead of using global
    singletons.


### RepositoryError

```go
type RepositoryError interface{ ... }
```


type RepositoryError interface {
	error
	// ErrorCode returns a domain-specific error code for programmatic handling
	ErrorCode() string
	// ErrorMessage returns a human-readable error message
	ErrorMessage() string
}
    RepositoryError defines the base interface for all repository-related
    errors. This enables consistent error handling and testing patterns across
    implementations.

    This interface abstracts MongoDB-specific errors (like mongo.ErrNoDocuments)
    into domain-appropriate error types that can be easily tested and handled.


### RepositoryFactoryFunc

```go
type RepositoryFactoryFunc func(deps RepositoryDependencies) (RepositoryInterface, error)
```


type RepositoryFactoryFunc func(deps RepositoryDependencies) (RepositoryInterface, error)
    RepositoryFactoryFunc defines the signature for repository factory
    functions. This enables both direct construction and container-based
    resolution patterns.

    Acceptance Criteria 5: Repository Factory Patterns support
    both NewRepositoryWithClient() direct construction and
    NewRepositoryFromContainer() for DI container resolution.


### RepositoryInterface

```go
type RepositoryInterface interface{ ... }
```


type RepositoryInterface interface {
	// Create inserts a new resource into the repository
	Create(ctx context.Context, resource any) error
	// Read retrieves a resource by ID and populates the result
	Read(ctx context.Context, id string, result any) error
	// Update modifies an existing resource in the repository
	Update(ctx context.Context, resource any) error
	// Delete removes a resource from the repository by ID
	Delete(ctx context.Context, id string) error
	// List retrieves multiple resources matching the filter criteria
	List(ctx context.Context, filter map[string]any, results any) error
}
    RepositoryInterface defines the core contract for all repository
    implementations. This interface enables dependency injection, testing with
    mocks, and service composition while maintaining the framework's automatic
    CRUD capabilities.

    Acceptance Criteria 2: Repository implementations satisfy
    RepositoryInterface from interfaces package and can be mocked in tests.


### RepositoryOptions

```go
type RepositoryOptions interface{ ... }
```


type RepositoryOptions interface {
	// GetAutoGenerateID returns whether IDs should be auto-generated
	GetAutoGenerateID() bool
	// SetAutoGenerateID configures ID generation behavior
	SetAutoGenerateID(auto bool)
}
    RepositoryOptions defines configuration options that repositories should
    support. This enables consistent configuration patterns across different
    repository implementations.


### RepositoryPattern

```go
type RepositoryPattern interface{}
```


type RepositoryPattern interface {
}
    RepositoryPattern defines the general contract for CRUD repository
    operations. This interface establishes the pattern that all repositories
    should follow without defining specific method signatures that would cause
    import cycles.

    Concrete implementations in the sdk package implement the actual method
    signatures while following this pattern for testability and dependency
    injection.


### ResourceInstanceInterface

```go
type ResourceInstanceInterface interface{ ... }
```


type ResourceInstanceInterface interface {
	GetID() *string
	SetID(id string)
}
    ResourceInstanceInterface defines the contract that all resource instances
    must implement. This matches the existing interface definition in the
    framework.


### ResourceInstanceSpecializedInterface

```go
type ResourceInstanceSpecializedInterface interface{ ... }
```


type ResourceInstanceSpecializedInterface interface {
	GetCategoryType() *string
	SetCategoryType(categoryType string)
}
    ResourceInstanceSpecializedInterface defines the contract for
    category-specialized resources. This matches the existing interface
    definition in the framework.


### Response[T

```go
type Response[T any] struct{ ... }
```


### RootSchema

```go
type RootSchema struct{ ... }
```


type RootSchema struct {
	Schema
}
    RootSchema extends Schema with additional root-level metadata.


### Schema

```go
type Schema struct{ ... }
```


type Schema struct {
	Type        SchemaType         `json:"type,omitempty"`
	Properties  *map[string]Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Reference   string             `json:"$ref,omitempty"`
	Enum        *[]string          `json:"enum,omitempty"`
	Definitions map[string]Schema  `json:"definitions,omitempty"`
	Format      string             `json:"format,omitempty"`
	Default     interface{}        `json:"default,omitempty"`
	Examples    []interface{}      `json:"examples,omitempty"`
}
    Schema represents a JSON schema for data validation and documentation. Used
    throughout the framework for automatic API documentation and validation.


### SchemaType

```go
type SchemaType string
```


type SchemaType string
    SchemaType represents the type of a schema field.

const StringType SchemaType = "string" ...

### Session

```go
type Session struct{ ... }
```


type Session struct {
	Id          string // Session identifier
	Username    string // Authenticated username
	Development bool   // Development mode flag
}
    Session contains user session information for authenticated requests.


### SingleResultInterface

```go
type SingleResultInterface interface{ ... }
```


type SingleResultInterface interface {
	// Decode decodes the result into the provided value
	Decode(v interface{}) error
	// Err returns any error from the operation
	Err() error
}
    SingleResultInterface abstracts MongoDB single result operations. This
    enables mocking of single result decoding and error handling in tests.


### StructuredLoggerInterface

```go
type StructuredLoggerInterface interface{ ... }
```


type StructuredLoggerInterface interface {
	LoggerInterface

	// LogWithContext logs a message with rich contextual information.
	// This method supports complex structured data and metadata.
	LogWithContext(level LogLevel, msg string, context map[string]interface{})

	// SetLevel configures the minimum log level for this logger instance.
	// Messages below this level will be filtered out.
	SetLevel(level LogLevel)

	// GetLevel returns the current minimum log level for this logger.
	GetLevel() LogLevel
}
    StructuredLoggerInterface extends LoggerInterface for implementations that
    support structured logging with rich context and metadata.


### TestableRepository

```go
type TestableRepository interface{ ... }
```


type TestableRepository interface {
	RepositoryPattern
}
    TestableRepository defines the interface for repository implementations
    that support testing. This interface enables dependency injection and mock
    creation for unit testing.

    Repositories implementing this pattern can be easily mocked and tested
    without requiring actual database connections or external dependencies.


### TransactionInterface

```go
type TransactionInterface interface{ ... }
```


type TransactionInterface interface {
	// Commit commits the transaction
	Commit(ctx context.Context) error
	// Abort aborts the transaction
	Abort(ctx context.Context) error
	// WithTransaction executes a function within the transaction context
	WithTransaction(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error)
}
    TransactionInterface abstracts MongoDB transaction operations for dependency
    injection. This enables transaction support to work through injected client
    interfaces.

    Acceptance Criteria 7: Transaction handling works with injected database
    client interfaces.


### ValidationError

```go
type ValidationError interface{ ... }
```


type ValidationError interface {
	RepositoryError
	// ValidationDetails returns specific validation failure information
	// Key is the field name, value is the validation error message
	ValidationDetails() map[string]string
}
    ValidationError indicates that data validation failed. This abstracts
    database constraint violations and input validation failures.

    Instead of MongoDB validation errors or constraint violations, repositories
    should return implementations of this interface.


---

*Generated by [endor-sdk-go documentation generator](https://github.com/mattiabonardi/endor-sdk-go/tree/main/tools/gendocs)*
