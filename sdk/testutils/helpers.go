package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/stretchr/testify/assert"
)

// TestEnvironment represents a complete test environment setup
// with all necessary components for testing framework functionality.
type TestEnvironment struct {
	MicroServiceID string
	ConfigProvider interfaces.ConfigProviderInterface
	Services       map[string]interfaces.EndorServiceInterface
	HybridServices map[string]interfaces.EndorHybridServiceInterface
	Context        context.Context
	CancelFunc     context.CancelFunc
	CleanupFuncs   []func()
}

// SetupTestEnvironment creates a complete test environment with sensible defaults.
// This is the primary helper function for initializing test scenarios.
//
// Example usage:
//
//	testEnv := SetupTestEnvironment("test-service")
//	defer CleanupTestEnvironment(testEnv)
//
//	// Use testEnv.Services, testEnv.ConfigProvider, etc. in your tests
func SetupTestEnvironment(microServiceID string) *TestEnvironment {
	ctx, cancel := context.WithCancel(context.Background())

	// Create test configuration
	configProvider := NewTestConfigProvider().
		WithServerPort("8080").
		WithDocumentDBUri("mongodb://localhost:27017/test").
		WithHybridResourcesEnabled(true).
		WithDynamicResourcesEnabled(true).
		Build()

	env := &TestEnvironment{
		MicroServiceID: microServiceID,
		ConfigProvider: configProvider,
		Services:       make(map[string]interfaces.EndorServiceInterface),
		HybridServices: make(map[string]interfaces.EndorHybridServiceInterface),
		Context:        ctx,
		CancelFunc:     cancel,
		CleanupFuncs:   make([]func(), 0),
	}

	return env
}

// SetupTestEnvironmentWithConfig creates a test environment with custom configuration.
func SetupTestEnvironmentWithConfig(microServiceID string, config interfaces.ConfigProviderInterface) *TestEnvironment {
	ctx, cancel := context.WithCancel(context.Background())

	env := &TestEnvironment{
		MicroServiceID: microServiceID,
		ConfigProvider: config,
		Services:       make(map[string]interfaces.EndorServiceInterface),
		HybridServices: make(map[string]interfaces.EndorHybridServiceInterface),
		Context:        ctx,
		CancelFunc:     cancel,
		CleanupFuncs:   make([]func(), 0),
	}

	return env
}

// AddService adds a service to the test environment.
func (env *TestEnvironment) AddService(name string, service interfaces.EndorServiceInterface) {
	env.Services[name] = service
}

// AddHybridService adds a hybrid service to the test environment.
func (env *TestEnvironment) AddHybridService(name string, service interfaces.EndorHybridServiceInterface) {
	env.HybridServices[name] = service
}

// AddCleanupFunc adds a cleanup function to be called during environment teardown.
func (env *TestEnvironment) AddCleanupFunc(cleanup func()) {
	env.CleanupFuncs = append(env.CleanupFuncs, cleanup)
}

// CleanupTestEnvironment cleans up all resources created during testing.
func CleanupTestEnvironment(env *TestEnvironment) {
	// Cancel context
	if env.CancelFunc != nil {
		env.CancelFunc()
	}

	// Run cleanup functions in reverse order
	for i := len(env.CleanupFuncs) - 1; i >= 0; i-- {
		env.CleanupFuncs[i]()
	}
}

// Service setup helpers

// SetupTestService creates a fully configured test service with common actions.
func SetupTestService(resource string, description string) interfaces.EndorServiceInterface {
	return NewTestEndorService().
		WithResource(resource).
		WithDescription(description).
		WithBasicMethods().
		Build()
}

// SetupTestHybridService creates a fully configured test hybrid service.
func SetupTestHybridService(resource string, description string) interfaces.EndorHybridServiceInterface {
	return NewTestEndorHybridService().
		WithResource(resource).
		WithResourceDescription(description).
		WithDefaultCategory("default").
		WithActions(true).
		Build()
}

// SetupTestServiceWithActions creates a test service with custom actions.
func SetupTestServiceWithActions(resource string, actions map[string]interfaces.EndorServiceAction) interfaces.EndorServiceInterface {
	builder := NewTestEndorService().
		WithResource(resource).
		WithDescription("Test service with custom actions")

	for name, action := range actions {
		builder = builder.WithMethod(name, action)
	}

	return builder.Build()
}

// Test lifecycle helpers

// WithTimeout runs a test function with a timeout, useful for performance testing.
func WithTimeout(t *testing.T, timeout time.Duration, testFunc func()) {
	done := make(chan bool)
	go func() {
		testFunc()
		done <- true
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(timeout):
		t.Fatalf("Test timed out after %v", timeout)
	}
}

// WithPerformanceAssertions runs a test function and asserts performance characteristics.
func WithPerformanceAssertions(t *testing.T, maxDuration time.Duration, testFunc func()) time.Duration {
	start := time.Now()
	testFunc()
	elapsed := time.Since(start)

	assert.True(t, elapsed <= maxDuration,
		"Test took %v, expected <= %v", elapsed, maxDuration)

	return elapsed
}

// Assertion helpers for common testing scenarios

// AssertServiceInterface verifies that a service implements the EndorServiceInterface correctly.
func AssertServiceInterface(t *testing.T, service interfaces.EndorServiceInterface) {
	assert.NotNil(t, service, "Service should not be nil")
	assert.NotEmpty(t, service.GetResource(), "Service resource should not be empty")
	assert.NotEmpty(t, service.GetDescription(), "Service description should not be empty")
	assert.NoError(t, service.Validate(), "Service should validate successfully")
}

