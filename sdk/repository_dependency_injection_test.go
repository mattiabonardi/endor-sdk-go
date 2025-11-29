package sdk

import (
	"context"
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
)

// MockDatabaseClient implements DatabaseClientInterface for testing
type MockDatabaseClient struct {
	mock.Mock
}

func (m *MockDatabaseClient) Collection(name string) interfaces.CollectionInterface {
	args := m.Called(name)
	return args.Get(0).(interfaces.CollectionInterface)
}

func (m *MockDatabaseClient) Database(name string) interfaces.DatabaseInterface {
	args := m.Called(name)
	return args.Get(0).(interfaces.DatabaseInterface)
}

func (m *MockDatabaseClient) StartTransaction(ctx context.Context) (interfaces.TransactionInterface, error) {
	args := m.Called(ctx)
	return args.Get(0).(interfaces.TransactionInterface), args.Error(1)
}

func (m *MockDatabaseClient) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDatabaseClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockCollectionDI implements CollectionInterface for testing (different from existing MockCollection)
type MockCollectionDI struct {
	mock.Mock
}

func (m *MockCollectionDI) FindOne(ctx context.Context, filter interface{}) interfaces.SingleResultInterface {
	args := m.Called(ctx, filter)
	return args.Get(0).(interfaces.SingleResultInterface)
}

func (m *MockCollectionDI) Find(ctx context.Context, filter interface{}) (interfaces.CursorInterface, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(interfaces.CursorInterface), args.Error(1)
}

func (m *MockCollectionDI) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollectionDI) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCollectionDI) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollectionDI) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCollectionDI) Aggregate(ctx context.Context, pipeline interface{}) (interfaces.CursorInterface, error) {
	args := m.Called(ctx, pipeline)
	return args.Get(0).(interfaces.CursorInterface), args.Error(1)
}

// Test repository dependency injection with mock dependencies
// Validates Acceptance Criteria 1: Repository constructors accept DatabaseClientInterface and ConfigInterface
func TestEndorServiceRepositoryWithDependencyInjection(t *testing.T) {
	// Setup mock dependencies - reuse existing MockConfigProvider and MockLogger
	mockDBClient := &MockDatabaseClient{}
	mockConfig := &MockConfigProvider{} // Defined in existing test files
	mockLogger := &MockLogger{}         // Defined in existing test files

	// Configure mock expectations
	mockConfig.On("IsHybridResourcesEnabled").Return(false)
	mockConfig.On("IsDynamicResourcesEnabled").Return(false)

	deps := interfaces.RepositoryDependencies{
		DatabaseClient: mockDBClient,
		Config:         mockConfig,
		Logger:         mockLogger,
		MicroServiceID: "test-service",
	}

	// Test repository creation with dependency injection
	repo, err := NewEndorServiceRepositoryWithDependencies(deps, nil, nil)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, "test-service", repo.microServiceId)
	assert.Equal(t, deps, repo.dependencies)

	// Verify mock expectations
	mockConfig.AssertExpectations(t)
}

// Test repository dependency validation
// Validates Acceptance Criteria 1: Add repository dependency validation with structured error handling
func TestRepositoryDependencyValidation(t *testing.T) {
	t.Run("Valid dependencies", func(t *testing.T) {
		deps := interfaces.RepositoryDependencies{
			DatabaseClient: &MockDatabaseClient{},
			Config:         &MockConfigProvider{},
			Logger:         &MockLogger{},
			MicroServiceID: "test-service",
		}

		err := validateRepositoryDependencies(deps)
		assert.NoError(t, err)
	})

	t.Run("Missing config", func(t *testing.T) {
		deps := interfaces.RepositoryDependencies{
			DatabaseClient: &MockDatabaseClient{},
			Config:         nil,
			Logger:         &MockLogger{},
			MicroServiceID: "test-service",
		}

		err := validateRepositoryDependencies(deps)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ConfigProviderInterface is required")
	})

	t.Run("Missing microservice ID", func(t *testing.T) {
		deps := interfaces.RepositoryDependencies{
			DatabaseClient: &MockDatabaseClient{},
			Config:         &MockConfigProvider{},
			Logger:         &MockLogger{},
			MicroServiceID: "",
		}

		err := validateRepositoryDependencies(deps)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MicroServiceID is required")
	})

	t.Run("Optional dependencies allowed", func(t *testing.T) {
		deps := interfaces.RepositoryDependencies{
			DatabaseClient: nil, // Database client can be nil
			Config:         &MockConfigProvider{},
			Logger:         nil, // Logger can be nil
			MicroServiceID: "test-service",
		}

		err := validateRepositoryDependencies(deps)
		assert.NoError(t, err)
	})
}

