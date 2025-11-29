package testutils

import (
	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/stretchr/testify/mock"
)

// MockEndorService provides a mock implementation of EndorServiceInterface
// for testing purposes. It uses testify/mock for behavior verification and
// call tracking.
//
// Example usage:
//
//	mockService := &MockEndorService{}
//	mockService.On("GetResource").Return("users")
//	mockService.On("GetDescription").Return("User service")
//	mockService.On("Validate").Return(nil)
//
//	// Use in your test
//	result := myFunction(mockService)
//
//	// Assert expectations
//	mockService.AssertExpectations(t)
type MockEndorService struct {
	mock.Mock
}

// Compile-time interface compliance check
var _ interfaces.EndorServiceInterface = (*MockEndorService)(nil)

func (m *MockEndorService) GetResource() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEndorService) GetDescription() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEndorService) GetMethods() map[string]interfaces.EndorServiceAction {
	args := m.Called()
	if actions := args.Get(0); actions != nil {
		return actions.(map[string]interfaces.EndorServiceAction)
	}
	return map[string]interfaces.EndorServiceAction{}
}

func (m *MockEndorService) GetPriority() *int {
	args := m.Called()
	if priority := args.Get(0); priority != nil {
		return priority.(*int)
	}
	return nil
}

func (m *MockEndorService) GetVersion() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEndorService) Validate() error {
	args := m.Called()
	return args.Error(0)
}

// MockEndorHybridService provides a mock implementation of EndorHybridServiceInterface
// for testing hybrid services with category-based specialization and dynamic actions.
//
// Example usage:
//
//	mockHybridService := &MockEndorHybridService{}
//	mockHybridService.On("GetResource").Return("products")
//	mockHybridService.On("GetResourceDescription").Return("Product management")
//	mockHybridService.On("WithCategories", mock.Anything).Return(mockHybridService)
//	mockHybridService.On("ToEndorService", mock.Anything).Return(mockEndorService)
type MockEndorHybridService struct {
	mock.Mock
}

// Compile-time interface compliance check
var _ interfaces.EndorHybridServiceInterface = (*MockEndorHybridService)(nil)

func (m *MockEndorHybridService) GetResource() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEndorHybridService) GetResourceDescription() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEndorHybridService) GetPriority() *int {
	args := m.Called()
	if priority := args.Get(0); priority != nil {
		return priority.(*int)
	}
	return nil
}

func (m *MockEndorHybridService) WithCategories(categories []interfaces.EndorHybridServiceCategory) interfaces.EndorHybridServiceInterface {
	args := m.Called(categories)
	if result := args.Get(0); result != nil {
		return result.(interfaces.EndorHybridServiceInterface)
	}
	return m
}

func (m *MockEndorHybridService) WithActions(fn func(getSchema func() interfaces.RootSchema) map[string]interfaces.EndorServiceAction) interfaces.EndorHybridServiceInterface {
	args := m.Called(fn)
	if result := args.Get(0); result != nil {
		return result.(interfaces.EndorHybridServiceInterface)
	}
	return m
}

func (m *MockEndorHybridService) ToEndorService(metadataSchema interfaces.Schema) interfaces.EndorServiceInterface {
	args := m.Called(metadataSchema)
	return args.Get(0).(interfaces.EndorServiceInterface)
}

func (m *MockEndorHybridService) Validate() error {
	args := m.Called()
	return args.Error(0)
}

// MockConfigProvider provides a mock implementation of ConfigProviderInterface
// for testing configuration-dependent functionality.
//
// Example usage:
//
//	mockConfig := &MockConfigProvider{}
//	mockConfig.On("GetServerPort").Return("8080")
//	mockConfig.On("GetDocumentDBUri").Return("mongodb://test:27017")
//	mockConfig.On("IsHybridResourcesEnabled").Return(true)
type MockConfigProvider struct {
	mock.Mock
}

// Compile-time interface compliance check
var _ interfaces.ConfigProviderInterface = (*MockConfigProvider)(nil)

