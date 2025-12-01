# Lifecycle Management

> Package documentation for Lifecycle Management

**Import Path:** `github.com/mattiabonardi/endor-sdk-go/sdk/lifecycle`
**Generated:** 2025-12-01 10:07:54 UTC

---

type BaseHook struct{ ... }
    func NewBaseHook(config HookConfiguration) *BaseHook
type CachedHealthStatus struct{ ... }
type CircuitBreaker struct{ ... }
    func NewCircuitBreaker(config CircuitBreakerConfiguration) *CircuitBreaker
type CircuitBreakerConfiguration struct{ ... }
    func DefaultCircuitBreakerConfiguration() CircuitBreakerConfiguration
type CircuitBreakerState int
    const CircuitClosed CircuitBreakerState = iota ...
type CompositeHealthStatus struct{ ... }
type DefaultLifecycleManager struct{ ... }
type DependencyGraph struct{ ... }
    func NewDependencyGraph() *DependencyGraph
type DependencyHealth struct{ ... }
type HealthAggregationPolicy int
    const AllHealthyPolicy HealthAggregationPolicy = iota ...
type HealthMonitor struct{ ... }
    func NewHealthMonitor(policy HealthAggregationPolicy) *HealthMonitor
type HealthStatus struct{ ... }
type HookConfiguration struct{ ... }
    func DefaultHookConfiguration() HookConfiguration
type HookFailurePolicy int
    const FailFast HookFailurePolicy = iota ...
type HookManager struct{ ... }
    func NewHookManager(config HookConfiguration) *HookManager
type HookPhase int
    const BeforeStartPhase HookPhase = iota ...
type LifecycleHook interface{ ... }
type LifecycleManager interface{ ... }
    func NewLifecycleManager(healthPolicy HealthAggregationPolicy) LifecycleManager
type RecoveryConfiguration struct{ ... }
    func DefaultRecoveryConfiguration() RecoveryConfiguration
type RecoveryManager struct{ ... }
    func NewRecoveryManager() *RecoveryManager
type RecoveryStrategy int
    const NoRecovery RecoveryStrategy = iota ...
type ServiceHealthConfiguration struct{ ... }
    func DefaultHealthConfiguration() ServiceHealthConfiguration
type ServiceHealthStatus int
    const Healthy ServiceHealthStatus = iota ...
type ServiceLifecycleInterface interface{ ... }
type ServiceState int
    const Created ServiceState = iota ...

## Package Overview

package lifecycle // import "github.com/mattiabonardi/endor-sdk-go/sdk/lifecycle"

Package lifecycle provides service lifecycle management for composed services
with dependency-aware ordering, health aggregation, and graceful degradation.

type BaseHook struct{ ... }
    func NewBaseHook(config HookConfiguration) *BaseHook
type CachedHealthStatus struct{ ... }
type CircuitBreaker struct{ ... }
    func NewCircuitBreaker(config CircuitBreakerConfiguration) *CircuitBreaker
type CircuitBreakerConfiguration struct{ ... }
    func DefaultCircuitBreakerConfiguration() CircuitBreakerConfiguration
type CircuitBreakerState int
    const CircuitClosed CircuitBreakerState = iota ...
type CompositeHealthStatus struct{ ... }
type DefaultLifecycleManager struct{ ... }
type DependencyGraph struct{ ... }
    func NewDependencyGraph() *DependencyGraph
type DependencyHealth struct{ ... }
type HealthAggregationPolicy int

## Exported Types

### BaseHook

```go
type BaseHook struct{ ... }
```


type BaseHook struct {
	// Has unexported fields.
}
    BaseHook provides a default implementation of the LifecycleHook interface

func NewBaseHook(config HookConfiguration) *BaseHook
func (h *BaseHook) AfterStart(ctx context.Context, service ServiceLifecycleInterface) error
func (h *BaseHook) AfterStop(ctx context.Context, service ServiceLifecycleInterface) error
func (h *BaseHook) BeforeStart(ctx context.Context, service ServiceLifecycleInterface) error
func (h *BaseHook) BeforeStop(ctx context.Context, service ServiceLifecycleInterface) error
func (h *BaseHook) GetTimeout() time.Duration
func (h *BaseHook) IsCritical() bool

### CachedHealthStatus

```go
type CachedHealthStatus struct{ ... }
```


type CachedHealthStatus struct {
	Status    HealthStatus
	ExpiresAt time.Time
}
    CachedHealthStatus contains a cached health check result

func (c *CachedHealthStatus) IsExpired() bool

