# SDK Core

> Package documentation for SDK Core

**Import Path:** `github.com/mattiabonardi/endor-sdk-go/sdk`
**Generated:** 2025-12-01 10:07:47 UTC

---

const COLLECTION_RESOURCES = "resources"
func CreateSwaggerConfiguration(microServiceId string, microServiceAddress string, services []EndorService, ...) (string, error)
func DefaultDatabaseClient() (interfaces.DatabaseClientInterface, error)
func GetMongoClient() (*mongo.Client, error)
func InitializeApiGatewayConfiguration(microServiceId string, microServiceAddress string, services []EndorService) error
func NewDefaultConfigProvider() interfaces.ConfigProviderInterface
func NewDefaultDatabaseClient() (interfaces.DatabaseClientInterface, error)
func NewDefaultLogger() interfaces.LoggerInterface
func NewDefaultRepositoryDependencies(microServiceID string) (interfaces.RepositoryDependencies, error)
func NewEndorRepositoryWithDependencies(deps interfaces.RepositoryDependencies) (interfaces.RepositoryInterface, error)
func NewMongoDatabaseClient(client *mongo.Client) interfaces.DatabaseClientInterface
func NewRepositoryFromContainer(container di.Container) (interfaces.RepositoryInterface, error)
func NewRepositoryWithClient(client interfaces.DatabaseClientInterface, ...) (interfaces.RepositoryInterface, error)
func RegisterRepositoryFactories(container di.Container) error
type ApiGatewayConfiguration struct{ ... }
type ApiGatewayConfigurationHTTP struct{ ... }
type ApiGatewayConfigurationLoadBalancer struct{ ... }
type ApiGatewayConfigurationRouter struct{ ... }
type ApiGatewayConfigurationServer struct{ ... }
type ApiGatewayConfigurationService struct{ ... }
type Category struct{ ... }
type CreateDTO[T any] struct{ ... }
type DSLDAO struct{ ... }
    func NewDSLDAO(username string, development bool) *DSLDAO
type DecoratedService struct{ ... }
type DefaultConfigProvider struct{}
type DefaultLogger struct{ ... }
type DynamicResource struct{ ... }
type DynamicResourceSpecialized struct{ ... }
type Endor struct{ ... }
type EndorContext[T any] struct{ ... }
type EndorError struct{ ... }
    func NewBadRequestError(err error) *EndorError
    func NewConflictError(err error) *EndorError
    func NewForbiddenError(err error) *EndorError
    func NewGenericError(status int, err error) *EndorError
    func NewInternalServerError(err error) *EndorError
    func NewNotFoundError(err error) *EndorError
    func NewUnauthorizedError(err error) *EndorError
type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)
type EndorHybridService interface{ ... }
    func NewHybridService[T ResourceInstanceInterface](resource, resourceDescription string) EndorHybridService
type EndorHybridServiceCategory interface{ ... }
    func NewEndorHybridServiceCategory[T ResourceInstanceInterface, R ResourceInstanceSpecializedInterface](category Category) EndorHybridServiceCategory
type EndorHybridServiceCategoryImpl[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct{ ... }
type EndorHybridServiceDependencies struct{ ... }
type EndorHybridServiceError struct{ ... }
type EndorHybridServiceImpl[T ResourceInstanceInterface] struct{ ... }
    func NewEndorHybridServiceFromContainer[T ResourceInstanceInterface](container di.Container, resource string, resourceDescription string) (*EndorHybridServiceImpl[T], error)
    func NewEndorHybridServiceWithDeps[T ResourceInstanceInterface](resource string, resourceDescription string, ...) (*EndorHybridServiceImpl[T], error)
type EndorInitializer struct{ ... }
    func NewEndorInitializer() *EndorInitializer
type EndorInitializerDependencies struct{ ... }
type EndorInitializerError struct{ ... }
type EndorRepositoryAdapter struct{ ... }
type EndorService struct{ ... }
    func NewEndorService(resource string, description string, methods map[string]EndorServiceAction) EndorService
    func NewEndorServiceFromContainer(container di.Container, resource string, description string, ...) (*EndorService, error)
    func NewEndorServiceWithDeps(resource string, description string, methods map[string]EndorServiceAction, ...) (*EndorService, error)
    func NewResourceActionService(microServiceId string, services *[]EndorService, ...) *EndorService
    func NewResourceService(microServiceId string, services *[]EndorService, ...) *EndorService
type EndorServiceAction interface{ ... }
    func NewAction[T any, R any](handler EndorHandlerFunc[T, R], description string) EndorServiceAction
    func NewConfigurableAction[T any, R any](options EndorServiceActionOptions, handler EndorHandlerFunc[T, R]) EndorServiceAction
type EndorServiceActionDictionary struct{ ... }
type EndorServiceActionOptions struct{ ... }
type EndorServiceDependencies struct{ ... }
type EndorServiceDictionary struct{ ... }
type EndorServiceError struct{ ... }
type EndorServiceRepository struct{ ... }
    func NewEndorServiceRepository(microServiceId string, internalEndorServices *[]EndorService, ...) *EndorServiceRepository
    func NewEndorServiceRepositoryFromContainer(container interfaces.DIContainerInterface, ...) (*EndorServiceRepository, error)
    func NewEndorServiceRepositoryWithDependencies(deps interfaces.RepositoryDependencies, internalEndorServices *[]EndorService, ...) (*EndorServiceRepository, error)
type Message struct{ ... }
    func NewMessage(gravity MessageGravity, value string) Message
type MessageGravity string
    const Info MessageGravity = "Info" ...
type Meta struct{ ... }
type MongoCollectionAdapter struct{ ... }
type MongoCursorAdapter struct{ ... }
type MongoDatabaseAdapter struct{ ... }
type MongoDatabaseClient struct{ ... }
type MongoResourceInstanceRepository[T ResourceInstanceInterface] struct{ ... }
    func NewMongoResourceInstanceRepository[T ResourceInstanceInterface](resourceId string, options ResourceInstanceRepositoryOptions) *MongoResourceInstanceRepository[T]
type MongoResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct{ ... }
    func NewMongoResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](resourceId string, options ResourceInstanceRepositoryOptions) *MongoResourceInstanceSpecializedRepository[T, C]
