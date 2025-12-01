# Dependency Injection

> Package documentation for Dependency Injection

**Import Path:** `github.com/mattiabonardi/endor-sdk-go/sdk/di`
**Generated:** 2025-12-01 10:07:51 UTC

---

func IsCommonFrameworkDependency(interfaceType reflect.Type) bool
func Register[T any](container Container, impl T, scope Scope) error
func RegisterFactory[T any](container Container, factory func(Container) (T, error), scope Scope) error
func Resolve[T any](container Container) (T, error)
func ResolveWithContext[T any](ctx context.Context, container Container) (T, error)
type CircuitBreaker struct{ ... }
    func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker
type CircuitBreakerConfig struct{ ... }
    func DefaultCircuitBreakerConfig() CircuitBreakerConfig
type CircuitBreakerError struct{ ... }
type CircuitBreakerManager struct{ ... }
    func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager
type CircuitState int
    const CircuitClosed CircuitState = iota ...
type CircularDependencyError struct{ ... }
    func NewCircularDependencyError(typ reflect.Type, path []string) *CircularDependencyError
type ConfigUpdatable interface{ ... }
type Container interface{ ... }
    func NewContainer() Container
type ContextKeyGenerator interface{ ... }
type DefaultContextKeyGenerator struct{}
type DependencyError struct{ ... }
    func NewDependencyError(typ reflect.Type, operation, message string, context map[string]interface{}) *DependencyError
type DependencyHealth struct{ ... }
type DependencyLifecycle struct{ ... }
type DependencyMemoryDetail struct{ ... }
type DependencyMetrics struct{ ... }
type DependencyPool struct{ ... }
    func NewDependencyPool(newFunc func() interface{}) *DependencyPool
type DependencyPoolManager struct{ ... }
    func NewDependencyPoolManager() *DependencyPoolManager
type DependencyUpdate struct{ ... }
type DependencyUpdateManager struct{ ... }
    func NewDependencyUpdateManager(container *containerImpl) *DependencyUpdateManager
type DependencyWithCircuitBreaker struct{ ... }
    func NewDependencyWithCircuitBreaker(dependency interface{}, dependencyType string, breaker *CircuitBreaker) *DependencyWithCircuitBreaker
type HealthAwareCircuitBreakerListener struct{ ... }
    func NewHealthAwareCircuitBreakerListener(cbm *CircuitBreakerManager) *HealthAwareCircuitBreakerListener
type HealthCheck struct{ ... }
type HealthChecker interface{ ... }
type HealthEvent struct{ ... }
type HealthListener interface{ ... }
type HealthMonitor struct{ ... }
    func NewHealthMonitor() *HealthMonitor
type HealthOption func(*DependencyHealth)
    func WithCheckInterval(interval time.Duration) HealthOption
    func WithMaxFailures(maxFailures int) HealthOption
type HealthStatus int
    const HealthUnknown HealthStatus = iota ...
type HotReloadable interface{ ... }
type LifecycleAware interface{ ... }
type LifecycleEvent struct{ ... }
type LifecycleListener interface{ ... }
type LifecycleManager struct{ ... }
    func NewLifecycleManager() *LifecycleManager
type LifecycleManagerHealthListener struct{ ... }
    func NewLifecycleManagerHealthListener(lm *LifecycleManager) *LifecycleManagerHealthListener
type LifecycleOption func(*DependencyLifecycle)
    func WithDependencies(dependencies ...string) LifecycleOption
    func WithPriority(priority LifecyclePriority) LifecycleOption
type LifecyclePriority int
    const PriorityLowest LifecyclePriority = iota ...
type LifecycleState int
    const LifecycleStateUnknown LifecycleState = iota ...
type MemoryOptimizationReport struct{ ... }
type MemoryProfiler struct{ ... }
    func NewMemoryProfiler(enabled bool) *MemoryProfiler
type MemorySnapshot struct{ ... }
type MemoryStats struct{ ... }
type MemoryTracker struct{ ... }
    func NewMemoryTracker() *MemoryTracker
