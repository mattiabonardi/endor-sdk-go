package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Address struct {
	City  string `json:"city"`
	State string `json:"state"`
}

type Car struct {
	Id string `json:"id"`
}

type User struct {
	ID      primitive.ObjectID `json:"id"`
	Name    string             `json:"name"`
	Email   string             `json:"email"`
	Age     int                `json:"age"`
	Active  bool               `json:"active"`
	Hobbies []string           `json:"hobbies"`
	Address Address            `json:"address"`
	Cars    []Car              `json:"cars"`
}

func TestSchemaTypes(t *testing.T) {
	schema := sdk.NewSchema(&User{})
	if schema.Type != sdk.ObjectType {
		t.Fatalf("Received %v", schema.Type)
	}
	if len(schema.Properties) != 8 {
		t.Fatalf("Received %v", len(schema.Properties))
	}
	if schema.Properties["id"].Type != sdk.StringType {
		t.Fatalf("Received %v", schema.Properties["id"].Type)
	}
	if schema.Properties["age"].Type != sdk.IntegerType {
		t.Fatalf("Received %v", schema.Properties["age"].Type)
	}
	if schema.Properties["active"].Type != sdk.BooleanType {
		t.Fatalf("Received %v", schema.Properties["active"].Type)
	}
	if schema.Properties["hobbies"].Type != sdk.ArrayType {
		t.Fatalf("Received %v", schema.Properties["hobbies"].Type)
	}
	if schema.Properties["hobbies"].Items.Type != sdk.StringType {
		t.Fatalf("Received %v", schema.Properties["hobbies"].Items.Type)
	}
	if schema.Properties["address"].Type != sdk.ObjectType {
		t.Fatalf("Received %v", schema.Properties["address"].Type)
	}
	if schema.Properties["cars"].Type != sdk.ArrayType {
		t.Fatalf("Received %v", schema.Properties["cars"].Type)
	}
	if schema.Properties["cars"].Items.Type != sdk.ObjectType {
		t.Fatalf("Received %v", schema.Properties["cars"].Items.Type)
	}
	if schema.Properties["cars"].Items.Properties["id"].Type != sdk.StringType {
		t.Fatalf("Received %v", schema.Properties["cars"].Items.Properties["id"].Type)
	}
}
