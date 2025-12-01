package errors

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ContextualError represents a framework runtime error with rich contextual information
type ContextualError struct {
	// OriginalError is the underlying error that occurred
	OriginalError error

	// ServiceContext contains information about the service where the error occurred
	ServiceContext ServiceErrorContext

	// OperationContext contains information about the operation being performed
	OperationContext OperationErrorContext

	// StackTrace contains framework-aware stack trace information
	StackTrace []StackFrame

	// RecoverySuggestions provides actionable suggestions to resolve the error
	RecoverySuggestions []string

	// Timestamp when the error occurred
	Timestamp time.Time

	// RequestID for correlation across service boundaries
	RequestID string
}

// ServiceErrorContext contains metadata about the service where the error occurred
type ServiceErrorContext struct {
	// ServiceName is the name of the service (e.g., "UserService", "OrderService")
	ServiceName string

	// ServiceType indicates if this is an EndorService or EndorHybridService
	ServiceType ServiceType

	// Dependencies lists the service's dependencies that were involved
	Dependencies []string

	// ResourceType is the resource type for hybrid services (e.g., "User", "Order")
	ResourceType string

	// Configuration contains relevant service configuration
	Configuration map[string]interface{}

	// InstanceID uniquely identifies this service instance
	InstanceID string
}

// OperationErrorContext contains information about the operation being performed
type OperationErrorContext struct {
	// Operation describes the high-level operation (e.g., "CreateUser", "ListOrders")
	Operation string

	// OperationType categorizes the operation (CRUD, Middleware, Lifecycle, etc.)
	OperationType OperationType

	// HTTPMethod for HTTP-related operations
	HTTPMethod string

	// HTTPPath for HTTP-related operations
	HTTPPath string

	// Parameters contains operation parameters (sanitized for security)
	Parameters map[string]interface{}

	// Phase indicates which phase of the operation failed
	Phase OperationPhase

	// Duration shows how long the operation ran before failing
	Duration time.Duration

	// RetryAttempt indicates if this was a retry attempt
	RetryAttempt int
}

// StackFrame represents a single frame in a framework-aware stack trace
type StackFrame struct {
	// Function is the function name
	Function string

	// File is the source file
	File string

	// Line is the line number
	Line int

	// FrameworkComponent indicates if this is a framework component
	FrameworkComponent bool

	// ComponentType describes the framework component type
	ComponentType string

	// UserCode indicates if this is user/application code
	UserCode bool
}

// ServiceType represents the type of service
type ServiceType string

const (
	ServiceTypeEndor       ServiceType = "EndorService"
	ServiceTypeHybrid      ServiceType = "EndorHybridService"
	ServiceTypeComposition ServiceType = "ComposedService"
	ServiceTypeMiddleware  ServiceType = "MiddlewareService"
	ServiceTypeUnknown     ServiceType = "Unknown"
)

// OperationType represents the category of operation
type OperationType string

const (
	OperationTypeCRUD       OperationType = "CRUD"
	OperationTypeMiddleware OperationType = "Middleware"
	OperationTypeLifecycle  OperationType = "Lifecycle"
	OperationTypeValidation OperationType = "Validation"
	OperationTypeSchema     OperationType = "Schema"
	OperationTypeHTTP       OperationType = "HTTP"
	OperationTypeDatabase   OperationType = "Database"
	OperationTypeConfig     OperationType = "Configuration"
	OperationTypeUnknown    OperationType = "Unknown"
)

// OperationPhase represents the phase where the operation failed
type OperationPhase string

const (
	PhaseInitialization OperationPhase = "Initialization"
	PhaseValidation     OperationPhase = "Validation"
	PhaseExecution      OperationPhase = "Execution"
	PhasePostProcessing OperationPhase = "PostProcessing"
	PhaseCleanup        OperationPhase = "Cleanup"
	PhaseUnknown        OperationPhase = "Unknown"
)