type MongoSingleResultAdapter struct{ ... }
type MongoStaticResourceInstanceRepository[T ResourceInstanceInterface] struct{ ... }
    func NewMongoStaticResourceInstanceRepository[T ResourceInstanceInterface](resourceId string, options StaticResourceInstanceRepositoryOptions) *MongoStaticResourceInstanceRepository[T]
type MongoTransactionAdapter struct{ ... }
type NoPayload struct{}
type ObjectID primitive.ObjectID
    func NewObjectID() ObjectID
type OpenAPIConfiguration struct{ ... }
    func CreateSwaggerDefinition(microServiceId string, microServiceAddress string, services []EndorService, ...) (OpenAPIConfiguration, error)
type OpenAPIInfo struct{ ... }
type OpenAPIMediaType struct{ ... }
type OpenAPIOperation struct{ ... }
type OpenAPIParameter struct{ ... }
type OpenAPIRequestBody struct{ ... }
type OpenAPIServer struct{ ... }
type OpenAPITag struct{ ... }
type OpenApiAuth struct{ ... }
type OpenApiComponents struct{ ... }
type OpenApiResponse struct{ ... }
type OpenApiResponses map[string]OpenApiResponse
type Presentation struct{ ... }
type ReadDTO struct{ ... }
type ReadInstanceDTO struct{ ... }
type Resource struct{ ... }
type ResourceAction struct{ ... }
type ResourceActionService struct{ ... }
type ResourceInstance[T ResourceInstanceInterface] struct{ ... }
type ResourceInstanceInterface interface{ ... }
type ResourceInstanceRepository[T ResourceInstanceInterface] struct{ ... }
    func NewResourceInstanceRepository[T ResourceInstanceInterface](resourceId string, options ResourceInstanceRepositoryOptions) *ResourceInstanceRepository[T]
type ResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface{ ... }
type ResourceInstanceRepositoryOptions struct{ ... }
type ResourceInstanceSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct{ ... }
type ResourceInstanceSpecializedInterface interface{ ... }
type ResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct{ ... }
    func NewResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](resourceId string, options ResourceInstanceRepositoryOptions) *ResourceInstanceSpecializedRepository[T, C]
type ResourceInstanceSpecializedRepositoryInterface[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] interface{ ... }
type ResourceService struct{ ... }
type Response[T any] struct{ ... }
type ResponseBuilder[T any] struct{ ... }
    func NewDefaultResponseBuilder() *ResponseBuilder[map[string]any]
    func NewResponseBuilder[T any]() *ResponseBuilder[T]
type RootSchema struct{ ... }
    func NewSchema(model any) *RootSchema
    func NewSchemaByType(t reflect.Type) *RootSchema
type Schema struct{ ... }
type SchemaFormatName string
    const DateTimeFormat SchemaFormatName = "date-time" ...
    func NewSchemaFormat(f SchemaFormatName) *SchemaFormatName
type SchemaTypeName string
    const StringType SchemaTypeName = "string" ...
type ServerConfig struct{ ... }
    func GetConfig() *ServerConfig
type Session struct{ ... }
type StaticResourceInstanceRepository[T ResourceInstanceInterface] struct{ ... }
    func NewStaticResourceInstanceRepository[T ResourceInstanceInterface](resourceId string, options StaticResourceInstanceRepositoryOptions) *StaticResourceInstanceRepository[T]
type StaticResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface{ ... }
type StaticResourceInstanceRepositoryOptions struct{ ... }
type UISchema struct{ ... }
type UpdateByIdDTO[T any] struct{ ... }

## Package Overview

package sdk // import "github.com/mattiabonardi/endor-sdk-go/sdk"

Package sdk provides the MongoDB implementation of database client
interfaces. This implementation wraps the MongoDB driver to satisfy the
DatabaseClientInterface while maintaining all existing functionality and
performance characteristics.

Package sdk provides repository factory functions for dependency injection.
These factories enable both direct construction and container-based resolution
patterns for repository instances.

const COLLECTION_RESOURCES = "resources"
func CreateSwaggerConfiguration(microServiceId string, microServiceAddress string, services []EndorService, ...) (string, error)
func DefaultDatabaseClient() (interfaces.DatabaseClientInterface, error)
func GetMongoClient() (*mongo.Client, error)
func InitializeApiGatewayConfiguration(microServiceId string, microServiceAddress string, services []EndorService) error
func NewDefaultConfigProvider() interfaces.ConfigProviderInterface
func NewDefaultDatabaseClient() (interfaces.DatabaseClientInterface, error)
func NewDefaultLogger() interfaces.LoggerInterface
func NewDefaultRepositoryDependencies(microServiceID string) (interfaces.RepositoryDependencies, error)

## Exported Types

### CreateSwaggerConfiguration

```go
func CreateSwaggerConfiguration(microServiceId string, microServiceAddress string, services []EndorService, ...) (string, error)
```

### DefaultDatabaseClient

```go
func DefaultDatabaseClient() (interfaces.DatabaseClientInterface, error)
```

### GetMongoClient

```go
func GetMongoClient() (*mongo.Client, error)
```