type MemoryTrendAnalysis struct{ ... }
type Scope int
    const Singleton Scope = iota ...
    func RecommendScope(interfaceType reflect.Type) Scope
type SharedDependencyManager struct{ ... }
    func NewSharedDependencyManager() *SharedDependencyManager
type Startable interface{ ... }
type Stoppable interface{ ... }
type UpdateListener interface{ ... }
type UpdatePropagationStrategy int
    const UpdatePropagationImmediate UpdatePropagationStrategy = iota ...
type UpdateType int
    const UpdateTypeUnknown UpdateType = iota ...
type ValidationError struct{ ... }
    func NewValidationError(errors []error, graph map[string][]string) *ValidationError
type Versionable interface{ ... }

## Package Overview

package di // import "github.com/mattiabonardi/endor-sdk-go/sdk/di"

func IsCommonFrameworkDependency(interfaceType reflect.Type) bool
func Register[T any](container Container, impl T, scope Scope) error
func RegisterFactory[T any](container Container, factory func(Container) (T, error), scope Scope) error
func Resolve[T any](container Container) (T, error)
func ResolveWithContext[T any](ctx context.Context, container Container) (T, error)
type CircuitBreaker struct{ ... }
    func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker
type CircuitBreakerConfig struct{ ... }
    func DefaultCircuitBreakerConfig() CircuitBreakerConfig
type CircuitBreakerError struct{ ... }
type CircuitBreakerManager struct{ ... }
    func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager
type CircuitState int
    const CircuitClosed CircuitState = iota ...
type CircularDependencyError struct{ ... }
    func NewCircularDependencyError(typ reflect.Type, path []string) *CircularDependencyError
type ConfigUpdatable interface{ ... }
type Container interface{ ... }

## Exported Types

### IsCommonFrameworkDependency

```go
func IsCommonFrameworkDependency(interfaceType reflect.Type) bool
```

### Register[T

```go
func Register[T any](container Container, impl T, scope Scope) error
```

### RegisterFactory[T

```go
func RegisterFactory[T any](container Container, factory func(Container) (T, error), scope Scope) error
```

### Resolve[T

```go
func Resolve[T any](container Container) (T, error)
```

### ResolveWithContext[T

```go
func ResolveWithContext[T any](ctx context.Context, container Container) (T, error)
```

### CircuitBreaker

```go
type CircuitBreaker struct{ ... }
```


type CircuitBreaker struct {
	// Has unexported fields.
}
    CircuitBreaker provides circuit breaker functionality for dependencies

func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker
func (cb *CircuitBreaker) Execute(fn func() error) error
func (cb *CircuitBreaker) GetFailureCount() int64
func (cb *CircuitBreaker) GetState() CircuitState
func (cb *CircuitBreaker) Reset()

### CircuitBreakerConfig

```go
type CircuitBreakerConfig struct{ ... }
```


type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures needed to open the circuit
	FailureThreshold int
	// OpenTimeout is how long the circuit stays open before trying to close
	OpenTimeout time.Duration
	// SuccessThreshold is the number of consecutive successes needed to close the circuit from half-open
	SuccessThreshold int
}
    CircuitBreakerConfig holds configuration for a circuit breaker

func DefaultCircuitBreakerConfig() CircuitBreakerConfig

### CircuitBreakerError

```go
type CircuitBreakerError struct{ ... }
```


type CircuitBreakerError struct {
	DependencyType string
	State          CircuitState
}
    CircuitBreakerError represents an error when the circuit is open

func (e *CircuitBreakerError) Error() string

### CircuitBreakerManager

```go
type CircuitBreakerManager struct{ ... }
```


type CircuitBreakerManager struct {
	// Has unexported fields.
}
    CircuitBreakerManager manages circuit breakers for different dependencies

func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager
func (cbm *CircuitBreakerManager) ExecuteWithBreaker(dependencyType string, fn func() error) error
func (cbm *CircuitBreakerManager) GetBreakerState(dependencyType string) CircuitState
func (cbm *CircuitBreakerManager) GetCircuitBreaker(dependencyType string) *CircuitBreaker
func (cbm *CircuitBreakerManager) GetOpenBreakers() []string
func (cbm *CircuitBreakerManager) Reset()

