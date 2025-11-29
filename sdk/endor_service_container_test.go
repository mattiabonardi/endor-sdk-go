package sdk

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk/di"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/stretchr/testify/assert"
)

// Test cases for AC 5: Container Integration

func TestNewEndorServiceFromContainer_ValidContainer_CreatesService(t *testing.T) {
	// AC 5: Service constructors work seamlessly with DI container for automatic dependency resolution
	container := di.NewContainer()

	// Register dependencies in container
	mockRepo := &MockRepository{}
	mockConfig := &MockConfig{}
	mockLogger := &MockLogger{}

	err := di.Register[interfaces.RepositoryPattern](container, mockRepo, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.ConfigProviderInterface](container, mockConfig, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
	assert.NoError(t, err)

	// Create service from container
	methods := map[string]EndorServiceAction{
		"test": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().Build(), nil
		}, "Test action"),
	}

	service, err := NewEndorServiceFromContainer(container, "products", "Product service", methods)

	assert.NoError(t, err, "Should create service from container")
	assert.NotNil(t, service, "Should return service instance")
	assert.Equal(t, "products", service.Resource, "Should set resource name")
	assert.Equal(t, "Product service", service.Description, "Should set description")
	assert.Equal(t, mockRepo, service.GetRepository(), "Should resolve repository from container")
	assert.Equal(t, mockConfig, service.GetConfig(), "Should resolve config from container")
	assert.Equal(t, mockLogger, service.GetLogger(), "Should resolve logger from container")
}

func TestNewEndorServiceFromContainer_MissingRepository_ReturnsError(t *testing.T) {
	// AC 5: Container validation for EndorService dependency requirements
	container := di.NewContainer()

	// Register only some dependencies - missing repository
	mockConfig := &MockConfig{}
	mockLogger := &MockLogger{}

	err := di.Register[interfaces.ConfigProviderInterface](container, mockConfig, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
	assert.NoError(t, err)

	methods := map[string]EndorServiceAction{}

	_, err = NewEndorServiceFromContainer(container, "products", "Product service", methods)

	assert.Error(t, err, "Should return error when repository not registered")
	assert.Contains(t, err.Error(), "Repository", "Error should mention Repository dependency")
}

func TestNewEndorServiceFromContainer_MissingConfig_ReturnsError(t *testing.T) {
	// AC 5: Container validation handles missing dependencies
	container := di.NewContainer()

	// Register only some dependencies - missing config
	mockRepo := &MockRepository{}
	mockLogger := &MockLogger{}

	err := di.Register[interfaces.RepositoryPattern](container, mockRepo, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.LoggerInterface](container, mockLogger, di.Singleton)
	assert.NoError(t, err)

	methods := map[string]EndorServiceAction{}

	_, err = NewEndorServiceFromContainer(container, "products", "Product service", methods)

	assert.Error(t, err, "Should return error when config not registered")
	assert.Contains(t, err.Error(), "ConfigProvider", "Error should mention ConfigProvider dependency")
}

func TestNewEndorServiceFromContainer_MissingLogger_ReturnsError(t *testing.T) {
	// AC 5: Container validation handles missing dependencies
	container := di.NewContainer()

	// Register only some dependencies - missing logger
	mockRepo := &MockRepository{}
	mockConfig := &MockConfig{}

	err := di.Register[interfaces.RepositoryPattern](container, mockRepo, di.Singleton)
	assert.NoError(t, err)

	err = di.Register[interfaces.ConfigProviderInterface](container, mockConfig, di.Singleton)
	assert.NoError(t, err)

	methods := map[string]EndorServiceAction{}

	_, err = NewEndorServiceFromContainer(container, "products", "Product service", methods)

	assert.Error(t, err, "Should return error when logger not registered")
	assert.Contains(t, err.Error(), "Logger", "Error should mention Logger dependency")
}

func TestNewEndorServiceFromContainer_FactoryPattern_CreatesService(t *testing.T) {
	// AC 5: Support dependency override patterns for testing
	container := di.NewContainer()

	// Register dependencies using factories
	err := di.RegisterFactory[interfaces.RepositoryPattern](container, func(c di.Container) (interfaces.RepositoryPattern, error) {
		return &MockRepository{}, nil
	}, di.Singleton)
	assert.NoError(t, err)

	err = di.RegisterFactory[interfaces.ConfigProviderInterface](container, func(c di.Container) (interfaces.ConfigProviderInterface, error) {
		return &MockConfig{}, nil
	}, di.Singleton)
	assert.NoError(t, err)

	err = di.RegisterFactory[interfaces.LoggerInterface](container, func(c di.Container) (interfaces.LoggerInterface, error) {
		return &MockLogger{}, nil
	}, di.Singleton)
	assert.NoError(t, err)

	methods := map[string]EndorServiceAction{
		"test": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().Build(), nil
		}, "Test action"),
	}

	service, err := NewEndorServiceFromContainer(container, "orders", "Order service", methods)

	assert.NoError(t, err, "Should create service using factory dependencies")
	assert.NotNil(t, service, "Should return service instance")
	assert.Equal(t, "orders", service.Resource, "Should set resource name")
	assert.NotNil(t, service.GetRepository(), "Should have factory-created repository")
	assert.NotNil(t, service.GetConfig(), "Should have factory-created config")
	assert.NotNil(t, service.GetLogger(), "Should have factory-created logger")
}