### InitializeApiGatewayConfiguration

```go
func InitializeApiGatewayConfiguration(microServiceId string, microServiceAddress string, services []EndorService) error
```

### NewDefaultConfigProvider

```go
func NewDefaultConfigProvider() interfaces.ConfigProviderInterface
```

### NewDefaultDatabaseClient

```go
func NewDefaultDatabaseClient() (interfaces.DatabaseClientInterface, error)
```

### NewDefaultLogger

```go
func NewDefaultLogger() interfaces.LoggerInterface
```

### NewDefaultRepositoryDependencies

```go
func NewDefaultRepositoryDependencies(microServiceID string) (interfaces.RepositoryDependencies, error)
```

### NewEndorRepositoryWithDependencies

```go
func NewEndorRepositoryWithDependencies(deps interfaces.RepositoryDependencies) (interfaces.RepositoryInterface, error)
```

### NewMongoDatabaseClient

```go
func NewMongoDatabaseClient(client *mongo.Client) interfaces.DatabaseClientInterface
```

### NewRepositoryFromContainer

```go
func NewRepositoryFromContainer(container di.Container) (interfaces.RepositoryInterface, error)
```

### NewRepositoryWithClient

```go
func NewRepositoryWithClient(client interfaces.DatabaseClientInterface, ...) (interfaces.RepositoryInterface, error)
```

### RegisterRepositoryFactories

```go
func RegisterRepositoryFactories(container di.Container) error
```

### ApiGatewayConfiguration

```go
type ApiGatewayConfiguration struct{ ... }
```


type ApiGatewayConfiguration struct {
	HTTP ApiGatewayConfigurationHTTP `yaml:"http"`
}


### ApiGatewayConfigurationHTTP

```go
type ApiGatewayConfigurationHTTP struct{ ... }
```


type ApiGatewayConfigurationHTTP struct {
	Routers  map[string]ApiGatewayConfigurationRouter  `yaml:"routers"`
	Services map[string]ApiGatewayConfigurationService `yaml:"services"`
}


### ApiGatewayConfigurationLoadBalancer

```go
type ApiGatewayConfigurationLoadBalancer struct{ ... }
```


type ApiGatewayConfigurationLoadBalancer struct {
	Servers []ApiGatewayConfigurationServer `yaml:"servers"`
}


### ApiGatewayConfigurationRouter

```go
type ApiGatewayConfigurationRouter struct{ ... }
```


type ApiGatewayConfigurationRouter struct {
	Rule        string    `yaml:"rule"`
	Service     string    `yaml:"service"`
	Priority    *int      `yaml:"priority,omitempty"`
	EntryPoints []string  `yaml:"entryPoints"`
	Middlewares *[]string `yaml:"middlewares,omitempty"`
}


### ApiGatewayConfigurationServer

```go
type ApiGatewayConfigurationServer struct{ ... }
```


type ApiGatewayConfigurationServer struct {
	URL string `yaml:"url"`
}


### ApiGatewayConfigurationService

```go
type ApiGatewayConfigurationService struct{ ... }
```


type ApiGatewayConfigurationService struct {
	LoadBalancer ApiGatewayConfigurationLoadBalancer `yaml:"loadBalancer"`
}


### Category

```go
type Category struct{ ... }
```


type Category struct {
	ID                   string `json:"id" bson:"id" schema:"title=Category ID"`
	Description          string `json:"description" bson:"description" schema:"title=Category Description"`
	AdditionalAttributes string `json:"additionalAttributes" bson:"additionalAttributes" schema:"title=Additional category attributes schema,format=yaml"`
}
    Category rappresenta una categoria con attributi dinamici specifici

func (c *Category) UnmarshalAdditionalAttributes() (*RootSchema, error)

### CreateDTO[T

```go
type CreateDTO[T any] struct{ ... }
```


### DSLDAO

```go
type DSLDAO struct{ ... }
```


type DSLDAO struct {
	BasePath string
}

func NewDSLDAO(username string, development bool) *DSLDAO
func (dao *DSLDAO) Create(fileName string, content string) error
func (dao *DSLDAO) Delete(fileName string) error
func (dao *DSLDAO) GetBasePath() string
func (dao *DSLDAO) Instace(fileName string) (string, error)
func (dao *DSLDAO) ListAll() ([]string, error)
func (dao *DSLDAO) ListFile() ([]string, error)
func (dao *DSLDAO) ListFolders() ([]string, error)
func (dao *DSLDAO) Rename(oldName, newName string) error
func (dao *DSLDAO) Update(fileName string, content string) error

### DecoratedService

```go
type DecoratedService struct{ ... }
```


type DecoratedService struct {
	*EndorService // Embedded service for method delegation
	// Has unexported fields.
}
    DecoratedService wraps an EndorService with middleware pipeline support.
    This type implements the decorator pattern to add cross-cutting concerns
    without modifying the original service implementation.

func (d *DecoratedService) GetMethods() map[string]EndorServiceAction

### DefaultConfigProvider

```go
type DefaultConfigProvider struct{}
```


type DefaultConfigProvider struct{}
    DefaultConfigProvider adapts the singleton ServerConfig to implement
    ConfigProviderInterface This provides backward compatibility while enabling
    dependency injection.

func (d *DefaultConfigProvider) GetDocumentDBUri() string
func (d *DefaultConfigProvider) GetDynamicResourceDocumentDBName() string
func (d *DefaultConfigProvider) GetServerPort() string
func (d *DefaultConfigProvider) IsDynamicResourcesEnabled() bool
func (d *DefaultConfigProvider) IsHybridResourcesEnabled() bool
func (d *DefaultConfigProvider) Reload() error
func (d *DefaultConfigProvider) Validate() error