### CircuitState

```go
type CircuitState int
```


type CircuitState int
    CircuitState represents the state of a circuit breaker

const CircuitClosed CircuitState = iota ...
func (c CircuitState) String() string

### CircularDependencyError

```go
type CircularDependencyError struct{ ... }
```


type CircularDependencyError struct {
	// Path shows the dependency resolution path that led to the cycle
	Path []string
	// Type is the interface type where the cycle was detected
	Type reflect.Type
}
    CircularDependencyError represents a circular dependency detection error

func NewCircularDependencyError(typ reflect.Type, path []string) *CircularDependencyError
func (e *CircularDependencyError) Error() string

### ConfigUpdatable

```go
type ConfigUpdatable interface{ ... }
```


type ConfigUpdatable interface {
	UpdateConfig(config map[string]interface{}) error
}
    ConfigUpdatable represents a dependency that can update its configuration


### Container

```go
type Container interface{ ... }
```


type Container interface {
	// RegisterType registers an implementation for an interface type with the specified scope.
	// interfaceType must be an interface type, and impl must implement that interface.
	// Returns error if registration fails (e.g., interfaceType is not interface, circular dependency).
	RegisterType(interfaceType reflect.Type, impl interface{}, scope Scope) error

	// RegisterFactoryType registers a factory function for an interface type with the specified scope.
	// The factory function receives the container instance to resolve its own dependencies.
	// interfaceType must be an interface type, and the factory must return an implementation of that type.
	RegisterFactoryType(interfaceType reflect.Type, factory interface{}, scope Scope) error

	// ResolveType resolves a dependency by interface type, returning the registered implementation.
	// interfaceType must be an interface type that has been previously registered.
	// Returns error if the dependency is not registered or resolution fails.
	ResolveType(interfaceType reflect.Type) (interface{}, error)

	// ResolveTypeWithContext resolves a dependency by interface type with context for scoped dependencies.
	// For Singleton scope, context is ignored. For Scoped scope, context determines the scope boundary.
	ResolveTypeWithContext(ctx context.Context, interfaceType reflect.Type) (interface{}, error)

	// Validate checks the completeness and correctness of the dependency graph.
	// Returns slice of errors for any issues found (missing dependencies, circular dependencies, etc.).
	// An empty slice indicates a valid dependency graph.
	Validate() []error

	// GetDependencyGraph returns a map representation of the dependency graph for debugging.
	// Keys are type names, values are lists of dependencies for that type.
	GetDependencyGraph() map[string][]string

	// Reset clears all registered dependencies. Primarily used for testing scenarios.
	Reset()

	// GetHealthMonitor returns the health monitor instance for dependency health tracking
	GetHealthMonitor() *HealthMonitor

	// GetMemoryTracker returns the memory tracker for dependency optimization analysis
	GetMemoryTracker() *MemoryTracker

	// GetMemoryProfiler returns the memory profiler for memory usage analysis
	GetMemoryProfiler() *MemoryProfiler

	// GetPoolManager returns the dependency pool manager for resource pooling
	GetPoolManager() *DependencyPoolManager

	// GetLifecycleManager returns the lifecycle manager for dependency startup/shutdown
	GetLifecycleManager() *LifecycleManager

	// GetUpdateManager returns the dependency update manager for hot-reload and update propagation
	GetUpdateManager() *DependencyUpdateManager
}
    Container defines the dependency injection container interface with
    registration and resolution capabilities.

func NewContainer() Container

### ContextKeyGenerator

```go
type ContextKeyGenerator interface{ ... }
```


type ContextKeyGenerator interface {
	GenerateKey(ctx context.Context) string
}
    ContextKeyGenerator generates unique context keys for scoped dependencies


### DefaultContextKeyGenerator

```go
type DefaultContextKeyGenerator struct{}
```


type DefaultContextKeyGenerator struct{}
    DefaultContextKeyGenerator uses request ID or correlation ID from context

