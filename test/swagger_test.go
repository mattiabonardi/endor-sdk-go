package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	services_test "github.com/mattiabonardi/endor-sdk-go/test/services"
)

func TestCreateSwaggerDefinition(t *testing.T) {
	def, err := sdk.CreateSwaggerDefinition("endor-sdk-service", "endorsdkservice.com", []sdk.EndorService{services_test.NewService1()}, "/api/:app")
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
	// security scheme
	if len(def.Components.SecuritySchemas) != 1 {
		t.Fatalf("Received %v", len(def.Components.SecuritySchemas))
	}
	// check definitions
	if len(def.Components.Schemas) != 4 {
		t.Fatalf("Received %v", len(def.Components.Schemas))
	}
	// check paths
	if len(def.Paths) != 3 {
		t.Fatalf("Received %v", len(def.Paths))
	}
}

func TestAdaptSwaggerSchemaToSchema(t *testing.T) {
	def, err := sdk.CreateSwaggerDefinition("endor-sdk-service", "endorsdkservice.com", []sdk.EndorService{services_test.NewService1()}, "/api/:app")
	if err != nil {
		t.Fail()
	}
	// test payload 2
	path := def.Paths["/api/{app}/v1/test/test2"].Post.RequestBody.Content["application/json"].Schema
	payload1 := sdk.AdaptSwaggerSchemaToSchema(def.Components, &path)
	if len(payload1.Definitions) != 2 {
		t.Fatalf("Received %v", len(payload1.Definitions))
	}
	if payload1.Reference != "#/$defs/Test2Payload" {
		t.Fatalf("Received %v", payload1.Reference)
	}
	if payload1.Definitions["Test2Payload"].Type != sdk.ObjectType {
		t.Fatalf("Received %v", payload1.Type)
	}
	prop := *payload1.Definitions["Test2Payload"].Properties
	if prop["objectArray"].Items.Reference != "#/$defs/Test2PayloadArrayIteam" {
		t.Fatalf("Received %v", prop["objectArray"].Items.Reference)
	}
}
