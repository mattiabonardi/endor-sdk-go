package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// containerImpl implements the Container interface
type containerImpl struct {
	// registrations stores all dependency registrations by interface type string
	registrations map[string]*registration
	// singletons stores singleton instances by interface type string
	singletons map[string]interface{}
	// sharedManager manages scoped and singleton dependency sharing
	sharedManager *SharedDependencyManager
	// healthMonitor manages dependency health monitoring
	healthMonitor *HealthMonitor
	// circuitBreakerManager manages circuit breakers for dependencies
	circuitBreakerManager *CircuitBreakerManager
	// memoryTracker tracks memory usage optimization
	memoryTracker *MemoryTracker
	// memoryProfiler profiles memory usage patterns
	memoryProfiler *MemoryProfiler
	// poolManager manages dependency pools for expensive resources
	poolManager *DependencyPoolManager
	// lifecycleManager manages dependency startup and shutdown ordering
	lifecycleManager *LifecycleManager
	// updateManager manages dependency updates and propagation
	updateManager *DependencyUpdateManager
	// suggestionEngine provides intelligent error suggestions
	suggestionEngine *SuggestionEngine
	// mutex protects concurrent access to container state
	mutex sync.RWMutex
}

// newContainerImpl creates a new container implementation
func newContainerImpl() Container {
	healthMonitor := NewHealthMonitor()
	circuitBreakerManager := NewCircuitBreakerManager(DefaultCircuitBreakerConfig())
	memoryTracker := NewMemoryTracker()
	memoryProfiler := NewMemoryProfiler(true) // Enable profiling by default
	poolManager := NewDependencyPoolManager()
	lifecycleManager := NewLifecycleManager()
	suggestionEngine := NewSuggestionEngine()

	// Add circuit breaker listener to health monitor
	cbListener := NewHealthAwareCircuitBreakerListener(circuitBreakerManager)
	healthMonitor.AddListener(cbListener)

	container := &containerImpl{
		registrations:         make(map[string]*registration),
		singletons:            make(map[string]interface{}),
		sharedManager:         NewSharedDependencyManager(),
		healthMonitor:         healthMonitor,
		circuitBreakerManager: circuitBreakerManager,
		memoryTracker:         memoryTracker,
		memoryProfiler:        memoryProfiler,
		poolManager:           poolManager,
		lifecycleManager:      lifecycleManager,
		suggestionEngine:      suggestionEngine,
	}

	// Initialize update manager after container is created
	updateManager := NewDependencyUpdateManager(container)
	container.updateManager = updateManager

	return container
}

// updateSuggestionEngine updates the suggestion engine with current registrations
func (c *containerImpl) updateSuggestionEngine() {
	registeredTypes := make(map[string]reflect.Type)
	for typeKey, reg := range c.registrations {
		registeredTypes[typeKey] = reg.interfaceType
	}
	c.suggestionEngine.UpdateRegisteredTypes(registeredTypes)
}

// RegisterType registers an implementation for an interface type
func (c *containerImpl) RegisterType(interfaceType reflect.Type, impl interface{}, scope Scope) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Validate that interfaceType is actually an interface
	if interfaceType.Kind() != reflect.Interface {
		diErr := NewDIError(
			interfaceType,
			"registration",
			"type is not an interface",
		).WithContext("type_kind", interfaceType.Kind().String()).
			WithContext("type_name", interfaceType.String())

		suggestions := []string{
			"Ensure you're registering an interface type, not a concrete type",
			"Use reflect.TypeOf((*MyInterface)(nil)).Elem() to get the interface type",
			"Check that your type declaration uses 'interface{}' keyword",
		}
		return diErr.WithSuggestions(suggestions)
	}

	// Validate that impl implements the interface
	implType := reflect.TypeOf(impl)
	if !implType.Implements(interfaceType) {
		diErr := NewDIError(
			interfaceType,
			"registration",
			"implementation does not implement the interface",
		).WithContext("interface_type", interfaceType.String()).
			WithContext("implementation_type", implType.String())

		suggestions := []string{
			fmt.Sprintf("Ensure %s implements all methods of %s", implType.String(), interfaceType.String()),
			"Check method signatures match exactly (including receiver types)",
			"Verify generic type parameters are correctly specified",
		}
		return diErr.WithSuggestions(suggestions)
	}

	typeKey := interfaceType.String()
	c.registrations[typeKey] = &registration{
		instance:      impl,
		scope:         scope,
		interfaceType: interfaceType,
		isFactory:     false,
	}

	// Update suggestion engine with new registration
	c.updateSuggestionEngine()

	return nil
}

