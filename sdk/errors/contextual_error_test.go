package errors

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextualError_Error(t *testing.T) {
	tests := []struct {
		name     string
		setupErr func() *ContextualError
		contains []string
	}{
		{
			name: "basic contextual error",
			setupErr: func() *ContextualError {
				return NewContextualError(fmt.Errorf("database connection failed")).
					WithService("UserService", ServiceTypeEndor).
					WithOperation("CreateUser", OperationTypeCRUD).
					WithOperationPhase(PhaseExecution).
					Build()
			},
			contains: []string{
				"Framework runtime error in UserService.CreateUser",
				"database connection failed",
				"Service: UserService (EndorService)",
				"Operation: CreateUser (CRUD, Phase: Execution)",
			},
		},
		{
			name: "hybrid service with resource type",
			setupErr: func() *ContextualError {
				return NewContextualError(fmt.Errorf("validation failed")).
					WithService("OrderService", ServiceTypeHybrid).
					WithResourceType("Order").
					WithOperation("UpdateOrder", OperationTypeValidation).
					WithOperationPhase(PhaseValidation).
					Build()
			},
			contains: []string{
				"Framework runtime error in OrderService.UpdateOrder",
				"validation failed",
				"Service: OrderService (EndorHybridService) [Resource: Order]",
				"Operation: UpdateOrder (Validation, Phase: Validation)",
			},
		},
		{
			name: "HTTP operation with context",
			setupErr: func() *ContextualError {
				return NewContextualError(fmt.Errorf("request timeout")).
					WithService("APIGateway", ServiceTypeMiddleware).
					WithOperation("ProcessRequest", OperationTypeHTTP).
					WithHTTPContext("POST", "/api/users").
					WithDuration(5 * time.Second).
					WithOperationPhase(PhaseExecution).
					Build()
			},
			contains: []string{
				"Framework runtime error in APIGateway.ProcessRequest",
				"request timeout",
				"Service: APIGateway (MiddlewareService)",
				"Operation: ProcessRequest (HTTP, Phase: Execution)",
				"HTTP: POST /api/users",
				"Duration: 5s",
			},
		},
		{
			name: "with recovery suggestions",
			setupErr: func() *ContextualError {
				return NewContextualError(fmt.Errorf("connection refused")).
					WithService("DatabaseService", ServiceTypeEndor).
					WithOperation("Connect", OperationTypeDatabase).
					WithOperationPhase(PhaseInitialization).
					WithRecoverySuggestions([]string{
						"Check database server status",
						"Verify connection string",
					}).
					Build()
			},
			contains: []string{
				"Recovery suggestions:",
				"- Check database server status",
				"- Verify connection string",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setupErr()
			errorStr := err.Error()

			for _, expected := range tt.contains {
				assert.Contains(t, errorStr, expected,
					"Expected error message to contain: %s\nActual error: %s", expected, errorStr)
			}
		})
	}
}

func TestContextualError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	contextualErr := NewContextualError(originalErr).
		WithService("TestService", ServiceTypeEndor).
		WithOperation("TestOperation", OperationTypeCRUD).
		Build()

	assert.Equal(t, originalErr, errors.Unwrap(contextualErr))
}

