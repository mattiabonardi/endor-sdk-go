package testutils

import (
	"testing"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/stretchr/testify/assert"
)

// TestExampleUsagePatterns demonstrates complete usage examples of the test utilities.
// This serves as both test validation and documentation.

// businessLogic represents a typical business function that uses framework interfaces
func businessLogic(service interfaces.EndorServiceInterface) (string, error) {
	if err := service.Validate(); err != nil {
		return "", err
	}
	return service.GetResource() + " is operational", nil
}

// TestExampleServiceMocking demonstrates basic service mocking patterns
func TestExampleServiceMocking(t *testing.T) {
	t.Run("Successful Service Operation", func(t *testing.T) {
		// Setup: Create and configure mock
		mockService := &MockEndorService{}
		mockService.On("Validate").Return(nil)
		mockService.On("GetResource").Return("users")

		// Execute: Test business logic
		result, err := businessLogic(mockService)

		// Assert: Verify results
		assert.NoError(t, err)
		assert.Equal(t, "users is operational", result)

		// Verify: Check mock expectations
		mockService.AssertExpectations(t)
	})

	t.Run("Service Validation Failure", func(t *testing.T) {
		// Setup: Configure mock to return error
		mockService := &MockEndorService{}
		validationError := SimulateServiceError("validation")
		mockService.On("Validate").Return(validationError)

		// Execute: Test error handling
		result, err := businessLogic(mockService)

		// Assert: Verify error handling
		assert.Error(t, err)
		assert.Empty(t, result)
		mockService.AssertExpectations(t)
	})
}

// configurationUsage demonstrates configuration-dependent logic
func configurationUsage(config interfaces.ConfigProviderInterface) string {
	if config.IsHybridResourcesEnabled() {
		return "hybrid:" + config.GetServerPort()
	}
	return "standard:" + config.GetServerPort()
}

// TestExampleConfigurationMocking demonstrates configuration testing patterns
func TestExampleConfigurationMocking(t *testing.T) {
	t.Run("Hybrid Resources Enabled", func(t *testing.T) {
		// Setup: Use builder pattern for configuration
		testConfig := NewTestConfigProvider().
			WithServerPort("8080").
			WithHybridResourcesEnabled(true).
			Build()

		// Execute and Assert
		result := configurationUsage(testConfig)
		assert.Equal(t, "hybrid:8080", result)
	})

	t.Run("Hybrid Resources Disabled", func(t *testing.T) {
		// Setup: Different configuration
		testConfig := NewTestConfigProvider().
			WithServerPort("9090").
			WithHybridResourcesEnabled(false).
			Build()

		// Execute and Assert
		result := configurationUsage(testConfig)
		assert.Equal(t, "standard:9090", result)
	})
}

// contextHandler demonstrates context-based request processing
func contextHandler(ctx interfaces.EndorContextInterface[TestUserPayload]) (string, error) {
	serviceId := ctx.GetMicroServiceId()
	payload := ctx.GetPayload()

	if !payload.Active {
		return "", SimulateServiceError("unauthorized")
	}

	return serviceId + " processed " + payload.Name, nil
}

// TestExampleContextMocking demonstrates context testing patterns
func TestExampleContextMocking(t *testing.T) {
	t.Run("Active User Processing", func(t *testing.T) {
		// Setup: Create test context with active user
		testPayload := TestUserPayload{
			Name:   "Alice Johnson",
			Email:  "alice@example.com",
			Active: true,
		}

		mockContext := NewTestEndorContext[TestUserPayload]().
			WithMicroServiceId("user-processor").
			WithPayload(testPayload).
			Build()

		// Execute
		result, err := contextHandler(mockContext)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "user-processor processed Alice Johnson", result)
	})

	t.Run("Inactive User Rejection", func(t *testing.T) {
		// Setup: Create test context with inactive user
		testPayload := TestUserPayload{
			Name:   "Bob Wilson",
			Email:  "bob@example.com",
			Active: false, // Inactive user should be rejected
		}

		mockContext := NewTestEndorContext[TestUserPayload]().
			WithMicroServiceId("user-processor").
			WithPayload(testPayload).
			Build()

		// Execute
		result, err := contextHandler(mockContext)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "Access denied")
	})
}

// TestExamplePerformanceTesting demonstrates performance testing patterns
func TestExamplePerformanceTesting(t *testing.T) {
	t.Run("Service Response Time", func(t *testing.T) {
		// Setup: Create performance mock with controlled latency
		perfService := NewPerformanceMockService(25 * time.Millisecond)
		perfService.On("GetResource").Return("performance-test-service")
		perfService.On("Validate").Return(nil)

		// Execute with performance measurement
		elapsed := WithPerformanceAssertions(t, 100*time.Millisecond, func() {
			// Simulate service usage
			resource := perfService.GetResource()
			assert.Equal(t, "performance-test-service", resource)

			err := perfService.Validate()
			assert.NoError(t, err)
		})

		// Assert: Verify performance characteristics
		assert.GreaterOrEqual(t, elapsed, 25*time.Millisecond, "Should have simulated latency")
		assert.LessOrEqual(t, elapsed, 100*time.Millisecond, "Should complete within timeout")
	})

	t.Run("Timeout Testing", func(t *testing.T) {
		// Test that operations complete within reasonable time
		WithTimeout(t, 100*time.Millisecond, func() {
			// Simulate network operation
			SimulateNetworkLatency(10*time.Millisecond, 20*time.Millisecond)

			// Test completes successfully within timeout
			assert.True(t, true)
		})
	})
}