### DefaultLogger

```go
type DefaultLogger struct{ ... }
```


type DefaultLogger struct {
	// Has unexported fields.
}
    DefaultLogger adapts Go's standard log package to implement LoggerInterface
    This provides backward compatibility while enabling dependency injection.

func (d *DefaultLogger) Debug(msg string, keysAndValues ...interface{})
func (d *DefaultLogger) Error(msg string, keysAndValues ...interface{})
func (d *DefaultLogger) Fatal(msg string, keysAndValues ...interface{})
func (d *DefaultLogger) Info(msg string, keysAndValues ...interface{})
func (d *DefaultLogger) Warn(msg string, keysAndValues ...interface{})
func (d *DefaultLogger) With(keysAndValues ...interface{}) interfaces.LoggerInterface
func (d *DefaultLogger) WithName(name string) interfaces.LoggerInterface

### DynamicResource

```go
type DynamicResource struct{ ... }
```


type DynamicResource struct {
	Id          string `json:"id" bson:"_id" schema:"title=Id,readOnly=true" ui-schema:"hidden=true"`
	Description string `json:"description" bson:"description" schema:"title=Description"`
}

func (h *DynamicResource) GetID() *string
func (h *DynamicResource) SetID(id string)

### DynamicResourceSpecialized

```go
type DynamicResourceSpecialized struct{ ... }
```


type DynamicResourceSpecialized struct {
	CategoryType string `json:"categoryType" bson:"categoryType" schema:"title=Type,readOnly=true"`
}

func (h *DynamicResourceSpecialized) GetCategoryType() *string
func (h *DynamicResourceSpecialized) SetCategoryType(categoryType string)

### Endor

```go
type Endor struct{ ... }
```


type Endor struct {
	// Has unexported fields.
}

func (h *Endor) Init(microserviceId string)
func (h *Endor) Shutdown() error

### EndorContext[T

```go
type EndorContext[T any] struct{ ... }
```


### EndorError

```go
type EndorError struct{ ... }
```


type EndorError struct {
	StatusCode  int
	InternalErr error
}

func NewBadRequestError(err error) *EndorError
func NewConflictError(err error) *EndorError
func NewForbiddenError(err error) *EndorError
func NewGenericError(status int, err error) *EndorError
func NewInternalServerError(err error) *EndorError
func NewNotFoundError(err error) *EndorError
func NewUnauthorizedError(err error) *EndorError
func (e *EndorError) Error() string
func (e *EndorError) Unwrap() error

### EndorHandlerFunc[T

```go
type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)
```


### EndorHybridService

```go
type EndorHybridService interface{ ... }
```


type EndorHybridService interface {
	GetResource() string
	GetResourceDescription() string
	GetPriority() *int
	WithCategories(categories []EndorHybridServiceCategory) EndorHybridService
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceAction) EndorHybridService
	WithMiddleware(middleware ...middleware.MiddlewareInterface) EndorHybridService
	ToEndorService(metadataSchema Schema) EndorService
	// AC 1: Service Embedding Interface - EndorHybridService provides EmbedService method
	EmbedService(prefix string, service interfaces.EndorServiceInterface) error
	// AC 1: Service discovery method for embedded services
	GetEmbeddedServices() map[string]interfaces.EndorServiceInterface
}

func NewHybridService[T ResourceInstanceInterface](resource, resourceDescription string) EndorHybridService

### EndorHybridServiceCategory

```go
type EndorHybridServiceCategory interface{ ... }
```


type EndorHybridServiceCategory interface {
	GetID() string
	CreateDefaultActions(resource string, resourceDescription string, metadataSchema Schema) map[string]EndorServiceAction
}

func NewEndorHybridServiceCategory[T ResourceInstanceInterface, R ResourceInstanceSpecializedInterface](category Category) EndorHybridServiceCategory

### EndorHybridServiceCategoryImpl[T

```go
type EndorHybridServiceCategoryImpl[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct{ ... }
```


### EndorHybridServiceDependencies

```go
type EndorHybridServiceDependencies struct{ ... }
```


type EndorHybridServiceDependencies struct {
	Repository interfaces.RepositoryPattern       // Required: Data access layer
	Config     interfaces.ConfigProviderInterface // Required: Configuration access
	Logger     interfaces.LoggerInterface         // Required: Logging interface
}
    EndorHybridServiceDependencies contains all required dependencies for
    EndorHybridService. This struct provides type safety and clear documentation
    of service dependencies.


### EndorHybridServiceError

```go
type EndorHybridServiceError struct{ ... }
```


type EndorHybridServiceError struct {
	Field   string // Name of the missing dependency field
	Message string // Human-readable error message
}
    EndorHybridServiceError represents dependency validation errors for
    EndorHybridService construction.

func (e *EndorHybridServiceError) Error() string

### EndorHybridServiceImpl[T

```go
type EndorHybridServiceImpl[T ResourceInstanceInterface] struct{ ... }
```


### EndorInitializer

```go
type EndorInitializer struct{ ... }
```


type EndorInitializer struct {
	// Has unexported fields.
}

func NewEndorInitializer() *EndorInitializer
func (b *EndorInitializer) Build() (*Endor, error)
func (b *EndorInitializer) WithContainer(container di.Container) *EndorInitializer
func (b *EndorInitializer) WithCustomConfig(config interfaces.ConfigProviderInterface) *EndorInitializer
func (b *EndorInitializer) WithCustomRepository(repo interfaces.RepositoryInterface) *EndorInitializer
func (b *EndorInitializer) WithEndorServices(services *[]EndorService) *EndorInitializer
func (b *EndorInitializer) WithHybridServices(services *[]EndorHybridService) *EndorInitializer
func (b *EndorInitializer) WithPostInitFunc(f func()) *EndorInitializer

