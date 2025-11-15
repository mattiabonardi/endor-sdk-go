package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

func TestNewResourceDefinitionFromYAML(t *testing.T) {
	yamlInput := `name:
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

type TestRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (t *TestRes) GetID() *string  { return &t.ID }
func (t *TestRes) SetID(id string) { t.ID = id }

func TestResourceInstance_ToSchema_MetadataFlat(t *testing.T) {
	// Metadata schema
	metaProps := map[string]sdk.Schema{
		"meta1": {Type: sdk.StringType},
		"meta2": {Type: sdk.StringType},
	}
	metadataSchema := &sdk.Schema{
		Type:       sdk.ObjectType,
		Properties: &metaProps,
	}

	inst := &sdk.ResourceInstance[*TestRes]{This: &TestRes{ID: "1", Name: "foo"}}
	rootSchema := inst.ToSchema(metadataSchema)

	if rootSchema.Schema.Properties == nil {
		t.Fatal("Properties should not be nil")
	}
	props := *rootSchema.Schema.Properties

	if _, ok := props["meta1"]; !ok {
		t.Error("meta1 should be at root level")
	}
	if _, ok := props["meta2"]; !ok {
		t.Error("meta2 should be at root level")
	}
	if _, ok := props["metadata"]; ok {
		t.Error("metadata should not be a root property")
	}
}
