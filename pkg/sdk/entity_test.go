package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestMergeSchemas(t *testing.T) {
	baseYAML := `type: object
properties:
  id:
    type: string`
	additionalYAML := `type: object
properties:
  name:
    type: string
  surname:
    type: string`

	merged := sdk.MergeSchemas(baseYAML, additionalYAML)

	assert.Contains(t, merged, "id", "merged schema should retain 'id' property")
	assert.Contains(t, merged, "name", "merged schema should contain 'name' property")
	assert.Contains(t, merged, "surname", "merged schema should contain 'surname' property")
}
