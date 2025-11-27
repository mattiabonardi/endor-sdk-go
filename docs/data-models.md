# Endor SDK Go - Data Models & Persistence

## Overview

The Endor SDK implements a sophisticated data persistence layer using MongoDB with a generic repository pattern. The architecture supports both wrapped resource instances with metadata and specialized category-based resources.

## Core Data Architecture

### Resource Instance Pattern

All data entities in the Endor SDK follow the Resource Instance pattern, which wraps business entities with standardized metadata:

```go
type ResourceInstance[T ResourceInstanceInterface] struct {
    This     T              `bson:",inline"`
    Metadata map[string]any `bson:"metadata,omitempty"`
}
```

### Resource Instance Interface

Business entities must implement the base interface:

```go
type ResourceInstanceInterface interface {
    GetId() string
    SetId(id string)
}
```

### Specialized Resource Pattern

For category-based resources, a specialized pattern is used:

```go
type ResourceInstanceSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct {
    This         T                      `json:",inline" bson:"this"`
    CategoryThis C                      `json:",inline" bson:"categoryThis"`
    Metadata     map[string]interface{} `json:",inline" bson:"metadata,omitempty"`
}
```

## MongoDB Integration

### Connection Management

The SDK uses a singleton pattern for MongoDB connection management:

```go
type MongoClient struct {
    Client   *mongo.Client
    Database *mongo.Database
}

func GetMongo() *MongoClient {
    // Singleton implementation with automatic reconnection
}
```

### Configuration

MongoDB configuration is managed through environment variables:

```go
type ServerConfig struct {
    DocumentDBUri                 string // MongoDB connection string
    DynamicResourceDocumentDBName string // Database name for dynamic resources
    HybridResourcesEnabled        bool   // Enable MongoDB features
    DynamicResourcesEnabled       bool   // Enable runtime resource creation
}
```

### Collection Strategy

- **One collection per resource type**: Each service resource maps to a dedicated MongoDB collection
- **Collection naming**: Collections are named after the service resource identifier
- **Index management**: Automatic index creation for ID fields and common query patterns

## Repository Interfaces

### Standard Repository Interface

```go
type ResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface {
    Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstance[T], error)
    List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error)
    Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error)
    Delete(ctx context.Context, dto ReadInstanceDTO) error
    Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T]]) (*ResourceInstance[T], error)
}
```

### Specialized Repository Interface

For category-based resources:

```go
type ResourceInstanceSpecializedRepositoryInterface[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] interface {
    Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstanceSpecialized[T, C], error)
    List(ctx context.Context, dto ReadDTO) ([]ResourceInstanceSpecialized[T, C], error)
    Create(ctx context.Context, dto CreateDTO[ResourceInstanceSpecialized[T, C]]) (*ResourceInstanceSpecialized[T, C], error)
    Delete(ctx context.Context, dto ReadInstanceDTO) error
    Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstanceSpecialized[T, C]]) (*ResourceInstanceSpecialized[T, C], error)
}
```

### Static Resource Repository Interface

For read-only static resources:

```go
type StaticResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface {
    Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstance[T], error)
    List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error)
}
```

## Repository Implementations

### MongoDB Repository

```go
type MongoResourceInstanceRepository[T ResourceInstanceInterface] struct {
    Collection *mongo.Collection
}

func NewMongoResourceInstanceRepository[T ResourceInstanceInterface](
    collectionName string,
) ResourceInstanceRepositoryInterface[T] {
    return &MongoResourceInstanceRepository[T]{
        Collection: GetMongo().Database.Collection(collectionName),
    }
}
```

### MongoDB Specialized Repository

```go
type MongoResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct {
    Collection *mongo.Collection
}
```

## Data Transfer Objects (DTOs)

### Read Operations

```go
type ReadDTO struct {
    Limit  *int64  `json:"limit" bson:"limit"`
    Skip   *int64  `json:"skip" bson:"skip"`
    Sort   *string `json:"sort" bson:"sort"`
}

type ReadInstanceDTO struct {
    Id string `json:"id" bson:"id" validate:"required"`
}
```

### Write Operations

```go
type CreateDTO[T any] struct {
    Value T `json:"value" bson:"value" validate:"required"`
}

type UpdateByIdDTO[T any] struct {
    Id    string `json:"id" bson:"id" validate:"required"`
    Value T      `json:"value" bson:"value" validate:"required"`
}
```

## ID Management Strategies

### Automatic ID Generation

The SDK supports multiple ID generation strategies:

#### MongoDB ObjectID (Default)
```go
func GenerateObjectId() string {
    return primitive.NewObjectID().Hex()
}
```

#### Custom String IDs
```go
func (entity *CustomEntity) SetId(id string) {
    entity.ID = id
}

func (entity *CustomEntity) GetId() string {
    return entity.ID
}
```

### ID Validation

- **ObjectID Format**: 24-character hexadecimal strings
- **Custom Validation**: Configurable through struct tags
- **Unique Constraints**: Enforced at database level

## Metadata Management

### Standard Metadata Fields

Every resource instance includes standard metadata:

