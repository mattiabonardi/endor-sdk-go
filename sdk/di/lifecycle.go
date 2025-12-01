package di

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// LifecycleState represents the current state of a dependency lifecycle
type LifecycleState int

const (
	LifecycleStateUnknown LifecycleState = iota
	LifecycleStateRegistered
	LifecycleStateStarting
	LifecycleStateStarted
	LifecycleStateStopping
	LifecycleStateStopped
	LifecycleStateError
)

func (l LifecycleState) String() string {
	switch l {
	case LifecycleStateRegistered:
		return "Registered"
	case LifecycleStateStarting:
		return "Starting"
	case LifecycleStateStarted:
		return "Started"
	case LifecycleStateStopping:
		return "Stopping"
	case LifecycleStateStopped:
		return "Stopped"
	case LifecycleStateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Startable interface for dependencies that need startup logic
type Startable interface {
	Start(ctx context.Context) error
}

// Stoppable interface for dependencies that need shutdown logic
type Stoppable interface {
	Stop(ctx context.Context) error
}

// LifecycleAware interface for dependencies that need both startup and shutdown
type LifecycleAware interface {
	Startable
	Stoppable
}

// LifecyclePriority defines startup/shutdown ordering priority
type LifecyclePriority int

const (
	PriorityLowest LifecyclePriority = iota // Start last, stop first
	PriorityLow
	PriorityNormal
	PriorityHigh
	PriorityHighest // Start first, stop last
)

// DependencyLifecycle tracks lifecycle state for a dependency
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
	mu             sync.RWMutex
}

// LifecycleEvent represents a lifecycle state change event
type LifecycleEvent struct {
	DependencyType string
	OldState       LifecycleState
	NewState       LifecycleState
	Error          error
	Timestamp      time.Time
}

// LifecycleListener interface for notifications about lifecycle events
type LifecycleListener interface {
	OnLifecycleChanged(event LifecycleEvent)
}

// LifecycleManager manages centralized dependency lifecycle operations
type LifecycleManager struct {
	dependencies map[string]*DependencyLifecycle
	listeners    []LifecycleListener
	started      bool
	mu           sync.RWMutex
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		dependencies: make(map[string]*DependencyLifecycle),
		listeners:    make([]LifecycleListener, 0),
	}
}

// RegisterDependency registers a dependency for lifecycle management
func (lm *LifecycleManager) RegisterDependency(dependencyType string, instance interface{}, options ...LifecycleOption) {
	lifecycle := &DependencyLifecycle{
		DependencyType: dependencyType,
		Instance:       instance,
		State:          LifecycleStateRegistered,
		Priority:       PriorityNormal,
		Dependencies:   make([]string, 0),
		Dependents:     make([]string, 0),
	}

	// Apply options
	for _, option := range options {
		option(lifecycle)
	}

	lm.mu.Lock()
	lm.dependencies[dependencyType] = lifecycle
	lm.mu.Unlock()

	// Notify listeners after releasing the lock
	lm.notifyListeners(LifecycleEvent{
		DependencyType: dependencyType,
		OldState:       LifecycleStateUnknown,
		NewState:       LifecycleStateRegistered,
		Timestamp:      time.Now(),
	})
}

// AddDependency adds a dependency relationship
func (lm *LifecycleManager) AddDependency(dependent, dependency string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if dependentLifecycle, exists := lm.dependencies[dependent]; exists {
		dependentLifecycle.Dependencies = append(dependentLifecycle.Dependencies, dependency)
	}

	if dependencyLifecycle, exists := lm.dependencies[dependency]; exists {
		dependencyLifecycle.Dependents = append(dependencyLifecycle.Dependents, dependent)
	}
}

// StartAll starts all dependencies in proper order based on dependency graph
func (lm *LifecycleManager) StartAll(ctx context.Context) error {
	lm.mu.Lock()
	if lm.started {
		lm.mu.Unlock()
		return fmt.Errorf("lifecycle manager already started")
	}
	lm.started = true
	lm.mu.Unlock()

	// Calculate startup order
	startupOrder, err := lm.calculateStartupOrder()
	if err != nil {
		lm.mu.Lock()
		lm.started = false
		lm.mu.Unlock()
		return fmt.Errorf("failed to calculate startup order: %w", err)
	}

	// Start dependencies in order
	for _, dependencyType := range startupOrder {
		if err := lm.startDependency(ctx, dependencyType); err != nil {
			// Attempt to stop already started dependencies
			lm.stopStartedDependencies(ctx)
			lm.mu.Lock()
			lm.started = false
			lm.mu.Unlock()
			return fmt.Errorf("failed to start dependency %s: %w", dependencyType, err)
		}
	}

	return nil
}

