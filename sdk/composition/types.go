package composition

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// CompositionPattern defines the core interface for all service composition utilities.
// It provides a unified contract for executing, validating, and introspecting composed services.
type CompositionPattern interface {
	// Execute runs the composed services with the given context and request.
	// Returns the result of the composition or an error if execution fails.
	Execute(ctx context.Context, request interface{}) (interface{}, error)

	// Validate checks that the composition is properly configured and all services are compatible.
	// Returns an error if validation fails with detailed information about the issue.
	Validate() error

	// GetServices returns the list of services involved in this composition for introspection.
	GetServices() []interfaces.EndorServiceInterface
}

// CompositionValidator provides validation capabilities for service compositions.
// It ensures type safety and interface compatibility across service boundaries.
type CompositionValidator interface {
	// ValidateChain verifies that services in a chain have compatible input/output types.
	ValidateChain(services []interfaces.EndorServiceInterface) error

	// ValidateTypes ensures that the input type is compatible with the service interfaces.
	ValidateTypes(inputType reflect.Type, services []interfaces.EndorServiceInterface) error

	// AnalyzeDependencyGraph analyzes the composition for circular dependencies and ordering issues.
	AnalyzeDependencyGraph(composition CompositionPattern) ([]string, error)
}

// ProxyInterceptor defines the interface for request/response interception in service proxies.
// It enables transformation of requests and responses as they flow through proxy layers.
type ProxyInterceptor interface {
	// BeforeRequest is called before the request is forwarded to the target service.
	// It can modify the request or return an error to terminate the proxy operation.
	BeforeRequest(ctx context.Context, request interface{}) (interface{}, error)

	// AfterResponse is called after the response is received from the target service.
	// It can modify the response or return an error to indicate proxy processing failure.
	AfterResponse(ctx context.Context, request interface{}, response interface{}) (interface{}, error)
}

// BranchRouter defines the interface for conditional service selection in branched compositions.
// It analyzes the request context and data to determine which service should handle the request.
type BranchRouter interface {
	// SelectService analyzes the request and returns the key of the service that should handle it.
	// Returns an error if no suitable service can be determined or if routing fails.
	SelectService(ctx context.Context, request interface{}) (string, error)

	// GetDefaultService returns the key of the service to use when no specific route is found.
	// Returns empty string if no default service is configured.
	GetDefaultService() string
}

// ResultMerger defines the interface for aggregating results from multiple services.
// It provides flexible strategies for combining outputs from parallel service executions.
type ResultMerger interface {
	// MergeResults combines results from multiple service executions into a single result.
	// The inputs slice contains the results from each service in the order they were executed.
	MergeResults(ctx context.Context, request interface{}, results []interface{}) (interface{}, error)

	// ShouldWaitForAll returns true if the merger requires all services to complete before merging.
	// Returns false if the merger can work with partial results (e.g., first-wins strategy).
	ShouldWaitForAll() bool

	// GetTimeout returns the maximum time to wait for service executions.
	// Returns zero duration if no timeout is configured.
	GetTimeout() time.Duration
}

// CompositionError represents errors that occur during service composition operations.
// It provides detailed context about which service failed and why.
type CompositionError struct {
	Operation    string                 // The composition operation that failed (chain, proxy, branch, merge)
	ServiceName  string                 // The name of the service that caused the error
	ServiceIndex int                    // The index of the service in the composition (if applicable)
	Cause        error                  // The underlying error that caused the failure
	Context      map[string]interface{} // Additional context about the failure
}

func (e *CompositionError) Error() string {
	if e.ServiceName != "" {
		return fmt.Sprintf("composition %s failed at service '%s' (index %d): %v",
			e.Operation, e.ServiceName, e.ServiceIndex, e.Cause)
	}
	return fmt.Sprintf("composition %s failed: %v", e.Operation, e.Cause)
}

func (e *CompositionError) Unwrap() error {
	return e.Cause
}

// NewCompositionError creates a new composition error with the specified details.
func NewCompositionError(operation, serviceName string, serviceIndex int, cause error) *CompositionError {
	return &CompositionError{
		Operation:    operation,
		ServiceName:  serviceName,
		ServiceIndex: serviceIndex,
		Cause:        cause,
		Context:      make(map[string]interface{}),
	}
}

// WithContext adds contextual information to the composition error.
func (e *CompositionError) WithContext(key string, value interface{}) *CompositionError {
	e.Context[key] = value
	return e
}

// CompositionConfig provides configuration options for composition patterns.
type CompositionConfig struct {
	// Timeout specifies the maximum time to wait for composition execution.
	Timeout time.Duration

	// MaxConcurrency limits the number of concurrent service executions (for mergers).
	MaxConcurrency int

	// ErrorHandling specifies how errors should be handled (fail-fast, continue-on-error, etc.).
	ErrorHandling ErrorHandlingStrategy

	// EnableMetrics indicates whether to collect performance metrics for the composition.
	EnableMetrics bool

	// ValidationMode specifies how strict the composition validation should be.
	ValidationMode ValidationMode
}

// ErrorHandlingStrategy defines how compositions should handle service failures.
type ErrorHandlingStrategy int

const (
	// FailFast terminates the composition immediately when any service fails.
	FailFast ErrorHandlingStrategy = iota

	// ContinueOnError continues execution even when services fail, collecting errors for final report.
	ContinueOnError

	// RetryOnError attempts to retry failed services before giving up.
	RetryOnError
)

// ValidationMode defines how strict composition validation should be.
type ValidationMode int

const (
	// Strict requires full type compatibility and interface validation.
	Strict ValidationMode = iota

	// Lenient allows compositions with loose type checking.
	Lenient

	// Disabled skips validation entirely for maximum performance.
	Disabled
)

// defaultValidator provides a basic implementation of CompositionValidator.
type defaultValidator struct{}

// ValidateChain validates that services in a chain have compatible interfaces.
func (v *defaultValidator) ValidateChain(services []interfaces.EndorServiceInterface) error {
	if len(services) == 0 {
		return fmt.Errorf("cannot validate empty service chain")
	}

	for i, service := range services {
		if service == nil {
			return fmt.Errorf("service at index %d is nil", i)
		}

		if err := service.Validate(); err != nil {
			return fmt.Errorf("service at index %d failed validation: %w", i, err)
		}
	}

	return nil
}

// ValidateTypes ensures input types are compatible with service requirements.
func (v *defaultValidator) ValidateTypes(inputType reflect.Type, services []interfaces.EndorServiceInterface) error {
	// Basic validation - check that we have services and a valid input type
	if inputType == nil {
		return fmt.Errorf("input type cannot be nil")
	}

	if len(services) == 0 {
		return fmt.Errorf("cannot validate types against empty service list")
	}

	// For now, we perform basic validation. More sophisticated type checking
	// would require reflection analysis of service handler signatures.
	return nil
}

// AnalyzeDependencyGraph checks for circular dependencies in compositions.
func (v *defaultValidator) AnalyzeDependencyGraph(composition CompositionPattern) ([]string, error) {
	services := composition.GetServices()
	warnings := make([]string, 0)

	// Basic analysis - check for duplicate services which could indicate circular references
	serviceNames := make(map[string]int)
	for i, service := range services {
		name := service.GetResource()
		if prevIndex, exists := serviceNames[name]; exists {
			warnings = append(warnings, fmt.Sprintf("service '%s' appears multiple times (indices %d and %d)", name, prevIndex, i))
		}
		serviceNames[name] = i
	}

	return warnings, nil
}

// DefaultValidator returns the default composition validator instance.
func DefaultValidator() CompositionValidator {
	return &defaultValidator{}
}