### EndorInitializerDependencies

```go
type EndorInitializerDependencies struct{ ... }
```


type EndorInitializerDependencies struct {
	Container  di.Container                       // Optional: Custom DI container, uses default if nil
	Repository interfaces.RepositoryInterface     // Optional: Custom repository, resolved from container if nil
	Config     interfaces.ConfigProviderInterface // Optional: Custom config, uses default if nil
	Logger     interfaces.LoggerInterface         // Optional: Custom logger, uses default if nil
}
    EndorInitializerDependencies contains all required dependencies for
    EndorInitializer. This struct provides type safety and clear documentation
    of initializer dependencies.


### EndorInitializerError

```go
type EndorInitializerError struct{ ... }
```


type EndorInitializerError struct {
	Field   string // Name of the problematic dependency field
	Message string // Human-readable error message
	Cause   error  // Underlying error if any
}
    EndorInitializerError represents dependency validation errors for
    EndorInitializer construction.

func (e *EndorInitializerError) Error() string

### EndorRepositoryAdapter

```go
type EndorRepositoryAdapter struct{ ... }
```


type EndorRepositoryAdapter struct {
	// Has unexported fields.
}
    EndorRepositoryAdapter adapts EndorServiceRepository to implement
    RepositoryInterface. This enables the EndorServiceRepository to be used
    through the standard repository interface.

func (r *EndorRepositoryAdapter) Create(ctx context.Context, resource any) error
func (r *EndorRepositoryAdapter) Delete(ctx context.Context, id string) error
func (r *EndorRepositoryAdapter) List(ctx context.Context, filter map[string]any, results any) error
func (r *EndorRepositoryAdapter) Read(ctx context.Context, id string, result any) error
func (r *EndorRepositoryAdapter) Update(ctx context.Context, resource any) error

### EndorService

```go
type EndorService struct{ ... }
```


type EndorService struct {
	Resource         string
	Description      string
	Methods          map[string]EndorServiceAction
	Priority         *int
	ResourceMetadata bool

	// optionals
	Version string

	// Has unexported fields.
}

func NewEndorService(resource string, description string, methods map[string]EndorServiceAction) EndorService
func NewEndorServiceFromContainer(container di.Container, resource string, description string, ...) (*EndorService, error)
func NewEndorServiceWithDeps(resource string, description string, methods map[string]EndorServiceAction, ...) (*EndorService, error)
func NewResourceActionService(microServiceId string, services *[]EndorService, ...) *EndorService
func NewResourceService(microServiceId string, services *[]EndorService, ...) *EndorService
func (s *EndorService) GetConfig() interfaces.ConfigProviderInterface
func (s *EndorService) GetDescription() string
func (s *EndorService) GetLogger() interfaces.LoggerInterface
func (s *EndorService) GetMethods() map[string]EndorServiceAction
func (s *EndorService) GetPriority() *int
func (s *EndorService) GetRepository() interfaces.RepositoryPattern
func (s *EndorService) GetResource() string
func (s *EndorService) GetVersion() string
func (s *EndorService) Validate() error
func (s *EndorService) WithMiddleware(middlewares ...middleware.MiddlewareInterface) *DecoratedService

### EndorServiceAction

```go
type EndorServiceAction interface{ ... }
```


type EndorServiceAction interface {
	CreateHTTPCallback(microserviceId string) func(c *gin.Context)
	GetOptions() EndorServiceActionOptions
}

func NewAction[T any, R any](handler EndorHandlerFunc[T, R], description string) EndorServiceAction
func NewConfigurableAction[T any, R any](options EndorServiceActionOptions, handler EndorHandlerFunc[T, R]) EndorServiceAction

### EndorServiceActionDictionary

```go
type EndorServiceActionDictionary struct{ ... }
```


type EndorServiceActionDictionary struct {
	EndorServiceAction EndorServiceAction
	// Has unexported fields.
}


### EndorServiceActionOptions

```go
type EndorServiceActionOptions struct{ ... }
```


type EndorServiceActionOptions struct {
	Description     string
	Public          bool
	ValidatePayload bool
	InputSchema     *RootSchema
}


### EndorServiceDependencies

```go
type EndorServiceDependencies struct{ ... }
```


type EndorServiceDependencies struct {
	Repository interfaces.RepositoryPattern       // Required: Data access layer
	Config     interfaces.ConfigProviderInterface // Required: Configuration access
	Logger     interfaces.LoggerInterface         // Required: Logging interface
}
    EndorServiceDependencies contains all required dependencies for
    EndorService. This struct provides type safety and clear documentation of
    service dependencies.


### EndorServiceDictionary

```go
type EndorServiceDictionary struct{ ... }
```


type EndorServiceDictionary struct {
	EndorService EndorService
	// Has unexported fields.
}


### EndorServiceError

```go
type EndorServiceError struct{ ... }
```


type EndorServiceError struct {
	Field   string // Name of the missing dependency field
	Message string // Human-readable error message
}
    EndorServiceError represents dependency validation errors for EndorService
    construction.

func (e *EndorServiceError) Error() string

### EndorServiceRepository

```go
type EndorServiceRepository struct{ ... }
```


type EndorServiceRepository struct {
	// Has unexported fields.
}

