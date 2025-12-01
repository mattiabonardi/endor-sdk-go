package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthChecker interface for dependency health monitoring
type HealthChecker interface {
	// HealthCheck performs a health check and returns error if unhealthy
	HealthCheck() error

	// OnHealthChange registers a callback for health status changes
	OnHealthChange(callback func(healthy bool))
}

// HealthStatus represents the health status of a dependency
type HealthStatus int

const (
	Healthy HealthStatus = iota
	Degraded
	Unhealthy
)

func (s HealthStatus) String() string {
	switch s {
	case Healthy:
		return "Healthy"
	case Degraded:
		return "Degraded"
	case Unhealthy:
		return "Unhealthy"
	default:
		return "Unknown"
	}
}

// HealthReport contains detailed health information
type HealthReport struct {
	Status    HealthStatus
	Message   string
	CheckedAt time.Time
	Metadata  map[string]interface{}
}

// HealthAggregator aggregates health status across multiple dependencies
type HealthAggregator interface {
	// AddChecker adds a health checker to be monitored
	AddChecker(name string, checker HealthChecker)

	// GetOverallHealth returns the aggregate health status
	GetOverallHealth() HealthReport

	// GetDetailedHealth returns individual health status for all checkers
	GetDetailedHealth() map[string]HealthReport
}

// DefaultHealthChecker provides a basic health checker implementation
type DefaultHealthChecker struct {
	checkFunc   func() error
	callbacks   []func(bool)
	lastHealthy bool
	mutex       sync.RWMutex
}

// NewDefaultHealthChecker creates a health checker with a custom check function
func NewDefaultHealthChecker(checkFunc func() error) *DefaultHealthChecker {
	return &DefaultHealthChecker{
		checkFunc:   checkFunc,
		callbacks:   make([]func(bool), 0),
		lastHealthy: true,
	}
}

// HealthCheck performs the health check
func (hc *DefaultHealthChecker) HealthCheck() error {
	err := hc.checkFunc()

	hc.mutex.Lock()
	currentlyHealthy := err == nil
	wasHealthy := hc.lastHealthy
	hc.lastHealthy = currentlyHealthy
	hc.mutex.Unlock()

	// Notify listeners if health status changed
	if currentlyHealthy != wasHealthy {
		hc.mutex.RLock()
		for _, callback := range hc.callbacks {
			go callback(currentlyHealthy) // Async to prevent blocking
		}
		hc.mutex.RUnlock()
	}

	return err
}

// OnHealthChange registers a callback for health status changes
func (hc *DefaultHealthChecker) OnHealthChange(callback func(healthy bool)) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.callbacks = append(hc.callbacks, callback)
}

// CircuitBreaker implements a circuit breaker pattern for unhealthy dependencies
type CircuitBreaker struct {
	name                 string
	checker              HealthChecker
	failureThreshold     int
	recoveryTimeout      time.Duration
	consecutiveFailures  int
	lastFailureTime      time.Time
	state                CircuitState
	mutex                sync.RWMutex
	stateChangeCallbacks []func(CircuitState)
}

type CircuitState int

const (
	Closed   CircuitState = iota // Normal operation
	Open                         // Circuit is open, calls fail fast
	HalfOpen                     // Testing if dependency has recovered
)

func (s CircuitState) String() string {
	switch s {
	case Closed:
		return "Closed"
	case Open:
		return "Open"
	case HalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, checker HealthChecker, failureThreshold int, recoveryTimeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:             name,
		checker:          checker,
		failureThreshold: failureThreshold,
		recoveryTimeout:  recoveryTimeout,
		state:            Closed,
	}

	// Register for health change notifications
	checker.OnHealthChange(cb.onHealthChange)

	return cb
}

// Execute attempts to execute an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	if !cb.allowExecution() {
		return &CircuitBreakerOpenError{
			Name:  cb.name,
			State: cb.getState(),
		}
	}

	err := operation()
	cb.recordResult(err)
	return err
}

// allowExecution determines if execution should be allowed
func (cb *CircuitBreaker) allowExecution() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case Closed:
		return true
	case Open:
		// Check if we should try to recover
		if time.Since(cb.lastFailureTime) > cb.recoveryTimeout {
			// Transition to half-open
			cb.setState(HalfOpen)
			return true
		}
		return false
	case HalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.consecutiveFailures++
		cb.lastFailureTime = time.Now()

		if cb.consecutiveFailures >= cb.failureThreshold && cb.state == Closed {
			cb.setState(Open)
		} else if cb.state == HalfOpen {
			cb.setState(Open)
		}
	} else {
		cb.consecutiveFailures = 0
		if cb.state == HalfOpen {
			cb.setState(Closed)
		}
	}
}

// onHealthChange handles health status changes from the underlying dependency
func (cb *CircuitBreaker) onHealthChange(healthy bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if !healthy {
		cb.consecutiveFailures++
		cb.lastFailureTime = time.Now()

		if cb.consecutiveFailures >= cb.failureThreshold {
			cb.setState(Open)
		}
	} else {
		cb.consecutiveFailures = 0
		if cb.state == Open || cb.state == HalfOpen {
			cb.setState(Closed)
		}
	}
}

// setState transitions the circuit breaker to a new state
func (cb *CircuitBreaker) setState(newState CircuitState) {
	cb.state = newState

	// Notify state change listeners
	for _, callback := range cb.stateChangeCallbacks {
		go callback(newState)
	}
}

// getState returns the current circuit breaker state
func (cb *CircuitBreaker) getState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetState returns the current circuit breaker state (public method)
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.getState()
}

// OnStateChange registers a callback for circuit breaker state changes
func (cb *CircuitBreaker) OnStateChange(callback func(CircuitState)) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.stateChangeCallbacks = append(cb.stateChangeCallbacks, callback)
}

// CircuitBreakerOpenError represents an error when circuit breaker is open
type CircuitBreakerOpenError struct {
	Name  string
	State CircuitState
}

func (e *CircuitBreakerOpenError) Error() string {
	return fmt.Sprintf("circuit breaker '%s' is %s", e.Name, e.State.String())
}