// RegisterFactoryType registers a factory function for an interface type
func (c *containerImpl) RegisterFactoryType(interfaceType reflect.Type, factory interface{}, scope Scope) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Validate that interfaceType is actually an interface
	if interfaceType.Kind() != reflect.Interface {
		diErr := NewDIError(
			interfaceType,
			"factory registration",
			"type is not an interface",
		).WithContext("type_kind", interfaceType.Kind().String()).
			WithContext("type_name", interfaceType.String())

		suggestions := []string{
			"Ensure you're registering an interface type, not a concrete type",
			"Use reflect.TypeOf((*MyInterface)(nil)).Elem() to get the interface type",
		}
		return diErr.WithSuggestions(suggestions)
	}

	// Validate factory function signature
	factoryType := reflect.TypeOf(factory)
	if factoryType.Kind() != reflect.Func {
		diErr := NewDIError(
			interfaceType,
			"factory registration",
			"factory must be a function",
		).WithContext("factory_type", factoryType.String())

		suggestions := []string{
			"Provide a function with signature: func(Container) (YourInterface, error)",
			"Factory functions must be functions, not other types",
		}
		return diErr.WithSuggestions(suggestions)
	}

	// Factory should accept Container and return (T, error)
	if factoryType.NumIn() != 1 || factoryType.NumOut() != 2 {
		diErr := NewDIError(
			interfaceType,
			"factory registration",
			"factory must have signature func(Container) (T, error)",
		).WithContext("factory_signature", factoryType.String()).
			WithContext("expected_in", 1).
			WithContext("actual_in", factoryType.NumIn()).
			WithContext("expected_out", 2).
			WithContext("actual_out", factoryType.NumOut())

		suggestions := []string{
			fmt.Sprintf("Correct signature: func(Container) (%s, error)", getShortTypeName(interfaceType.String())),
			"Factory must accept exactly one parameter (Container) and return exactly two values (interface, error)",
		}
		return diErr.WithSuggestions(suggestions)
	}

	// Check that second return type is error
	if factoryType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		diErr := NewDIError(
			interfaceType,
			"factory registration",
			"factory second return value must be error",
		).WithContext("factory_signature", factoryType.String()).
			WithContext("second_return", factoryType.Out(1).String())

		suggestions := []string{
			"Factory function must return (YourInterface, error)",
			"Second return value must be the built-in error type",
		}
		return diErr.WithSuggestions(suggestions)
	}

	typeKey := interfaceType.String()
	c.registrations[typeKey] = &registration{
		factory:       factory,
		scope:         scope,
		interfaceType: interfaceType,
		isFactory:     true,
	}

	// Update suggestion engine with new registration
	c.updateSuggestionEngine()

	return nil
}

// ResolveType resolves a dependency by interface type
func (c *containerImpl) ResolveType(interfaceType reflect.Type) (interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create a local resolution path for this resolution chain
	resolutionPath := make(map[string]bool)
	return c.resolveTypeInternal(interfaceType, resolutionPath, "")
}

// ResolveTypeWithContext resolves a dependency by interface type with context for scoped dependencies
func (c *containerImpl) ResolveTypeWithContext(ctx context.Context, interfaceType reflect.Type) (interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Generate context key for scoped dependencies
	contextKey := ""
	if ctx != nil {
		keyGen := &DefaultContextKeyGenerator{}
		contextKey = keyGen.GenerateKey(ctx)
	}

	// Create a local resolution path for this resolution chain
	resolutionPath := make(map[string]bool)
	return c.resolveTypeInternal(interfaceType, resolutionPath, contextKey)
}

