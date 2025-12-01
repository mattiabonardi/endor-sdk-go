# Service Composition

> Package documentation for Service Composition

**Import Path:** `github.com/mattiabonardi/endor-sdk-go/sdk/composition`
**Generated:** 2025-12-01 10:07:52 UTC

---

type AllResultsMerger struct{ ... }
type BranchBuilder struct{ ... }
    func NewBranchBuilder(router BranchRouter) *BranchBuilder
type BranchRouter interface{ ... }
type BranchedService struct{ ... }
    func ServiceBranch(router BranchRouter, services map[string]interfaces.EndorServiceInterface) *BranchedService
type ChainBuilder struct{ ... }
    func NewChainBuilder() *ChainBuilder
type ChainedService struct{ ... }
    func ServiceChain(services ...interfaces.EndorServiceInterface) *ChainedService
type CompositionConfig struct{ ... }
type CompositionError struct{ ... }
    func NewCompositionError(operation, serviceName string, serviceIndex int, cause error) *CompositionError
type CompositionPattern interface{ ... }
type CompositionValidator interface{ ... }
    func DefaultValidator() CompositionValidator
type ContextRouter struct{ ... }
type ErrorHandlingStrategy int
    const FailFast ErrorHandlingStrategy = iota ...
type FirstWinsMerger struct{ ... }
type FunctionMerger struct{ ... }
type FunctionRouter struct{ ... }
type LoggingInterceptor struct{ ... }
type MergedService struct{ ... }
    func ServiceMerger(services []interfaces.EndorServiceInterface, merger ResultMerger) *MergedService
type MergerBuilder struct{ ... }
    func NewMergerBuilder() *MergerBuilder
type NoOpInterceptor struct{}
type ProxiedService struct{ ... }
    func ServiceProxy(target interfaces.EndorServiceInterface, interceptor ProxyInterceptor) *ProxiedService
type ProxyBuilder struct{ ... }
    func NewProxyBuilder(target interfaces.EndorServiceInterface) *ProxyBuilder
type ProxyInterceptor interface{ ... }
type ResultMerger interface{ ... }
type SimpleRouter struct{ ... }
type TransformationInterceptor struct{ ... }
type ValidationMode int
    const Strict ValidationMode = iota ...

## Package Overview

package composition // import "github.com/mattiabonardi/endor-sdk-go/sdk/composition"

type AllResultsMerger struct{ ... }
type BranchBuilder struct{ ... }
    func NewBranchBuilder(router BranchRouter) *BranchBuilder
type BranchRouter interface{ ... }
type BranchedService struct{ ... }
    func ServiceBranch(router BranchRouter, services map[string]interfaces.EndorServiceInterface) *BranchedService
type ChainBuilder struct{ ... }
    func NewChainBuilder() *ChainBuilder
type ChainedService struct{ ... }
    func ServiceChain(services ...interfaces.EndorServiceInterface) *ChainedService
type CompositionConfig struct{ ... }
type CompositionError struct{ ... }
    func NewCompositionError(operation, serviceName string, serviceIndex int, cause error) *CompositionError
type CompositionPattern interface{ ... }
type CompositionValidator interface{ ... }
    func DefaultValidator() CompositionValidator
type ContextRouter struct{ ... }
type ErrorHandlingStrategy int

## Exported Types

### AllResultsMerger

```go
type AllResultsMerger struct{ ... }
```


type AllResultsMerger struct {
	Timeout time.Duration
}
    AllResultsMerger collects all results into a slice.

func (a *AllResultsMerger) GetTimeout() time.Duration
func (a *AllResultsMerger) MergeResults(ctx context.Context, request interface{}, results []interface{}) (interface{}, error)
func (a *AllResultsMerger) ShouldWaitForAll() bool

### BranchBuilder

```go
type BranchBuilder struct{ ... }
```


type BranchBuilder struct {
	// Has unexported fields.
}
    BranchBuilder provides a fluent API for building complex service branches.

func NewBranchBuilder(router BranchRouter) *BranchBuilder
func (b *BranchBuilder) AddService(key string, service interfaces.EndorServiceInterface) *BranchBuilder
func (b *BranchBuilder) AddServices(services map[string]interfaces.EndorServiceInterface) *BranchBuilder
func (b *BranchBuilder) Build() *BranchedService
func (b *BranchBuilder) WithErrorHandling(strategy ErrorHandlingStrategy) *BranchBuilder
func (b *BranchBuilder) WithTimeout(timeout time.Duration) *BranchBuilder

### BranchRouter

```go
type BranchRouter interface{ ... }
```