func NewEndorServiceRepository(microServiceId string, internalEndorServices *[]EndorService, ...) *EndorServiceRepository
func NewEndorServiceRepositoryFromContainer(container interfaces.DIContainerInterface, ...) (*EndorServiceRepository, error)
func NewEndorServiceRepositoryWithDependencies(deps interfaces.RepositoryDependencies, internalEndorServices *[]EndorService, ...) (*EndorServiceRepository, error)
func (h *EndorServiceRepository) ActionInstance(dto ReadInstanceDTO) (*EndorServiceActionDictionary, error)
func (h *EndorServiceRepository) ActionMap() (map[string]EndorServiceActionDictionary, error)
func (h *EndorServiceRepository) Create(dto CreateDTO[Resource]) error
func (h *EndorServiceRepository) DeleteOne(dto ReadInstanceDTO) error
func (h *EndorServiceRepository) DynamiResourceList() ([]Resource, error)
func (h *EndorServiceRepository) EndorServiceList() ([]EndorService, error)
func (h *EndorServiceRepository) Instance(dto ReadInstanceDTO) (*EndorServiceDictionary, error)
func (h *EndorServiceRepository) Map() (map[string]EndorServiceDictionary, error)
func (h *EndorServiceRepository) ResourceActionList() ([]ResourceAction, error)
func (h *EndorServiceRepository) ResourceList() ([]Resource, error)
func (h *EndorServiceRepository) UpdateOne(dto UpdateByIdDTO[Resource]) (*Resource, error)

### Message

```go
type Message struct{ ... }
```


type Message struct {
	Gravity MessageGravity `json:"gravity"`
	Value   string         `json:"value"`
}

func NewMessage(gravity MessageGravity, value string) Message

### MessageGravity

```go
type MessageGravity string
```


type MessageGravity string
    Message Gravity

const Info MessageGravity = "Info" ...

### Meta

```go
type Meta struct{ ... }
```


type Meta struct {
	Default  Presentation            `json:"default"`
	Elements map[string]Presentation `json:"elements"`
}


### MongoCollectionAdapter

```go
type MongoCollectionAdapter struct{ ... }
```


type MongoCollectionAdapter struct {
	// Has unexported fields.
}
    MongoCollectionAdapter adapts MongoDB collection to CollectionInterface.

func (m *MongoCollectionAdapter) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (interfaces.CursorInterface, error)
func (m *MongoCollectionAdapter) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
func (m *MongoCollectionAdapter) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
func (m *MongoCollectionAdapter) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (interfaces.CursorInterface, error)
func (m *MongoCollectionAdapter) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) interfaces.SingleResultInterface
func (m *MongoCollectionAdapter) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
func (m *MongoCollectionAdapter) UpdateOne(ctx context.Context, filter interface{}, update interface{}, ...) (*mongo.UpdateResult, error)

### MongoCursorAdapter

```go
type MongoCursorAdapter struct{ ... }
```


type MongoCursorAdapter struct {
	// Has unexported fields.
}
    MongoCursorAdapter adapts MongoDB cursor to CursorInterface.

func (m *MongoCursorAdapter) All(ctx context.Context, results interface{}) error
func (m *MongoCursorAdapter) Close(ctx context.Context) error
func (m *MongoCursorAdapter) Decode(val interface{}) error
func (m *MongoCursorAdapter) Err() error
func (m *MongoCursorAdapter) Next(ctx context.Context) bool

### MongoDatabaseAdapter

```go
type MongoDatabaseAdapter struct{ ... }
```


type MongoDatabaseAdapter struct {
	// Has unexported fields.
}
    MongoDatabaseAdapter adapts MongoDB database to DatabaseInterface.

func (m *MongoDatabaseAdapter) Collection(name string) interfaces.CollectionInterface
func (m *MongoDatabaseAdapter) Name() string
func (m *MongoDatabaseAdapter) RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) interfaces.SingleResultInterface

### MongoDatabaseClient

```go
type MongoDatabaseClient struct{ ... }
```


type MongoDatabaseClient struct {
	// Has unexported fields.
}
    MongoDatabaseClient implements DatabaseClientInterface using MongoDB driver.
    This provides the concrete implementation for dependency injection while
    maintaining the same performance and functionality as direct MongoDB client
    usage.

    Acceptance Criteria 3: All MongoDB operations use injected
    DatabaseClientInterface.

func (m *MongoDatabaseClient) Close(ctx context.Context) error
func (m *MongoDatabaseClient) Collection(name string) interfaces.CollectionInterface
func (m *MongoDatabaseClient) Database(name string) interfaces.DatabaseInterface
func (m *MongoDatabaseClient) Ping(ctx context.Context) error
func (m *MongoDatabaseClient) StartTransaction(ctx context.Context) (interfaces.TransactionInterface, error)

### MongoResourceInstanceRepository[T

```go
type MongoResourceInstanceRepository[T ResourceInstanceInterface] struct{ ... }
```


### MongoResourceInstanceSpecializedRepository[T

```go
type MongoResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct{ ... }
```


### MongoSingleResultAdapter

```go
type MongoSingleResultAdapter struct{ ... }
```


type MongoSingleResultAdapter struct {
	// Has unexported fields.
}
    MongoSingleResultAdapter adapts MongoDB SingleResult to
    SingleResultInterface.

func (m *MongoSingleResultAdapter) Decode(v interface{}) error
func (m *MongoSingleResultAdapter) Err() error

### MongoStaticResourceInstanceRepository[T

```go
type MongoStaticResourceInstanceRepository[T ResourceInstanceInterface] struct{ ... }
```


### MongoTransactionAdapter

```go
type MongoTransactionAdapter struct{ ... }
```


type MongoTransactionAdapter struct {
	// Has unexported fields.
}
    MongoTransactionAdapter adapts MongoDB session to TransactionInterface.

func (m *MongoTransactionAdapter) Abort(ctx context.Context) error
func (m *MongoTransactionAdapter) Commit(ctx context.Context) error
func (m *MongoTransactionAdapter) WithTransaction(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error)

