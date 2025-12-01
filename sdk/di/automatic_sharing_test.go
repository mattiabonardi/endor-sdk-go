package di

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test automatic dependency sharing detection
func TestAutomaticDependencySharing(t *testing.T) {
	t.Run("IsCommonFrameworkDependency_DetectsCommonTypes", func(t *testing.T) {
		// Create mock interface types representing common dependencies
		configType := reflect.TypeOf((*ConfigProviderInterface)(nil)).Elem()
		loggerType := reflect.TypeOf((*LoggerInterface)(nil)).Elem()
		repoType := reflect.TypeOf((*RepositoryInterface)(nil)).Elem()

		// Print the actual type names for debugging
		t.Logf("Config type: %s", configType.String())
		t.Logf("Logger type: %s", loggerType.String())
		t.Logf("Repo type: %s", repoType.String())

		// These should be detected as common framework dependencies
		assert.True(t, IsCommonFrameworkDependency(configType))
		assert.True(t, IsCommonFrameworkDependency(loggerType))
		assert.True(t, IsCommonFrameworkDependency(repoType))

		// Custom interfaces should not be detected
		customType := reflect.TypeOf((*ScopedTestInterface)(nil)).Elem()
		assert.False(t, IsCommonFrameworkDependency(customType))
	})

	t.Run("RecommendScope_SuggestsCorrectScopes", func(t *testing.T) {
		// Configuration providers should be singleton
		configType := reflect.TypeOf((*ConfigProviderInterface)(nil)).Elem()
		assert.Equal(t, Singleton, RecommendScope(configType))

		// Loggers should be singleton
		loggerType := reflect.TypeOf((*LoggerInterface)(nil)).Elem()
		assert.Equal(t, Singleton, RecommendScope(loggerType))

		// Repository interfaces should default to singleton
		repoType := reflect.TypeOf((*RepositoryInterface)(nil)).Elem()
		assert.Equal(t, Singleton, RecommendScope(repoType))

		// Unknown types should default to transient
		customType := reflect.TypeOf((*ScopedTestInterface)(nil)).Elem()
		assert.Equal(t, Transient, RecommendScope(customType))
	})
}

// Test dependency sharing metrics and monitoring
func TestDependencyMetrics(t *testing.T) {
	t.Run("MetricsTracking_SingletonResolutions", func(t *testing.T) {
		container := NewContainer()

		// Register a singleton
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: "singleton-metrics"}, nil
		}, Singleton)
		require.NoError(t, err)

		// Get initial metrics
		containerImpl := container.(*containerImpl)
		initialMetrics := containerImpl.sharedManager.GetMetrics()

		// Resolve multiple times
		for i := 0; i < 5; i++ {
			_, err := Resolve[ScopedTestInterface](container)
			require.NoError(t, err)
		}

		// Check metrics
		finalMetrics := containerImpl.sharedManager.GetMetrics()
		assert.Greater(t, finalMetrics.SingletonResolutions, initialMetrics.SingletonResolutions)
		assert.Greater(t, finalMetrics.SingletonCacheHits, int64(0)) // Should have cache hits after first resolution
		assert.Greater(t, finalMetrics.SharedInstanceCount, initialMetrics.SharedInstanceCount)
		assert.Greater(t, finalMetrics.ResolutionCount, initialMetrics.ResolutionCount)
	})

	t.Run("MetricsTracking_ScopedResolutions", func(t *testing.T) {
		container := NewContainer()

		// Register a scoped dependency
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: time.Now().Format("20060102150405.000000")}, nil
		}, Scoped)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), "request-id", "metrics-test")
		containerImpl := container.(*containerImpl)
		initialMetrics := containerImpl.sharedManager.GetMetrics()

		// Resolve multiple times in same context
		for i := 0; i < 3; i++ {
			_, err := ResolveWithContext[ScopedTestInterface](ctx, container)
			require.NoError(t, err)
		}

		// Check metrics
		finalMetrics := containerImpl.sharedManager.GetMetrics()
		assert.Greater(t, finalMetrics.ScopedResolutions, initialMetrics.ScopedResolutions)
		assert.Greater(t, finalMetrics.ScopedCacheHits, int64(0)) // Should have cache hits after first resolution
	})

	t.Run("PerformanceMetrics_AverageResolutionTime", func(t *testing.T) {
		container := NewContainer()

		// Register a dependency with a slight delay to measure timing
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			time.Sleep(1 * time.Microsecond) // Minimal delay for timing measurement
			return &scopedTestImpl{id: "performance-test"}, nil
		}, Singleton)
		require.NoError(t, err)

		// Resolve multiple times
		for i := 0; i < 10; i++ {
			_, err := Resolve[ScopedTestInterface](container)
			require.NoError(t, err)
		}

		// Check performance metrics
		containerImpl := container.(*containerImpl)
		metrics := containerImpl.sharedManager.GetMetrics()

		assert.Greater(t, metrics.ResolutionCount, int64(0))
		assert.Greater(t, metrics.TotalResolutionTime, time.Duration(0))
		assert.Greater(t, metrics.AverageResolutionTime, time.Duration(0))

		// Performance should still be well within targets
		assert.Less(t, metrics.AverageResolutionTime, 10*time.Microsecond) // Should be much faster than 10μs
	})
}

