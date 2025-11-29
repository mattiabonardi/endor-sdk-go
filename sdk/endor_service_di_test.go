package sdk

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetID() *string {
	args := m.Called()
	if id := args.Get(0); id != nil {
		return id.(*string)
	}
	return nil
}

func (m *MockRepository) SetID(id string) {
	m.Called(id)
}

func (m *MockRepository) GetCategoryType() *string {
	args := m.Called()
	if cat := args.Get(0); cat != nil {
		return cat.(*string)
	}
	return nil
}

func (m *MockRepository) SetCategoryType(categoryType string) {
	m.Called(categoryType)
}

type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetServerPort() string {
	return m.Called().String(0)
}

func (m *MockConfig) GetDocumentDBUri() string {
	return m.Called().String(0)
}

func (m *MockConfig) IsHybridResourcesEnabled() bool {
	return m.Called().Bool(0)
}

func (m *MockConfig) IsDynamicResourcesEnabled() bool {
	return m.Called().Bool(0)
}

func (m *MockConfig) GetDynamicResourceDocumentDBName() string {
	return m.Called().String(0)
}

func (m *MockConfig) Reload() error {
	return m.Called().Error(0)
}

func (m *MockConfig) Validate() error {
	return m.Called().Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) With(keysAndValues ...interface{}) interfaces.LoggerInterface {
	m.Called(keysAndValues...)
	return m
}

func (m *MockLogger) WithName(name string) interfaces.LoggerInterface {
	m.Called(name)
	return m
}

// Test cases for AC 1: Constructor Dependency Injection

func TestNewEndorServiceWithDeps_ValidDependencies_CreatesService(t *testing.T) {
	// AC 1: NewEndorServiceWithDeps() accepts all required dependencies as interface parameters with type safety
	mockRepo := &MockRepository{}
	mockConfig := &MockConfig{}
	mockLogger := &MockLogger{}

	deps := EndorServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	methods := map[string]EndorServiceAction{
		"test": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().Build(), nil
		}, "Test action"),
	}

	service, err := NewEndorServiceWithDeps("users", "User service", methods, deps)

	assert.NoError(t, err, "Should create service with valid dependencies")
	assert.NotNil(t, service, "Should return service instance")
	assert.Equal(t, "users", service.Resource, "Should set resource name")
	assert.Equal(t, "User service", service.Description, "Should set description")
	assert.Equal(t, methods, service.Methods, "Should set methods")
	assert.Equal(t, mockRepo, service.GetRepository(), "Should inject repository")
	assert.Equal(t, mockConfig, service.GetConfig(), "Should inject config")
	assert.Equal(t, mockLogger, service.GetLogger(), "Should inject logger")
}

// Test cases for AC 4: Dependency Validation

func TestNewEndorServiceWithDeps_NilRepository_ReturnsError(t *testing.T) {
	// AC 4: Constructor validates required dependencies and provides clear error messages
	mockConfig := &MockConfig{}
	mockLogger := &MockLogger{}

	deps := EndorServiceDependencies{
		Repository: nil, // Missing required dependency
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	methods := map[string]EndorServiceAction{}

	_, err := NewEndorServiceWithDeps("users", "User service", methods, deps)

	assert.Error(t, err, "Should return error for nil repository")
	assert.Contains(t, err.Error(), "Repository", "Error should mention Repository field")
	assert.Contains(t, err.Error(), "required for data access", "Error should explain purpose")
}

func TestNewEndorServiceWithDeps_NilConfig_ReturnsError(t *testing.T) {
	// AC 4: Dependency validation with structured error messages
	mockRepo := &MockRepository{}
	mockLogger := &MockLogger{}

	deps := EndorServiceDependencies{
		Repository: mockRepo,
		Config:     nil, // Missing required dependency
		Logger:     mockLogger,
	}

	methods := map[string]EndorServiceAction{}

	_, err := NewEndorServiceWithDeps("users", "User service", methods, deps)

	assert.Error(t, err, "Should return error for nil config")
	assert.Contains(t, err.Error(), "Config", "Error should mention Config field")
	assert.Contains(t, err.Error(), "configuration access", "Error should explain purpose")
}

func TestNewEndorServiceWithDeps_NilLogger_ReturnsError(t *testing.T) {
	// AC 4: Dependency validation with structured error messages
	mockRepo := &MockRepository{}
	mockConfig := &MockConfig{}

	deps := EndorServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     nil, // Missing required dependency
	}

	methods := map[string]EndorServiceAction{}

	_, err := NewEndorServiceWithDeps("users", "User service", methods, deps)

	assert.Error(t, err, "Should return error for nil logger")
	assert.Contains(t, err.Error(), "Logger", "Error should mention Logger field")
	assert.Contains(t, err.Error(), "service logging", "Error should explain purpose")
}