// StopAll stops all dependencies in reverse startup order
func (lm *LifecycleManager) StopAll(ctx context.Context) error {
	lm.mu.Lock()
	if !lm.started {
		lm.mu.Unlock()
		return nil // Already stopped
	}
	lm.started = false
	lm.mu.Unlock()

	// Calculate shutdown order (reverse of startup)
	startupOrder, err := lm.calculateStartupOrder()
	if err != nil {
		return fmt.Errorf("failed to calculate shutdown order: %w", err)
	}

	// Reverse the order for shutdown
	shutdownOrder := make([]string, len(startupOrder))
	for i, dep := range startupOrder {
		shutdownOrder[len(startupOrder)-1-i] = dep
	}

	var lastError error
	// Stop all dependencies, collecting errors but continuing
	for _, dependencyType := range shutdownOrder {
		if err := lm.stopDependency(ctx, dependencyType); err != nil {
			lastError = err
		}
	}

	return lastError
}

// GetLifecycleState returns the current lifecycle state of a dependency
func (lm *LifecycleManager) GetLifecycleState(dependencyType string) (LifecycleState, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lifecycle, exists := lm.dependencies[dependencyType]; exists {
		lifecycle.mu.RLock()
		state := lifecycle.State
		lifecycle.mu.RUnlock()
		return state, true
	}
	return LifecycleStateUnknown, false
}

// GetDependencyGraph returns the dependency graph for analysis
func (lm *LifecycleManager) GetDependencyGraph() map[string][]string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	graph := make(map[string][]string)
	for name, lifecycle := range lm.dependencies {
		lifecycle.mu.RLock()
		deps := make([]string, len(lifecycle.Dependencies))
		copy(deps, lifecycle.Dependencies)
		lifecycle.mu.RUnlock()
		graph[name] = deps
	}
	return graph
}

// AddListener adds a lifecycle event listener
func (lm *LifecycleManager) AddListener(listener LifecycleListener) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.listeners = append(lm.listeners, listener)
}

// calculateStartupOrder determines the proper startup order using topological sort
func (lm *LifecycleManager) calculateStartupOrder() ([]string, error) {
	// Build adjacency list and in-degree map
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize
	for name := range lm.dependencies {
		graph[name] = make([]string, 0)
		inDegree[name] = 0
	}

	// Build graph and count in-degrees
	for name, lifecycle := range lm.dependencies {
		lifecycle.mu.RLock()
		for _, dep := range lifecycle.Dependencies {
			if _, exists := lm.dependencies[dep]; exists {
				graph[dep] = append(graph[dep], name)
				inDegree[name]++
			}
		}
		lifecycle.mu.RUnlock()
	}

	// Group by priority for stable ordering within priority levels
	priorityGroups := make(map[LifecyclePriority][]string)
	for name, lifecycle := range lm.dependencies {
		lifecycle.mu.RLock()
		priority := lifecycle.Priority
		lifecycle.mu.RUnlock()
		priorityGroups[priority] = append(priorityGroups[priority], name)
	}

	// Sort priorities (highest first for startup)
	priorities := []LifecyclePriority{PriorityHighest, PriorityHigh, PriorityNormal, PriorityLow, PriorityLowest}

	var result []string
	queue := make([]string, 0)

	// Add nodes with no dependencies, ordered by priority
	for _, priority := range priorities {
		if deps, exists := priorityGroups[priority]; exists {
			sort.Strings(deps) // Stable ordering within priority
			for _, dep := range deps {
				if inDegree[dep] == 0 {
					queue = append(queue, dep)
				}
			}
		}
	}

	// Topological sort
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree for dependents
		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				// Insert in priority order
				inserted := false
				depLifecycle := lm.dependencies[dependent]
				depLifecycle.mu.RLock()
				currentPriority := depLifecycle.Priority
				depLifecycle.mu.RUnlock()

				for i, existing := range queue {
					existingLifecycle := lm.dependencies[existing]
					existingLifecycle.mu.RLock()
					existingPriority := existingLifecycle.Priority
					existingLifecycle.mu.RUnlock()

					if currentPriority > existingPriority ||
						(currentPriority == existingPriority && dependent < existing) {
						queue = append(queue[:i], append([]string{dependent}, queue[i:]...)...)
						inserted = true
						break
					}
				}
				if !inserted {
					queue = append(queue, dependent)
				}
			}
		}
	}

	// Check for circular dependencies
	if len(result) != len(lm.dependencies) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// startDependency starts a single dependency
func (lm *LifecycleManager) startDependency(ctx context.Context, dependencyType string) error {
	lm.mu.RLock()
	lifecycle, exists := lm.dependencies[dependencyType]
	lm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("dependency %s not found", dependencyType)
	}

	lifecycle.mu.Lock()
	if lifecycle.State != LifecycleStateRegistered {
		lifecycle.mu.Unlock()
		return nil // Already started or in error state
	}

	lifecycle.State = LifecycleStateStarting
	lifecycle.mu.Unlock()

	lm.notifyListeners(LifecycleEvent{
		DependencyType: dependencyType,
		OldState:       LifecycleStateRegistered,
		NewState:       LifecycleStateStarting,
		Timestamp:      time.Now(),
	})

	var err error
	if startable, ok := lifecycle.Instance.(Startable); ok {
		err = startable.Start(ctx)
	}

	lifecycle.mu.Lock()
	if err != nil {
		lifecycle.State = LifecycleStateError
		lifecycle.LastError = err
	} else {
		lifecycle.State = LifecycleStateStarted
		lifecycle.StartedAt = time.Now()
	}
	newState := lifecycle.State
	lifecycle.mu.Unlock()

	lm.notifyListeners(LifecycleEvent{
		DependencyType: dependencyType,
		OldState:       LifecycleStateStarting,
		NewState:       newState,
		Error:          err,
		Timestamp:      time.Now(),
	})

	return err
}

