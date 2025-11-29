package middleware

import (
	"fmt"
	"reflect"
	"time"
)

// MiddlewareInterface defines the contract for middleware components that can intercept
// and modify service requests and responses. Middleware enables cross-cutting concerns
// like authentication, logging, and metrics without modifying core service logic.
//
// Before is called before the service handler executes and can modify the request context.
// It can return an error to short-circuit execution and prevent the handler from running.
//
// After is called after the service handler executes and can inspect or modify the response.
// It can return an error to indicate middleware processing failures, but the original
// handler result is preserved for consistency.
//
// Example implementation:
//
//	type LoggingMiddleware struct {
//	    Logger interfaces.LoggerInterface
//	}
//
//	func (l *LoggingMiddleware) Before(ctx *sdk.EndorContext[any]) error {
//	    l.Logger.Info("Request started", "resource", ctx.GetMicroServiceId())
//	    return nil
//	}
//
//	func (l *LoggingMiddleware) After(ctx *sdk.EndorContext[any], response *sdk.Response[any]) error {
//	    l.Logger.Info("Request completed", "resource", ctx.GetMicroServiceId())
//	    return nil
//	}
type MiddlewareInterface interface {
	// Before is called before the service handler executes.
	// It receives the request context and can modify it or terminate the request early.
	// Returning an error will short-circuit the middleware pipeline and prevent
	// the service handler from executing.
	Before(ctx interface{}) error

	// After is called after the service handler executes.
	// It receives both the request context and the response for inspection or modification.
	// Returning an error indicates middleware processing failure but preserves the original response.
	After(ctx interface{}, response interface{}) error
}

// MiddlewareExecution tracks middleware execution details for performance monitoring.
type MiddlewareExecution struct {
	Name           string        // Middleware name for identification
	BeforeDuration time.Duration // Time taken by Before() method
	AfterDuration  time.Duration // Time taken by After() method
	Success        bool          // Whether middleware executed without errors
	Error          error         // Error if middleware failed
}

// MiddlewarePipeline manages the execution of a middleware chain.
// It handles execution order, short-circuiting, and performance tracking.
type MiddlewarePipeline struct {
	middlewares []MiddlewareInterface
	executions  []MiddlewareExecution // Performance tracking data
}

// NewMiddlewarePipeline creates a new middleware pipeline with the given middleware components.
// Middleware will be executed in the order provided.
func NewMiddlewarePipeline(middlewares ...MiddlewareInterface) *MiddlewarePipeline {
	return &MiddlewarePipeline{
		middlewares: middlewares,
		executions:  make([]MiddlewareExecution, 0, len(middlewares)),
	}
}

// ExecuteBefore runs the Before() method of all middleware in order.
// If any middleware returns an error, execution stops and the error is returned.
// This implements the short-circuiting behavior required for auth failures and validation.
func (p *MiddlewarePipeline) ExecuteBefore(ctx interface{}) error {
	p.executions = p.executions[:0] // Reset execution tracking

	for i, middleware := range p.middlewares {
		start := time.Now()
		execution := MiddlewareExecution{
			Name:    middlewareName(middleware, i),
			Success: true,
		}

		err := middleware.Before(ctx)
		execution.BeforeDuration = time.Since(start)

		if err != nil {
			execution.Success = false
			execution.Error = err
			p.executions = append(p.executions, execution)
			return NewMiddlewareError("before", execution.Name, err)
		}

		p.executions = append(p.executions, execution)
	}

	return nil
}

// ExecuteAfter runs the After() method of all middleware in reverse order.
// This ensures proper cleanup and response processing symmetry.
// All After() methods are called even if some return errors, but the first error is returned.
func (p *MiddlewarePipeline) ExecuteAfter(ctx interface{}, response interface{}) error {
	var firstError error

	// Execute in reverse order for symmetric cleanup
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		middleware := p.middlewares[i]
		start := time.Now()

		err := middleware.After(ctx, response)

		// Update execution tracking
		if i < len(p.executions) {
			p.executions[i].AfterDuration = time.Since(start)
			if err != nil && p.executions[i].Success {
				p.executions[i].Success = false
				p.executions[i].Error = err
			}
		}

		// Capture first error but continue processing
		if err != nil && firstError == nil {
			firstError = NewMiddlewareError("after", middlewareName(middleware, i), err)
		}
	}

	return firstError
}

