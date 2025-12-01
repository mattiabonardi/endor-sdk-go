package di

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock types for memory testing
type MockExpensiveResource struct {
	ID      string
	Data    []byte
	Created time.Time
}

func NewMockExpensiveResource(id string) *MockExpensiveResource {
	return &MockExpensiveResource{
		ID:      id,
		Data:    make([]byte, 1024), // 1KB of data
		Created: time.Now(),
	}
}

type MockExpensiveResourceInterface interface {
	GetID() string
	GetSize() int
}

func (m *MockExpensiveResource) GetID() string {
	return m.ID
}

func (m *MockExpensiveResource) GetSize() int {
	return len(m.Data)
}

func TestMemoryTracking(t *testing.T) {
	t.Run("TrackDependencyCreation_RecordsMemoryUsage", func(t *testing.T) {
		tracker := NewMemoryTracker()
		resource := NewMockExpensiveResource("test-1")

		tracker.TrackDependencyCreation("test-resource", resource, Singleton)

		stats := tracker.GetMemoryStats()
		assert.Len(t, stats, 1)
		assert.Contains(t, stats, "test-resource")

		resourceStats := stats["test-resource"]
		assert.Equal(t, "test-resource", resourceStats.DependencyType)
		assert.Equal(t, Singleton, resourceStats.Scope)
		assert.Equal(t, int64(1), resourceStats.ReferenceCount)
		assert.True(t, resourceStats.AllocationSize > 0, "Should track non-zero allocation size")
		assert.Equal(t, int64(0), resourceStats.ShareCount)
	})

	t.Run("TrackDependencyShare_IncrementsShareCount", func(t *testing.T) {
		tracker := NewMemoryTracker()
		resource := NewMockExpensiveResource("test-1")

		// First create the dependency
		tracker.TrackDependencyCreation("test-resource", resource, Singleton)

		// Then track several shares
		for i := 0; i < 5; i++ {
			tracker.TrackDependencyShare("test-resource")
		}

		stats := tracker.GetMemoryStats()
		resourceStats := stats["test-resource"]
		assert.Equal(t, int64(5), resourceStats.ShareCount)
	})

	t.Run("GetSharingEfficiency_CalculatesCorrectRatio", func(t *testing.T) {
		tracker := NewMemoryTracker()

		// Create 2 dependencies
		tracker.TrackDependencyCreation("resource-1", "test", Singleton)
		tracker.TrackDependencyCreation("resource-2", "test", Singleton)

		// Track 8 shares total
		for i := 0; i < 5; i++ {
			tracker.TrackDependencyShare("resource-1")
		}
		for i := 0; i < 3; i++ {
			tracker.TrackDependencyShare("resource-2")
		}

		efficiency := tracker.GetSharingEfficiency()
		// 8 shares / (8 shares + 2 creations) = 0.8
		assert.InDelta(t, 0.8, efficiency, 0.01)
	})

	t.Run("GetMemoryOptimizationReport_ProvidesDetailedAnalysis", func(t *testing.T) {
		tracker := NewMemoryTracker()

		// Create a large resource
		resource := NewMockExpensiveResource("large-resource")
		tracker.TrackDependencyCreation("large-resource", resource, Singleton)

		// Track multiple shares
		for i := 0; i < 10; i++ {
			tracker.TrackDependencyShare("large-resource")
		}

		report := tracker.GetMemoryOptimizationReport()
		assert.Equal(t, int64(10), report.TotalShares)
		assert.Equal(t, int64(1), report.TotalCreations)
		assert.Equal(t, 10.0/11.0, report.SharingEfficiency) // 10/(10+1)
		assert.Len(t, report.DependencyDetails, 1)

		detail := report.DependencyDetails[0]
		assert.Equal(t, "large-resource", detail.DependencyType)
		assert.Equal(t, int64(10), detail.ShareCount)
		assert.Equal(t, Singleton, detail.Scope)

		// Memory saved should be allocation_size * share_count
		expectedSaved := detail.AllocationSize * uint64(detail.ShareCount)
		assert.Equal(t, expectedSaved, detail.MemorySaved)
		assert.Equal(t, expectedSaved, report.TotalMemorySaved)
	})

	t.Run("ReleaseDependencyReference_CleansUpMemoryTracking", func(t *testing.T) {
		tracker := NewMemoryTracker()
		resource := NewMockExpensiveResource("test-resource")

		tracker.TrackDependencyCreation("test-resource", resource, Singleton)
		initialAllocations := tracker.GetTotalAllocations()
		assert.True(t, initialAllocations > 0)

		tracker.ReleaseDependencyReference("test-resource")

		finalAllocations := tracker.GetTotalAllocations()
		assert.Equal(t, uint64(0), finalAllocations)

		stats := tracker.GetMemoryStats()
		assert.Len(t, stats, 0, "Should have removed dependency from tracking")
	})
}

