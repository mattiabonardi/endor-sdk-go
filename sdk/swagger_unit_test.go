//go:build unit

package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/mattiabonardi/endor-sdk-go/sdk/testutils"
	"github.com/stretchr/testify/assert"
)

// TestSwaggerInterfaceCompliance_Unit tests interface patterns for swagger-compatible services
func TestSwaggerInterfaceCompliance_Unit(t *testing.T) {
	// Arrange: Create mock service that implements interface
	mockService := testutils.NewTestEndorService().
		WithResource("test-resource").
		WithDescription("Test service for swagger").
		Build()

	// Verify it implements the interface
	var serviceInterface interfaces.EndorServiceInterface = mockService

	// Test interface compliance
	testServiceForSwagger(t, serviceInterface)

	// Create concrete service for swagger generation (since swagger expects concrete types)
	concreteService := sdk.EndorService{
		Resource:    "test-resource",
		Description: "Test service for swagger",
		Methods:     make(map[string]sdk.EndorServiceAction),
	}

	services := []sdk.EndorService{concreteService}
	def, err := sdk.CreateSwaggerDefinition("test-service", "test.com", services, "/api")

	// Assert: Verify swagger generation works with interface-tested patterns
	assert.NoError(t, err, "Swagger generation should succeed with interface-tested patterns")
	assert.Equal(t, "3.1.0", def.OpenAPI, "Should use OpenAPI 3.1.0")
	assert.Equal(t, "test-service", def.Info.Title, "Should use provided service name")
}

// TestConfigIntegrationPattern_Unit demonstrates config interface pattern for service configuration
func TestConfigIntegrationPattern_Unit(t *testing.T) {
	// Arrange: Create mock config for service configuration
	mockConfig := &testutils.MockConfigProvider{}

	// Configure behavior
	mockConfig.On("GetServerPort").Return("8080")
	mockConfig.On("GetDocumentDBUri").Return("mongodb://test:27017")
	mockConfig.On("IsHybridResourcesEnabled").Return(true)
	mockConfig.On("IsDynamicResourcesEnabled").Return(false)

	// Act: Use config in service-like pattern (demonstration)
	serviceConfig := configureServiceWithInterface(mockConfig)

	// Assert: Verify configuration applied correctly
	assert.Equal(t, "8080", serviceConfig.Port)
	assert.Equal(t, "mongodb://test:27017", serviceConfig.DBUri)
	assert.True(t, serviceConfig.HybridEnabled)
	assert.False(t, serviceConfig.DynamicEnabled)

	// Verify expectations
	mockConfig.AssertExpectations(t)
}

// Helper function to test service interface for swagger compatibility
func testServiceForSwagger(t *testing.T, service interfaces.EndorServiceInterface) {
	// Test interface methods required for swagger generation compatibility
	resource := service.GetResource()
	assert.NotEmpty(t, resource, "Service should have resource for swagger")

	description := service.GetDescription()
	assert.NotEmpty(t, description, "Service should have description for swagger")

	methods := service.GetMethods()
	assert.NotNil(t, methods, "Service should have methods for swagger")

	err := service.Validate()
	assert.NoError(t, err, "Service should validate for swagger generation")
}

// Helper struct for config demonstration
type ServiceConfig struct {
	Port           string
	DBUri          string
	HybridEnabled  bool
	DynamicEnabled bool
}

// Helper function demonstrating config interface usage
func configureServiceWithInterface(config interfaces.ConfigProviderInterface) ServiceConfig {
	return ServiceConfig{
		Port:           config.GetServerPort(),
		DBUri:          config.GetDocumentDBUri(),
		HybridEnabled:  config.IsHybridResourcesEnabled(),
		DynamicEnabled: config.IsDynamicResourcesEnabled(),
	}
}