// resolveTypeInternal is the internal resolution method (assumes lock is held)
func (c *containerImpl) resolveTypeInternal(interfaceType reflect.Type, resolutionPath map[string]bool, contextKey string) (interface{}, error) {
	typeKey := interfaceType.String()

	// Build current dependency chain for error reporting
	dependencyChain := make([]string, 0, len(resolutionPath)+1)
	for dep := range resolutionPath {
		dependencyChain = append(dependencyChain, dep)
	}
	dependencyChain = append(dependencyChain, typeKey)

	// Check for circular dependency
	if resolutionPath[typeKey] {
		diErr := NewDIError(
			interfaceType,
			"resolution",
			"circular dependency detected",
		).WithDependencyChain(dependencyChain)

		suggestions := []string{
			"Break the circular dependency by introducing an intermediate interface",
			"Use factory registration to defer one of the dependencies",
			"Consider redesigning the dependency relationship",
		}
		return nil, diErr.WithSuggestions(suggestions)
	}

	// Find registration
	reg, exists := c.registrations[typeKey]
	if !exists {
		// Get available registrations for error reporting
		availableRegistrations := make([]string, 0, len(c.registrations))
		for regKey := range c.registrations {
			availableRegistrations = append(availableRegistrations, regKey)
		}

		diErr := NewDIError(
			interfaceType,
			"resolution",
			"no registration found for interface type",
		).WithDependencyChain(dependencyChain).
			WithAvailableRegistrations(availableRegistrations).
			WithContext("type", typeKey)

		// Generate intelligent suggestions
		suggestions := c.suggestionEngine.GenerateSuggestions(diErr)
		return nil, diErr.WithSuggestions(suggestions)
	}

	// Handle singleton scope - use container-level synchronization
	if reg.scope == Singleton {
		// Record metrics for singleton resolution
		startTime := time.Now()
		defer func() {
			c.sharedManager.recordResolutionMetrics(startTime, "singleton")
		}()

		// Check if we already have a singleton instance
		if instance := c.sharedManager.GetExistingSingleton(interfaceType); instance != nil {
			// Track memory optimization - this was a share, not a new creation
			c.memoryTracker.TrackDependencyShare(interfaceType.String())
			c.memoryTracker.TrackDependencyAccess(interfaceType.String())
			return instance, nil
		}

		// Create the instance
		instance, err := c.createInstanceUnsafe(reg, contextKey, resolutionPath, dependencyChain)
		if err != nil {
			return nil, err
		}

		// Store the singleton instance
		c.sharedManager.SetSingleton(interfaceType, instance)

		// Track memory usage for the new singleton
		c.memoryTracker.TrackDependencyCreation(interfaceType.String(), instance, Singleton)

		// Register for health monitoring if it implements HealthChecker or is a common framework dependency
		c.registerForHealthMonitoring(interfaceType.String(), instance)

		// Register for lifecycle management if it implements lifecycle interfaces
		c.registerForLifecycleManagement(interfaceType.String(), instance)

		return instance, nil
	}

	// Handle scoped dependencies - use container-level synchronization
	if reg.scope == Scoped && contextKey != "" {
		// Record metrics for scoped resolution
		startTime := time.Now()
		defer func() {
			c.sharedManager.recordResolutionMetrics(startTime, "scoped")
		}()

		// Check if we already have a scoped instance
		if instance := c.sharedManager.GetExistingScoped(contextKey, interfaceType); instance != nil {
			// Track memory optimization - this was a share, not a new creation
			scopedDependencyType := fmt.Sprintf("%s[%s]", interfaceType.String(), contextKey)
			c.memoryTracker.TrackDependencyShare(scopedDependencyType)
			c.memoryTracker.TrackDependencyAccess(scopedDependencyType)
			return instance, nil
		}

		// Create the instance
		instance, err := c.createInstanceUnsafe(reg, contextKey, resolutionPath, dependencyChain)
		if err != nil {
			return nil, err
		}

		// Store the scoped instance
		c.sharedManager.SetScoped(contextKey, interfaceType, instance)

		// Track memory usage for the new scoped instance
		scopedDependencyType := fmt.Sprintf("%s[%s]", interfaceType.String(), contextKey)
		c.memoryTracker.TrackDependencyCreation(scopedDependencyType, instance, Scoped)

		// Register for health monitoring if it implements HealthChecker or is a common framework dependency
		c.registerForHealthMonitoring(scopedDependencyType, instance)

		// Register for lifecycle management if it implements lifecycle interfaces
		c.registerForLifecycleManagement(scopedDependencyType, instance)

		return instance, nil
	} // Handle transient dependencies or fallback
	c.sharedManager.recordResolutionMetrics(time.Now(), "transient")
	instance, err := c.createInstance(reg, contextKey, resolutionPath, dependencyChain)
	if err == nil && instance != nil {
		// Track memory usage for transient instance (no sharing optimization)
		c.memoryTracker.TrackDependencyCreation(interfaceType.String(), instance, Transient)
	}
	return instance, err
}

