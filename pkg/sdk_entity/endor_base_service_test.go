package sdk_entity_test

import (
	"testing"

	test_utils_services "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/services"
	"github.com/stretchr/testify/assert"
)

func TestEndorBaseService(t *testing.T) {
	baseService := test_utils_services.NewBaseServiceService()
	endorService := baseService.ToEndorService()

	// check attribute
	assert.Equal(t, endorService.Entity, "base-service")
	assert.Equal(t, endorService.EntityDescription, "Base Service (EndorBaseService)")
	// check schema
	assert.Equal(t, len(*endorService.EntitySchema.Properties), 2)

	// check methods
	assert.Len(t, endorService.Actions, 3, "Expected 3 actions")
	_, actionExists := endorService.Actions["action1"]
	assert.True(t, actionExists, "method 'action1' not found in endorService methods map")
	_, categoryActionExists := endorService.Actions["cat_1/action1"]
	assert.True(t, categoryActionExists, "method 'cat_1/action1' not found in endorService methods map")
	_, publicActionExists := endorService.Actions["public-action"]
	assert.True(t, publicActionExists, "method 'public-action' not found in endorService methods map")
}