type BranchRouter interface {
	// SelectService analyzes the request and returns the key of the service that should handle it.
	// Returns an error if no suitable service can be determined or if routing fails.
	SelectService(ctx context.Context, request interface{}) (string, error)

	// GetDefaultService returns the key of the service to use when no specific route is found.
	// Returns empty string if no default service is configured.
	GetDefaultService() string
}
    BranchRouter defines the interface for conditional service selection in
    branched compositions. It analyzes the request context and data to determine
    which service should handle the request.


### BranchedService

```go
type BranchedService struct{ ... }
```


type BranchedService struct {
	// Has unexported fields.
}
    BranchedService represents a conditional routing service that selects which
    service to execute based on request analysis and routing logic.

func ServiceBranch(router BranchRouter, services map[string]interfaces.EndorServiceInterface) *BranchedService
func (b *BranchedService) Execute(ctx context.Context, request interface{}) (interface{}, error)
func (b *BranchedService) GetConfig() CompositionConfig
func (b *BranchedService) GetRouter() BranchRouter
func (b *BranchedService) GetServiceKeys() []string
func (b *BranchedService) GetServices() []interfaces.EndorServiceInterface
func (b *BranchedService) Validate() error
func (b *BranchedService) WithConfig(config CompositionConfig) *BranchedService
func (b *BranchedService) WithValidator(validator CompositionValidator) *BranchedService

### ChainBuilder

```go
type ChainBuilder struct{ ... }
```


type ChainBuilder struct {
	// Has unexported fields.
}
    ChainBuilder provides a fluent API for building complex service chains.

func NewChainBuilder() *ChainBuilder
func (b *ChainBuilder) Add(service interfaces.EndorServiceInterface) *ChainBuilder
func (b *ChainBuilder) AddServices(services ...interfaces.EndorServiceInterface) *ChainBuilder
func (b *ChainBuilder) Build() *ChainedService
func (b *ChainBuilder) WithErrorHandling(strategy ErrorHandlingStrategy) *ChainBuilder
func (b *ChainBuilder) WithMetrics(enabled bool) *ChainBuilder
func (b *ChainBuilder) WithTimeout(timeout time.Duration) *ChainBuilder

### ChainedService

```go
type ChainedService struct{ ... }
```


type ChainedService struct {
	// Has unexported fields.
}
    ChainedService represents a composition of services that execute in
    sequence. Each service in the chain receives the output of the previous
    service as its input.

func ServiceChain(services ...interfaces.EndorServiceInterface) *ChainedService
func (c *ChainedService) Execute(ctx context.Context, request interface{}) (interface{}, error)
func (c *ChainedService) GetConfig() CompositionConfig
func (c *ChainedService) GetServices() []interfaces.EndorServiceInterface
func (c *ChainedService) Length() int
func (c *ChainedService) Validate() error
func (c *ChainedService) WithConfig(config CompositionConfig) *ChainedService
func (c *ChainedService) WithValidator(validator CompositionValidator) *ChainedService

### CompositionConfig

```go
type CompositionConfig struct{ ... }
```


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
    CompositionConfig provides configuration options for composition patterns.


### CompositionError

```go
type CompositionError struct{ ... }
```


type CompositionError struct {
	Operation    string                 // The composition operation that failed (chain, proxy, branch, merge)
	ServiceName  string                 // The name of the service that caused the error
	ServiceIndex int                    // The index of the service in the composition (if applicable)
	Cause        error                  // The underlying error that caused the failure
	Context      map[string]interface{} // Additional context about the failure
}
    CompositionError represents errors that occur during service composition
    operations. It provides detailed context about which service failed and why.

func NewCompositionError(operation, serviceName string, serviceIndex int, cause error) *CompositionError
func (e *CompositionError) Error() string
func (e *CompositionError) Unwrap() error
func (e *CompositionError) WithContext(key string, value interface{}) *CompositionError

### CompositionPattern

```go
type CompositionPattern interface{ ... }
```


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
    CompositionPattern defines the core interface for all service composition
    utilities. It provides a unified contract for executing, validating,
    and introspecting composed services.


### CompositionValidator

```go
type CompositionValidator interface{ ... }
```


type CompositionValidator interface {
	// ValidateChain verifies that services in a chain have compatible input/output types.
	ValidateChain(services []interfaces.EndorServiceInterface) error

	// ValidateTypes ensures that the input type is compatible with the service interfaces.
	ValidateTypes(inputType reflect.Type, services []interfaces.EndorServiceInterface) error

	// AnalyzeDependencyGraph analyzes the composition for circular dependencies and ordering issues.
	AnalyzeDependencyGraph(composition CompositionPattern) ([]string, error)
}
    CompositionValidator provides validation capabilities for service
    compositions. It ensures type safety and interface compatibility across
    service boundaries.

