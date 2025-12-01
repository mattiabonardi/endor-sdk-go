package di

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for lifecycle testing

// MockLifecycleService implements both Startable and Stoppable
type MockLifecycleService struct {
	name      string
	started   bool
	stopped   bool
	startedAt time.Time
	stoppedAt time.Time
	startErr  error
	stopErr   error
	mu        sync.RWMutex
}

func NewMockLifecycleService(name string) *MockLifecycleService {
	return &MockLifecycleService{
		name: name,
	}
}

func (m *MockLifecycleService) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.startErr != nil {
		return m.startErr
	}

	m.started = true
	m.startedAt = time.Now()
	return nil
}

func (m *MockLifecycleService) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopErr != nil {
		return m.stopErr
	}

	m.stopped = true
	m.stoppedAt = time.Now()
	return nil
}

func (m *MockLifecycleService) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

func (m *MockLifecycleService) IsStopped() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stopped
}

func (m *MockLifecycleService) GetName() string {
	return m.name
}

func (m *MockLifecycleService) SetStartError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startErr = err
}

func (m *MockLifecycleService) SetStopError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopErr = err
}

func (m *MockLifecycleService) GetStartedAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.startedAt
}

func (m *MockLifecycleService) GetStoppedAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stoppedAt
}

// Mock lifecycle listener for testing
type MockLifecycleListener struct {
	events []LifecycleEvent
	mu     sync.Mutex
}

func (m *MockLifecycleListener) OnLifecycleChanged(event LifecycleEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

func (m *MockLifecycleListener) GetEvents() []LifecycleEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	events := make([]LifecycleEvent, len(m.events))
	copy(events, m.events)
	return events
}

func (m *MockLifecycleListener) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = nil
}

