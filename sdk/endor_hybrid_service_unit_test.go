//go:build unit

package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/mattiabonardi/endor-sdk-go/sdk/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestEndorHybridServiceInterface_Unit tests the hybrid service using interface mocks
// This demonstrates proper interface usage patterns and enables fast unit testing
func TestEndorHybridServiceInterface_Unit(t *testing.T) {
	// Arrange: Create mocks using testutils
	mockHybridService := &testutils.MockEndorHybridService{}
	mockEndorService := &testutils.MockEndorService{}

	// Configure mock expectations - only for methods that will be called
	mockHybridService.On("ToEndorService", mock.Anything).Return(mockEndorService)

	// Configure mock expectations for generated EndorService
	mockEndorService.On("GetMethods").Return(map[string]interfaces.EndorServiceAction{
		"schema":   testutils.NewTestAction("Get schema", true),
		"instance": testutils.NewTestAction("Get instance", true),
		"list":     testutils.NewTestAction("List resources", true),
		"create":   testutils.NewTestAction("Create resource", true),
		"update":   testutils.NewTestAction("Update resource", true),
		"delete":   testutils.NewTestAction("Delete resource", true),
		"action-1": testutils.NewTestAction("Custom action", true),
	})

	// Act: Convert hybrid service to endor service using interface
	endorService := mockHybridService.ToEndorService(interfaces.Schema{})
	methods := endorService.GetMethods()

	// Assert: Verify default CRUD methods are present
	assert.Contains(t, methods, "schema", "Default schema method should be present")
	assert.Contains(t, methods, "instance", "Default instance method should be present")
	assert.Contains(t, methods, "list", "Default list method should be present")
	assert.Contains(t, methods, "create", "Default create method should be present")
	assert.Contains(t, methods, "update", "Default update method should be present")
	assert.Contains(t, methods, "delete", "Default delete method should be present")

	// Assert: Verify custom actions are present
	assert.Contains(t, methods, "action-1", "Custom action should be present")

	// Verify all expectations were met
	mockHybridService.AssertExpectations(t)
	mockEndorService.AssertExpectations(t)
}

// TestEndorHybridServiceWithCategories_Unit tests category-based service specialization using mocks
func TestEndorHybridServiceWithCategories_Unit(t *testing.T) {
	// Arrange: Create hybrid service with category support
	mockHybridService := &testutils.MockEndorHybridService{}
	mockEndorService := &testutils.MockEndorService{}

	categories := []interfaces.EndorHybridServiceCategory{
		testutils.NewTestCategory("cat-1"),
		testutils.NewTestCategory("cat-2"),
	}

	// Configure mock to return itself when categories are added (fluent API)
	mockHybridService.On("WithCategories", categories).Return(mockHybridService)
	mockHybridService.On("ToEndorService", mock.Anything).Return(mockEndorService)

	// Configure mock service to include category-specific methods
	mockEndorService.On("GetMethods").Return(map[string]interfaces.EndorServiceAction{
		// Default methods
		"schema":   testutils.NewTestAction("Get schema", true),
		"instance": testutils.NewTestAction("Get instance", true),
		"list":     testutils.NewTestAction("List resources", true),
		"create":   testutils.NewTestAction("Create resource", true),
		"update":   testutils.NewTestAction("Update resource", true),
		"delete":   testutils.NewTestAction("Delete resource", true),

		// Category-specific methods
		"cat-1/schema":   testutils.NewTestAction("Cat1 schema", true),
		"cat-1/instance": testutils.NewTestAction("Cat1 instance", true),
		"cat-1/list":     testutils.NewTestAction("Cat1 list", true),
		"cat-1/create":   testutils.NewTestAction("Cat1 create", true),
		"cat-1/update":   testutils.NewTestAction("Cat1 update", true),
		"cat-1/delete":   testutils.NewTestAction("Cat1 delete", true),

		"cat-2/create": testutils.NewTestAction("Cat2 create", true),
		"cat-2/update": testutils.NewTestAction("Cat2 update", true),
		"cat-2/delete": testutils.NewTestAction("Cat2 delete", true),
	})

	// Act: Configure hybrid service with categories
	serviceWithCategories := mockHybridService.WithCategories(categories)
	endorService := serviceWithCategories.ToEndorService(interfaces.Schema{})
	methods := endorService.GetMethods()

	// Assert: Verify category-specific methods are present
	assert.Contains(t, methods, "cat-1/schema", "Category 1 schema method should be present")
	assert.Contains(t, methods, "cat-1/instance", "Category 1 instance method should be present")
	assert.Contains(t, methods, "cat-1/list", "Category 1 list method should be present")
	assert.Contains(t, methods, "cat-1/create", "Category 1 create method should be present")
	assert.Contains(t, methods, "cat-1/update", "Category 1 update method should be present")
	assert.Contains(t, methods, "cat-1/delete", "Category 1 delete method should be present")

	assert.Contains(t, methods, "cat-2/create", "Category 2 create method should be present")
	assert.Contains(t, methods, "cat-2/update", "Category 2 update method should be present")
	assert.Contains(t, methods, "cat-2/delete", "Category 2 delete method should be present")

	// Verify all expectations were met
	mockHybridService.AssertExpectations(t)
	mockEndorService.AssertExpectations(t)
}

// TestServiceInterfaceCompliance_Unit validates that mocks properly implement interfaces
func TestServiceInterfaceCompliance_Unit(t *testing.T) {
	// This test ensures our mocks properly implement the required interfaces
	// and demonstrates the testing pattern for interface compliance

	// Test EndorServiceInterface compliance
	var _ interfaces.EndorServiceInterface = (*testutils.MockEndorService)(nil)

	// Test EndorHybridServiceInterface compliance
	var _ interfaces.EndorHybridServiceInterface = (*testutils.MockEndorHybridService)(nil)

	// Create instances to verify they can be used as interfaces
	mockEndorService := &testutils.MockEndorService{}
	mockHybridService := &testutils.MockEndorHybridService{}

	// Configure basic expectations
	mockEndorService.On("GetResource").Return("test")
	mockEndorService.On("Validate").Return(nil)
	mockHybridService.On("GetResource").Return("test")

	// Test that they work as interfaces
	testServiceInterface(t, mockEndorService)
	testHybridServiceInterface(t, mockHybridService)

	// Verify expectations
	mockEndorService.AssertExpectations(t)
	mockHybridService.AssertExpectations(t)
}

// Helper function that accepts EndorServiceInterface to test interface usage
func testServiceInterface(t *testing.T, service interfaces.EndorServiceInterface) {
	resource := service.GetResource()
	assert.NotEmpty(t, resource, "Service should have a resource name")

	err := service.Validate()
	assert.NoError(t, err, "Service validation should succeed")
}

// Helper function that accepts EndorHybridServiceInterface to test interface usage
func testHybridServiceInterface(t *testing.T, service interfaces.EndorHybridServiceInterface) {
	resource := service.GetResource()
	assert.NotEmpty(t, resource, "Hybrid service should have a resource name")
}