// Test cases for AC 3: Backward Compatibility

func TestNewEndorService_BackwardCompatibility_CreatesServiceWithDefaults(t *testing.T) {
	// AC 3: NewEndorService() convenience function maintains existing creation patterns with default implementations
	methods := map[string]EndorServiceAction{
		"test": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().Build(), nil
		}, "Test action"),
	}

	service := NewEndorService("users", "User service", methods)

	assert.Equal(t, "users", service.Resource, "Should set resource name")
	assert.Equal(t, "User service", service.Description, "Should set description")
	assert.Equal(t, methods, service.Methods, "Should set methods")
	assert.NotNil(t, service.GetConfig(), "Should have default config")
	assert.NotNil(t, service.GetLogger(), "Should have default logger")
	// Repository is nil until repository refactoring is complete
}

// Test cases for AC 6: Method Compatibility

func TestEndorService_GetResource_ReturnsResourceName(t *testing.T) {
	// AC 6: All existing EndorService methods work unchanged using injected interface dependencies
	service := NewEndorService("products", "Product service", map[string]EndorServiceAction{})

	result := service.GetResource()

	assert.Equal(t, "products", result, "Should return resource name")
}

func TestEndorService_GetDescription_ReturnsDescription(t *testing.T) {
	// AC 6: Method compatibility maintained
	service := NewEndorService("products", "Product management", map[string]EndorServiceAction{})

	result := service.GetDescription()

	assert.Equal(t, "Product management", result, "Should return description")
}

func TestEndorService_Validate_WithValidService_ReturnsNoError(t *testing.T) {
	// AC 6: Validation method works with injected dependencies
	methods := map[string]EndorServiceAction{
		"create": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().Build(), nil
		}, "Create action"),
	}
	service := NewEndorService("products", "Product service", methods)

	err := service.Validate()

	assert.NoError(t, err, "Valid service should pass validation")
}

func TestEndorService_Validate_WithEmptyResource_ReturnsError(t *testing.T) {
	// AC 6: Validation uses structured error messages
	methods := map[string]EndorServiceAction{
		"test": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().Build(), nil
		}, "Test action"),
	}
	service := NewEndorService("", "Empty resource service", methods)

	err := service.Validate()

	assert.Error(t, err, "Should return error for empty resource name")
	assert.Contains(t, err.Error(), "Resource", "Error should mention Resource field")
}

func TestEndorService_Validate_WithEmptyMethods_ReturnsError(t *testing.T) {
	// AC 6: Validation checks required fields
	service := NewEndorService("products", "Product service", map[string]EndorServiceAction{})

	err := service.Validate()

	assert.Error(t, err, "Should return error for empty methods")
	assert.Contains(t, err.Error(), "Methods", "Error should mention Methods field")
}

// Test cases for AC 2: Interface-based Internal Dependencies

func TestEndorServiceWithDeps_UsesInjectedDependencies(t *testing.T) {
	// AC 2: EndorService struct holds interface references instead of concrete types
	mockRepo := &MockRepository{}
	mockConfig := &MockConfig{}
	mockLogger := &MockLogger{}

	// Set up mock expectations
	mockConfig.On("Validate").Return(nil)
	mockLogger.On("Debug", "EndorService validation", "resource", "users", "methods", 1).Return()

	deps := EndorServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	methods := map[string]EndorServiceAction{
		"test": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().Build(), nil
		}, "Test action"),
	}

	service, err := NewEndorServiceWithDeps("users", "User service", methods, deps)
	assert.NoError(t, err)

	// Test that validation uses injected dependencies
	err = service.Validate()
	assert.NoError(t, err)

	// Verify mock calls
	mockConfig.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