func (m *MockConfigProvider) GetServerPort() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigProvider) GetDocumentDBUri() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigProvider) IsHybridResourcesEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConfigProvider) IsDynamicResourcesEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConfigProvider) GetDynamicResourceDocumentDBName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigProvider) Reload() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfigProvider) Validate() error {
	args := m.Called()
	return args.Error(0)
}

// MockEndorContext provides a mock implementation of EndorContextInterface[T]
// for testing context propagation and request handling.
//
// Example usage:
//
//	mockContext := &MockEndorContext[UserPayload]{}
//	mockContext.On("GetMicroServiceId").Return("user-service")
//	mockContext.On("GetSession").Return(testSession)
//	mockContext.On("GetPayload").Return(UserPayload{Name: "Test"})
type MockEndorContext[T any] struct {
	mock.Mock
}

// Compile-time interface compliance check
var _ interfaces.EndorContextInterface[any] = (*MockEndorContext[any])(nil)

func (m *MockEndorContext[T]) GetMicroServiceId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEndorContext[T]) GetSession() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockEndorContext[T]) GetPayload() T {
	args := m.Called()
	return args.Get(0).(T)
}

func (m *MockEndorContext[T]) SetPayload(payload T) {
	m.Called(payload)
}

func (m *MockEndorContext[T]) GetResourceMetadataSchema() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockEndorContext[T]) GetCategoryID() *string {
	args := m.Called()
	if categoryID := args.Get(0); categoryID != nil {
		return categoryID.(*string)
	}
	return nil
}

func (m *MockEndorContext[T]) SetCategoryID(categoryID *string) {
	m.Called(categoryID)
}

func (m *MockEndorContext[T]) GetGinContext() *gin.Context {
	args := m.Called()
	if ginCtx := args.Get(0); ginCtx != nil {
		return ginCtx.(*gin.Context)
	}
	return nil
}

// MockEndorServiceAction provides a mock implementation of EndorServiceAction
// for testing service action behavior and HTTP callback creation.
//
// Example usage:
//
//	mockAction := &MockEndorServiceAction{}
//	mockAction.On("GetOptions").Return(interfaces.EndorServiceActionOptions{
//		Description: "Test action",
//		Public: true,
//	})
//	mockAction.On("CreateHTTPCallback", "test-service").Return(testHandler)
type MockEndorServiceAction struct {
	mock.Mock
}

// Compile-time interface compliance check
var _ interfaces.EndorServiceAction = (*MockEndorServiceAction)(nil)

func (m *MockEndorServiceAction) CreateHTTPCallback(microserviceId string) func(c *gin.Context) {
	args := m.Called(microserviceId)
	if callback := args.Get(0); callback != nil {
		return callback.(func(c *gin.Context))
	}
	return func(c *gin.Context) {}
}

func (m *MockEndorServiceAction) GetOptions() interfaces.EndorServiceActionOptions {
	args := m.Called()
	return args.Get(0).(interfaces.EndorServiceActionOptions)
}

// MockEndorHybridServiceCategory provides a mock implementation of EndorHybridServiceCategory
// for testing category-based specializations in hybrid services.
//
// Example usage:
//
//	mockCategory := &MockEndorHybridServiceCategory{}
//	mockCategory.On("GetID").Return("admin")
//	mockCategory.On("CreateDefaultActions", "users", "User management", mock.Anything).Return(testActions)
type MockEndorHybridServiceCategory struct {
	mock.Mock
}

// Compile-time interface compliance check
var _ interfaces.EndorHybridServiceCategory = (*MockEndorHybridServiceCategory)(nil)

func (m *MockEndorHybridServiceCategory) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEndorHybridServiceCategory) CreateDefaultActions(resource string, resourceDescription string, metadataSchema interfaces.Schema) map[string]interfaces.EndorServiceAction {
	args := m.Called(resource, resourceDescription, metadataSchema)
	if actions := args.Get(0); actions != nil {
		return actions.(map[string]interfaces.EndorServiceAction)
	}
	return map[string]interfaces.EndorServiceAction{}
}