func (g *DefaultContextKeyGenerator) GenerateKey(ctx context.Context) string

### DependencyError

```go
type DependencyError struct{ ... }
```


type DependencyError struct {
	// Type is the interface type that failed to resolve
	Type reflect.Type
	// Operation describes what operation failed (registration, resolution, etc.)
	Operation string
	// Message provides human-readable error details
	Message string
	// Context provides additional error context for debugging
	Context map[string]interface{}
}
    DependencyError represents errors that occur during dependency operations

func NewDependencyError(typ reflect.Type, operation, message string, context map[string]interface{}) *DependencyError
func (e *DependencyError) Error() string

### DependencyHealth

```go
type DependencyHealth struct{ ... }
```


type DependencyHealth struct {
	DependencyType   string
	Instance         interface{}
	LastCheck        HealthCheck
	ConsecutiveFails int
	CheckInterval    time.Duration
	MaxFailures      int
	// Has unexported fields.
}
    DependencyHealth tracks health status for a specific dependency


### DependencyLifecycle

```go
type DependencyLifecycle struct{ ... }
```


type DependencyLifecycle struct {
	DependencyType string
	Instance       interface{}
	State          LifecycleState
	Priority       LifecyclePriority
	Dependencies   []string // Dependencies this component depends on
	Dependents     []string // Components that depend on this one
	StartedAt      time.Time
	StoppedAt      time.Time
	LastError      error
	// Has unexported fields.
}
    DependencyLifecycle tracks lifecycle state for a dependency


### DependencyMemoryDetail

```go
type DependencyMemoryDetail struct{ ... }
```


