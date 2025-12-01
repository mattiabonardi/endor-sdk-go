package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RecoveryStrategy defines the strategy for recovering from service failures
type RecoveryStrategy int

const (
	// NoRecovery does not attempt to recover from failures
	NoRecovery RecoveryStrategy = iota
	// ImmediateRestart attempts immediate restart without delay
	ImmediateRestart
	// ExponentialBackoff uses exponential backoff for restart attempts
	ExponentialBackoff
	// LinearBackoff uses linear backoff for restart attempts
	LinearBackoff
)

// String returns the string representation of the recovery strategy
func (r RecoveryStrategy) String() string {
	switch r {
	case NoRecovery:
		return "NoRecovery"
	case ImmediateRestart:
		return "ImmediateRestart"
	case ExponentialBackoff:
		return "ExponentialBackoff"
	case LinearBackoff:
		return "LinearBackoff"
	default:
		return "Unknown"
	}
}

// RecoveryConfiguration contains configuration for service recovery
type RecoveryConfiguration struct {
	// Strategy defines the recovery strategy to use
	Strategy RecoveryStrategy
	// MaxAttempts is the maximum number of recovery attempts
	MaxAttempts int
	// InitialDelay is the initial delay before first recovery attempt
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between recovery attempts
	MaxDelay time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff
	BackoffMultiplier float64
	// LinearIncrement is the increment for linear backoff
	LinearIncrement time.Duration
	// RecoveryTimeout is the maximum time to wait for recovery
	RecoveryTimeout time.Duration
	// HealthCheckInterval is how often to check if recovery is needed
	HealthCheckInterval time.Duration
}

// DefaultRecoveryConfiguration returns the default recovery configuration
func DefaultRecoveryConfiguration() RecoveryConfiguration {
	return RecoveryConfiguration{
		Strategy:            ExponentialBackoff,
		MaxAttempts:         3,
		InitialDelay:        1 * time.Second,
		MaxDelay:            60 * time.Second,
		BackoffMultiplier:   2.0,
		LinearIncrement:     5 * time.Second,
		RecoveryTimeout:     300 * time.Second, // 5 minutes
		HealthCheckInterval: 10 * time.Second,
	}
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// CircuitClosed allows requests to pass through
	CircuitClosed CircuitBreakerState = iota
	// CircuitOpen blocks requests and triggers recovery
	CircuitOpen
	// CircuitHalfOpen allows limited requests to test recovery
	CircuitHalfOpen
)

// String returns the string representation of the circuit breaker state
func (c CircuitBreakerState) String() string {
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

// CircuitBreakerConfiguration contains configuration for circuit breaker
type CircuitBreakerConfiguration struct {
	// FailureThreshold is the number of failures that trigger the circuit to open
	FailureThreshold int
	// SuccessThreshold is the number of successes needed to close the circuit
	SuccessThreshold int
	// Timeout is how long to wait before transitioning from Open to HalfOpen
	Timeout time.Duration
	// MaxConcurrentRequests is the maximum number of requests allowed in HalfOpen state
	MaxConcurrentRequests int
}

// DefaultCircuitBreakerConfiguration returns the default circuit breaker configuration
func DefaultCircuitBreakerConfiguration() CircuitBreakerConfiguration {
	return CircuitBreakerConfiguration{
		FailureThreshold:      5,
		SuccessThreshold:      3,
		Timeout:               30 * time.Second,
		MaxConcurrentRequests: 2,
	}
}

// CircuitBreaker implements the circuit breaker pattern for service failures
type CircuitBreaker struct {
	config          CircuitBreakerConfiguration
	state           CircuitBreakerState
	failures        int
	successes       int
	lastFailureTime time.Time
	requestCount    int
	mu              sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfiguration) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitClosed,
	}
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
			cb.state = CircuitHalfOpen
			cb.requestCount = 0
			cb.successes = 0
			return cb.requestCount < cb.config.MaxConcurrentRequests
		}
		return false

	case CircuitHalfOpen:
		return cb.requestCount < cb.config.MaxConcurrentRequests

	default:
		return false
	}
}

// OnSuccess records a successful operation
func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		cb.failures = 0 // Reset failure count on success

	case CircuitHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			// Transition to closed state
			cb.state = CircuitClosed
			cb.failures = 0
			cb.successes = 0
			cb.requestCount = 0
		}
	}
}

// OnFailure records a failed operation
func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitClosed:
		if cb.failures >= cb.config.FailureThreshold {
			cb.state = CircuitOpen
		}

	case CircuitHalfOpen:
		// Transition back to open state
		cb.state = CircuitOpen
		cb.successes = 0
		cb.requestCount = 0
	}
}

// OnRequest increments the request count (for HalfOpen state)
func (cb *CircuitBreaker) OnRequest() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitHalfOpen {
		cb.requestCount++
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns the current statistics
func (cb *CircuitBreaker) GetStats() (int, int, time.Time) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures, cb.successes, cb.lastFailureTime
}

