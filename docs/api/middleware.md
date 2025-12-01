# Middleware

> Package documentation for Middleware

**Import Path:** `github.com/mattiabonardi/endor-sdk-go/sdk/middleware`
**Generated:** 2025-12-01 10:07:53 UTC

---

type AuthMiddleware struct{ ... }
    func NewAuthMiddleware(deps AuthMiddlewareDependencies) *AuthMiddleware
type AuthMiddlewareDependencies struct{ ... }
type LoggingMiddleware struct{ ... }
    func NewLoggingMiddleware(deps LoggingMiddlewareDependencies) *LoggingMiddleware
type LoggingMiddlewareDependencies struct{ ... }
type MetricsMiddleware struct{ ... }
    func NewMetricsMiddleware(deps MetricsMiddlewareDependencies) *MetricsMiddleware
type MetricsMiddlewareDependencies struct{ ... }
type MiddlewareError struct{ ... }
    func NewMiddlewareError(phase, middleware string, cause error) *MiddlewareError
type MiddlewareExecution struct{ ... }
type MiddlewareInterface interface{ ... }
type MiddlewarePipeline struct{ ... }
    func NewMiddlewarePipeline(middlewares ...MiddlewareInterface) *MiddlewarePipeline

## Package Overview

package middleware // import "github.com/mattiabonardi/endor-sdk-go/sdk/middleware"

type AuthMiddleware struct{ ... }
    func NewAuthMiddleware(deps AuthMiddlewareDependencies) *AuthMiddleware
type AuthMiddlewareDependencies struct{ ... }
type LoggingMiddleware struct{ ... }
    func NewLoggingMiddleware(deps LoggingMiddlewareDependencies) *LoggingMiddleware
type LoggingMiddlewareDependencies struct{ ... }
type MetricsMiddleware struct{ ... }
    func NewMetricsMiddleware(deps MetricsMiddlewareDependencies) *MetricsMiddleware
type MetricsMiddlewareDependencies struct{ ... }
type MiddlewareError struct{ ... }
    func NewMiddlewareError(phase, middleware string, cause error) *MiddlewareError
type MiddlewareExecution struct{ ... }
type MiddlewareInterface interface{ ... }
type MiddlewarePipeline struct{ ... }
    func NewMiddlewarePipeline(middlewares ...MiddlewareInterface) *MiddlewarePipeline

## Exported Types

### AuthMiddleware

```go
type AuthMiddleware struct{ ... }
```


type AuthMiddleware struct {
	Config      interfaces.ConfigProviderInterface // For auth configuration
	Logger      interfaces.LoggerInterface         // For auth logging
	RequireAuth bool                               // Whether authentication is mandatory
}
    AuthMiddleware provides authentication and authorization middleware.
    It validates session tokens and ensures proper security context propagation.

func NewAuthMiddleware(deps AuthMiddlewareDependencies) *AuthMiddleware
func (a *AuthMiddleware) After(ctx interface{}, response interface{}) error
func (a *AuthMiddleware) Before(ctx interface{}) error

### AuthMiddlewareDependencies

```go
type AuthMiddlewareDependencies struct{ ... }
```


type AuthMiddlewareDependencies struct {
	Config      interfaces.ConfigProviderInterface
	Logger      interfaces.LoggerInterface
	RequireAuth bool
}
    AuthMiddlewareDependencies contains all required dependencies for
    AuthMiddleware.


### LoggingMiddleware

```go
type LoggingMiddleware struct{ ... }
```


type LoggingMiddleware struct {
	Logger           interfaces.LoggerInterface // For structured logging
	IncludeHeaders   bool                       // Whether to log request headers
	IncludePayload   bool                       // Whether to log request payload (be careful with sensitive data)
	CorrelationIDKey string                     // Header key for correlation ID (default: "x-correlation-id")
}
    LoggingMiddleware provides request/response logging with correlation ID
    support. It logs request start, completion, and performance metrics.

func NewLoggingMiddleware(deps LoggingMiddlewareDependencies) *LoggingMiddleware
func (l *LoggingMiddleware) After(ctx interface{}, response interface{}) error
func (l *LoggingMiddleware) Before(ctx interface{}) error

### LoggingMiddlewareDependencies

```go
type LoggingMiddlewareDependencies struct{ ... }
```


type LoggingMiddlewareDependencies struct {
	Logger           interfaces.LoggerInterface
	IncludeHeaders   bool
	IncludePayload   bool
	CorrelationIDKey string // Optional, defaults to "x-correlation-id"
}
    LoggingMiddlewareDependencies contains all required dependencies for
    LoggingMiddleware.