// Test memory optimization through efficient sharing
func TestMemoryOptimization(t *testing.T) {
	t.Run("SharedInstances_ReduceMemoryFootprint", func(t *testing.T) {
		container := NewContainer()

		// Register singleton that would normally create multiple instances
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: "shared-memory-test"}, nil
		}, Singleton)
		require.NoError(t, err)

		containerImpl := container.(*containerImpl)
		initialCount := containerImpl.sharedManager.GetMetrics().SharedInstanceCount

		// Resolve multiple times - should create only one instance
		instances := make([]ScopedTestInterface, 10)
		for i := 0; i < 10; i++ {
			instance, err := Resolve[ScopedTestInterface](container)
			require.NoError(t, err)
			instances[i] = instance
		}

		// Verify all instances are the same (memory sharing)
		for i := 1; i < len(instances); i++ {
			assert.Equal(t, instances[0].GetID(), instances[i].GetID())
		}

		// Should have created only one additional shared instance
		finalCount := containerImpl.sharedManager.GetMetrics().SharedInstanceCount
		assert.Equal(t, initialCount+1, finalCount)
	})

	t.Run("ScopedInstances_ShareWithinContext", func(t *testing.T) {
		container := NewContainer()

		// Register scoped dependency
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: time.Now().Format("20060102150405.000000")}, nil
		}, Scoped)
		require.NoError(t, err)

		ctx1 := context.WithValue(context.Background(), "request-id", "scope-test-1")
		ctx2 := context.WithValue(context.Background(), "request-id", "scope-test-2")

		// Resolve multiple times within each context
		var ctx1Instance1, ctx1Instance2 ScopedTestInterface
		var ctx2Instance1, ctx2Instance2 ScopedTestInterface

		ctx1Instance1, _ = ResolveWithContext[ScopedTestInterface](ctx1, container)
		ctx1Instance2, _ = ResolveWithContext[ScopedTestInterface](ctx1, container)
		ctx2Instance1, _ = ResolveWithContext[ScopedTestInterface](ctx2, container)
		ctx2Instance2, _ = ResolveWithContext[ScopedTestInterface](ctx2, container)

		// Within same context should be shared
		assert.Equal(t, ctx1Instance1.GetID(), ctx1Instance2.GetID())
		assert.Equal(t, ctx2Instance1.GetID(), ctx2Instance2.GetID())

		// Between different contexts should be different
		assert.NotEqual(t, ctx1Instance1.GetID(), ctx2Instance1.GetID())
	})
}

// Mock interfaces for testing common framework dependency detection
type ConfigProviderInterface interface {
	GetServerPort() string
	GetDocumentDBUri() string
	IsHybridResourcesEnabled() bool
}

type LoggerInterface interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
}

type RepositoryInterface interface {
	Create(ctx context.Context, resource interface{}) error
	Read(ctx context.Context, id string, result interface{}) error
	Update(ctx context.Context, resource interface{}) error
	Delete(ctx context.Context, id string) error
}