func TestErrorContextBuilder_FluentInterface(t *testing.T) {
	originalErr := fmt.Errorf("test error")

	// Test that all methods return the builder for chaining
	builder := NewContextualError(originalErr).
		WithService("TestService", ServiceTypeHybrid).
		WithServiceInstance("instance-123", []string{"dep1", "dep2"}).
		WithResourceType("TestResource").
		WithConfiguration(map[string]interface{}{
			"timeout": "30s",
			"retries": 3,
		}).
		WithOperation("TestOperation", OperationTypeCRUD).
		WithOperationPhase(PhaseValidation).
		WithHTTPContext("POST", "/test").
		WithParameters(map[string]interface{}{
			"id": "123",
		}).
		WithDuration(100 * time.Millisecond).
		WithRequestID("req-456").
		WithRetryAttempt(2).
		WithStackTrace().
		WithRecoverySuggestions([]string{"Try again"})

	contextualErr := builder.Build()

	// Verify all fields are set
	assert.Equal(t, originalErr, contextualErr.OriginalError)
	assert.Equal(t, "TestService", contextualErr.ServiceContext.ServiceName)
	assert.Equal(t, ServiceTypeHybrid, contextualErr.ServiceContext.ServiceType)
	assert.Equal(t, "instance-123", contextualErr.ServiceContext.InstanceID)
	assert.Equal(t, []string{"dep1", "dep2"}, contextualErr.ServiceContext.Dependencies)
	assert.Equal(t, "TestResource", contextualErr.ServiceContext.ResourceType)
	assert.Equal(t, "TestOperation", contextualErr.OperationContext.Operation)
	assert.Equal(t, OperationTypeCRUD, contextualErr.OperationContext.OperationType)
	assert.Equal(t, PhaseValidation, contextualErr.OperationContext.Phase)
	assert.Equal(t, "POST", contextualErr.OperationContext.HTTPMethod)
	assert.Equal(t, "/test", contextualErr.OperationContext.HTTPPath)
	assert.Equal(t, 100*time.Millisecond, contextualErr.OperationContext.Duration)
	assert.Equal(t, "req-456", contextualErr.RequestID)
	assert.Equal(t, 2, contextualErr.OperationContext.RetryAttempt)
	assert.NotEmpty(t, contextualErr.StackTrace)
	assert.Contains(t, contextualErr.RecoverySuggestions, "Try again")
	assert.NotZero(t, contextualErr.Timestamp)
}

func TestGenerateRecoverySuggestions(t *testing.T) {
	tests := []struct {
		name                string
		errorMsg            string
		operationType       OperationType
		operationPhase      OperationPhase
		expectedSuggestions []string
	}{
		{
			name:          "database connection error",
			errorMsg:      "connection refused",
			operationType: OperationTypeDatabase,
			expectedSuggestions: []string{
				"Check database connection configuration",
				"Verify database server is running and accessible",
				"Review connection timeout settings",
			},
		},
		{
			name:          "validation error",
			errorMsg:      "field validation failed",
			operationType: OperationTypeValidation,
			expectedSuggestions: []string{
				"Review input data format and constraints",
				"Check validation rules configuration",
				"Verify required fields are provided",
			},
		},
		{
			name:          "HTTP 404 error",
			errorMsg:      "404 not found",
			operationType: OperationTypeHTTP,
			expectedSuggestions: []string{
				"Verify the endpoint URL is correct",
				"Check if the resource exists",
				"Review route registration",
			},
		},
		{
			name:          "HTTP 400 error",
			errorMsg:      "400 bad request",
			operationType: OperationTypeHTTP,
			expectedSuggestions: []string{
				"Validate request payload format",
				"Check required headers and parameters",
				"Review API documentation for correct request format",
			},
		},
		{
			name:          "configuration error",
			errorMsg:      "config not found",
			operationType: OperationTypeConfig,
			expectedSuggestions: []string{
				"Check environment variables are set correctly",
				"Verify configuration file format",
				"Review default values and required fields",
			},
		},
		{
			name:          "dependency injection error",
			errorMsg:      "dependency not found",
			operationType: OperationTypeUnknown,
			expectedSuggestions: []string{
				"Verify all required dependencies are registered",
				"Check dependency registration order",
				"Review service construction parameters",
			},
		},
		{
			name:          "middleware error",
			errorMsg:      "middleware failed",
			operationType: OperationTypeMiddleware,
			expectedSuggestions: []string{
				"Check middleware configuration and order",
				"Verify middleware dependencies are available",
				"Review middleware execution chain",
			},
		},
		{
			name:           "initialization phase error",
			errorMsg:       "initialization failed",
			operationType:  OperationTypeUnknown,
			operationPhase: PhaseInitialization,
			expectedSuggestions: []string{
				"Review service initialization and dependencies",
			},
		},
		{
			name:           "validation phase error",
			errorMsg:       "validation failed",
			operationType:  OperationTypeUnknown,
			operationPhase: PhaseValidation,
			expectedSuggestions: []string{
				"Check input validation rules and data format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextualErr := &ContextualError{
				OriginalError: fmt.Errorf(tt.errorMsg),
				OperationContext: OperationErrorContext{
					OperationType: tt.operationType,
					Phase:         tt.operationPhase,
				},
			}

			suggestions := generateRecoverySuggestions(contextualErr)

			for _, expected := range tt.expectedSuggestions {
				assert.Contains(t, suggestions, expected,
					"Expected suggestions to contain: %s\nActual suggestions: %v", expected, suggestions)
			}

			// All suggestions should include the monitoring suggestion
			assert.Contains(t, suggestions,
				"Check application logs and monitoring dashboards for additional context")
		})
	}
}

