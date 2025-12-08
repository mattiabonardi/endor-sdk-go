package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
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
	assert.Empty(t, schema.Reference, "Expected no root reference")

	// Definitions should be empty
	assert.Empty(t, schema.Definitions, "Expected empty definitions")

	// Root should be object type with properties
	assert.Equal(t, sdk.SchemaTypeObject, schema.Type, "Expected object type at root")

	assert.NotNil(t, schema.Properties, "Expected properties to be present")

	userSchemaProperties := *schema.Properties
	assert.Len(t, userSchemaProperties, 11, "Expected 11 properties")

	assert.Equal(t, sdk.SchemaTypeString, userSchemaProperties["id"].Type, "Expected id to be string type")
	assert.Equal(t, sdk.SchemaTypeInteger, userSchemaProperties["age"].Type, "Expected age to be integer type")
	assert.Equal(t, sdk.SchemaTypeBoolean, userSchemaProperties["active"].Type, "Expected active to be boolean type")
	assert.Equal(t, sdk.SchemaTypeArray, userSchemaProperties["hobbies"].Type, "Expected hobbies to be array type")
	assert.Equal(t, sdk.SchemaTypeString, userSchemaProperties["hobbies"].Items.Type, "Expected hobbies items to be string type")

	// address should be expanded object, not reference
	assert.Empty(t, userSchemaProperties["address"].Reference, "Expected address to be expanded, not referenced")
	assert.Equal(t, sdk.SchemaTypeObject, userSchemaProperties["address"].Type, "Expected address to be object type")
	assert.NotNil(t, userSchemaProperties["address"].Properties, "Expected address properties to be expanded")
	addressSchemaProperties := *userSchemaProperties["address"].Properties
	assert.Len(t, addressSchemaProperties, 2, "Expected 2 address properties")

	// cars array should have expanded items, not references
	assert.Equal(t, sdk.SchemaTypeArray, userSchemaProperties["cars"].Type, "Expected cars to be array type")
	assert.Empty(t, userSchemaProperties["cars"].Items.Reference, "Expected cars items to be expanded, not referenced")
	assert.Equal(t, sdk.SchemaTypeObject, userSchemaProperties["cars"].Items.Type, "Expected cars items to be object type")
	assert.NotNil(t, userSchemaProperties["cars"].Items.Properties, "Expected cars items properties to be expanded")
	carSchemaProperties := *userSchemaProperties["cars"].Items.Properties
	assert.Len(t, carSchemaProperties, 1, "Expected 1 car property")

	// date time
	assert.Equal(t, sdk.SchemaTypeString, userSchemaProperties["dateTime"].Type, "Expected dateTime to be string type")
	assert.Equal(t, sdk.SchemaFormatDateTime, *userSchemaProperties["dateTime"].Format, "Expected dateTime format to be datetime")

	// date (with decorators)
	assert.Equal(t, sdk.SchemaTypeString, userSchemaProperties["date"].Type, "Expected date to be string type")
	assert.Equal(t, "The Date", *userSchemaProperties["date"].Description, "Expected correct date description")
	assert.Equal(t, sdk.SchemaFormatDate, *userSchemaProperties["date"].Format, "Expected date format to be date")

	// ui schema
	assert.Equal(t, "xxx", *userSchemaProperties["date"].UISchema.Resource, "Expected correct UI schema resource")
	assert.True(t, *userSchemaProperties["date"].UISchema.Hidden, "Expected UI schema hidden to be true")

	// Root UI schema order
	assert.Len(t, *schema.UISchema.Order, 11, "Expected 11 ordered fields")
	order := *schema.UISchema.Order
	assert.Equal(t, "id", order[0], "Expected first field to be 'id'")
	assert.Equal(t, "name", order[1], "Expected second field to be 'name'")
	assert.Equal(t, "active", order[4], "Expected fifth field to be 'active'")
	assert.Equal(t, "car", order[8], "Expected ninth field to be 'car'")
}