### CircuitBreaker

```go
type CircuitBreaker struct{ ... }
```


type CircuitBreaker struct {
	// Has unexported fields.
}
    CircuitBreaker implements the circuit breaker pattern for service failures

func NewCircuitBreaker(config CircuitBreakerConfiguration) *CircuitBreaker
func (cb *CircuitBreaker) CanExecute() bool
func (cb *CircuitBreaker) GetState() CircuitBreakerState
func (cb *CircuitBreaker) GetStats() (int, int, time.Time)
func (cb *CircuitBreaker) OnFailure()
func (cb *CircuitBreaker) OnRequest()
func (cb *CircuitBreaker) OnSuccess()

### CircuitBreakerConfiguration

```go
type CircuitBreakerConfiguration struct{ ... }
```


type CircuitBreakerConfiguration struct {
	// FailureThreshold is the number of failures that trigger the circuit to open
	FailureThreshold int
	// SuccessThreshold is the number of successes needed to close the circuit
	SuccessThreshold int
	// Timeout is how long to wait before transitioning from Open to HalfOpen
	Timeout time.Duration
	// MaxConcurrentRequests is the maximum number of requests allowed in HalfOpen state
	MaxConcurrentRequests int
}
    CircuitBreakerConfiguration contains configuration for circuit breaker

func DefaultCircuitBreakerConfiguration() CircuitBreakerConfiguration

### CircuitBreakerState

```go
type CircuitBreakerState int
```


type CircuitBreakerState int
    CircuitBreakerState represents the state of a circuit breaker

const CircuitClosed CircuitBreakerState = iota ...
func (c CircuitBreakerState) String() string

### CompositeHealthStatus

```go
type CompositeHealthStatus struct{ ... }
```


type CompositeHealthStatus struct {
	// OverallStatus is the aggregated health status
	OverallStatus ServiceHealthStatus `json:"overallStatus"`
	// Services contains the health status of individual services
	Services map[string]HealthStatus `json:"services"`
	// Policy is the aggregation policy used
	Policy HealthAggregationPolicy `json:"policy"`
	// LastCheck is when the composite health was last calculated
	LastCheck time.Time `json:"lastCheck"`
	// HealthySources is the number of healthy services
	HealthySources int `json:"healthySources"`
	// TotalSources is the total number of services
	TotalSources int `json:"totalSources"`
}
    CompositeHealthStatus represents the aggregated health status of multiple
    services


### DefaultLifecycleManager

```go
type DefaultLifecycleManager struct{ ... }
```


type DefaultLifecycleManager struct {
	// Has unexported fields.
}
    DefaultLifecycleManager is the default implementation of LifecycleManager

func (lm *DefaultLifecycleManager) AddServiceDependency(dependent, dependency string) error
func (lm *DefaultLifecycleManager) GetDependencyGraph() *DependencyGraph
func (lm *DefaultLifecycleManager) GetHealth(ctx context.Context) CompositeHealthStatus
func (lm *DefaultLifecycleManager) GetServiceState(serviceName string) (ServiceState, error)
func (lm *DefaultLifecycleManager) RegisterService(name string, service ServiceLifecycleInterface) error
func (lm *DefaultLifecycleManager) RestartService(ctx context.Context, serviceName string) error
func (lm *DefaultLifecycleManager) StartAll(ctx context.Context) error
func (lm *DefaultLifecycleManager) StopAll(ctx context.Context) error

### DependencyGraph

```go
type DependencyGraph struct{ ... }
```


type DependencyGraph struct {
	// Nodes maps service names to their dependencies
	Nodes map[string][]string
	// ReverseDependencies maps service names to services that depend on them
	ReverseDependencies map[string][]string
}
    DependencyGraph represents the dependency relationships between services

func NewDependencyGraph() *DependencyGraph
func (dg *DependencyGraph) AddDependency(dependent, dependency string)
func (dg *DependencyGraph) AddService(serviceName string)
func (dg *DependencyGraph) GetDependencies(serviceName string) []string
func (dg *DependencyGraph) GetDependents(serviceName string) []string
func (dg *DependencyGraph) TopologicalSort() ([]string, error)

### DependencyHealth

```go
type DependencyHealth struct{ ... }
```


type DependencyHealth struct {
	Name   string              `json:"name"`
	Status ServiceHealthStatus `json:"status"`
	Error  string              `json:"error,omitempty"`
}
    DependencyHealth represents the health status of a dependency


### HealthAggregationPolicy

```go
type HealthAggregationPolicy int
```


