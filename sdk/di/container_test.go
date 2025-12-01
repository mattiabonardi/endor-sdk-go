package di

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test interfaces for testing
type TestInterface interface {
	DoSomething() string
}

type AnotherInterface interface {
	DoSomethingElse() int
}

// Test implementations
type testImpl struct {
	value string
}

func (t *testImpl) DoSomething() string {
	return t.value
}

type anotherImpl struct {
	number int
}

func (a *anotherImpl) DoSomethingElse() int {
	return a.number
}

// Test structs that don't implement interfaces
type nonInterfaceImpl struct{}

func TestNewContainer(t *testing.T) {
	container := NewContainer()
	assert.NotNil(t, container)
}

func TestRegister_Success(t *testing.T) {
	container := NewContainer()
	impl := &testImpl{value: "test"}

	err := Register[TestInterface](container, impl, Singleton)
	assert.NoError(t, err)

	// Verify we can resolve it
	resolved, err := Resolve[TestInterface](container)
	require.NoError(t, err)
	assert.Equal(t, "test", resolved.DoSomething())
}

func TestRegister_TransientScope(t *testing.T) {
	container := NewContainer()
	impl := &testImpl{value: "transient"}

	err := Register[TestInterface](container, impl, Transient)
	assert.NoError(t, err)

	// Resolve multiple times - should get the same instance for direct registration
	// (transient vs singleton mainly affects factory functions)
	resolved1, err1 := Resolve[TestInterface](container)
	resolved2, err2 := Resolve[TestInterface](container)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, "transient", resolved1.DoSomething())
	assert.Equal(t, "transient", resolved2.DoSomething())
}

func TestRegister_NonInterface(t *testing.T) {
	container := NewContainer()
	impl := &nonInterfaceImpl{}

	// This should fail because nonInterfaceImpl is not an interface type
	err := Register[*nonInterfaceImpl](container, impl, Singleton)
	assert.Error(t, err)

	var diErr *DIError
	assert.True(t, errors.As(err, &diErr))
	assert.Contains(t, diErr.Error(), "type is not an interface")
}

func TestResolve_NotRegistered(t *testing.T) {
	container := NewContainer()

	_, err := Resolve[TestInterface](container)
	assert.Error(t, err)

	var diErr *DIError
	assert.True(t, errors.As(err, &diErr))
	assert.Contains(t, diErr.Error(), "no registration found")
}

func TestRegisterFactory_Success(t *testing.T) {
	container := NewContainer()

	factory := func(c Container) (TestInterface, error) {
		return &testImpl{value: "factory"}, nil
	}

	err := RegisterFactory[TestInterface](container, factory, Singleton)
	assert.NoError(t, err)

	resolved, err := Resolve[TestInterface](container)
	require.NoError(t, err)
	assert.Equal(t, "factory", resolved.DoSomething())
}

func TestRegisterFactory_TransientScope(t *testing.T) {
	container := NewContainer()

	counter := 0
	factory := func(c Container) (TestInterface, error) {
		counter++
		return &testImpl{value: "factory"}, nil
	}

	err := RegisterFactory[TestInterface](container, factory, Transient)
	assert.NoError(t, err)

	// Resolve multiple times - should call factory each time for transient
	_, err1 := Resolve[TestInterface](container)
	_, err2 := Resolve[TestInterface](container)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, 2, counter, "Factory should be called twice for transient scope")
}

func TestRegisterFactory_SingletonScope(t *testing.T) {
	container := NewContainer()

	counter := 0
	factory := func(c Container) (TestInterface, error) {
		counter++
		return &testImpl{value: "singleton"}, nil
	}

	err := RegisterFactory[TestInterface](container, factory, Singleton)
	assert.NoError(t, err)

	// Resolve multiple times - should call factory only once for singleton
	resolved1, err1 := Resolve[TestInterface](container)
	resolved2, err2 := Resolve[TestInterface](container)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, 1, counter, "Factory should be called only once for singleton scope")
	assert.Same(t, resolved1, resolved2, "Should return the same instance")
}