// Error implements the error interface
func (e *ContextualError) Error() string {
	var builder strings.Builder

	// Primary error message
	builder.WriteString(fmt.Sprintf("Framework runtime error in %s.%s: %s",
		e.ServiceContext.ServiceName, e.OperationContext.Operation, e.OriginalError.Error()))

	// Add service context
	builder.WriteString(fmt.Sprintf("\nService: %s (%s)",
		e.ServiceContext.ServiceName, e.ServiceContext.ServiceType))

	if e.ServiceContext.ResourceType != "" {
		builder.WriteString(fmt.Sprintf(" [Resource: %s]", e.ServiceContext.ResourceType))
	}

	// Add operation context
	builder.WriteString(fmt.Sprintf("\nOperation: %s (%s, Phase: %s)",
		e.OperationContext.Operation, e.OperationContext.OperationType, e.OperationContext.Phase))

	if e.OperationContext.HTTPMethod != "" {
		builder.WriteString(fmt.Sprintf("\nHTTP: %s %s",
			e.OperationContext.HTTPMethod, e.OperationContext.HTTPPath))
	}

	if e.OperationContext.Duration > 0 {
		builder.WriteString(fmt.Sprintf("\nDuration: %v", e.OperationContext.Duration))
	}

	// Add framework stack trace (only user code and framework boundaries)
	if len(e.StackTrace) > 0 {
		builder.WriteString("\nStack trace:")
		for _, frame := range e.StackTrace {
			if frame.UserCode || frame.FrameworkComponent {
				prefix := "  "
				if frame.FrameworkComponent {
					prefix = fmt.Sprintf("  [%s] ", frame.ComponentType)
				}
				builder.WriteString(fmt.Sprintf("\n%s%s (%s:%d)",
					prefix, frame.Function, frame.File, frame.Line))
			}
		}
	}

	// Add recovery suggestions
	if len(e.RecoverySuggestions) > 0 {
		builder.WriteString("\nRecovery suggestions:")
		for _, suggestion := range e.RecoverySuggestions {
			builder.WriteString(fmt.Sprintf("\n  - %s", suggestion))
		}
	}

	return builder.String()
}

// Unwrap returns the original error for error unwrapping
func (e *ContextualError) Unwrap() error {
	return e.OriginalError
}

// ErrorContextBuilder builds contextual errors with a fluent interface
type ErrorContextBuilder struct {
	contextualError *ContextualError
}

// NewContextualError creates a new contextual error builder
func NewContextualError(originalError error) *ErrorContextBuilder {
	return &ErrorContextBuilder{
		contextualError: &ContextualError{
			OriginalError:       originalError,
			Timestamp:           time.Now(),
			RecoverySuggestions: make([]string, 0),
		},
	}
}

// WithService adds service context information
func (b *ErrorContextBuilder) WithService(name string, serviceType ServiceType) *ErrorContextBuilder {
	b.contextualError.ServiceContext.ServiceName = name
	b.contextualError.ServiceContext.ServiceType = serviceType
	return b
}

// WithServiceInstance adds service instance information
func (b *ErrorContextBuilder) WithServiceInstance(instanceID string, dependencies []string) *ErrorContextBuilder {
	b.contextualError.ServiceContext.InstanceID = instanceID
	b.contextualError.ServiceContext.Dependencies = dependencies
	return b
}

// WithResourceType adds resource type for hybrid services
func (b *ErrorContextBuilder) WithResourceType(resourceType string) *ErrorContextBuilder {
	b.contextualError.ServiceContext.ResourceType = resourceType
	return b
}

// WithConfiguration adds relevant service configuration
func (b *ErrorContextBuilder) WithConfiguration(config map[string]interface{}) *ErrorContextBuilder {
	b.contextualError.ServiceContext.Configuration = config
	return b
}

// WithOperation adds operation context
func (b *ErrorContextBuilder) WithOperation(operation string, operationType OperationType) *ErrorContextBuilder {
	b.contextualError.OperationContext.Operation = operation
	b.contextualError.OperationContext.OperationType = operationType
	return b
}

// WithOperationPhase adds the phase where the operation failed
func (b *ErrorContextBuilder) WithOperationPhase(phase OperationPhase) *ErrorContextBuilder {
	b.contextualError.OperationContext.Phase = phase
	return b
}

// WithHTTPContext adds HTTP-specific context
func (b *ErrorContextBuilder) WithHTTPContext(method, path string) *ErrorContextBuilder {
	b.contextualError.OperationContext.HTTPMethod = method
	b.contextualError.OperationContext.HTTPPath = path
	return b
}

