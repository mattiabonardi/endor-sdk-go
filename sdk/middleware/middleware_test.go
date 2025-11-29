package middleware

import (
	"errors"
	"testing"
	"time"
)

// mockMiddleware implements MiddlewareInterface for testing
type mockMiddleware struct {
	name           string
	beforeError    error
	afterError     error
	beforeCalled   bool
	afterCalled    bool
	beforeDuration time.Duration
	afterDuration  time.Duration
}

func (m *mockMiddleware) Before(ctx interface{}) error {
	m.beforeCalled = true
	if m.beforeDuration > 0 {
		time.Sleep(m.beforeDuration)
	}
	return m.beforeError
}

func (m *mockMiddleware) After(ctx interface{}, response interface{}) error {
	m.afterCalled = true
	if m.afterDuration > 0 {
		time.Sleep(m.afterDuration)
	}
	return m.afterError
}

func TestMiddlewarePipeline_ExecuteBefore_Success(t *testing.T) {
	// Create mock middleware
	middleware1 := &mockMiddleware{name: "middleware1"}
	middleware2 := &mockMiddleware{name: "middleware2"}

	pipeline := NewMiddlewarePipeline(middleware1, middleware2)

	// Execute Before hooks
	err := pipeline.ExecuteBefore("test-context")

	// Verify no error
	if err != nil {
		t.Errorf("ExecuteBefore() failed: %v", err)
	}

	// Verify both middleware were called
	if !middleware1.beforeCalled {
		t.Error("middleware1.Before() was not called")
	}
	if !middleware2.beforeCalled {
		t.Error("middleware2.Before() was not called")
	}

	// Verify execution tracking
	executions := pipeline.GetExecutions()
	if len(executions) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(executions))
	}

	for i, exec := range executions {
		if !exec.Success {
			t.Errorf("Execution %d should be successful", i)
		}
		if exec.Error != nil {
			t.Errorf("Execution %d should not have error, got: %v", i, exec.Error)
		}
	}
}

func TestMiddlewarePipeline_ExecuteBefore_ShortCircuit(t *testing.T) {
	// Create middleware where first one fails
	middleware1 := &mockMiddleware{name: "middleware1", beforeError: errors.New("auth failed")}
	middleware2 := &mockMiddleware{name: "middleware2"}

	pipeline := NewMiddlewarePipeline(middleware1, middleware2)

	// Execute Before hooks
	err := pipeline.ExecuteBefore("test-context")

	// Verify error is returned
	if err == nil {
		t.Error("ExecuteBefore() should have returned an error")
	}

	// Verify first middleware was called
	if !middleware1.beforeCalled {
		t.Error("middleware1.Before() was not called")
	}

	// Verify second middleware was NOT called (short-circuit)
	if middleware2.beforeCalled {
		t.Error("middleware2.Before() should not have been called due to short-circuit")
	}

	// Verify execution tracking shows the failure
	executions := pipeline.GetExecutions()
	if len(executions) != 1 {
		t.Errorf("Expected 1 execution, got %d", len(executions))
	}

	if executions[0].Success {
		t.Error("First execution should be marked as failed")
	}
	if executions[0].Error == nil {
		t.Error("First execution should have error recorded")
	}
}
