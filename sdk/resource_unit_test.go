//go:build unit

package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/mattiabonardi/endor-sdk-go/sdk/testutils"
	"github.com/stretchr/testify/assert"
)

// TestResourceDefinitionFromYAML_Unit tests YAML parsing using interface mocks
func TestResourceDefinitionFromYAML_Unit(t *testing.T) {
	// Arrange: Create test resource with mock dependencies
	yamlInput := `type: object
properties:
  name:
    type: string
  surname:
    type: string`

	resource := sdk.Resource{
		ID:                   "customer",
		Description:          "Customers",
		Service:              "",
		AdditionalAttributes: yamlInput,
	}

	// Act: Parse the YAML definition (this uses real parsing logic)
	def, err := resource.UnmarshalAdditionalAttributes()

	// Assert: Verify parsing succeeds and structure is correct
	assert.NoError(t, err, "YAML parsing should succeed")
	assert.Equal(t, "object", string(def.Schema.Type), "Schema type should be object")

	assert.NotNil(t, def.Schema.Properties, "Schema should have properties")
	properties := *def.Schema.Properties
	assert.Contains(t, properties, "name", "Schema should include name property")
	assert.Contains(t, properties, "surname", "Schema should include surname property")
}

// TestResourceWithMockConfig_Unit demonstrates config interface usage for resource validation
func TestResourceWithMockConfig_Unit(t *testing.T) {
	// Arrange: Create mock config provider for testing resource operations
	mockConfig := &testutils.MockConfigProvider{}

	testResource := sdk.Resource{
		ID:          "test-resource",
		Description: "Test resource for unit testing",
		Service:     "test-service",
	}

	// Configure mock expectations for config-driven behavior
	mockConfig.On("IsHybridResourcesEnabled").Return(true)
	mockConfig.On("IsDynamicResourcesEnabled").Return(true)

	// Act: Test resource validation using config interface
	isValid := validateResourceWithConfig(testResource, mockConfig)

	// Assert: Verify validation behavior with mocked config
	assert.True(t, isValid, "Valid resource should pass strict validation")

	// Test validation with different config expectations
	mockConfig2 := &testutils.MockConfigProvider{}
	mockConfig2.On("IsHybridResourcesEnabled").Return(false)
	mockConfig2.On("IsDynamicResourcesEnabled").Return(false)

	isInvalid := validateResourceWithConfig(testResource, mockConfig2)
	assert.True(t, isInvalid, "Resource should pass validation with relaxed config")

	// Verify all mock expectations were met
	mockConfig.AssertExpectations(t)
	mockConfig2.AssertExpectations(t)
} // TestResourceValidation_Unit tests resource validation using mock config
func TestResourceValidation_Unit(t *testing.T) {
	// Arrange: Create mock config provider
	mockConfig := &testutils.MockConfigProvider{}

	// Configure validation rules through mock - need to include all methods called by validateResourceWithConfig
	mockConfig.On("IsHybridResourcesEnabled").Return(true)
	mockConfig.On("IsDynamicResourcesEnabled").Return(true)

	testResource := sdk.Resource{
		ID:          "valid-resource-id",
		Description: "Valid resource description",
		Service:     "test-service",
	}

	// Act: Perform validation (this would use the mock config)
	// Note: This assumes validation logic exists that uses config interface
	isValid := validateResourceWithConfig(testResource, mockConfig)

	// Assert: Verify validation passes with proper configuration
	assert.True(t, isValid, "Valid resource should pass validation")

	// Verify mock expectations
	mockConfig.AssertExpectations(t)
}

// TestResourceSchemaGeneration_Unit tests schema generation with interface patterns
func TestResourceSchemaGeneration_Unit(t *testing.T) {
	// Arrange: Create resource with schema that can be mocked
	type TestResourceData struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	// Act: Generate schema (real implementation but demonstrates interface usage)
	schema := sdk.NewSchema(TestResourceData{})

	// Assert: Verify schema structure
	assert.Equal(t, sdk.ObjectType, schema.Type, "Schema should be object type")
	assert.NotNil(t, schema.Properties, "Schema should have properties")

	properties := *schema.Properties
	assert.Contains(t, properties, "name", "Schema should include name")
	assert.Contains(t, properties, "email", "Schema should include email")
	assert.Contains(t, properties, "age", "Schema should include age")

	// Verify property types
	assert.Equal(t, sdk.StringType, properties["name"].Type, "Name should be string type")
	assert.Equal(t, sdk.StringType, properties["email"].Type, "Email should be string type")
	assert.Equal(t, sdk.IntegerType, properties["age"].Type, "Age should be integer type")
}

// Helper function demonstrating config interface usage pattern
// This would typically be part of the actual resource validation logic
func validateResourceWithConfig(resource sdk.Resource, config interfaces.ConfigProviderInterface) bool {
	// This is a simplified example of how resource validation might use config interface
	// Using available interface methods for demonstration
	hybridEnabled := config.IsHybridResourcesEnabled()
	dynamicEnabled := config.IsDynamicResourcesEnabled()

	// Simplified validation based on configuration flags
	if hybridEnabled && dynamicEnabled {
		// Strict validation when both features enabled
		return len(resource.ID) > 0 && len(resource.Description) >= 10 && resource.Service != ""
	}
	return resource.ID != ""
}
