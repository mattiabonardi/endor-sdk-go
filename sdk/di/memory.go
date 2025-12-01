package di

import (
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// MemoryStats tracks memory usage for dependency sharing
type MemoryStats struct {
	// AllocationSize is the estimated memory size of the dependency instance
	AllocationSize uint64
	// ReferenceCount is the number of active references to this dependency
	ReferenceCount int64
	// CreatedAt is when this dependency was first created
	CreatedAt time.Time
	// LastAccessed is the last time this dependency was accessed
	LastAccessed time.Time
	// ShareCount is the number of times this dependency was shared vs created new
	ShareCount int64
	// Type information
	DependencyType string
	Scope          Scope
}

// MemoryTracker tracks memory usage and optimization metrics for dependencies
type MemoryTracker struct {
	// dependencyStats maps dependency type to memory statistics
	dependencyStats map[string]*MemoryStats
	// totalAllocations tracks total memory allocated for dependencies
	totalAllocations uint64
	// totalShares tracks how many times dependencies were shared vs created
	totalShares int64
	// totalCreations tracks total dependency creations
	totalCreations int64
	mu             sync.RWMutex
}

// NewMemoryTracker creates a new memory tracker
func NewMemoryTracker() *MemoryTracker {
	return &MemoryTracker{
		dependencyStats: make(map[string]*MemoryStats),
	}
}

// TrackDependencyCreation records the creation of a new dependency
func (mt *MemoryTracker) TrackDependencyCreation(dependencyType string, instance interface{}, scope Scope) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	size := estimateMemorySize(instance)
	mt.totalAllocations += size
	mt.totalCreations++

	stats := mt.dependencyStats[dependencyType]
	if stats == nil {
		stats = &MemoryStats{
			DependencyType: dependencyType,
			Scope:          scope,
			CreatedAt:      time.Now(),
		}
		mt.dependencyStats[dependencyType] = stats
	}

	stats.AllocationSize = size
	stats.ReferenceCount++
	stats.LastAccessed = time.Now()
}

// TrackDependencyShare records when a dependency is shared instead of created
func (mt *MemoryTracker) TrackDependencyShare(dependencyType string) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.totalShares++

	if stats, exists := mt.dependencyStats[dependencyType]; exists {
		stats.ShareCount++
		stats.LastAccessed = time.Now()
	}
}

// TrackDependencyAccess records access to a dependency (for LRU tracking)
func (mt *MemoryTracker) TrackDependencyAccess(dependencyType string) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if stats, exists := mt.dependencyStats[dependencyType]; exists {
		stats.LastAccessed = time.Now()
	}
}

// ReleaseDependencyReference decrements reference count
func (mt *MemoryTracker) ReleaseDependencyReference(dependencyType string) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if stats, exists := mt.dependencyStats[dependencyType]; exists {
		stats.ReferenceCount--
		if stats.ReferenceCount <= 0 {
			mt.totalAllocations -= stats.AllocationSize
			delete(mt.dependencyStats, dependencyType)
		}
	}
}

// GetMemoryStats returns current memory statistics
func (mt *MemoryTracker) GetMemoryStats() map[string]*MemoryStats {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	statsCopy := make(map[string]*MemoryStats)
	for k, v := range mt.dependencyStats {
		statsCopy[k] = &MemoryStats{
			AllocationSize: v.AllocationSize,
			ReferenceCount: v.ReferenceCount,
			CreatedAt:      v.CreatedAt,
			LastAccessed:   v.LastAccessed,
			ShareCount:     v.ShareCount,
			DependencyType: v.DependencyType,
			Scope:          v.Scope,
		}
	}
	return statsCopy
}

// GetTotalAllocations returns the total memory allocated for dependencies
func (mt *MemoryTracker) GetTotalAllocations() uint64 {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.totalAllocations
}

// GetSharingEfficiency returns the ratio of shares to total accesses
func (mt *MemoryTracker) GetSharingEfficiency() float64 {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	totalAccesses := mt.totalShares + mt.totalCreations
	if totalAccesses == 0 {
		return 0.0
	}
	return float64(mt.totalShares) / float64(totalAccesses)
}