type DependencyMemoryDetail struct {
	DependencyType string    `json:"dependency_type"`
	AllocationSize uint64    `json:"allocation_size"`
	ReferenceCount int64     `json:"reference_count"`
	ShareCount     int64     `json:"share_count"`
	MemorySaved    uint64    `json:"memory_saved"`
	Scope          Scope     `json:"scope"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessed   time.Time `json:"last_accessed"`
}
    DependencyMemoryDetail provides detailed memory information for a specific
    dependency


### DependencyMetrics

```go
type DependencyMetrics struct{ ... }
```


type DependencyMetrics struct {
	// Total resolutions by scope
	SingletonResolutions int64
	ScopedResolutions    int64
	TransientResolutions int64

	// Cache hit rates
	SingletonCacheHits int64
	ScopedCacheHits    int64

	// Performance metrics
	AverageResolutionTime time.Duration
	TotalResolutionTime   time.Duration
	ResolutionCount       int64

	// Memory optimization metrics
	SharedInstanceCount int64
	DuplicatesPrevented int64

	// Has unexported fields.
}
    DependencyMetrics tracks metrics for dependency sharing and resolution


### DependencyPool

```go
type DependencyPool struct{ ... }
```


type DependencyPool struct {
	// Has unexported fields.
}
    DependencyPool manages a pool of reusable dependency instances

func NewDependencyPool(newFunc func() interface{}) *DependencyPool
func (dp *DependencyPool) Get() interface{}
func (dp *DependencyPool) Put(instance interface{})

### DependencyPoolManager

```go
type DependencyPoolManager struct{ ... }
```


type DependencyPoolManager struct {
	// Has unexported fields.
}
    DependencyPoolManager manages multiple dependency pools

func NewDependencyPoolManager() *DependencyPoolManager
func (dpm *DependencyPoolManager) GetFromPool(dependencyType string) (interface{}, bool)
func (dpm *DependencyPoolManager) RegisterPool(dependencyType string, newFunc func() interface{})
func (dpm *DependencyPoolManager) ReturnToPool(dependencyType string, instance interface{}) bool

### DependencyUpdate

```go
type DependencyUpdate struct{ ... }
```


type DependencyUpdate struct {
	DependencyType  string
	UpdateType      UpdateType
	OldVersion      string
	NewVersion      string
	OldInstance     interface{}
	NewInstance     interface{}
	Configuration   map[string]interface{}
	PropagationPath []string
	Timestamp       time.Time
	Error           error
	RequiresRestart bool
}
    DependencyUpdate represents an update to a dependency


### DependencyUpdateManager

```go
type DependencyUpdateManager struct{ ... }
```


type DependencyUpdateManager struct {
	// Has unexported fields.
}
    DependencyUpdateManager manages dependency updates and propagation

func NewDependencyUpdateManager(container *containerImpl) *DependencyUpdateManager
func (dum *DependencyUpdateManager) AddDependencyRelationship(dependency, dependent string)
func (dum *DependencyUpdateManager) AddUpdateListener(listener UpdateListener)
func (dum *DependencyUpdateManager) Close()
func (dum *DependencyUpdateManager) ForceRefresh(dependencyType string) error
func (dum *DependencyUpdateManager) GetDependencyGraph() map[string][]string
func (dum *DependencyUpdateManager) GetDependencyVersion(dependencyType string) (string, bool)
func (dum *DependencyUpdateManager) GetUpdateHistory() []DependencyUpdate
func (dum *DependencyUpdateManager) RegisterDependencyVersion(dependencyType string, version string)
func (dum *DependencyUpdateManager) SetBatchTimeout(timeout time.Duration)
func (dum *DependencyUpdateManager) SetUpdatePropagationStrategy(strategy UpdatePropagationStrategy)
func (dum *DependencyUpdateManager) TriggerConfigurationUpdate(dependencyType string, config map[string]interface{}) error
func (dum *DependencyUpdateManager) TriggerDependencyUpdate(dependencyType string, updateType UpdateType, newInstance interface{}, ...) error

### DependencyWithCircuitBreaker

```go
type DependencyWithCircuitBreaker struct{ ... }
```


type DependencyWithCircuitBreaker struct {
	// Has unexported fields.
}
    DependencyWithCircuitBreaker wraps a dependency with circuit breaker
    functionality

func NewDependencyWithCircuitBreaker(dependency interface{}, dependencyType string, breaker *CircuitBreaker) *DependencyWithCircuitBreaker
func (d *DependencyWithCircuitBreaker) Execute(operation func(interface{}) error) error
func (d *DependencyWithCircuitBreaker) GetDependency() (interface{}, error)
func (d *DependencyWithCircuitBreaker) GetState() CircuitState
func (d *DependencyWithCircuitBreaker) IsAvailable() bool

### HealthAwareCircuitBreakerListener

```go
type HealthAwareCircuitBreakerListener struct{ ... }
```


type HealthAwareCircuitBreakerListener struct {
	// Has unexported fields.
}
    HealthAwareCircuitBreakerListener implements HealthListener to integrate
    circuit breakers with health monitoring

func NewHealthAwareCircuitBreakerListener(cbm *CircuitBreakerManager) *HealthAwareCircuitBreakerListener
func (h *HealthAwareCircuitBreakerListener) OnHealthChanged(dependencyType string, oldStatus, newStatus HealthStatus, ...)

### HealthCheck

```go
type HealthCheck struct{ ... }
```


type HealthCheck struct {
	Status        HealthStatus           `json:"status"`
	Message       string                 `json:"message,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CheckDuration time.Duration          `json:"check_duration"`
}
    HealthCheck represents a health check result


### HealthChecker

```go
type HealthChecker interface{ ... }
```


type HealthChecker interface {
	// HealthCheck performs a health check and returns the result
	HealthCheck(ctx context.Context) HealthCheck
}
    HealthChecker interface for dependencies that can be health checked


### HealthEvent

```go
type HealthEvent struct{ ... }
```


type HealthEvent struct {
	DependencyType string
	OldStatus      HealthStatus
	NewStatus      HealthStatus
	HealthCheck    HealthCheck
	Timestamp      time.Time
}
    HealthEvent represents a health status change event


### HealthListener

```go
type HealthListener interface{ ... }
```


type HealthListener interface {
	OnHealthChanged(dependencyType string, oldStatus, newStatus HealthStatus, health HealthCheck)
}
    HealthListener interface for notifications about dependency health changes


### HealthMonitor