// TestExampleEnvironmentSetup demonstrates integration testing setup
func TestExampleEnvironmentSetup(t *testing.T) {
	t.Run("Complete Environment Test", func(t *testing.T) {
		// Setup: Create complete test environment
		testEnv := SetupTestEnvironment("example-microservice")
		defer CleanupTestEnvironment(testEnv)

		// Add services to environment
		userService := SetupTestService("users", "User management service")
		productService := SetupTestHybridService("products", "Product catalog service")

		testEnv.AddService("users", userService)
		testEnv.AddHybridService("products", productService)

		// Test environment configuration
		assert.Equal(t, "example-microservice", testEnv.MicroServiceID)
		assert.NotNil(t, testEnv.ConfigProvider)
		assert.Len(t, testEnv.Services, 1)
		assert.Len(t, testEnv.HybridServices, 1)

		// Test individual services
		AssertServiceInterface(t, testEnv.Services["users"])
		AssertHybridServiceInterface(t, testEnv.HybridServices["products"])
		AssertConfigProvider(t, testEnv.ConfigProvider)
	})
}

// TestExampleTestDataUsage demonstrates how to use test fixtures effectively
func TestExampleTestDataUsage(t *testing.T) {
	t.Run("Using Test Fixtures", func(t *testing.T) {
		// Get predefined test data
		users := GetTestUsers()
		products := GetTestProducts()
		scenarios := GetCRUDTestScenarios()

		// Validate fixture data quality
		assert.Len(t, users, 3, "Should have multiple test users")
		assert.Len(t, products, 3, "Should have multiple test products")
		assert.NotEmpty(t, scenarios, "Should have test scenarios")

		// Use fixtures in tests
		for _, user := range users {
			assert.NotEmpty(t, user.Name, "User should have name")
			assert.NotEmpty(t, user.Email, "User should have email")
		}

		for _, product := range products {
			assert.Greater(t, product.Price, 0.0, "Product should have positive price")
			assert.NotEmpty(t, product.Name, "Product should have name")
		}
	})

	t.Run("Schema Validation", func(t *testing.T) {
		// Test schema fixtures
		userSchema := GetTestUserSchema()
		productSchema := GetTestProductSchema()

		assert.Equal(t, interfaces.ObjectType, userSchema.Type)
		assert.Contains(t, userSchema.Required, "name")
		assert.Contains(t, userSchema.Required, "email")

		assert.Equal(t, interfaces.ObjectType, productSchema.Type)
		assert.Contains(t, productSchema.Required, "name")
		assert.Contains(t, productSchema.Required, "price")
	})
}

// TestExampleErrorSimulation demonstrates error testing patterns
func TestExampleErrorSimulation(t *testing.T) {
	errorTypes := []struct {
		errorType   string
		expectedMsg string
	}{
		{"validation", "Validation failed"},
		{"not_found", "not found"},
		{"unauthorized", "Access denied"},
		{"internal", "Internal server error"},
	}

	for _, tt := range errorTypes {
		t.Run("Error_"+tt.errorType, func(t *testing.T) {
			// Generate specific error type
			err := SimulateServiceError(tt.errorType)

			// Verify error characteristics
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedMsg)
		})
	}
}

// TestExampleBuilderPatterns demonstrates builder usage patterns
func TestExampleBuilderPatterns(t *testing.T) {
	t.Run("Service Builder", func(t *testing.T) {
		// Build service with specific configuration
		testService := NewTestEndorService().
			WithResource("advanced-service").
			WithDescription("Advanced service description").
			WithVersion("2.0").
			WithPriority(5).
			Build()

		// Validate builder results
		assert.Equal(t, "advanced-service", testService.GetResource())
		assert.Equal(t, "Advanced service description", testService.GetDescription())
		assert.Equal(t, "2.0", testService.GetVersion())
		assert.NotNil(t, testService.GetPriority())
		assert.Equal(t, 5, *testService.GetPriority())

		// Note: We don't call AssertExpectations here because builders
		// set up the mocks but don't require specific usage patterns
	})

	t.Run("Hybrid Service Builder", func(t *testing.T) {
		// Build hybrid service with categories
		testHybridService := NewTestEndorHybridService().
			WithResource("advanced-hybrid").
			WithResourceDescription("Advanced hybrid service").
			WithPriority(3).
			Build()

		// Validate builder results
		assert.Equal(t, "advanced-hybrid", testHybridService.GetResource())
		assert.Equal(t, "Advanced hybrid service", testHybridService.GetResourceDescription())
		assert.NotNil(t, testHybridService.GetPriority())
		assert.Equal(t, 3, *testHybridService.GetPriority())

		// Note: We don't call AssertExpectations here because builders
		// set up the mocks but don't require specific usage patterns
	})
}
