package lifecycle

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockServiceLifecycle is a mock implementation of ServiceLifecycleInterface
type MockServiceLifecycle struct {
	mock.Mock
	state ServiceState
	hooks []LifecycleHook
}

// NewMockServiceLifecycle creates a new mock service
func NewMockServiceLifecycle(initialState ServiceState) *MockServiceLifecycle {
	return &MockServiceLifecycle{
		state: initialState,
		hooks: make([]LifecycleHook, 0),
	}
}

// Start mocks the Start method
func (m *MockServiceLifecycle) Start(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.state = Running
	} else {
		m.state = Failed
	}
	return args.Error(0)
}

// Stop mocks the Stop method
func (m *MockServiceLifecycle) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.state = Stopped
	} else {
		m.state = Failed
	}
	return args.Error(0)
}

// HealthCheck mocks the HealthCheck method
func (m *MockServiceLifecycle) HealthCheck(ctx context.Context) HealthStatus {
	args := m.Called(ctx)
	return args.Get(0).(HealthStatus)
}

// GetState returns the current state
func (m *MockServiceLifecycle) GetState() ServiceState {
	return m.state
}

// AddHook adds a lifecycle hook
func (m *MockServiceLifecycle) AddHook(hook LifecycleHook) error {
	args := m.Called(hook)
	if args.Error(0) == nil {
		m.hooks = append(m.hooks, hook)
	}
	return args.Error(0)
}

// SetState allows setting the state for testing
func (m *MockServiceLifecycle) SetState(state ServiceState) {
	m.state = state
}

// MockHook is a mock implementation of LifecycleHook
type MockHook struct {
	mock.Mock
	critical bool
	timeout  time.Duration
}

// NewMockHook creates a new mock hook
func NewMockHook(critical bool, timeout time.Duration) *MockHook {
	return &MockHook{
		critical: critical,
		timeout:  timeout,
	}
}

func (m *MockHook) BeforeStart(ctx context.Context, service ServiceLifecycleInterface) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockHook) AfterStart(ctx context.Context, service ServiceLifecycleInterface) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockHook) BeforeStop(ctx context.Context, service ServiceLifecycleInterface) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockHook) AfterStop(ctx context.Context, service ServiceLifecycleInterface) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockHook) IsCritical() bool {
	return m.critical
}

func (m *MockHook) GetTimeout() time.Duration {
	return m.timeout
}