// WithParameters adds operation parameters (be careful with sensitive data)
func (b *ErrorContextBuilder) WithParameters(params map[string]interface{}) *ErrorContextBuilder {
	b.contextualError.OperationContext.Parameters = params
	return b
}

// WithDuration adds the duration the operation ran before failing
func (b *ErrorContextBuilder) WithDuration(duration time.Duration) *ErrorContextBuilder {
	b.contextualError.OperationContext.Duration = duration
	return b
}

// WithRequestID adds a request ID for correlation
func (b *ErrorContextBuilder) WithRequestID(requestID string) *ErrorContextBuilder {
	b.contextualError.RequestID = requestID
	return b
}

// WithRetryAttempt indicates this was a retry attempt
func (b *ErrorContextBuilder) WithRetryAttempt(attempt int) *ErrorContextBuilder {
	b.contextualError.OperationContext.RetryAttempt = attempt
	return b
}

// WithStackTrace captures and annotates the current stack trace
func (b *ErrorContextBuilder) WithStackTrace() *ErrorContextBuilder {
	b.contextualError.StackTrace = captureFrameworkAwareStackTrace()
	return b
}

// WithRecoverySuggestions adds recovery suggestions
func (b *ErrorContextBuilder) WithRecoverySuggestions(suggestions []string) *ErrorContextBuilder {
	b.contextualError.RecoverySuggestions = suggestions
	return b
}

// Build creates the final contextual error with recovery suggestions
func (b *ErrorContextBuilder) Build() *ContextualError {
	// Generate recovery suggestions based on error patterns
	b.contextualError.RecoverySuggestions = append(
		b.contextualError.RecoverySuggestions,
		generateRecoverySuggestions(b.contextualError)...)

	return b.contextualError
}

// captureFrameworkAwareStackTrace captures stack trace with framework component annotation
func captureFrameworkAwareStackTrace() []StackFrame {
	var frames []StackFrame

	// Skip the first frame (this function itself)
	skip := 1
	for i := skip; i < skip+20; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		funcName := runtime.FuncForPC(pc).Name()

		frame := StackFrame{
			Function: funcName,
			File:     file,
			Line:     line,
		}

		// Annotate framework components
		frame.FrameworkComponent, frame.ComponentType = isFrameworkComponent(funcName, file)
		frame.UserCode = isUserCode(file)

		frames = append(frames, frame)

		// Stop when we've collected enough relevant frames
		if len(frames) >= 15 {
			break
		}
	}

	return frames
} // isFrameworkComponent determines if a function is part of the framework
func isFrameworkComponent(funcName, file string) (bool, string) {
	// Check for endor-sdk-go framework components
	if strings.Contains(file, "endor-sdk-go/sdk") {
		if strings.Contains(funcName, ".EndorService") || strings.Contains(funcName, ".endorService") {
			return true, "EndorService"
		}
		if strings.Contains(funcName, ".EndorHybridService") || strings.Contains(funcName, ".endorHybridService") {
			return true, "EndorHybridService"
		}
		if strings.Contains(file, "/middleware/") {
			return true, "Middleware"
		}
		if strings.Contains(file, "/di/") {
			return true, "DependencyInjection"
		}
		if strings.Contains(file, "/validation/") {
			return true, "Validation"
		}
		if strings.Contains(file, "/repository/") {
			return true, "Repository"
		}
		if strings.Contains(file, "/composition/") {
			return true, "ServiceComposition"
		}
		return true, "Framework"
	}

	// Check for other framework components
	if strings.Contains(file, "/gin-gonic/gin/") {
		return true, "HTTP"
	}
	if strings.Contains(file, "/mongo-driver/") {
		return true, "Database"
	}

	return false, ""
}

// isUserCode determines if a file is user/application code
func isUserCode(file string) bool {
	// User code is typically not in vendor directories or standard library
	return !strings.Contains(file, "/vendor/") &&
		!strings.Contains(file, "/usr/local/go/") &&
		!strings.Contains(file, "/go/pkg/mod/") &&
		!strings.Contains(file, "endor-sdk-go/sdk")
}