// RecoveryManager handles service recovery and circuit breaking
type RecoveryManager struct {
	services           map[string]ServiceLifecycleInterface
	configs            map[string]RecoveryConfiguration
	circuitBreakers    map[string]*CircuitBreaker
	recoveryInProgress map[string]bool
	stopChannels       map[string]chan struct{}
	mu                 sync.RWMutex
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		services:           make(map[string]ServiceLifecycleInterface),
		configs:            make(map[string]RecoveryConfiguration),
		circuitBreakers:    make(map[string]*CircuitBreaker),
		recoveryInProgress: make(map[string]bool),
		stopChannels:       make(map[string]chan struct{}),
	}
}

// RegisterService registers a service for recovery management
func (rm *RecoveryManager) RegisterService(name string, service ServiceLifecycleInterface, config RecoveryConfiguration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.services[name] = service
	rm.configs[name] = config
	rm.circuitBreakers[name] = NewCircuitBreaker(DefaultCircuitBreakerConfiguration())
	rm.recoveryInProgress[name] = false
	rm.stopChannels[name] = make(chan struct{})
}

// UnregisterService removes a service from recovery management
func (rm *RecoveryManager) UnregisterService(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Stop any ongoing recovery
	if stopCh, exists := rm.stopChannels[name]; exists {
		close(stopCh)
	}

	delete(rm.services, name)
	delete(rm.configs, name)
	delete(rm.circuitBreakers, name)
	delete(rm.recoveryInProgress, name)
	delete(rm.stopChannels, name)
}

// StartRecoveryMonitoring starts monitoring and recovery for all registered services
func (rm *RecoveryManager) StartRecoveryMonitoring(ctx context.Context) {
	rm.mu.RLock()
	services := make([]string, 0, len(rm.services))
	for serviceName := range rm.services {
		services = append(services, serviceName)
	}
	rm.mu.RUnlock()

	for _, serviceName := range services {
		go rm.monitorService(ctx, serviceName)
	}
}

// StopRecoveryMonitoring stops monitoring for all services
func (rm *RecoveryManager) StopRecoveryMonitoring() {
	rm.mu.RLock()
	stopChannels := make([]chan struct{}, 0, len(rm.stopChannels))
	for _, stopCh := range rm.stopChannels {
		stopChannels = append(stopChannels, stopCh)
	}
	rm.mu.RUnlock()

	for _, stopCh := range stopChannels {
		select {
		case <-stopCh:
			// Already closed
		default:
			close(stopCh)
		}
	}
}

// monitorService monitors a service and triggers recovery when needed
func (rm *RecoveryManager) monitorService(ctx context.Context, serviceName string) {
	rm.mu.RLock()
	config, exists := rm.configs[serviceName]
	stopCh, stopExists := rm.stopChannels[serviceName]
	rm.mu.RUnlock()

	if !exists || !stopExists {
		return
	}

	ticker := time.NewTicker(config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if rm.needsRecovery(ctx, serviceName) {
				go rm.recoverService(ctx, serviceName)
			}

		case <-stopCh:
			return

		case <-ctx.Done():
			return
		}
	}
}

// needsRecovery checks if a service needs recovery
func (rm *RecoveryManager) needsRecovery(ctx context.Context, serviceName string) bool {
	rm.mu.RLock()
	service, exists := rm.services[serviceName]
	inProgress := rm.recoveryInProgress[serviceName]
	circuitBreaker := rm.circuitBreakers[serviceName]
	rm.mu.RUnlock()

	if !exists || inProgress {
		return false
	}

	// Check service state
	state := service.GetState()
	if state == Failed || state == Stopped {
		return true
	}

	// Check health status
	health := service.HealthCheck(ctx)
	if health.Status == Unhealthy {
		circuitBreaker.OnFailure()
		return circuitBreaker.GetState() == CircuitOpen
	}

	circuitBreaker.OnSuccess()
	return false
}

