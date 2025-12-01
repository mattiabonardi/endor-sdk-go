package lifecycle

import (
	"context"
	"fmt"
	"sync"
)

// DependencyGraph represents the dependency relationships between services
type DependencyGraph struct {
	// Nodes maps service names to their dependencies
	Nodes map[string][]string
	// ReverseDependencies maps service names to services that depend on them
	ReverseDependencies map[string][]string
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes:               make(map[string][]string),
		ReverseDependencies: make(map[string][]string),
	}
}

// AddService adds a service to the dependency graph without any dependencies
func (dg *DependencyGraph) AddService(serviceName string) {
	if _, exists := dg.Nodes[serviceName]; !exists {
		dg.Nodes[serviceName] = make([]string, 0)
	}
}

// AddDependency adds a dependency relationship (dependent depends on dependency)
func (dg *DependencyGraph) AddDependency(dependent, dependency string) {
	// Add to forward dependencies
	if deps, exists := dg.Nodes[dependent]; exists {
		// Check if dependency already exists to avoid duplicates
		for _, existing := range deps {
			if existing == dependency {
				return
			}
		}
		dg.Nodes[dependent] = append(deps, dependency)
	} else {
		dg.Nodes[dependent] = []string{dependency}
	}

	// Add to reverse dependencies
	if deps, exists := dg.ReverseDependencies[dependency]; exists {
		// Check if dependent already exists to avoid duplicates
		for _, existing := range deps {
			if existing == dependent {
				return
			}
		}
		dg.ReverseDependencies[dependency] = append(deps, dependent)
	} else {
		dg.ReverseDependencies[dependency] = []string{dependent}
	}

	// Ensure both services exist in nodes map
	if _, exists := dg.Nodes[dependency]; !exists {
		dg.Nodes[dependency] = make([]string, 0)
	}
}