func TestRicorsionTypes(t *testing.T) {
	schema := sdk.NewSchema(&CarTreeNode{})

	// With expanded schema, no definitions and no root reference
	assert.Empty(t, schema.Definitions, "Expected empty definitions")
	assert.Empty(t, schema.Reference, "Expected no root reference")

	// Root should be object type with properties
	assert.Equal(t, sdk.SchemaTypeObject, schema.Type, "Expected object type at root")

	assert.NotNil(t, schema.Properties, "Expected properties to be present")

	carTreeNodeSchemaProperties := *schema.Properties

	// Check value property (should be expanded Car object)
	valueProp := carTreeNodeSchemaProperties["value"]
	assert.Equal(t, sdk.SchemaTypeObject, valueProp.Type, "Expected value to be object type")
	assert.NotNil(t, valueProp.Properties, "Expected value properties to be expanded")

	// Check children property (array with recursive handling)
	assert.Equal(t, sdk.SchemaTypeArray, carTreeNodeSchemaProperties["children"].Type, "Expected children to be array type")

	// The recursive reference should be handled as a simple string schema
	childrenItems := carTreeNodeSchemaProperties["children"].Items
	assert.Equal(t, sdk.SchemaTypeString, childrenItems.Type, "Expected recursive children items to be string type (recursion prevention)")
	assert.NotNil(t, childrenItems.Description, "Expected recursive description to be present")
	assert.Contains(t, *childrenItems.Description, "Recursive reference", "Expected recursive description to contain 'Recursive reference'")
}

func TestNoPayload(t *testing.T) {
	schema := sdk.NewSchema(sdk.NoPayload{})

	// With expanded schema, no definitions and no root reference
	assert.Empty(t, schema.Definitions, "Expected empty definitions")
	assert.Empty(t, schema.Reference, "Expected no root reference")

	// Should be object type with empty properties
	assert.Equal(t, sdk.SchemaTypeObject, schema.Type, "Expected object type")
	assert.NotNil(t, schema.Properties, "Expected properties to be present (even if empty)")

	noPayloadSchemaProperties := *schema.Properties
	assert.Empty(t, noPayloadSchemaProperties, "Expected 0 properties")
}

func TestWithGenerics(t *testing.T) {
	schema := sdk.NewSchema(&GenericCar[Car]{})

	// With expanded schema, no definitions and no root reference
	assert.Empty(t, schema.Definitions, "Expected empty definitions")
	assert.Empty(t, schema.Reference, "Expected no root reference")

	// Should be object type with properties
	assert.Equal(t, sdk.SchemaTypeObject, schema.Type, "Expected object type")
	assert.NotNil(t, schema.Properties, "Expected properties to be present")

	genericCarSchemaProperties := *schema.Properties

	// Value property should be expanded Car object, not reference
	valueProp := genericCarSchemaProperties["value"]
	assert.Empty(t, valueProp.Reference, "Expected value to be expanded, not referenced")
	assert.Equal(t, sdk.SchemaTypeObject, valueProp.Type, "Expected value to be object type")
	assert.NotNil(t, valueProp.Properties, "Expected value properties to be expanded")
}

func TestExpandedSchema(t *testing.T) {
	// Test the new expanded schema functionality
	schema := sdk.NewSchema(&User{})

	// Should not have any definitions (empty map)
	assert.Empty(t, schema.Definitions, "Expected empty definitions")

	// Should not have a reference at root level
	assert.Empty(t, schema.Reference, "Expected no root reference")

	// Should have type object at root
	assert.Equal(t, sdk.SchemaTypeObject, schema.Type, "Expected object type at root")

	// Should have properties directly expanded
	assert.NotNil(t, schema.Properties, "Expected properties to be present")

	props := *schema.Properties
	assert.Len(t, props, 11, "Expected 11 properties")

	// Check nested object is expanded, not referenced
	addressProp := props["address"]
	assert.Empty(t, addressProp.Reference, "Expected address to be expanded, not referenced")
	assert.Equal(t, sdk.SchemaTypeObject, addressProp.Type, "Expected address to be object type")
	assert.NotNil(t, addressProp.Properties, "Expected address properties to be expanded")

	addressProps := *addressProp.Properties
	assert.Len(t, addressProps, 2, "Expected 2 address properties")
}