// recoverService attempts to recover a failed service
func (rm *RecoveryManager) recoverService(ctx context.Context, serviceName string) {
	rm.mu.Lock()
	service, exists := rm.services[serviceName]
	config, configExists := rm.configs[serviceName]

	if !exists || !configExists || rm.recoveryInProgress[serviceName] {
		rm.mu.Unlock()
		return
	}

	rm.recoveryInProgress[serviceName] = true
	rm.mu.Unlock()

	defer func() {
		rm.mu.Lock()
		rm.recoveryInProgress[serviceName] = false
		rm.mu.Unlock()
	}()

	fmt.Printf("Starting recovery for service '%s' using strategy %s\n", serviceName, config.Strategy.String())

	// Create recovery context with timeout
	recoveryCtx, cancel := context.WithTimeout(ctx, config.RecoveryTimeout)
	defer cancel()

	// Attempt recovery based on strategy
	var recovered bool
	switch config.Strategy {
	case NoRecovery:
		fmt.Printf("No recovery strategy configured for service '%s'\n", serviceName)
		return

	case ImmediateRestart:
		recovered = rm.attemptImmediateRestart(recoveryCtx, serviceName, service)

	case ExponentialBackoff:
		recovered = rm.attemptExponentialBackoffRestart(recoveryCtx, serviceName, service, config)

	case LinearBackoff:
		recovered = rm.attemptLinearBackoffRestart(recoveryCtx, serviceName, service, config)

	default:
		fmt.Printf("Unknown recovery strategy %s for service '%s'\n", config.Strategy.String(), serviceName)
		return
	}

	if recovered {
		fmt.Printf("Successfully recovered service '%s'\n", serviceName)
		rm.mu.RLock()
		circuitBreaker := rm.circuitBreakers[serviceName]
		rm.mu.RUnlock()
		circuitBreaker.OnSuccess()
	} else {
		fmt.Printf("Failed to recover service '%s' after all attempts\n", serviceName)
	}
}

// attemptImmediateRestart attempts immediate restart without delay
func (rm *RecoveryManager) attemptImmediateRestart(ctx context.Context, serviceName string, service ServiceLifecycleInterface) bool {
	// Stop the service first
	if err := service.Stop(ctx); err != nil {
		fmt.Printf("Failed to stop service '%s' during recovery: %v\n", serviceName, err)
	}

	// Start the service
	if err := service.Start(ctx); err != nil {
		fmt.Printf("Failed to restart service '%s': %v\n", serviceName, err)
		return false
	}

	return service.GetState() == Running
}

// attemptExponentialBackoffRestart attempts restart with exponential backoff
func (rm *RecoveryManager) attemptExponentialBackoffRestart(ctx context.Context, serviceName string, service ServiceLifecycleInterface, config RecoveryConfiguration) bool {
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		fmt.Printf("Recovery attempt %d/%d for service '%s' (delay: %v)\n", attempt, config.MaxAttempts, serviceName, delay)

		// Wait for the calculated delay
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return false
		}

		// Stop the service first
		if err := service.Stop(ctx); err != nil {
			fmt.Printf("Failed to stop service '%s' during recovery attempt %d: %v\n", serviceName, attempt, err)
		}

		// Start the service
		if err := service.Start(ctx); err != nil {
			fmt.Printf("Failed to restart service '%s' on attempt %d: %v\n", serviceName, attempt, err)
		} else if service.GetState() == Running {
			// Verify the service is actually healthy
			health := service.HealthCheck(ctx)
			if health.Status == Healthy || health.Status == Degraded {
				return true
			}
		}

		// Calculate next delay with exponential backoff
		nextDelay := time.Duration(float64(delay) * config.BackoffMultiplier)
		if nextDelay > config.MaxDelay {
			nextDelay = config.MaxDelay
		}
		delay = nextDelay
	}

	return false
}

// attemptLinearBackoffRestart attempts restart with linear backoff
func (rm *RecoveryManager) attemptLinearBackoffRestart(ctx context.Context, serviceName string, service ServiceLifecycleInterface, config RecoveryConfiguration) bool {
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		fmt.Printf("Recovery attempt %d/%d for service '%s' (delay: %v)\n", attempt, config.MaxAttempts, serviceName, delay)

		// Wait for the calculated delay
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return false
		}

		// Stop the service first
		if err := service.Stop(ctx); err != nil {
			fmt.Printf("Failed to stop service '%s' during recovery attempt %d: %v\n", serviceName, attempt, err)
		}

		// Start the service
		if err := service.Start(ctx); err != nil {
			fmt.Printf("Failed to restart service '%s' on attempt %d: %v\n", serviceName, attempt, err)
		} else if service.GetState() == Running {
			// Verify the service is actually healthy
			health := service.HealthCheck(ctx)
			if health.Status == Healthy || health.Status == Degraded {
				return true
			}
		}

		// Calculate next delay with linear backoff
		nextDelay := delay + config.LinearIncrement
		if nextDelay > config.MaxDelay {
			nextDelay = config.MaxDelay
		}
		delay = nextDelay
	}

	return false
}

// GetCircuitBreakerState returns the current state of a service's circuit breaker
func (rm *RecoveryManager) GetCircuitBreakerState(serviceName string) (CircuitBreakerState, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	circuitBreaker, exists := rm.circuitBreakers[serviceName]
	if !exists {
		return CircuitClosed, fmt.Errorf("service '%s' is not registered", serviceName)
	}

	return circuitBreaker.GetState(), nil
}

// IsRecoveryInProgress returns true if recovery is currently in progress for the service
func (rm *RecoveryManager) IsRecoveryInProgress(serviceName string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.recoveryInProgress[serviceName]
}
