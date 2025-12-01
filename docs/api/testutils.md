# Test Utilities

> Package documentation for Test Utilities

**Import Path:** `github.com/mattiabonardi/endor-sdk-go/sdk/testutils`
**Generated:** 2025-12-01 10:07:53 UTC

---

func AssertConfigProvider(t *testing.T, config interfaces.ConfigProviderInterface)
func AssertHybridServiceInterface(t *testing.T, service interfaces.EndorHybridServiceInterface)
func AssertServiceInterface(t *testing.T, service interfaces.EndorServiceInterface)
func BenchmarkServiceOperation(b *testing.B, operation func())
func BenchmarkServiceWithLoad(b *testing.B, concurrency int, operation func())
func CleanupTestEnvironment(env *TestEnvironment)
func CreateTestGinContext() (*gin.Context, *gin.Engine)
func CreateTestGinContextWithAuth(sessionID, userID string, development bool) *gin.Context
func GetTestDeveloperSession() map[string]interface{}
func GetTestOrderSchema() interfaces.RootSchema
func GetTestProductSchema() interfaces.RootSchema
func GetTestSession() map[string]interface{}
func GetTestUserSchema() interfaces.RootSchema
func GetTestUserSession() map[string]interface{}
func NewTestAction(description string, public bool) interfaces.EndorServiceAction
func NewTestCategory(categoryID string) interfaces.EndorHybridServiceCategory
func SetupTestHybridService(resource string, description string) interfaces.EndorHybridServiceInterface
func SetupTestService(resource string, description string) interfaces.EndorServiceInterface
func SetupTestServiceWithActions(resource string, actions map[string]interfaces.EndorServiceAction) interfaces.EndorServiceInterface
func SimulateNetworkLatency(min, max time.Duration)
func SimulateServiceError(errorType string) error
func TestHybridServiceComposition(t *testing.T, services []interfaces.EndorHybridServiceInterface)
func TestServiceComposition(t *testing.T, services []interfaces.EndorServiceInterface)
func WithPerformanceAssertions(t *testing.T, maxDuration time.Duration, testFunc func()) time.Duration
func WithTimeout(t *testing.T, timeout time.Duration, testFunc func())
type GenericError struct{ ... }
type InMemoryRepository[T any] struct{ ... }
    func NewInMemoryRepository[T any]() *InMemoryRepository[T]
type IntegrationTestDatabase struct{ ... }
    func NewIntegrationTestDatabase(databaseName string) *IntegrationTestDatabase
type IntegrationTestSuite struct{ ... }
    func NewIntegrationTestSuite(suiteName string) *IntegrationTestSuite
type InternalError struct{ ... }
type MockConfigProvider struct{ ... }
type MockEndorContext[T any] struct{ ... }
type MockEndorHybridService struct{ ... }
type MockEndorHybridServiceCategory struct{ ... }
type MockEndorService struct{ ... }
type MockEndorServiceAction struct{ ... }
type MockRepositoryOptions struct{ ... }
    func DefaultMockRepositoryOptions() MockRepositoryOptions
type NoPayload struct{}
type NotFoundError struct{ ... }
type PerformanceMockService struct{ ... }
    func NewPerformanceMockService(latency time.Duration) *PerformanceMockService
type TestConfigProviderBuilder struct{ ... }
    func NewTestConfigProvider() *TestConfigProviderBuilder
type TestEndorContextBuilder[T any] struct{ ... }
    func NewTestEndorContext[T any]() *TestEndorContextBuilder[T]
type TestEndorHybridServiceBuilder struct{ ... }
    func NewTestEndorHybridService() *TestEndorHybridServiceBuilder
type TestEndorServiceBuilder struct{ ... }
    func NewTestEndorService() *TestEndorServiceBuilder
type TestEnvironment struct{ ... }
    func SetupTestEnvironment(microServiceID string) *TestEnvironment
    func SetupTestEnvironmentWithConfig(microServiceID string, config interfaces.ConfigProviderInterface) *TestEnvironment
type TestLifecycleManager struct{ ... }
    func NewTestLifecycleManager() *TestLifecycleManager
type TestOrderItem struct{ ... }
type TestOrderPayload struct{ ... }
    func CloneTestOrder(order TestOrderPayload) TestOrderPayload
    func GetTestOrders() []TestOrderPayload
