package di

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (c CircuitState) String() string {
	switch c {
	case CircuitClosed:
		return "Closed"
	case CircuitOpen:
		return "Open"
	case CircuitHalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures needed to open the circuit
	FailureThreshold int
	// OpenTimeout is how long the circuit stays open before trying to close
	OpenTimeout time.Duration
	// SuccessThreshold is the number of consecutive successes needed to close the circuit from half-open
	SuccessThreshold int
}

// DefaultCircuitBreakerConfig returns a sensible default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		OpenTimeout:      30 * time.Second,
		SuccessThreshold: 3,
	}
}

// CircuitBreaker provides circuit breaker functionality for dependencies
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitState
	failureCount    int64
	successCount    int64
	lastFailureTime time.Time
	lastSuccessTime time.Time
	mu              sync.RWMutex
}

// CircuitBreakerError represents an error when the circuit is open
type CircuitBreakerError struct {
	DependencyType string
	State          CircuitState
}

func (e *CircuitBreakerError) Error() string {
	return fmt.Sprintf("circuit breaker is %s for dependency %s", e.State.String(), e.DependencyType)
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.canExecute() {
		return &CircuitBreakerError{
			State: cb.GetState(),
		}
	}

	err := fn()
	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}

	return err
}

// canExecute determines if the function can be executed based on circuit state
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if enough time has passed to try half-open
		if time.Since(cb.lastFailureTime) > cb.config.OpenTimeout {
			cb.state = CircuitHalfOpen
			cb.successCount = 0
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// recordFailure records a failed execution
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()
	newFailureCount := atomic.AddInt64(&cb.failureCount, 1)

	switch cb.state {
	case CircuitClosed:
		if newFailureCount >= int64(cb.config.FailureThreshold) {
			cb.state = CircuitOpen
		}
	case CircuitHalfOpen:
		cb.state = CircuitOpen
		atomic.StoreInt64(&cb.failureCount, 1) // Reset for next cycle
	}
}

// recordSuccess records a successful execution
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastSuccessTime = time.Now()

	switch cb.state {
	case CircuitClosed:
		atomic.StoreInt64(&cb.failureCount, 0) // Reset failure count on success
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= int64(cb.config.SuccessThreshold) {
			cb.state = CircuitClosed
			atomic.StoreInt64(&cb.failureCount, 0)
			cb.successCount = 0
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int64 {
	return atomic.LoadInt64(&cb.failureCount)
}

// Reset resets the circuit breaker to its initial state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitClosed
	atomic.StoreInt64(&cb.failureCount, 0)
	cb.successCount = 0
}

// CircuitBreakerManager manages circuit breakers for different dependencies
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	config   CircuitBreakerConfig
	mu       sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		config:   config,
	}
}

// GetCircuitBreaker gets or creates a circuit breaker for a dependency
func (cbm *CircuitBreakerManager) GetCircuitBreaker(dependencyType string) *CircuitBreaker {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[dependencyType]
	cbm.mu.RUnlock()

	if exists {
		return breaker
	}

	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	// Double-check pattern
	if breaker, exists := cbm.breakers[dependencyType]; exists {
		return breaker
	}

	breaker = NewCircuitBreaker(cbm.config)
	cbm.breakers[dependencyType] = breaker
	return breaker
}

// ExecuteWithBreaker executes a function with circuit breaker protection for a specific dependency
func (cbm *CircuitBreakerManager) ExecuteWithBreaker(dependencyType string, fn func() error) error {
	breaker := cbm.GetCircuitBreaker(dependencyType)

	err := breaker.Execute(fn)
	if cbErr, ok := err.(*CircuitBreakerError); ok {
		cbErr.DependencyType = dependencyType
	}

	return err
}

// GetBreakerState returns the current state of a circuit breaker for a dependency
func (cbm *CircuitBreakerManager) GetBreakerState(dependencyType string) CircuitState {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[dependencyType]
	cbm.mu.RUnlock()

	if !exists {
		return CircuitClosed // Default state for non-existent breakers
	}

	return breaker.GetState()
}

// GetOpenBreakers returns a list of dependencies with open circuit breakers
func (cbm *CircuitBreakerManager) GetOpenBreakers() []string {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	var openBreakers []string
	for dependencyType, breaker := range cbm.breakers {
		if breaker.GetState() == CircuitOpen {
			openBreakers = append(openBreakers, dependencyType)
		}
	}

	return openBreakers
}

// Reset resets all circuit breakers
func (cbm *CircuitBreakerManager) Reset() {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	for _, breaker := range cbm.breakers {
		breaker.Reset()
	}
}

// HealthAwareCircuitBreakerListener implements HealthListener to integrate circuit breakers with health monitoring
type HealthAwareCircuitBreakerListener struct {
	circuitBreakerManager *CircuitBreakerManager
}

// NewHealthAwareCircuitBreakerListener creates a new health-aware circuit breaker listener
func NewHealthAwareCircuitBreakerListener(cbm *CircuitBreakerManager) *HealthAwareCircuitBreakerListener {
	return &HealthAwareCircuitBreakerListener{
		circuitBreakerManager: cbm,
	}
}

// OnHealthChanged implements HealthListener interface
func (h *HealthAwareCircuitBreakerListener) OnHealthChanged(dependencyType string, oldStatus, newStatus HealthStatus, healthCheck HealthCheck) {
	breaker := h.circuitBreakerManager.GetCircuitBreaker(dependencyType)

	switch newStatus {
	case HealthUnhealthy:
		// Record failure when dependency becomes unhealthy
		breaker.recordFailure()
	case HealthHealthy:
		// Record success when dependency becomes healthy
		if oldStatus == HealthUnhealthy || oldStatus == HealthDegraded {
			breaker.recordSuccess()
		}
	case HealthDegraded:
		// For degraded state, we don't change circuit breaker state
		// This allows for performance degradation without complete circuit opening
	}
}

// DependencyWithCircuitBreaker wraps a dependency with circuit breaker functionality
type DependencyWithCircuitBreaker struct {
	dependency     interface{}
	dependencyType string
	breaker        *CircuitBreaker
}

// NewDependencyWithCircuitBreaker creates a new dependency wrapper with circuit breaker
func NewDependencyWithCircuitBreaker(dependency interface{}, dependencyType string, breaker *CircuitBreaker) *DependencyWithCircuitBreaker {
	return &DependencyWithCircuitBreaker{
		dependency:     dependency,
		dependencyType: dependencyType,
		breaker:        breaker,
	}
}

// Execute executes an operation on the dependency with circuit breaker protection
func (d *DependencyWithCircuitBreaker) Execute(operation func(interface{}) error) error {
	return d.breaker.Execute(func() error {
		return operation(d.dependency)
	})
}

// GetDependency returns the underlying dependency if the circuit is not open
func (d *DependencyWithCircuitBreaker) GetDependency() (interface{}, error) {
	if d.breaker.GetState() == CircuitOpen {
		return nil, &CircuitBreakerError{
			DependencyType: d.dependencyType,
			State:          CircuitOpen,
		}
	}
	return d.dependency, nil
}

// IsAvailable returns true if the dependency is available (circuit not open)
func (d *DependencyWithCircuitBreaker) IsAvailable() bool {
	return d.breaker.GetState() != CircuitOpen
}

// GetState returns the current circuit breaker state
func (d *DependencyWithCircuitBreaker) GetState() CircuitState {
	return d.breaker.GetState()
}
