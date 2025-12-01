package di

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing health monitoring

// MockHealthyService implements HealthChecker with healthy responses
type MockHealthyService struct {
	checkCount int64
	mu         sync.Mutex
}

func (m *MockHealthyService) HealthCheck(ctx context.Context) HealthCheck {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkCount++

	return HealthCheck{
		Status:        HealthHealthy,
		Message:       "Service is healthy",
		Timestamp:     time.Now(),
		CheckDuration: 10 * time.Millisecond,
		Metadata: map[string]interface{}{
			"check_count": m.checkCount,
		},
	}
}

func (m *MockHealthyService) GetCheckCount() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.checkCount
}

// MockUnhealthyService implements HealthChecker with unhealthy responses
type MockUnhealthyService struct {
	checkCount int64
	mu         sync.Mutex
}

func (m *MockUnhealthyService) HealthCheck(ctx context.Context) HealthCheck {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkCount++

	return HealthCheck{
		Status:        HealthUnhealthy,
		Message:       "Service is down",
		Timestamp:     time.Now(),
		CheckDuration: 5 * time.Millisecond,
		Metadata: map[string]interface{}{
			"check_count": m.checkCount,
			"error":       "connection failed",
		},
	}
}

// MockFlappingService switches between healthy and unhealthy
type MockFlappingService struct {
	checkCount int64
	mu         sync.Mutex
}

func (m *MockFlappingService) HealthCheck(ctx context.Context) HealthCheck {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkCount++

	// Alternate between healthy and unhealthy
	if m.checkCount%2 == 0 {
		return HealthCheck{
			Status:        HealthHealthy,
			Message:       "Service recovered",
			Timestamp:     time.Now(),
			CheckDuration: 8 * time.Millisecond,
		}
	}

	return HealthCheck{
		Status:        HealthUnhealthy,
		Message:       "Service temporarily down",
		Timestamp:     time.Now(),
		CheckDuration: 12 * time.Millisecond,
	}
}

// MockHealthListener captures health change events for testing
type MockHealthListener struct {
	events []HealthEvent
	mu     sync.Mutex
}

func (m *MockHealthListener) OnHealthChanged(dependencyType string, oldStatus, newStatus HealthStatus, health HealthCheck) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = append(m.events, HealthEvent{
		DependencyType: dependencyType,
		OldStatus:      oldStatus,
		NewStatus:      newStatus,
		HealthCheck:    health,
		Timestamp:      time.Now(),
	})
}

func (m *MockHealthListener) GetEvents() []HealthEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	events := make([]HealthEvent, len(m.events))
	copy(events, m.events)
	return events
}

func (m *MockHealthListener) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = nil
}

