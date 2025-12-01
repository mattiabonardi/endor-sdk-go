package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mattiabonardi/endor-sdk-go/sdk/health"
)

// SharedDependencyManager manages scoped dependency lifecycle coordination
type SharedDependencyManager struct {
	// singletonCache stores singleton instances across application lifetime
	singletonCache map[reflect.Type]interface{}
	// scopedCache stores scoped instances within request/operation context
	scopedCache map[string]map[reflect.Type]interface{}
	// healthCheckers stores dependency health monitors
	healthCheckers map[reflect.Type]health.HealthChecker
	// metrics tracks dependency sharing metrics
	metrics *DependencyMetrics
	// mutex protects concurrent access to shared state
	mutex sync.RWMutex
	// scopedMutexes protects individual scoped context caches
	scopedMutexes sync.Map
}

// DependencyMetrics tracks metrics for dependency sharing and resolution
type DependencyMetrics struct {
	// Total resolutions by scope
	SingletonResolutions int64
	ScopedResolutions    int64
	TransientResolutions int64

	// Cache hit rates
	SingletonCacheHits int64
	ScopedCacheHits    int64

	// Performance metrics
	AverageResolutionTime time.Duration
	TotalResolutionTime   time.Duration
	ResolutionCount       int64

	// Memory optimization metrics
	SharedInstanceCount int64
	DuplicatesPrevented int64

	mutex sync.RWMutex
}

// NewSharedDependencyManager creates a new shared dependency manager
func NewSharedDependencyManager() *SharedDependencyManager {
	return &SharedDependencyManager{
		singletonCache: make(map[reflect.Type]interface{}),
		scopedCache:    make(map[string]map[reflect.Type]interface{}),
		healthCheckers: make(map[reflect.Type]health.HealthChecker),
		metrics:        &DependencyMetrics{},
	}
}

// GetScoped retrieves or creates a scoped dependency instance
func (sdm *SharedDependencyManager) GetScoped(contextKey string, interfaceType reflect.Type, factory func() (interface{}, error)) (interface{}, error) {
	startTime := time.Now()
	defer func() {
		sdm.recordResolutionMetrics(startTime, "scoped")
	}()

	sdm.mutex.RLock()
	scopeCache, exists := sdm.scopedCache[contextKey]
	if exists {
		if instance, found := scopeCache[interfaceType]; found {
			sdm.mutex.RUnlock()
			atomic.AddInt64(&sdm.metrics.ScopedCacheHits, 1)
			return instance, nil
		}
	}
	sdm.mutex.RUnlock()

	// Need to create instance - acquire write lock
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	// Double-check pattern
	scopeCache, exists = sdm.scopedCache[contextKey]
	if !exists {
		sdm.scopedCache[contextKey] = make(map[reflect.Type]interface{})
		scopeCache = sdm.scopedCache[contextKey]
	} else if instance, found := scopeCache[interfaceType]; found {
		atomic.AddInt64(&sdm.metrics.ScopedCacheHits, 1)
		return instance, nil
	}

	// Create new instance
	instance, err := factory()
	if err != nil {
		return nil, fmt.Errorf("failed to create scoped instance for %s: %w", interfaceType.String(), err)
	}

	scopeCache[interfaceType] = instance
	atomic.AddInt64(&sdm.metrics.SharedInstanceCount, 1)
	return instance, nil
} // GetSingleton retrieves or creates a singleton dependency instance
func (sdm *SharedDependencyManager) GetSingleton(interfaceType reflect.Type, factory func() (interface{}, error)) (interface{}, error) {
	startTime := time.Now()
	defer func() {
		sdm.recordResolutionMetrics(startTime, "singleton")
	}()

	sdm.mutex.RLock()
	if instance, exists := sdm.singletonCache[interfaceType]; exists {
		sdm.mutex.RUnlock()
		atomic.AddInt64(&sdm.metrics.SingletonCacheHits, 1)
		return instance, nil
	}
	sdm.mutex.RUnlock()

	// Need to create instance - acquire write lock
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	// Double-check pattern
	if instance, exists := sdm.singletonCache[interfaceType]; exists {
		atomic.AddInt64(&sdm.metrics.SingletonCacheHits, 1)
		return instance, nil
	}

	// Create new instance
	instance, err := factory()
	if err != nil {
		return nil, fmt.Errorf("failed to create singleton instance for %s: %w", interfaceType.String(), err)
	}

	sdm.singletonCache[interfaceType] = instance
	atomic.AddInt64(&sdm.metrics.SharedInstanceCount, 1)
	return instance, nil
} // StoreSingleton stores a singleton instance in the cache
func (sdm *SharedDependencyManager) StoreSingleton(interfaceType reflect.Type, instance interface{}) error {
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	// Check if already exists (double-check pattern)
	if _, exists := sdm.singletonCache[interfaceType]; exists {
		return nil // Already stored, no error
	}

	sdm.singletonCache[interfaceType] = instance
	atomic.AddInt64(&sdm.metrics.SharedInstanceCount, 1)
	return nil
}