### NoPayload

```go
type NoPayload struct{}
```


type NoPayload struct{}


### ObjectID

```go
type ObjectID primitive.ObjectID
```


type ObjectID primitive.ObjectID
    ObjectID è un wrapper per primitive.ObjectID che mantiene lo stesso
    comportamento con MongoDB ma serializza/deserializza in JSON come stringa
    esadecimale.

func NewObjectID() ObjectID
func (oid ObjectID) Hex() string
func (oid ObjectID) IsZero() bool
func (oid ObjectID) MarshalBSONValue() (bsontype.Type, []byte, error)
func (oid ObjectID) MarshalJSON() ([]byte, error)
func (oid ObjectID) ToPrimitive() primitive.ObjectID
func (oid *ObjectID) UnmarshalBSONValue(t bsontype.Type, data []byte) error
func (oid *ObjectID) UnmarshalJSON(data []byte) error

### OpenAPIConfiguration

```go
type OpenAPIConfiguration struct{ ... }
```


type OpenAPIConfiguration struct {
	OpenAPI    string                                 `json:"openapi"`
	Info       OpenAPIInfo                            `json:"info"`
	Servers    []OpenAPIServer                        `json:"servers"`
	Tags       []OpenAPITag                           `json:"tags"`
	Paths      map[string]map[string]OpenAPIOperation `json:"paths"`
	Components OpenApiComponents                      `json:"components"`
}

func CreateSwaggerDefinition(microServiceId string, microServiceAddress string, services []EndorService, ...) (OpenAPIConfiguration, error)

### OpenAPIInfo

```go
type OpenAPIInfo struct{ ... }
```


type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}


### OpenAPIMediaType

```go
type OpenAPIMediaType struct{ ... }
```


type OpenAPIMediaType struct {
	Schema Schema `json:"schema"`
}


### OpenAPIOperation

```go
type OpenAPIOperation struct{ ... }
```


type OpenAPIOperation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	OperationID string              `json:"operationId,omitempty"`
	Parameters  []OpenAPIParameter  `json:"parameters"`
	RequestBody *OpenAPIRequestBody `json:"requestBody,omitempty"`
	Responses   OpenApiResponses    `json:"responses"`
}


### OpenAPIParameter

```go
type OpenAPIParameter struct{ ... }
```


type OpenAPIParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Schema      Schema `json:"schema"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}


### OpenAPIRequestBody

```go
type OpenAPIRequestBody struct{ ... }
```


type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content"`
	Required    bool                        `json:"required,omitempty"`
}


### OpenAPIServer

```go
type OpenAPIServer struct{ ... }
```


type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}


### OpenAPITag

```go
type OpenAPITag struct{ ... }
```


type OpenAPITag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}


### OpenApiAuth

```go
type OpenApiAuth struct{ ... }
```


type OpenApiAuth struct {
	Type string `json:"type"`
	In   string `json:"in"`
	Name string `json:"name"`
}


### OpenApiComponents

```go
type OpenApiComponents struct{ ... }
```


type OpenApiComponents struct {
	Schemas map[string]Schema `json:"schemas"`
}


### OpenApiResponse

```go
type OpenApiResponse struct{ ... }
```


type OpenApiResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenAPIMediaType `json:"content"`
}


### OpenApiResponses

```go
type OpenApiResponses map[string]OpenApiResponse
```


type OpenApiResponses map[string]OpenApiResponse


### Presentation

```go
type Presentation struct{ ... }
```


type Presentation struct {
	Entity string `json:"entity"`
	Icon   string `json:"icon"`
}


### ReadDTO

```go
type ReadDTO struct{ ... }
```


type ReadDTO struct {
	Filter     map[string]interface{} `json:"filter"`
	Projection map[string]interface{} `json:"projection"`
}


### ReadInstanceDTO

```go
type ReadInstanceDTO struct{ ... }
```


type ReadInstanceDTO struct {
	Id string `json:"id,omitempty"`
}


### Resource

```go
type Resource struct{ ... }
```


type Resource struct {
	ID                   string     `json:"id" bson:"_id" schema:"title=Id"`
	Description          string     `json:"description" schema:"title=Description"`
	Service              string     `json:"service" schema:"title=Service" ui-schema:"resource=microservice"`
	AdditionalAttributes string     `json:"additionalAttributes" schema:"title=Additional attributes schema,format=yaml"` // YAML string, raw
	Categories           []Category `json:"categories,omitempty" bson:"categories,omitempty" schema:"title=Categories"`
}

func (h *Resource) GetCategoryByID(categoryID string) (*Category, bool)
func (h *Resource) GetCategorySchema(categoryID string) (*RootSchema, error)
func (h *Resource) UnmarshalAdditionalAttributes() (*RootSchema, error)

### ResourceAction

```go
type ResourceAction struct{ ... }
```


type ResourceAction struct {
	// version/resource/action
	ID          string `json:"id" schema:"title=Id"`
	Resource    string `json:"resource" schema:"title=Resource" ui-schema:"resource=resource"`
	Description string `json:"description" schema:"title=Description"`
	InputSchema string `json:"inputSchema" schema:"title=Input schema,format=yaml"`
}


### ResourceActionService

```go
type ResourceActionService struct{ ... }
```


type ResourceActionService struct {
	// Has unexported fields.
}


### ResourceInstance[T

```go
type ResourceInstance[T ResourceInstanceInterface] struct{ ... }
```


### ResourceInstanceInterface

```go
type ResourceInstanceInterface interface{ ... }
```


type ResourceInstanceInterface interface {
	GetID() *string
	SetID(id string)
}