// Test backward compatibility with existing constructor
// Validates Acceptance Criteria 6: Backward compatibility with convenience constructors
func TestBackwardCompatibilityConstructor(t *testing.T) {
	// This test ensures the old constructor still works
	repo := NewEndorServiceRepository("test-service", nil, nil)

	assert.NotNil(t, repo)
	assert.Equal(t, "test-service", repo.microServiceId)
	// The repository should be created even if dependencies are nil (for backward compatibility)
}

// Test MongoDB database client implementation
// Validates Acceptance Criteria 3: All MongoDB operations use injected DatabaseClientInterface
func TestMongoDatabaseClientImplementation(t *testing.T) {
	// Note: This would typically use a test MongoDB instance
	// For unit testing, we're validating the interface compliance

	t.Run("Interface compliance", func(t *testing.T) {
		// This test verifies that MongoDatabaseClient implements DatabaseClientInterface
		var client interfaces.DatabaseClientInterface
		client = NewMongoDatabaseClient(nil) // nil client for interface test only
		assert.NotNil(t, client)
	})
}

// Test repository factory patterns
// Validates Acceptance Criteria 5: Repository Factory Patterns support both patterns
func TestRepositoryFactoryPatterns(t *testing.T) {
	t.Run("NewRepositoryWithClient direct construction", func(t *testing.T) {
		mockDBClient := &MockDatabaseClient{}
		mockConfig := &MockConfigProvider{}

		// Set up mock expectations
		mockConfig.On("IsHybridResourcesEnabled").Return(false)
		mockConfig.On("IsDynamicResourcesEnabled").Return(false)

		repo, err := NewRepositoryWithClient(mockDBClient, mockConfig, "test-service")

		assert.NoError(t, err)
		assert.NotNil(t, repo)

		// Verify it implements RepositoryInterface
		var repoInterface interfaces.RepositoryInterface = repo
		assert.NotNil(t, repoInterface)

		// Verify mock expectations
		mockConfig.AssertExpectations(t)
	})

	t.Run("Default dependencies creation", func(t *testing.T) {
		deps, err := NewDefaultRepositoryDependencies("test-service")

		// Should succeed even if MongoDB is not available (backward compatibility)
		assert.NoError(t, err)
		assert.NotNil(t, deps.Config)
		assert.NotNil(t, deps.Logger)
		assert.Equal(t, "test-service", deps.MicroServiceID)
		// DatabaseClient may be nil if MongoDB is not available - this is acceptable
	})
}

// Test error handling and domain error abstraction
// Validates repository error patterns
func TestRepositoryErrorHandling(t *testing.T) {
	t.Run("Repository error creation", func(t *testing.T) {
		err := interfaces.NewDatabaseRepositoryError(
			interfaces.RepositoryErrorCodeInvalidDependencies,
			"Test error message",
			nil,
		)

		assert.Equal(t, "Test error message", err.Error())
		assert.Equal(t, interfaces.RepositoryErrorCodeInvalidDependencies, err.ErrorCode())
		assert.Equal(t, "Test error message", err.ErrorMessage())
	})

	t.Run("Repository error with cause", func(t *testing.T) {
		cause := assert.AnError
		err := interfaces.NewDatabaseRepositoryError(
			interfaces.RepositoryErrorCodeDatabaseConnection,
			"Database connection failed",
			cause,
		)

		assert.Contains(t, err.Error(), "Database connection failed")
		assert.Contains(t, err.Error(), cause.Error())
	})
}
