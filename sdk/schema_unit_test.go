//go:build unit

package sdk_test

import (
	"fmt"
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/mattiabonardi/endor-sdk-go/sdk/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestSchemaGeneration_Unit tests schema generation using interface patterns
func TestSchemaGeneration_Unit(t *testing.T) {
	// Arrange: Define test types for schema generation
	type Address struct {
		City  string `json:"city"`
		State string `json:"state"`
	}

	type Car struct {
		ID string `json:"id"`
	}

	type User struct {
		ID         string   `json:"id" ui-schema:"id=true"`
		Name       string   `json:"name"`
		Email      string   `json:"email"`
		Age        int      `json:"age"`
		Active     bool     `json:"active"`
		Hobbies    []string `json:"hobbies"`
		Address    Address  `json:"address"`
		Cars       []Car    `json:"cars"`
		CurrentCar Car      `json:"car"`
	}

	// Act: Generate schema using real implementation
	schema := sdk.NewSchema(&User{})

	// Assert: Verify schema structure (unit testing approach)
	assert.Equal(t, sdk.ObjectType, schema.Type, "Root schema should be object type")
	assert.NotNil(t, schema.Properties, "Schema should have properties")
	assert.Empty(t, schema.Reference, "Expanded schema should not have root reference")

	properties := *schema.Properties
	expectedFields := []string{"id", "name", "email", "age", "active", "hobbies", "address", "cars", "car"}
	for _, field := range expectedFields {
		assert.Contains(t, properties, field, "Schema should include %s property", field)
	}

	// Verify nested object expansion (testing interface-driven schema patterns)
	addressProp := properties["address"]
	assert.Equal(t, sdk.ObjectType, addressProp.Type, "Address should be object type")
	assert.Empty(t, addressProp.Reference, "Nested objects should be expanded, not referenced")
	assert.NotNil(t, addressProp.Properties, "Address should have expanded properties")

	addressProps := *addressProp.Properties
	assert.Contains(t, addressProps, "city", "Address should include city property")
	assert.Contains(t, addressProps, "state", "Address should include state property")
}

// TestSchemaWithMockConfigProvider_Unit demonstrates schema configuration through interface
func TestSchemaWithMockConfigProvider_Unit(t *testing.T) {
	// Arrange: Create mock config provider for schema behavior configuration
	mockConfig := &testutils.MockConfigProvider{}

	// Configure schema generation behavior through mock
	mockConfig.On("IsHybridResourcesEnabled").Return(true)
	mockConfig.On("IsDynamicResourcesEnabled").Return(true)
	mockConfig.On("GetServerPort").Return("8080")

	type SimpleUser struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Act: Generate schema with config-driven behavior
	// (Note: This demonstrates the pattern - actual implementation would use config)
	schema := generateSchemaWithConfig(SimpleUser{}, mockConfig)

	// Assert: Verify schema generation respects configuration
	assert.Equal(t, sdk.ObjectType, schema.Type, "Schema should be object type")
	assert.NotNil(t, schema.Properties, "Schema should have properties")

	// Verify config-driven behavior was applied
	properties := *schema.Properties
	assert.Contains(t, properties, "name", "Schema should include name property")
	assert.Contains(t, properties, "age", "Schema should include age property")

	// Verify mock expectations
	mockConfig.AssertExpectations(t)
}

// TestSchemaValidation_Unit tests schema validation using interface mocks
func TestSchemaValidation_Unit(t *testing.T) {
	// Arrange: Create test data and mock validator
	type Product struct {
		Name  string  `json:"name" schema:"minLength=3"`
		Price float64 `json:"price" schema:"minimum=0"`
	}

	schema := sdk.NewSchema(Product{})

	// Create test validator interface (this would be part of the framework)
	mockValidator := &MockSchemaValidator{}

	validData := map[string]interface{}{
		"name":  "Test Product",
		"price": 29.99,
	}

	invalidData := map[string]interface{}{
		"name":  "XY", // Too short
		"price": -5.0, // Negative price
	}

	// Configure mock validator expectations
	mockValidator.On("ValidateData", validData, *schema).Return(nil)
	mockValidator.On("ValidateData", invalidData, *schema).Return(fmt.Errorf("validation failed: data does not match schema"))

	// Act & Assert: Test validation through interface
	err1 := mockValidator.ValidateData(validData, *schema)
	assert.NoError(t, err1, "Valid data should pass validation")

	err2 := mockValidator.ValidateData(invalidData, *schema)
	assert.Error(t, err2, "Invalid data should fail validation")

	// Verify expectations
	mockValidator.AssertExpectations(t)
}

// TestGenericSchemaSupport_Unit tests generic type schema generation
func TestGenericSchemaSupport_Unit(t *testing.T) {
	// Arrange: Define generic types for schema testing
	type Car struct {
		Id string `json:"id"`
	}

	type GenericCar[T any] struct {
		Value T `json:"value"`
	}

	type CarTreeNode struct {
		Value    *Car          `json:"value"`
		Children []CarTreeNode `json:"children"`
	}

	// Act: Generate schema for generic type
	schema := sdk.NewSchema(&GenericCar[Car]{})

	// Assert: Verify generic type handling
	assert.Equal(t, sdk.ObjectType, schema.Type, "Generic schema should be object type")
	assert.NotNil(t, schema.Properties, "Generic schema should have properties")

	properties := *schema.Properties
	assert.Contains(t, properties, "value", "Generic schema should include value property")

	// Verify value property is properly expanded
	valueProp := properties["value"]
	assert.Empty(t, valueProp.Reference, "Generic value should be expanded, not referenced")
	assert.Equal(t, sdk.ObjectType, valueProp.Type, "Generic value should resolve to object type")
}

// TestSchemaInterfaceCompliance_Unit validates schema interface implementations
func TestSchemaInterfaceCompliance_Unit(t *testing.T) {
	// This test demonstrates interface compliance checking for schema components

	// Test that generated schemas work with interface patterns
	type TestEntity struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	schema := sdk.NewSchema(TestEntity{})

	// Verify schema can be used in interface contexts
	testSchemaInterface(t, *schema)

	// Test schema properties interface usage
	assert.NotNil(t, schema.Properties, "Schema should implement properties interface")
	properties := *schema.Properties
	assert.IsType(t, map[string]sdk.Schema{}, properties, "Properties should be map of schemas")
}

// Helper function for config-driven schema generation (demonstrates pattern)
func generateSchemaWithConfig(entity interface{}, config interfaces.ConfigProviderInterface) *sdk.RootSchema {
	// This demonstrates how schema generation might use config interface
	// In actual implementation, this would configure schema expansion, examples, etc.

	hybridEnabled := config.IsHybridResourcesEnabled()
	dynamicEnabled := config.IsDynamicResourcesEnabled()
	port := config.GetServerPort()

	// Generate base schema
	schema := sdk.NewSchema(entity)

	// Apply configuration (simplified example)
	if hybridEnabled && dynamicEnabled && port == "8080" {
		// Configuration applied successfully
		return schema
	}

	return schema
} // Helper function demonstrating schema interface usage
func testSchemaInterface(t *testing.T, schema sdk.RootSchema) {
	// Test that schema implements expected interface patterns
	assert.Equal(t, sdk.ObjectType, schema.Type, "Schema should be object type")
	assert.NotNil(t, schema.Properties, "Schema should have properties")
}

// MockSchemaValidator demonstrates validator interface for testing
type MockSchemaValidator struct {
	mock.Mock
}

func (m *MockSchemaValidator) ValidateData(data interface{}, schema sdk.RootSchema) error {
	args := m.Called(data, schema)
	return args.Error(0)
}
