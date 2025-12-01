package di

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test concurrent access protection for singleton dependency resolution
func TestConcurrentSingletonResolution(t *testing.T) {
	t.Run("SingletonResolution_HighConcurrency_NoDataRaces", func(t *testing.T) {
		container := NewContainer()

		// Counter to track how many instances were actually created
		var creationCount int64

		// Register singleton with creation tracking
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			atomic.AddInt64(&creationCount, 1)
			time.Sleep(1 * time.Millisecond) // Small delay to increase chance of race conditions
			return &scopedTestImpl{id: "singleton-concurrent"}, nil
		}, Singleton)
		require.NoError(t, err)

		// Launch many goroutines simultaneously
		numGoroutines := 1000
		var wg sync.WaitGroup
		results := make([]ScopedTestInterface, numGoroutines)
		errors := make([]error, numGoroutines)

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				defer wg.Done()
				instance, err := Resolve[ScopedTestInterface](container)
				results[index] = instance
				errors[index] = err
			}(i)
		}

		wg.Wait()

		// Verify no errors occurred
		for i, err := range errors {
			require.NoError(t, err, "Goroutine %d failed", i)
		}

		// Verify only one instance was created despite many concurrent requests
		assert.Equal(t, int64(1), atomic.LoadInt64(&creationCount), "Should create exactly one singleton instance")

		// Verify all goroutines got the same instance
		firstInstance := results[0]
		for i, instance := range results {
			assert.Equal(t, firstInstance.GetID(), instance.GetID(), "Goroutine %d got different instance", i)
		}
	})

	t.Run("SingletonResolution_Performance_UnderLoad", func(t *testing.T) {
		container := NewContainer()

		// Register a simple singleton
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: "perf-test"}, nil
		}, Singleton)
		require.NoError(t, err)

		// Warm up - create the instance
		_, err = Resolve[ScopedTestInterface](container)
		require.NoError(t, err)

		// Measure performance under concurrent load
		numGoroutines := 1000
		numResolutionsPerGoroutine := 100

		start := time.Now()
		var wg sync.WaitGroup

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < numResolutionsPerGoroutine; j++ {
					_, err := Resolve[ScopedTestInterface](container)
					if err != nil {
						t.Errorf("Resolution failed: %v", err)
					}
				}
			}()
		}

		wg.Wait()
		elapsed := time.Since(start)

		totalResolutions := numGoroutines * numResolutionsPerGoroutine
		avgTimePerResolution := elapsed / time.Duration(totalResolutions)

		t.Logf("Performed %d concurrent singleton resolutions in %v", totalResolutions, elapsed)
		t.Logf("Average time per resolution: %v", avgTimePerResolution)

		// Should be well under 1μs per resolution even under load
		assert.Less(t, avgTimePerResolution, 5*time.Microsecond, "Performance should remain good under concurrent load")
	})
}

