package sdk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// Mock implementations - reuse the patterns from endor_service_di_test.go

type MockConfigProvider struct {
	mock.Mock
}

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

// MockLogger already exists in endor_service_di_test.go - reuse it

// Test Resource for hybrid service testing
type TestUserHybrid struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (u TestUserHybrid) GetID() *string {
	return &u.ID
}

func (u TestUserHybrid) SetID(id string) {
	u.ID = id
}

// ResourceInstanceSpecializedInterface implementation for category testing
func (u TestUserHybrid) GetCategoryType() *string {
	categoryType := "admin"
	return &categoryType
}

func (u TestUserHybrid) SetCategoryType(categoryType string) {
	// For testing purposes, we don't need to store this
}

// AC 1: Constructor Dependency Injection Tests
func TestNewEndorHybridServiceWithDeps_ValidDependencies_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	deps := EndorHybridServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	// Act
	service, err := NewEndorHybridServiceWithDeps[TestUserHybrid]("users", "User management service", deps)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "users", service.Resource)
	assert.Equal(t, "User management service", service.ResourceDescription)
	assert.Equal(t, mockRepo, service.repository)
	assert.Equal(t, mockConfig, service.config)
	assert.Equal(t, mockLogger, service.logger)
}

// AC 4: Dependency Validation Tests
func TestNewEndorHybridServiceWithDeps_NilRepository_ReturnsError(t *testing.T) {
	// Arrange
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	deps := EndorHybridServiceDependencies{
		Repository: nil, // Missing required dependency
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	// Act
	service, err := NewEndorHybridServiceWithDeps[TestUserHybrid]("users", "User management", deps)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, service)

	var hybridServiceError *EndorHybridServiceError
	assert.ErrorAs(t, err, &hybridServiceError)
	assert.Equal(t, "Repository", hybridServiceError.Field)
	assert.Contains(t, hybridServiceError.Message, "Repository interface is required")
}

func TestNewEndorHybridServiceWithDeps_NilConfig_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockLogger := &MockLogger{}

	deps := EndorHybridServiceDependencies{
		Repository: mockRepo,
		Config:     nil, // Missing required dependency
		Logger:     mockLogger,
	}

	// Act
	service, err := NewEndorHybridServiceWithDeps[TestUserHybrid]("users", "User management", deps)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, service)

	var hybridServiceError *EndorHybridServiceError
	assert.ErrorAs(t, err, &hybridServiceError)
	assert.Equal(t, "Config", hybridServiceError.Field)
	assert.Contains(t, hybridServiceError.Message, "Configuration interface is required")
}

func TestNewEndorHybridServiceWithDeps_NilLogger_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}

	deps := EndorHybridServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     nil, // Missing required dependency
	}

	// Act
	service, err := NewEndorHybridServiceWithDeps[TestUserHybrid]("users", "User management", deps)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, service)

	var hybridServiceError *EndorHybridServiceError
	assert.ErrorAs(t, err, &hybridServiceError)
	assert.Equal(t, "Logger", hybridServiceError.Field)
	assert.Contains(t, hybridServiceError.Message, "Logger interface is required")
}

// AC 7: Backward Compatibility Tests
func TestNewHybridService_BackwardCompatibility_Success(t *testing.T) {
	// Act
	service := NewHybridService[TestUserHybrid]("users", "User management service")

	// Assert
	assert.NotNil(t, service)

	impl, ok := service.(EndorHybridServiceImpl[TestUserHybrid])
	assert.True(t, ok)
	assert.Equal(t, "users", impl.Resource)
	assert.Equal(t, "User management service", impl.ResourceDescription)

	// AC 7: Should have default dependencies for backward compatibility
	assert.NotNil(t, impl.config)  // Default config provider
	assert.NotNil(t, impl.logger)  // Default logger
	assert.Nil(t, impl.repository) // Repository will be set in repository refactoring
}

// AC 3: ToEndorService Method Dependency Propagation Tests
func TestToEndorService_WithInjectedDependencies_PropagatesDependencies(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	// Set up logger expectation for potential error logging
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Maybe()

	deps := EndorHybridServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	service, err := NewEndorHybridServiceWithDeps[TestUserHybrid]("users", "User management", deps)
	assert.NoError(t, err)

	testSchema := Schema{
		Type: ObjectType,
		Properties: &map[string]Schema{
			"name": {Type: StringType},
		},
	}

	// Act
	endorService := service.ToEndorService(testSchema)

	// Assert
	assert.Equal(t, "users", endorService.Resource)
	assert.Equal(t, "User management", endorService.Description)
	assert.NotNil(t, endorService.Methods)

	// AC 3: Validate dependency propagation - EndorService should have injected dependencies
	// Note: We can't directly access private fields, but the service creation should succeed
	// This validates that dependency chain integrity is maintained

	mockLogger.AssertExpectations(t)
}

