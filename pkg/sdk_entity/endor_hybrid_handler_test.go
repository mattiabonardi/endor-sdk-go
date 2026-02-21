package sdk_entity_test

import (
	"testing"

	test_utils_handlers "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/handlers"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

type AdditionalAttributesMock struct {
	AdditionalAttribute string `json:"additionalAttribute"`
}

func TestEndorHybridHandler(t *testing.T) {
	hybridHandler := test_utils_handlers.NewHybridHandler()
	endorHandler := hybridHandler.ToEndorHandler(
		*sdk.NewSchema(AdditionalAttributesMock{}),
	)

	// check attribute
	assert.Equal(t, endorHandler.Entity, "hybrid-service")
	assert.Equal(t, endorHandler.EntityDescription, "Hybrid Handler (EndorHybridHandler)")
	// check schema
	assert.Equal(t, len(*endorHandler.EntitySchema.Properties), 3)

	// check default methods
	_, schemaExists := endorHandler.Actions["schema"]
	assert.True(t, schemaExists, "method 'schema' not found in endorHandler methods map")
	_, instanceExists := endorHandler.Actions["instance"]
	assert.True(t, instanceExists, "method 'instance' not found in endorHandler methods map")
	_, idPropertyExists := (*endorHandler.Actions["instance"].GetOptions().InputSchema.Properties)["id"]
	assert.True(t, idPropertyExists, "'id' property not found in input schema for method 'instance'")
	_, listExists := endorHandler.Actions["list"]
	assert.True(t, listExists, "method 'list' not found in endorHandler methods map")
	_, createExists := endorHandler.Actions["create"]
	assert.True(t, createExists, "method 'create' not found in endorHandler methods map")
	if dataSchema, ok := (*endorHandler.Actions["create"].GetOptions().InputSchema.Properties)["data"]; ok {
		_, idExists := (*dataSchema.Properties)["id"]
		assert.True(t, idExists, "input schema for method 'create' missing 'id'")
		_, attributeExists := (*dataSchema.Properties)["attribute"]
		assert.True(t, attributeExists, "input schema for method 'create' missing 'attribute'")
		_, additionalAttributeExists := (*dataSchema.Properties)["additionalAttribute"]
		assert.True(t, additionalAttributeExists, "input schema for method 'create' missing 'additionalAttribute'")
	} else {
		assert.Fail(t, "'data' property not found in input schema for method 'create'")
	}
	_, updateExists := endorHandler.Actions["update"]
	assert.True(t, updateExists, "method 'update' not found in endorHandler methods map")
	if dataSchema, ok := (*endorHandler.Actions["update"].GetOptions().InputSchema.Properties)["data"]; ok {
		_, idExists := (*dataSchema.Properties)["id"]
		assert.True(t, idExists, "input schema for method 'update' missing 'id'")
		_, attributeExists := (*dataSchema.Properties)["attribute"]
		assert.True(t, attributeExists, "input schema for method 'update' missing 'attribute'")
		_, additionalAttributeExists := (*dataSchema.Properties)["additionalAttribute"]
		assert.True(t, additionalAttributeExists, "input schema for method 'update' missing 'additionalAttribute'")
	} else {
		assert.Fail(t, "'data' property not found in input schema for method 'update'")
	}
	_, updateIdExists := (*endorHandler.Actions["update"].GetOptions().InputSchema.Properties)["id"]
	assert.True(t, updateIdExists, "'id' property not found in input schema for method 'update'")
	_, deleteExists := endorHandler.Actions["delete"]
	assert.True(t, deleteExists, "method 'delete' not found in endorHandler methods map")
	_, deleteIdExists := (*endorHandler.Actions["delete"].GetOptions().InputSchema.Properties)["id"]
	assert.True(t, deleteIdExists, "'id' property not found in input schema for method 'delete'")
	_, action1Exists := endorHandler.Actions["action-1"]
	assert.True(t, action1Exists, "method 'action-1' not found in endorHandler methods map")
}