func TestRegisterFactory_FactoryError(t *testing.T) {
	container := NewContainer()

	factory := func(c Container) (TestInterface, error) {
		return nil, errors.New("factory failed")
	}

	err := RegisterFactory[TestInterface](container, factory, Singleton)
	assert.NoError(t, err)

	_, err = Resolve[TestInterface](container)
	assert.Error(t, err)

	var diErr *DIError
	assert.True(t, errors.As(err, &diErr))
	assert.Contains(t, diErr.Error(), "factory function returned error")
}

func TestValidate_Success(t *testing.T) {
	container := NewContainer()
	impl := &testImpl{value: "test"}

	err := Register[TestInterface](container, impl, Singleton)
	require.NoError(t, err)

	validationErrors := container.Validate()
	assert.Empty(t, validationErrors)
}

func TestValidate_InterfaceComplianceError(t *testing.T) {
	container := NewContainer()

	// Try to register something that doesn't implement the interface
	interfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem()
	wrongImpl := &nonInterfaceImpl{}

	// Registration should fail because nonInterfaceImpl doesn't implement TestInterface
	err := container.RegisterType(interfaceType, wrongImpl, Singleton)
	assert.Error(t, err) // Registration should fail with our validation

	var diErr *DIError
	assert.True(t, errors.As(err, &diErr))
	assert.Contains(t, diErr.Error(), "does not implement the interface")

	// Since registration failed, validation should succeed (no invalid registrations)
	validationErrors := container.Validate()
	assert.Empty(t, validationErrors)
}

func TestGetDependencyGraph(t *testing.T) {
	container := NewContainer()
	impl := &testImpl{value: "test"}

	err := Register[TestInterface](container, impl, Singleton)
	require.NoError(t, err)

	factory := func(c Container) (AnotherInterface, error) {
		return &anotherImpl{number: 42}, nil
	}
	err = RegisterFactory[AnotherInterface](container, factory, Transient)
	require.NoError(t, err)

	graph := container.GetDependencyGraph()

	testInterfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem().String()
	anotherInterfaceType := reflect.TypeOf((*AnotherInterface)(nil)).Elem().String()

	assert.Contains(t, graph, testInterfaceType)
	assert.Contains(t, graph, anotherInterfaceType)
	assert.Equal(t, []string{"<direct>"}, graph[testInterfaceType])
	assert.Equal(t, []string{"<factory>"}, graph[anotherInterfaceType])
}

func TestReset(t *testing.T) {
	container := NewContainer()
	impl := &testImpl{value: "test"}

	err := Register[TestInterface](container, impl, Singleton)
	require.NoError(t, err)

	// Verify it's registered
	_, err = Resolve[TestInterface](container)
	assert.NoError(t, err)

	// Reset
	container.Reset()

	// Should no longer be registered
	_, err = Resolve[TestInterface](container)
	assert.Error(t, err)
}

func TestCircularDependencyDetection(t *testing.T) {
	// This test is more complex and requires interfaces that depend on each other
	// For now, we'll test the circular dependency error structure

	interfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem()
	path := []string{"TypeA", "TypeB", "TypeA"}

	err := NewCircularDependencyError(interfaceType, path)
	assert.Contains(t, err.Error(), "circular dependency detected")
	assert.Contains(t, err.Error(), "TestInterface")
	assert.Contains(t, err.Error(), "TypeA")
}

func TestScopeString(t *testing.T) {
	assert.Equal(t, "Singleton", Singleton.String())
	assert.Equal(t, "Transient", Transient.String())
	assert.Equal(t, "Scoped", Scoped.String())

	unknownScope := Scope(999)
	assert.Equal(t, "Unknown", unknownScope.String())
}

func TestConcurrentAccess(t *testing.T) {
	container := NewContainer()
	impl := &testImpl{value: "concurrent"}

	err := Register[TestInterface](container, impl, Singleton)
	require.NoError(t, err)

	// Test concurrent resolution
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			resolved, err := Resolve[TestInterface](container)
			assert.NoError(t, err)
			assert.Equal(t, "concurrent", resolved.DoSomething())
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