// createInstanceUnsafe creates a new instance of the dependency (assumes locks are held)
func (c *containerImpl) createInstanceUnsafe(reg *registration, contextKey string, resolutionPath map[string]bool, dependencyChain []string) (interface{}, error) {
	typeKey := reg.interfaceType.String()

	// Mark this type as being resolved (for circular dependency detection)
	resolutionPath[typeKey] = true
	defer delete(resolutionPath, typeKey)

	var instance interface{}
	var err error

	if reg.isFactory {
		// Create a wrapper that can handle the recursive resolution
		// Note: We keep the current lock held and use a special internal resolution method
		wrapper := &factoryWrapper{
			container:      c,
			resolutionPath: resolutionPath,
			contextKey:     contextKey,
		}

		// Call factory function
		factoryValue := reflect.ValueOf(reg.factory)
		containerValue := reflect.ValueOf(wrapper)
		results := factoryValue.Call([]reflect.Value{containerValue})

		// Check for error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
			diErr := NewDIError(
				reg.interfaceType,
				"factory resolution",
				"factory function returned error",
			).WithDependencyChain(dependencyChain).
				WithContext("factory_error", err.Error())

			suggestions := []string{
				"Check the factory function implementation for errors",
				"Ensure all dependencies required by the factory are properly registered",
				"Review factory function parameters and return types",
			}
			return nil, diErr.WithSuggestions(suggestions)
		}

		instance = results[0].Interface()
	} else {
		// Use registered instance directly
		instance = reg.instance
	}

	return instance, nil
}

// createInstance creates a new instance of the dependency
func (c *containerImpl) createInstance(reg *registration, contextKey string, resolutionPath map[string]bool, dependencyChain []string) (interface{}, error) {
	typeKey := reg.interfaceType.String()

	// Mark this type as being resolved (for circular dependency detection)
	resolutionPath[typeKey] = true
	defer delete(resolutionPath, typeKey)

	var instance interface{}
	var err error

	if reg.isFactory {
		// Create a wrapper that can handle the recursive resolution
		// Note: We keep the current lock held and use a special internal resolution method
		wrapper := &factoryWrapper{
			container:      c,
			resolutionPath: resolutionPath,
			contextKey:     contextKey,
		}

		// Call factory function
		factoryValue := reflect.ValueOf(reg.factory)
		containerValue := reflect.ValueOf(wrapper)
		results := factoryValue.Call([]reflect.Value{containerValue})

		// Check for error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
			diErr := NewDIError(
				reg.interfaceType,
				"factory resolution",
				"factory function returned error",
			).WithDependencyChain(dependencyChain).
				WithContext("factory_error", err.Error())

			suggestions := []string{
				"Check the factory function implementation for errors",
				"Ensure all dependencies required by the factory are properly registered",
				"Review factory function parameters and return types",
			}
			return nil, diErr.WithSuggestions(suggestions)
		}

		instance = results[0].Interface()
	} else {
		// Use registered instance directly
		instance = reg.instance
	}

	return instance, nil
}