type TestProductPayload struct{ ... }
    func CloneTestProduct(product TestProductPayload) TestProductPayload
    func GetTestProducts() []TestProductPayload
type TestScenario struct{ ... }
    func GetAuthorizationTestScenarios() []TestScenario
    func GetCRUDTestScenarios() []TestScenario
type TestUserPayload struct{ ... }
    func CloneTestUser(user TestUserPayload) TestUserPayload
    func GetTestUsers() []TestUserPayload
type UnauthorizedError struct{ ... }
type ValidationError struct{ ... }

## Package Overview

package testutils // import "github.com/mattiabonardi/endor-sdk-go/sdk/testutils"

Package testutils provides comprehensive testing utilities and mock
implementations for the endor-sdk-go framework. This package includes
ready-to-use mock implementations for all framework interfaces, fluent API
builders, test data fixtures, helper functions, and complete integration testing
infrastructure, enabling developers to quickly write both unit and integration
tests without external dependencies.

The package follows the framework's established patterns for interface-based
dependency injection and testing, supporting both behavior verification (using
testify/mock) and state testing.

Key Components:

- Mock implementations: MockEndorService, MockEndorHybridService,
MockConfigProvider, etc. - Fluent API builders: TestEndorServiceBuilder,
TestConfigProviderBuilder with method chaining - Test data fixtures: Realistic
test data sets and generators for different service types - Helper functions:
Environment setup, assertion helpers, error simulation utilities - Integration

## Exported Types

### AssertConfigProvider

```go
func AssertConfigProvider(t *testing.T, config interfaces.ConfigProviderInterface)
```

### AssertHybridServiceInterface

```go
func AssertHybridServiceInterface(t *testing.T, service interfaces.EndorHybridServiceInterface)
```

### AssertServiceInterface

```go
func AssertServiceInterface(t *testing.T, service interfaces.EndorServiceInterface)
```

### BenchmarkServiceOperation

```go
func BenchmarkServiceOperation(b *testing.B, operation func())
```

### BenchmarkServiceWithLoad

```go
func BenchmarkServiceWithLoad(b *testing.B, concurrency int, operation func())
```

### CleanupTestEnvironment

```go
func CleanupTestEnvironment(env *TestEnvironment)
```

### CreateTestGinContext

```go
func CreateTestGinContext() (*gin.Context, *gin.Engine)
```

### CreateTestGinContextWithAuth

```go
func CreateTestGinContextWithAuth(sessionID, userID string, development bool) *gin.Context
```

### GetTestDeveloperSession

```go
func GetTestDeveloperSession() map[string]interface{}
```

### GetTestOrderSchema

```go
func GetTestOrderSchema() interfaces.RootSchema
```

### GetTestProductSchema

```go
func GetTestProductSchema() interfaces.RootSchema
```

### GetTestSession

```go
func GetTestSession() map[string]interface{}
```

### GetTestUserSchema

```go
func GetTestUserSchema() interfaces.RootSchema
```

### GetTestUserSession

```go
func GetTestUserSession() map[string]interface{}
```

### NewTestAction

```go
func NewTestAction(description string, public bool) interfaces.EndorServiceAction
```

### NewTestCategory

```go
func NewTestCategory(categoryID string) interfaces.EndorHybridServiceCategory
```

### SetupTestHybridService

```go
func SetupTestHybridService(resource string, description string) interfaces.EndorHybridServiceInterface
```

### SetupTestService

```go
func SetupTestService(resource string, description string) interfaces.EndorServiceInterface
```

### SetupTestServiceWithActions

```go
func SetupTestServiceWithActions(resource string, actions map[string]interfaces.EndorServiceAction) interfaces.EndorServiceInterface
```

### SimulateNetworkLatency

```go
func SimulateNetworkLatency(min, max time.Duration)
```

### SimulateServiceError

```go
func SimulateServiceError(errorType string) error
```

### TestHybridServiceComposition

```go
func TestHybridServiceComposition(t *testing.T, services []interfaces.EndorHybridServiceInterface)
```

### TestServiceComposition

```go
func TestServiceComposition(t *testing.T, services []interfaces.EndorServiceInterface)
```

### WithPerformanceAssertions

```go
func WithPerformanceAssertions(t *testing.T, maxDuration time.Duration, testFunc func()) time.Duration
```