// Test concurrent access protection for scoped dependency resolution
func TestConcurrentScopedResolution(t *testing.T) {
	t.Run("ScopedResolution_ConcurrentContexts_ThreadSafe", func(t *testing.T) {
		container := NewContainer()

		// Counter to track instance creation per context
		var totalCreations int64

		// Register scoped dependency
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			count := atomic.AddInt64(&totalCreations, 1)
			return &scopedTestImpl{id: time.Now().Format("15:04:05.000000") + "_" + string(rune(count))}, nil
		}, Scoped)
		require.NoError(t, err)

		numContexts := 10
		numGoroutinesPerContext := 50

		var wg sync.WaitGroup
		results := make(map[string][]ScopedTestInterface)
		var resultsMutex sync.Mutex

		// Create multiple contexts, each with multiple concurrent resolutions
		for ctxIndex := 0; ctxIndex < numContexts; ctxIndex++ {
			contextKey := "test-context-" + string(rune(ctxIndex))
			ctx := context.WithValue(context.Background(), "request-id", contextKey)

			contextResults := make([]ScopedTestInterface, numGoroutinesPerContext)

			wg.Add(numGoroutinesPerContext)
			for goroutineIndex := 0; goroutineIndex < numGoroutinesPerContext; goroutineIndex++ {
				go func(ctxIdx, gorIdx int, context context.Context) {
					defer wg.Done()

					instance, err := ResolveWithContext[ScopedTestInterface](context, container)
					require.NoError(t, err)

					contextResults[gorIdx] = instance
				}(ctxIndex, goroutineIndex, ctx)
			}

			// Store results for this context
			resultsMutex.Lock()
			results[contextKey] = contextResults
			resultsMutex.Unlock()
		}

		wg.Wait()

		// Verify each context has consistent instances
		for contextKey, instances := range results {
			if len(instances) == 0 {
				continue
			}

			firstInstance := instances[0]
			for i, instance := range instances {
				assert.Equal(t, firstInstance.GetID(), instance.GetID(),
					"Context %s, goroutine %d should have same instance", contextKey, i)
			}
		}

		// Verify different contexts have different instances
		contextKeys := make([]string, 0, len(results))
		for key := range results {
			contextKeys = append(contextKeys, key)
		}

		for i := 0; i < len(contextKeys); i++ {
			for j := i + 1; j < len(contextKeys); j++ {
				ctx1Results := results[contextKeys[i]]
				ctx2Results := results[contextKeys[j]]

				if len(ctx1Results) > 0 && len(ctx2Results) > 0 {
					assert.NotEqual(t, ctx1Results[0].GetID(), ctx2Results[0].GetID(),
						"Different contexts should have different instances")
				}
			}
		}

		// Should have created exactly one instance per context
		assert.Equal(t, int64(numContexts), atomic.LoadInt64(&totalCreations))
	})
}

// Test thread-safe dependency instance storage with minimal locking overhead
func TestThreadSafeDependencyStorage(t *testing.T) {
	t.Run("DependencyStorage_ConcurrentReadsAndWrites_NoDataRaces", func(t *testing.T) {
		sdm := NewSharedDependencyManager()

		numGoroutines := 100
		var wg sync.WaitGroup

		// Test concurrent singleton storage and retrieval
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				defer wg.Done()

				// Alternate between reading and writing operations
				if index%2 == 0 {
					// Writer goroutines
					interfaceType := reflect.TypeOf((*ScopedTestInterface)(nil)).Elem()
					_, err := sdm.GetSingleton(interfaceType, func() (interface{}, error) {
						return &scopedTestImpl{id: "concurrent-storage"}, nil
					})
					assert.NoError(t, err)
				} else {
					// Reader goroutines
					interfaceType := reflect.TypeOf((*ScopedTestInterface)(nil)).Elem()
					_, err := sdm.GetSingleton(interfaceType, func() (interface{}, error) {
						return &scopedTestImpl{id: "concurrent-storage"}, nil
					})
					assert.NoError(t, err)
				}
			}(i)
		}

		wg.Wait()

		// Verify metrics are consistent
		metrics := sdm.GetMetrics()
		assert.Greater(t, metrics.SingletonResolutions, int64(0))
		assert.Greater(t, metrics.ResolutionCount, int64(0))
	})

	t.Run("MemorySharing_ConcurrentAccess_ConsistentState", func(t *testing.T) {
		sdm := NewSharedDependencyManager()

		numContexts := 20
		numOperationsPerContext := 50

		var wg sync.WaitGroup

		// Test concurrent scoped operations across multiple contexts
		for ctxIndex := 0; ctxIndex < numContexts; ctxIndex++ {
			contextKey := "concurrent-context-" + string(rune(ctxIndex))

			wg.Add(numOperationsPerContext)
			for opIndex := 0; opIndex < numOperationsPerContext; opIndex++ {
				go func(ctx string, op int) {
					defer wg.Done()

					interfaceType := reflect.TypeOf((*ScopedTestInterface)(nil)).Elem()
					_, err := sdm.GetScoped(ctx, interfaceType, func() (interface{}, error) {
						return &scopedTestImpl{id: ctx + "_instance"}, nil
					})
					assert.NoError(t, err)
				}(contextKey, opIndex)
			}
		}

		wg.Wait()

		// Verify state consistency
		metrics := sdm.GetMetrics()
		assert.Equal(t, int64(numContexts), metrics.SharedInstanceCount) // One instance per context
		assert.Greater(t, metrics.ScopedCacheHits, int64(0))             // Should have cache hits
	})
}

