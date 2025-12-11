package sdk_resource_test

import (
	"testing"

	test_utils_services "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/services"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

type HybridCategory1AdditionalSchema struct {
	AdditionalAttributeCat1 string `json:"additionalAttributeCat1"`
}

type HybridCategory2AdditionalSchema struct {
	AdditionalAttributeCat2 string `json:"additionalAttributeCat2"`
}

func TestEndorHybridSpecializedService(t *testing.T) {
	category1AdditionalSchema := sdk.NewSchema(HybridCategory1AdditionalSchema{})
	category2AdditionalSchema := sdk.NewSchema(HybridCategory2AdditionalSchema{})

	hybridService := test_utils_services.NewHybridSpecializedService()
	endorService := hybridService.ToEndorService(
		sdk.NewSchema(AdditionalAttributesMock{}).Schema,
		map[string]sdk.Schema{
			"cat-1": category1AdditionalSchema.Schema,
			"cat-2": category2AdditionalSchema.Schema,
		},
	)

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
	assert.False(t, createExists, "method 'create' defined in endorService methods map")
	_, updateExists := endorService.Actions["update"]
	assert.False(t, updateExists, "method 'update' defined in endorService methods map")
	_, action1Exists := endorService.Actions["action-1"]
	assert.True(t, action1Exists, "method 'action-1' not found in endorService methods map")
	// categories
	// check categories default methods (cat-1)
	_, cat1SchemaExists := endorService.Actions["cat-1/schema"]
	assert.True(t, cat1SchemaExists, "method 'cat-1/schema' not found in endorService methods map")
	_, cat1InstanceExists := endorService.Actions["cat-1/instance"]
	assert.True(t, cat1InstanceExists, "method 'cat-1/instance' not found in endorService methods map")
	_, cat1InstanceIdExists := (*endorService.Actions["cat-1/instance"].GetOptions().InputSchema.Properties)["id"]
	assert.True(t, cat1InstanceIdExists, "'id' property not found in input schema for method 'cat-1/instance'")
	_, cat1ListExists := endorService.Actions["cat-1/list"]
	assert.True(t, cat1ListExists, "method 'cat-1/list' not found in endorService methods map")
	_, cat1CreateExists := endorService.Actions["cat-1/create"]
	assert.True(t, cat1CreateExists, "method 'cat-1/create' not found in endorService methods map")
	if dataSchema, ok := (*endorService.Actions["cat-1/create"].GetOptions().InputSchema.Properties)["data"]; ok {
		_, idExists := (*dataSchema.Properties)["id"]
		assert.True(t, idExists, "input schema for method 'cat-1/create' missing 'id'")
		_, typeExists := (*dataSchema.Properties)["type"]
		assert.True(t, typeExists, "input schema for method 'create' missing 'type'")
		_, attributeExists := (*dataSchema.Properties)["attribute"]
		assert.True(t, attributeExists, "input schema for method 'cat-1/create' missing 'attribute'")
		_, additionalAttributeExists := (*dataSchema.Properties)["additionalAttribute"]
		assert.True(t, additionalAttributeExists, "input schema for method 'cat-1/create' missing 'additionalAttribute'")
		_, additionalAttributeCat1Exists := (*dataSchema.Properties)["additionalAttributeCat1"]
		assert.True(t, additionalAttributeCat1Exists, "input schema for method 'cat-1/create' missing 'additionalAttributeCat1'")
	} else {
		assert.Fail(t, "'data' property not found in input schema for method 'cat-1/create'")
	}
	if dataSchema, ok := (*endorService.Actions["cat-1/update"].GetOptions().InputSchema.Properties)["data"]; ok {
		_, idExists := (*dataSchema.Properties)["id"]
		assert.True(t, idExists, "input schema for method 'cat-1/update' missing 'id'")
		_, typeExists := (*dataSchema.Properties)["type"]
		assert.True(t, typeExists, "input schema for method 'create' missing 'type'")
		_, attributeExists := (*dataSchema.Properties)["attribute"]
		assert.True(t, attributeExists, "input schema for method 'cat-1/update' missing 'attribute'")
		_, additionalAttributeExists := (*dataSchema.Properties)["additionalAttribute"]
		assert.True(t, additionalAttributeExists, "input schema for method 'cat-1/update' missing 'additionalAttribute'")
		_, additionalAttributeCat1Exists := (*dataSchema.Properties)["additionalAttributeCat1"]
		assert.True(t, additionalAttributeCat1Exists, "input schema for method 'cat-1/update' missing 'additionalAttributeCat1'")
	} else {
		assert.Fail(t, "'data' property not found in input schema for method 'cat-1/update'")
	}
	_, cat1UpdateIdExists := (*endorService.Actions["cat-1/update"].GetOptions().InputSchema.Properties)["id"]
	assert.True(t, cat1UpdateIdExists, "'id' property not found in input schema for method 'cat-1/update'")
	// check categories default methods (cat-2)
	_, cat2CreateExists := endorService.Actions["cat-2/create"]
	assert.True(t, cat2CreateExists, "method 'cat-2/create' not found in endorService methods map")
	if dataSchema, ok := (*endorService.Actions["cat-2/create"].GetOptions().InputSchema.Properties)["data"]; ok {
		_, idExists := (*dataSchema.Properties)["id"]
		assert.True(t, idExists, "input schema for method 'cat-2/create' missing 'id'")
		_, typeExists := (*dataSchema.Properties)["type"]
		assert.True(t, typeExists, "input schema for method 'create' missing 'type'")
		_, attributeExists := (*dataSchema.Properties)["attribute"]
		assert.True(t, attributeExists, "input schema for method 'cat-2/create' missing 'attribute'")
		_, additionalAttributeExists := (*dataSchema.Properties)["additionalAttribute"]
		assert.True(t, additionalAttributeExists, "input schema for method 'cat-2/create' missing 'additionalAttribute'")
		_, attributeCat2Exists := (*dataSchema.Properties)["attributeCat2"]
		assert.True(t, attributeCat2Exists, "input schema for method 'cat-2/create' missing 'attributeCat2'")
		_, additionalAttributeCat2Exists := (*dataSchema.Properties)["additionalAttributeCat2"]
		assert.True(t, additionalAttributeCat2Exists, "input schema for method 'cat-2/create' missing 'additionalAttributeCat2'")
	} else {
		assert.Fail(t, "'data' property not found in input schema for method 'cat-2/create'")
	}
}