func TestDependencyPoolManager(t *testing.T) {
	t.Run("RegisterPool_CreatesPoolForDependencyType", func(t *testing.T) {
		poolManager := NewDependencyPoolManager()

		poolManager.RegisterPool("expensive-resource", func() interface{} {
			return NewMockExpensiveResource("pooled")
		})

		// Should be able to get an instance from the pool
		instance, exists := poolManager.GetFromPool("expensive-resource")
		assert.True(t, exists)
		assert.NotNil(t, instance)

		resource, ok := instance.(*MockExpensiveResource)
		require.True(t, ok)
		assert.Equal(t, "pooled", resource.GetID())
	})

	t.Run("ReturnToPool_ReusesInstances", func(t *testing.T) {
		poolManager := NewDependencyPoolManager()

		poolManager.RegisterPool("expensive-resource", func() interface{} {
			return NewMockExpensiveResource("pooled")
		})

		// Get an instance
		instance1, exists := poolManager.GetFromPool("expensive-resource")
		require.True(t, exists)

		// Modify it
		resource1 := instance1.(*MockExpensiveResource)
		resource1.ID = "modified"

		// Return it to the pool
		success := poolManager.ReturnToPool("expensive-resource", instance1)
		assert.True(t, success)

		// Get another instance - should be the same (reused)
		instance2, exists := poolManager.GetFromPool("expensive-resource")
		require.True(t, exists)

		resource2 := instance2.(*MockExpensiveResource)
		assert.Equal(t, "modified", resource2.ID, "Should reuse the same instance")
	})

	t.Run("GetFromPool_NonExistentPool_ReturnsFalse", func(t *testing.T) {
		poolManager := NewDependencyPoolManager()

		instance, exists := poolManager.GetFromPool("non-existent")
		assert.False(t, exists)
		assert.Nil(t, instance)
	})

	t.Run("ReturnToPool_NonExistentPool_ReturnsFalse", func(t *testing.T) {
		poolManager := NewDependencyPoolManager()

		success := poolManager.ReturnToPool("non-existent", "some-instance")
		assert.False(t, success)
	})
}

func TestMemoryProfiler(t *testing.T) {
	t.Run("TakeSnapshot_CapturesMemoryState", func(t *testing.T) {
		profiler := NewMemoryProfiler(true)
		tracker := NewMemoryTracker()

		// Create some dependencies to track
		resource1 := NewMockExpensiveResource("resource-1")
		resource2 := NewMockExpensiveResource("resource-2")
		tracker.TrackDependencyCreation("resource-1", resource1, Singleton)
		tracker.TrackDependencyCreation("resource-2", resource2, Singleton)

		profiler.TakeSnapshot(tracker)

		snapshots := profiler.GetSnapshots()
		assert.Len(t, snapshots, 1)

		snapshot := snapshots[0]
		assert.True(t, snapshot.HeapAlloc > 0)
		assert.True(t, snapshot.HeapSys > 0)
		assert.True(t, snapshot.DependencyMem > 0)
		assert.Len(t, snapshot.Dependencies, 2)
		assert.Contains(t, snapshot.Dependencies, "resource-1")
		assert.Contains(t, snapshot.Dependencies, "resource-2")
	})

	t.Run("AnalyzeTrend_DetectsMemoryGrowth", func(t *testing.T) {
		profiler := NewMemoryProfiler(true)
		tracker := NewMemoryTracker()

		// Take initial snapshot
		profiler.TakeSnapshot(tracker)

		// Add more dependencies
		for i := 0; i < 10; i++ {
			resource := NewMockExpensiveResource("resource")
			tracker.TrackDependencyCreation("resource", resource, Singleton)
		}

		// Take second snapshot
		profiler.TakeSnapshot(tracker)

		analysis := profiler.AnalyzeTrend()
		assert.Equal(t, "increasing", analysis.DependencyTrend)
		assert.True(t, analysis.DependencyMemChange > 0)
		assert.Equal(t, 2, analysis.SnapshotCount)
	})

	t.Run("GetLatestSnapshot_ReturnsNilWhenEmpty", func(t *testing.T) {
		profiler := NewMemoryProfiler(true)

		latest := profiler.GetLatestSnapshot()
		assert.Nil(t, latest)
	})

	t.Run("DisabledProfiler_DoesNotCaptureSnapshots", func(t *testing.T) {
		profiler := NewMemoryProfiler(false) // Disabled
		tracker := NewMemoryTracker()

		profiler.TakeSnapshot(tracker)

		snapshots := profiler.GetSnapshots()
		assert.Len(t, snapshots, 0)
	})
}

