package sdk_test

import (
	"strings"
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

	// With expanded schema, no root reference
	if schema.Reference != "" {
		t.Fatalf("Expected no root reference, got %v", schema.Reference)
	}

	// Definitions should be empty
	if len(schema.Definitions) != 0 {
		t.Fatalf("Expected empty definitions, got %d", len(schema.Definitions))
	}

	// Root should be object type with properties
	if schema.Type != sdk.ObjectType {
		t.Fatalf("Expected object type at root, got %v", schema.Type)
	}

	if schema.Properties == nil {
		t.Fatalf("Expected properties to be present")
	}

	userSchemaProperties := *schema.Properties
	if len(userSchemaProperties) != 11 {
		t.Fatalf("Expected 11 properties, got %d", len(userSchemaProperties))
	}

	if userSchemaProperties["id"].Type != sdk.StringType {
		t.Fatalf("Expected id to be string type, got %v", userSchemaProperties["id"].Type)
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

	// address should be expanded object, not reference
	if userSchemaProperties["address"].Reference != "" {
		t.Fatalf("Expected address to be expanded, not referenced")
	}
	if userSchemaProperties["address"].Type != sdk.ObjectType {
		t.Fatalf("Expected address to be object type, got %v", userSchemaProperties["address"].Type)
	}
	if userSchemaProperties["address"].Properties == nil {
		t.Fatalf("Expected address properties to be expanded")
	}
	addressSchemaProperties := *userSchemaProperties["address"].Properties
	if len(addressSchemaProperties) != 2 {
		t.Fatalf("Expected 2 address properties, got %d", len(addressSchemaProperties))
	}

	// cars array should have expanded items, not references
	if userSchemaProperties["cars"].Type != sdk.ArrayType {
		t.Fatalf("Received %v", userSchemaProperties["cars"].Type)
	}
	if userSchemaProperties["cars"].Items.Reference != "" {
		t.Fatalf("Expected cars items to be expanded, not referenced")
	}
	if userSchemaProperties["cars"].Items.Type != sdk.ObjectType {
		t.Fatalf("Expected cars items to be object type, got %v", userSchemaProperties["cars"].Items.Type)
	}
	if userSchemaProperties["cars"].Items.Properties == nil {
		t.Fatalf("Expected cars items properties to be expanded")
	}
	carSchemaProperties := *userSchemaProperties["cars"].Items.Properties
	if len(carSchemaProperties) != 1 {
		t.Fatalf("Expected 1 car property, got %d", len(carSchemaProperties))
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

	// Root UI schema order
	if len(*schema.UISchema.Order) != 11 {
		t.Fatalf("Expected 11 ordered fields, got %d", len(*schema.UISchema.Order))
	}
	order := *schema.UISchema.Order
	if order[0] != "id" {
		t.Fatalf("Expected first field to be 'id', got %v", order[0])
	}
	if order[1] != "name" {
		t.Fatalf("Expected second field to be 'name', got %v", order[1])
	}
	if order[4] != "active" {
		t.Fatalf("Expected fifth field to be 'active', got %v", order[4])
	}
	if order[8] != "car" {
		t.Fatalf("Expected ninth field to be 'car', got %v", order[8])
	}
	if *schema.UISchema.Id != "id" {
		t.Fatalf("Expected UI schema id to be 'id', got %v", *schema.UISchema.Id)
	}
}

func TestRicorsionTypes(t *testing.T) {
	schema := sdk.NewSchema(&CarTreeNode{})

	// With expanded schema, no definitions and no root reference
	if len(schema.Definitions) != 0 {
		t.Fatalf("Expected empty definitions, got %d", len(schema.Definitions))
	}
	if schema.Reference != "" {
		t.Fatalf("Expected no root reference, got %v", schema.Reference)
	}

	// Root should be object type with properties
	if schema.Type != sdk.ObjectType {
		t.Fatalf("Expected object type at root, got %v", schema.Type)
	}

	if schema.Properties == nil {
		t.Fatalf("Expected properties to be present")
	}

	carTreeNodeSchemaProperties := *schema.Properties

	// Check value property (should be expanded Car object)
	valueProp := carTreeNodeSchemaProperties["value"]
	if valueProp.Type != sdk.ObjectType {
		t.Fatalf("Expected value to be object type, got %v", valueProp.Type)
	}
	if valueProp.Properties == nil {
		t.Fatalf("Expected value properties to be expanded")
	}

	// Check children property (array with recursive handling)
	if carTreeNodeSchemaProperties["children"].Type != sdk.ArrayType {
		t.Fatalf("Expected children to be array type, got %v", carTreeNodeSchemaProperties["children"].Type)
	}

	// The recursive reference should be handled as a simple string schema
	childrenItems := carTreeNodeSchemaProperties["children"].Items
	if childrenItems.Type != sdk.StringType {
		t.Fatalf("Expected recursive children items to be string type (recursion prevention), got %v", childrenItems.Type)
	}
	if childrenItems.Description == nil || !strings.Contains(*childrenItems.Description, "Recursive reference") {
		t.Fatalf("Expected recursive description to be present")
	}
}

func TestNoPayload(t *testing.T) {
	schema := sdk.NewSchema(sdk.NoPayload{})

	// With expanded schema, no definitions and no root reference
	if len(schema.Definitions) != 0 {
		t.Fatalf("Expected empty definitions, got %d", len(schema.Definitions))
	}
	if schema.Reference != "" {
		t.Fatalf("Expected no root reference, got %v", schema.Reference)
	}

	// Should be object type with empty properties
	if schema.Type != sdk.ObjectType {
		t.Fatalf("Expected object type, got %v", schema.Type)
	}
	if schema.Properties == nil {
		t.Fatalf("Expected properties to be present (even if empty)")
	}

	noPayloadSchemaProperties := *schema.Properties
	if len(noPayloadSchemaProperties) != 0 {
		t.Fatalf("Expected 0 properties, got %d", len(noPayloadSchemaProperties))
	}
}

func TestWithGenerics(t *testing.T) {
	schema := sdk.NewSchema(&GenericCar[Car]{})

	// With expanded schema, no definitions and no root reference
	if len(schema.Definitions) != 0 {
		t.Fatalf("Expected empty definitions, got %d", len(schema.Definitions))
	}
	if schema.Reference != "" {
		t.Fatalf("Expected no root reference, got %v", schema.Reference)
	}

	// Should be object type with properties
	if schema.Type != sdk.ObjectType {
		t.Fatalf("Expected object type, got %v", schema.Type)
	}
	if schema.Properties == nil {
		t.Fatalf("Expected properties to be present")
	}

	genericCarSchemaProperties := *schema.Properties

	// Value property should be expanded Car object, not reference
	valueProp := genericCarSchemaProperties["value"]
	if valueProp.Reference != "" {
		t.Fatalf("Expected value to be expanded, not referenced")
	}
	if valueProp.Type != sdk.ObjectType {
		t.Fatalf("Expected value to be object type, got %v", valueProp.Type)
	}
	if valueProp.Properties == nil {
		t.Fatalf("Expected value properties to be expanded")
	}
}

func TestExpandedSchema(t *testing.T) {
	// Test the new expanded schema functionality
	schema := sdk.NewSchema(&User{})

	// Should not have any definitions (empty map)
	if len(schema.Definitions) != 0 {
		t.Fatalf("Expected empty definitions, got %d", len(schema.Definitions))
	}

	// Should not have a reference at root level
	if schema.Reference != "" {
		t.Fatalf("Expected no root reference, got %v", schema.Reference)
	}

	// Should have type object at root
	if schema.Type != sdk.ObjectType {
		t.Fatalf("Expected object type at root, got %v", schema.Type)
	}

	// Should have properties directly expanded
	if schema.Properties == nil {
		t.Fatalf("Expected properties to be present")
	}

	props := *schema.Properties
	if len(props) != 11 {
		t.Fatalf("Expected 11 properties, got %d", len(props))
	}

	// Check nested object is expanded, not referenced
	addressProp := props["address"]
	if addressProp.Reference != "" {
		t.Fatalf("Expected address to be expanded, not referenced")
	}
	if addressProp.Type != sdk.ObjectType {
		t.Fatalf("Expected address to be object type, got %v", addressProp.Type)
	}
	if addressProp.Properties == nil {
		t.Fatalf("Expected address properties to be expanded")
	}

	addressProps := *addressProp.Properties
	if len(addressProps) != 2 {
		t.Fatalf("Expected 2 address properties, got %d", len(addressProps))
	}
}
