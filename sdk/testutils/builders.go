package testutils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/stretchr/testify/mock"
)

// TestEndorServiceBuilder provides a fluent API for creating EndorService instances
// for testing purposes. It supports method chaining to configure all service properties.
//
// Example usage:
//
//	testService := NewTestEndorService().
//		WithResource("products").
//		WithDescription("Product management service").
//		WithVersion("1.0").
//		WithPriority(10).
//		WithMethod("create", testCreateAction).
//		Build()
type TestEndorServiceBuilder struct {
	resource    string
	description string
	methods     map[string]interfaces.EndorServiceAction
	priority    *int
	version     string
}

// NewTestEndorService creates a new TestEndorServiceBuilder with sensible defaults.
func NewTestEndorService() *TestEndorServiceBuilder {
	return &TestEndorServiceBuilder{
		resource:    "test-resource",
		description: "Test service for unit testing",
		methods:     make(map[string]interfaces.EndorServiceAction),
		version:     "1.0",
	}
}

// WithResource sets the resource name for the service.
func (b *TestEndorServiceBuilder) WithResource(resource string) *TestEndorServiceBuilder {
	b.resource = resource
	return b
}

// WithDescription sets the description for the service.
func (b *TestEndorServiceBuilder) WithDescription(description string) *TestEndorServiceBuilder {
	b.description = description
	return b
}

// WithVersion sets the version for the service.
func (b *TestEndorServiceBuilder) WithVersion(version string) *TestEndorServiceBuilder {
	b.version = version
	return b
}

// WithPriority sets the priority for the service.
func (b *TestEndorServiceBuilder) WithPriority(priority int) *TestEndorServiceBuilder {
	b.priority = &priority
	return b
}

// WithMethod adds a method/action to the service.
func (b *TestEndorServiceBuilder) WithMethod(name string, action interfaces.EndorServiceAction) *TestEndorServiceBuilder {
	b.methods[name] = action
	return b
}

// WithBasicMethods adds common CRUD methods with mock implementations.
func (b *TestEndorServiceBuilder) WithBasicMethods() *TestEndorServiceBuilder {
	b.methods["create"] = NewTestAction("Create operation", true)
	b.methods["read"] = NewTestAction("Read operation", true)
	b.methods["update"] = NewTestAction("Update operation", true)
	b.methods["delete"] = NewTestAction("Delete operation", true)
	b.methods["list"] = NewTestAction("List operation", true)
	return b
}

// Build creates a MockEndorService configured with the builder settings.
func (b *TestEndorServiceBuilder) Build() *MockEndorService {
	mockService := &MockEndorService{}
	mockService.On("GetResource").Return(b.resource)
	mockService.On("GetDescription").Return(b.description)
	mockService.On("GetMethods").Return(b.methods)
	mockService.On("GetPriority").Return(b.priority)
	mockService.On("GetVersion").Return(b.version)
	mockService.On("Validate").Return(nil)
	return mockService
}

// TestEndorHybridServiceBuilder provides a fluent API for creating EndorHybridService instances
// for testing purposes with support for categories and custom actions.
//
// Example usage:
//
//	testHybridService := NewTestEndorHybridService().
//		WithResource("users").
//		WithResourceDescription("User management").
//		WithPriority(5).
//		WithCategory("admin", testAdminCategory).
//		Build()
type TestEndorHybridServiceBuilder struct {
	resource            string
	resourceDescription string
	priority            *int
	categories          []interfaces.EndorHybridServiceCategory
	hasActions          bool
}

// NewTestEndorHybridService creates a new TestEndorHybridServiceBuilder with sensible defaults.
func NewTestEndorHybridService() *TestEndorHybridServiceBuilder {
	return &TestEndorHybridServiceBuilder{
		resource:            "test-hybrid-resource",
		resourceDescription: "Test hybrid service for unit testing",
		categories:          make([]interfaces.EndorHybridServiceCategory, 0),
		hasActions:          false,
	}
}

// WithResource sets the resource name for the hybrid service.
func (b *TestEndorHybridServiceBuilder) WithResource(resource string) *TestEndorHybridServiceBuilder {
	b.resource = resource
	return b
}

// WithResourceDescription sets the resource description for the hybrid service.
func (b *TestEndorHybridServiceBuilder) WithResourceDescription(description string) *TestEndorHybridServiceBuilder {
	b.resourceDescription = description
	return b
}

// WithPriority sets the priority for the hybrid service.
func (b *TestEndorHybridServiceBuilder) WithPriority(priority int) *TestEndorHybridServiceBuilder {
	b.priority = &priority
	return b
}

// WithCategory adds a category to the hybrid service.
func (b *TestEndorHybridServiceBuilder) WithCategory(categoryID string, category interfaces.EndorHybridServiceCategory) *TestEndorHybridServiceBuilder {
	b.categories = append(b.categories, category)
	return b
}

// WithDefaultCategory adds a default test category with basic actions.
func (b *TestEndorHybridServiceBuilder) WithDefaultCategory(categoryID string) *TestEndorHybridServiceBuilder {
	category := NewTestCategory(categoryID)
	b.categories = append(b.categories, category)
	return b
}