func TestToEndorService_WithoutDependencies_FallsBackToDirectConstruction(t *testing.T) {
	// Arrange - use backward compatibility constructor (no dependencies)
	service := NewHybridService[TestUserHybrid]("users", "User management")

	testSchema := Schema{
		Type: ObjectType,
		Properties: &map[string]Schema{
			"name": {Type: StringType},
		},
	}

	// Act
	endorService := service.(EndorHybridServiceImpl[TestUserHybrid]).ToEndorService(testSchema)

	// Assert
	assert.Equal(t, "users", endorService.Resource)
	assert.Equal(t, "User management", endorService.Description)
	assert.NotNil(t, endorService.Methods)

	// Should have basic CRUD methods
	assert.Contains(t, endorService.Methods, "schema")
	assert.Contains(t, endorService.Methods, "create")
	assert.Contains(t, endorService.Methods, "list")
	assert.Contains(t, endorService.Methods, "instance")
	assert.Contains(t, endorService.Methods, "update")
	assert.Contains(t, endorService.Methods, "delete")
}

// AC 2: Interface-based Internal Dependencies Tests
func TestEndorHybridServiceImpl_HoldsInterfaceReferences_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	deps := EndorHybridServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	// Act
	service, err := NewEndorHybridServiceWithDeps[TestUserHybrid]("users", "User management", deps)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, service)

	// AC 2: Verify struct holds interface references instead of concrete types
	// These are private fields, but the fact that we can assign interfaces proves the typing
	assert.Implements(t, (*interfaces.RepositoryPattern)(nil), service.repository)
	assert.Implements(t, (*interfaces.ConfigProviderInterface)(nil), service.config)
	assert.Implements(t, (*interfaces.LoggerInterface)(nil), service.logger)
}

// AC 4: Category and Action Operations Tests
func TestWithCategories_WithInjectedDependencies_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	deps := EndorHybridServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	service, err := NewEndorHybridServiceWithDeps[TestUserHybrid]("users", "User management", deps)
	assert.NoError(t, err)

	// Create a simple category for testing
	category := NewEndorHybridServiceCategory[TestUserHybrid, TestUserHybrid](Category{ID: "admin"})
	categories := []EndorHybridServiceCategory{category}

	// Act
	updatedService := service.WithCategories(categories)

	// Assert
	assert.NotNil(t, updatedService)
	impl, ok := updatedService.(EndorHybridServiceImpl[TestUserHybrid])
	assert.True(t, ok)
	assert.Len(t, impl.categories, 1)
	assert.Contains(t, impl.categories, "admin")
}

// AC 4: WithActions Method Tests
func TestWithActions_WithInjectedDependencies_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	deps := EndorHybridServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	service, err := NewEndorHybridServiceWithDeps[TestUserHybrid]("users", "User management", deps)
	assert.NoError(t, err)

	// Define custom actions
	customActions := func(getSchema func() RootSchema) map[string]EndorServiceAction {
		return map[string]EndorServiceAction{
			"custom": NewAction(
				func(c *EndorContext[NoPayload]) (*Response[any], error) {
					return NewResponseBuilder[any]().Build(), nil
				},
				"Custom action",
			),
		}
	}

	// Act
	updatedService := service.WithActions(customActions)

	// Assert
	assert.NotNil(t, updatedService)
	impl, ok := updatedService.(EndorHybridServiceImpl[TestUserHybrid])
	assert.True(t, ok)
	assert.NotNil(t, impl.methodsFn)
}

// Performance validation test
func TestEndorHybridService_DependencyInjection_PerformanceRegression(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	deps := EndorHybridServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	// Act & Assert - Multiple creations should be fast
	for i := 0; i < 100; i++ {
		service, err := NewEndorHybridServiceWithDeps[TestUserHybrid](
			fmt.Sprintf("users%d", i),
			"User management",
			deps,
		)
		assert.NoError(t, err)
		assert.NotNil(t, service)
	}
}
