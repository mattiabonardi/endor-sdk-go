package sdk

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	services_test "github.com/mattiabonardi/endor-sdk-go/test/services"
)

type AdditionalAttributesMock struct {
	AdditionalAttribute string `json:"additionalAttribute"`
}

func TestEndorHybridService(t *testing.T) {

	hybridService := services_test.NewService2()
	endorService := hybridService.ToEndorService(
		sdk.NewSchema(AdditionalAttributesMock{}).Schema,
	)

	// check default methods
	if _, ok := endorService.Methods["schema"]; !ok {
		t.Errorf("method 'schema' not found in endorService methods map")
	}
	if _, ok := endorService.Methods["instance"]; !ok {
		t.Errorf("method 'instance' not found in endorService methods map")
	}
	if _, ok := (*endorService.Methods["instance"].GetOptions().InputSchema.Properties)["id"]; !ok {
		t.Errorf("'id' property not found in input schema for method 'instance'")
	}
	if _, ok := endorService.Methods["list"]; !ok {
		t.Errorf("method 'list' not found in endorService methods map")
	}
	if _, ok := endorService.Methods["create"]; !ok {
		t.Errorf("method 'create' not found in endorService methods map")
	}
	if dataSchema, ok := (*endorService.Methods["create"].GetOptions().InputSchema.Properties)["data"]; ok {
		if _, ok := (*dataSchema.Properties)["id"]; !ok {
			t.Errorf("input schema for method 'create' wrong")
		}
		if _, ok := (*dataSchema.Properties)["attribute"]; !ok {
			t.Errorf("input schema for method 'create' wrong")
		}
		if _, ok := (*dataSchema.Properties)["additionalAttribute"]; !ok {
			t.Errorf("input schema for method 'create' wrong")
		}
	} else {
		t.Errorf("'data' property not found in input schema for method 'create'")
	}
	if _, ok := endorService.Methods["update"]; !ok {
		t.Errorf("method 'update' not found in endorService methods map")
	}
	if dataSchema, ok := (*endorService.Methods["update"].GetOptions().InputSchema.Properties)["data"]; ok {
		if _, ok := (*dataSchema.Properties)["id"]; !ok {
			t.Errorf("input schema for method 'update' wrong")
		}
		if _, ok := (*dataSchema.Properties)["attribute"]; !ok {
			t.Errorf("input schema for method 'update' wrong")
		}
		if _, ok := (*dataSchema.Properties)["additionalAttribute"]; !ok {
			t.Errorf("input schema for method 'update' wrong")
		}
	} else {
		t.Errorf("'data' property not found in input schema for method 'create'")
	}
	if _, ok := (*endorService.Methods["update"].GetOptions().InputSchema.Properties)["id"]; !ok {
		t.Errorf("'id' property not found in input schema for method 'update'")
	}
	if _, ok := endorService.Methods["delete"]; !ok {
		t.Errorf("method 'delete' not found in endorService methods map")
	}
	if _, ok := (*endorService.Methods["delete"].GetOptions().InputSchema.Properties)["id"]; !ok {
		t.Errorf("'id' property not found in input schema for method 'delete'")
	}
	// check additional method
	if _, ok := endorService.Methods["action-1"]; !ok {
		t.Errorf("method 'action-1' not found in endorService methods map")
	}
	// categories
	// check categories default methods (cat-1)
	if _, ok := endorService.Methods["cat-1/schema"]; !ok {
		t.Errorf("method 'cat-1/schema' not found in endorService methods map")
	}
	if _, ok := endorService.Methods["cat-1/instance"]; !ok {
		t.Errorf("method 'cat-1/instance' not found in endorService methods map")
	}
	if _, ok := (*endorService.Methods["cat-1/instance"].GetOptions().InputSchema.Properties)["id"]; !ok {
		t.Errorf("'id' property not found in input schema for method 'cat-1/instance'")
	}
	if _, ok := endorService.Methods["cat-1/list"]; !ok {
		t.Errorf("method 'cat-1/list' not found in endorService methods map")
	}
	if _, ok := endorService.Methods["cat-1/create"]; !ok {
		t.Errorf("method 'cat-1/create' not found in endorService methods map")
	}
	if dataSchema, ok := (*endorService.Methods["cat-1/create"].GetOptions().InputSchema.Properties)["data"]; ok {
		if _, ok := (*dataSchema.Properties)["id"]; !ok {
			t.Errorf("input schema for method 'cat-1/create' wrong")
		}
		if _, ok := (*dataSchema.Properties)["attribute"]; !ok {
			t.Errorf("input schema for method 'cat-1/create' wrong")
		}
		if _, ok := (*dataSchema.Properties)["additionalAttribute"]; !ok {
			t.Errorf("input schema for method 'cat-1/create' wrong")
		}
		if _, ok := (*dataSchema.Properties)["additionalAttributeCat1"]; !ok {
			t.Errorf("input schema for method 'cat-1/create' wrong")
		}
	} else {
		t.Errorf("'data' property not found in input schema for method 'cat-1/create'")
	}
	if dataSchema, ok := (*endorService.Methods["cat-1/update"].GetOptions().InputSchema.Properties)["data"]; ok {
		if _, ok := (*dataSchema.Properties)["id"]; !ok {
			t.Errorf("input schema for method 'cat-1/' wrong")
		}
		if _, ok := (*dataSchema.Properties)["attribute"]; !ok {
			t.Errorf("input schema for method 'cat-1/' wrong")
		}
		if _, ok := (*dataSchema.Properties)["additionalAttribute"]; !ok {
			t.Errorf("input schema for method 'cat-1/' wrong")
		}
		if _, ok := (*dataSchema.Properties)["additionalAttributeCat1"]; !ok {
			t.Errorf("input schema for method 'cat-1/create' wrong")
		}
	} else {
		t.Errorf("'data' property not found in input schema for method 'cat-1/'")
	}
	if _, ok := (*endorService.Methods["cat-1/update"].GetOptions().InputSchema.Properties)["id"]; !ok {
		t.Errorf("'id' property not found in input schema for method 'cat-1/update'")
	}
	if _, ok := endorService.Methods["cat-1/delete"]; !ok {
		t.Errorf("method 'cat-1/delete' not found in endorService methods map")
	}
	if _, ok := (*endorService.Methods["cat-1/delete"].GetOptions().InputSchema.Properties)["id"]; !ok {
		t.Errorf("'id' property not found in input schema for method 'cat-1/delete'")
	}
	// check categories default methods (cat-2)
	if _, ok := endorService.Methods["cat-2/create"]; !ok {
		t.Errorf("method 'cat-2/create' not found in endorService methods map")
	}
	if dataSchema, ok := (*endorService.Methods["cat-2/create"].GetOptions().InputSchema.Properties)["data"]; ok {
		if _, ok := (*dataSchema.Properties)["id"]; !ok {
			t.Errorf("input schema for method 'cat-2/create' wrong")
		}
		if _, ok := (*dataSchema.Properties)["attribute"]; !ok {
			t.Errorf("input schema for method 'cat-2/create' wrong")
		}
		if _, ok := (*dataSchema.Properties)["additionalAttribute"]; !ok {
			t.Errorf("input schema for method 'cat-2/create' wrong")
		}
		if _, ok := (*dataSchema.Properties)["additionalAttributeCat2"]; !ok {
			t.Errorf("input schema for method 'cat-2/create' wrong")
		}
	} else {
		t.Errorf("'data' property not found in input schema for method 'cat-2/create'")
	}
}