```go
type HealthMonitor struct{ ... }
```


type HealthMonitor struct {
	// Has unexported fields.
}
    HealthMonitor manages health checking for all dependencies

func NewHealthMonitor() *HealthMonitor
func (h *HealthMonitor) AddListener(listener HealthListener)
func (h *HealthMonitor) CheckAllHealth(ctx context.Context)
func (h *HealthMonitor) CheckHealth(ctx context.Context, dependencyType string) error
func (h *HealthMonitor) GetHealth(dependencyType string) (HealthCheck, bool)
func (h *HealthMonitor) GetOverallHealth() HealthCheck
func (h *HealthMonitor) GetUnhealthyDependencies() []string
func (h *HealthMonitor) IsHealthy(dependencyType string) bool
func (h *HealthMonitor) RegisterDependency(dependencyType string, instance interface{}, options ...HealthOption)
func (h *HealthMonitor) StartMonitoring(ctx context.Context)
func (h *HealthMonitor) StopMonitoring()

### HealthOption

```go
type HealthOption func(*DependencyHealth)
```


type HealthOption func(*DependencyHealth)
    HealthOption configures health monitoring for a dependency

func WithCheckInterval(interval time.Duration) HealthOption
func WithMaxFailures(maxFailures int) HealthOption

### HealthStatus

```go
type HealthStatus int
```


type HealthStatus int
    HealthStatus represents the health state of a dependency

const HealthUnknown HealthStatus = iota ...
func (h HealthStatus) String() string

### HotReloadable

```go
type HotReloadable interface{ ... }
```


type HotReloadable interface {
	Reload(ctx context.Context, config map[string]interface{}) error
}
    HotReloadable represents a dependency that supports hot reloading


### LifecycleAware

```go
type LifecycleAware interface{ ... }
```


type LifecycleAware interface {
	Startable
	Stoppable
}
    LifecycleAware interface for dependencies that need both startup and
    shutdown


### LifecycleEvent

```go
type LifecycleEvent struct{ ... }
```


type LifecycleEvent struct {
	DependencyType string
	OldState       LifecycleState
	NewState       LifecycleState
	Error          error
	Timestamp      time.Time
}
    LifecycleEvent represents a lifecycle state change event


### LifecycleListener

```go
type LifecycleListener interface{ ... }
```


type LifecycleListener interface {
	OnLifecycleChanged(event LifecycleEvent)
}
    LifecycleListener interface for notifications about lifecycle events


### LifecycleManager

```go
type LifecycleManager struct{ ... }
```


type LifecycleManager struct {
	// Has unexported fields.
}
    LifecycleManager manages centralized dependency lifecycle operations

func NewLifecycleManager() *LifecycleManager
func (lm *LifecycleManager) AddDependency(dependent, dependency string)
func (lm *LifecycleManager) AddListener(listener LifecycleListener)
func (lm *LifecycleManager) GetDependencyGraph() map[string][]string
func (lm *LifecycleManager) GetLifecycleState(dependencyType string) (LifecycleState, bool)
func (lm *LifecycleManager) RegisterDependency(dependencyType string, instance interface{}, options ...LifecycleOption)
func (lm *LifecycleManager) StartAll(ctx context.Context) error
func (lm *LifecycleManager) StopAll(ctx context.Context) error

### LifecycleManagerHealthListener

```go
type LifecycleManagerHealthListener struct{ ... }
```


type LifecycleManagerHealthListener struct {
	// Has unexported fields.
}
    LifecycleManagerHealthListener integrates lifecycle management with health
    monitoring

func NewLifecycleManagerHealthListener(lm *LifecycleManager) *LifecycleManagerHealthListener
func (l *LifecycleManagerHealthListener) OnHealthChanged(dependencyType string, oldStatus, newStatus HealthStatus, ...)

### LifecycleOption

```go
type LifecycleOption func(*DependencyLifecycle)
```


type LifecycleOption func(*DependencyLifecycle)
    LifecycleOption configures lifecycle management for a dependency

func WithDependencies(dependencies ...string) LifecycleOption
func WithPriority(priority LifecyclePriority) LifecycleOption

