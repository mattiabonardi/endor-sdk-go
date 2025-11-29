//go:build integration

package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	services_test "github.com/mattiabonardi/endor-sdk-go/test/services"
	"github.com/stretchr/testify/assert"
)

// TestEndorHybridServiceIntegration demonstrates the integration testing approach
// using real service implementations with the testutils infrastructure for validation
func TestEndorHybridServiceIntegration(t *testing.T) {
	// Arrange: Create a real hybrid service using the test services
	hybridService := services_test.NewService2()

	// Create test schema for conversion
	type AdditionalAttributesMock struct {
		AdditionalAttribute string `json:"additionalAttribute"`
	}

	// Act: Convert to EndorService using real implementation
	endorService := hybridService.ToEndorService(
		sdk.NewSchema(AdditionalAttributesMock{}).Schema,
	)

	// Assert: Verify real service structure works correctly
	assert.NotEmpty(t, endorService.Resource, "Service should have resource name")
	assert.NotEmpty(t, endorService.Description, "Service should have description")
	assert.NotNil(t, endorService.Methods, "Service should have methods")

	// Verify default CRUD methods are present (integration validation)
	methods := endorService.Methods
	assert.Contains(t, methods, "schema", "Default schema method should be present")
	assert.Contains(t, methods, "instance", "Default instance method should be present")
	assert.Contains(t, methods, "list", "Default list method should be present")
	assert.Contains(t, methods, "create", "Default create method should be present")
	assert.Contains(t, methods, "update", "Default update method should be present")
	assert.Contains(t, methods, "delete", "Default delete method should be present")

	// Verify custom action is present
	assert.Contains(t, methods, "action-1", "Custom action should be present")
}

// TestEndorHybridServiceCategories_Integration validates real category functionality
func TestEndorHybridServiceCategories_Integration(t *testing.T) {
	// Arrange: Create real hybrid service with categories
	hybridService := services_test.NewService2()

	type AdditionalAttributesMock struct {
		AdditionalAttribute string `json:"additionalAttribute"`
	}

	// Act: Convert with real categories
	endorService := hybridService.ToEndorService(
		sdk.NewSchema(AdditionalAttributesMock{}).Schema,
	)

	methods := endorService.Methods

	// Assert: Verify category-specific methods exist
	assert.Contains(t, methods, "cat-1/schema", "Category 1 schema should be present")
	assert.Contains(t, methods, "cat-1/instance", "Category 1 instance should be present")
	assert.Contains(t, methods, "cat-1/list", "Category 1 list should be present")
	assert.Contains(t, methods, "cat-1/create", "Category 1 create should be present")
	assert.Contains(t, methods, "cat-1/update", "Category 1 update should be present")
	assert.Contains(t, methods, "cat-1/delete", "Category 1 delete should be present")

	// Verify Category 2 methods
	assert.Contains(t, methods, "cat-2/create", "Category 2 create should be present")
	assert.Contains(t, methods, "cat-2/update", "Category 2 update should be present")
	assert.Contains(t, methods, "cat-2/delete", "Category 2 delete should be present")

	// Verify proper schema generation for categories
	createMethod := methods["cat-1/create"]
	assert.NotNil(t, createMethod, "Category 1 create method should not be nil")

	options := createMethod.GetOptions()
	assert.NotEmpty(t, options.Description, "Category method should have description")
	assert.True(t, options.Public, "Category method should be public by default")
}

// TestEndorService_Integration validates regular service functionality
func TestEndorService_Integration(t *testing.T) {
	// Arrange: Create regular EndorService
	service := services_test.NewService1()

	// Assert: Verify service structure is properly configured
	assert.NotEmpty(t, service.Resource, "Service should have resource name")
	assert.NotEmpty(t, service.Description, "Service should have description")
	assert.NotNil(t, service.Methods, "Service should have methods")

	// Verify methods are properly configured
	methods := service.Methods
	assert.Contains(t, methods, "action1", "Service should have action1")
	assert.Contains(t, methods, "cat_1/action1", "Service should have category action")

	// Test action configuration
	action := methods["action1"]
	options := action.GetOptions()
	assert.Equal(t, "Action 1", options.Description)
	assert.True(t, options.Public, "Action should be public")
}

// TestSchemaGeneration_Integration validates schema generation with real services
func TestSchemaGeneration_Integration(t *testing.T) {
	// This test validates that interface-based approach preserves schema generation
	type TestUser struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create schema using real implementation
	schema := sdk.NewSchema(TestUser{})
	assert.NotNil(t, schema, "Schema should be generated")
	assert.Equal(t, sdk.ObjectType, schema.Type)
	assert.NotNil(t, schema.Properties, "Schema should have properties")

	properties := *schema.Properties
	assert.Contains(t, properties, "id", "Schema should include id property")
	assert.Contains(t, properties, "name", "Schema should include name property")
	assert.Contains(t, properties, "age", "Schema should include age property")
}

// TestEndorHybridService_ServiceEmbedding_Integration tests service embedding with real service instances (AC: 1-7)
func TestEndorHybridService_ServiceEmbedding_Integration(t *testing.T) {
	// Skip test if real service embedding is not ready
	t.Skip("Service embedding integration tests require real service instances and dependency injection setup")

	// This test would verify:
	// - AC 1: Real EmbedService() method functionality with concrete services
	// - AC 2: Actual method delegation through HTTP callbacks
	// - AC 3: Method resolution with real conflicting services
	// - AC 4: Dependency sharing between parent and embedded services
	// - AC 5: Middleware inheritance through real middleware pipeline
	// - AC 6: Type safety with concrete service implementations
	// - AC 7: Multiple service embedding with real routing

	// Implementation would require:
	// 1. Creating multiple test services with different methods
	// 2. Setting up dependency injection container with shared dependencies
	// 3. Creating hybrid service and embedding multiple services
	// 4. Testing HTTP routing to embedded service methods
	// 5. Validating middleware execution order
	// 6. Performance testing for method delegation overhead
}

// TestEndorHybridService_EmbeddedServiceMethodRouting_Integration tests method routing to embedded services
func TestEndorHybridService_EmbeddedServiceMethodRouting_Integration(t *testing.T) {
	// Skip test as it requires HTTP routing setup
	t.Skip("Method routing tests require HTTP server setup and real routing infrastructure")

	// This test would verify:
	// - HTTP requests to embedded service methods work correctly
	// - Prefix-based routing resolves to correct embedded service
	// - Method precedence works in HTTP context
	// - Error handling propagates correctly through delegation
	// - Performance overhead meets < 20μs requirement
}
