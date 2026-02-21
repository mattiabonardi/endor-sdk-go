package sdk_entity_test

import (
	"testing"

	test_utils_handlers "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/handlers"
	"github.com/stretchr/testify/assert"
)

func TestEndorBaseSpecializedHandler(t *testing.T) {
	baseSpecializedHandler := test_utils_handlers.NewBaseSpecializedHandler()
	endorHandler := baseSpecializedHandler.ToEndorHandler()

	// check attribute
	assert.Equal(t, endorHandler.Entity, "base-specialized-handler")
	assert.Equal(t, endorHandler.EntityDescription, "Base Specialized Handler (EndorBaseSpecializedHandler)")
	// check schema
	assert.Equal(t, 3, len(*endorHandler.EntitySchema.Properties))

	// check method
	_, actionExists := endorHandler.Actions["action-1"]
	assert.True(t, actionExists, "method 'action-1' not found in endorHandler methods map")
	// check method category 1
	_, category1ActionExists := endorHandler.Actions["cat-1/action-1"]
	assert.True(t, category1ActionExists, "method 'action-1' not found in endorHandler methods map for category 1")
	// check method category 2
	_, category2ActionExists := endorHandler.Actions["cat-2/action-1"]
	assert.True(t, category2ActionExists, "method 'action-1' not found in endorHandler methods map for category 2")
}
