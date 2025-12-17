package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntityDefinitionFromYAML(t *testing.T) {
	yamlInput := `type: object
properties:
  name:
    type: string
  surname:
    type: string`
	entity := sdk.EntityHybrid{
		Entity: sdk.Entity{
			ID:          "customer",
			Description: "Customers",
			Service:     "",
		},
		AdditionalAttributes: yamlInput,
	}
	def, err := entity.UnmarshalAdditionalAttributes()
	require.NoError(t, err, "Error parsing definition")

	assert.Equal(t, sdk.SchemaTypeObject, def.Schema.Type, "Expected schema type 'object'")

	properties := *def.Schema.Properties
	assert.Contains(t, properties, "name", "Schema properties missing 'name'")
}