// StoreScoped stores a scoped instance in the cache
func (sdm *SharedDependencyManager) StoreScoped(contextKey string, interfaceType reflect.Type, instance interface{}) error {
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	// Ensure scope cache exists
	scopeCache, exists := sdm.scopedCache[contextKey]
	if !exists {
		sdm.scopedCache[contextKey] = make(map[reflect.Type]interface{})
		scopeCache = sdm.scopedCache[contextKey]
	}

	// Check if already exists (double-check pattern)
	if _, found := scopeCache[interfaceType]; found {
		return nil // Already stored, no error
	}

	scopeCache[interfaceType] = instance
	atomic.AddInt64(&sdm.metrics.SharedInstanceCount, 1)
	return nil
}

// ClearScope removes all dependencies for a specific scope context
func (sdm *SharedDependencyManager) ClearScope(contextKey string) {
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	delete(sdm.scopedCache, contextKey)
}

// GetExistingSingleton retrieves a singleton instance if it exists, returns nil if not found
func (sdm *SharedDependencyManager) GetExistingSingleton(interfaceType reflect.Type) interface{} {
	sdm.mutex.RLock()
	defer sdm.mutex.RUnlock()

	if instance, exists := sdm.singletonCache[interfaceType]; exists {
		atomic.AddInt64(&sdm.metrics.SingletonCacheHits, 1)
		return instance
	}
	return nil
}

// SetSingleton stores a singleton instance (assumes container lock is held)
func (sdm *SharedDependencyManager) SetSingleton(interfaceType reflect.Type, instance interface{}) {
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	// Double-check pattern
	if _, exists := sdm.singletonCache[interfaceType]; !exists {
		sdm.singletonCache[interfaceType] = instance
		atomic.AddInt64(&sdm.metrics.SharedInstanceCount, 1)
	}
}

// GetExistingScoped retrieves a scoped instance if it exists, returns nil if not found
func (sdm *SharedDependencyManager) GetExistingScoped(contextKey string, interfaceType reflect.Type) interface{} {
	sdm.mutex.RLock()
	defer sdm.mutex.RUnlock()

	if scopeCache, exists := sdm.scopedCache[contextKey]; exists {
		if instance, found := scopeCache[interfaceType]; found {
			atomic.AddInt64(&sdm.metrics.ScopedCacheHits, 1)
			return instance
		}
	}
	return nil
}

// SetScoped stores a scoped instance (assumes container lock is held)
func (sdm *SharedDependencyManager) SetScoped(contextKey string, interfaceType reflect.Type, instance interface{}) {
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	// Ensure scope cache exists
	scopeCache, exists := sdm.scopedCache[contextKey]
	if !exists {
		sdm.scopedCache[contextKey] = make(map[reflect.Type]interface{})
		scopeCache = sdm.scopedCache[contextKey]
	}

	// Double-check pattern
	if _, found := scopeCache[interfaceType]; !found {
		scopeCache[interfaceType] = instance
		atomic.AddInt64(&sdm.metrics.SharedInstanceCount, 1)
	}
}

// RegisterHealthChecker registers a health checker for a dependency type
func (sdm *SharedDependencyManager) RegisterHealthChecker(interfaceType reflect.Type, checker health.HealthChecker) {
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	sdm.healthCheckers[interfaceType] = checker
}

// CheckHealth performs health checks on all registered dependencies
func (sdm *SharedDependencyManager) CheckHealth() map[reflect.Type]error {
	sdm.mutex.RLock()
	defer sdm.mutex.RUnlock()

	results := make(map[reflect.Type]error)
	for interfaceType, checker := range sdm.healthCheckers {
		results[interfaceType] = checker.HealthCheck()
	}

	return results
}

// ContextKeyGenerator generates unique context keys for scoped dependencies
type ContextKeyGenerator interface {
	GenerateKey(ctx context.Context) string
}

// DefaultContextKeyGenerator uses request ID or correlation ID from context
type DefaultContextKeyGenerator struct{}

