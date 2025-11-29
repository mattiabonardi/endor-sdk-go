package testutils

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
	"github.com/stretchr/testify/assert"
)

// TestInterfaceCompliance verifies that all mock implementations satisfy their respective interfaces.
// This test ensures that mocks can be used as drop-in replacements for concrete implementations.

// TestMockEndorServiceCompliance verifies MockEndorService implements EndorServiceInterface.
func TestMockEndorServiceCompliance(t *testing.T) {
	// Compile-time interface compliance check
	var _ interfaces.EndorServiceInterface = (*MockEndorService)(nil)

	// Runtime behavior verification
	mockService := &MockEndorService{}
	mockService.On("GetResource").Return("test-resource")
	mockService.On("GetDescription").Return("Test service")
	mockService.On("GetMethods").Return(map[string]interfaces.EndorServiceAction{})
	mockService.On("GetPriority").Return((*int)(nil))
	mockService.On("GetVersion").Return("1.0")
	mockService.On("Validate").Return(nil)

	// Test interface methods
	assert.Equal(t, "test-resource", mockService.GetResource())
	assert.Equal(t, "Test service", mockService.GetDescription())
	assert.NotNil(t, mockService.GetMethods())
	assert.Nil(t, mockService.GetPriority())
	assert.Equal(t, "1.0", mockService.GetVersion())
	assert.NoError(t, mockService.Validate())

	// Verify all expectations were met
	mockService.AssertExpectations(t)
}

// TestMockEndorHybridServiceCompliance verifies MockEndorHybridService implements EndorHybridServiceInterface.
func TestMockEndorHybridServiceCompliance(t *testing.T) {
	// Compile-time interface compliance check
	var _ interfaces.EndorHybridServiceInterface = (*MockEndorHybridService)(nil)

	// Runtime behavior verification
	mockHybridService := &MockEndorHybridService{}
	mockHybridService.On("GetResource").Return("test-hybrid-resource")
	mockHybridService.On("GetResourceDescription").Return("Test hybrid service")
	mockHybridService.On("GetPriority").Return((*int)(nil))
	mockHybridService.On("WithCategories", []interfaces.EndorHybridServiceCategory{}).Return(mockHybridService)
	mockHybridService.On("Validate").Return(nil)

	// Test interface methods
	assert.Equal(t, "test-hybrid-resource", mockHybridService.GetResource())
	assert.Equal(t, "Test hybrid service", mockHybridService.GetResourceDescription())
	assert.Nil(t, mockHybridService.GetPriority())

	// Test method chaining
	result := mockHybridService.WithCategories([]interfaces.EndorHybridServiceCategory{})
	assert.NotNil(t, result)
	assert.NoError(t, mockHybridService.Validate())

	// Verify all expectations were met
	mockHybridService.AssertExpectations(t)
}

// TestMockConfigProviderCompliance verifies MockConfigProvider implements ConfigProviderInterface.
func TestMockConfigProviderCompliance(t *testing.T) {
	// Compile-time interface compliance check
	var _ interfaces.ConfigProviderInterface = (*MockConfigProvider)(nil)

	// Runtime behavior verification
	mockConfig := &MockConfigProvider{}
	mockConfig.On("GetServerPort").Return("8080")
	mockConfig.On("GetDocumentDBUri").Return("mongodb://test:27017")
	mockConfig.On("IsHybridResourcesEnabled").Return(true)
	mockConfig.On("IsDynamicResourcesEnabled").Return(true)
	mockConfig.On("GetDynamicResourceDocumentDBName").Return("test-db")
	mockConfig.On("Reload").Return(nil)
	mockConfig.On("Validate").Return(nil)

	// Test interface methods
	assert.Equal(t, "8080", mockConfig.GetServerPort())
	assert.Equal(t, "mongodb://test:27017", mockConfig.GetDocumentDBUri())
	assert.True(t, mockConfig.IsHybridResourcesEnabled())
	assert.True(t, mockConfig.IsDynamicResourcesEnabled())
	assert.Equal(t, "test-db", mockConfig.GetDynamicResourceDocumentDBName())
	assert.NoError(t, mockConfig.Reload())
	assert.NoError(t, mockConfig.Validate())

	// Verify all expectations were met
	mockConfig.AssertExpectations(t)
}

