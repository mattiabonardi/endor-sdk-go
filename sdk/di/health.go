package di

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthStatus represents the health state of a dependency
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthHealthy
	HealthDegraded
	HealthUnhealthy
)

func (h HealthStatus) String() string {
	switch h {
	case HealthHealthy:
		return "Healthy"
	case HealthDegraded:
		return "Degraded"
	case HealthUnhealthy:
		return "Unhealthy"
	case HealthUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// HealthCheck represents a health check result
type HealthCheck struct {
	Status        HealthStatus           `json:"status"`
	Message       string                 `json:"message,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CheckDuration time.Duration          `json:"check_duration"`
}

// HealthChecker interface for dependencies that can be health checked
type HealthChecker interface {
	// HealthCheck performs a health check and returns the result
	HealthCheck(ctx context.Context) HealthCheck
}

// DependencyHealth tracks health status for a specific dependency
type DependencyHealth struct {
	DependencyType   string
	Instance         interface{}
	LastCheck        HealthCheck
	ConsecutiveFails int
	CheckInterval    time.Duration
	MaxFailures      int
	mu               sync.RWMutex
}

// HealthMonitor manages health checking for all dependencies
type HealthMonitor struct {
	dependencies map[string]*DependencyHealth
	listeners    []HealthListener
	ticker       *time.Ticker
	stopChan     chan struct{}
	mu           sync.RWMutex
}

// HealthListener interface for notifications about dependency health changes
type HealthListener interface {
	OnHealthChanged(dependencyType string, oldStatus, newStatus HealthStatus, health HealthCheck)
}

// HealthEvent represents a health status change event
type HealthEvent struct {
	DependencyType string
	OldStatus      HealthStatus
	NewStatus      HealthStatus
	HealthCheck    HealthCheck
	Timestamp      time.Time
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		dependencies: make(map[string]*DependencyHealth),
		listeners:    make([]HealthListener, 0),
		stopChan:     make(chan struct{}),
	}
}

// RegisterDependency registers a dependency for health monitoring
func (h *HealthMonitor) RegisterDependency(dependencyType string, instance interface{}, options ...HealthOption) {
	h.mu.Lock()
	defer h.mu.Unlock()

	health := &DependencyHealth{
		DependencyType: dependencyType,
		Instance:       instance,
		LastCheck: HealthCheck{
			Status:    HealthUnknown,
			Timestamp: time.Now(),
		},
		CheckInterval: 30 * time.Second, // Default check interval
		MaxFailures:   3,                // Default max consecutive failures
	}

	// Apply options
	for _, option := range options {
		option(health)
	}

	h.dependencies[dependencyType] = health
}

// GetHealth returns the current health status for a dependency
func (h *HealthMonitor) GetHealth(dependencyType string) (HealthCheck, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if dep, exists := h.dependencies[dependencyType]; exists {
		dep.mu.RLock()
		defer dep.mu.RUnlock()
		return dep.LastCheck, true
	}
	return HealthCheck{}, false
}

// GetOverallHealth returns aggregated health status of all dependencies
func (h *HealthMonitor) GetOverallHealth() HealthCheck {
	h.mu.RLock()
	defer h.mu.RUnlock()

	overallStatus := HealthHealthy
	messages := make([]string, 0)
	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0
	unknownCount := 0

	for _, dep := range h.dependencies {
		dep.mu.RLock()
		status := dep.LastCheck.Status
		dep.mu.RUnlock()

		switch status {
		case HealthHealthy:
			healthyCount++
		case HealthDegraded:
			degradedCount++
			if overallStatus == HealthHealthy {
				overallStatus = HealthDegraded
			}
			messages = append(messages, fmt.Sprintf("%s is degraded", dep.DependencyType))
		case HealthUnhealthy:
			unhealthyCount++
			overallStatus = HealthUnhealthy
			messages = append(messages, fmt.Sprintf("%s is unhealthy", dep.DependencyType))
		case HealthUnknown:
			unknownCount++
			if overallStatus == HealthHealthy {
				overallStatus = HealthDegraded
			}
			messages = append(messages, fmt.Sprintf("%s status is unknown", dep.DependencyType))
		}
	}

	message := ""
	if len(messages) > 0 {
		message = fmt.Sprintf("Issues: %v", messages)
	} else {
		message = fmt.Sprintf("All %d dependencies healthy", healthyCount)
	}

	return HealthCheck{
		Status:    overallStatus,
		Message:   message,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"healthy_count":   healthyCount,
			"degraded_count":  degradedCount,
			"unhealthy_count": unhealthyCount,
			"unknown_count":   unknownCount,
			"total_count":     len(h.dependencies),
		},
	}
}

// AddListener adds a health status change listener
func (h *HealthMonitor) AddListener(listener HealthListener) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners = append(h.listeners, listener)
}

// CheckHealth performs immediate health check for a specific dependency
func (h *HealthMonitor) CheckHealth(ctx context.Context, dependencyType string) error {
	h.mu.RLock()
	dep, exists := h.dependencies[dependencyType]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("dependency %s not registered", dependencyType)
	}

	h.performHealthCheck(ctx, dep)
	return nil
}

// CheckAllHealth performs health checks for all registered dependencies
func (h *HealthMonitor) CheckAllHealth(ctx context.Context) {
	h.mu.RLock()
	deps := make([]*DependencyHealth, 0, len(h.dependencies))
	for _, dep := range h.dependencies {
		deps = append(deps, dep)
	}
	h.mu.RUnlock()

	for _, dep := range deps {
		h.performHealthCheck(ctx, dep)
	}
}

// performHealthCheck executes health check for a dependency and notifies listeners
func (h *HealthMonitor) performHealthCheck(ctx context.Context, dep *DependencyHealth) {
	checker, ok := dep.Instance.(HealthChecker)
	if !ok {
		// If dependency doesn't implement HealthChecker, assume it's healthy
		dep.mu.Lock()
		oldStatus := dep.LastCheck.Status
		dep.LastCheck = HealthCheck{
			Status:        HealthHealthy,
			Message:       "No health check implemented - assumed healthy",
			Timestamp:     time.Now(),
			CheckDuration: 0,
		}
		dep.ConsecutiveFails = 0
		newStatus := dep.LastCheck.Status
		dep.mu.Unlock()

		if oldStatus != newStatus {
			h.notifyListeners(dep.DependencyType, oldStatus, newStatus, dep.LastCheck)
		}
		return
	}

	start := time.Now()

	// Create context with timeout for health check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	healthResult := checker.HealthCheck(checkCtx)
	healthResult.CheckDuration = time.Since(start)
	healthResult.Timestamp = time.Now()

	dep.mu.Lock()
	oldStatus := dep.LastCheck.Status
	dep.LastCheck = healthResult

	// Update consecutive failure count
	if healthResult.Status == HealthUnhealthy {
		dep.ConsecutiveFails++
	} else {
		dep.ConsecutiveFails = 0
	}

	newStatus := healthResult.Status
	dep.mu.Unlock()

	// Notify listeners if status changed
	if oldStatus != newStatus {
		h.notifyListeners(dep.DependencyType, oldStatus, newStatus, healthResult)
	}
}

// notifyListeners notifies all registered listeners about health status changes
func (h *HealthMonitor) notifyListeners(dependencyType string, oldStatus, newStatus HealthStatus, healthCheck HealthCheck) {
	h.mu.RLock()
	listeners := make([]HealthListener, len(h.listeners))
	copy(listeners, h.listeners)
	h.mu.RUnlock()

	for _, listener := range listeners {
		// Run listener in goroutine to avoid blocking health checks
		go func(l HealthListener) {
			defer func() {
				if r := recover(); r != nil {
					// Log panic but don't let it crash the health monitor
					fmt.Printf("Health listener panicked: %v\n", r)
				}
			}()
			l.OnHealthChanged(dependencyType, oldStatus, newStatus, healthCheck)
		}(listener)
	}
}

// StartMonitoring starts continuous health monitoring
func (h *HealthMonitor) StartMonitoring(ctx context.Context) {
	h.mu.Lock()
	if h.ticker != nil {
		h.mu.Unlock()
		return // Already running
	}

	// Use the shortest check interval among all dependencies
	minInterval := 30 * time.Second
	for _, dep := range h.dependencies {
		if dep.CheckInterval < minInterval {
			minInterval = dep.CheckInterval
		}
	}

	h.ticker = time.NewTicker(minInterval)
	h.mu.Unlock()

	go func() {
		for {
			select {
			case <-h.ticker.C:
				h.CheckAllHealth(ctx)
			case <-h.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// StopMonitoring stops continuous health monitoring
func (h *HealthMonitor) StopMonitoring() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.ticker != nil {
		h.ticker.Stop()
		h.ticker = nil
		close(h.stopChan)
		h.stopChan = make(chan struct{}) // Reset for future use
	}
}

// IsHealthy returns true if a dependency is currently healthy
func (h *HealthMonitor) IsHealthy(dependencyType string) bool {
	if health, exists := h.GetHealth(dependencyType); exists {
		return health.Status == HealthHealthy
	}
	return false // Unknown dependencies are considered unhealthy
}

// GetUnhealthyDependencies returns a list of dependencies that are currently unhealthy
func (h *HealthMonitor) GetUnhealthyDependencies() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var unhealthy []string
	for _, dep := range h.dependencies {
		dep.mu.RLock()
		if dep.LastCheck.Status == HealthUnhealthy {
			unhealthy = append(unhealthy, dep.DependencyType)
		}
		dep.mu.RUnlock()
	}

	return unhealthy
}

// HealthOption configures health monitoring for a dependency
type HealthOption func(*DependencyHealth)

// WithCheckInterval sets the health check interval for a dependency
func WithCheckInterval(interval time.Duration) HealthOption {
	return func(h *DependencyHealth) {
		h.CheckInterval = interval
	}
}

// WithMaxFailures sets the maximum consecutive failures before marking as unhealthy
func WithMaxFailures(maxFailures int) HealthOption {
	return func(h *DependencyHealth) {
		h.MaxFailures = maxFailures
	}
}
