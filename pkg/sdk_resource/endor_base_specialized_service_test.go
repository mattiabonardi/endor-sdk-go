package sdk_resource_test

import (
	"testing"

	test_utils_services "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/services"
	"github.com/stretchr/testify/assert"
)

func TestEndorBaseSpecializedService(t *testing.T) {
	baseService := test_utils_services.NewBaseSpecializedService()
	endorService := baseService.ToEndorService()

	// check attribute
	assert.Equal(t, endorService.Resource, "base-specialized-service")
	assert.Equal(t, endorService.ResourceDescription, "Base Specialized Service (EndorBaseSpecializedService)")
	// check schema
	assert.Equal(t, 3, len(*endorService.ResourceSchema.Properties))

	// check method
	_, actionExists := endorService.Actions["action-1"]
	assert.True(t, actionExists, "method 'action-1' not found in endorService methods map")
	// check method category 1
	_, category1ActionExists := endorService.Actions["cat-1/action-1"]
	assert.True(t, category1ActionExists, "method 'action-1' not found in endorService methods map for category 1")
	// check method category 2
	_, category2ActionExists := endorService.Actions["cat-2/action-1"]
	assert.True(t, category2ActionExists, "method 'action-1' not found in endorService methods map for category 2")
}