// WithActions indicates that the hybrid service has custom actions.
func (b *TestEndorHybridServiceBuilder) WithActions(hasActions bool) *TestEndorHybridServiceBuilder {
	b.hasActions = hasActions
	return b
}

// Build creates a MockEndorHybridService configured with the builder settings.
func (b *TestEndorHybridServiceBuilder) Build() *MockEndorHybridService {
	mockHybridService := &MockEndorHybridService{}
	mockHybridService.On("GetResource").Return(b.resource)
	mockHybridService.On("GetResourceDescription").Return(b.resourceDescription)
	mockHybridService.On("GetPriority").Return(b.priority)
	mockHybridService.On("WithCategories", mock.Anything).Return(mockHybridService)
	mockHybridService.On("Validate").Return(nil)

	// Mock WithActions behavior
	if b.hasActions {
		mockHybridService.On("WithActions", mock.Anything).Return(mockHybridService)
	}

	// Mock ToEndorService conversion
	convertedService := NewTestEndorService().
		WithResource(b.resource).
		WithDescription(b.resourceDescription).
		WithBasicMethods().
		Build()

	// Only set priority if it's not nil
	if b.priority != nil {
		convertedService = NewTestEndorService().
			WithResource(b.resource).
			WithDescription(b.resourceDescription).
			WithPriority(*b.priority).
			WithBasicMethods().
			Build()
	}

	mockHybridService.On("ToEndorService", mock.Anything).Return(convertedService)

	return mockHybridService
} // TestConfigProviderBuilder provides a fluent API for creating ConfigProvider instances
// for testing different configuration scenarios.
//
// Example usage:
//
//	testConfig := NewTestConfigProvider().
//		WithServerPort("9999").
//		WithDocumentDBUri("mongodb://test:27017").
//		WithHybridResourcesEnabled(true).
//		Build()
type TestConfigProviderBuilder struct {
	serverPort                    string
	documentDBUri                 string
	hybridResourcesEnabled        bool
	dynamicResourcesEnabled       bool
	dynamicResourceDocumentDBName string
}

// NewTestConfigProvider creates a new TestConfigProviderBuilder with sensible defaults.
func NewTestConfigProvider() *TestConfigProviderBuilder {
	return &TestConfigProviderBuilder{
		serverPort:                    "8080",
		documentDBUri:                 "mongodb://localhost:27017",
		hybridResourcesEnabled:        true,
		dynamicResourcesEnabled:       true,
		dynamicResourceDocumentDBName: "test-db",
	}
}

// WithServerPort sets the server port configuration.
func (b *TestConfigProviderBuilder) WithServerPort(port string) *TestConfigProviderBuilder {
	b.serverPort = port
	return b
}

// WithDocumentDBUri sets the MongoDB connection URI.
func (b *TestConfigProviderBuilder) WithDocumentDBUri(uri string) *TestConfigProviderBuilder {
	b.documentDBUri = uri
	return b
}

// WithHybridResourcesEnabled sets the hybrid resources feature flag.
func (b *TestConfigProviderBuilder) WithHybridResourcesEnabled(enabled bool) *TestConfigProviderBuilder {
	b.hybridResourcesEnabled = enabled
	return b
}

// WithDynamicResourcesEnabled sets the dynamic resources feature flag.
func (b *TestConfigProviderBuilder) WithDynamicResourcesEnabled(enabled bool) *TestConfigProviderBuilder {
	b.dynamicResourcesEnabled = enabled
	return b
}

// WithDynamicResourceDocumentDBName sets the dynamic resources database name.
func (b *TestConfigProviderBuilder) WithDynamicResourceDocumentDBName(dbName string) *TestConfigProviderBuilder {
	b.dynamicResourceDocumentDBName = dbName
	return b
}

// Build creates a MockConfigProvider configured with the builder settings.
func (b *TestConfigProviderBuilder) Build() *MockConfigProvider {
	mockConfig := &MockConfigProvider{}
	mockConfig.On("GetServerPort").Return(b.serverPort)
	mockConfig.On("GetDocumentDBUri").Return(b.documentDBUri)
	mockConfig.On("IsHybridResourcesEnabled").Return(b.hybridResourcesEnabled)
	mockConfig.On("IsDynamicResourcesEnabled").Return(b.dynamicResourcesEnabled)
	mockConfig.On("GetDynamicResourceDocumentDBName").Return(b.dynamicResourceDocumentDBName)
	mockConfig.On("Reload").Return(nil)
	mockConfig.On("Validate").Return(nil)
	return mockConfig
}

// TestEndorContextBuilder provides a fluent API for creating EndorContext instances
// for testing request handling and context propagation.
//
// Example usage:
//
//	type UserPayload struct { Name string }
//	testContext := NewTestEndorContext[UserPayload]().
//		WithMicroServiceId("user-service").
//		WithSession(testSession).
//		WithPayload(UserPayload{Name: "Test User"}).
//		Build()
type TestEndorContextBuilder[T any] struct {
	microServiceId         string
	session                interface{}
	payload                T
	resourceMetadataSchema interface{}
	categoryID             *string
	ginContext             *gin.Context
}