// stopDependency stops a single dependency
func (lm *LifecycleManager) stopDependency(ctx context.Context, dependencyType string) error {
	lm.mu.RLock()
	lifecycle, exists := lm.dependencies[dependencyType]
	lm.mu.RUnlock()

	if !exists {
		return nil
	}

	lifecycle.mu.Lock()
	if lifecycle.State != LifecycleStateStarted {
		lifecycle.mu.Unlock()
		return nil // Not started or already stopped
	}

	lifecycle.State = LifecycleStateStopping
	lifecycle.mu.Unlock()

	lm.notifyListeners(LifecycleEvent{
		DependencyType: dependencyType,
		OldState:       LifecycleStateStarted,
		NewState:       LifecycleStateStopping,
		Timestamp:      time.Now(),
	})

	var err error
	if stoppable, ok := lifecycle.Instance.(Stoppable); ok {
		err = stoppable.Stop(ctx)
	}

	lifecycle.mu.Lock()
	if err != nil {
		lifecycle.State = LifecycleStateError
		lifecycle.LastError = err
	} else {
		lifecycle.State = LifecycleStateStopped
		lifecycle.StoppedAt = time.Now()
	}
	newState := lifecycle.State
	lifecycle.mu.Unlock()

	lm.notifyListeners(LifecycleEvent{
		DependencyType: dependencyType,
		OldState:       LifecycleStateStopping,
		NewState:       newState,
		Error:          err,
		Timestamp:      time.Now(),
	})

	return err
}

// stopStartedDependencies stops all dependencies that were successfully started
func (lm *LifecycleManager) stopStartedDependencies(ctx context.Context) {
	lm.mu.RLock()
	var startedDeps []string
	for name, lifecycle := range lm.dependencies {
		lifecycle.mu.RLock()
		if lifecycle.State == LifecycleStateStarted {
			startedDeps = append(startedDeps, name)
		}
		lifecycle.mu.RUnlock()
	}
	lm.mu.RUnlock()

	// Stop in reverse order
	for i := len(startedDeps) - 1; i >= 0; i-- {
		lm.stopDependency(ctx, startedDeps[i])
	}
}

// notifyListeners notifies all registered listeners about lifecycle events
func (lm *LifecycleManager) notifyListeners(event LifecycleEvent) {
	lm.mu.RLock()
	listeners := make([]LifecycleListener, len(lm.listeners))
	copy(listeners, lm.listeners)
	lm.mu.RUnlock()

	for _, listener := range listeners {
		// Call listener directly with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Log panic but don't let it crash the lifecycle manager
					fmt.Printf("Lifecycle listener panicked: %v\n", r)
				}
			}()
			listener.OnLifecycleChanged(event)
		}()
	}
}

// LifecycleOption configures lifecycle management for a dependency
type LifecycleOption func(*DependencyLifecycle)

// WithPriority sets the lifecycle priority
func WithPriority(priority LifecyclePriority) LifecycleOption {
	return func(l *DependencyLifecycle) {
		l.Priority = priority
	}
}

// WithDependencies sets the dependencies for this component
func WithDependencies(dependencies ...string) LifecycleOption {
	return func(l *DependencyLifecycle) {
		l.Dependencies = append(l.Dependencies, dependencies...)
	}
}

// LifecycleManagerHealthListener integrates lifecycle management with health monitoring
type LifecycleManagerHealthListener struct {
	lifecycleManager *LifecycleManager
}

// NewLifecycleManagerHealthListener creates a new health-aware lifecycle listener
func NewLifecycleManagerHealthListener(lm *LifecycleManager) *LifecycleManagerHealthListener {
	return &LifecycleManagerHealthListener{
		lifecycleManager: lm,
	}
}

// OnHealthChanged implements HealthListener interface
func (l *LifecycleManagerHealthListener) OnHealthChanged(dependencyType string, oldStatus, newStatus HealthStatus, healthCheck HealthCheck) {
	// If a critical dependency becomes unhealthy, consider lifecycle implications
	if newStatus == HealthUnhealthy {
		if state, exists := l.lifecycleManager.GetLifecycleState(dependencyType); exists && state == LifecycleStateStarted {
			// Could trigger shutdown of dependent services here in a more advanced implementation
			fmt.Printf("Critical dependency %s became unhealthy while started\n", dependencyType)
		}
	}
}