// GetExecutions returns the performance tracking data for the last pipeline execution.
// This enables performance monitoring and debugging of middleware overhead.
func (p *MiddlewarePipeline) GetExecutions() []MiddlewareExecution {
	// Return copy to prevent external modification
	result := make([]MiddlewareExecution, len(p.executions))
	copy(result, p.executions)
	return result
}

// GetTotalDuration returns the total execution time for all middleware components.
func (p *MiddlewarePipeline) GetTotalDuration() time.Duration {
	var total time.Duration
	for _, exec := range p.executions {
		total += exec.BeforeDuration + exec.AfterDuration
	}
	return total
}

// middlewareName extracts a human-readable name from middleware for tracking.
// Falls back to index-based naming if type name extraction fails.
func middlewareName(middleware MiddlewareInterface, index int) string {
	// Use type name for identification
	// This could be enhanced with a NamedMiddleware interface if needed
	return fmt.Sprintf("%T", middleware)
}

// middlewareContext adapts request context for middleware interface compatibility.
// This abstracts the underlying HTTP framework (Gin) from middleware implementations.
type middlewareContext struct {
	ginContext interface{} // *gin.Context stored as interface{} for flexibility
}

// middlewareResponse wraps response data for middleware interface compatibility.
type middlewareResponse struct {
	statusCode int
	headers    map[string][]string
	data       interface{}
}

// Helper functions for framework-agnostic middleware implementations
// These use reflection to work with different HTTP framework contexts

// extractHeaders extracts HTTP headers from the context using reflection
func extractHeaders(ctx interface{}) map[string]string {
	if ctx == nil {
		return nil
	}

	// Try to access headers using reflection for gin.Context
	v := reflect.ValueOf(ctx)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Look for Request field (gin.Context.Request)
	requestField := v.FieldByName("Request")
	if !requestField.IsValid() {
		return nil
	}

	// Get Header from Request
	headerField := requestField.FieldByName("Header")
	if !headerField.IsValid() {
		return nil
	}

	// Convert http.Header to simple map[string]string
	headers := make(map[string]string)
	headerInterface := headerField.Interface()
	if headerMap, ok := headerInterface.(map[string][]string); ok {
		for k, v := range headerMap {
			if len(v) > 0 {
				headers[k] = v[0] // Take first value
			}
		}
	}

	return headers
}

// getHeader gets a header value from the headers map
func getHeader(headers map[string]string, key string) string {
	if headers == nil {
		return ""
	}
	return headers[key]
}

// setContextValue sets a value in the context using reflection
func setContextValue(ctx interface{}, key string, value interface{}) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	// Use reflection to call Set method if available (gin.Context.Set)
	v := reflect.ValueOf(ctx)
	setMethod := v.MethodByName("Set")
	if !setMethod.IsValid() {
		return fmt.Errorf("context does not have Set method")
	}

	// Call Set(key, value)
	args := []reflect.Value{
		reflect.ValueOf(key),
		reflect.ValueOf(value),
	}

	setMethod.Call(args)
	return nil
}

// MiddlewareError represents errors that occur during middleware execution.
type MiddlewareError struct {
	Phase      string // "before" or "after"
	Middleware string // Middleware name
	Cause      error  // Original error
}

func (e *MiddlewareError) Error() string {
	return fmt.Sprintf("middleware error in %s phase of %s: %v", e.Phase, e.Middleware, e.Cause)
}

func (e *MiddlewareError) Unwrap() error {
	return e.Cause
}

// NewMiddlewareError creates a new middleware execution error.
func NewMiddlewareError(phase, middleware string, cause error) *MiddlewareError {
	return &MiddlewareError{
		Phase:      phase,
		Middleware: middleware,
		Cause:      cause,
	}
}