// TopologicalSort returns services in dependency order (dependencies first)
func (dg *DependencyGraph) TopologicalSort() ([]string, error) {
	// Kahn's algorithm for topological sorting
	inDegree := make(map[string]int)

	// Initialize in-degree for all nodes
	for node := range dg.Nodes {
		inDegree[node] = 0
	}

	// Calculate in-degrees (count incoming edges)
	for node, dependencies := range dg.Nodes {
		for range dependencies {
			inDegree[node]++
		}
	}

	// Find all nodes with no incoming edges (no dependencies)
	queue := make([]string, 0)
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	result := make([]string, 0, len(dg.Nodes))

	for len(queue) > 0 {
		// Remove a node from the queue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// For each service that depends on the current node,
		// reduce its in-degree and add to queue if it reaches 0
		for _, dependent := range dg.ReverseDependencies[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycles
	if len(result) != len(dg.Nodes) {
		return nil, fmt.Errorf("circular dependency detected in service graph")
	}

	return result, nil
} // GetDependencies returns the direct dependencies of a service
func (dg *DependencyGraph) GetDependencies(serviceName string) []string {
	deps, exists := dg.Nodes[serviceName]
	if !exists {
		return make([]string, 0)
	}

	// Return a copy to prevent external modification
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// GetDependents returns the services that depend on a given service
func (dg *DependencyGraph) GetDependents(serviceName string) []string {
	deps, exists := dg.ReverseDependencies[serviceName]
	if !exists {
		return make([]string, 0)
	}

	// Return a copy to prevent external modification
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// LifecycleManager coordinates the lifecycle of multiple services with dependency management
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

// DefaultLifecycleManager is the default implementation of LifecycleManager
type DefaultLifecycleManager struct {
	services        map[string]ServiceLifecycleInterface
	serviceStates   map[string]ServiceState
	dependencyGraph *DependencyGraph
	healthMonitor   *HealthMonitor
	hookManager     *HookManager
	startupOrder    []string
	mu              sync.RWMutex
	started         bool
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(healthPolicy HealthAggregationPolicy) LifecycleManager {
	return &DefaultLifecycleManager{
		services:        make(map[string]ServiceLifecycleInterface),
		serviceStates:   make(map[string]ServiceState),
		dependencyGraph: NewDependencyGraph(),
		healthMonitor:   NewHealthMonitor(healthPolicy),
		hookManager:     NewHookManager(DefaultHookConfiguration()),
	}
}

// RegisterService registers a service for lifecycle management
func (lm *DefaultLifecycleManager) RegisterService(name string, service ServiceLifecycleInterface) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Check if service is already registered
	if _, exists := lm.services[name]; exists {
		return fmt.Errorf("service '%s' is already registered", name)
	}

	lm.services[name] = service
	lm.serviceStates[name] = Created

	// Add service to dependency graph
	lm.dependencyGraph.AddService(name)

	// Register with health monitor
	healthConfig := DefaultHealthConfiguration()
	lm.healthMonitor.RegisterService(name, service, healthConfig)

	// Clear cached startup order since dependencies may have changed
	lm.startupOrder = nil

	return nil
}

// AddServiceDependency adds a dependency relationship between services
func (lm *DefaultLifecycleManager) AddServiceDependency(dependent, dependency string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Validate that both services are registered
	if _, exists := lm.services[dependent]; !exists {
		return fmt.Errorf("dependent service '%s' is not registered", dependent)
	}

	if _, exists := lm.services[dependency]; !exists {
		return fmt.Errorf("dependency service '%s' is not registered", dependency)
	}

	// Add dependency to graph
	lm.dependencyGraph.AddDependency(dependent, dependency)

	// Clear cached startup order since dependencies have changed
	lm.startupOrder = nil

	return nil
}

// calculateStartupOrder calculates the order in which services should be started
func (lm *DefaultLifecycleManager) calculateStartupOrder() ([]string, error) {
	if lm.startupOrder != nil {
		return lm.startupOrder, nil
	}

	order, err := lm.dependencyGraph.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate startup order: %w", err)
	}

	lm.startupOrder = order
	return order, nil
}

// StartAll starts all services in dependency order
func (lm *DefaultLifecycleManager) StartAll(ctx context.Context) error {
	lm.mu.Lock()
	if lm.started {
		lm.mu.Unlock()
		return fmt.Errorf("services are already started")
	}
	lm.started = true
	lm.mu.Unlock()

	// Calculate startup order
	startupOrder, err := lm.calculateStartupOrder()
	if err != nil {
		lm.mu.Lock()
		lm.started = false
		lm.mu.Unlock()
		return err
	}

	// Start services in dependency order
	for _, serviceName := range startupOrder {
		if err := lm.startService(ctx, serviceName); err != nil {
			// Attempt to stop already started services
			lm.stopStartedServices(ctx)
			lm.mu.Lock()
			lm.started = false
			lm.mu.Unlock()
			return fmt.Errorf("failed to start service '%s': %w", serviceName, err)
		}
	}

	// Start health monitoring
	if err := lm.healthMonitor.StartMonitoring(ctx); err != nil {
		// Services are started but health monitoring failed
		return fmt.Errorf("services started but health monitoring failed: %w", err)
	}

	return nil
}

// startService starts a single service with hooks
func (lm *DefaultLifecycleManager) startService(ctx context.Context, serviceName string) error {
	lm.mu.Lock()
	service, exists := lm.services[serviceName]
	if !exists {
		lm.mu.Unlock()
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	currentState := lm.serviceStates[serviceName]
	if currentState != Created && currentState != Stopped {
		lm.mu.Unlock()
		return fmt.Errorf("service '%s' is in state %s, cannot start", serviceName, currentState.String())
	}

	lm.serviceStates[serviceName] = Starting
	lm.mu.Unlock()

	// Execute before start hooks
	if err := lm.hookManager.ExecuteHooks(ctx, BeforeStartPhase, service); err != nil {
		lm.mu.Lock()
		lm.serviceStates[serviceName] = Failed
		lm.mu.Unlock()
		return fmt.Errorf("before start hooks failed for service '%s': %w", serviceName, err)
	}

	// Start the service
	if err := service.Start(ctx); err != nil {
		lm.mu.Lock()
		lm.serviceStates[serviceName] = Failed
		lm.mu.Unlock()
		return fmt.Errorf("failed to start service '%s': %w", serviceName, err)
	}

	// Execute after start hooks
	if err := lm.hookManager.ExecuteHooks(ctx, AfterStartPhase, service); err != nil {
		// Service started but hooks failed - attempt to stop the service
		_ = service.Stop(ctx)
		lm.mu.Lock()
		lm.serviceStates[serviceName] = Failed
		lm.mu.Unlock()
		return fmt.Errorf("after start hooks failed for service '%s': %w", serviceName, err)
	}

	lm.mu.Lock()
	lm.serviceStates[serviceName] = Running
	lm.mu.Unlock()

	return nil
}

// StopAll stops all services in reverse dependency order
func (lm *DefaultLifecycleManager) StopAll(ctx context.Context) error {
	lm.mu.Lock()
	if !lm.started {
		lm.mu.Unlock()
		return nil // Already stopped
	}
	lm.started = false
	lm.mu.Unlock()

	// Stop health monitoring first
	_ = lm.healthMonitor.StopMonitoring()

	// Get startup order and reverse it for shutdown
	startupOrder, err := lm.calculateStartupOrder()
	if err != nil {
		return fmt.Errorf("failed to calculate shutdown order: %w", err)
	}

	// Reverse the order for shutdown (dependents stop before dependencies)
	shutdownOrder := make([]string, len(startupOrder))
	for i, j := 0, len(startupOrder)-1; i <= j; i, j = i+1, j-1 {
		shutdownOrder[i], shutdownOrder[j] = startupOrder[j], startupOrder[i]
	}

	var lastErr error

	// Stop services in reverse dependency order
	for _, serviceName := range shutdownOrder {
		if err := lm.stopService(ctx, serviceName); err != nil {
			// Continue stopping other services but remember the error
			lastErr = fmt.Errorf("failed to stop service '%s': %w", serviceName, err)
		}
	}

	return lastErr
}

// stopService stops a single service with hooks
func (lm *DefaultLifecycleManager) stopService(ctx context.Context, serviceName string) error {
	lm.mu.Lock()
	service, exists := lm.services[serviceName]
	if !exists {
		lm.mu.Unlock()
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	currentState := lm.serviceStates[serviceName]
	if currentState != Running && currentState != Failed {
		lm.mu.Unlock()
		return nil // Already stopped or not startable
	}

	lm.serviceStates[serviceName] = Stopping
	lm.mu.Unlock()

	// Execute before stop hooks
	if err := lm.hookManager.ExecuteHooks(ctx, BeforeStopPhase, service); err != nil {
		// Log error but continue with stopping
		fmt.Printf("Warning: before stop hooks failed for service '%s': %v\n", serviceName, err)
	}

	// Stop the service
	if err := service.Stop(ctx); err != nil {
		lm.mu.Lock()
		lm.serviceStates[serviceName] = Failed
		lm.mu.Unlock()
		return fmt.Errorf("failed to stop service '%s': %w", serviceName, err)
	}

	// Execute after stop hooks
	if err := lm.hookManager.ExecuteHooks(ctx, AfterStopPhase, service); err != nil {
		// Log error but don't fail the stop operation
		fmt.Printf("Warning: after stop hooks failed for service '%s': %v\n", serviceName, err)
	}

	lm.mu.Lock()
	lm.serviceStates[serviceName] = Stopped
	lm.mu.Unlock()

	return nil
}

// stopStartedServices stops all services that are currently running (used during startup failure)
func (lm *DefaultLifecycleManager) stopStartedServices(ctx context.Context) {
	lm.mu.RLock()
	runningServices := make([]string, 0)
	for serviceName, state := range lm.serviceStates {
		if state == Running {
			runningServices = append(runningServices, serviceName)
		}
	}
	lm.mu.RUnlock()

	// Sort running services in reverse dependency order for shutdown
	startupOrder, _ := lm.calculateStartupOrder()
	shutdownOrder := make([]string, 0)

	// Create shutdown order from startup order (reverse)
	for i := len(startupOrder) - 1; i >= 0; i-- {
		serviceName := startupOrder[i]
		for _, running := range runningServices {
			if serviceName == running {
				shutdownOrder = append(shutdownOrder, serviceName)
				break
			}
		}
	}

	// Stop services
	for _, serviceName := range shutdownOrder {
		_ = lm.stopService(ctx, serviceName)
	}
}

// GetHealth returns the composite health status of all services
func (lm *DefaultLifecycleManager) GetHealth(ctx context.Context) CompositeHealthStatus {
	return lm.healthMonitor.CheckCompositeHealth(ctx)
}

// GetDependencyGraph returns the current dependency graph
func (lm *DefaultLifecycleManager) GetDependencyGraph() *DependencyGraph {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Return a copy to prevent external modification
	graph := NewDependencyGraph()
	for service, deps := range lm.dependencyGraph.Nodes {
		// Add the service first (even if it has no dependencies)
		graph.AddService(service)

		// Then add its dependencies
		for _, dep := range deps {
			graph.AddDependency(service, dep)
		}
	}

	return graph
}

// GetServiceState returns the current state of a specific service
func (lm *DefaultLifecycleManager) GetServiceState(serviceName string) (ServiceState, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	state, exists := lm.serviceStates[serviceName]
	if !exists {
		return Created, fmt.Errorf("service '%s' is not registered", serviceName)
	}

	return state, nil
}

// RestartService restarts a specific service and its dependents
func (lm *DefaultLifecycleManager) RestartService(ctx context.Context, serviceName string) error {
	lm.mu.RLock()
	_, exists := lm.services[serviceName]
	if !exists {
		lm.mu.RUnlock()
		return fmt.Errorf("service '%s' is not registered", serviceName)
	}
	lm.mu.RUnlock()

	// Find all services that depend on this one (directly or indirectly)
	dependents := lm.findAllDependents(serviceName)

	// Include the target service itself
	servicesToRestart := append([]string{serviceName}, dependents...)

	// Stop services in reverse dependency order (dependents first)
	for i := len(servicesToRestart) - 1; i >= 0; i-- {
		if err := lm.stopService(ctx, servicesToRestart[i]); err != nil {
			return fmt.Errorf("failed to stop service '%s' during restart: %w", servicesToRestart[i], err)
		}
	}

	// Start services in dependency order (dependencies first)
	for _, service := range servicesToRestart {
		if err := lm.startService(ctx, service); err != nil {
			return fmt.Errorf("failed to start service '%s' during restart: %w", service, err)
		}
	}

	return nil
}

// findAllDependents recursively finds all services that depend on the given service
func (lm *DefaultLifecycleManager) findAllDependents(serviceName string) []string {
	visited := make(map[string]bool)
	dependents := make([]string, 0)

	lm.findDependentsRecursive(serviceName, visited, &dependents)

	return dependents
}

// findDependentsRecursive recursively finds dependents
func (lm *DefaultLifecycleManager) findDependentsRecursive(serviceName string, visited map[string]bool, dependents *[]string) {
	if visited[serviceName] {
		return
	}
	visited[serviceName] = true

	// Get direct dependents
	directDependents := lm.dependencyGraph.GetDependents(serviceName)

	for _, dependent := range directDependents {
		if !visited[dependent] {
			*dependents = append(*dependents, dependent)
			lm.findDependentsRecursive(dependent, visited, dependents)
		}
	}
}