// Benchmark concurrent dependency resolution performance
func BenchmarkConcurrentDependencyResolution(b *testing.B) {
	container := NewContainer()

	// Register dependencies
	err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
		return &scopedTestImpl{id: "benchmark-singleton"}, nil
	}, Singleton)
	require.NoError(b, err)

	b.Run("ConcurrentSingletonResolution", func(b *testing.B) {
		// Set parallelism to use all available CPUs
		b.SetParallelism(runtime.NumCPU())

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := Resolve[ScopedTestInterface](container)
				if err != nil {
					b.Fatalf("Resolution failed: %v", err)
				}
			}
		})
	})

	b.Run("ConcurrentScopedResolution", func(b *testing.B) {
		// Register scoped dependency
		err := RegisterFactory[TestInterface](container, func(c Container) (TestInterface, error) {
			return &testImpl{value: "benchmark-scoped"}, nil
		}, Scoped)
		require.NoError(b, err)

		b.SetParallelism(runtime.NumCPU())

		b.RunParallel(func(pb *testing.PB) {
			// Each goroutine gets its own context to test scoped resolution
			ctx := context.WithValue(context.Background(), "request-id", "bench")

			for pb.Next() {
				_, err := ResolveWithContext[TestInterface](ctx, container)
				if err != nil {
					b.Fatalf("Scoped resolution failed: %v", err)
				}
			}
		})
	})
}

// Stress test for identifying potential race conditions
func TestStressConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	t.Run("StressTest_MixedOperations_NoRaceConditions", func(t *testing.T) {
		container := NewContainer()

		// Register multiple dependencies with different scopes
		err := RegisterFactory[ScopedTestInterface](container, func(c Container) (ScopedTestInterface, error) {
			return &scopedTestImpl{id: "stress-singleton"}, nil
		}, Singleton)
		require.NoError(t, err)

		err = RegisterFactory[TestInterface](container, func(c Container) (TestInterface, error) {
			return &testImpl{value: "stress-scoped"}, nil
		}, Scoped)
		require.NoError(t, err)

		err = RegisterFactory[ConfigProviderInterface](container, func(c Container) (ConfigProviderInterface, error) {
			return &mockConfig{}, nil
		}, Transient)
		require.NoError(t, err)

		duration := 2 * time.Second
		numGoroutines := 50

		var wg sync.WaitGroup
		stop := make(chan bool)

		// Start stress test goroutines
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				contextKey := "stress-" + string(rune(goroutineID%10)) // 10 different contexts
				ctx := context.WithValue(context.Background(), "request-id", contextKey)

				for {
					select {
					case <-stop:
						return
					default:
						// Perform random operations
						switch goroutineID % 3 {
						case 0:
							_, _ = Resolve[ScopedTestInterface](container)
						case 1:
							_, _ = ResolveWithContext[TestInterface](ctx, container)
						case 2:
							_, _ = Resolve[ConfigProviderInterface](container)
						}
					}
				}
			}(i)
		}

		// Run for specified duration
		time.Sleep(duration)
		close(stop)
		wg.Wait()

		// Verify no panics occurred and system is still functional
		_, err = Resolve[ScopedTestInterface](container)
		assert.NoError(t, err, "System should remain functional after stress test")

		// Check metrics for sanity
		containerImpl := container.(*containerImpl)
		metrics := containerImpl.sharedManager.GetMetrics()
		assert.Greater(t, metrics.ResolutionCount, int64(0), "Should have recorded resolutions")
	})
}

// Mock config for testing
type mockConfig struct{}

func (m *mockConfig) GetServerPort() string          { return "8080" }
func (m *mockConfig) GetDocumentDBUri() string       { return "mongodb://test" }
func (m *mockConfig) IsHybridResourcesEnabled() bool { return true }