### LifecyclePriority

```go
type LifecyclePriority int
```


type LifecyclePriority int
    LifecyclePriority defines startup/shutdown ordering priority

const PriorityLowest LifecyclePriority = iota ...

### LifecycleState

```go
type LifecycleState int
```


type LifecycleState int
    LifecycleState represents the current state of a dependency lifecycle

const LifecycleStateUnknown LifecycleState = iota ...
func (l LifecycleState) String() string

### MemoryOptimizationReport

```go
type MemoryOptimizationReport struct{ ... }
```


type MemoryOptimizationReport struct {
	TotalAllocations  uint64                   `json:"total_allocations"`
	TotalShares       int64                    `json:"total_shares"`
	TotalCreations    int64                    `json:"total_creations"`
	SharingEfficiency float64                  `json:"sharing_efficiency"`
	TotalMemorySaved  uint64                   `json:"total_memory_saved"`
	DependencyDetails []DependencyMemoryDetail `json:"dependency_details"`
	Timestamp         time.Time                `json:"timestamp"`
}
    MemoryOptimizationReport provides detailed memory usage analysis


### MemoryProfiler

```go
type MemoryProfiler struct{ ... }
```


type MemoryProfiler struct {
	// Has unexported fields.
}
    MemoryProfiler provides memory profiling tools for dependency analysis

func NewMemoryProfiler(enabled bool) *MemoryProfiler
func (mp *MemoryProfiler) AnalyzeTrend() MemoryTrendAnalysis
func (mp *MemoryProfiler) GetLatestSnapshot() *MemorySnapshot
func (mp *MemoryProfiler) GetSnapshots() []MemorySnapshot
func (mp *MemoryProfiler) TakeSnapshot(memoryTracker *MemoryTracker)

### MemorySnapshot

```go
type MemorySnapshot struct{ ... }
```


type MemorySnapshot struct {
	Timestamp     time.Time               `json:"timestamp"`
	HeapAlloc     uint64                  `json:"heap_alloc"`
	HeapSys       uint64                  `json:"heap_sys"`
	NumGC         uint32                  `json:"num_gc"`
	DependencyMem uint64                  `json:"dependency_mem"`
	Dependencies  map[string]*MemoryStats `json:"dependencies"`
}
    MemorySnapshot captures memory state at a point in time


### MemoryStats

```go
type MemoryStats struct{ ... }
```


type MemoryStats struct {
	// AllocationSize is the estimated memory size of the dependency instance
	AllocationSize uint64
	// ReferenceCount is the number of active references to this dependency
	ReferenceCount int64
	// CreatedAt is when this dependency was first created
	CreatedAt time.Time
	// LastAccessed is the last time this dependency was accessed
	LastAccessed time.Time
	// ShareCount is the number of times this dependency was shared vs created new
	ShareCount int64
	// Type information
	DependencyType string
	Scope          Scope
}
    MemoryStats tracks memory usage for dependency sharing


### MemoryTracker

```go
type MemoryTracker struct{ ... }
```


type MemoryTracker struct {
	// Has unexported fields.
}
    MemoryTracker tracks memory usage and optimization metrics for dependencies

func NewMemoryTracker() *MemoryTracker
func (mt *MemoryTracker) GetMemoryOptimizationReport() MemoryOptimizationReport
func (mt *MemoryTracker) GetMemoryStats() map[string]*MemoryStats
func (mt *MemoryTracker) GetSharingEfficiency() float64
func (mt *MemoryTracker) GetTotalAllocations() uint64
func (mt *MemoryTracker) ReleaseDependencyReference(dependencyType string)
func (mt *MemoryTracker) TrackDependencyAccess(dependencyType string)
func (mt *MemoryTracker) TrackDependencyCreation(dependencyType string, instance interface{}, scope Scope)
func (mt *MemoryTracker) TrackDependencyShare(dependencyType string)

### MemoryTrendAnalysis

```go
type MemoryTrendAnalysis struct{ ... }
```


