// Package testutils provides comprehensive testing utilities and mock implementations for the endor-sdk-go framework.
// This package includes ready-to-use mock implementations for all framework interfaces, fluent API builders,
// test data fixtures, helper functions, and complete integration testing infrastructure, enabling
// developers to quickly write both unit and integration tests without external dependencies.
//
// The package follows the framework's established patterns for interface-based dependency injection
// and testing, supporting both behavior verification (using testify/mock) and state testing.
//
// Key Components:
//
// - Mock implementations: MockEndorService, MockEndorHybridService, MockConfigProvider, etc.
// - Fluent API builders: TestEndorServiceBuilder, TestConfigProviderBuilder with method chaining
// - Test data fixtures: Realistic test data sets and generators for different service types
// - Helper functions: Environment setup, assertion helpers, error simulation utilities
// - Integration testing: InMemoryRepository, IntegrationTestSuite, TestLifecycleManager
// - Performance simulation: Mocks that simulate realistic latency for integration testing
//
// Example usage:
//
//	// Basic mock usage with testify/mock
//	mockService := &testutils.MockEndorService{}
//	mockService.On("GetResource").Return("users")
//	mockService.On("GetDescription").Return("Mock user service")
//	mockService.On("Validate").Return(nil)
//
//	// Test data builder usage
//	testService := testutils.NewTestEndorService().
//		WithResource("products").
//		WithDescription("Product management").
//		Build()
//
//	// Integration testing with in-memory repositories
//	repo := testutils.NewInMemoryRepository[TestUserPayload]()
//	id, err := repo.Create(ctx, testUser)
//
//	// Complete integration test suite
//	suite := testutils.NewIntegrationTestSuite("my_test")
//	suite.Manager.RegisterService("users", userService)
//	err := suite.Setup()
//	defer suite.Teardown()
//
//	// Helper function usage
//	testEnv := testutils.SetupTestEnvironment("test-service")
//	defer testutils.CleanupTestEnvironment(testEnv)
//
// All mock implementations satisfy their respective interface contracts and include
// interface compliance validation tests to ensure they work correctly with the framework.
package testutils
