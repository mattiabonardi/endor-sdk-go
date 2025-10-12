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
	ID         primitive.ObjectID `json:"id" ui-schema:"id=true"`
	Name       string             `json:"name"`
	Email      string             `json:"email"`
	Age        int                `json:"age"`
	Active     bool               `json:"active"`
	Hobbies    []string           `json:"hobbies"`
	Address    Address            `json:"address"`
	Cars       []Car              `json:"cars"`
	CurrentCar Car                `json:"car"`
	DateTime   primitive.DateTime `json:"dateTime"`
	Date       string             `json:"date" schema:"description=The Date,format=date" ui-schema:"resource=xxx,hidden=true"`
}

type CarTreeNode struct {
	Value    *Car          `json:"value"`
	Children []CarTreeNode `json:"children"`
}

type GenericCar[T any] struct {
	Value T `json:"value"`
}

func TestSchemaTypes(t *testing.T) {
	schema := sdk.NewSchema(&User{})
	if schema.Reference != "#/$defs/User" {
		t.Fatalf("Received %v", schema.Reference)
	}
	if len(schema.Definitions) != 3 {
		t.Fatalf("Received %v", schema.Definitions)
	}
	userSchema := schema.Definitions["User"]
	userSchemaProperties := *userSchema.Properties
	if userSchema.Type != sdk.ObjectType {
		t.Fatalf("Received %v", schema.Type)
	}
	if len(userSchemaProperties) != 11 {
		t.Fatalf("Received %v", len(userSchemaProperties))
	}
	if userSchemaProperties["id"].Type != sdk.StringType {
		t.Fatalf("Received %v", userSchemaProperties["id"].Type)
	}
	if userSchemaProperties["age"].Type != sdk.IntegerType {
		t.Fatalf("Received %v", userSchemaProperties["age"].Type)
	}
	if userSchemaProperties["active"].Type != sdk.BooleanType {
		t.Fatalf("Received %v", userSchemaProperties["active"].Type)
	}
	if userSchemaProperties["hobbies"].Type != sdk.ArrayType {
		t.Fatalf("Received %v", userSchemaProperties["hobbies"].Type)
	}
	if userSchemaProperties["hobbies"].Items.Type != sdk.StringType {
		t.Fatalf("Received %v", userSchemaProperties["hobbies"].Items.Type)
	}
	// object --> ref
	if userSchemaProperties["address"].Reference != "#/$defs/Address" {
		t.Fatalf("Received %v", userSchemaProperties["address"].Reference)
	}
	addressSchema := schema.Definitions["Address"]
	addressSchemaProperties := *addressSchema.Properties
	if len(addressSchemaProperties) != 2 {
		t.Fatalf("Received %v", len(addressSchemaProperties))
	}
	// object array --> items ref
	if userSchemaProperties["cars"].Type != sdk.ArrayType {
		t.Fatalf("Received %v", userSchemaProperties["cars"].Type)
	}
	if userSchemaProperties["cars"].Items.Reference != "#/$defs/Car" {
		t.Fatalf("Received %v", userSchemaProperties["cars"].Items.Reference)
	}
	carSchema := schema.Definitions["Car"]
	carSchemaProperties := *carSchema.Properties
	if len(carSchemaProperties) != 1 {
		t.Fatalf("Received %v", len(carSchemaProperties))
	}
	// date time
	if userSchemaProperties["dateTime"].Type != sdk.StringType {
		t.Fatalf("Received %v", userSchemaProperties["dateTime"].Type)
	}
	if *userSchemaProperties["dateTime"].Format != sdk.DateTimeFormat {
		t.Fatalf("Received %v", userSchemaProperties["dateTime"].Format)
	}
	// date (with decorators)
	if userSchemaProperties["date"].Type != sdk.StringType {
		t.Fatalf("Received %v", userSchemaProperties["date"].Type)
	}
	if *userSchemaProperties["date"].Description != "The Date" {
		t.Fatalf("Received %v", userSchemaProperties["date"].Description)
	}
	if *userSchemaProperties["date"].Format != sdk.DateFormat {
		t.Fatalf("Received %v", userSchemaProperties["date"].Format)
	}
	// ui schema
	if *userSchemaProperties["date"].UISchema.Resource != "xxx" {
		t.Fatalf("Received %v", userSchemaProperties["date"].UISchema.Resource)
	}
	if !*userSchemaProperties["date"].UISchema.Hidden {
		t.Fatalf("Received %v", userSchemaProperties["date"].UISchema.Hidden)
	}
	if len(*userSchema.UISchema.Order) != 11 {
		t.Fatalf("Received %v", len(*userSchema.UISchema.Order))
	}
	order := *userSchema.UISchema.Order
	if order[0] != "id" {
		t.Fatalf("Received %v", order[0])
	}
	if order[1] != "name" {
		t.Fatalf("Received %v", order[1])
	}
	if order[4] != "active" {
		t.Fatalf("Received %v", order[4])
	}
	if order[8] != "car" {
		t.Fatalf("Received %v", order[8])
	}
	if *userSchema.UISchema.Id != "id" {
		t.Fatalf("Received %v", len(*userSchema.UISchema.Id))
	}
}

func TestRicorsionTypes(t *testing.T) {
	schema := sdk.NewSchema(&CarTreeNode{})
	carTreeNodeSchema := schema.Definitions["CarTreeNode"]
	carTreeNodeSchemaProperties := *carTreeNodeSchema.Properties
	if len(schema.Definitions) != 2 {
		t.Fatalf("Received %v", schema.Definitions)
	}
	if schema.Reference != "#/$defs/CarTreeNode" {
		t.Fatalf("Received %v", schema.Reference)
	}
	if carTreeNodeSchemaProperties["value"].Reference != "#/$defs/Car" {
		t.Fatalf("Received %v", carTreeNodeSchemaProperties["value"].Reference)
	}
	if carTreeNodeSchemaProperties["children"].Type != sdk.ArrayType {
		t.Fatalf("Received %v", carTreeNodeSchemaProperties["children"].Type)
	}
	if carTreeNodeSchemaProperties["children"].Items.Reference != "#/$defs/CarTreeNode" {
		t.Fatalf("Received %v", carTreeNodeSchemaProperties["children"].Items.Reference)
	}
}

func TestNoPayload(t *testing.T) {
	schema := sdk.NewSchema(sdk.NoPayload{})
	noPayloadSchema := schema.Definitions["NoPayload"]
	noPayloadSchemaProperties := *noPayloadSchema.Properties
	if noPayloadSchema.Type != sdk.ObjectType {
		t.Fatalf("Received %v", schema.Type)
	}
	if len(noPayloadSchemaProperties) != 0 {
		t.Fatalf("Received %v", len(noPayloadSchemaProperties))
	}
}

func TestWithGenerics(t *testing.T) {
	schema := sdk.NewSchema(&GenericCar[Car]{})
	if len(schema.Definitions) != 2 {
		t.Fatalf("Received %v", len(schema.Definitions))
	}
	if schema.Reference != "#/$defs/GenericCar_Car" {
		t.Fatalf("Received %v", schema.Reference)
	}
	genericCarSchema := schema.Definitions["GenericCar_Car"]
	genericCarSchemaProperties := *genericCarSchema.Properties
	if genericCarSchemaProperties["value"].Reference != "#/$defs/Car" {
		t.Fatalf("Received %v", genericCarSchemaProperties["value"].Reference)
	}
}
