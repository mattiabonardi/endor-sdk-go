package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HookPhase represents the phase when a lifecycle hook is executed
type HookPhase int

const (
	// BeforeStartPhase executes before service startup
	BeforeStartPhase HookPhase = iota
	// AfterStartPhase executes after successful service startup
	AfterStartPhase
	// BeforeStopPhase executes before service shutdown
	BeforeStopPhase
	// AfterStopPhase executes after successful service shutdown
	AfterStopPhase
)

// String returns the string representation of the hook phase
func (h HookPhase) String() string {
	switch h {
	case BeforeStartPhase:
		return "BeforeStart"
	case AfterStartPhase:
		return "AfterStart"
	case BeforeStopPhase:
		return "BeforeStop"
	case AfterStopPhase:
		return "AfterStop"
	default:
		return "Unknown"
	}
}

// HookFailurePolicy defines how to handle hook execution failures
type HookFailurePolicy int

const (
	// FailFast stops execution immediately if a hook fails
	FailFast HookFailurePolicy = iota
	// Continue continues execution even if a hook fails
	Continue
	// FailOnCritical only fails if the hook is marked as critical
	FailOnCritical
)

// LifecycleHook defines the interface for lifecycle hooks
type LifecycleHook interface {
	// BeforeStart is called before service startup
	BeforeStart(ctx context.Context, service ServiceLifecycleInterface) error

	// AfterStart is called after successful service startup
	AfterStart(ctx context.Context, service ServiceLifecycleInterface) error

	// BeforeStop is called before service shutdown
	BeforeStop(ctx context.Context, service ServiceLifecycleInterface) error

	// AfterStop is called after successful service shutdown
	AfterStop(ctx context.Context, service ServiceLifecycleInterface) error

	// IsCritical returns true if this hook is critical for service operation
	IsCritical() bool

	// GetTimeout returns the maximum execution time for this hook
	GetTimeout() time.Duration
}

// HookConfiguration contains configuration for hook execution
type HookConfiguration struct {
	// Timeout is the maximum time to wait for hook execution
	Timeout time.Duration
	// FailurePolicy defines how to handle hook failures
	FailurePolicy HookFailurePolicy
	// Critical indicates if this hook is critical for service operation
	Critical bool
}

// DefaultHookConfiguration returns the default configuration for hooks
func DefaultHookConfiguration() HookConfiguration {
	return HookConfiguration{
		Timeout:       30 * time.Second,
		FailurePolicy: FailOnCritical,
		Critical:      false,
	}
}

// BaseHook provides a default implementation of the LifecycleHook interface
type BaseHook struct {
	config HookConfiguration
}

// NewBaseHook creates a new base hook with the given configuration
func NewBaseHook(config HookConfiguration) *BaseHook {
	return &BaseHook{
		config: config,
	}
}

// BeforeStart provides a default no-op implementation
func (h *BaseHook) BeforeStart(ctx context.Context, service ServiceLifecycleInterface) error {
	return nil
}

// AfterStart provides a default no-op implementation
func (h *BaseHook) AfterStart(ctx context.Context, service ServiceLifecycleInterface) error {
	return nil
}

// BeforeStop provides a default no-op implementation
func (h *BaseHook) BeforeStop(ctx context.Context, service ServiceLifecycleInterface) error {
	return nil
}

// AfterStop provides a default no-op implementation
func (h *BaseHook) AfterStop(ctx context.Context, service ServiceLifecycleInterface) error {
	return nil
}

// IsCritical returns whether this hook is critical
func (h *BaseHook) IsCritical() bool {
	return h.config.Critical
}

// GetTimeout returns the hook timeout
func (h *BaseHook) GetTimeout() time.Duration {
	return h.config.Timeout
}

// HookManager manages lifecycle hooks for a service
type HookManager struct {
	hooks  map[HookPhase][]LifecycleHook
	config HookConfiguration
	mu     sync.RWMutex
}

// NewHookManager creates a new hook manager
func NewHookManager(config HookConfiguration) *HookManager {
	return &HookManager{
		hooks:  make(map[HookPhase][]LifecycleHook),
		config: config,
	}
}

// AddHook adds a lifecycle hook for the specified phase
func (hm *HookManager) AddHook(phase HookPhase, hook LifecycleHook) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.hooks[phase] = append(hm.hooks[phase], hook)
	return nil
}

// ExecuteHooks executes all hooks for the specified phase
func (hm *HookManager) ExecuteHooks(ctx context.Context, phase HookPhase, service ServiceLifecycleInterface) error {
	hm.mu.RLock()
	hooks := hm.hooks[phase]
	hm.mu.RUnlock()

	for i, hook := range hooks {
		if err := hm.executeHook(ctx, phase, hook, service); err != nil {
			// Handle failure based on policy
			switch hm.config.FailurePolicy {
			case FailFast:
				return fmt.Errorf("hook %d in phase %s failed: %w", i, phase.String(), err)
			case FailOnCritical:
				if hook.IsCritical() {
					return fmt.Errorf("critical hook %d in phase %s failed: %w", i, phase.String(), err)
				}
				// Continue for non-critical hooks
			case Continue:
				// Log error but continue
				fmt.Printf("Warning: hook %d in phase %s failed: %v\n", i, phase.String(), err)
			}
		}
	}

	return nil
}

// executeHook executes a single hook with timeout handling
func (hm *HookManager) executeHook(ctx context.Context, phase HookPhase, hook LifecycleHook, service ServiceLifecycleInterface) error {
	// Create context with timeout
	timeout := hook.GetTimeout()
	if timeout == 0 {
		timeout = hm.config.Timeout
	}

	hookCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute hook based on phase
	switch phase {
	case BeforeStartPhase:
		return hook.BeforeStart(hookCtx, service)
	case AfterStartPhase:
		return hook.AfterStart(hookCtx, service)
	case BeforeStopPhase:
		return hook.BeforeStop(hookCtx, service)
	case AfterStopPhase:
		return hook.AfterStop(hookCtx, service)
	default:
		return fmt.Errorf("unknown hook phase: %v", phase)
	}
}

// GetHooksCount returns the number of hooks registered for each phase
func (hm *HookManager) GetHooksCount() map[HookPhase]int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	counts := make(map[HookPhase]int)
	for phase, hooks := range hm.hooks {
		counts[phase] = len(hooks)
	}
	return counts
}