// generateRecoverySuggestions generates recovery suggestions based on error patterns
func generateRecoverySuggestions(err *ContextualError) []string {
	var suggestions []string

	errorMsg := strings.ToLower(err.OriginalError.Error())

	// Database-related errors
	if strings.Contains(errorMsg, "connection") && err.OperationContext.OperationType == OperationTypeDatabase {
		suggestions = append(suggestions, []string{
			"Check database connection configuration",
			"Verify database server is running and accessible",
			"Review connection timeout settings",
		}...)
	}

	// Validation errors
	if err.OperationContext.OperationType == OperationTypeValidation {
		suggestions = append(suggestions, []string{
			"Review input data format and constraints",
			"Check validation rules configuration",
			"Verify required fields are provided",
		}...)
	}

	// HTTP errors
	if err.OperationContext.OperationType == OperationTypeHTTP {
		if strings.Contains(errorMsg, "404") || strings.Contains(errorMsg, "not found") {
			suggestions = append(suggestions, []string{
				"Verify the endpoint URL is correct",
				"Check if the resource exists",
				"Review route registration",
			}...)
		} else if strings.Contains(errorMsg, "400") || strings.Contains(errorMsg, "bad request") {
			suggestions = append(suggestions, []string{
				"Validate request payload format",
				"Check required headers and parameters",
				"Review API documentation for correct request format",
			}...)
		}
	}

	// Configuration errors
	if err.OperationContext.OperationType == OperationTypeConfig {
		suggestions = append(suggestions, []string{
			"Check environment variables are set correctly",
			"Verify configuration file format",
			"Review default values and required fields",
		}...)
	}

	// Dependency injection errors
	if strings.Contains(errorMsg, "dependency") || strings.Contains(errorMsg, "inject") {
		suggestions = append(suggestions, []string{
			"Verify all required dependencies are registered",
			"Check dependency registration order",
			"Review service construction parameters",
		}...)
	}

	// Middleware errors
	if err.OperationContext.OperationType == OperationTypeMiddleware {
		suggestions = append(suggestions, []string{
			"Check middleware configuration and order",
			"Verify middleware dependencies are available",
			"Review middleware execution chain",
		}...)
	}

	// Generic suggestions based on operation phase
	switch err.OperationContext.Phase {
	case PhaseInitialization:
		suggestions = append(suggestions, "Review service initialization and dependencies")
	case PhaseValidation:
		suggestions = append(suggestions, "Check input validation rules and data format")
	case PhaseExecution:
		suggestions = append(suggestions, "Review business logic and external service dependencies")
	case PhasePostProcessing:
		suggestions = append(suggestions, "Check post-processing steps and data transformation")
	case PhaseCleanup:
		suggestions = append(suggestions, "Review resource cleanup and disposal logic")
	}

	// Add monitoring suggestion
	suggestions = append(suggestions, "Check application logs and monitoring dashboards for additional context")

	return suggestions
}

// Helper functions for extracting context from various sources

// ExtractServiceContextFromContext extracts service context from a Go context
func ExtractServiceContextFromContext(ctx context.Context) ServiceErrorContext {
	serviceCtx := ServiceErrorContext{}

	// Try to extract service information from context values
	if serviceName, ok := ctx.Value("serviceName").(string); ok {
		serviceCtx.ServiceName = serviceName
	}
	if serviceType, ok := ctx.Value("serviceType").(ServiceType); ok {
		serviceCtx.ServiceType = serviceType
	}
	if resourceType, ok := ctx.Value("resourceType").(string); ok {
		serviceCtx.ResourceType = resourceType
	}
	if instanceID, ok := ctx.Value("instanceID").(string); ok {
		serviceCtx.InstanceID = instanceID
	}

	return serviceCtx
}

// ExtractOperationContextFromContext extracts operation context from a Go context
func ExtractOperationContextFromContext(ctx context.Context) OperationErrorContext {
	opCtx := OperationErrorContext{}

	// Try to extract operation information from context values
	if operation, ok := ctx.Value("operation").(string); ok {
		opCtx.Operation = operation
	}
	if operationType, ok := ctx.Value("operationType").(OperationType); ok {
		opCtx.OperationType = operationType
	}
	if httpMethod, ok := ctx.Value("httpMethod").(string); ok {
		opCtx.HTTPMethod = httpMethod
	}
	if httpPath, ok := ctx.Value("httpPath").(string); ok {
		opCtx.HTTPPath = httpPath
	}
	if phase, ok := ctx.Value("operationPhase").(OperationPhase); ok {
		opCtx.Phase = phase
	}

	return opCtx
}