### ResourceInstanceRepository[T

```go
type ResourceInstanceRepository[T ResourceInstanceInterface] struct{ ... }
```


### ResourceInstanceRepositoryInterface[T

```go
type ResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface{ ... }
```


### ResourceInstanceRepositoryOptions

```go
type ResourceInstanceRepositoryOptions struct{ ... }
```


type ResourceInstanceRepositoryOptions struct {
	// AutoGenerateID determines whether IDs should be auto-generated or provided by the user
	// When true (default): Empty IDs are auto-generated using primitive.ObjectID.Hex()
	// When false: IDs must be provided by the user, empty IDs cause BadRequestError
	AutoGenerateID *bool
}
    ResourceInstanceRepositoryOptions defines configuration options for
    ResourceInstanceRepository


### ResourceInstanceSpecialized[T

```go
type ResourceInstanceSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct{ ... }
```


### ResourceInstanceSpecializedInterface

```go
type ResourceInstanceSpecializedInterface interface{ ... }
```


type ResourceInstanceSpecializedInterface interface {
	GetCategoryType() *string
	SetCategoryType(categoryType string)
}


### ResourceInstanceSpecializedRepository[T

```go
type ResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct{ ... }
```


### ResourceInstanceSpecializedRepositoryInterface[T

```go
type ResourceInstanceSpecializedRepositoryInterface[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] interface{ ... }
```


### ResourceService

```go
type ResourceService struct{ ... }
```


type ResourceService struct {
	// Has unexported fields.
}


### Response[T

```go
type Response[T any] struct{ ... }
```


### ResponseBuilder[T

```go
type ResponseBuilder[T any] struct{ ... }
```


### RootSchema

```go
type RootSchema struct{ ... }
```


type RootSchema struct {
	Schema      `json:",inline" yaml:",inline"`
	Definitions map[string]Schema `json:"$defs,omitempty" yaml:"$defs,omitempty"`
}

func NewSchema(model any) *RootSchema
func NewSchemaByType(t reflect.Type) *RootSchema
func (h *RootSchema) ToYAML() (string, error)

### Schema

```go
type Schema struct{ ... }
```


type Schema struct {
	Reference   string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type        SchemaTypeName     `json:"type,omitempty" yaml:"type,omitempty"`
	Properties  *map[string]Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	Enum        *[]string          `json:"enum,omitempty" yaml:"enum,omitempty"`
	Title       *string            `json:"title,omitempty" yaml:"title,omitempty"`
	Description *string            `json:"description,omitempty" yaml:"description,omitempty"`
	Format      *SchemaFormatName  `json:"format,omitempty" yaml:"format,omitempty"`
	ReadOnly    *bool              `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly   *bool              `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`

	// field dimension
	MinLength *int `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength *int `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`

	UISchema *UISchema `json:"x-ui,omitempty" yaml:"x-ui,omitempty"`
}


### SchemaFormatName

```go
type SchemaFormatName string
```


type SchemaFormatName string

const DateTimeFormat SchemaFormatName = "date-time" ...
func NewSchemaFormat(f SchemaFormatName) *SchemaFormatName

### SchemaTypeName

```go
type SchemaTypeName string
```


type SchemaTypeName string

const StringType SchemaTypeName = "string" ...

### ServerConfig

```go
type ServerConfig struct{ ... }
```


type ServerConfig struct {
	ServerPort                    string
	DocumentDBUri                 string
	HybridResourcesEnabled        bool
	DynamicResourcesEnabled       bool
	DynamicResourceDocumentDBName string
}

func GetConfig() *ServerConfig
func (c *ServerConfig) GetDocumentDBUri() string
func (c *ServerConfig) GetDynamicResourceDocumentDBName() string
func (c *ServerConfig) GetServerPort() string
func (c *ServerConfig) IsDynamicResourcesEnabled() bool
func (c *ServerConfig) IsHybridResourcesEnabled() bool
func (c *ServerConfig) Reload() error
func (c *ServerConfig) Validate() error

### Session

```go
type Session struct{ ... }
```


type Session struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	Development bool   `json:"development"`
}


### StaticResourceInstanceRepository[T

```go
type StaticResourceInstanceRepository[T ResourceInstanceInterface] struct{ ... }
```


### StaticResourceInstanceRepositoryInterface[T

```go
type StaticResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface{ ... }
```


### StaticResourceInstanceRepositoryOptions

```go
type StaticResourceInstanceRepositoryOptions struct{ ... }
```


type StaticResourceInstanceRepositoryOptions struct {
	// AutoGenerateID determines whether IDs should be auto-generated or provided by the user
	// When true (default): Empty IDs are auto-generated using primitive.ObjectID.Hex()
	// When false: IDs must be provided by the user, empty IDs cause BadRequestError
	AutoGenerateID *bool
}
    StaticResourceInstanceRepositoryOptions defines configuration options for
    StaticResourceInstanceRepository Mirrors ResourceInstanceRepositoryOptions
    for consistency


### UISchema

```go
type UISchema struct{ ... }
```


type UISchema struct {
	Resource *string   `json:"resource,omitempty" yaml:"resource,omitempty"` // define the reference resource
	Order    *[]string `json:"order,omitempty" yaml:"order,omitempty"`       // define the order of the attributes
	Hidden   *bool     `json:"hidden,omitempty" yaml:"hidden,omitempty"`     // define if the property is displayable
}


### UpdateByIdDTO[T

```go
type UpdateByIdDTO[T any] struct{ ... }
```


---

*Generated by [endor-sdk-go documentation generator](https://github.com/mattiabonardi/endor-sdk-go/tree/main/tools/gendocs)*
