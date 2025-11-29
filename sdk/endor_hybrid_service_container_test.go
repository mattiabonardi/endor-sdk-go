package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattiabonardi/endor-sdk-go/sdk/di"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// AC 8: Container Integration Tests for EndorHybridService

func TestNewEndorHybridServiceFromContainer_ValidContainer_Success(t *testing.T) {
	// Arrange
	container := di.NewContainer()

	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	// Register dependencies in container
	err := di.Register[interfaces.RepositoryPattern](container, mockRepo, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.ConfigProviderInterface](container, mockConfig, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
	assert.NoError(t, err)

	// Act
	service, err := NewEndorHybridServiceFromContainer[TestUserHybrid](
		container,
		"users",
		"User management service",
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "users", service.Resource)
	assert.Equal(t, "User management service", service.ResourceDescription)
	assert.Equal(t, mockRepo, service.repository)
	assert.Equal(t, mockConfig, service.config)
	assert.Equal(t, mockLogger, service.logger)
}

func TestNewEndorHybridServiceFromContainer_MissingRepositoryDependency_ReturnsError(t *testing.T) {
	// Arrange
	container := di.NewContainer()

	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	// Register only config and logger, missing repository
	err := di.Register[interfaces.ConfigProviderInterface](container, mockConfig, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
	assert.NoError(t, err)

	// Act
	service, err := NewEndorHybridServiceFromContainer[TestUserHybrid](
		container,
		"users",
		"User management service",
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to resolve Repository dependency")
}

func TestNewEndorHybridServiceFromContainer_MissingConfigDependency_ReturnsError(t *testing.T) {
	// Arrange
	container := di.NewContainer()

	mockRepo := &MockRepository{}
	mockLogger := &MockLogger{}

	// Register only repository and logger, missing config
	err := di.Register[interfaces.RepositoryPattern](container, mockRepo, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
	assert.NoError(t, err)

	// Act
	service, err := NewEndorHybridServiceFromContainer[TestUserHybrid](
		container,
		"users",
		"User management service",
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to resolve ConfigProvider dependency")
}

func TestNewEndorHybridServiceFromContainer_MissingLoggerDependency_ReturnsError(t *testing.T) {
	// Arrange
	container := di.NewContainer()

	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}

	// Register only repository and config, missing logger
	err := di.Register[interfaces.RepositoryPattern](container, mockRepo, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.ConfigProviderInterface](container, mockConfig, di.Singleton)
	assert.NoError(t, err)

	// Act
	service, err := NewEndorHybridServiceFromContainer[TestUserHybrid](
		container,
		"users",
		"User management service",
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to resolve Logger dependency")
}

// Integration test validating the complete dependency chain
func TestEndorHybridService_ContainerIntegration_ToEndorServiceChain(t *testing.T) {
	// Arrange
	container := di.NewContainer()

	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	// Set up logger expectation for potential error logging in ToEndorService
	mockLogger.On("Error", "Failed to create EndorService with dependencies", "error", "mock error").Maybe()

	// Register dependencies in container
	err := di.Register[interfaces.RepositoryPattern](container, mockRepo, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.ConfigProviderInterface](container, mockConfig, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
	assert.NoError(t, err)

	// Create hybrid service from container
	hybridService, err := NewEndorHybridServiceFromContainer[TestUserHybrid](
		container,
		"users",
		"User management service",
	)
	assert.NoError(t, err)

	testSchema := Schema{
		Type: ObjectType,
		Properties: &map[string]Schema{
			"name": {Type: StringType},
		},
	}

	// Act - Test dependency propagation through ToEndorService
	endorService := hybridService.ToEndorService(testSchema)

	// Assert
	assert.Equal(t, "users", endorService.Resource)
	assert.Equal(t, "User management service", endorService.Description)
	assert.NotNil(t, endorService.Methods)

	// Validate that the service has expected CRUD methods
	assert.Contains(t, endorService.Methods, "schema")
	assert.Contains(t, endorService.Methods, "create")
	assert.Contains(t, endorService.Methods, "list")
	assert.Contains(t, endorService.Methods, "instance")
	assert.Contains(t, endorService.Methods, "update")
	assert.Contains(t, endorService.Methods, "delete")

	mockLogger.AssertExpectations(t)
}

// Test validating container scope management
func TestEndorHybridService_ContainerScopes_SingletonBehavior(t *testing.T) {
	// Arrange
	container := di.NewContainer()

	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	// Register dependencies as singletons
	err := di.Register[interfaces.RepositoryPattern](container, mockRepo, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.ConfigProviderInterface](container, mockConfig, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
	assert.NoError(t, err)

	// Act - Create multiple services from same container
	service1, err1 := NewEndorHybridServiceFromContainer[TestUserHybrid](
		container,
		"users",
		"User management 1",
	)
	assert.NoError(t, err1)

	service2, err2 := NewEndorHybridServiceFromContainer[TestUserHybrid](
		container,
		"accounts",
		"Account management 2",
	)
	assert.NoError(t, err2)

	// Assert - Both services should share the same singleton dependencies
	assert.Same(t, service1.repository, service2.repository)
	assert.Same(t, service1.config, service2.config)
	assert.Same(t, service1.logger, service2.logger)

	// But services themselves should be different instances
	assert.NotSame(t, service1, service2)
	assert.NotEqual(t, service1.Resource, service2.Resource)
}

// Performance validation for container resolution
func TestEndorHybridService_ContainerResolution_Performance(t *testing.T) {
	// Arrange
	container := di.NewContainer()

	mockRepo := &MockRepository{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	err := di.Register[interfaces.RepositoryPattern](container, mockRepo, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.ConfigProviderInterface](container, mockConfig, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
	assert.NoError(t, err)

	// Act & Assert - Multiple resolutions should be fast
	for i := 0; i < 50; i++ {
		service, err := NewEndorHybridServiceFromContainer[TestUserHybrid](
			container,
			"testservice",
			"Test service",
		)
		assert.NoError(t, err)
		assert.NotNil(t, service)
	}
}