### WithTimeout

```go
func WithTimeout(t *testing.T, timeout time.Duration, testFunc func())
```

### GenericError

```go
type GenericError struct{ ... }
```


type GenericError struct {
	// Has unexported fields.
}
    GenericError simulates generic errors.

func (e *GenericError) Error() string

### InMemoryRepository[T

```go
type InMemoryRepository[T any] struct{ ... }
```


### IntegrationTestDatabase

```go
type IntegrationTestDatabase struct{ ... }
```


type IntegrationTestDatabase struct {
	ConnectionString string
	DatabaseName     string
	CollectionPrefix string
	CleanupAfter     bool

	// Has unexported fields.
}
    IntegrationTestDatabase provides utilities for MongoDB integration testing

func NewIntegrationTestDatabase(databaseName string) *IntegrationTestDatabase
func (db *IntegrationTestDatabase) Cleanup() error
func (db *IntegrationTestDatabase) CreateTestCollection(name string) string
func (db *IntegrationTestDatabase) GetCollectionName(name string) string

### IntegrationTestSuite

```go
type IntegrationTestSuite struct{ ... }
```


type IntegrationTestSuite struct {
	Manager       *TestLifecycleManager
	Database      *IntegrationTestDatabase
	Config        interfaces.ConfigProviderInterface
	Timeout       time.Duration
	SetupTasks    []func() error
	TeardownTasks []func() error
}
    IntegrationTestSuite provides a complete test suite for integration testing

func NewIntegrationTestSuite(suiteName string) *IntegrationTestSuite
func (suite *IntegrationTestSuite) AddSetupTask(task func() error)
func (suite *IntegrationTestSuite) AddTeardownTask(task func() error)
func (suite *IntegrationTestSuite) Setup() error
func (suite *IntegrationTestSuite) Teardown() error
func (suite *IntegrationTestSuite) WithCleanupDisabled() *IntegrationTestSuite
func (suite *IntegrationTestSuite) WithTimeout(timeout time.Duration) *IntegrationTestSuite

### InternalError

```go
type InternalError struct{ ... }
```


type InternalError struct {
	// Has unexported fields.
}
    InternalError simulates internal server errors.

func (e *InternalError) Error() string

### MockConfigProvider

```go
type MockConfigProvider struct{ ... }
```


type MockConfigProvider struct {
	mock.Mock
}
    MockConfigProvider provides a mock implementation of ConfigProviderInterface
    for testing configuration-dependent functionality.

    Example usage:

        mockConfig := &MockConfigProvider{}
        mockConfig.On("GetServerPort").Return("8080")
        mockConfig.On("GetDocumentDBUri").Return("mongodb://test:27017")
        mockConfig.On("IsHybridResourcesEnabled").Return(true)

func (m *MockConfigProvider) GetDocumentDBUri() string
func (m *MockConfigProvider) GetDynamicResourceDocumentDBName() string
func (m *MockConfigProvider) GetServerPort() string
func (m *MockConfigProvider) IsDynamicResourcesEnabled() bool
func (m *MockConfigProvider) IsHybridResourcesEnabled() bool
func (m *MockConfigProvider) Reload() error
func (m *MockConfigProvider) Validate() error

### MockEndorContext[T

```go
type MockEndorContext[T any] struct{ ... }
```


### MockEndorHybridService

```go
type MockEndorHybridService struct{ ... }
```


type MockEndorHybridService struct {
	mock.Mock
}
    MockEndorHybridService provides a mock implementation of
    EndorHybridServiceInterface for testing hybrid services with category-based
    specialization and dynamic actions.

    Example usage:

        mockHybridService := &MockEndorHybridService{}
        mockHybridService.On("GetResource").Return("products")
        mockHybridService.On("GetResourceDescription").Return("Product management")
        mockHybridService.On("WithCategories", mock.Anything).Return(mockHybridService)
        mockHybridService.On("ToEndorService", mock.Anything).Return(mockEndorService)

