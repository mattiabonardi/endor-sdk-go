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
	if len(def.Components.Schemas) != 3 {
		t.Fatalf("Received %v", len(def.Components.Schemas))
	}
	// check paths
	if len(def.Paths) != 3 {
		t.Fatalf("Received %v", len(def.Paths))
	}
}
