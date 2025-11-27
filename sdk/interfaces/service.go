// Package interfaces provides the core interfaces for endor-sdk-go framework components.
// These interfaces enable dependency injection, testing with mocks, and service composition
// while maintaining the framework's dual-service architecture (EndorService + EndorHybridService).
package interfaces

import (
	"github.com/gin-gonic/gin"
)

// EndorServiceInterface defines the contract for EndorService components.
// This interface enables mocking and testing of services without requiring
// concrete implementations or external dependencies like databases.
//
// The interface follows Go's "accept interfaces, return structs" philosophy
// and supports the framework's manual control service pattern.
//
// Example usage:
//
//	// Production usage with concrete implementation
//	service := sdk.EndorService{
//		Resource: "users",
//		Description: "User management service",
//		Methods: map[string]sdk.EndorServiceAction{
//			"create": sdk.NewAction(createUserHandler, "Create a new user"),
//		},
//	}
//
//	// Testing usage with mock
//	mockService := &MockEndorService{}
//	mockService.On("GetResource").Return("users")
//	mockService.On("GetDescription").Return("Mock user service")
type EndorServiceInterface interface {
	// GetResource returns the resource name that this service manages.
	// This is used for API routing and service identification.
	GetResource() string

	// GetDescription returns a human-readable description of the service.
	// Used in API documentation and service discovery.
	GetDescription() string

	// GetMethods returns the map of available actions for this service.
	// Each action corresponds to an API endpoint and handler function.
	GetMethods() map[string]EndorServiceAction

	// GetPriority returns the service priority for conflict resolution.
	// Services with higher priority take precedence during registration.
	// Returns nil if no specific priority is set.
	GetPriority() *int

	// GetVersion returns the API version for this service.
	// Used for API versioning and backward compatibility.
	// Returns empty string if not specified.
	GetVersion() string

	// Validate performs service configuration validation.
	// Should check that all required fields are properly configured
	// and that the service can be safely registered.
	Validate() error
}

// EndorHybridServiceInterface defines the contract for EndorHybridService components.
// This interface enables testing of hybrid services that provide automated CRUD operations
// with category-based specialization and dynamic schema generation.
//
// Hybrid services combine the automation of CRUD operations with the flexibility
// of custom actions, supporting the framework's type-safe generic architecture.
//
// Example usage:
//
//	// Production usage
//	hybridService := sdk.NewHybridService[User]("users", "User management")
//	hybridService = hybridService.WithCategories([]sdk.EndorHybridServiceCategory{
//		sdk.NewEndorHybridServiceCategory[User, AdminUser](adminCategory),
//	}).WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
//		return map[string]sdk.EndorServiceAction{
//			"custom": sdk.NewAction(customHandler, "Custom user action"),
//		}
//	})
//
//	// Testing usage with mock
//	mockHybridService := &MockEndorHybridService{}
//	mockHybridService.On("GetResource").Return("users")
//	mockHybridService.On("ToEndorService", mock.Any).Return(mockEndorService)
type EndorHybridServiceInterface interface {
	// GetResource returns the resource name that this hybrid service manages.
	GetResource() string

	// GetResourceDescription returns a human-readable description of the resource.
	GetResourceDescription() string

	// GetPriority returns the service priority for conflict resolution.
	// Returns nil if no specific priority is set.
	GetPriority() *int

	// WithCategories configures the hybrid service with category-based specializations.
	// Each category provides specialized CRUD operations and schema extensions.
	// Returns a new hybrid service instance with the categories applied.
	WithCategories(categories []EndorHybridServiceCategory) EndorHybridServiceInterface

	// WithActions configures the hybrid service with custom actions beyond CRUD.
	// The provided function receives a schema getter and returns custom actions.
	// This enables dynamic action creation based on the resolved schema.
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceAction) EndorHybridServiceInterface

	// ToEndorService converts the hybrid service to a concrete EndorService.
	// This performs the final resolution of schema, categories, and actions
	// into a complete service ready for registration and routing.
	ToEndorService(metadataSchema Schema) EndorServiceInterface

	// Validate performs hybrid service configuration validation.
	// Should verify that categories and actions are properly configured.
	Validate() error
}

// EndorServiceAction represents an action that can be performed on a service.
// This interface abstracts the HTTP callback creation and configuration options.
type EndorServiceAction interface {
	// CreateHTTPCallback creates a Gin HTTP handler for this action.
	// The microserviceId is used for request tracking and service identification.
	CreateHTTPCallback(microserviceId string) func(c *gin.Context)

	// GetOptions returns the configuration options for this action.
	// Includes description, visibility, validation settings, and input schema.
	GetOptions() EndorServiceActionOptions
}

// EndorServiceActionOptions contains configuration for service actions.
type EndorServiceActionOptions struct {
	Description     string      // Human-readable description for documentation
	Public          bool        // Whether the action is publicly accessible
	ValidatePayload bool        // Whether to validate request payload
	InputSchema     *RootSchema // Schema for request validation
}

// EndorHybridServiceCategory represents a category-based specialization.
// Categories enable automatic CRUD operations with specialized schemas and behaviors.
type EndorHybridServiceCategory interface {
	// GetID returns the unique identifier for this category.
	GetID() string

	// CreateDefaultActions generates the default CRUD actions for this category.
	// Uses the provided resource information and metadata schema to create
	// specialized endpoints like "categoryId/create", "categoryId/list", etc.
	CreateDefaultActions(resource string, resourceDescription string, metadataSchema Schema) map[string]EndorServiceAction
}

// Schema represents a JSON schema for data validation and documentation.
// Used throughout the framework for automatic API documentation and validation.
type Schema struct {
	Type        SchemaType         `json:"type,omitempty"`
	Properties  *map[string]Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Reference   string             `json:"$ref,omitempty"`
	Enum        *[]string          `json:"enum,omitempty"`
	Definitions map[string]Schema  `json:"definitions,omitempty"`
	Format      string             `json:"format,omitempty"`
	Default     interface{}        `json:"default,omitempty"`
	Examples    []interface{}      `json:"examples,omitempty"`
}

// RootSchema extends Schema with additional root-level metadata.
type RootSchema struct {
	Schema
}

// SchemaType represents the type of a schema field.
type SchemaType string

const (
	StringType  SchemaType = "string"
	NumberType  SchemaType = "number"
	IntegerType SchemaType = "integer"
	BooleanType SchemaType = "boolean"
	ArrayType   SchemaType = "array"
	ObjectType  SchemaType = "object"
)

// EndorContext represents the request context passed to action handlers.
// Contains session information, request data, and framework context.
type EndorContext[T any] struct {
	MicroServiceId string
	Session        Session
	CategoryID     *string
	GinContext     *gin.Context
	Payload        T
}

// Session contains user session information for authenticated requests.
type Session struct {
	Id          string // Session identifier
	Username    string // Authenticated username
	Development bool   // Development mode flag
}

// Response represents the standardized response format for all framework actions.
type Response[T any] struct {
	Data     *T          `json:"data,omitempty"`
	Messages []Message   `json:"messages,omitempty"`
	Schema   *RootSchema `json:"schema,omitempty"`
}

// Message represents a response message with severity level.
type Message struct {
	Gravity MessageGravity `json:"gravity"`
	Value   string         `json:"value"`
}

// MessageGravity indicates the severity level of a message.
type MessageGravity string

const (
	Info    MessageGravity = "info"
	Warning MessageGravity = "warning"
	Error   MessageGravity = "error"
	Fatal   MessageGravity = "fatal"
)

// NoPayload is used for actions that don't require input payload.
type NoPayload struct{}