func (m *MockEndorHybridService) EmbedService(prefix string, service interfaces.EndorServiceInterface) error
func (m *MockEndorHybridService) GetEmbeddedServices() map[string]interfaces.EndorServiceInterface
func (m *MockEndorHybridService) GetPriority() *int
func (m *MockEndorHybridService) GetResource() string
func (m *MockEndorHybridService) GetResourceDescription() string
func (m *MockEndorHybridService) ToEndorService(metadataSchema interfaces.Schema) interfaces.EndorServiceInterface
func (m *MockEndorHybridService) Validate() error
func (m *MockEndorHybridService) WithActions(...) interfaces.EndorHybridServiceInterface
func (m *MockEndorHybridService) WithCategories(categories []interfaces.EndorHybridServiceCategory) interfaces.EndorHybridServiceInterface

### MockEndorHybridServiceCategory

```go
type MockEndorHybridServiceCategory struct{ ... }
```


type MockEndorHybridServiceCategory struct {
	mock.Mock
}
    MockEndorHybridServiceCategory provides a mock implementation of
    EndorHybridServiceCategory for testing category-based specializations in
    hybrid services.

    Example usage:

        mockCategory := &MockEndorHybridServiceCategory{}
        mockCategory.On("GetID").Return("admin")
        mockCategory.On("CreateDefaultActions", "users", "User management", mock.Anything).Return(testActions)

func (m *MockEndorHybridServiceCategory) CreateDefaultActions(resource string, resourceDescription string, metadataSchema interfaces.Schema) map[string]interfaces.EndorServiceAction
func (m *MockEndorHybridServiceCategory) GetID() string

### MockEndorService

```go
type MockEndorService struct{ ... }
```


type MockEndorService struct {
	mock.Mock
}
    MockEndorService provides a mock implementation of EndorServiceInterface for
    testing purposes. It uses testify/mock for behavior verification and call
    tracking.

    Example usage:

        mockService := &MockEndorService{}
        mockService.On("GetResource").Return("users")
        mockService.On("GetDescription").Return("User service")
        mockService.On("Validate").Return(nil)

        // Use in your test
        result := myFunction(mockService)

        // Assert expectations
        mockService.AssertExpectations(t)

func (m *MockEndorService) GetDescription() string
func (m *MockEndorService) GetMethods() map[string]interfaces.EndorServiceAction
func (m *MockEndorService) GetPriority() *int
func (m *MockEndorService) GetResource() string
func (m *MockEndorService) GetVersion() string
func (m *MockEndorService) Validate() error

### MockEndorServiceAction

```go
type MockEndorServiceAction struct{ ... }
```


type MockEndorServiceAction struct {
	mock.Mock
}
    MockEndorServiceAction provides a mock implementation of EndorServiceAction
    for testing service action behavior and HTTP callback creation.

    Example usage:

        mockAction := &MockEndorServiceAction{}
        mockAction.On("GetOptions").Return(interfaces.EndorServiceActionOptions{
        	Description: "Test action",
        	Public: true,
        })
        mockAction.On("CreateHTTPCallback", "test-service").Return(testHandler)

func (m *MockEndorServiceAction) CreateHTTPCallback(microserviceId string) func(c *gin.Context)
func (m *MockEndorServiceAction) GetOptions() interfaces.EndorServiceActionOptions

### MockRepositoryOptions

```go
type MockRepositoryOptions struct{ ... }
```


type MockRepositoryOptions struct {
	AutoGenerateID  bool
	SimulateLatency time.Duration
	ErrorRate       float64 // 0.0 to 1.0
	MaxResults      int
}
    MockRepositoryOptions provides configuration for repository mock behavior.

func DefaultMockRepositoryOptions() MockRepositoryOptions

### NoPayload

```go
type NoPayload struct{}
```


type NoPayload struct{}
    NoPayload represents actions that don't require input payload.


### NotFoundError

```go
type NotFoundError struct{ ... }
```


type NotFoundError struct {
	// Has unexported fields.
}
    NotFoundError simulates resource not found errors.

func (e *NotFoundError) Error() string

### PerformanceMockService

```go
type PerformanceMockService struct{ ... }
```


type PerformanceMockService struct {
	*MockEndorService
	// Has unexported fields.
}
    PerformanceMockService wraps a MockEndorService with latency simulation for
    integration testing scenarios.

func NewPerformanceMockService(latency time.Duration) *PerformanceMockService
func (p *PerformanceMockService) GetDescription() string
func (p *PerformanceMockService) GetResource() string
func (p *PerformanceMockService) Validate() error

### TestConfigProviderBuilder

```go
type TestConfigProviderBuilder struct{ ... }
```


