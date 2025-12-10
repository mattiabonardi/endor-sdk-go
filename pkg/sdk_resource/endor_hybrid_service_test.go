package sdk_resource_test

import (
	"testing"

	test_utils_services "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/services"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

type AdditionalAttributesMock struct {
	AdditionalAttribute string `json:"additionalAttribute"`
}

func TestEndorHybridService(t *testing.T) {
	hybridService := test_utils_services.NewHybridService()
	endorService := hybridService.ToEndorService(
		sdk.NewSchema(AdditionalAttributesMock{}).Schema,
	)

	// check attribute
	assert.Equal(t, endorService.Resource, "hybrid-service")
	assert.Equal(t, endorService.ResourceDescription, "Hybrid Service (EndorHybridService)")
	// check schema
	assert.Equal(t, len(*endorService.ResourceSchema.Properties), 3)

	// check default methods
	_, schemaExists := endorService.Actions["schema"]
	assert.True(t, schemaExists, "method 'schema' not found in endorService methods map")
	_, instanceExists := endorService.Actions["instance"]
	assert.True(t, instanceExists, "method 'instance' not found in endorService methods map")
	_, idPropertyExists := (*endorService.Actions["instance"].GetOptions().InputSchema.Properties)["id"]
	assert.True(t, idPropertyExists, "'id' property not found in input schema for method 'instance'")
	_, listExists := endorService.Actions["list"]
	assert.True(t, listExists, "method 'list' not found in endorService methods map")
	_, createExists := endorService.Actions["create"]
	assert.True(t, createExists, "method 'create' not found in endorService methods map")
	if dataSchema, ok := (*endorService.Actions["create"].GetOptions().InputSchema.Properties)["data"]; ok {
		_, idExists := (*dataSchema.Properties)["id"]
		assert.True(t, idExists, "input schema for method 'create' missing 'id'")
		_, attributeExists := (*dataSchema.Properties)["attribute"]
		assert.True(t, attributeExists, "input schema for method 'create' missing 'attribute'")
		_, additionalAttributeExists := (*dataSchema.Properties)["additionalAttribute"]
		assert.True(t, additionalAttributeExists, "input schema for method 'create' missing 'additionalAttribute'")
	} else {
		assert.Fail(t, "'data' property not found in input schema for method 'create'")
	}
	_, updateExists := endorService.Actions["update"]
	assert.True(t, updateExists, "method 'update' not found in endorService methods map")
	if dataSchema, ok := (*endorService.Actions["update"].GetOptions().InputSchema.Properties)["data"]; ok {
		_, idExists := (*dataSchema.Properties)["id"]
		assert.True(t, idExists, "input schema for method 'update' missing 'id'")
		_, attributeExists := (*dataSchema.Properties)["attribute"]
		assert.True(t, attributeExists, "input schema for method 'update' missing 'attribute'")
		_, additionalAttributeExists := (*dataSchema.Properties)["additionalAttribute"]
		assert.True(t, additionalAttributeExists, "input schema for method 'update' missing 'additionalAttribute'")
	} else {
		assert.Fail(t, "'data' property not found in input schema for method 'update'")
	}
	_, updateIdExists := (*endorService.Actions["update"].GetOptions().InputSchema.Properties)["id"]
	assert.True(t, updateIdExists, "'id' property not found in input schema for method 'update'")
	_, deleteExists := endorService.Actions["delete"]
	assert.True(t, deleteExists, "method 'delete' not found in endorService methods map")
	_, deleteIdExists := (*endorService.Actions["delete"].GetOptions().InputSchema.Properties)["id"]
	assert.True(t, deleteIdExists, "'id' property not found in input schema for method 'delete'")
	_, action1Exists := endorService.Actions["action-1"]
	assert.True(t, action1Exists, "method 'action-1' not found in endorService methods map")
}
