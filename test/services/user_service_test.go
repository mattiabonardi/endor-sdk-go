package services_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"github.com/stretchr/testify/assert"
)

// TestUserService_Creation tests basic service creation
func TestUserService_Creation(t *testing.T) {
	// Create the service
	userService := NewUserService()

	// Verify basic properties
	assert.Equal(t, "users", userService.GetResource())
	assert.Equal(t, "User Management System", userService.GetResourceDescription())
	assert.Nil(t, userService.GetPriority()) // No priority set
}

// TestUserService_ConvertToEndorService tests the hybrid service conversion
func TestUserService_ConvertToEndorService(t *testing.T) {
	userService := NewUserService()

	// Convert to traditional EndorService with empty metadata schema
	endorService := userService.ToEndorService(sdk.Schema{})

	// Verify the conversion
	assert.Equal(t, "users", endorService.GetResource())
	assert.Equal(t, "User Management System", endorService.GetDescription())

	// Verify that actions are available (should include CRUD + custom actions)
	methods := endorService.GetMethods()

	// Standard CRUD actions should be present
	assert.Contains(t, methods, "schema")
	assert.Contains(t, methods, "list")
	assert.Contains(t, methods, "create")
	assert.Contains(t, methods, "instance")
	assert.Contains(t, methods, "update")
	assert.Contains(t, methods, "delete")

	// Custom actions should be present
	assert.Contains(t, methods, "promote-user")
	assert.Contains(t, methods, "send-notification")
	assert.Contains(t, methods, "bulk-operations")

	// Category-specific actions should be present
	assert.Contains(t, methods, "admin/schema")
	assert.Contains(t, methods, "admin/list")
	assert.Contains(t, methods, "admin/create")
	assert.Contains(t, methods, "premium/schema")
	assert.Contains(t, methods, "premium/list")
	assert.Contains(t, methods, "premium/create")
}

// TestUserService_Categories tests category-based specialization
func TestUserService_Categories(t *testing.T) {
	userService := NewUserService()
	endorService := userService.ToEndorService(sdk.Schema{})
	methods := endorService.GetMethods()

	// Test that admin category actions exist
	adminSchemaAction, exists := methods["admin/schema"]
	assert.True(t, exists, "Admin schema action should exist")
	assert.NotNil(t, adminSchemaAction)

	adminCreateAction, exists := methods["admin/create"]
	assert.True(t, exists, "Admin create action should exist")
	assert.NotNil(t, adminCreateAction)

	// Test that premium category actions exist
	premiumSchemaAction, exists := methods["premium/schema"]
	assert.True(t, exists, "Premium schema action should exist")
	assert.NotNil(t, premiumSchemaAction)

	premiumCreateAction, exists := methods["premium/create"]
	assert.True(t, exists, "Premium create action should exist")
	assert.NotNil(t, premiumCreateAction)
}

// TestUserService_CustomActions tests custom action availability
func TestUserService_CustomActions(t *testing.T) {
	userService := NewUserService()
	endorService := userService.ToEndorService(sdk.Schema{})
	methods := endorService.GetMethods()

	// Test promote-user action
	promoteAction, exists := methods["promote-user"]
	assert.True(t, exists, "Promote user action should exist")
	assert.NotNil(t, promoteAction)

	options := promoteAction.GetOptions()
	assert.Equal(t, "Promote a user to admin with specified level", options.Description)
	assert.False(t, options.Public) // Actions are private by default
	assert.True(t, options.ValidatePayload)

	// Test send-notification action
	notifyAction, exists := methods["send-notification"]
	assert.True(t, exists, "Send notification action should exist")
	assert.NotNil(t, notifyAction)

	notifyOptions := notifyAction.GetOptions()
	assert.Equal(t, "Send notification to a specific user", notifyOptions.Description)

	// Test bulk-operations action
	bulkAction, exists := methods["bulk-operations"]
	assert.True(t, exists, "Bulk operations action should exist")
	assert.NotNil(t, bulkAction)

	bulkOptions := bulkAction.GetOptions()
	assert.Equal(t, "Perform bulk operations on multiple users", bulkOptions.Description)
}

// Example showing how the new architecture enables testing patterns
func TestUserService_MockedActionExecution(t *testing.T) {
	// This test demonstrates how you could test action execution
	// with mocked dependencies in a real application

	userService := NewUserService()
	endorService := userService.ToEndorService(sdk.Schema{})

	// Get the promote-user action
	promoteAction := endorService.GetMethods()["promote-user"]
	assert.NotNil(t, promoteAction)

	// In a real test, you would:
	// 1. Create a mock HTTP context
	// 2. Call the HTTP callback
	// 3. Verify the response
	// This demonstrates the testing capabilities of the new architecture
}
