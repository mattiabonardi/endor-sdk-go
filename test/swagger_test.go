package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	services_test "github.com/mattiabonardi/endor-sdk-go/test/services"
)

func TestCreateSwaggerDefinition(t *testing.T) {
	def, err := sdk.CreateSwaggerDefinition("endor-sdk-service", "endorsdkservice.com", []sdk.EndorService{services_test.NewService1()}, "/api")
	if err != nil {
		t.Fail()
	}
	if def.OpenAPI != "3.1.0" {
		t.Fatalf("Received %v", def.OpenAPI)
	}
	if def.Info.Title != "endor-sdk-service" {
		t.Fatalf("Received %v", def.Info.Title)
	}
	if def.Info.Description != "endor-sdk-service docs" {
		t.Fatalf("Received %v", def.Info.Description)
	}
	if def.Servers[0].URL != "/" {
		t.Fatalf("Received %v", def.Servers[0].URL)
	}
	// endor resources
	if def.Tags[0].Description != "Testing resource" {
		t.Fatalf("Received %v", def.Tags[0].Description)
	}
	// check definitions - with expanded schemas, should be minimal (only default response)
	if len(def.Components.Schemas) != 1 {
		t.Fatalf("Expected 1 schema component (default response), got %d", len(def.Components.Schemas))
	}
	// check paths
	if len(def.Paths) != 5 {
		t.Fatalf("Received %v", len(def.Paths))
	}
	// check generics - with expanded schemas, should be inline objects, not references
	test4RequestSchema := def.Paths["/api/v1/test/test4"]["post"].RequestBody.Content["application/json"].Schema
	if test4RequestSchema.Reference != "" {
		t.Fatalf("Expected inline schema, not reference, got %v", test4RequestSchema.Reference)
	}
	if test4RequestSchema.Type != sdk.ObjectType {
		t.Fatalf("Expected object type, got %v", test4RequestSchema.Type)
	}
	if test4RequestSchema.Properties == nil {
		t.Fatalf("Expected properties to be present in inline schema")
	}

	test4Properties := *test4RequestSchema.Properties
	if valueProp := test4Properties["value"]; valueProp.Type != sdk.ObjectType {
		t.Fatalf("Expected value property to be object type, got %v", valueProp.Type)
	}
}

func TestCreateSwaggerDefinitionWithCategories(t *testing.T) {
	def, err := sdk.CreateSwaggerDefinition("endor-sdk-service", "endorsdkservice.com", []sdk.EndorService{services_test.NewService3()}, "/api")
	if err != nil {
		t.Fail()
	}
	if _, ok := def.Paths["/api/v1/resource-3/action1"]; !ok {
		t.Fatalf("Expected /api/v1/resource-3/action1 to exist")
	}
	if _, ok := def.Paths["/api/v1/resource-3/cat_1/action1"]; !ok {
		t.Fatalf("Expected /api/v1/resource-3/cat_1/action1 to exist")
	}
}