func TestHealthMonitoring(t *testing.T) {
	t.Run("RegisterDependency_RegistersSuccessfully", func(t *testing.T) {
		monitor := NewHealthMonitor()
		service := &MockHealthyService{}

		monitor.RegisterDependency("test-service", service)

		health, exists := monitor.GetHealth("test-service")
		assert.True(t, exists)
		assert.Equal(t, HealthUnknown, health.Status) // Should be unknown before first check
	})

	t.Run("CheckHealth_HealthyService_ReturnsHealthy", func(t *testing.T) {
		monitor := NewHealthMonitor()
		service := &MockHealthyService{}

		monitor.RegisterDependency("test-service", service)

		ctx := context.Background()
		err := monitor.CheckHealth(ctx, "test-service")
		require.NoError(t, err)

		health, exists := monitor.GetHealth("test-service")
		require.True(t, exists)
		assert.Equal(t, HealthHealthy, health.Status)
		assert.Equal(t, "Service is healthy", health.Message)
		assert.Equal(t, int64(1), service.GetCheckCount())
	})

	t.Run("CheckHealth_UnhealthyService_ReturnsUnhealthy", func(t *testing.T) {
		monitor := NewHealthMonitor()
		service := &MockUnhealthyService{}

		monitor.RegisterDependency("test-service", service)

		ctx := context.Background()
		err := monitor.CheckHealth(ctx, "test-service")
		require.NoError(t, err)

		health, exists := monitor.GetHealth("test-service")
		require.True(t, exists)
		assert.Equal(t, HealthUnhealthy, health.Status)
		assert.Equal(t, "Service is down", health.Message)
	})

	t.Run("HealthListener_ReceivesStatusChanges", func(t *testing.T) {
		monitor := NewHealthMonitor()
		service := &MockFlappingService{}
		listener := &MockHealthListener{}

		monitor.AddListener(listener)
		monitor.RegisterDependency("flapping-service", service)

		ctx := context.Background()

		// First check - should change from Unknown to Unhealthy
		err := monitor.CheckHealth(ctx, "flapping-service")
		require.NoError(t, err)

		// Second check - should change from Unhealthy to Healthy
		err = monitor.CheckHealth(ctx, "flapping-service")
		require.NoError(t, err)

		// Give listeners time to process
		time.Sleep(100 * time.Millisecond)

		events := listener.GetEvents()
		require.Len(t, events, 2, "Expected exactly 2 health events")

		t.Logf("Event 0: %s -> %s", events[0].OldStatus.String(), events[0].NewStatus.String())
		t.Logf("Event 1: %s -> %s", events[1].OldStatus.String(), events[1].NewStatus.String())

		// Verify that both expected transitions occurred (order may vary due to goroutines)
		transitions := make(map[string]bool)
		for _, event := range events {
			assert.Equal(t, "flapping-service", event.DependencyType)
			transitionKey := fmt.Sprintf("%s->%s", event.OldStatus.String(), event.NewStatus.String())
			transitions[transitionKey] = true
		}

		assert.True(t, transitions["Unknown->Unhealthy"], "Should have Unknown -> Unhealthy transition")
		assert.True(t, transitions["Unhealthy->Healthy"], "Should have Unhealthy -> Healthy transition")
	})

	t.Run("GetOverallHealth_MixedStatuses_ReturnsAggregatedHealth", func(t *testing.T) {
		monitor := NewHealthMonitor()
		healthyService := &MockHealthyService{}
		unhealthyService := &MockUnhealthyService{}

		monitor.RegisterDependency("healthy-service", healthyService)
		monitor.RegisterDependency("unhealthy-service", unhealthyService)

		ctx := context.Background()
		monitor.CheckAllHealth(ctx)

		overallHealth := monitor.GetOverallHealth()
		assert.Equal(t, HealthUnhealthy, overallHealth.Status) // Should be unhealthy due to one unhealthy service
		assert.Contains(t, overallHealth.Message, "unhealthy-service is unhealthy")

		metadata := overallHealth.Metadata
		assert.Equal(t, 1, metadata["healthy_count"])
		assert.Equal(t, 1, metadata["unhealthy_count"])
		assert.Equal(t, 2, metadata["total_count"])
	})

	t.Run("IsHealthy_ChecksServiceStatus", func(t *testing.T) {
		monitor := NewHealthMonitor()
		service := &MockHealthyService{}

		monitor.RegisterDependency("test-service", service)

		// Before check - should be false (unknown status)
		assert.False(t, monitor.IsHealthy("test-service"))

		ctx := context.Background()
		err := monitor.CheckHealth(ctx, "test-service")
		require.NoError(t, err)

		// After check - should be true
		assert.True(t, monitor.IsHealthy("test-service"))
	})

	t.Run("GetUnhealthyDependencies_ReturnsUnhealthyOnly", func(t *testing.T) {
		monitor := NewHealthMonitor()
		healthyService := &MockHealthyService{}
		unhealthyService := &MockUnhealthyService{}

		monitor.RegisterDependency("healthy-service", healthyService)
		monitor.RegisterDependency("unhealthy-service", unhealthyService)

		ctx := context.Background()
		monitor.CheckAllHealth(ctx)

		unhealthyDeps := monitor.GetUnhealthyDependencies()
		assert.Len(t, unhealthyDeps, 1)
		assert.Contains(t, unhealthyDeps, "unhealthy-service")
	})
}