```go
type Metadata struct {
    CreatedAt   time.Time              `bson:"createdAt"`
    UpdatedAt   time.Time              `bson:"updatedAt"`
    CreatedBy   string                 `bson:"createdBy"`
    UpdatedBy   string                 `bson:"updatedBy"`
    Version     int                    `bson:"version"`
    Tags        []string               `bson:"tags,omitempty"`
    Custom      map[string]interface{} `bson:"custom,omitempty"`
}
```

### Automatic Metadata Population

- **Creation timestamp**: Automatically set on resource creation
- **Update timestamp**: Updated on every modification
- **User tracking**: Populated from session context
- **Version control**: Incremented on updates for optimistic locking

## Query Patterns

### Basic Queries

```go
// List with pagination
dto := ReadDTO{
    Limit: int64Ptr(20),
    Skip:  int64Ptr(40),
    Sort:  stringPtr("-createdAt"), // Descending by creation date
}
results, err := repository.List(ctx, dto)

// Find by ID
instanceDTO := ReadInstanceDTO{Id: "64a7b8c9d1e2f3a4b5c6d7e8"}
resource, err := repository.Instance(ctx, instanceDTO)
```

### Advanced Query Support

The MongoDB implementation supports:

- **Sorting**: Ascending (`+field`) and descending (`-field`)
- **Pagination**: Limit and skip parameters
- **Filtering**: Through metadata fields
- **Text Search**: On indexed text fields
- **Aggregation**: Via custom repository methods

## Category System Data Model

### Category Definition

```go
type EndorHybridServiceCategory struct {
    CategoryId          string     `json:"categoryId" bson:"categoryId"`
    CategoryDescription string     `json:"categoryDescription" bson:"categoryDescription"`
    CategorySchema      RootSchema `json:"categorySchema" bson:"categorySchema"`
}
```

### Category-Specialized Resources

```go
type ProductCategory struct {
    Type        string  `json:"type" bson:"type"`
    Commission  float64 `json:"commission" bson:"commission"`
    Restrictions []string `json:"restrictions" bson:"restrictions"`
}

type Product struct {
    ID          string  `json:"id" bson:"_id"`
    Name        string  `json:"name" bson:"name"`
    Description string  `json:"description" bson:"description"`
    Price       float64 `json:"price" bson:"price"`
}

// Combined specialized resource
type SpecializedProduct = ResourceInstanceSpecialized[Product, ProductCategory]
```

## Schema System Integration

### Dynamic Schema Storage

Schemas are stored alongside resources and can be dynamically updated:

```go
type RootSchema struct {
    Type        string                    `json:"type" bson:"type"`
    Properties  map[string]PropertySchema `json:"properties" bson:"properties"`
    Required    []string                  `json:"required" bson:"required"`
    Title       string                    `json:"title,omitempty" bson:"title,omitempty"`
    Description string                    `json:"description,omitempty" bson:"description,omitempty"`
}
```

### Schema Validation

- **Pre-write validation**: Payload validated against schema before persistence
- **Schema evolution**: Support for schema migration and versioning
- **Type safety**: Compile-time type checking with runtime validation

## Database Operations

### Transaction Support

```go
func (r *MongoResourceInstanceRepository[T]) WithTransaction(
    ctx context.Context,
    operation func(mongo.SessionContext) error,
) error {
    session, err := GetMongo().Client.StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(ctx)
    
    return mongo.WithSession(ctx, session, operation)
}
```

### Error Handling

```go
type DatabaseError struct {
    Operation string
    Resource  string
    Cause     error
}

func (e *DatabaseError) Error() string {
    return fmt.Sprintf("Database %s failed for %s: %v", 
        e.Operation, e.Resource, e.Cause)
}
```

## Performance Considerations

### Connection Pooling

```go
clientOptions := options.Client().
    ApplyURI(config.DocumentDBUri).
    SetMaxPoolSize(100).
    SetMinPoolSize(10).
    SetMaxConnIdleTime(30 * time.Second)
```

### Indexing Strategy

- **Primary keys**: Automatic unique index on ID field
- **Query optimization**: Indexes on frequently queried metadata fields
- **Text search**: Full-text indexes for searchable content
- **Compound indexes**: For complex query patterns

### Query Optimization

- **Projection**: Return only required fields
- **Aggregation pipelines**: For complex data transformations
- **Caching**: Repository-level caching for frequently accessed data
- **Batch operations**: Bulk insert/update capabilities

## Testing Strategies

### Repository Testing

```go
func TestMongoRepository(t *testing.T) {
    // Setup test database
    testDB := setupTestDatabase()
    defer testDB.Cleanup()
    
    repository := NewMongoResourceInstanceRepository[TestEntity]("test_collection")
    
    // Test CRUD operations
    entity := TestEntity{Name: "Test"}
    created, err := repository.Create(context.Background(), CreateDTO[ResourceInstance[TestEntity]]{
        Value: ResourceInstance[TestEntity]{This: entity},
    })
    
    assert.NoError(t, err)
    assert.NotEmpty(t, created.This.GetId())
}
```

### Integration Testing

- **In-memory MongoDB**: Using `testcontainers` for integration tests
- **Schema validation**: Testing schema generation and validation
- **Transaction testing**: Verifying ACID properties
- **Performance testing**: Load testing with realistic data volumes

---

*This data model architecture provides a robust foundation for building scalable microservices with flexible schema management and type-safe operations.*