// factoryWrapper wraps the container to provide the resolution path context for factory functions
type factoryWrapper struct {
	container      *containerImpl
	resolutionPath map[string]bool
	contextKey     string
}

// Implementation of Container interface for factoryWrapper
func (fw *factoryWrapper) RegisterType(interfaceType reflect.Type, impl interface{}, scope Scope) error {
	return fw.container.RegisterType(interfaceType, impl, scope)
}

func (fw *factoryWrapper) RegisterFactoryType(interfaceType reflect.Type, factory interface{}, scope Scope) error {
	return fw.container.RegisterFactoryType(interfaceType, factory, scope)
}

func (fw *factoryWrapper) ResolveType(interfaceType reflect.Type) (interface{}, error) {
	// Factory wrapper uses internal resolution that assumes the lock is already held
	return fw.container.resolveTypeInternal(interfaceType, fw.resolutionPath, fw.contextKey)
}

func (fw *factoryWrapper) ResolveTypeWithContext(ctx context.Context, interfaceType reflect.Type) (interface{}, error) {
	// For factory resolution, use the provided context key from the wrapper
	return fw.container.resolveTypeInternal(interfaceType, fw.resolutionPath, fw.contextKey)
}

func (fw *factoryWrapper) Validate() []error {
	return fw.container.Validate()
}

func (fw *factoryWrapper) GetDependencyGraph() map[string][]string {
	return fw.container.GetDependencyGraph()
}

func (fw *factoryWrapper) Reset() {
	fw.container.Reset()
}

func (fw *factoryWrapper) GetHealthMonitor() *HealthMonitor {
	return fw.container.GetHealthMonitor()
}

func (fw *factoryWrapper) GetMemoryTracker() *MemoryTracker {
	return fw.container.GetMemoryTracker()
}

func (fw *factoryWrapper) GetMemoryProfiler() *MemoryProfiler {
	return fw.container.GetMemoryProfiler()
}

func (fw *factoryWrapper) GetPoolManager() *DependencyPoolManager {
	return fw.container.GetPoolManager()
}

func (fw *factoryWrapper) GetLifecycleManager() *LifecycleManager {
	return fw.container.GetLifecycleManager()
}

func (fw *factoryWrapper) GetUpdateManager() *DependencyUpdateManager {
	return fw.container.GetUpdateManager()
}

// Validate checks the dependency graph for issues
func (c *containerImpl) Validate() []error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var errors []error

	// Check each registration for validity
	for _, reg := range c.registrations {
		// Verify interface compliance for direct registrations
		if !reg.isFactory && reg.instance != nil {
			implType := reflect.TypeOf(reg.instance)
			if !implType.Implements(reg.interfaceType) {
				diErr := NewDIError(
					reg.interfaceType,
					"validation",
					"registered instance does not implement interface",
				).WithContext("interface_type", reg.interfaceType.String()).
					WithContext("implementation_type", implType.String())

				suggestions := []string{
					fmt.Sprintf("Ensure %s implements all methods of %s", implType.String(), reg.interfaceType.String()),
					"Check method signatures match exactly (including receiver types)",
					"Verify generic type parameters are correctly specified",
				}
				errors = append(errors, diErr.WithSuggestions(suggestions))
			}
		}

		// For factory registrations, validate the factory signature
		if reg.isFactory && reg.factory != nil {
			factoryType := reflect.TypeOf(reg.factory)
			if factoryType.Kind() != reflect.Func {
				diErr := NewDIError(
					reg.interfaceType,
					"validation",
					"factory is not a function",
				).WithContext("factory_type", factoryType.String())

				suggestions := []string{
					"Ensure the factory is a function, not another type",
					"Factory should have signature: func(Container) (YourInterface, error)",
				}
				errors = append(errors, diErr.WithSuggestions(suggestions))
			}
		}
	}

	// TODO: Add more sophisticated dependency graph analysis
	// - Check for missing dependencies in factory functions
	// - Validate complete dependency chains

	return errors
}