func TestCircuitBreaker(t *testing.T) {
	t.Run("Execute_SuccessfulOperation_RemainsClosedState", func(t *testing.T) {
		breaker := NewCircuitBreaker(DefaultCircuitBreakerConfig())

		err := breaker.Execute(func() error {
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, CircuitClosed, breaker.GetState())
		assert.Equal(t, int64(0), breaker.GetFailureCount())
	})

	t.Run("Execute_RepeatedFailures_OpensCircuit", func(t *testing.T) {
		config := CircuitBreakerConfig{
			FailureThreshold: 3,
			OpenTimeout:      1 * time.Second,
			SuccessThreshold: 2,
		}
		breaker := NewCircuitBreaker(config)

		failureErr := errors.New("operation failed")

		// Execute failures up to threshold
		for i := 0; i < 3; i++ {
			err := breaker.Execute(func() error {
				return failureErr
			})
			assert.Equal(t, failureErr, err)
		}

		assert.Equal(t, CircuitOpen, breaker.GetState())
		assert.Equal(t, int64(3), breaker.GetFailureCount())
	})

	t.Run("Execute_CircuitOpen_ReturnsCircuitBreakerError", func(t *testing.T) {
		config := CircuitBreakerConfig{
			FailureThreshold: 1,
			OpenTimeout:      1 * time.Second,
			SuccessThreshold: 2,
		}
		breaker := NewCircuitBreaker(config)

		// Cause circuit to open
		breaker.Execute(func() error {
			return errors.New("failure")
		})

		assert.Equal(t, CircuitOpen, breaker.GetState())

		// Try to execute - should get circuit breaker error
		err := breaker.Execute(func() error {
			return nil
		})

		assert.Error(t, err)
		assert.IsType(t, &CircuitBreakerError{}, err)
	})

	t.Run("Execute_HalfOpenToClosedTransition", func(t *testing.T) {
		config := CircuitBreakerConfig{
			FailureThreshold: 1,
			OpenTimeout:      100 * time.Millisecond,
			SuccessThreshold: 2,
		}
		breaker := NewCircuitBreaker(config)

		// Open the circuit
		breaker.Execute(func() error {
			return errors.New("failure")
		})
		assert.Equal(t, CircuitOpen, breaker.GetState())

		// Wait for open timeout
		time.Sleep(150 * time.Millisecond)

		// Execute successful operations to close circuit
		for i := 0; i < 2; i++ {
			err := breaker.Execute(func() error {
				return nil
			})
			assert.NoError(t, err)
		}

		assert.Equal(t, CircuitClosed, breaker.GetState())
	})

	t.Run("Reset_ResetsCircuitState", func(t *testing.T) {
		breaker := NewCircuitBreaker(DefaultCircuitBreakerConfig())

		// Cause some failures
		for i := 0; i < 3; i++ {
			breaker.Execute(func() error {
				return errors.New("failure")
			})
		}

		breaker.Reset()

		assert.Equal(t, CircuitClosed, breaker.GetState())
		assert.Equal(t, int64(0), breaker.GetFailureCount())
	})
}

func TestCircuitBreakerManager(t *testing.T) {
	t.Run("GetCircuitBreaker_CreatesAndReusesBreakers", func(t *testing.T) {
		manager := NewCircuitBreakerManager(DefaultCircuitBreakerConfig())

		breaker1 := manager.GetCircuitBreaker("service1")
		breaker2 := manager.GetCircuitBreaker("service1")
		breaker3 := manager.GetCircuitBreaker("service2")

		assert.Same(t, breaker1, breaker2)
		assert.NotSame(t, breaker1, breaker3)
	})

	t.Run("ExecuteWithBreaker_UsesCorrectBreaker", func(t *testing.T) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{
			FailureThreshold: 1,
			OpenTimeout:      1 * time.Second,
			SuccessThreshold: 1,
		})

		// Fail service1
		err := manager.ExecuteWithBreaker("service1", func() error {
			return errors.New("service1 failed")
		})
		assert.Error(t, err)
		assert.Equal(t, CircuitOpen, manager.GetBreakerState("service1"))

		// Service2 should still be closed
		assert.Equal(t, CircuitClosed, manager.GetBreakerState("service2"))

		// Execute on service2 should succeed
		err = manager.ExecuteWithBreaker("service2", func() error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("GetOpenBreakers_ReturnsOpenBreakerNames", func(t *testing.T) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{
			FailureThreshold: 1,
			OpenTimeout:      1 * time.Second,
			SuccessThreshold: 1,
		})

		// Open two breakers
		manager.ExecuteWithBreaker("service1", func() error {
			return errors.New("failure")
		})
		manager.ExecuteWithBreaker("service2", func() error {
			return errors.New("failure")
		})

		openBreakers := manager.GetOpenBreakers()
		assert.Len(t, openBreakers, 2)
		assert.Contains(t, openBreakers, "service1")
		assert.Contains(t, openBreakers, "service2")
	})
}

