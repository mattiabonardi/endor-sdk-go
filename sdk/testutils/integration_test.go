package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

// TestIntegrationInMemoryRepository demonstrates in-memory repository testing
func TestIntegrationInMemoryRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("CRUD Operations", func(t *testing.T) {
		// Setup: Create repository
		repo := NewInMemoryRepository[TestUserPayload]()

		// Test Create
		user := TestUserPayload{
			Name:   "Integration Test User",
			Email:  "integration@example.com",
			Active: true,
		}

		id, err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)

		// Test FindByID
		retrievedUser, err := repo.FindByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, user.Name, retrievedUser.Name)
		assert.Equal(t, user.Email, retrievedUser.Email)
		assert.Equal(t, user.Active, retrievedUser.Active)

		// Test Update
		updatedUser := TestUserPayload{
			Name:   "Updated Integration User",
			Email:  "updated@example.com",
			Active: false,
		}
		err = repo.Update(ctx, id, updatedUser)
		assert.NoError(t, err)

		// Verify update
		retrievedUpdatedUser, err := repo.FindByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, updatedUser.Name, retrievedUpdatedUser.Name)
		assert.Equal(t, updatedUser.Email, retrievedUpdatedUser.Email)
		assert.Equal(t, updatedUser.Active, retrievedUpdatedUser.Active)

		// Test FindAll
		users, err := repo.FindAll(ctx, bson.M{})
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		// Test Count
		count, err := repo.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Test Delete
		err = repo.Delete(ctx, id)
		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.FindByID(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Verify empty repository
		count, err = repo.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Multiple Items", func(t *testing.T) {
		// Setup: Create repository with multiple items
		repo := NewInMemoryRepository[TestProductPayload]()

		products := GetTestProducts()
		var ids []string

		// Create multiple products
		for _, product := range products {
			id, err := repo.Create(ctx, product)
			assert.NoError(t, err)
			ids = append(ids, id)
		}

		// Verify count
		count, err := repo.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(products)), count)

		// Test FindAll
		retrievedProducts, err := repo.FindAll(ctx, bson.M{})
		assert.NoError(t, err)
		assert.Len(t, retrievedProducts, len(products))

		// Test clearing repository
		repo.Clear()
		count, err = repo.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// TestIntegrationDatabase demonstrates database integration testing
func TestIntegrationDatabase(t *testing.T) {
	t.Run("Database Setup", func(t *testing.T) {
		// Setup: Create test database
		testDB := NewIntegrationTestDatabase("test_integration")

		// Test collection creation
		userCollection := testDB.CreateTestCollection("users")
		productCollection := testDB.CreateTestCollection("products")

		assert.Equal(t, "test_users", userCollection)
		assert.Equal(t, "test_products", productCollection)

		// Test collection name generation
		orderCollection := testDB.GetCollectionName("orders")
		assert.Equal(t, "test_orders", orderCollection)

		// Test cleanup
		err := testDB.Cleanup()
		assert.NoError(t, err)
	})

	t.Run("Database Configuration", func(t *testing.T) {
		// Test different database configurations
		testDB := NewIntegrationTestDatabase("custom_test")
		testDB.ConnectionString = "mongodb://localhost:27018"
		testDB.CollectionPrefix = "custom_"
		testDB.CleanupAfter = false

		collectionName := testDB.CreateTestCollection("items")
		assert.Equal(t, "custom_items", collectionName)
	})
}

// TestIntegrationLifecycleManager demonstrates service lifecycle management
func TestIntegrationLifecycleManager(t *testing.T) {
	t.Run("Service Registration and Management", func(t *testing.T) {
		// Setup: Create lifecycle manager
		manager := NewTestLifecycleManager()

		// Create test services
		userService := NewTestEndorService().
			WithResource("users").
			WithDescription("User management service").
			Build()

		productService := NewTestEndorHybridService().
			WithResource("products").
			WithResourceDescription("Product catalog service").
			Build()

		// Register services
		manager.RegisterService("users", userService)
		manager.RegisterHybridService("products", productService)

		// Verify service retrieval
		retrievedUserService, exists := manager.GetService("users")
		assert.True(t, exists)
		assert.NotNil(t, retrievedUserService)

		retrievedProductService, exists := manager.GetHybridService("products")
		assert.True(t, exists)
		assert.NotNil(t, retrievedProductService)

		// Test non-existent service
		_, exists = manager.GetService("nonexistent")
		assert.False(t, exists)

		// Test repository registration
		userRepo := NewInMemoryRepository[TestUserPayload]()
		manager.RegisterRepository("user_repo", userRepo)

		retrievedRepo, exists := manager.GetRepository("user_repo")
		assert.True(t, exists)
		assert.NotNil(t, retrievedRepo)

		// Test service startup
		err := manager.StartAll()
		assert.NoError(t, err)

		// Test cleanup
		err = manager.StopAll()
		assert.NoError(t, err)
	})

	t.Run("Error Handling", func(t *testing.T) {
		// Setup: Create manager with failing service
		manager := NewTestLifecycleManager()

		// Create a service that will fail validation
		failingService := &MockEndorService{}
		failingService.On("Validate").Return(SimulateServiceError("validation"))

		manager.RegisterService("failing", failingService)

		// Test startup failure
		err := manager.StartAll()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate service failing")
	})
}