func TestIsFrameworkComponent(t *testing.T) {
	tests := []struct {
		name          string
		funcName      string
		file          string
		isFramework   bool
		componentType string
	}{
		{
			name:          "EndorService component",
			funcName:      "main.(*UserService).EndorService.CreateUser",
			file:          "/path/to/endor-sdk-go/sdk/endor_service.go",
			isFramework:   true,
			componentType: "EndorService",
		},
		{
			name:          "EndorHybridService component",
			funcName:      "main.(*OrderService).EndorHybridService.UpdateOrder",
			file:          "/path/to/endor-sdk-go/sdk/endor_hybrid_service.go",
			isFramework:   true,
			componentType: "EndorHybridService",
		},
		{
			name:          "Middleware component",
			funcName:      "main.authMiddleware",
			file:          "/path/to/endor-sdk-go/sdk/middleware/auth.go",
			isFramework:   true,
			componentType: "Middleware",
		},
		{
			name:          "DI component",
			funcName:      "main.(*Container).Resolve",
			file:          "/path/to/endor-sdk-go/sdk/di/container.go",
			isFramework:   true,
			componentType: "DependencyInjection",
		},
		{
			name:          "Validation component",
			funcName:      "main.(*ConfigValidator).Validate",
			file:          "/path/to/endor-sdk-go/sdk/validation/config_validator.go",
			isFramework:   true,
			componentType: "Validation",
		},
		{
			name:          "Repository component",
			funcName:      "main.(*MongoRepository).FindById",
			file:          "/path/to/endor-sdk-go/sdk/repository/mongo.go",
			isFramework:   true,
			componentType: "Repository",
		},
		{
			name:          "Gin HTTP framework",
			funcName:      "github.com/gin-gonic/gin.(*Engine).handleHTTPRequest",
			file:          "/path/to/gin-gonic/gin/gin.go",
			isFramework:   true,
			componentType: "HTTP",
		},
		{
			name:          "MongoDB driver",
			funcName:      "go.mongodb.org/mongo-driver/mongo.(*Collection).FindOne",
			file:          "/path/to/mongo-driver/mongo/collection.go",
			isFramework:   true,
			componentType: "Database",
		},
		{
			name:          "User code",
			funcName:      "main.(*UserController).CreateUser",
			file:          "/home/user/myapp/controllers/user_controller.go",
			isFramework:   false,
			componentType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isFramework, componentType := isFrameworkComponent(tt.funcName, tt.file)
			assert.Equal(t, tt.isFramework, isFramework)
			assert.Equal(t, tt.componentType, componentType)
		})
	}
}