func TestHealthAwareCircuitBreakerIntegration(t *testing.T) {
	t.Run("HealthMonitor_IntegratesWithCircuitBreaker", func(t *testing.T) {
		monitor := NewHealthMonitor()
		cbManager := NewCircuitBreakerManager(CircuitBreakerConfig{
			FailureThreshold: 2,
			OpenTimeout:      100 * time.Millisecond,
			SuccessThreshold: 2,
		})

		cbListener := NewHealthAwareCircuitBreakerListener(cbManager)
		monitor.AddListener(cbListener)

		// Register a flapping service
		service := &MockFlappingService{}
		monitor.RegisterDependency("flapping-service", service)

		ctx := context.Background()

		// First check - should be unhealthy (odd check count)
		err := monitor.CheckHealth(ctx, "flapping-service")
		require.NoError(t, err)

		// Give listeners time to process
		time.Sleep(50 * time.Millisecond)

		// Check circuit breaker state - should have recorded one failure
		breaker := cbManager.GetCircuitBreaker("flapping-service")
		assert.Equal(t, int64(1), breaker.GetFailureCount())
		assert.Equal(t, CircuitClosed, breaker.GetState()) // Still closed as we need 2 failures

		// Second check - should be healthy (even check count)
		err = monitor.CheckHealth(ctx, "flapping-service")
		require.NoError(t, err)

		// Give listeners time to process
		time.Sleep(50 * time.Millisecond)

		// Circuit breaker should have recorded success and reset failure count
		assert.Equal(t, int64(0), breaker.GetFailureCount())
		assert.Equal(t, CircuitClosed, breaker.GetState())
	})
}

func TestDependencyHealthMonitoring(t *testing.T) {
	t.Run("Container_IntegratesHealthMonitoring", func(t *testing.T) {
		container := NewContainer()
		healthMonitor := container.GetHealthMonitor()

		// Register a health-checked service
		service := &MockHealthyService{}
		err := Register[HealthChecker](container, service, Singleton)
		require.NoError(t, err)

		// Resolve the service
		resolvedService, err := Resolve[HealthChecker](container)
		require.NoError(t, err)
		require.NotNil(t, resolvedService)

		// Check that it's registered for health monitoring
		// The dependency type should be the interface name
		dependencyType := "di.HealthChecker"
		health, exists := healthMonitor.GetHealth(dependencyType)
		assert.True(t, exists, "Service should be registered for health monitoring")

		// Initially should be unknown
		assert.Equal(t, HealthUnknown, health.Status)

		// Perform health check
		ctx := context.Background()
		err = healthMonitor.CheckHealth(ctx, dependencyType)
		require.NoError(t, err)

		// Should now be healthy
		health, exists = healthMonitor.GetHealth(dependencyType)
		require.True(t, exists)
		assert.Equal(t, HealthHealthy, health.Status)
		assert.Equal(t, "Service is healthy", health.Message)
	})
}
