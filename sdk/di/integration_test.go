package di

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test interfaces that simulate real dependencies

type MockRepository interface {
	Save(ctx context.Context, data interface{}) error
	Load(ctx context.Context, id string) (interface{}, error)
}

type MockConfig interface {
	GetDatabaseURL() string
	GetPort() int
}

type MockLogger interface {
	Info(message string)
	Error(message string, err error)
}

type MockService interface {
	ProcessData(data interface{}) error
	GetStatus() string
}

// Mock implementations

type mockRepositoryImpl struct {
	data map[string]interface{}
}

func (r *mockRepositoryImpl) Save(ctx context.Context, data interface{}) error {
	r.data["test"] = data
	return nil
}

func (r *mockRepositoryImpl) Load(ctx context.Context, id string) (interface{}, error) {
	return r.data[id], nil
}

type mockConfigImpl struct{}

func (c *mockConfigImpl) GetDatabaseURL() string {
	return "mongodb://localhost:27017"
}

func (c *mockConfigImpl) GetPort() int {
	return 8080
}

type mockLoggerImpl struct {
	logs []string
}

func (l *mockLoggerImpl) Info(message string) {
	l.logs = append(l.logs, "INFO: "+message)
}

func (l *mockLoggerImpl) Error(message string, err error) {
	l.logs = append(l.logs, "ERROR: "+message)
}

type mockServiceImpl struct {
	repo   MockRepository
	config MockConfig
	logger MockLogger
}

func (s *mockServiceImpl) ProcessData(data interface{}) error {
	s.logger.Info("Processing data")
	ctx := context.Background()
	return s.repo.Save(ctx, data)
}

func (s *mockServiceImpl) GetStatus() string {
	return "running on port " + string(rune(s.config.GetPort()))
}

func TestIntegration_ServiceWithDependencies(t *testing.T) {
	container := NewContainer()

	// Register dependencies
	repo := &mockRepositoryImpl{data: make(map[string]interface{})}
	err := Register[MockRepository](container, repo, Singleton)
	require.NoError(t, err)

	config := &mockConfigImpl{}
	err = Register[MockConfig](container, config, Singleton)
	require.NoError(t, err)

	logger := &mockLoggerImpl{logs: make([]string, 0)}
	err = Register[MockLogger](container, logger, Singleton)
	require.NoError(t, err)

	// Register service with factory that uses dependencies
	serviceFactory := func(c Container) (MockService, error) {
		repo, err := Resolve[MockRepository](c)
		if err != nil {
			return nil, err
		}

		config, err := Resolve[MockConfig](c)
		if err != nil {
			return nil, err
		}

		logger, err := Resolve[MockLogger](c)
		if err != nil {
			return nil, err
		}

		return &mockServiceImpl{
			repo:   repo,
			config: config,
			logger: logger,
		}, nil
	}

	err = RegisterFactory[MockService](container, serviceFactory, Singleton)
	require.NoError(t, err)

	// Validate the dependency graph
	validationErrors := container.Validate()
	assert.Empty(t, validationErrors, "Dependency graph should be valid")

	// Resolve and use the service
	service, err := Resolve[MockService](container)
	require.NoError(t, err)

	err = service.ProcessData("test data")
	assert.NoError(t, err)

	// Verify dependencies were used correctly
	resolvedLogger, err := Resolve[MockLogger](container)
	require.NoError(t, err)
	loggerImpl := resolvedLogger.(*mockLoggerImpl)
	assert.Contains(t, loggerImpl.logs, "INFO: Processing data")

	resolvedRepo, err := Resolve[MockRepository](container)
	require.NoError(t, err)
	repoImpl := resolvedRepo.(*mockRepositoryImpl)
	assert.Equal(t, "test data", repoImpl.data["test"])
}

func TestIntegration_DependencyGraphVisualization(t *testing.T) {
	container := NewContainer()

	// Register a hierarchy of dependencies
	err := Register[MockConfig](container, &mockConfigImpl{}, Singleton)
	require.NoError(t, err)

	err = Register[MockRepository](container, &mockRepositoryImpl{data: make(map[string]interface{})}, Singleton)
	require.NoError(t, err)

	loggerFactory := func(c Container) (MockLogger, error) {
		return &mockLoggerImpl{logs: make([]string, 0)}, nil
	}
	err = RegisterFactory[MockLogger](container, loggerFactory, Transient)
	require.NoError(t, err)

	serviceFactory := func(c Container) (MockService, error) {
		// This factory depends on other services
		repo, _ := Resolve[MockRepository](c)
		config, _ := Resolve[MockConfig](c)
		logger, _ := Resolve[MockLogger](c)

		return &mockServiceImpl{
			repo:   repo,
			config: config,
			logger: logger,
		}, nil
	}
	err = RegisterFactory[MockService](container, serviceFactory, Singleton)
	require.NoError(t, err)

	// Get dependency graph
	graph := container.GetDependencyGraph()

	// Verify all dependencies are represented
	assert.Len(t, graph, 4, "Should have 4 registered types")

	// Check that each dependency type is present
	foundTypes := make(map[string]bool)
	for typeName := range graph {
		if typeName == "di.MockConfig" || typeName == "di.MockRepository" {
			assert.Equal(t, []string{"<direct>"}, graph[typeName])
			foundTypes["direct"] = true
		} else if typeName == "di.MockLogger" || typeName == "di.MockService" {
			assert.Equal(t, []string{"<factory>"}, graph[typeName])
			foundTypes["factory"] = true
		}
	}

	assert.True(t, foundTypes["direct"], "Should have direct registrations")
	assert.True(t, foundTypes["factory"], "Should have factory registrations")
}

