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
	// check definitions
	if len(def.Components.Schemas) != 6 {
		t.Fatalf("Received %v", len(def.Components.Schemas))
	}
	// check paths
	if len(def.Paths) != 5 {
		t.Fatalf("Received %v", len(def.Paths))
	}
	// check generics
	if def.Paths["/api/v1/test/test4"]["post"].RequestBody.Content["application/json"].Schema.Reference != "#/components/schemas/Test4Payload_GenericPayload" {
		t.Fatalf("Received %v", def.Paths["/api/v1/test/test4"]["post"].RequestBody.Content["application/json"].Schema.Reference)
	}
	test4Payload := def.Components.Schemas["Test4Payload_GenericPayload"]
	test4PayloadProperties := *test4Payload.Properties
	if test4PayloadProperties["value"].Reference != "#/components/schemas/GenericPayload" {
		t.Fatalf("Received %v", test4PayloadProperties["value"].Reference)
	}
}
