package sdk_entity_test

import (
	"testing"

	examples_handlers "github.com/mattiabonardi/endor-sdk-go/internal/examples/handlers"
	"github.com/stretchr/testify/assert"
)

func TestEndorBaseHandler(t *testing.T) {
	baseHandler := examples_handlers.NewBaseEntityHandler()
	endorHandler := baseHandler.ToEndorHandler()

	// check attribute
	assert.Equal(t, endorHandler.Entity, "base-entity")
	assert.Equal(t, endorHandler.EntityTitle, "${t.examples.base-entity.handler.title}")
	// check schema
	assert.Equal(t, len(*endorHandler.EntitySchema.Properties), 2)

	// check methods
	assert.Len(t, endorHandler.Actions, 2, "Expected 2 actions")
	_, actionExists := endorHandler.Actions["action1"]
	assert.True(t, actionExists, "method 'action1' not found in endorHandler methods map")
	_, publicActionExists := endorHandler.Actions["public-action"]
	assert.True(t, publicActionExists, "method 'public-action' not found in endorHandler methods map")
}