// NewTestEndorContext creates a new TestEndorContextBuilder with sensible defaults.
func NewTestEndorContext[T any]() *TestEndorContextBuilder[T] {
	return &TestEndorContextBuilder[T]{
		microServiceId: "test-service",
		session: map[string]interface{}{
			"id":          "test-session-id",
			"username":    "test-user",
			"development": false,
		},
	}
}

// WithMicroServiceId sets the microservice identifier.
func (b *TestEndorContextBuilder[T]) WithMicroServiceId(id string) *TestEndorContextBuilder[T] {
	b.microServiceId = id
	return b
}

// WithSession sets the session information.
func (b *TestEndorContextBuilder[T]) WithSession(session interface{}) *TestEndorContextBuilder[T] {
	b.session = session
	return b
}

// WithPayload sets the request payload.
func (b *TestEndorContextBuilder[T]) WithPayload(payload T) *TestEndorContextBuilder[T] {
	b.payload = payload
	return b
}

// WithResourceMetadataSchema sets the resource metadata schema.
func (b *TestEndorContextBuilder[T]) WithResourceMetadataSchema(schema interface{}) *TestEndorContextBuilder[T] {
	b.resourceMetadataSchema = schema
	return b
}

// WithCategoryID sets the category identifier.
func (b *TestEndorContextBuilder[T]) WithCategoryID(categoryID string) *TestEndorContextBuilder[T] {
	b.categoryID = &categoryID
	return b
}

// WithGinContext sets the Gin HTTP context.
func (b *TestEndorContextBuilder[T]) WithGinContext(ginCtx *gin.Context) *TestEndorContextBuilder[T] {
	b.ginContext = ginCtx
	return b
}

// Build creates a MockEndorContext configured with the builder settings.
func (b *TestEndorContextBuilder[T]) Build() *MockEndorContext[T] {
	mockContext := &MockEndorContext[T]{}
	mockContext.On("GetMicroServiceId").Return(b.microServiceId)
	mockContext.On("GetSession").Return(b.session)
	mockContext.On("GetPayload").Return(b.payload)
	mockContext.On("SetPayload", b.payload).Return()
	mockContext.On("GetResourceMetadataSchema").Return(b.resourceMetadataSchema)
	mockContext.On("GetCategoryID").Return(b.categoryID)
	mockContext.On("SetCategoryID", b.categoryID).Return()
	mockContext.On("GetGinContext").Return(b.ginContext)
	return mockContext
}

// Helper functions for creating test data

// NewTestAction creates a test EndorServiceAction with basic configuration.
func NewTestAction(description string, public bool) interfaces.EndorServiceAction {
	mockAction := &MockEndorServiceAction{}
	options := interfaces.EndorServiceActionOptions{
		Description:     description,
		Public:          public,
		ValidatePayload: true,
	}
	mockAction.On("GetOptions").Return(options)
	mockAction.On("CreateHTTPCallback", "test-service").Return(func(c *gin.Context) {
		c.JSON(200, map[string]string{"result": "test"})
	})
	return mockAction
}

// NewTestCategory creates a test EndorHybridServiceCategory with basic configuration.
func NewTestCategory(categoryID string) interfaces.EndorHybridServiceCategory {
	mockCategory := &MockEndorHybridServiceCategory{}
	mockCategory.On("GetID").Return(categoryID)
	mockCategory.On("CreateDefaultActions", "test-resource", "Test resource", interfaces.Schema{}).Return(
		map[string]interfaces.EndorServiceAction{
			categoryID + "/create": NewTestAction("Category create", true),
			categoryID + "/read":   NewTestAction("Category read", true),
			categoryID + "/update": NewTestAction("Category update", true),
			categoryID + "/delete": NewTestAction("Category delete", true),
		},
	)
	return mockCategory
}

// Performance simulation helpers

// PerformanceMockService wraps a MockEndorService with latency simulation
// for integration testing scenarios.
type PerformanceMockService struct {
	*MockEndorService
	latency time.Duration
}

// NewPerformanceMockService creates a MockEndorService with simulated latency.
// The latency parameter controls how long each method call should take.
func NewPerformanceMockService(latency time.Duration) *PerformanceMockService {
	return &PerformanceMockService{
		MockEndorService: &MockEndorService{},
		latency:          latency,
	}
}

// GetResource simulates realistic latency for service resource retrieval.
func (p *PerformanceMockService) GetResource() string {
	time.Sleep(p.latency)
	return p.MockEndorService.GetResource()
}

// GetDescription simulates realistic latency for service description retrieval.
func (p *PerformanceMockService) GetDescription() string {
	time.Sleep(p.latency)
	return p.MockEndorService.GetDescription()
}

// Validate simulates realistic latency for service validation.
func (p *PerformanceMockService) Validate() error {
	time.Sleep(p.latency)
	return p.MockEndorService.Validate()
}