func TestEndorServiceDI_IntegrationWithExistingWorkflow_WorksCorrectly(t *testing.T) {
	// Integration test demonstrating that dependency injection works with existing patterns

	// 1. Create service using traditional constructor (backward compatibility)
	traditionalService := NewEndorService("legacy", "Legacy service", map[string]EndorServiceAction{
		"action1": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().AddMessage(NewMessage(Info, "Legacy action")).Build(), nil
		}, "Legacy action"),
	})

	// 2. Create service using dependency injection
	mockRepo := &MockRepository{}
	mockConfig := &MockConfig{}
	mockLogger := &MockLogger{}

	deps := EndorServiceDependencies{
		Repository: mockRepo,
		Config:     mockConfig,
		Logger:     mockLogger,
	}

	diService, err := NewEndorServiceWithDeps("modern", "Modern service", map[string]EndorServiceAction{
		"action1": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().AddMessage(NewMessage(Info, "Modern action")).Build(), nil
		}, "Modern action"),
	}, deps)
	assert.NoError(t, err)

	// 3. Create service using container
	container := di.NewContainer()
	err = di.Register[interfaces.RepositoryPattern](container, &MockRepository{}, di.Singleton)
	assert.NoError(t, err)
	err = di.Register[interfaces.ConfigProviderInterface](container, &MockConfig{}, di.Singleton)
	assert.NoError(t, err)
	err = di.Register[interfaces.LoggerInterface](container, &MockLogger{}, di.Singleton)
	assert.NoError(t, err)

	containerService, err := NewEndorServiceFromContainer(container, "container", "Container service", map[string]EndorServiceAction{
		"action1": NewAction(func(c *EndorContext[NoPayload]) (*Response[any], error) {
			return NewResponseBuilder[any]().AddMessage(NewMessage(Info, "Container action")).Build(), nil
		}, "Container action"),
	})
	assert.NoError(t, err)

	// All services should be valid and work correctly
	services := []EndorService{traditionalService, *diService, *containerService}

	for _, service := range services {
		assert.NotEmpty(t, service.Resource, "Service should have resource name")
		assert.NotEmpty(t, service.Description, "Service should have description")
		assert.NotEmpty(t, service.Methods, "Service should have methods")

		// Interface methods should work
		assert.Equal(t, service.Resource, service.GetResource())
		assert.Equal(t, service.Description, service.GetDescription())
		assert.Equal(t, service.Methods, service.GetMethods())
		assert.NotNil(t, service.GetMethods()["action1"], "Should have test action")
	}
}