// GetMemoryOptimizationReport returns a detailed memory optimization report
func (mt *MemoryTracker) GetMemoryOptimizationReport() MemoryOptimizationReport {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	report := MemoryOptimizationReport{
		TotalAllocations:  mt.totalAllocations,
		TotalShares:       mt.totalShares,
		TotalCreations:    mt.totalCreations,
		SharingEfficiency: mt.GetSharingEfficiency(),
		DependencyDetails: make([]DependencyMemoryDetail, 0, len(mt.dependencyStats)),
		Timestamp:         time.Now(),
	}

	var totalSavings uint64
	for _, stats := range mt.dependencyStats {
		potentialAllocations := uint64(stats.ShareCount + 1) // +1 for the initial creation
		actualAllocations := uint64(1)                       // Only one actual allocation for shared deps
		savings := (potentialAllocations - actualAllocations) * stats.AllocationSize
		totalSavings += savings

		detail := DependencyMemoryDetail{
			DependencyType: stats.DependencyType,
			AllocationSize: stats.AllocationSize,
			ReferenceCount: stats.ReferenceCount,
			ShareCount:     stats.ShareCount,
			MemorySaved:    savings,
			Scope:          stats.Scope,
			CreatedAt:      stats.CreatedAt,
			LastAccessed:   stats.LastAccessed,
		}
		report.DependencyDetails = append(report.DependencyDetails, detail)
	}

	report.TotalMemorySaved = totalSavings
	return report
}

// MemoryOptimizationReport provides detailed memory usage analysis
type MemoryOptimizationReport struct {
	TotalAllocations  uint64                   `json:"total_allocations"`
	TotalShares       int64                    `json:"total_shares"`
	TotalCreations    int64                    `json:"total_creations"`
	SharingEfficiency float64                  `json:"sharing_efficiency"`
	TotalMemorySaved  uint64                   `json:"total_memory_saved"`
	DependencyDetails []DependencyMemoryDetail `json:"dependency_details"`
	Timestamp         time.Time                `json:"timestamp"`
}

