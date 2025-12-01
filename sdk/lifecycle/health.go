package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthAggregationPolicy defines how to aggregate health status from multiple services
type HealthAggregationPolicy int

const (
	// AllHealthyPolicy requires all services to be healthy for overall healthy status
	AllHealthyPolicy HealthAggregationPolicy = iota
	// MajorityHealthyPolicy requires majority of services to be healthy
	MajorityHealthyPolicy
	// CriticalServicesHealthyPolicy requires only critical services to be healthy
	CriticalServicesHealthyPolicy
)

// String returns the string representation of the health aggregation policy
func (h HealthAggregationPolicy) String() string {
	switch h {
	case AllHealthyPolicy:
		return "AllHealthy"
	case MajorityHealthyPolicy:
		return "MajorityHealthy"
	case CriticalServicesHealthyPolicy:
		return "CriticalServicesHealthy"
	default:
		return "Unknown"
	}
}

// ServiceHealthConfiguration contains configuration for service health checking
type ServiceHealthConfiguration struct {
	// CheckInterval is the frequency of health checks
	CheckInterval time.Duration
	// CheckTimeout is the maximum time to wait for a health check
	CheckTimeout time.Duration
	// CacheTimeout is how long to cache health check results
	CacheTimeout time.Duration
	// IsCritical indicates if this service is critical for overall system health
	IsCritical bool
	// RetryAttempts is the number of retry attempts for failed health checks
	RetryAttempts int
	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration
}

// DefaultHealthConfiguration returns the default health configuration
func DefaultHealthConfiguration() ServiceHealthConfiguration {
	return ServiceHealthConfiguration{
		CheckInterval: 30 * time.Second,
		CheckTimeout:  10 * time.Second,
		CacheTimeout:  10 * time.Second,
		IsCritical:    false,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}
}

// CompositeHealthStatus represents the aggregated health status of multiple services
type CompositeHealthStatus struct {
	// OverallStatus is the aggregated health status
	OverallStatus ServiceHealthStatus `json:"overallStatus"`
	// Services contains the health status of individual services
	Services map[string]HealthStatus `json:"services"`
	// Policy is the aggregation policy used
	Policy HealthAggregationPolicy `json:"policy"`
	// LastCheck is when the composite health was last calculated
	LastCheck time.Time `json:"lastCheck"`
	// HealthySources is the number of healthy services
	HealthySources int `json:"healthySources"`
	// TotalSources is the total number of services
	TotalSources int `json:"totalSources"`
}

// CachedHealthStatus contains a cached health check result
type CachedHealthStatus struct {
	Status    HealthStatus
	ExpiresAt time.Time
}

// IsExpired returns true if the cached status has expired
func (c *CachedHealthStatus) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// HealthMonitor manages health checking for services with caching and aggregation
type HealthMonitor struct {
	services   map[string]ServiceLifecycleInterface
	configs    map[string]ServiceHealthConfiguration
	cache      map[string]*CachedHealthStatus
	policy     HealthAggregationPolicy
	mu         sync.RWMutex
	stopCh     chan struct{}
	monitoring bool
}

// NewHealthMonitor creates a new health monitor with the given aggregation policy
func NewHealthMonitor(policy HealthAggregationPolicy) *HealthMonitor {
	return &HealthMonitor{
		services: make(map[string]ServiceLifecycleInterface),
		configs:  make(map[string]ServiceHealthConfiguration),
		cache:    make(map[string]*CachedHealthStatus),
		policy:   policy,
		stopCh:   make(chan struct{}),
	}
}

// RegisterService registers a service for health monitoring
func (hm *HealthMonitor) RegisterService(name string, service ServiceLifecycleInterface, config ServiceHealthConfiguration) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.services[name] = service
	hm.configs[name] = config
}

// UnregisterService removes a service from health monitoring
func (hm *HealthMonitor) UnregisterService(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	delete(hm.services, name)
	delete(hm.configs, name)
	delete(hm.cache, name)
}

// CheckHealth checks the health of a specific service with caching
func (hm *HealthMonitor) CheckHealth(ctx context.Context, serviceName string) (HealthStatus, error) {
	hm.mu.RLock()
	service, exists := hm.services[serviceName]
	config, configExists := hm.configs[serviceName]
	cached, hasCached := hm.cache[serviceName]
	hm.mu.RUnlock()

	if !exists {
		return HealthStatus{}, fmt.Errorf("service %s not registered", serviceName)
	}

	if !configExists {
		config = DefaultHealthConfiguration()
	}

	// Check cache first
	if hasCached && !cached.IsExpired() {
		return cached.Status, nil
	}

	// Perform health check with timeout
	checkCtx, cancel := context.WithTimeout(ctx, config.CheckTimeout)
	defer cancel()

	var status HealthStatus
	var err error

	// Retry logic for health checks
	for attempt := 0; attempt <= config.RetryAttempts; attempt++ {
		status = service.HealthCheck(checkCtx)

		// If health check succeeded or this is the last attempt, break
		if status.Status == Healthy || status.Status == Degraded || attempt == config.RetryAttempts {
			break
		}

		// Wait before retry
		if attempt < config.RetryAttempts {
			select {
			case <-time.After(config.RetryDelay):
			case <-ctx.Done():
				return HealthStatus{Status: Unknown}, ctx.Err()
			}
		}
	}

	// Update cache
	hm.mu.Lock()
	hm.cache[serviceName] = &CachedHealthStatus{
		Status:    status,
		ExpiresAt: time.Now().Add(config.CacheTimeout),
	}
	hm.mu.Unlock()

	return status, err
}