// GetDependencyGraph returns a representation of the dependency graph
func (c *containerImpl) GetDependencyGraph() map[string][]string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	graph := make(map[string][]string)

	for typeKey, reg := range c.registrations {
		var dependencies []string

		if reg.isFactory {
			// For factory registrations, we could potentially analyze the factory
			// function to determine its dependencies, but that's complex.
			// For now, we'll just indicate it's a factory.
			dependencies = []string{"<factory>"}
		} else {
			// For direct registrations, the instance has no container dependencies
			dependencies = []string{"<direct>"}
		}

		graph[typeKey] = dependencies
	}

	return graph
}

// Reset clears all registrations and singletons
func (c *containerImpl) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.registrations = make(map[string]*registration)
	c.singletons = make(map[string]interface{})
	// Stop health monitoring and reset the health monitor
	c.healthMonitor.StopMonitoring()
	c.healthMonitor = NewHealthMonitor()
	// Reset circuit breakers
	c.circuitBreakerManager.Reset()
	// Reset suggestion engine
	c.suggestionEngine = NewSuggestionEngine()
}

// GetHealthMonitor returns the health monitor instance
func (c *containerImpl) GetHealthMonitor() *HealthMonitor {
	return c.healthMonitor
}

// GetMemoryTracker returns the memory tracker instance
func (c *containerImpl) GetMemoryTracker() *MemoryTracker {
	return c.memoryTracker
}

// GetMemoryProfiler returns the memory profiler instance
func (c *containerImpl) GetMemoryProfiler() *MemoryProfiler {
	return c.memoryProfiler
}

// GetPoolManager returns the dependency pool manager
func (c *containerImpl) GetPoolManager() *DependencyPoolManager {
	return c.poolManager
}

// GetLifecycleManager returns the lifecycle manager
func (c *containerImpl) GetLifecycleManager() *LifecycleManager {
	return c.lifecycleManager
}

// GetUpdateManager returns the dependency update manager
func (c *containerImpl) GetUpdateManager() *DependencyUpdateManager {
	return c.updateManager
}

// registerForHealthMonitoring registers a dependency instance for health monitoring
func (c *containerImpl) registerForHealthMonitoring(dependencyType string, instance interface{}) {
	// Determine if this dependency should be monitored
	shouldMonitor := false

	// Check if it implements HealthChecker interface
	if _, ok := instance.(HealthChecker); ok {
		shouldMonitor = true
	}

	// Check if it's a common framework dependency that should be monitored
	if !shouldMonitor {
		shouldMonitor = IsCommonFrameworkDependency(reflect.TypeOf(instance))
	}

	if shouldMonitor {
		// Use shorter intervals for critical dependencies
		interval := 30 * time.Second
		if IsCommonFrameworkDependency(reflect.TypeOf(instance)) {
			interval = 15 * time.Second // More frequent checks for framework dependencies
		}

		c.healthMonitor.RegisterDependency(
			dependencyType,
			instance,
			WithCheckInterval(interval),
			WithMaxFailures(3),
		)
	}
}

// registerForLifecycleManagement registers a dependency instance for lifecycle management
func (c *containerImpl) registerForLifecycleManagement(dependencyType string, instance interface{}) {
	// Check if it implements lifecycle interfaces
	var priority LifecyclePriority = PriorityNormal

	// Determine priority based on dependency type and interfaces
	if IsCommonFrameworkDependency(reflect.TypeOf(instance)) {
		priority = PriorityHigh // Framework dependencies should start early
	}

	// Check if it implements lifecycle interfaces
	if _, implementsStartable := instance.(Startable); implementsStartable {
		c.lifecycleManager.RegisterDependency(dependencyType, instance, WithPriority(priority))
	} else if _, implementsStoppable := instance.(Stoppable); implementsStoppable {
		c.lifecycleManager.RegisterDependency(dependencyType, instance, WithPriority(priority))
	} else if _, implementsLifecycleAware := instance.(LifecycleAware); implementsLifecycleAware {
		c.lifecycleManager.RegisterDependency(dependencyType, instance, WithPriority(priority))
	}
}
