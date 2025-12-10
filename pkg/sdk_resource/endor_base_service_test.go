package sdk_resource_test

import (
	"testing"

	test_utils_services "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/services"
	"github.com/stretchr/testify/assert"
)

func TestEndorBaseService(t *testing.T) {
	baseService := test_utils_services.NewBaseServiceService()
	endorService := baseService.ToEndorService()

	// check attribute
	assert.Equal(t, endorService.Resource, "base-service")
	assert.Equal(t, endorService.ResourceDescription, "Base Service (EndorBaseService)")
	// check schema
	assert.Equal(t, len(*endorService.ResourceSchema.Properties), 2)

	// check method
	_, actionExists := endorService.Actions["action-1"]
	assert.True(t, actionExists, "method 'action-1' not found in endorService methods map")
}
