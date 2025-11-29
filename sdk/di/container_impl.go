package di

import (
	"reflect"
	"sync"
)

// containerImpl implements the Container interface
type containerImpl struct {
	// registrations stores all dependency registrations by interface type string
	registrations map[string]*registration
	// singletons stores singleton instances by interface type string
	singletons map[string]interface{}
	// mutex protects concurrent access to container state
	mutex sync.RWMutex
}

// newContainerImpl creates a new container implementation
func newContainerImpl() Container {
	return &containerImpl{
		registrations: make(map[string]*registration),
		singletons:    make(map[string]interface{}),
	}
}

// RegisterType registers an implementation for an interface type
func (c *containerImpl) RegisterType(interfaceType reflect.Type, impl interface{}, scope Scope) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Validate that interfaceType is actually an interface
	if interfaceType.Kind() != reflect.Interface {
		return NewDependencyError(
			interfaceType,
			"registration",
			"type is not an interface",
			map[string]interface{}{
				"type_kind": interfaceType.Kind().String(),
				"type_name": interfaceType.String(),
			},
		)
	}

	// Validate that impl implements the interface
	implType := reflect.TypeOf(impl)
	if !implType.Implements(interfaceType) {
		return NewDependencyError(
			interfaceType,
			"registration",
			"implementation does not implement the interface",
			map[string]interface{}{
				"interface_type":      interfaceType.String(),
				"implementation_type": implType.String(),
			},
		)
	}

	typeKey := interfaceType.String()
	c.registrations[typeKey] = &registration{
		instance:      impl,
		scope:         scope,
		interfaceType: interfaceType,
		isFactory:     false,
	}

	return nil
}

// RegisterFactoryType registers a factory function for an interface type
func (c *containerImpl) RegisterFactoryType(interfaceType reflect.Type, factory interface{}, scope Scope) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Validate that interfaceType is actually an interface
	if interfaceType.Kind() != reflect.Interface {
		return NewDependencyError(
			interfaceType,
			"factory registration",
			"type is not an interface",
			map[string]interface{}{
				"type_kind": interfaceType.Kind().String(),
				"type_name": interfaceType.String(),
			},
		)
	}

	// Validate factory function signature
	factoryType := reflect.TypeOf(factory)
	if factoryType.Kind() != reflect.Func {
		return NewDependencyError(
			interfaceType,
			"factory registration",
			"factory must be a function",
			map[string]interface{}{
				"factory_type": factoryType.String(),
			},
		)
	}

	// Factory should accept Container and return (T, error)
	if factoryType.NumIn() != 1 || factoryType.NumOut() != 2 {
		return NewDependencyError(
			interfaceType,
			"factory registration",
			"factory must have signature func(Container) (T, error)",
			map[string]interface{}{
				"factory_signature": factoryType.String(),
				"expected_in":       1,
				"actual_in":         factoryType.NumIn(),
				"expected_out":      2,
				"actual_out":        factoryType.NumOut(),
			},
		)
	}

	// Check that second return type is error
	if factoryType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return NewDependencyError(
			interfaceType,
			"factory registration",
			"factory second return value must be error",
			map[string]interface{}{
				"factory_signature": factoryType.String(),
				"second_return":     factoryType.Out(1).String(),
			},
		)
	}

	typeKey := interfaceType.String()
	c.registrations[typeKey] = &registration{
		factory:       factory,
		scope:         scope,
		interfaceType: interfaceType,
		isFactory:     true,
	}

	return nil
}

// ResolveType resolves a dependency by interface type
func (c *containerImpl) ResolveType(interfaceType reflect.Type) (interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create a local resolution path for this resolution chain
	resolutionPath := make(map[string]bool)
	return c.resolveTypeInternal(interfaceType, resolutionPath)
}

// resolveTypeInternal is the internal resolution method (assumes lock is held)
func (c *containerImpl) resolveTypeInternal(interfaceType reflect.Type, resolutionPath map[string]bool) (interface{}, error) {
	typeKey := interfaceType.String()

	// Check for circular dependency
	if resolutionPath[typeKey] {
		// Build path for error reporting
		path := make([]string, 0, len(resolutionPath))
		for dep := range resolutionPath {
			path = append(path, dep)
		}
		path = append(path, typeKey)
		return nil, NewCircularDependencyError(interfaceType, path)
	}

	// Find registration
	reg, exists := c.registrations[typeKey]
	if !exists {
		return nil, NewDependencyError(
			interfaceType,
			"resolution",
			"no registration found for interface type",
			map[string]interface{}{
				"type": typeKey,
			},
		)
	}

	// Handle singleton scope - check if we already have an instance
	if reg.scope == Singleton {
		if singleton, exists := c.singletons[typeKey]; exists {
			return singleton, nil
		}
	}

	// Mark this type as being resolved (for circular dependency detection)
	resolutionPath[typeKey] = true
	defer delete(resolutionPath, typeKey)

	var instance interface{}
	var err error

	if reg.isFactory {
		// For factory calls, we need to release the lock temporarily to avoid deadlock
		// when the factory calls back into the container to resolve its own dependencies
		c.mutex.Unlock()

		// Create a wrapper that can handle the recursive resolution
		wrapper := &factoryWrapper{container: c, resolutionPath: resolutionPath}

		// Call factory function
		factoryValue := reflect.ValueOf(reg.factory)
		containerValue := reflect.ValueOf(wrapper)
		results := factoryValue.Call([]reflect.Value{containerValue})

		// Reacquire the lock
		c.mutex.Lock()

		// Check for error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
			return nil, NewDependencyError(
				interfaceType,
				"factory resolution",
				"factory function returned error",
				map[string]interface{}{
					"factory_error": err.Error(),
				},
			)
		}

		instance = results[0].Interface()
	} else {
		// Use registered instance directly
		instance = reg.instance
	}

	// Store singleton instance
	if reg.scope == Singleton {
		c.singletons[typeKey] = instance
	}

	return instance, nil
}

// factoryWrapper wraps the container to provide the resolution path context for factory functions
type factoryWrapper struct {
	container      *containerImpl
	resolutionPath map[string]bool
}

// Implementation of Container interface for factoryWrapper
func (fw *factoryWrapper) RegisterType(interfaceType reflect.Type, impl interface{}, scope Scope) error {
	return fw.container.RegisterType(interfaceType, impl, scope)
}

func (fw *factoryWrapper) RegisterFactoryType(interfaceType reflect.Type, factory interface{}, scope Scope) error {
	return fw.container.RegisterFactoryType(interfaceType, factory, scope)
}

func (fw *factoryWrapper) ResolveType(interfaceType reflect.Type) (interface{}, error) {
	fw.container.mutex.Lock()
	defer fw.container.mutex.Unlock()

	return fw.container.resolveTypeInternal(interfaceType, fw.resolutionPath)
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
				errors = append(errors, NewDependencyError(
					reg.interfaceType,
					"validation",
					"registered instance does not implement interface",
					map[string]interface{}{
						"interface_type":      reg.interfaceType.String(),
						"implementation_type": implType.String(),
					},
				))
			}
		}

		// For factory registrations, validate the factory signature
		if reg.isFactory && reg.factory != nil {
			factoryType := reflect.TypeOf(reg.factory)
			if factoryType.Kind() != reflect.Func {
				errors = append(errors, NewDependencyError(
					reg.interfaceType,
					"validation",
					"factory is not a function",
					map[string]interface{}{
						"factory_type": factoryType.String(),
					},
				))
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
}