### MetricsMiddleware

```go
type MetricsMiddleware struct{ ... }
```


type MetricsMiddleware struct {
	Logger             interfaces.LoggerInterface         // For error logging
	Config             interfaces.ConfigProviderInterface // For metrics configuration
	RequestCountMetric string                             // Name of request count metric
	DurationMetric     string                             // Name of duration metric
	ErrorCountMetric   string                             // Name of error count metric
	EnabledMetricTypes []string                           // Which metric types to collect
}
    MetricsMiddleware provides request metrics collection with Prometheus
    integration. It tracks request counts, durations, and success rates.

func NewMetricsMiddleware(deps MetricsMiddlewareDependencies) *MetricsMiddleware
func (m *MetricsMiddleware) After(ctx interface{}, response interface{}) error
func (m *MetricsMiddleware) Before(ctx interface{}) error

### MetricsMiddlewareDependencies

```go
type MetricsMiddlewareDependencies struct{ ... }
```


type MetricsMiddlewareDependencies struct {
	Logger             interfaces.LoggerInterface
	Config             interfaces.ConfigProviderInterface
	RequestCountMetric string   // Optional, defaults to "http_requests_total"
	DurationMetric     string   // Optional, defaults to "http_request_duration_seconds"
	ErrorCountMetric   string   // Optional, defaults to "http_request_errors_total"
	EnabledMetricTypes []string // Optional, defaults to all types
}
    MetricsMiddlewareDependencies contains all required dependencies for
    MetricsMiddleware.


### MiddlewareError

```go
type MiddlewareError struct{ ... }
```


type MiddlewareError struct {
	Phase      string // "before" or "after"
	Middleware string // Middleware name
	Cause      error  // Original error
}
    MiddlewareError represents errors that occur during middleware execution.

func NewMiddlewareError(phase, middleware string, cause error) *MiddlewareError
func (e *MiddlewareError) Error() string
func (e *MiddlewareError) Unwrap() error

### MiddlewareExecution

```go
type MiddlewareExecution struct{ ... }
```


type MiddlewareExecution struct {
	Name           string        // Middleware name for identification
	BeforeDuration time.Duration // Time taken by Before() method
	AfterDuration  time.Duration // Time taken by After() method
	Success        bool          // Whether middleware executed without errors
	Error          error         // Error if middleware failed
}
    MiddlewareExecution tracks middleware execution details for performance
    monitoring.


### MiddlewareInterface

```go
type MiddlewareInterface interface{ ... }
```


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
    MiddlewareInterface defines the contract for middleware components that
    can intercept and modify service requests and responses. Middleware enables
    cross-cutting concerns like authentication, logging, and metrics without
    modifying core service logic.

    Before is called before the service handler executes and can modify the
    request context. It can return an error to short-circuit execution and
    prevent the handler from running.

    After is called after the service handler executes and can inspect or modify
    the response. It can return an error to indicate middleware processing
    failures, but the original handler result is preserved for consistency.

    Example implementation:

        type LoggingMiddleware struct {
            Logger interfaces.LoggerInterface
        }

        func (l *LoggingMiddleware) Before(ctx *sdk.EndorContext[any]) error {
            l.Logger.Info("Request started", "resource", ctx.GetMicroServiceId())
            return nil
        }

        func (l *LoggingMiddleware) After(ctx *sdk.EndorContext[any], response *sdk.Response[any]) error {
            l.Logger.Info("Request completed", "resource", ctx.GetMicroServiceId())
            return nil
        }


### MiddlewarePipeline

```go
type MiddlewarePipeline struct{ ... }
```


type MiddlewarePipeline struct {
	// Has unexported fields.
}
    MiddlewarePipeline manages the execution of a middleware chain. It handles
    execution order, short-circuiting, and performance tracking.

func NewMiddlewarePipeline(middlewares ...MiddlewareInterface) *MiddlewarePipeline
func (p *MiddlewarePipeline) ExecuteAfter(ctx interface{}, response interface{}) error
func (p *MiddlewarePipeline) ExecuteBefore(ctx interface{}) error
func (p *MiddlewarePipeline) GetExecutions() []MiddlewareExecution
func (p *MiddlewarePipeline) GetTotalDuration() time.Duration

---

*Generated by [endor-sdk-go documentation generator](https://github.com/mattiabonardi/endor-sdk-go/tree/main/tools/gendocs)*