func TestContainerMemoryIntegration(t *testing.T) {
	t.Run("Container_TracksMemoryUsageForDependencies", func(t *testing.T) {
		container := NewContainer()
		memoryTracker := container.GetMemoryTracker()

		// Register expensive resource as singleton
		resource := NewMockExpensiveResource("singleton-resource")
		err := Register[MockExpensiveResourceInterface](container, resource, Singleton)
		require.NoError(t, err)

		// Resolve it multiple times
		for i := 0; i < 5; i++ {
			resolved, err := Resolve[MockExpensiveResourceInterface](container)
			require.NoError(t, err)
			require.NotNil(t, resolved)
			assert.Equal(t, "singleton-resource", resolved.GetID())
		}

		// Check memory tracking
		stats := memoryTracker.GetMemoryStats()
		assert.Len(t, stats, 1)

		dependencyType := "di.MockExpensiveResourceInterface"
		assert.Contains(t, stats, dependencyType)

		resourceStats := stats[dependencyType]
		assert.Equal(t, int64(1), resourceStats.ReferenceCount)
		assert.Equal(t, int64(4), resourceStats.ShareCount) // 5 resolutions - 1 creation = 4 shares
		assert.True(t, resourceStats.AllocationSize > 0)

		// Check sharing efficiency
		efficiency := memoryTracker.GetSharingEfficiency()
		assert.Equal(t, 4.0/5.0, efficiency) // 4 shares out of 5 total accesses
	})

	t.Run("Container_TracksTransientDependencies", func(t *testing.T) {
		container := NewContainer()
		memoryTracker := container.GetMemoryTracker()

		// Register factory for transient resource
		err := RegisterFactory[MockExpensiveResourceInterface](container, func(c Container) (MockExpensiveResourceInterface, error) {
			return NewMockExpensiveResource("transient-resource"), nil
		}, Transient)
		require.NoError(t, err)

		// Resolve multiple times
		for i := 0; i < 3; i++ {
			resolved, err := Resolve[MockExpensiveResourceInterface](container)
			require.NoError(t, err)
			require.NotNil(t, resolved)
		}

		// For transient dependencies, each resolution creates a new instance
		report := memoryTracker.GetMemoryOptimizationReport()
		assert.Equal(t, int64(3), report.TotalCreations)
		assert.Equal(t, int64(0), report.TotalShares) // No sharing for transient
		assert.Equal(t, 0.0, report.SharingEfficiency)
	})

	t.Run("Container_MemoryProfilerIntegration", func(t *testing.T) {
		container := NewContainer()
		profiler := container.GetMemoryProfiler()
		tracker := container.GetMemoryTracker()

		// Register and resolve some dependencies
		resource := NewMockExpensiveResource("profiled-resource")
		err := Register[MockExpensiveResourceInterface](container, resource, Singleton)
		require.NoError(t, err)

		profiler.TakeSnapshot(tracker)

		// Resolve the dependency
		_, err = Resolve[MockExpensiveResourceInterface](container)
		require.NoError(t, err)

		profiler.TakeSnapshot(tracker)

		snapshots := profiler.GetSnapshots()
		assert.Len(t, snapshots, 2)

		// First snapshot should have no dependency memory
		assert.Equal(t, uint64(0), snapshots[0].DependencyMem)

		// Second snapshot should have dependency memory
		assert.True(t, snapshots[1].DependencyMem > 0)
		assert.Len(t, snapshots[1].Dependencies, 1)
	})
}