func TestIsUserCode(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		expected bool
	}{
		{
			name:     "user application code",
			file:     "/home/user/myapp/services/user_service.go",
			expected: true,
		},
		{
			name:     "vendor code",
			file:     "/home/user/myapp/vendor/github.com/gin-gonic/gin/gin.go",
			expected: false,
		},
		{
			name:     "Go standard library",
			file:     "/usr/local/go/src/net/http/server.go",
			expected: false,
		},
		{
			name:     "Go module cache",
			file:     "/home/user/go/pkg/mod/github.com/stretchr/testify@v1.8.4/assert/assertions.go",
			expected: false,
		},
		{
			name:     "Framework SDK code",
			file:     "/home/user/myapp/endor-sdk-go/sdk/endor_service.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUserCode(tt.file)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractContextFromContext(t *testing.T) {
	// Test extracting service context
	ctx := context.Background()
	ctx = context.WithValue(ctx, "serviceName", "TestService")
	ctx = context.WithValue(ctx, "serviceType", ServiceTypeHybrid)
	ctx = context.WithValue(ctx, "resourceType", "User")
	ctx = context.WithValue(ctx, "instanceID", "instance-123")

	serviceCtx := ExtractServiceContextFromContext(ctx)
	assert.Equal(t, "TestService", serviceCtx.ServiceName)
	assert.Equal(t, ServiceTypeHybrid, serviceCtx.ServiceType)
	assert.Equal(t, "User", serviceCtx.ResourceType)
	assert.Equal(t, "instance-123", serviceCtx.InstanceID)

	// Test extracting operation context
	ctx = context.WithValue(ctx, "operation", "CreateUser")
	ctx = context.WithValue(ctx, "operationType", OperationTypeCRUD)
	ctx = context.WithValue(ctx, "httpMethod", "POST")
	ctx = context.WithValue(ctx, "httpPath", "/api/users")
	ctx = context.WithValue(ctx, "operationPhase", PhaseValidation)

	opCtx := ExtractOperationContextFromContext(ctx)
	assert.Equal(t, "CreateUser", opCtx.Operation)
	assert.Equal(t, OperationTypeCRUD, opCtx.OperationType)
	assert.Equal(t, "POST", opCtx.HTTPMethod)
	assert.Equal(t, "/api/users", opCtx.HTTPPath)
	assert.Equal(t, PhaseValidation, opCtx.Phase)
}

func TestContextualError_WithStackTrace(t *testing.T) {
	err := NewContextualError(fmt.Errorf("test error")).
		WithService("TestService", ServiceTypeEndor).
		WithOperation("TestOperation", OperationTypeCRUD).
		WithStackTrace().
		Build()

	assert.NotEmpty(t, err.StackTrace)

	// Debug: Print the stack trace to understand what we're getting
	t.Logf("Stack trace contains %d frames:", len(err.StackTrace))
	for i, frame := range err.StackTrace {
		t.Logf("  [%d] %s (%s:%d) - Framework: %t (%s), User: %t",
			i, frame.Function, frame.File, frame.Line, frame.FrameworkComponent, frame.ComponentType, frame.UserCode)
	}

	// Check that we captured some stack frames
	assert.Greater(t, len(err.StackTrace), 0, "Should have captured some stack frames")

	// Check that at least one frame has a meaningful function name
	hasMeaningfulFrame := false
	for _, frame := range err.StackTrace {
		if frame.Function != "" && !strings.Contains(frame.Function, "runtime.") {
			hasMeaningfulFrame = true
			break
		}
	}
	assert.True(t, hasMeaningfulFrame, "Should have at least one meaningful stack frame")
}

func TestContextualError_TimestampAndRequestID(t *testing.T) {
	before := time.Now()

	err := NewContextualError(fmt.Errorf("test error")).
		WithService("TestService", ServiceTypeEndor).
		WithOperation("TestOperation", OperationTypeCRUD).
		WithRequestID("test-request-123").
		Build()

	after := time.Now()

	assert.True(t, err.Timestamp.After(before) || err.Timestamp.Equal(before))
	assert.True(t, err.Timestamp.Before(after) || err.Timestamp.Equal(after))
	assert.Equal(t, "test-request-123", err.RequestID)
}
