package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleLifecycleManager(t *testing.T) {
	t.Run("BasicInstantiation", func(t *testing.T) {
		lm := NewLifecycleManager()
		assert.NotNil(t, lm)
	})

	t.Run("RegisterWithoutListeners", func(t *testing.T) {
		lm := NewLifecycleManager()
		service := NewMockLifecycleService("simple-test")

		// Don't add any listeners to avoid goroutine issues
		lm.RegisterDependency("simple-test", service)

		state, exists := lm.GetLifecycleState("simple-test")
		assert.True(t, exists)
		assert.Equal(t, LifecycleStateRegistered, state)
	})
}