// CheckCompositeHealth checks the health of all registered services and returns composite status
func (hm *HealthMonitor) CheckCompositeHealth(ctx context.Context) CompositeHealthStatus {
	hm.mu.RLock()
	serviceNames := make([]string, 0, len(hm.services))
	for name := range hm.services {
		serviceNames = append(serviceNames, name)
	}
	policy := hm.policy
	hm.mu.RUnlock()

	composite := CompositeHealthStatus{
		Services:     make(map[string]HealthStatus),
		Policy:       policy,
		LastCheck:    time.Now(),
		TotalSources: len(serviceNames),
	}

	// Check health of each service
	healthyCount := 0
	criticalHealthyCount := 0
	totalCritical := 0

	for _, serviceName := range serviceNames {
		status, err := hm.CheckHealth(ctx, serviceName)
		if err != nil {
			status = HealthStatus{
				Status:    Unknown,
				Details:   map[string]interface{}{"error": err.Error()},
				LastCheck: time.Now(),
			}
		}

		composite.Services[serviceName] = status

		// Count healthy services
		if status.Status == Healthy || status.Status == Degraded {
			healthyCount++

			// Check if this is a critical service
			hm.mu.RLock()
			config, exists := hm.configs[serviceName]
			hm.mu.RUnlock()

			if exists && config.IsCritical {
				criticalHealthyCount++
			}
		}

		// Count total critical services
		hm.mu.RLock()
		config, exists := hm.configs[serviceName]
		hm.mu.RUnlock()

		if exists && config.IsCritical {
			totalCritical++
		}
	}

	composite.HealthySources = healthyCount

	// Determine overall health based on policy
	switch policy {
	case AllHealthyPolicy:
		if healthyCount == len(serviceNames) {
			composite.OverallStatus = Healthy
		} else if healthyCount > 0 {
			composite.OverallStatus = Degraded
		} else {
			composite.OverallStatus = Unhealthy
		}

	case MajorityHealthyPolicy:
		if healthyCount > len(serviceNames)/2 {
			composite.OverallStatus = Healthy
		} else if healthyCount > 0 {
			composite.OverallStatus = Degraded
		} else {
			composite.OverallStatus = Unhealthy
		}

	case CriticalServicesHealthyPolicy:
		if totalCritical == 0 {
			// No critical services, use all services
			if healthyCount == len(serviceNames) {
				composite.OverallStatus = Healthy
			} else if healthyCount > 0 {
				composite.OverallStatus = Degraded
			} else {
				composite.OverallStatus = Unhealthy
			}
		} else {
			// Critical services exist, check them
			if criticalHealthyCount == totalCritical {
				composite.OverallStatus = Healthy
			} else if criticalHealthyCount > 0 {
				composite.OverallStatus = Degraded
			} else {
				composite.OverallStatus = Unhealthy
			}
		}

	default:
		composite.OverallStatus = Unknown
	}

	return composite
}

// StartMonitoring starts periodic health monitoring for all registered services
func (hm *HealthMonitor) StartMonitoring(ctx context.Context) error {
	hm.mu.Lock()
	if hm.monitoring {
		hm.mu.Unlock()
		return fmt.Errorf("health monitoring already started")
	}
	hm.monitoring = true
	hm.mu.Unlock()

	go hm.monitoringLoop(ctx)
	return nil
}

// StopMonitoring stops the periodic health monitoring
func (hm *HealthMonitor) StopMonitoring() error {
	hm.mu.Lock()
	if !hm.monitoring {
		hm.mu.Unlock()
		return fmt.Errorf("health monitoring not running")
	}
	hm.monitoring = false
	hm.mu.Unlock()

	close(hm.stopCh)
	return nil
}

// monitoringLoop runs the periodic health monitoring
func (hm *HealthMonitor) monitoringLoop(ctx context.Context) {
	// Find the shortest check interval among all services
	hm.mu.RLock()
	minInterval := 60 * time.Second // Default
	for _, config := range hm.configs {
		if config.CheckInterval < minInterval {
			minInterval = config.CheckInterval
		}
	}
	hm.mu.RUnlock()

	ticker := time.NewTicker(minInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if any services need health checks
			hm.performScheduledChecks(ctx)

		case <-hm.stopCh:
			return

		case <-ctx.Done():
			return
		}
	}
}

// performScheduledChecks checks health for services that are due for checking
func (hm *HealthMonitor) performScheduledChecks(ctx context.Context) {
	hm.mu.RLock()
	servicesToCheck := make([]string, 0)

	for serviceName, config := range hm.configs {
		cached, hasCached := hm.cache[serviceName]

		// Check if health check is due
		if !hasCached {
			servicesToCheck = append(servicesToCheck, serviceName)
		} else {
			// Check if enough time has passed since last check
			timeSinceCheck := time.Since(cached.Status.LastCheck)
			if timeSinceCheck >= config.CheckInterval {
				servicesToCheck = append(servicesToCheck, serviceName)
			}
		}
	}
	hm.mu.RUnlock()

	// Perform health checks for due services
	for _, serviceName := range servicesToCheck {
		// Perform health check in background to avoid blocking
		go func(name string) {
			_, _ = hm.CheckHealth(ctx, name)
		}(serviceName)
	}
}

// GetCacheStatus returns the current cache status for debugging
func (hm *HealthMonitor) GetCacheStatus() map[string]bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	status := make(map[string]bool)
	for serviceName, cached := range hm.cache {
		status[serviceName] = !cached.IsExpired()
	}
	return status
}