func TestIntegration_FactoryDependencyChain(t *testing.T) {
	container := NewContainer()

	// Create a chain where each factory depends on the previous one
	err := Register[MockConfig](container, &mockConfigImpl{}, Singleton)
	require.NoError(t, err)

	repoFactory := func(c Container) (MockRepository, error) {
		// Repository depends on config
		config, err := Resolve[MockConfig](c)
		if err != nil {
			return nil, err
		}

		_ = config // Use config (in real scenario, might use for connection string)
		return &mockRepositoryImpl{data: make(map[string]interface{})}, nil
	}
	err = RegisterFactory[MockRepository](container, repoFactory, Singleton)
	require.NoError(t, err)

	loggerFactory := func(c Container) (MockLogger, error) {
		// Logger depends on config
		config, err := Resolve[MockConfig](c)
		if err != nil {
			return nil, err
		}

		_ = config // Use config (in real scenario, might configure log level)
		return &mockLoggerImpl{logs: make([]string, 0)}, nil
	}
	err = RegisterFactory[MockLogger](container, loggerFactory, Transient)
	require.NoError(t, err)

	serviceFactory := func(c Container) (MockService, error) {
		// Service depends on repo and logger
		repo, err := Resolve[MockRepository](c)
		if err != nil {
			return nil, err
		}

		logger, err := Resolve[MockLogger](c)
		if err != nil {
			return nil, err
		}

		config, err := Resolve[MockConfig](c)
		if err != nil {
			return nil, err
		}

		return &mockServiceImpl{
			repo:   repo,
			config: config,
			logger: logger,
		}, nil
	}
	err = RegisterFactory[MockService](container, serviceFactory, Singleton)
	require.NoError(t, err)

	// Resolve the service - this should trigger the entire dependency chain
	service, err := Resolve[MockService](container)
	require.NoError(t, err)
	assert.NotNil(t, service)

	// Test that the service works with all its dependencies
	err = service.ProcessData("chain test")
	assert.NoError(t, err)

	// Verify singleton behavior - resolving again should return same instance
	service2, err := Resolve[MockService](container)
	require.NoError(t, err)
	assert.Same(t, service, service2, "Should return same singleton instance")
}

func TestIntegration_ErrorPropagation(t *testing.T) {
	container := NewContainer()

	// Register a factory that will fail
	failingFactory := func(c Container) (MockService, error) {
		return nil, assert.AnError // Return a known error
	}

	err := RegisterFactory[MockService](container, failingFactory, Singleton)
	require.NoError(t, err)

	// Try to resolve - should get the factory error wrapped in a dependency error
	_, err = Resolve[MockService](container)
	assert.Error(t, err)

	var diErr *DIError
	assert.True(t, errors.As(err, &diErr))
	assert.Contains(t, diErr.Error(), "factory function returned error")
}

func TestIntegration_ScopeInteraction(t *testing.T) {
	container := NewContainer()

	// Register config as singleton
	err := Register[MockConfig](container, &mockConfigImpl{}, Singleton)
	require.NoError(t, err)

	// Register logger factory as transient
	loggerFactory := func(c Container) (MockLogger, error) {
		return &mockLoggerImpl{logs: make([]string, 0)}, nil
	}
	err = RegisterFactory[MockLogger](container, loggerFactory, Transient)
	require.NoError(t, err)

	// Resolve multiple times
	config1, err := Resolve[MockConfig](container)
	require.NoError(t, err)
	config2, err := Resolve[MockConfig](container)
	require.NoError(t, err)
	assert.Same(t, config1, config2, "Singleton should return same instance")

	logger1, err := Resolve[MockLogger](container)
	require.NoError(t, err)
	logger2, err := Resolve[MockLogger](container)
	require.NoError(t, err)
	assert.NotSame(t, logger1, logger2, "Transient should return different instances")
}