func (g *DefaultContextKeyGenerator) GenerateKey(ctx context.Context) string {
	// Try to extract request ID from context
	if requestID, ok := ctx.Value("request-id").(string); ok {
		return requestID
	}
	if correlationID, ok := ctx.Value("correlation-id").(string); ok {
		return correlationID
	}
	// Fallback to context pointer as string (not ideal but functional)
	return fmt.Sprintf("ctx_%p", ctx)
}

// recordResolutionMetrics records performance metrics for dependency resolution
func (sdm *SharedDependencyManager) recordResolutionMetrics(startTime time.Time, scope string) {
	duration := time.Since(startTime)

	atomic.AddInt64(&sdm.metrics.ResolutionCount, 1)
	atomic.AddInt64((*int64)(&sdm.metrics.TotalResolutionTime), int64(duration))

	switch scope {
	case "singleton":
		atomic.AddInt64(&sdm.metrics.SingletonResolutions, 1)
	case "scoped":
		atomic.AddInt64(&sdm.metrics.ScopedResolutions, 1)
	case "transient":
		atomic.AddInt64(&sdm.metrics.TransientResolutions, 1)
	}

	// Update average resolution time
	totalCount := atomic.LoadInt64(&sdm.metrics.ResolutionCount)
	if totalCount > 0 {
		totalTime := time.Duration(atomic.LoadInt64((*int64)(&sdm.metrics.TotalResolutionTime)))
		avgTime := totalTime / time.Duration(totalCount)
		atomic.StoreInt64((*int64)(&sdm.metrics.AverageResolutionTime), int64(avgTime))
	}
}

// GetMetrics returns a copy of the current dependency metrics
func (sdm *SharedDependencyManager) GetMetrics() DependencyMetrics {
	return DependencyMetrics{
		SingletonResolutions:  atomic.LoadInt64(&sdm.metrics.SingletonResolutions),
		ScopedResolutions:     atomic.LoadInt64(&sdm.metrics.ScopedResolutions),
		TransientResolutions:  atomic.LoadInt64(&sdm.metrics.TransientResolutions),
		SingletonCacheHits:    atomic.LoadInt64(&sdm.metrics.SingletonCacheHits),
		ScopedCacheHits:       atomic.LoadInt64(&sdm.metrics.ScopedCacheHits),
		AverageResolutionTime: time.Duration(atomic.LoadInt64((*int64)(&sdm.metrics.AverageResolutionTime))),
		TotalResolutionTime:   time.Duration(atomic.LoadInt64((*int64)(&sdm.metrics.TotalResolutionTime))),
		ResolutionCount:       atomic.LoadInt64(&sdm.metrics.ResolutionCount),
		SharedInstanceCount:   atomic.LoadInt64(&sdm.metrics.SharedInstanceCount),
		DuplicatesPrevented:   atomic.LoadInt64(&sdm.metrics.DuplicatesPrevented),
	}
}

// IsCommonFrameworkDependency checks if a type is a common framework dependency that should be shared
func IsCommonFrameworkDependency(interfaceType reflect.Type) bool {
	typeName := interfaceType.String()

	// List of common framework dependency patterns
	commonDependencies := []string{
		"interfaces.ConfigProviderInterface",
		"interfaces.LoggerInterface",
		"interfaces.RepositoryInterface",
		"interfaces.DatabaseClientInterface",
		"interfaces.EndorServiceInterface",
		"interfaces.EndorHybridServiceInterface",
		// Also handle local test interfaces
		"di.ConfigProviderInterface",
		"di.LoggerInterface",
		"di.RepositoryInterface",
	}

	for _, common := range commonDependencies {
		if typeName == common {
			return true
		}
	}

	return false
}

// RecommendScope suggests the appropriate scope for a dependency based on its type
func RecommendScope(interfaceType reflect.Type) Scope {
	if IsCommonFrameworkDependency(interfaceType) {
		typeName := interfaceType.String()

		// Configuration and database clients should be singleton
		if typeName == "interfaces.ConfigProviderInterface" ||
			typeName == "di.ConfigProviderInterface" ||
			typeName == "interfaces.DatabaseClientInterface" ||
			typeName == "interfaces.LoggerInterface" ||
			typeName == "di.LoggerInterface" {
			return Singleton
		}

		// Repository interfaces could be singleton or scoped depending on use case
		if typeName == "interfaces.RepositoryInterface" ||
			typeName == "di.RepositoryInterface" {
			return Singleton // Default to singleton for shared access
		}

		// Service interfaces are typically transient unless explicitly shared
		return Transient
	}

	// Default recommendation for unknown types
	return Transient
}