func DefaultValidator() CompositionValidator

### ContextRouter

```go
type ContextRouter struct{ ... }
```


type ContextRouter struct {
	ContextKey     string
	RouteMap       map[interface{}]string
	DefaultService string
}
    ContextRouter provides routing based on context values.

func (r *ContextRouter) GetDefaultService() string
func (r *ContextRouter) SelectService(ctx context.Context, request interface{}) (string, error)

### ErrorHandlingStrategy

```go
type ErrorHandlingStrategy int
```


type ErrorHandlingStrategy int
    ErrorHandlingStrategy defines how compositions should handle service
    failures.

const FailFast ErrorHandlingStrategy = iota ...

### FirstWinsMerger

```go
type FirstWinsMerger struct{ ... }
```


type FirstWinsMerger struct {
	Timeout time.Duration
}
    FirstWinsMerger returns the first successful result received.

func (f *FirstWinsMerger) GetTimeout() time.Duration
func (f *FirstWinsMerger) MergeResults(ctx context.Context, request interface{}, results []interface{}) (interface{}, error)
func (f *FirstWinsMerger) ShouldWaitForAll() bool

### FunctionMerger

```go
type FunctionMerger struct{ ... }
```


type FunctionMerger struct {
	MergeFunc  func(context.Context, interface{}, []interface{}) (interface{}, error)
	WaitForAll bool
	Timeout    time.Duration
}
    FunctionMerger uses a custom function to merge results.

func (f *FunctionMerger) GetTimeout() time.Duration
func (f *FunctionMerger) MergeResults(ctx context.Context, request interface{}, results []interface{}) (interface{}, error)
func (f *FunctionMerger) ShouldWaitForAll() bool

### FunctionRouter

```go
type FunctionRouter struct{ ... }
```


type FunctionRouter struct {
	RoutingFunc    func(context.Context, interface{}) (string, error)
	DefaultService string
}
    FunctionRouter provides routing based on a custom function.

func (r *FunctionRouter) GetDefaultService() string
func (r *FunctionRouter) SelectService(ctx context.Context, request interface{}) (string, error)

### LoggingInterceptor

```go
type LoggingInterceptor struct{ ... }
```


type LoggingInterceptor struct {
	ServiceName string
}
    LoggingInterceptor provides basic request/response logging functionality.

func (l *LoggingInterceptor) AfterResponse(ctx context.Context, request interface{}, response interface{}) (interface{}, error)
func (l *LoggingInterceptor) BeforeRequest(ctx context.Context, request interface{}) (interface{}, error)

### MergedService

```go
type MergedService struct{ ... }
```


type MergedService struct {
	// Has unexported fields.
}
    MergedService represents a composition that executes multiple services in
    parallel and merges their results using a configurable merging strategy.

func ServiceMerger(services []interfaces.EndorServiceInterface, merger ResultMerger) *MergedService
func (m *MergedService) Execute(ctx context.Context, request interface{}) (interface{}, error)
func (m *MergedService) GetConfig() CompositionConfig
func (m *MergedService) GetMerger() ResultMerger
func (m *MergedService) GetServices() []interfaces.EndorServiceInterface
func (m *MergedService) Validate() error
func (m *MergedService) WithConfig(config CompositionConfig) *MergedService
func (m *MergedService) WithValidator(validator CompositionValidator) *MergedService

### MergerBuilder

```go
type MergerBuilder struct{ ... }
```


type MergerBuilder struct {
	// Has unexported fields.
}
    MergerBuilder provides a fluent API for building complex service mergers.

func NewMergerBuilder() *MergerBuilder
func (b *MergerBuilder) AddService(service interfaces.EndorServiceInterface) *MergerBuilder
func (b *MergerBuilder) AddServices(services ...interfaces.EndorServiceInterface) *MergerBuilder
func (b *MergerBuilder) Build() *MergedService
func (b *MergerBuilder) WithAllResults(timeout time.Duration) *MergerBuilder
func (b *MergerBuilder) WithConcurrency(maxConcurrency int) *MergerBuilder
func (b *MergerBuilder) WithErrorHandling(strategy ErrorHandlingStrategy) *MergerBuilder
func (b *MergerBuilder) WithFirstWins(timeout time.Duration) *MergerBuilder
func (b *MergerBuilder) WithMerger(merger ResultMerger) *MergerBuilder
func (b *MergerBuilder) WithTimeout(timeout time.Duration) *MergerBuilder