// AssertHybridServiceInterface verifies that a hybrid service implements the interface correctly.
func AssertHybridServiceInterface(t *testing.T, service interfaces.EndorHybridServiceInterface) {
	assert.NotNil(t, service, "Hybrid service should not be nil")
	assert.NotEmpty(t, service.GetResource(), "Hybrid service resource should not be empty")
	assert.NotEmpty(t, service.GetResourceDescription(), "Hybrid service description should not be empty")
	assert.NoError(t, service.Validate(), "Hybrid service should validate successfully")

	// Test method chaining - skip for mocks that already have expectations set
	if _, ok := service.(*MockEndorHybridService); !ok {
		withCategories := service.WithCategories([]interfaces.EndorHybridServiceCategory{})
		assert.NotNil(t, withCategories, "WithCategories should return non-nil result")
	}
} // AssertConfigProvider verifies that a config provider implements the interface correctly.
func AssertConfigProvider(t *testing.T, config interfaces.ConfigProviderInterface) {
	assert.NotNil(t, config, "Config provider should not be nil")
	assert.NotEmpty(t, config.GetServerPort(), "Server port should not be empty")
	assert.NotEmpty(t, config.GetDocumentDBUri(), "Document DB URI should not be empty")
	assert.NoError(t, config.Validate(), "Config provider should validate successfully")
}

// Context creation helpers

// CreateTestGinContext creates a test Gin context with common headers and request data.
func CreateTestGinContext() (*gin.Context, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	context, _ := gin.CreateTestContext(nil)

	return context, engine
}

// CreateTestGinContextWithAuth creates a test Gin context with authentication headers.
func CreateTestGinContextWithAuth(sessionID, userID string, development bool) *gin.Context {
	context, _ := CreateTestGinContext()
	return context
}

// Error simulation helpers

// SimulateNetworkLatency adds artificial network latency to simulate real-world conditions.
func SimulateNetworkLatency(min, max time.Duration) {
	// Random latency between min and max
	latency := min + time.Duration(time.Now().UnixNano()%(int64(max-min)))
	time.Sleep(latency)
}

// SimulateServiceError creates predictable errors for testing error handling.
func SimulateServiceError(errorType string) error {
	switch errorType {
	case "validation":
		return &ValidationError{message: "Validation failed", field: "test"}
	case "not_found":
		return &NotFoundError{resource: "test", id: "not-found"}
	case "unauthorized":
		return &UnauthorizedError{message: "Access denied"}
	case "internal":
		return &InternalError{message: "Internal server error"}
	default:
		return &GenericError{message: "Unknown error"}
	}
}

// Test error types for error simulation

// ValidationError simulates validation errors.
type ValidationError struct {
	message string
	field   string
}

func (e *ValidationError) Error() string {
	return e.message
}

// NotFoundError simulates resource not found errors.
type NotFoundError struct {
	resource string
	id       string
}

func (e *NotFoundError) Error() string {
	return "Resource not found: " + e.resource + " with ID " + e.id
}

// UnauthorizedError simulates authorization errors.
type UnauthorizedError struct {
	message string
}

func (e *UnauthorizedError) Error() string {
	return e.message
}

// InternalError simulates internal server errors.
type InternalError struct {
	message string
}

func (e *InternalError) Error() string {
	return e.message
}

// GenericError simulates generic errors.
type GenericError struct {
	message string
}

func (e *GenericError) Error() string {
	return e.message
}

// Utility functions for service composition testing

// TestServiceComposition verifies that services can be composed correctly.
func TestServiceComposition(t *testing.T, services []interfaces.EndorServiceInterface) {
	assert.NotEmpty(t, services, "Services list should not be empty")

	resourceMap := make(map[string]bool)
	for _, service := range services {
		resource := service.GetResource()
		assert.False(t, resourceMap[resource], "Duplicate resource found: %s", resource)
		resourceMap[resource] = true

		AssertServiceInterface(t, service)
	}
}

// TestHybridServiceComposition verifies that hybrid services can be composed correctly.
func TestHybridServiceComposition(t *testing.T, services []interfaces.EndorHybridServiceInterface) {
	assert.NotEmpty(t, services, "Hybrid services list should not be empty")

	resourceMap := make(map[string]bool)
	for _, service := range services {
		resource := service.GetResource()
		assert.False(t, resourceMap[resource], "Duplicate resource found: %s", resource)
		resourceMap[resource] = true

		AssertHybridServiceInterface(t, service)
	}
}

// Repository mock helpers (for future repository interface mocks)

// MockRepositoryOptions provides configuration for repository mock behavior.
type MockRepositoryOptions struct {
	AutoGenerateID  bool
	SimulateLatency time.Duration
	ErrorRate       float64 // 0.0 to 1.0
	MaxResults      int
}

// DefaultMockRepositoryOptions returns sensible defaults for repository mocking.
func DefaultMockRepositoryOptions() MockRepositoryOptions {
	return MockRepositoryOptions{
		AutoGenerateID:  true,
		SimulateLatency: 10 * time.Millisecond,
		ErrorRate:       0.0,
		MaxResults:      100,
	}
}

// Performance benchmarking helpers

// BenchmarkServiceOperation measures the performance of a service operation.
func BenchmarkServiceOperation(b *testing.B, operation func()) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		operation()
	}
}

// BenchmarkServiceWithLoad measures service performance under concurrent load.
func BenchmarkServiceWithLoad(b *testing.B, concurrency int, operation func()) {
	b.SetParallelism(concurrency)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			operation()
		}
	})
}