// TestMockEndorContextCompliance verifies MockEndorContext implements EndorContextInterface.
func TestMockEndorContextCompliance(t *testing.T) {
	// Compile-time interface compliance check
	var _ interfaces.EndorContextInterface[TestUserPayload] = (*MockEndorContext[TestUserPayload])(nil)

	// Runtime behavior verification
	mockContext := &MockEndorContext[TestUserPayload]{}
	testPayload := TestUserPayload{Name: "Test User", Email: "test@example.com", Active: true}
	testSession := GetTestSession()
	categoryID := "test-category"

	mockContext.On("GetMicroServiceId").Return("test-service")
	mockContext.On("GetSession").Return(testSession)
	mockContext.On("GetPayload").Return(testPayload)
	mockContext.On("SetPayload", testPayload).Return()
	mockContext.On("GetResourceMetadataSchema").Return(GetTestUserSchema())
	mockContext.On("GetCategoryID").Return(&categoryID)
	mockContext.On("SetCategoryID", &categoryID).Return()
	mockContext.On("GetGinContext").Return((*gin.Context)(nil))

	// Test interface methods
	assert.Equal(t, "test-service", mockContext.GetMicroServiceId())
	assert.Equal(t, testSession, mockContext.GetSession())
	assert.Equal(t, testPayload, mockContext.GetPayload())
	assert.NotNil(t, mockContext.GetResourceMetadataSchema()) // Call the mocked method
	assert.Equal(t, &categoryID, mockContext.GetCategoryID())
	assert.Nil(t, mockContext.GetGinContext())

	mockContext.SetPayload(testPayload)
	mockContext.SetCategoryID(&categoryID)

	// Verify all expectations were met
	mockContext.AssertExpectations(t)
}

// TestMockEndorServiceActionCompliance verifies MockEndorServiceAction implements EndorServiceAction.
func TestMockEndorServiceActionCompliance(t *testing.T) {
	// Compile-time interface compliance check
	var _ interfaces.EndorServiceAction = (*MockEndorServiceAction)(nil)

	// Runtime behavior verification
	mockAction := &MockEndorServiceAction{}
	expectedOptions := interfaces.EndorServiceActionOptions{
		Description:     "Test action",
		Public:          true,
		ValidatePayload: true,
	}

	mockAction.On("GetOptions").Return(expectedOptions)
	mockAction.On("CreateHTTPCallback", "test-service").Return(func(c *gin.Context) {})

	// Test interface methods
	options := mockAction.GetOptions()
	assert.Equal(t, expectedOptions.Description, options.Description)
	assert.Equal(t, expectedOptions.Public, options.Public)
	assert.Equal(t, expectedOptions.ValidatePayload, options.ValidatePayload)

	callback := mockAction.CreateHTTPCallback("test-service")
	assert.NotNil(t, callback)

	// Verify all expectations were met
	mockAction.AssertExpectations(t)
}

// TestMockEndorHybridServiceCategoryCompliance verifies MockEndorHybridServiceCategory implements EndorHybridServiceCategory.
func TestMockEndorHybridServiceCategoryCompliance(t *testing.T) {
	// Compile-time interface compliance check
	var _ interfaces.EndorHybridServiceCategory = (*MockEndorHybridServiceCategory)(nil)

	// Runtime behavior verification
	mockCategory := &MockEndorHybridServiceCategory{}
	testActions := map[string]interfaces.EndorServiceAction{
		"admin/create": NewTestAction("Admin create", true),
		"admin/read":   NewTestAction("Admin read", true),
	}

	mockCategory.On("GetID").Return("admin")
	mockCategory.On("CreateDefaultActions", "users", "User management", interfaces.Schema{}).Return(testActions)

	// Test interface methods
	assert.Equal(t, "admin", mockCategory.GetID())

	actions := mockCategory.CreateDefaultActions("users", "User management", interfaces.Schema{})
	assert.NotNil(t, actions)
	assert.Len(t, actions, 2)
	assert.Contains(t, actions, "admin/create")
	assert.Contains(t, actions, "admin/read")

	// Verify all expectations were met
	mockCategory.AssertExpectations(t)
}

// TestBuilderInterfaceCompliance verifies that builder-created mocks satisfy interfaces.
func TestBuilderInterfaceCompliance(t *testing.T) {
	// Test service builder compliance
	testService := NewTestEndorService().
		WithResource("products").
		WithDescription("Product service").
		WithBasicMethods().
		Build()

	var _ interfaces.EndorServiceInterface = testService
	AssertServiceInterface(t, testService)

	// Test hybrid service builder compliance
	testHybridService := NewTestEndorHybridService().
		WithResource("users").
		WithResourceDescription("User service").
		WithDefaultCategory("admin").
		Build()

	var _ interfaces.EndorHybridServiceInterface = testHybridService
	AssertHybridServiceInterface(t, testHybridService)

	// Test config builder compliance
	testConfig := NewTestConfigProvider().
		WithServerPort("9999").
		WithHybridResourcesEnabled(false).
		Build()

	var _ interfaces.ConfigProviderInterface = testConfig
	AssertConfigProvider(t, testConfig)

	// Test context builder compliance
	testContext := NewTestEndorContext[TestUserPayload]().
		WithMicroServiceId("test-service").
		WithPayload(TestUserPayload{Name: "Test", Email: "test@example.com", Active: true}).
		Build()

	var _ interfaces.EndorContextInterface[TestUserPayload] = testContext
}