type TestConfigProviderBuilder struct {
	// Has unexported fields.
}
    for testing different configuration scenarios.

    Example usage:

        testConfig := NewTestConfigProvider().
        	WithServerPort("9999").
        	WithDocumentDBUri("mongodb://test:27017").
        	WithHybridResourcesEnabled(true).
        	Build()

func NewTestConfigProvider() *TestConfigProviderBuilder
func (b *TestConfigProviderBuilder) Build() *MockConfigProvider
func (b *TestConfigProviderBuilder) WithDocumentDBUri(uri string) *TestConfigProviderBuilder
func (b *TestConfigProviderBuilder) WithDynamicResourceDocumentDBName(dbName string) *TestConfigProviderBuilder
func (b *TestConfigProviderBuilder) WithDynamicResourcesEnabled(enabled bool) *TestConfigProviderBuilder
func (b *TestConfigProviderBuilder) WithHybridResourcesEnabled(enabled bool) *TestConfigProviderBuilder
func (b *TestConfigProviderBuilder) WithServerPort(port string) *TestConfigProviderBuilder

### TestEndorContextBuilder[T

```go
type TestEndorContextBuilder[T any] struct{ ... }
```


### TestEndorHybridServiceBuilder

```go
type TestEndorHybridServiceBuilder struct{ ... }
```


type TestEndorHybridServiceBuilder struct {
	// Has unexported fields.
}
    TestEndorHybridServiceBuilder provides a fluent API for creating
    EndorHybridService instances for testing purposes with support for
    categories and custom actions.

    Example usage:

        testHybridService := NewTestEndorHybridService().
        	WithResource("users").
        	WithResourceDescription("User management").
        	WithPriority(5).
        	WithCategory("admin", testAdminCategory).
        	Build()

func NewTestEndorHybridService() *TestEndorHybridServiceBuilder
func (b *TestEndorHybridServiceBuilder) Build() *MockEndorHybridService
func (b *TestEndorHybridServiceBuilder) WithActions(hasActions bool) *TestEndorHybridServiceBuilder
func (b *TestEndorHybridServiceBuilder) WithCategory(categoryID string, category interfaces.EndorHybridServiceCategory) *TestEndorHybridServiceBuilder
func (b *TestEndorHybridServiceBuilder) WithDefaultCategory(categoryID string) *TestEndorHybridServiceBuilder
func (b *TestEndorHybridServiceBuilder) WithPriority(priority int) *TestEndorHybridServiceBuilder
func (b *TestEndorHybridServiceBuilder) WithResource(resource string) *TestEndorHybridServiceBuilder
func (b *TestEndorHybridServiceBuilder) WithResourceDescription(description string) *TestEndorHybridServiceBuilder

### TestEndorServiceBuilder

```go
type TestEndorServiceBuilder struct{ ... }
```


type TestEndorServiceBuilder struct {
	// Has unexported fields.
}
    TestEndorServiceBuilder provides a fluent API for creating EndorService
    instances for testing purposes. It supports method chaining to configure all
    service properties.

    Example usage:

        testService := NewTestEndorService().
        	WithResource("products").
        	WithDescription("Product management service").
        	WithVersion("1.0").
        	WithPriority(10).
        	WithMethod("create", testCreateAction).
        	Build()

func NewTestEndorService() *TestEndorServiceBuilder
func (b *TestEndorServiceBuilder) Build() *MockEndorService
func (b *TestEndorServiceBuilder) WithBasicMethods() *TestEndorServiceBuilder
func (b *TestEndorServiceBuilder) WithDescription(description string) *TestEndorServiceBuilder
func (b *TestEndorServiceBuilder) WithMethod(name string, action interfaces.EndorServiceAction) *TestEndorServiceBuilder
func (b *TestEndorServiceBuilder) WithPriority(priority int) *TestEndorServiceBuilder
func (b *TestEndorServiceBuilder) WithResource(resource string) *TestEndorServiceBuilder
func (b *TestEndorServiceBuilder) WithVersion(version string) *TestEndorServiceBuilder

### TestEnvironment

```go
type TestEnvironment struct{ ... }
```