func TestServiceState_String(t *testing.T) {
	tests := []struct {
		state    ServiceState
		expected string
	}{
		{Created, "Created"},
		{Starting, "Starting"},
		{Running, "Running"},
		{Stopping, "Stopping"},
		{Stopped, "Stopped"},
		{Failed, "Failed"},
		{ServiceState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestServiceHealthStatus_String(t *testing.T) {
	tests := []struct {
		status   ServiceHealthStatus
		expected string
	}{
		{Healthy, "Healthy"},
		{Degraded, "Degraded"},
		{Unhealthy, "Unhealthy"},
		{Unknown, "Unknown"},
		{ServiceHealthStatus(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestHookManager_AddHook(t *testing.T) {
	config := DefaultHookConfiguration()
	manager := NewHookManager(config)

	hook := NewMockHook(false, 5*time.Second)

	err := manager.AddHook(BeforeStartPhase, hook)
	assert.NoError(t, err)

	counts := manager.GetHooksCount()
	assert.Equal(t, 1, counts[BeforeStartPhase])
	assert.Equal(t, 0, counts[AfterStartPhase])
}

func TestHookManager_AddNilHook(t *testing.T) {
	config := DefaultHookConfiguration()
	manager := NewHookManager(config)

	err := manager.AddHook(BeforeStartPhase, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hook cannot be nil")
}

func TestHookManager_ExecuteHooks_Success(t *testing.T) {
	config := DefaultHookConfiguration()
	manager := NewHookManager(config)

	service := NewMockServiceLifecycle(Created)
	hook := NewMockHook(false, 5*time.Second)

	hook.On("BeforeStart", mock.Anything, service).Return(nil)

	err := manager.AddHook(BeforeStartPhase, hook)
	require.NoError(t, err)

	ctx := context.Background()
	err = manager.ExecuteHooks(ctx, BeforeStartPhase, service)
	assert.NoError(t, err)

	hook.AssertExpectations(t)
}

func TestHookManager_ExecuteHooks_Failure_FailFast(t *testing.T) {
	config := HookConfiguration{
		Timeout:       30 * time.Second,
		FailurePolicy: FailFast,
		Critical:      false,
	}
	manager := NewHookManager(config)

	service := NewMockServiceLifecycle(Created)
	hook := NewMockHook(false, 5*time.Second)

	hook.On("BeforeStart", mock.Anything, service).Return(assert.AnError)

	err := manager.AddHook(BeforeStartPhase, hook)
	require.NoError(t, err)

	ctx := context.Background()
	err = manager.ExecuteHooks(ctx, BeforeStartPhase, service)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hook 0 in phase BeforeStart failed")

	hook.AssertExpectations(t)
}

func TestHookManager_ExecuteHooks_Failure_FailOnCritical(t *testing.T) {
	config := HookConfiguration{
		Timeout:       30 * time.Second,
		FailurePolicy: FailOnCritical,
		Critical:      false,
	}
	manager := NewHookManager(config)

	service := NewMockServiceLifecycle(Created)

	// Non-critical hook that fails
	nonCriticalHook := NewMockHook(false, 5*time.Second)
	nonCriticalHook.On("BeforeStart", mock.Anything, service).Return(assert.AnError)

	// Critical hook that succeeds
	criticalHook := NewMockHook(true, 5*time.Second)
	criticalHook.On("BeforeStart", mock.Anything, service).Return(nil)

	err := manager.AddHook(BeforeStartPhase, nonCriticalHook)
	require.NoError(t, err)
	err = manager.AddHook(BeforeStartPhase, criticalHook)
	require.NoError(t, err)

	ctx := context.Background()
	err = manager.ExecuteHooks(ctx, BeforeStartPhase, service)
	assert.NoError(t, err) // Should not fail because non-critical hook failure is ignored

	nonCriticalHook.AssertExpectations(t)
	criticalHook.AssertExpectations(t)
}

func TestHookManager_ExecuteHooks_CriticalFailure(t *testing.T) {
	config := HookConfiguration{
		Timeout:       30 * time.Second,
		FailurePolicy: FailOnCritical,
		Critical:      false,
	}
	manager := NewHookManager(config)

	service := NewMockServiceLifecycle(Created)

	// Critical hook that fails
	criticalHook := NewMockHook(true, 5*time.Second)
	criticalHook.On("BeforeStart", mock.Anything, service).Return(assert.AnError)

	err := manager.AddHook(BeforeStartPhase, criticalHook)
	require.NoError(t, err)

	ctx := context.Background()
	err = manager.ExecuteHooks(ctx, BeforeStartPhase, service)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "critical hook 0 in phase BeforeStart failed")

	criticalHook.AssertExpectations(t)
}

func TestBaseHook_DefaultImplementation(t *testing.T) {
	config := DefaultHookConfiguration()
	hook := NewBaseHook(config)

	service := NewMockServiceLifecycle(Created)
	ctx := context.Background()

	// All default implementations should return nil
	assert.NoError(t, hook.BeforeStart(ctx, service))
	assert.NoError(t, hook.AfterStart(ctx, service))
	assert.NoError(t, hook.BeforeStop(ctx, service))
	assert.NoError(t, hook.AfterStop(ctx, service))

	assert.Equal(t, config.Critical, hook.IsCritical())
	assert.Equal(t, config.Timeout, hook.GetTimeout())
}

func TestDependencyGraph_AddDependency(t *testing.T) {
	graph := NewDependencyGraph()

	graph.AddDependency("serviceA", "serviceB")
	graph.AddDependency("serviceA", "serviceC")
	graph.AddDependency("serviceB", "serviceD")

	// Check forward dependencies
	depsA := graph.GetDependencies("serviceA")
	assert.ElementsMatch(t, []string{"serviceB", "serviceC"}, depsA)

	depsB := graph.GetDependencies("serviceB")
	assert.ElementsMatch(t, []string{"serviceD"}, depsB)

	// Check reverse dependencies
	dependentsB := graph.GetDependents("serviceB")
	assert.ElementsMatch(t, []string{"serviceA"}, dependentsB)

	dependentsC := graph.GetDependents("serviceC")
	assert.ElementsMatch(t, []string{"serviceA"}, dependentsC)

	dependentsD := graph.GetDependents("serviceD")
	assert.ElementsMatch(t, []string{"serviceB"}, dependentsD)
}

func TestDependencyGraph_AddDependency_NoDuplicates(t *testing.T) {
	graph := NewDependencyGraph()

	// Add same dependency twice
	graph.AddDependency("serviceA", "serviceB")
	graph.AddDependency("serviceA", "serviceB")

	deps := graph.GetDependencies("serviceA")
	assert.Len(t, deps, 1)
	assert.Equal(t, "serviceB", deps[0])

	dependents := graph.GetDependents("serviceB")
	assert.Len(t, dependents, 1)
	assert.Equal(t, "serviceA", dependents[0])
}

func TestDependencyGraph_TopologicalSort_Simple(t *testing.T) {
	graph := NewDependencyGraph()

	// Create dependency chain: A -> B -> C
	graph.AddDependency("serviceA", "serviceB")
	graph.AddDependency("serviceB", "serviceC")

	order, err := graph.TopologicalSort()
	assert.NoError(t, err)

	// serviceC should come first (no dependencies)
	// serviceB should come second (depends on C)
	// serviceA should come last (depends on B)
	assert.Equal(t, []string{"serviceC", "serviceB", "serviceA"}, order)
}

func TestDependencyGraph_TopologicalSort_Complex(t *testing.T) {
	graph := NewDependencyGraph()

	// Create complex dependency graph
	graph.AddDependency("A", "B")
	graph.AddDependency("A", "C")
	graph.AddDependency("B", "D")
	graph.AddDependency("C", "D")
	graph.AddDependency("D", "E")

	order, err := graph.TopologicalSort()
	assert.NoError(t, err)
	assert.Len(t, order, 5)

	// E should come first (no dependencies)
	assert.Equal(t, "E", order[0])

	// D should come before B and C
	dIndex := findIndex(order, "D")
	bIndex := findIndex(order, "B")
	cIndex := findIndex(order, "C")
	aIndex := findIndex(order, "A")

	assert.True(t, dIndex < bIndex)
	assert.True(t, dIndex < cIndex)
	assert.True(t, bIndex < aIndex)
	assert.True(t, cIndex < aIndex)
}

func TestDependencyGraph_TopologicalSort_CircularDependency(t *testing.T) {
	graph := NewDependencyGraph()

	// Create circular dependency: A -> B -> C -> A
	graph.AddDependency("A", "B")
	graph.AddDependency("B", "C")
	graph.AddDependency("C", "A")

	order, err := graph.TopologicalSort()
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "circular dependency detected")
}

func TestLifecycleManager_RegisterService(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	service := NewMockServiceLifecycle(Created)
	err := manager.RegisterService("testService", service)
	assert.NoError(t, err)

	state, err := manager.GetServiceState("testService")
	assert.NoError(t, err)
	assert.Equal(t, Created, state)
}

func TestLifecycleManager_RegisterService_EmptyName(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	service := NewMockServiceLifecycle(Created)
	err := manager.RegisterService("", service)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service name cannot be empty")
}

func TestLifecycleManager_RegisterService_NilService(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	err := manager.RegisterService("testService", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service cannot be nil")
}

func TestLifecycleManager_RegisterService_Duplicate(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	service := NewMockServiceLifecycle(Created)
	err := manager.RegisterService("testService", service)
	assert.NoError(t, err)

	// Try to register same name again
	service2 := NewMockServiceLifecycle(Created)
	err = manager.RegisterService("testService", service2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service 'testService' is already registered")
}

func TestLifecycleManager_AddServiceDependency(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	serviceA := NewMockServiceLifecycle(Created)
	serviceB := NewMockServiceLifecycle(Created)

	err := manager.RegisterService("serviceA", serviceA)
	require.NoError(t, err)
	err = manager.RegisterService("serviceB", serviceB)
	require.NoError(t, err)

	err = manager.AddServiceDependency("serviceA", "serviceB")
	assert.NoError(t, err)

	graph := manager.GetDependencyGraph()
	deps := graph.GetDependencies("serviceA")
	assert.ElementsMatch(t, []string{"serviceB"}, deps)
}

func TestLifecycleManager_AddServiceDependency_UnregisteredDependent(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	serviceB := NewMockServiceLifecycle(Created)
	err := manager.RegisterService("serviceB", serviceB)
	require.NoError(t, err)

	err = manager.AddServiceDependency("serviceA", "serviceB")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dependent service 'serviceA' is not registered")
}

func TestLifecycleManager_AddServiceDependency_UnregisteredDependency(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	serviceA := NewMockServiceLifecycle(Created)
	err := manager.RegisterService("serviceA", serviceA)
	require.NoError(t, err)

	err = manager.AddServiceDependency("serviceA", "serviceB")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dependency service 'serviceB' is not registered")
}

func TestLifecycleManager_StartAll_Simple(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	service := NewMockServiceLifecycle(Created)
	service.On("Start", mock.Anything).Return(nil)

	err := manager.RegisterService("testService", service)
	require.NoError(t, err)

	ctx := context.Background()
	err = manager.StartAll(ctx)
	assert.NoError(t, err)

	state, err := manager.GetServiceState("testService")
	assert.NoError(t, err)
	assert.Equal(t, Running, state)

	service.AssertExpectations(t)
}

func TestLifecycleManager_StartAll_WithDependencies(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	// Create services
	serviceA := NewMockServiceLifecycle(Created)
	serviceB := NewMockServiceLifecycle(Created)

	serviceA.On("Start", mock.Anything).Return(nil)
	serviceB.On("Start", mock.Anything).Return(nil)

	// Register services
	err := manager.RegisterService("serviceA", serviceA)
	require.NoError(t, err)
	err = manager.RegisterService("serviceB", serviceB)
	require.NoError(t, err)

	// Add dependency: A depends on B
	err = manager.AddServiceDependency("serviceA", "serviceB")
	require.NoError(t, err)

	ctx := context.Background()
	err = manager.StartAll(ctx)
	assert.NoError(t, err)

	// Both services should be running
	stateA, err := manager.GetServiceState("serviceA")
	assert.NoError(t, err)
	assert.Equal(t, Running, stateA)

	stateB, err := manager.GetServiceState("serviceB")
	assert.NoError(t, err)
	assert.Equal(t, Running, stateB)

	serviceA.AssertExpectations(t)
	serviceB.AssertExpectations(t)
}

func TestLifecycleManager_StopAll(t *testing.T) {
	manager := NewLifecycleManager(AllHealthyPolicy)

	service := NewMockServiceLifecycle(Created)
	service.On("Start", mock.Anything).Return(nil)
	service.On("Stop", mock.Anything).Return(nil)

	err := manager.RegisterService("testService", service)
	require.NoError(t, err)

	ctx := context.Background()
	err = manager.StartAll(ctx)
	require.NoError(t, err)

	err = manager.StopAll(ctx)
	assert.NoError(t, err)

	state, err := manager.GetServiceState("testService")
	assert.NoError(t, err)
	assert.Equal(t, Stopped, state)

	service.AssertExpectations(t)
}

// Helper function to find index of element in slice
func findIndex(slice []string, element string) int {
	for i, v := range slice {
		if v == element {
			return i
		}
	}
	return -1
}
