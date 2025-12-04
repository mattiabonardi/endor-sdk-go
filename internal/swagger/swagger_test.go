package swagger

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	services_test "github.com/mattiabonardi/endor-sdk-go/test/services"
)

func TestCreateSwaggerDefinition(t *testing.T) {
	def, err := CreateSwaggerDefinition("endor-sdk-service", "endorsdkservice.com", []sdk.EndorService{services_test.NewService1()}, "/api")
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
	if def.Tags[0].Description != "Resource 1 (EndorService with categories)" {
		t.Fatalf("Received %v", def.Tags[0].Description)
	}
	// check paths
	if len(def.Paths) != 2 {
		t.Fatalf("Received %v", len(def.Paths))
	}
	if _, ok := def.Paths["/api/v1/resource-1/action1"]; !ok {
		t.Fatalf("Expected /api/v1/resource-1/action1 to exist")
	}
	if _, ok := def.Paths["/api/v1/resource-1/cat_1/action1"]; !ok {
		t.Fatalf("Expected /api/v1/resource-1/cat_1/action1 to exist")
	}
}
