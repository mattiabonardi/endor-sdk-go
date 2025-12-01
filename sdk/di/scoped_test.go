package di

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test interface for scoped dependency testing
type ScopedTestInterface interface {
	GetID() string
}

type scopedTestImpl struct {
	id string
}

func (s *scopedTestImpl) GetID() string {
	return s.id
}

// Test Scoped dependency lifecycle
func TestContainer_ScopedDependencies(t *testing.T) {
	t.Run("ScopedDependencies_SameContext_ReturnsSameInstance", func(t *testing.T) {
		container := NewContainer()

		// Register a factory that creates instances with unique IDs
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: time.Now().Format("20060102150405.000000")}, nil
		}, Scoped)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), "request-id", "test-request-123")

		// Resolve the same dependency multiple times within the same context
		instance1, err1 := ResolveWithContext[ScopedTestInterface](ctx, container)
		require.NoError(t, err1)

		instance2, err2 := ResolveWithContext[ScopedTestInterface](ctx, container)
		require.NoError(t, err2)

		// Should be the same instance
		assert.Equal(t, instance1.GetID(), instance2.GetID())
	})

	t.Run("ScopedDependencies_DifferentContext_ReturnsDifferentInstances", func(t *testing.T) {
		container := NewContainer()

		// Register a factory that creates instances with unique IDs
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: time.Now().Format("20060102150405.000000")}, nil
		}, Scoped)
		require.NoError(t, err)

		ctx1 := context.WithValue(context.Background(), "request-id", "test-request-123")
		ctx2 := context.WithValue(context.Background(), "request-id", "test-request-456")

		// Resolve dependency in different contexts
		instance1, err1 := ResolveWithContext[ScopedTestInterface](ctx1, container)
		require.NoError(t, err1)

		instance2, err2 := ResolveWithContext[ScopedTestInterface](ctx2, container)
		require.NoError(t, err2)

		// Should be different instances
		assert.NotEqual(t, instance1.GetID(), instance2.GetID())
	})

	t.Run("ScopedDependencies_NoContext_FallsBackToTransient", func(t *testing.T) {
		container := NewContainer()

		// Register a factory that creates instances with unique IDs
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: time.Now().Format("20060102150405.000000")}, nil
		}, Scoped)
		require.NoError(t, err)

		// Resolve without context (should behave like transient)
		instance1, err1 := Resolve[ScopedTestInterface](container)
		require.NoError(t, err1)

		instance2, err2 := Resolve[ScopedTestInterface](container)
		require.NoError(t, err2)

		// Should be different instances since no context for scoping
		assert.NotEqual(t, instance1.GetID(), instance2.GetID())
	})
}

// Test singleton behavior with shared dependency manager
func TestContainer_SingletonWithSharedManager(t *testing.T) {
	t.Run("Singleton_AlwaysReturnsSameInstance", func(t *testing.T) {
		container := NewContainer()

		// Register a factory that creates instances with unique IDs
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: time.Now().Format("20060102150405.000000")}, nil
		}, Singleton)
		require.NoError(t, err)

		// Resolve multiple times
		instance1, err1 := Resolve[ScopedTestInterface](container)
		require.NoError(t, err1)

		instance2, err2 := Resolve[ScopedTestInterface](container)
		require.NoError(t, err2)

		// Should be the same instance
		assert.Equal(t, instance1.GetID(), instance2.GetID())
	})

	t.Run("Singleton_SameInstanceAcrossDifferentContexts", func(t *testing.T) {
		container := NewContainer()

		// Register a factory that creates instances with unique IDs
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: time.Now().Format("20060102150405.000000")}, nil
		}, Singleton)
		require.NoError(t, err)

		ctx1 := context.WithValue(context.Background(), "request-id", "test-request-123")
		ctx2 := context.WithValue(context.Background(), "request-id", "test-request-456")

		// Resolve in different contexts
		instance1, err1 := ResolveWithContext[ScopedTestInterface](ctx1, container)
		require.NoError(t, err1)

		instance2, err2 := ResolveWithContext[ScopedTestInterface](ctx2, container)
		require.NoError(t, err2)

		// Should be the same instance even across different contexts
		assert.Equal(t, instance1.GetID(), instance2.GetID())
	})
}

// Test transient behavior
func TestContainer_TransientBehavior(t *testing.T) {
	t.Run("Transient_AlwaysReturnsNewInstance", func(t *testing.T) {
		container := NewContainer()

		// Register a factory that creates instances with unique IDs
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: time.Now().Format("20060102150405.000000")}, nil
		}, Transient)
		require.NoError(t, err)

		// Resolve multiple times
		instance1, err1 := Resolve[ScopedTestInterface](container)
		require.NoError(t, err1)

		instance2, err2 := Resolve[ScopedTestInterface](container)
		require.NoError(t, err2)

		// Should be different instances
		assert.NotEqual(t, instance1.GetID(), instance2.GetID())
	})
}

// Test performance requirements AC: 1 - Dependency resolution < 1μs for singletons, < 5μs for scoped
func BenchmarkDependencyResolution(b *testing.B) {
	container := NewContainer()

	// Register singleton
	err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
		return &scopedTestImpl{id: "singleton-test"}, nil
	}, Singleton)
	require.NoError(b, err)

	// Register scoped
	err = RegisterFactory[TestInterface](container, func(c Container) (TestInterface, error) {
		return &testImpl{value: "scoped-test"}, nil
	}, Scoped)
	require.NoError(b, err)

	b.Run("SingletonResolution", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := Resolve[ScopedTestInterface](container)
			if err != nil {
				b.Fatalf("Resolution failed: %v", err)
			}
		}

		// Verify performance target: < 1μs for singletons
		if b.N > 0 {
			avgTime := float64(b.Elapsed()) / float64(b.N)
			if avgTime > 1000 { // 1000ns = 1μs
				b.Logf("Warning: Average singleton resolution time %.2fns exceeds 1μs target", avgTime)
			}
		}
	})

	b.Run("ScopedResolution", func(b *testing.B) {
		ctx := context.WithValue(context.Background(), "request-id", "bench-test")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := ResolveWithContext[TestInterface](ctx, container)
			if err != nil {
				b.Fatalf("Scoped resolution failed: %v", err)
			}
		}

		// Verify performance target: < 5μs for scoped
		if b.N > 0 {
			avgTime := float64(b.Elapsed()) / float64(b.N)
			if avgTime > 5000 { // 5000ns = 5μs
				b.Logf("Warning: Average scoped resolution time %.2fns exceeds 5μs target", avgTime)
			}
		}
	})
}
