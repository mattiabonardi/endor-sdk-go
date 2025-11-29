// Package interfaces provides configuration and context interfaces for endor-sdk-go framework.
// These interfaces enable dependency injection, testing with mocks, and flexible configuration
// management while maintaining the framework's configuration patterns and context propagation.
package interfaces

import (
	"github.com/gin-gonic/gin"
)

// ConfigProviderInterface defines the contract for configuration access in the framework.
// This interface abstracts configuration loading and access patterns, enabling testing
// with different configuration values and supporting both environment-based and custom
// configuration sources.
//
// The interface follows the framework's existing configuration access patterns while
// making them testable and mockable for unit testing scenarios.
//
// Example usage:
//
//	// Production usage with concrete implementation
//	config := sdk.GetConfig() // implements ConfigProviderInterface
//	port := config.GetServerPort()
//	dbUri := config.GetDocumentDBUri()
//
//	// Testing usage with mock configuration
//	mockConfig := &MockConfigProvider{}
//	mockConfig.On("GetServerPort").Return("8080")
//	mockConfig.On("GetDocumentDBUri").Return("mongodb://test:27017")
//	mockConfig.On("IsHybridResourcesEnabled").Return(true)
//
//	// Custom configuration for testing
//	testConfig := &TestConfigProvider{
//		ServerPort: "9999",
//		DocumentDBUri: "mongodb://localhost:27017/test",
//		HybridResourcesEnabled: false,
//	}
type ConfigProviderInterface interface {
	// GetServerPort returns the HTTP server port configuration.
	// Typically read from PORT environment variable or defaults to "8080".
	GetServerPort() string

	// GetDocumentDBUri returns the MongoDB connection URI.
	// Typically read from DOCUMENT_DB_URI environment variable.
	GetDocumentDBUri() string

	// IsHybridResourcesEnabled returns whether hybrid resource functionality is enabled.
	// Typically read from HYBRID_RESOURCES_ENABLED environment variable.
	IsHybridResourcesEnabled() bool

	// IsDynamicResourcesEnabled returns whether dynamic resource functionality is enabled.
	// Typically read from DYNAMIC_RESOURCES_ENABLED environment variable.
	IsDynamicResourcesEnabled() bool

	// GetDynamicResourceDocumentDBName returns the database name for dynamic resources.
	// This is typically set at runtime based on the microservice ID.
	GetDynamicResourceDocumentDBName() string

	// Reload forces configuration reload from sources (environment variables, files).
	// Useful for testing scenarios where configuration changes need to be applied.
	// Returns error if configuration reload fails.
	Reload() error

	// Validate performs configuration validation to ensure all required values are present
	// and properly formatted. Returns error if configuration is invalid.
	Validate() error
}

// EndorContextInterface defines the contract for request context management in the framework.
// This interface abstracts context operations and data access, enabling testing with
// different context scenarios and mocking context propagation patterns.
//
// The interface supports the framework's generic context pattern while making context
// operations testable and providing abstraction for session management and payload handling.
//
// Example usage:
//
//	// Production usage with concrete implementation
//	ctx := &sdk.EndorContext[UserPayload]{
//		MicroServiceId: "user-service",
//		Session: sdk.Session{Id: "sess123", Username: "testuser"},
//		Payload: UserPayload{Name: "John", Email: "john@example.com"},
//	}
//
//	// Testing usage with mock context
//	mockContext := &MockEndorContext[UserPayload]{}
//	mockContext.On("GetMicroServiceId").Return("test-service")
//	mockContext.On("GetSession").Return(testSession)
//	mockContext.On("GetPayload").Return(testPayload)
//
//	// Test context builder for unit tests
//	testContext := NewTestContext[UserPayload]().
//		WithMicroServiceId("test-service").
//		WithSession(Session{Username: "testuser"}).
//		WithPayload(UserPayload{Name: "Test User"}).
//		Build()
type EndorContextInterface[T any] interface {
	// GetMicroServiceId returns the identifier of the microservice handling the request.
	// Used for service identification, logging, and resource routing.
	GetMicroServiceId() string

	// GetSession returns the authentication session information for the request.
	// Contains user identification, authentication status, and session metadata.
	// Returns the session interface - actual type is sdk.Session
	GetSession() interface{}

	// GetPayload returns the typed payload data for the request.
	// The payload type is determined by the service action's expected input type.
	GetPayload() T

	// SetPayload updates the payload data for the request context.
	// Used during request processing to modify or enrich payload information.
	SetPayload(payload T)

	// GetResourceMetadataSchema returns the schema definition for the resource.
	// Used for validation, documentation generation, and API schema responses.
	// Returns the schema interface - actual type is sdk.RootSchema
	GetResourceMetadataSchema() interface{}

	// GetCategoryID returns the category identifier for specialized resource operations.
	// Returns nil if the request is not for a categorized resource.
	GetCategoryID() *string

	// SetCategoryID sets the category identifier for specialized resource operations.
	// Used when routing requests to category-specific implementations.
	SetCategoryID(categoryID *string)

	// GetGinContext returns the underlying Gin HTTP context for the request.
	// Provides access to HTTP-specific functionality like headers, query parameters, etc.
	GetGinContext() *gin.Context
}