type HealthAggregationPolicy int
    HealthAggregationPolicy defines how to aggregate health status from multiple
    services

const AllHealthyPolicy HealthAggregationPolicy = iota ...
func (h HealthAggregationPolicy) String() string

### HealthMonitor

```go
type HealthMonitor struct{ ... }
```


type HealthMonitor struct {
	// Has unexported fields.
}
    HealthMonitor manages health checking for services with caching and
    aggregation

func NewHealthMonitor(policy HealthAggregationPolicy) *HealthMonitor
func (hm *HealthMonitor) CheckCompositeHealth(ctx context.Context) CompositeHealthStatus
func (hm *HealthMonitor) CheckHealth(ctx context.Context, serviceName string) (HealthStatus, error)
func (hm *HealthMonitor) GetCacheStatus() map[string]bool
func (hm *HealthMonitor) RegisterService(name string, service ServiceLifecycleInterface, ...)
func (hm *HealthMonitor) StartMonitoring(ctx context.Context) error
func (hm *HealthMonitor) StopMonitoring() error
func (hm *HealthMonitor) UnregisterService(name string)

### HealthStatus

```go
type HealthStatus struct{ ... }
```


type HealthStatus struct {
	// Status is the overall health status of the service
	Status ServiceHealthStatus `json:"status"`
	// Details contains additional health information specific to the service
	Details map[string]interface{} `json:"details,omitempty"`
	// LastCheck is the timestamp of the last health check
	LastCheck time.Time `json:"lastCheck"`
	// Dependencies contains the health status of service dependencies
	Dependencies []DependencyHealth `json:"dependencies,omitempty"`
}
    HealthStatus represents the comprehensive health status of a service


### HookConfiguration

```go
type HookConfiguration struct{ ... }
```


type HookConfiguration struct {
	// Timeout is the maximum time to wait for hook execution
	Timeout time.Duration
	// FailurePolicy defines how to handle hook failures
	FailurePolicy HookFailurePolicy
	// Critical indicates if this hook is critical for service operation
	Critical bool
}
    HookConfiguration contains configuration for hook execution

func DefaultHookConfiguration() HookConfiguration

### HookFailurePolicy

```go
type HookFailurePolicy int
```


type HookFailurePolicy int
    HookFailurePolicy defines how to handle hook execution failures

const FailFast HookFailurePolicy = iota ...

### HookManager

```go
type HookManager struct{ ... }
```


type HookManager struct {
	// Has unexported fields.
}
    HookManager manages lifecycle hooks for a service

func NewHookManager(config HookConfiguration) *HookManager
func (hm *HookManager) AddHook(phase HookPhase, hook LifecycleHook) error
func (hm *HookManager) ExecuteHooks(ctx context.Context, phase HookPhase, service ServiceLifecycleInterface) error
func (hm *HookManager) GetHooksCount() map[HookPhase]int

### HookPhase

```go
type HookPhase int
```


type HookPhase int
    HookPhase represents the phase when a lifecycle hook is executed

const BeforeStartPhase HookPhase = iota ...
func (h HookPhase) String() string

### LifecycleHook

```go
type LifecycleHook interface{ ... }
```


type LifecycleHook interface {
	// BeforeStart is called before service startup
	BeforeStart(ctx context.Context, service ServiceLifecycleInterface) error

	// AfterStart is called after successful service startup
	AfterStart(ctx context.Context, service ServiceLifecycleInterface) error

	// BeforeStop is called before service shutdown
	BeforeStop(ctx context.Context, service ServiceLifecycleInterface) error

	// AfterStop is called after successful service shutdown
	AfterStop(ctx context.Context, service ServiceLifecycleInterface) error

	// IsCritical returns true if this hook is critical for service operation
	IsCritical() bool

	// GetTimeout returns the maximum execution time for this hook
	GetTimeout() time.Duration
}
    LifecycleHook defines the interface for lifecycle hooks


### LifecycleManager

```go
type LifecycleManager interface{ ... }
```


type LifecycleManager interface {
	// RegisterService registers a service for lifecycle management
	RegisterService(name string, service ServiceLifecycleInterface) error

	// AddServiceDependency adds a dependency relationship between services
	AddServiceDependency(dependent, dependency string) error

	// StartAll starts all services in dependency order
	StartAll(ctx context.Context) error

	// StopAll stops all services in reverse dependency order
	StopAll(ctx context.Context) error

	// GetHealth returns the composite health status of all services
	GetHealth(ctx context.Context) CompositeHealthStatus

	// GetDependencyGraph returns the current dependency graph
	GetDependencyGraph() *DependencyGraph

	// GetServiceState returns the current state of a specific service
	GetServiceState(serviceName string) (ServiceState, error)

	// RestartService restarts a specific service and its dependents
	RestartService(ctx context.Context, serviceName string) error
}
    LifecycleManager coordinates the lifecycle of multiple services with
    dependency management