// TestIntegrationTestSuite demonstrates complete integration test suite usage
func TestIntegrationTestSuite(t *testing.T) {
	t.Run("Complete Integration Test", func(t *testing.T) {
		// Setup: Create integration test suite
		suite := NewIntegrationTestSuite("complete_test").
			WithTimeout(10 * time.Second)

		// Add setup tasks
		setupExecuted := false
		suite.AddSetupTask(func() error {
			setupExecuted = true
			return nil
		})

		// Add services to the suite
		userService := NewTestEndorService().
			WithResource("users").
			WithDescription("User management service").
			Build()

		suite.Manager.RegisterService("users", userService)

		// Add repository to the suite
		userRepo := NewInMemoryRepository[TestUserPayload]()
		suite.Manager.RegisterRepository("user_repo", userRepo)

		// Add teardown tasks
		teardownExecuted := false
		suite.AddTeardownTask(func() error {
			teardownExecuted = true
			return nil
		})

		// Test suite setup
		err := suite.Setup()
		assert.NoError(t, err)
		assert.True(t, setupExecuted, "Setup task should have been executed")

		// Verify services are available
		retrievedService, exists := suite.Manager.GetService("users")
		assert.True(t, exists)
		assert.NotNil(t, retrievedService)

		// Verify repositories are available
		retrievedRepo, exists := suite.Manager.GetRepository("user_repo")
		assert.True(t, exists)
		assert.NotNil(t, retrievedRepo)

		// Test business logic with the suite
		ctx := context.Background()
		testUser := TestUserPayload{
			Name:   "Suite Test User",
			Email:  "suite@example.com",
			Active: true,
		}

		// Use the repository from the suite
		if repo, ok := retrievedRepo.(*InMemoryRepository[TestUserPayload]); ok {
			id, err := repo.Create(ctx, testUser)
			assert.NoError(t, err)
			assert.NotEmpty(t, id)

			retrievedUser, err := repo.FindByID(ctx, id)
			assert.NoError(t, err)
			assert.Equal(t, testUser.Name, retrievedUser.Name)
		}

		// Test suite teardown
		err = suite.Teardown()
		assert.NoError(t, err)
		assert.True(t, teardownExecuted, "Teardown task should have been executed")
	})

	t.Run("Suite with Database", func(t *testing.T) {
		// Create suite with database configuration
		suite := NewIntegrationTestSuite("db_test")

		// Verify database configuration
		assert.NotNil(t, suite.Database)
		assert.Equal(t, "test_db_test", suite.Database.DatabaseName)
		assert.Contains(t, suite.Config.GetDocumentDBUri(), "test_db_test")

		// Test database collection creation through suite
		collectionName := suite.Database.CreateTestCollection("test_entities")
		assert.Equal(t, "test_test_entities", collectionName)

		// Setup and teardown
		err := suite.Setup()
		assert.NoError(t, err)

		err = suite.Teardown()
		assert.NoError(t, err)
	})

	t.Run("Suite Error Handling", func(t *testing.T) {
		// Create suite with failing setup task
		suite := NewIntegrationTestSuite("error_test")

		suite.AddSetupTask(func() error {
			return SimulateServiceError("setup")
		})

		// Test setup failure
		err := suite.Setup()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "setup task 0 failed")
	})

	t.Run("Suite with Cleanup Disabled", func(t *testing.T) {
		// Create suite with cleanup disabled
		suite := NewIntegrationTestSuite("no_cleanup_test").
			WithCleanupDisabled()

		assert.False(t, suite.Database.CleanupAfter)

		// Setup and teardown should still work
		err := suite.Setup()
		assert.NoError(t, err)

		err = suite.Teardown()
		assert.NoError(t, err)
	})
}

// TestIntegrationPerformance demonstrates performance testing in integration context
func TestIntegrationPerformance(t *testing.T) {
	t.Run("Repository Performance", func(t *testing.T) {
		// Setup: Create repository with multiple items
		repo := NewInMemoryRepository[TestUserPayload]()
		ctx := context.Background()

		// Test bulk operations performance
		elapsed := WithPerformanceAssertions(t, 1000*time.Millisecond, func() {
			// Create many items
			users := make([]TestUserPayload, 100)
			for i := range users {
				users[i] = TestUserPayload{
					Name:   "Performance User " + string(rune(i)),
					Email:  "perf" + string(rune(i)) + "@example.com",
					Active: true,
				}
			}

			// Measure creation time
			for _, user := range users {
				_, err := repo.Create(ctx, user)
				assert.NoError(t, err)
			}

			// Measure retrieval time
			retrievedUsers, err := repo.FindAll(ctx, bson.M{})
			assert.NoError(t, err)
			assert.Len(t, retrievedUsers, 100)
		})

		// Assert performance characteristics
		assert.Less(t, elapsed, 500*time.Millisecond, "Bulk operations should be fast")
	})

	t.Run("Service Startup Performance", func(t *testing.T) {
		// Test service startup time
		WithTimeout(t, 5*time.Second, func() {
			suite := NewIntegrationTestSuite("perf_test")

			// Add multiple services
			for i := 0; i < 10; i++ {
				serviceName := "service_" + string(rune(i))
				service := NewTestEndorService().
					WithResource(serviceName).
					Build()
				suite.Manager.RegisterService(serviceName, service)
			}

			// Measure startup time
			err := suite.Setup()
			assert.NoError(t, err)

			err = suite.Teardown()
			assert.NoError(t, err)
		})
	})
}