type TestEnvironment struct {
	MicroServiceID string
	ConfigProvider interfaces.ConfigProviderInterface
	Services       map[string]interfaces.EndorServiceInterface
	HybridServices map[string]interfaces.EndorHybridServiceInterface
	Context        context.Context
	CancelFunc     context.CancelFunc
	CleanupFuncs   []func()
}
    TestEnvironment represents a complete test environment setup with all
    necessary components for testing framework functionality.

func SetupTestEnvironment(microServiceID string) *TestEnvironment
func SetupTestEnvironmentWithConfig(microServiceID string, config interfaces.ConfigProviderInterface) *TestEnvironment
func (env *TestEnvironment) AddCleanupFunc(cleanup func())
func (env *TestEnvironment) AddHybridService(name string, service interfaces.EndorHybridServiceInterface)
func (env *TestEnvironment) AddService(name string, service interfaces.EndorServiceInterface)

### TestLifecycleManager

```go
type TestLifecycleManager struct{ ... }
```


type TestLifecycleManager struct {
	// Has unexported fields.
}
    TestLifecycleManager handles the lifecycle of test services and resources

func NewTestLifecycleManager() *TestLifecycleManager
func (lm *TestLifecycleManager) GetHybridService(name string) (interfaces.EndorHybridServiceInterface, bool)
func (lm *TestLifecycleManager) GetRepository(name string) (interface{}, bool)
func (lm *TestLifecycleManager) GetService(name string) (interfaces.EndorServiceInterface, bool)
func (lm *TestLifecycleManager) RegisterHybridService(name string, service interfaces.EndorHybridServiceInterface)
func (lm *TestLifecycleManager) RegisterRepository(name string, repo interface{})
func (lm *TestLifecycleManager) RegisterService(name string, service interfaces.EndorServiceInterface)
func (lm *TestLifecycleManager) SetDatabase(db *IntegrationTestDatabase)
func (lm *TestLifecycleManager) StartAll() error
func (lm *TestLifecycleManager) StopAll() error

### TestOrderItem

```go
type TestOrderItem struct{ ... }
```


type TestOrderItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
    TestOrderItem represents an order item for testing order scenarios.


### TestOrderPayload

```go
type TestOrderPayload struct{ ... }
```


type TestOrderPayload struct {
	ID         string                 `json:"id,omitempty"`
	CustomerID string                 `json:"customerId"`
	Items      []TestOrderItem        `json:"items"`
	Total      float64                `json:"total"`
	Status     string                 `json:"status"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}
    TestOrderPayload represents a typical order data payload for testing.

func CloneTestOrder(order TestOrderPayload) TestOrderPayload
func GetTestOrders() []TestOrderPayload

### TestProductPayload

```go
type TestProductPayload struct{ ... }
```


type TestProductPayload struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Category    string   `json:"category"`
	InStock     bool     `json:"inStock"`
	Tags        []string `json:"tags,omitempty"`
}
    TestProductPayload represents a typical product data payload for testing.

func CloneTestProduct(product TestProductPayload) TestProductPayload
func GetTestProducts() []TestProductPayload

### TestScenario

```go
type TestScenario struct{ ... }
```


type TestScenario struct {
	Name        string
	Description string
	Context     interface{}
	Payload     interface{}
	Expected    interface{}
	ShouldError bool
	ErrorType   string
}
    TestScenario represents a complete test scenario with context and expected
    outcomes.

func GetAuthorizationTestScenarios() []TestScenario
func GetCRUDTestScenarios() []TestScenario

### TestUserPayload

```go
type TestUserPayload struct{ ... }
```


type TestUserPayload struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Role     string                 `json:"role,omitempty"`
	Active   bool                   `json:"active"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
    TestUserPayload represents a typical user data payload for testing.

func CloneTestUser(user TestUserPayload) TestUserPayload
func GetTestUsers() []TestUserPayload

### UnauthorizedError

```go
type UnauthorizedError struct{ ... }
```


type UnauthorizedError struct {
	// Has unexported fields.
}
    UnauthorizedError simulates authorization errors.

func (e *UnauthorizedError) Error() string

### ValidationError

```go
type ValidationError struct{ ... }
```


type ValidationError struct {
	// Has unexported fields.
}
    ValidationError simulates validation errors.

func (e *ValidationError) Error() string

---

*Generated by [endor-sdk-go documentation generator](https://github.com/mattiabonardi/endor-sdk-go/tree/main/tools/gendocs)*
