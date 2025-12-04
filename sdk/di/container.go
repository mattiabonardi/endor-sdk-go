package di

import (
	"context"
	"reflect"
)

// Container defines the dependency injection container interface with
// registration and resolution capabilities.
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
}

// Register provides type-safe registration of an interface implementation.
// T must be an interface type, and impl must implement that interface.
func Register[T any](container Container, impl T, scope Scope) error {
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	return container.RegisterType(interfaceType, impl, scope)
}

// RegisterFactory provides type-safe registration of a factory function.
// T must be an interface type, and the factory must return an implementation of T.
func RegisterFactory[T any](container Container, factory func(Container) (T, error), scope Scope) error {
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	return container.RegisterFactoryType(interfaceType, factory, scope)
}

// Resolve provides type-safe resolution of a dependency by interface type.
// T must be an interface type that has been previously registered.
func Resolve[T any](container Container) (T, error) {
	var zero T
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	result, err := container.ResolveType(interfaceType)
	if err != nil {
		return zero, err
	}

	typed, ok := result.(T)
	if !ok {
		return zero, NewDependencyError(
			interfaceType,
			"resolution",
			"resolved instance does not implement the requested interface",
			map[string]interface{}{
				"expected": interfaceType.String(),
				"actual":   reflect.TypeOf(result).String(),
			},
		)
	}

	return typed, nil
}

// ResolveWithContext provides type-safe resolution of a dependency with context for scoped dependencies.
// T must be an interface type that has been previously registered.
func ResolveWithContext[T any](ctx context.Context, container Container) (T, error) {
	var zero T
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	result, err := container.ResolveTypeWithContext(ctx, interfaceType)
	if err != nil {
		return zero, err
	}

	typed, ok := result.(T)
	if !ok {
		return zero, NewDependencyError(
			interfaceType,
			"scoped resolution",
			"resolved instance does not implement the requested interface",
			map[string]interface{}{
				"expected": interfaceType.String(),
				"actual":   reflect.TypeOf(result).String(),
			},
		)
	}

	return typed, nil
}

// NewContainer creates a new dependency injection container
func NewContainer() Container {
	return newContainerImpl()
}

// registration represents an internal dependency registration
type registration struct {
	// instance holds the registered implementation (for direct registration)
	instance interface{}
	// factory holds the factory function (for factory registration)
	factory interface{}
	// scope defines the lifecycle management strategy
	scope Scope
	// interfaceType holds the interface type this registration implements
	interfaceType reflect.Type
	// isFactory indicates if this is a factory registration
	isFactory bool
	// singletonInstance caches the singleton instance if scope is Singleton
	singletonInstance interface{}
}