type MemoryTrendAnalysis struct {
	TrendDirection      string        `json:"trend_direction"`
	DependencyTrend     string        `json:"dependency_trend"`
	HeapAllocChange     int64         `json:"heap_alloc_change"`
	DependencyMemChange int64         `json:"dependency_mem_change"`
	TimeSpan            time.Duration `json:"time_span"`
	SnapshotCount       int           `json:"snapshot_count"`
}
    MemoryTrendAnalysis provides analysis of memory usage trends


### Scope

```go
type Scope int
```


type Scope int
    Scope defines the lifecycle management strategy for dependencies

const Singleton Scope = iota ...
func RecommendScope(interfaceType reflect.Type) Scope
func (s Scope) String() string

### SharedDependencyManager

```go
type SharedDependencyManager struct{ ... }
```


type SharedDependencyManager struct {
	// Has unexported fields.
}
    SharedDependencyManager manages scoped dependency lifecycle coordination

func NewSharedDependencyManager() *SharedDependencyManager
func (sdm *SharedDependencyManager) CheckHealth() map[reflect.Type]error
func (sdm *SharedDependencyManager) ClearScope(contextKey string)
func (sdm *SharedDependencyManager) GetExistingScoped(contextKey string, interfaceType reflect.Type) interface{}
func (sdm *SharedDependencyManager) GetExistingSingleton(interfaceType reflect.Type) interface{}
func (sdm *SharedDependencyManager) GetMetrics() DependencyMetrics
func (sdm *SharedDependencyManager) GetScoped(contextKey string, interfaceType reflect.Type, ...) (interface{}, error)
func (sdm *SharedDependencyManager) GetSingleton(interfaceType reflect.Type, factory func() (interface{}, error)) (interface{}, error)
func (sdm *SharedDependencyManager) RegisterHealthChecker(interfaceType reflect.Type, checker health.HealthChecker)
func (sdm *SharedDependencyManager) SetScoped(contextKey string, interfaceType reflect.Type, instance interface{})
func (sdm *SharedDependencyManager) SetSingleton(interfaceType reflect.Type, instance interface{})
func (sdm *SharedDependencyManager) StoreScoped(contextKey string, interfaceType reflect.Type, instance interface{}) error
func (sdm *SharedDependencyManager) StoreSingleton(interfaceType reflect.Type, instance interface{}) error

### Startable

```go
type Startable interface{ ... }
```


type Startable interface {
	Start(ctx context.Context) error
}
    Startable interface for dependencies that need startup logic


### Stoppable

```go
type Stoppable interface{ ... }
```


type Stoppable interface {
	Stop(ctx context.Context) error
}
    Stoppable interface for dependencies that need shutdown logic


### UpdateListener

```go
type UpdateListener interface{ ... }
```


type UpdateListener interface {
	OnDependencyUpdated(update DependencyUpdate)
}
    UpdateListener handles dependency update events


### UpdatePropagationStrategy

```go
type UpdatePropagationStrategy int
```


type UpdatePropagationStrategy int
    UpdatePropagationStrategy defines how updates should be propagated

const UpdatePropagationImmediate UpdatePropagationStrategy = iota ...
func (ups UpdatePropagationStrategy) String() string

### UpdateType

```go
type UpdateType int
```


type UpdateType int
    UpdateType represents the type of dependency update

const UpdateTypeUnknown UpdateType = iota ...
func (ut UpdateType) String() string

### ValidationError

```go
type ValidationError struct{ ... }
```


type ValidationError struct {
	// Errors contains all validation errors found
	Errors []error
	// Graph provides the dependency graph for debugging
	Graph map[string][]string
}
    ValidationError represents a dependency graph validation error

func NewValidationError(errors []error, graph map[string][]string) *ValidationError
func (e *ValidationError) Error() string

### Versionable

```go
type Versionable interface{ ... }
```


type Versionable interface {
	GetVersion() string
}
    Versionable represents a dependency that supports versioning


---

*Generated by [endor-sdk-go documentation generator](https://github.com/mattiabonardi/endor-sdk-go/tree/main/tools/gendocs)*