### NoOpInterceptor

```go
type NoOpInterceptor struct{}
```


type NoOpInterceptor struct{}
    NoOpInterceptor provides a no-operation interceptor that passes requests and
    responses unchanged.

func (n *NoOpInterceptor) AfterResponse(ctx context.Context, request interface{}, response interface{}) (interface{}, error)
func (n *NoOpInterceptor) BeforeRequest(ctx context.Context, request interface{}) (interface{}, error)

### ProxiedService

```go
type ProxiedService struct{ ... }
```


type ProxiedService struct {
	// Has unexported fields.
}
    ProxiedService represents a service that forwards requests to a target
    service with optional interception and transformation of requests and
    responses.

func ServiceProxy(target interfaces.EndorServiceInterface, interceptor ProxyInterceptor) *ProxiedService
func (p *ProxiedService) Execute(ctx context.Context, request interface{}) (interface{}, error)
func (p *ProxiedService) GetConfig() CompositionConfig
func (p *ProxiedService) GetInterceptor() ProxyInterceptor
func (p *ProxiedService) GetServices() []interfaces.EndorServiceInterface
func (p *ProxiedService) GetTarget() interfaces.EndorServiceInterface
func (p *ProxiedService) Validate() error
func (p *ProxiedService) WithConfig(config CompositionConfig) *ProxiedService
func (p *ProxiedService) WithValidator(validator CompositionValidator) *ProxiedService

### ProxyBuilder

```go
type ProxyBuilder struct{ ... }
```


type ProxyBuilder struct {
	// Has unexported fields.
}
    ProxyBuilder provides a fluent API for building complex service proxies.

func NewProxyBuilder(target interfaces.EndorServiceInterface) *ProxyBuilder
func (b *ProxyBuilder) Build() *ProxiedService
func (b *ProxyBuilder) WithErrorHandling(strategy ErrorHandlingStrategy) *ProxyBuilder
func (b *ProxyBuilder) WithInterceptor(interceptor ProxyInterceptor) *ProxyBuilder
func (b *ProxyBuilder) WithLogging(serviceName string) *ProxyBuilder
func (b *ProxyBuilder) WithTimeout(timeout time.Duration) *ProxyBuilder

### ProxyInterceptor

```go
type ProxyInterceptor interface{ ... }
```


type ProxyInterceptor interface {
	// BeforeRequest is called before the request is forwarded to the target service.
	// It can modify the request or return an error to terminate the proxy operation.
	BeforeRequest(ctx context.Context, request interface{}) (interface{}, error)

	// AfterResponse is called after the response is received from the target service.
	// It can modify the response or return an error to indicate proxy processing failure.
	AfterResponse(ctx context.Context, request interface{}, response interface{}) (interface{}, error)
}
    ProxyInterceptor defines the interface for request/response interception in
    service proxies. It enables transformation of requests and responses as they
    flow through proxy layers.


### ResultMerger

```go
type ResultMerger interface{ ... }
```


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
    ResultMerger defines the interface for aggregating results from multiple
    services. It provides flexible strategies for combining outputs from
    parallel service executions.


### SimpleRouter

```go
type SimpleRouter struct{ ... }
```


type SimpleRouter struct {
	FieldName      string
	RouteMap       map[interface{}]string
	DefaultService string
}
    SimpleRouter provides a basic router implementation based on request field
    values.

func (r *SimpleRouter) GetDefaultService() string
func (r *SimpleRouter) SelectService(ctx context.Context, request interface{}) (string, error)

### TransformationInterceptor

```go
type TransformationInterceptor struct{ ... }
```


type TransformationInterceptor struct {
	RequestTransformer  func(context.Context, interface{}) (interface{}, error)
	ResponseTransformer func(context.Context, interface{}, interface{}) (interface{}, error)
}
    TransformationInterceptor provides request/response transformation
    capabilities.

func (t *TransformationInterceptor) AfterResponse(ctx context.Context, request interface{}, response interface{}) (interface{}, error)
func (t *TransformationInterceptor) BeforeRequest(ctx context.Context, request interface{}) (interface{}, error)

### ValidationMode

```go
type ValidationMode int
```


type ValidationMode int
    ValidationMode defines how strict composition validation should be.

const Strict ValidationMode = iota ...

---

*Generated by [endor-sdk-go documentation generator](https://github.com/mattiabonardi/endor-sdk-go/tree/main/tools/gendocs)*