func TestLifecycleManager(t *testing.T) {
	t.Run("RegisterDependency_RegistersSuccessfully", func(t *testing.T) {
		lm := NewLifecycleManager()
		service := NewMockLifecycleService("test-service")

		lm.RegisterDependency("test-service", service)

		state, exists := lm.GetLifecycleState("test-service")
		assert.True(t, exists)
		assert.Equal(t, LifecycleStateRegistered, state)
	})

	t.Run("StartAll_StartsServicesInOrder", func(t *testing.T) {
		lm := NewLifecycleManager()
		service1 := NewMockLifecycleService("service1")
		service2 := NewMockLifecycleService("service2")
		service3 := NewMockLifecycleService("service3")

		// Register with different priorities
		lm.RegisterDependency("service1", service1, WithPriority(PriorityLow))
		lm.RegisterDependency("service2", service2, WithPriority(PriorityHigh))
		lm.RegisterDependency("service3", service3, WithPriority(PriorityNormal))

		ctx := context.Background()
		err := lm.StartAll(ctx)
		require.NoError(t, err)

		// All services should be started
		assert.True(t, service1.IsStarted())
		assert.True(t, service2.IsStarted())
		assert.True(t, service3.IsStarted())

		// Check startup order (High, Normal, Low)
		service2StartTime := service2.GetStartedAt()
		service3StartTime := service3.GetStartedAt()
		service1StartTime := service1.GetStartedAt()

		assert.True(t, service2StartTime.Before(service3StartTime) || service2StartTime.Equal(service3StartTime))
		assert.True(t, service3StartTime.Before(service1StartTime) || service3StartTime.Equal(service1StartTime))
	})

	t.Run("StartAll_WithDependencies_RespectsOrder", func(t *testing.T) {
		lm := NewLifecycleManager()
		database := NewMockLifecycleService("database")
		cache := NewMockLifecycleService("cache")
		webServer := NewMockLifecycleService("webserver")

		// Register dependencies: webserver depends on database and cache
		lm.RegisterDependency("database", database, WithPriority(PriorityHigh))
		lm.RegisterDependency("cache", cache, WithPriority(PriorityHigh))
		lm.RegisterDependency("webserver", webServer, WithPriority(PriorityNormal))

		// Add dependencies
		lm.AddDependency("webserver", "database")
		lm.AddDependency("webserver", "cache")

		ctx := context.Background()
		err := lm.StartAll(ctx)
		require.NoError(t, err)

		// All services should be started
		assert.True(t, database.IsStarted())
		assert.True(t, cache.IsStarted())
		assert.True(t, webServer.IsStarted())

		// Database and cache should start before webserver
		dbStartTime := database.GetStartedAt()
		cacheStartTime := cache.GetStartedAt()
		webStartTime := webServer.GetStartedAt()

		assert.True(t, dbStartTime.Before(webStartTime))
		assert.True(t, cacheStartTime.Before(webStartTime))
	})

	t.Run("StartAll_HandlesStartupFailure", func(t *testing.T) {
		lm := NewLifecycleManager()
		goodService := NewMockLifecycleService("good-service")
		badService := NewMockLifecycleService("bad-service")

		badService.SetStartError(errors.New("startup failed"))

		lm.RegisterDependency("good-service", goodService, WithPriority(PriorityHigh))
		lm.RegisterDependency("bad-service", badService, WithPriority(PriorityNormal))

		ctx := context.Background()
		err := lm.StartAll(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "startup failed")

		// Good service should have been started but then stopped due to failure
		// Bad service should not be started
		state1, _ := lm.GetLifecycleState("good-service")
		state2, _ := lm.GetLifecycleState("bad-service")

		assert.Equal(t, LifecycleStateStopped, state1)
		assert.Equal(t, LifecycleStateError, state2)
	})

	t.Run("StopAll_StopsInReverseOrder", func(t *testing.T) {
		lm := NewLifecycleManager()
		service1 := NewMockLifecycleService("service1")
		service2 := NewMockLifecycleService("service2")

		lm.RegisterDependency("service1", service1, WithPriority(PriorityHigh))
		lm.RegisterDependency("service2", service2, WithPriority(PriorityLow))

		ctx := context.Background()
		err := lm.StartAll(ctx)
		require.NoError(t, err)

		// Clear the start times for testing stop order
		time.Sleep(10 * time.Millisecond)

		err = lm.StopAll(ctx)
		require.NoError(t, err)

		// Both services should be stopped
		assert.True(t, service1.IsStopped())
		assert.True(t, service2.IsStopped())

		// Stop order should be reverse of start order (Low priority stops first)
		service1StopTime := service1.GetStoppedAt()
		service2StopTime := service2.GetStoppedAt()

		assert.True(t, service2StopTime.Before(service1StopTime) || service2StopTime.Equal(service1StopTime))
	})

	t.Run("AddListener_ReceivesLifecycleEvents", func(t *testing.T) {
		lm := NewLifecycleManager()
		listener := &MockLifecycleListener{}
		service := NewMockLifecycleService("test-service")

		lm.AddListener(listener)
		lm.RegisterDependency("test-service", service)

		ctx := context.Background()
		err := lm.StartAll(ctx)
		require.NoError(t, err)

		// Give listeners time to process
		time.Sleep(100 * time.Millisecond)

		events := listener.GetEvents()
		assert.True(t, len(events) >= 3) // At least: Register, Starting, Started

		// Check for registration event
		var hasRegisterEvent bool
		var hasStartingEvent bool
		var hasStartedEvent bool

		for _, event := range events {
			if event.DependencyType == "test-service" {
				switch event.NewState {
				case LifecycleStateRegistered:
					hasRegisterEvent = true
				case LifecycleStateStarting:
					hasStartingEvent = true
				case LifecycleStateStarted:
					hasStartedEvent = true
				}
			}
		}

		assert.True(t, hasRegisterEvent, "Should have registration event")
		assert.True(t, hasStartingEvent, "Should have starting event")
		assert.True(t, hasStartedEvent, "Should have started event")
	})

	t.Run("GetDependencyGraph_ReturnsCorrectGraph", func(t *testing.T) {
		lm := NewLifecycleManager()
		database := NewMockLifecycleService("database")
		cache := NewMockLifecycleService("cache")
		webServer := NewMockLifecycleService("webserver")

		lm.RegisterDependency("database", database)
		lm.RegisterDependency("cache", cache)
		lm.RegisterDependency("webserver", webServer)

		lm.AddDependency("webserver", "database")
		lm.AddDependency("webserver", "cache")

		graph := lm.GetDependencyGraph()

		assert.Len(t, graph, 3)
		assert.Contains(t, graph, "database")
		assert.Contains(t, graph, "cache")
		assert.Contains(t, graph, "webserver")

		// Webserver should depend on database and cache
		webDeps := graph["webserver"]
		assert.Len(t, webDeps, 2)
		assert.Contains(t, webDeps, "database")
		assert.Contains(t, webDeps, "cache")

		// Database and cache should have no dependencies
		assert.Empty(t, graph["database"])
		assert.Empty(t, graph["cache"])
	})

	t.Run("CircularDependency_DetectedAndPrevented", func(t *testing.T) {
		lm := NewLifecycleManager()
		service1 := NewMockLifecycleService("service1")
		service2 := NewMockLifecycleService("service2")

		lm.RegisterDependency("service1", service1)
		lm.RegisterDependency("service2", service2)

		// Create circular dependency
		lm.AddDependency("service1", "service2")
		lm.AddDependency("service2", "service1")

		ctx := context.Background()
		err := lm.StartAll(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency")
	})
}

func TestContainerLifecycleIntegration(t *testing.T) {
	t.Run("Container_AutoRegistersLifecycleServices", func(t *testing.T) {
		container := NewContainer()
		lifecycleManager := container.GetLifecycleManager()

		service := NewMockLifecycleService("lifecycle-service")

		// Register a lifecycle-aware service
		err := Register[LifecycleAware](container, service, Singleton)
		require.NoError(t, err)

		// Resolve the service to trigger registration
		resolved, err := Resolve[LifecycleAware](container)
		require.NoError(t, err)
		require.NotNil(t, resolved)

		// Check that it was registered for lifecycle management
		dependencyType := "di.LifecycleAware"
		state, exists := lifecycleManager.GetLifecycleState(dependencyType)
		assert.True(t, exists, "Service should be registered for lifecycle management")
		assert.Equal(t, LifecycleStateRegistered, state)
	})

	t.Run("Container_LifecycleManagement_StartStopCycle", func(t *testing.T) {
		container := NewContainer()
		lifecycleManager := container.GetLifecycleManager()

		service := NewMockLifecycleService("test-service")

		// Register and resolve the service
		err := Register[LifecycleAware](container, service, Singleton)
		require.NoError(t, err)

		_, err = Resolve[LifecycleAware](container)
		require.NoError(t, err)

		// Start all services
		ctx := context.Background()
		err = lifecycleManager.StartAll(ctx)
		require.NoError(t, err)

		// Check that service was started
		assert.True(t, service.IsStarted())
		dependencyType := "di.LifecycleAware"
		state, exists := lifecycleManager.GetLifecycleState(dependencyType)
		require.True(t, exists)
		assert.Equal(t, LifecycleStateStarted, state)

		// Stop all services
		err = lifecycleManager.StopAll(ctx)
		require.NoError(t, err)

		// Check that service was stopped
		assert.True(t, service.IsStopped())
		state, exists = lifecycleManager.GetLifecycleState(dependencyType)
		require.True(t, exists)
		assert.Equal(t, LifecycleStateStopped, state)
	})
}

func TestLifecycleStateTransitions(t *testing.T) {
	t.Run("LifecycleStateString_ReturnsCorrectValues", func(t *testing.T) {
		assert.Equal(t, "Unknown", LifecycleStateUnknown.String())
		assert.Equal(t, "Registered", LifecycleStateRegistered.String())
		assert.Equal(t, "Starting", LifecycleStateStarting.String())
		assert.Equal(t, "Started", LifecycleStateStarted.String())
		assert.Equal(t, "Stopping", LifecycleStateStopping.String())
		assert.Equal(t, "Stopped", LifecycleStateStopped.String())
		assert.Equal(t, "Error", LifecycleStateError.String())
	})

	t.Run("LifecycleOptions_ConfigureProperly", func(t *testing.T) {
		lm := NewLifecycleManager()
		service := NewMockLifecycleService("test-service")

		lm.RegisterDependency("test-service", service,
			WithPriority(PriorityHigh),
			WithDependencies("dependency1", "dependency2"))

		// The dependency should be registered with the specified configuration
		state, exists := lm.GetLifecycleState("test-service")
		assert.True(t, exists)
		assert.Equal(t, LifecycleStateRegistered, state)

		graph := lm.GetDependencyGraph()
		deps := graph["test-service"]
		assert.Len(t, deps, 2)
		assert.Contains(t, deps, "dependency1")
		assert.Contains(t, deps, "dependency2")
	})
}