func NewLifecycleManager(healthPolicy HealthAggregationPolicy) LifecycleManager

### RecoveryConfiguration

```go
type RecoveryConfiguration struct{ ... }
```


type RecoveryConfiguration struct {
	// Strategy defines the recovery strategy to use
	Strategy RecoveryStrategy
	// MaxAttempts is the maximum number of recovery attempts
	MaxAttempts int
	// InitialDelay is the initial delay before first recovery attempt
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between recovery attempts
	MaxDelay time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff
	BackoffMultiplier float64
	// LinearIncrement is the increment for linear backoff
	LinearIncrement time.Duration
	// RecoveryTimeout is the maximum time to wait for recovery
	RecoveryTimeout time.Duration
	// HealthCheckInterval is how often to check if recovery is needed
	HealthCheckInterval time.Duration
}
    RecoveryConfiguration contains configuration for service recovery

func DefaultRecoveryConfiguration() RecoveryConfiguration

### RecoveryManager

```go
type RecoveryManager struct{ ... }
```


type RecoveryManager struct {
	// Has unexported fields.
}
    RecoveryManager handles service recovery and circuit breaking

func NewRecoveryManager() *RecoveryManager
func (rm *RecoveryManager) GetCircuitBreakerState(serviceName string) (CircuitBreakerState, error)
func (rm *RecoveryManager) IsRecoveryInProgress(serviceName string) bool
func (rm *RecoveryManager) RegisterService(name string, service ServiceLifecycleInterface, config RecoveryConfiguration)
func (rm *RecoveryManager) StartRecoveryMonitoring(ctx context.Context)
func (rm *RecoveryManager) StopRecoveryMonitoring()
func (rm *RecoveryManager) UnregisterService(name string)

### RecoveryStrategy

```go
type RecoveryStrategy int
```


type RecoveryStrategy int
    RecoveryStrategy defines the strategy for recovering from service failures

const NoRecovery RecoveryStrategy = iota ...
func (r RecoveryStrategy) String() string

### ServiceHealthConfiguration

```go
type ServiceHealthConfiguration struct{ ... }
```


type ServiceHealthConfiguration struct {
	// CheckInterval is the frequency of health checks
	CheckInterval time.Duration
	// CheckTimeout is the maximum time to wait for a health check
	CheckTimeout time.Duration
	// CacheTimeout is how long to cache health check results
	CacheTimeout time.Duration
	// IsCritical indicates if this service is critical for overall system health
	IsCritical bool
	// RetryAttempts is the number of retry attempts for failed health checks
	RetryAttempts int
	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration
}
    ServiceHealthConfiguration contains configuration for service health
    checking

func DefaultHealthConfiguration() ServiceHealthConfiguration

### ServiceHealthStatus

```go
type ServiceHealthStatus int
```


type ServiceHealthStatus int
    ServiceHealthStatus represents the health status of a service

const Healthy ServiceHealthStatus = iota ...
func (h ServiceHealthStatus) String() string

### ServiceLifecycleInterface

```go
type ServiceLifecycleInterface interface{ ... }
```


type ServiceLifecycleInterface interface {
	// Start initiates the service startup process
	// Returns an error if startup fails
	Start(ctx context.Context) error

	// Stop initiates the service shutdown process
	// Returns an error if shutdown fails
	Stop(ctx context.Context) error

	// HealthCheck performs a health check and returns the current health status
	// Returns the health status with dependency information
	HealthCheck(ctx context.Context) HealthStatus

	// GetState returns the current lifecycle state of the service
	GetState() ServiceState

	// AddHook registers a lifecycle hook for this service
	// Returns an error if the hook cannot be added
	AddHook(hook LifecycleHook) error
}
    ServiceLifecycleInterface defines the core lifecycle operations for services


### ServiceState

```go
type ServiceState int
```


type ServiceState int
    ServiceState represents the current state of a service in its lifecycle

const Created ServiceState = iota ...
func (s ServiceState) String() string

---

*Generated by [endor-sdk-go documentation generator](https://github.com/mattiabonardi/endor-sdk-go/tree/main/tools/gendocs)*
