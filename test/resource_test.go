package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

func TestNewResourceDefinitionFromYAML(t *testing.T) {
	yamlInput := `type: object
properties:
  name:
    type: string
  surname:
    type: string`
	resource := sdk.Resource{
		ID:                   "customer",
		Description:          "Customers",
		Service:              "",
		AdditionalAttributes: yamlInput,
	}
	def, err := resource.UnmarshalAdditionalAttributes()
	if err != nil {
		t.Fatalf("Error parsing definition: %v", err)
	}

	if def.Schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got %q", def.Schema.Type)
	}

	properties := *def.Schema.Properties
	if _, ok := properties["name"]; !ok {
		t.Errorf("Schema properties missing 'name'")
	}
}
