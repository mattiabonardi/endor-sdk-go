package testutils

import (
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// TestFixtures provides common test data sets and scenarios
// for different service types and testing patterns.

// Common test payloads for different scenarios

// TestUserPayload represents a typical user data payload for testing.
type TestUserPayload struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Role     string                 `json:"role,omitempty"`
	Active   bool                   `json:"active"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TestProductPayload represents a typical product data payload for testing.
type TestProductPayload struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Category    string   `json:"category"`
	InStock     bool     `json:"inStock"`
	Tags        []string `json:"tags,omitempty"`
}

// TestOrderPayload represents a typical order data payload for testing.
type TestOrderPayload struct {
	ID         string                 `json:"id,omitempty"`
	CustomerID string                 `json:"customerId"`
	Items      []TestOrderItem        `json:"items"`
	Total      float64                `json:"total"`
	Status     string                 `json:"status"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// TestOrderItem represents an order item for testing order scenarios.
type TestOrderItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// NoPayload represents actions that don't require input payload.
type NoPayload struct{}

// Test data generators

// GetTestUsers returns a set of test user payloads for various testing scenarios.
func GetTestUsers() []TestUserPayload {
	return []TestUserPayload{
		{
			ID:     "user-1",
			Name:   "John Doe",
			Email:  "john@example.com",
			Role:   "admin",
			Active: true,
			Metadata: map[string]interface{}{
				"department": "engineering",
				"level":      "senior",
			},
		},
		{
			ID:     "user-2",
			Name:   "Jane Smith",
			Email:  "jane@example.com",
			Role:   "user",
			Active: true,
		},
		{
			ID:     "user-3",
			Name:   "Bob Wilson",
			Email:  "bob@example.com",
			Role:   "user",
			Active: false,
			Metadata: map[string]interface{}{
				"suspended": true,
				"reason":    "violation",
			},
		},
	}
}

// GetTestProducts returns a set of test product payloads for various testing scenarios.
func GetTestProducts() []TestProductPayload {
	return []TestProductPayload{
		{
			ID:          "product-1",
			Name:        "Laptop Computer",
			Description: "High-performance laptop for development",
			Price:       1299.99,
			Category:    "electronics",
			InStock:     true,
			Tags:        []string{"laptop", "computer", "development"},
		},
		{
			ID:          "product-2",
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse",
			Price:       29.99,
			Category:    "accessories",
			InStock:     true,
			Tags:        []string{"mouse", "wireless", "ergonomic"},
		},
		{
			ID:          "product-3",
			Name:        "Monitor Stand",
			Description: "Adjustable monitor stand",
			Price:       49.99,
			Category:    "accessories",
			InStock:     false,
			Tags:        []string{"monitor", "stand", "adjustable"},
		},
	}
}

// GetTestOrders returns a set of test order payloads for various testing scenarios.
func GetTestOrders() []TestOrderPayload {
	return []TestOrderPayload{
		{
			ID:         "order-1",
			CustomerID: "user-1",
			Items: []TestOrderItem{
				{ProductID: "product-1", Quantity: 1, Price: 1299.99},
				{ProductID: "product-2", Quantity: 2, Price: 29.99},
			},
			Total:     1359.97,
			Status:    "completed",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:         "order-2",
			CustomerID: "user-2",
			Items: []TestOrderItem{
				{ProductID: "product-3", Quantity: 1, Price: 49.99},
			},
			Total:     49.99,
			Status:    "pending",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			Metadata: map[string]interface{}{
				"priority": "high",
				"notes":    "Rush order",
			},
		},
	}
}

// Test schemas for different payload types

// GetTestUserSchema returns a schema definition for user payloads.
func GetTestUserSchema() interfaces.RootSchema {
	return interfaces.RootSchema{
		Schema: interfaces.Schema{
			Type: interfaces.ObjectType,
			Properties: &map[string]interfaces.Schema{
				"id": {
					Type:   interfaces.StringType,
					Format: "uuid",
				},
				"name": {
					Type: interfaces.StringType,
				},
				"email": {
					Type:   interfaces.StringType,
					Format: "email",
				},
				"role": {
					Type: interfaces.StringType,
					Enum: &[]string{"admin", "user", "guest"},
				},
				"active": {
					Type: interfaces.BooleanType,
				},
				"metadata": {
					Type: interfaces.ObjectType,
				},
			},
			Required: []string{"name", "email", "active"},
		},
	}
}

// GetTestProductSchema returns a schema definition for product payloads.
func GetTestProductSchema() interfaces.RootSchema {
	return interfaces.RootSchema{
		Schema: interfaces.Schema{
			Type: interfaces.ObjectType,
			Properties: &map[string]interfaces.Schema{
				"id": {
					Type:   interfaces.StringType,
					Format: "uuid",
				},
				"name": {
					Type: interfaces.StringType,
				},
				"description": {
					Type: interfaces.StringType,
				},
				"price": {
					Type: interfaces.NumberType,
				},
				"category": {
					Type: interfaces.StringType,
				},
				"inStock": {
					Type: interfaces.BooleanType,
				},
				"tags": {
					Type: interfaces.ArrayType,
					Items: &interfaces.Schema{
						Type: interfaces.StringType,
					},
				},
			},
			Required: []string{"name", "price", "category", "inStock"},
		},
	}
}

// GetTestOrderSchema returns a schema definition for order payloads.
func GetTestOrderSchema() interfaces.RootSchema {
	return interfaces.RootSchema{
		Schema: interfaces.Schema{
			Type: interfaces.ObjectType,
			Properties: &map[string]interfaces.Schema{
				"id": {
					Type:   interfaces.StringType,
					Format: "uuid",
				},
				"customerId": {
					Type:   interfaces.StringType,
					Format: "uuid",
				},
				"items": {
					Type: interfaces.ArrayType,
					Items: &interfaces.Schema{
						Type: interfaces.ObjectType,
						Properties: &map[string]interfaces.Schema{
							"productId": {Type: interfaces.StringType},
							"quantity":  {Type: interfaces.IntegerType},
							"price":     {Type: interfaces.NumberType},
						},
						Required: []string{"productId", "quantity", "price"},
					},
				},
				"total": {
					Type: interfaces.NumberType,
				},
				"status": {
					Type: interfaces.StringType,
					Enum: &[]string{"pending", "processing", "completed", "cancelled"},
				},
				"createdAt": {
					Type:   interfaces.StringType,
					Format: "date-time",
				},
				"metadata": {
					Type: interfaces.ObjectType,
				},
			},
			Required: []string{"customerId", "items", "total", "status"},
		},
	}
}

// Test session fixtures

// GetTestSession returns a test session for admin user.
func GetTestSession() map[string]interface{} {
	return map[string]interface{}{
		"id":          "test-session-123",
		"username":    "test-admin",
		"development": false,
		"roles":       []string{"admin", "user"},
		"permissions": []string{"read", "write", "delete"},
	}
}

// GetTestUserSession returns a test session for regular user.
func GetTestUserSession() map[string]interface{} {
	return map[string]interface{}{
		"id":          "test-session-456",
		"username":    "test-user",
		"development": false,
		"roles":       []string{"user"},
		"permissions": []string{"read"},
	}
}

// GetTestDeveloperSession returns a test session for development mode.
func GetTestDeveloperSession() map[string]interface{} {
	return map[string]interface{}{
		"id":          "dev-session-789",
		"username":    "developer",
		"development": true,
		"roles":       []string{"admin", "developer"},
		"permissions": []string{"read", "write", "delete", "debug"},
	}
}

// Common test scenarios

// TestScenario represents a complete test scenario with context and expected outcomes.
type TestScenario struct {
	Name        string
	Description string
	Context     interface{}
	Payload     interface{}
	Expected    interface{}
	ShouldError bool
	ErrorType   string
}

// GetCRUDTestScenarios returns test scenarios for CRUD operations.
func GetCRUDTestScenarios() []TestScenario {
	testUsers := GetTestUsers()
	return []TestScenario{
		{
			Name:        "Create User Success",
			Description: "Successfully create a new user",
			Context:     GetTestSession(),
			Payload:     testUsers[0],
			Expected: map[string]interface{}{
				"status": "success",
				"id":     testUsers[0].ID,
			},
			ShouldError: false,
		},
		{
			Name:        "Create User Invalid Email",
			Description: "Fail to create user with invalid email",
			Context:     GetTestSession(),
			Payload: TestUserPayload{
				Name:   "Invalid User",
				Email:  "invalid-email",
				Active: true,
			},
			ShouldError: true,
			ErrorType:   "validation",
		},
		{
			Name:        "Read User Success",
			Description: "Successfully retrieve an existing user",
			Context:     GetTestSession(),
			Payload:     map[string]string{"id": "user-1"},
			Expected:    testUsers[0],
			ShouldError: false,
		},
		{
			Name:        "Read User Not Found",
			Description: "Fail to retrieve non-existent user",
			Context:     GetTestSession(),
			Payload:     map[string]string{"id": "non-existent"},
			ShouldError: true,
			ErrorType:   "not_found",
		},
		{
			Name:        "Update User Success",
			Description: "Successfully update an existing user",
			Context:     GetTestSession(),
			Payload: TestUserPayload{
				ID:     "user-1",
				Name:   "John Updated",
				Email:  "john.updated@example.com",
				Role:   "admin",
				Active: true,
			},
			Expected: map[string]interface{}{
				"status": "updated",
				"id":     "user-1",
			},
			ShouldError: false,
		},
		{
			Name:        "Delete User Success",
			Description: "Successfully delete an existing user",
			Context:     GetTestSession(),
			Payload:     map[string]string{"id": "user-3"},
			Expected: map[string]interface{}{
				"status": "deleted",
				"id":     "user-3",
			},
			ShouldError: false,
		},
	}
}

// GetAuthorizationTestScenarios returns test scenarios for authorization testing.
func GetAuthorizationTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:        "Admin Access Success",
			Description: "Admin user can access protected resource",
			Context:     GetTestSession(),
			Payload:     map[string]string{"resource": "admin-only"},
			Expected: map[string]interface{}{
				"access": "granted",
			},
			ShouldError: false,
		},
		{
			Name:        "User Access Denied",
			Description: "Regular user cannot access admin resource",
			Context:     GetTestUserSession(),
			Payload:     map[string]string{"resource": "admin-only"},
			ShouldError: true,
			ErrorType:   "unauthorized",
		},
		{
			Name:        "Development Mode Access",
			Description: "Developer session has special privileges",
			Context:     GetTestDeveloperSession(),
			Payload:     map[string]string{"resource": "debug-endpoint"},
			Expected: map[string]interface{}{
				"access":     "granted",
				"debug_info": "available",
			},
			ShouldError: false,
		},
	}
}

// Helper functions for test data manipulation

// CloneTestUser creates a deep copy of a test user payload for modification.
func CloneTestUser(user TestUserPayload) TestUserPayload {
	clone := user
	if user.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range user.Metadata {
			clone.Metadata[k] = v
		}
	}
	return clone
}

// CloneTestProduct creates a deep copy of a test product payload for modification.
func CloneTestProduct(product TestProductPayload) TestProductPayload {
	clone := product
	if product.Tags != nil {
		clone.Tags = make([]string, len(product.Tags))
		copy(clone.Tags, product.Tags)
	}
	return clone
}

// CloneTestOrder creates a deep copy of a test order payload for modification.
func CloneTestOrder(order TestOrderPayload) TestOrderPayload {
	clone := order
	if order.Items != nil {
		clone.Items = make([]TestOrderItem, len(order.Items))
		copy(clone.Items, order.Items)
	}
	if order.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range order.Metadata {
			clone.Metadata[k] = v
		}
	}
	return clone
}
