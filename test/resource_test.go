package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

func TestNewResourceDefinitionFromYAML(t *testing.T) {
	yamlInput := `
schema:
  type: object
  properties:
    name:
      type: string
    surname:
      type: string
dataSources:
  - name: main
    type: mongodb
    collection: customers
    mappings:
      name:
        path: name
      surname:
        path: surname
id: name
`
	resource := sdk.Resource{
		ID:          "customer",
		Description: "Customers",
		Service:     "",
		Definition:  yamlInput,
	}
	def, err := resource.UnmarshalDefinition()
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

	if len(def.DataSources) != 1 {
		t.Fatalf("Expected 1 data source, got %d", len(def.DataSources))
	}

	ds := def.DataSources[0]
	if ds.GetType() != "mongodb" {
		t.Errorf("Expected data source type 'mongodb', got %q", ds.GetType())
	}

	if mongoDS, ok := ds.(*sdk.MongoDataSource); ok {
		if mongoDS.Collection != "customers" {
			t.Errorf("Expected collection 'customers', got %q", mongoDS.Collection)
		}
		if path, ok := mongoDS.Mappings["name"]; !ok || path.Path != "name" {
			t.Errorf("Expected mapping for 'name' to path 'name', got %v", path)
		}
	} else {
		t.Errorf("DataSource is not MongoDataSource")
	}

	if def.Id != "name" {
		t.Errorf("Expected id 'name', got %s", def.Id)
	}
}