// DependencyMemoryDetail provides detailed memory information for a specific dependency
type DependencyMemoryDetail struct {
	DependencyType string    `json:"dependency_type"`
	AllocationSize uint64    `json:"allocation_size"`
	ReferenceCount int64     `json:"reference_count"`
	ShareCount     int64     `json:"share_count"`
	MemorySaved    uint64    `json:"memory_saved"`
	Scope          Scope     `json:"scope"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessed   time.Time `json:"last_accessed"`
}

// estimateMemorySize provides a rough estimate of memory usage for an interface{}
func estimateMemorySize(instance interface{}) uint64 {
	if instance == nil {
		return 0
	}

	// Basic size estimation - this is approximate
	baseSize := uint64(unsafe.Sizeof(instance))

	// For known types, provide more accurate estimates
	switch v := instance.(type) {
	case string:
		return baseSize + uint64(len(v))
	case []byte:
		return baseSize + uint64(len(v))
	case map[string]interface{}:
		// Rough estimate for maps
		return baseSize + uint64(len(v))*32 // Assume 32 bytes per entry
	default:
		// For other types, use sizeof plus some overhead
		return baseSize + 64 // Add some overhead for pointers and structure
	}
}

// DependencyPool manages a pool of reusable dependency instances
type DependencyPool struct {
	pool sync.Pool
	new  func() interface{}
}

// NewDependencyPool creates a new dependency pool
func NewDependencyPool(newFunc func() interface{}) *DependencyPool {
	return &DependencyPool{
		pool: sync.Pool{
			New: newFunc,
		},
		new: newFunc,
	}
}

// Get retrieves an instance from the pool
func (dp *DependencyPool) Get() interface{} {
	return dp.pool.Get()
}

// Put returns an instance to the pool
func (dp *DependencyPool) Put(instance interface{}) {
	dp.pool.Put(instance)
}

// DependencyPoolManager manages multiple dependency pools
type DependencyPoolManager struct {
	pools map[string]*DependencyPool
	mu    sync.RWMutex
}

// NewDependencyPoolManager creates a new pool manager
func NewDependencyPoolManager() *DependencyPoolManager {
	return &DependencyPoolManager{
		pools: make(map[string]*DependencyPool),
	}
}

// RegisterPool registers a pool for a specific dependency type
func (dpm *DependencyPoolManager) RegisterPool(dependencyType string, newFunc func() interface{}) {
	dpm.mu.Lock()
	defer dpm.mu.Unlock()
	dpm.pools[dependencyType] = NewDependencyPool(newFunc)
}

// GetFromPool retrieves an instance from the pool for the given dependency type
func (dpm *DependencyPoolManager) GetFromPool(dependencyType string) (interface{}, bool) {
	dpm.mu.RLock()
	pool, exists := dpm.pools[dependencyType]
	dpm.mu.RUnlock()

	if !exists {
		return nil, false
	}

	return pool.Get(), true
}

// ReturnToPool returns an instance to the pool
func (dpm *DependencyPoolManager) ReturnToPool(dependencyType string, instance interface{}) bool {
	dpm.mu.RLock()
	pool, exists := dpm.pools[dependencyType]
	dpm.mu.RUnlock()

	if !exists {
		return false
	}

	pool.Put(instance)
	return true
}

// MemoryProfiler provides memory profiling tools for dependency analysis
type MemoryProfiler struct {
	enabled   bool
	snapshots []MemorySnapshot
	mu        sync.RWMutex
}

// MemorySnapshot captures memory state at a point in time
type MemorySnapshot struct {
	Timestamp     time.Time               `json:"timestamp"`
	HeapAlloc     uint64                  `json:"heap_alloc"`
	HeapSys       uint64                  `json:"heap_sys"`
	NumGC         uint32                  `json:"num_gc"`
	DependencyMem uint64                  `json:"dependency_mem"`
	Dependencies  map[string]*MemoryStats `json:"dependencies"`
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(enabled bool) *MemoryProfiler {
	return &MemoryProfiler{
		enabled:   enabled,
		snapshots: make([]MemorySnapshot, 0),
	}
}

// TakeSnapshot captures current memory state
func (mp *MemoryProfiler) TakeSnapshot(memoryTracker *MemoryTracker) {
	if !mp.enabled {
		return
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	snapshot := MemorySnapshot{
		Timestamp:     time.Now(),
		HeapAlloc:     ms.HeapAlloc,
		HeapSys:       ms.HeapSys,
		NumGC:         ms.NumGC,
		DependencyMem: memoryTracker.GetTotalAllocations(),
		Dependencies:  memoryTracker.GetMemoryStats(),
	}

	mp.snapshots = append(mp.snapshots, snapshot)

	// Keep only the last 100 snapshots to prevent memory bloat
	if len(mp.snapshots) > 100 {
		mp.snapshots = mp.snapshots[len(mp.snapshots)-100:]
	}
}

// GetSnapshots returns all memory snapshots
func (mp *MemoryProfiler) GetSnapshots() []MemorySnapshot {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	snapshots := make([]MemorySnapshot, len(mp.snapshots))
	copy(snapshots, mp.snapshots)
	return snapshots
}

// GetLatestSnapshot returns the most recent memory snapshot
func (mp *MemoryProfiler) GetLatestSnapshot() *MemorySnapshot {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if len(mp.snapshots) == 0 {
		return nil
	}

	snapshot := mp.snapshots[len(mp.snapshots)-1]
	return &snapshot
}

// AnalyzeTrend analyzes memory usage trends over time
func (mp *MemoryProfiler) AnalyzeTrend() MemoryTrendAnalysis {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if len(mp.snapshots) < 2 {
		return MemoryTrendAnalysis{
			TrendDirection: "insufficient_data",
		}
	}

	first := mp.snapshots[0]
	last := mp.snapshots[len(mp.snapshots)-1]

	heapTrend := "stable"
	if last.HeapAlloc > first.HeapAlloc*11/10 { // >10% increase
		heapTrend = "increasing"
	} else if last.HeapAlloc < first.HeapAlloc*9/10 { // >10% decrease
		heapTrend = "decreasing"
	}

	dependencyTrend := "stable"
	if last.DependencyMem > first.DependencyMem*11/10 {
		dependencyTrend = "increasing"
	} else if last.DependencyMem < first.DependencyMem*9/10 {
		dependencyTrend = "decreasing"
	}

	return MemoryTrendAnalysis{
		TrendDirection:      heapTrend,
		DependencyTrend:     dependencyTrend,
		HeapAllocChange:     int64(last.HeapAlloc) - int64(first.HeapAlloc),
		DependencyMemChange: int64(last.DependencyMem) - int64(first.DependencyMem),
		TimeSpan:            last.Timestamp.Sub(first.Timestamp),
		SnapshotCount:       len(mp.snapshots),
	}
}

// MemoryTrendAnalysis provides analysis of memory usage trends
type MemoryTrendAnalysis struct {
	TrendDirection      string        `json:"trend_direction"`
	DependencyTrend     string        `json:"dependency_trend"`
	HeapAllocChange     int64         `json:"heap_alloc_change"`
	DependencyMemChange int64         `json:"dependency_mem_change"`
	TimeSpan            time.Duration `json:"time_span"`
	SnapshotCount       int           `json:"snapshot_count"`
}
