// Package testutils provides comprehensive examples and usage patterns for testing
// endor-sdk-go framework components. This file contains example code and patterns
// that developers can copy and adapt for their own tests.

package testutils

// Example Usage Patterns
//
// 1. Basic Service Testing:
//
//	func TestMyServiceLogic(t *testing.T) {
//		mockService := &testutils.MockEndorService{}
//		mockService.On("GetResource").Return("users")
//		mockService.On("Validate").Return(nil)
//
//		// Test your logic that uses the service
//		result := myFunction(mockService)
//		assert.Equal(t, "expected", result)
//
//		mockService.AssertExpectations(t)
//	}
//
// 2. Hybrid Service Testing:
//
//	func TestHybridServiceConversion(t *testing.T) {
//		mockHybridService := testutils.NewTestEndorHybridService().
//			WithResource("products").
//			WithResourceDescription("Product catalog").
//			Build()
//
//		// Test service conversion
//		schema := interfaces.Schema{Type: interfaces.ObjectType}
//		result := mockHybridService.ToEndorService(schema)
//		assert.NotNil(t, result)
//	}
//
// 3. Configuration Testing:
//
//	func TestConfigDependent(t *testing.T) {
//		testConfig := testutils.NewTestConfigProvider().
//			WithServerPort("9999").
//			WithHybridResourcesEnabled(false).
//			Build()
//
//		// Use config in your tests
//		service := NewMyService(testConfig)
//		assert.Equal(t, "http://localhost:9999", service.GetURL())
//	}
//
// 4. Context Testing:
//
//	func TestRequestHandler(t *testing.T) {
//		testPayload := testutils.TestUserPayload{
//			Name: "John", Email: "john@example.com", Active: true,
//		}
//
//		mockContext := testutils.NewTestEndorContext[testutils.TestUserPayload]().
//			WithMicroServiceId("user-service").
//			WithPayload(testPayload).
//			Build()
//
//		// Test your handler
//		result, err := myHandler(mockContext)
//		assert.NoError(t, err)
//		assert.Equal(t, "expected", result)
//	}
//
// 5. Error Scenario Testing:
//
//	func TestErrorHandling(t *testing.T) {
//		mockService := &testutils.MockEndorService{}
//		validationError := testutils.SimulateServiceError("validation")
//		mockService.On("Validate").Return(validationError)
//
//		// Test error handling
//		err := mockService.Validate()
//		assert.Error(t, err)
//		assert.Contains(t, err.Error(), "Validation failed")
//	}
//
// 6. Performance Testing:
//
//	func TestPerformance(t *testing.T) {
//		perfService := testutils.NewPerformanceMockService(50 * time.Millisecond)
//		perfService.On("GetResource").Return("slow-service")
//
//		elapsed := testutils.WithPerformanceAssertions(t, 100*time.Millisecond, func() {
//			result := perfService.GetResource()
//			assert.Equal(t, "slow-service", result)
//		})
//
//		assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)
//	}
//
// 7. Complete Environment Testing:
//
//	func TestIntegration(t *testing.T) {
//		testEnv := testutils.SetupTestEnvironment("test-service")
//		defer testutils.CleanupTestEnvironment(testEnv)
//
//		// Add services to environment
//		userService := testutils.SetupTestService("users", "User management")
//		testEnv.AddService("users", userService)
//
//		// Test complete environment
//		assert.Equal(t, "test-service", testEnv.MicroServiceID)
//		assert.NotNil(t, testEnv.Services["users"])
//	}
//
// 8. Test Data Usage:
//
//	func TestWithFixtures(t *testing.T) {
//		users := testutils.GetTestUsers()
//		scenarios := testutils.GetCRUDTestScenarios()
//
//		for _, scenario := range scenarios {
//			t.Run(scenario.Name, func(t *testing.T) {
//				// Use scenario data for comprehensive testing
//			})
//		}
//	}
//
// Common Testing Patterns:
//
// • Use builders (NewTestEndorService(), NewTestConfigProvider()) for flexible test setup
// • Use SetupTestEnvironment() for integration testing scenarios
// • Use performance simulation for testing under realistic conditions
// • Use error simulation to test error handling paths
// • Use test fixtures and scenarios for comprehensive coverage
// • Always call AssertExpectations() on mocks to verify behavior
// • Use assertion helpers (AssertServiceInterface()) for interface validation
//
// Best Practices:
//
// • Configure mock behavior before using mocks in tests
// • Use specific error simulation instead of generic errors
// • Clean up test environments with defer CleanupTestEnvironment()
// • Use performance assertions for operations with timing requirements
// • Test both success and failure scenarios
// • Verify all mock expectations to catch missing or unexpected calls
