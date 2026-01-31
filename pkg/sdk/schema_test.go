package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order is a test struct for SchemaTransformer tests
type Order struct {
	ID          string `json:"id" schema:"readOnly=true"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Date        string `json:"date" schema:"format=date"`
	WarehouseID string `json:"warehouseId"`
	Notes       string `json:"notes,omitempty"`
	ReceivedQty int    `json:"receivedQty,omitempty"`
}

func TestSchemaTransformer(t *testing.T) {
	t.Run("Require marks fields as required", func(t *testing.T) {
		schema := sdk.NewSchema(Order{})

		schema.Apply(sdk.Require("type", "date", "warehouseId"))

		assert.Len(t, schema.Required, 3)
		assert.Contains(t, schema.Required, "type")
		assert.Contains(t, schema.Required, "date")
		assert.Contains(t, schema.Required, "warehouseId")
	})

	t.Run("Require ignores non-existent fields", func(t *testing.T) {
		schema := sdk.NewSchema(Order{})

		schema.Apply(sdk.Require("type", "nonExistentField"))

		assert.Len(t, schema.Required, 1)
		assert.Contains(t, schema.Required, "type")
		assert.NotContains(t, schema.Required, "nonExistentField")
	})

	t.Run("Forbid removes fields from schema", func(t *testing.T) {
		schema := sdk.NewSchema(Order{})

		// Verify field exists before forbid
		assert.Contains(t, *schema.Properties, "receivedQty")

		schema.Apply(sdk.Forbid("receivedQty"))

		// Verify field is removed
		assert.NotContains(t, *schema.Properties, "receivedQty")
		// Verify other fields still exist
		assert.Contains(t, *schema.Properties, "type")
		assert.Contains(t, *schema.Properties, "id")
	})

	t.Run("Forbid removes fields from UISchema order", func(t *testing.T) {
		schema := sdk.NewSchema(Order{})

		// Verify field is in order before forbid
		assert.Contains(t, *schema.UISchema.Order, "receivedQty")

		schema.Apply(sdk.Forbid("receivedQty", "notes"))

		// Verify fields are removed from order
		assert.NotContains(t, *schema.UISchema.Order, "receivedQty")
		assert.NotContains(t, *schema.UISchema.Order, "notes")
		// Verify other fields still in order
		assert.Contains(t, *schema.UISchema.Order, "type")
	})

	t.Run("ReadOnly marks fields as read-only", func(t *testing.T) {
		schema := sdk.NewSchema(Order{})

		schema.Apply(sdk.ReadOnly("id", "status"))

		props := *schema.Properties
		assert.NotNil(t, props["id"].ReadOnly)
		assert.True(t, *props["id"].ReadOnly)
		assert.NotNil(t, props["status"].ReadOnly)
		assert.True(t, *props["status"].ReadOnly)
		// Verify other fields are not affected
		assert.Nil(t, props["type"].ReadOnly)
	})

	t.Run("Multiple transformers can be chained", func(t *testing.T) {
		schema := sdk.NewSchema(Order{})

		schema.Apply(
			sdk.Require("type", "date", "warehouseId"),
			sdk.Forbid("receivedQty"),
			sdk.ReadOnly("id", "status"),
		)

		// Check required
		assert.Len(t, schema.Required, 3)
		assert.Contains(t, schema.Required, "type")

		// Check forbidden
		assert.NotContains(t, *schema.Properties, "receivedQty")

		// Check read-only
		props := *schema.Properties
		assert.True(t, *props["id"].ReadOnly)
		assert.True(t, *props["status"].ReadOnly)
	})

	t.Run("CreateOrder use case schema", func(t *testing.T) {
		schema := sdk.NewSchema(Order{}).Apply(
			sdk.Require("type", "date", "warehouseId"),
			sdk.Forbid("receivedQty"),
			sdk.ReadOnly("id", "status"),
		)

		props := *schema.Properties

		// Required fields
		assert.Contains(t, schema.Required, "type")
		assert.Contains(t, schema.Required, "date")
		assert.Contains(t, schema.Required, "warehouseId")

		// Forbidden field removed
		assert.NotContains(t, props, "receivedQty")

		// Read-only fields
		assert.True(t, *props["id"].ReadOnly)
		assert.True(t, *props["status"].ReadOnly)

		// Other fields remain writable
		assert.Nil(t, props["type"].ReadOnly)
		assert.Nil(t, props["notes"].ReadOnly)
	})

	t.Run("ReceiveGoods use case schema", func(t *testing.T) {
		schema := sdk.NewSchema(Order{}).Apply(
			sdk.Require("receivedQty"),
			sdk.Forbid("type", "date", "warehouseId", "notes"),
			sdk.ReadOnly("id", "status"),
		)

		props := *schema.Properties

		// Required field
		assert.Contains(t, schema.Required, "receivedQty")

		// Forbidden fields removed
		assert.NotContains(t, props, "type")
		assert.NotContains(t, props, "date")
		assert.NotContains(t, props, "warehouseId")
		assert.NotContains(t, props, "notes")

		// Read-only fields
		assert.True(t, *props["id"].ReadOnly)
		assert.True(t, *props["status"].ReadOnly)

		// receivedQty remains and is writable
		assert.Contains(t, props, "receivedQty")
		assert.Nil(t, props["receivedQty"].ReadOnly)
	})

	t.Run("Apply returns the same RootSchema for chaining", func(t *testing.T) {
		schema := sdk.NewSchema(Order{})
		result := schema.Apply(sdk.Require("type"))

		assert.Same(t, schema, result)
	})

	t.Run("Transformers on nil properties do nothing", func(t *testing.T) {
		schema := &sdk.RootSchema{}

		// Should not panic
		assert.NotPanics(t, func() {
			schema.Apply(
				sdk.Require("field"),
				sdk.Forbid("field"),
				sdk.ReadOnly("field"),
			)
		})
	})
}

// OrderItem is a nested struct for testing nested transformations
type OrderItem struct {
	ProductID   string  `json:"productId"`
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Notes       string  `json:"notes,omitempty"`
}

// Warehouse is a nested struct for testing nested object transformations
type Warehouse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// OrderWithItems is a test struct with nested objects and arrays
type OrderWithItems struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Status    string      `json:"status"`
	Items     []OrderItem `json:"items"`
	Warehouse Warehouse   `json:"warehouse"`
}

func TestSchemaTransformerNested(t *testing.T) {
	t.Run("Require nested field in object", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{})

		schema.Apply(sdk.Require("warehouse.name", "warehouse.address"))

		// Check nested required
		warehouseProps := (*schema.Properties)["warehouse"]
		assert.Contains(t, warehouseProps.Required, "name")
		assert.Contains(t, warehouseProps.Required, "address")
	})

	t.Run("Require nested field in array items", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{})

		schema.Apply(sdk.Require("items.productId", "items.quantity"))

		// Check nested required in array items
		itemsProps := (*schema.Properties)["items"]
		assert.NotNil(t, itemsProps.Items)
		assert.Contains(t, itemsProps.Items.Required, "productId")
		assert.Contains(t, itemsProps.Items.Required, "quantity")
	})

	t.Run("Forbid nested field in object", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{})

		// Verify field exists before forbid
		warehouseProps := (*schema.Properties)["warehouse"]
		assert.Contains(t, *warehouseProps.Properties, "address")

		schema.Apply(sdk.Forbid("warehouse.address"))

		// Verify nested field is removed
		warehouseProps = (*schema.Properties)["warehouse"]
		assert.NotContains(t, *warehouseProps.Properties, "address")
		// Other nested fields still exist
		assert.Contains(t, *warehouseProps.Properties, "id")
		assert.Contains(t, *warehouseProps.Properties, "name")
	})

	t.Run("Forbid nested field in array items", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{})

		// Verify field exists before forbid
		itemsSchema := (*schema.Properties)["items"].Items
		assert.Contains(t, *itemsSchema.Properties, "notes")

		schema.Apply(sdk.Forbid("items.notes", "items.price"))

		// Verify nested fields are removed from array items
		itemsSchema = (*schema.Properties)["items"].Items
		assert.NotContains(t, *itemsSchema.Properties, "notes")
		assert.NotContains(t, *itemsSchema.Properties, "price")
		// Other nested fields still exist
		assert.Contains(t, *itemsSchema.Properties, "productId")
		assert.Contains(t, *itemsSchema.Properties, "quantity")
	})

	t.Run("ReadOnly nested field in object", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{})

		schema.Apply(sdk.ReadOnly("warehouse.id"))

		warehouseProps := *(*schema.Properties)["warehouse"].Properties
		assert.NotNil(t, warehouseProps["id"].ReadOnly)
		assert.True(t, *warehouseProps["id"].ReadOnly)
		// Other fields not affected
		assert.Nil(t, warehouseProps["name"].ReadOnly)
	})

	t.Run("ReadOnly nested field in array items", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{})

		schema.Apply(sdk.ReadOnly("items.productId"))

		itemsProps := *(*schema.Properties)["items"].Items.Properties
		assert.NotNil(t, itemsProps["productId"].ReadOnly)
		assert.True(t, *itemsProps["productId"].ReadOnly)
		// Other fields not affected
		assert.Nil(t, itemsProps["quantity"].ReadOnly)
	})

	t.Run("WriteOnly nested field", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{})

		schema.Apply(sdk.WriteOnly("warehouse.address", "items.price"))

		// Check object nested field
		warehouseProps := *(*schema.Properties)["warehouse"].Properties
		assert.NotNil(t, warehouseProps["address"].WriteOnly)
		assert.True(t, *warehouseProps["address"].WriteOnly)

		// Check array items nested field
		itemsProps := *(*schema.Properties)["items"].Items.Properties
		assert.NotNil(t, itemsProps["price"].WriteOnly)
		assert.True(t, *itemsProps["price"].WriteOnly)
	})

	t.Run("Mixed top-level and nested transformations", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{}).Apply(
			sdk.Require("type", "items.productId", "items.quantity", "warehouse.name"),
			sdk.Forbid("status", "items.notes"),
			sdk.ReadOnly("id", "warehouse.id"),
		)

		props := *schema.Properties

		// Top-level required
		assert.Contains(t, schema.Required, "type")

		// Top-level forbidden
		assert.NotContains(t, props, "status")

		// Top-level read-only
		assert.True(t, *props["id"].ReadOnly)

		// Nested required in array items
		assert.Contains(t, props["items"].Items.Required, "productId")
		assert.Contains(t, props["items"].Items.Required, "quantity")

		// Nested forbidden in array items
		assert.NotContains(t, *props["items"].Items.Properties, "notes")

		// Nested required in object
		assert.Contains(t, props["warehouse"].Required, "name")

		// Nested read-only in object
		warehouseProps := *props["warehouse"].Properties
		assert.True(t, *warehouseProps["id"].ReadOnly)
	})

	t.Run("Non-existent nested path does nothing", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{})

		// Should not panic
		assert.NotPanics(t, func() {
			schema.Apply(
				sdk.Require("nonexistent.field"),
				sdk.Forbid("items.nonexistent"),
				sdk.ReadOnly("warehouse.nonexistent"),
			)
		})
	})

	t.Run("CreateOrderWithItems use case", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{}).Apply(
			sdk.Require("type", "items", "items.productId", "items.quantity", "warehouse.name"),
			sdk.Forbid("items.notes"),
			sdk.ReadOnly("id", "status", "warehouse.id"),
		)

		props := *schema.Properties

		// Verify top-level
		assert.Contains(t, schema.Required, "type")
		assert.Contains(t, schema.Required, "items")
		assert.True(t, *props["id"].ReadOnly)
		assert.True(t, *props["status"].ReadOnly)

		// Verify nested array items
		itemsSchema := props["items"].Items
		assert.Contains(t, itemsSchema.Required, "productId")
		assert.Contains(t, itemsSchema.Required, "quantity")
		assert.NotContains(t, *itemsSchema.Properties, "notes")

		// Verify nested object
		warehouseSchema := props["warehouse"]
		assert.Contains(t, warehouseSchema.Required, "name")
		warehouseProps := *warehouseSchema.Properties
		assert.True(t, *warehouseProps["id"].ReadOnly)
	})
}

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
	Date       string             `json:"date" schema:"description=The Date,format=date" ui-schema:"entity=xxx,hidden=true"`
}

type CarTreeNode struct {
	Value    *Car          `json:"value"`
	Children []CarTreeNode `json:"children"`
}

type GenericCar[T any] struct {
	Value T `json:"value"`
}

type Map struct {
	StringFilter map[string]string `json:"stringFilter"`
	AnyFilter    map[string]any    `json:"anyFilter"`
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
	assert.Equal(t, "xxx", *userSchemaProperties["date"].UISchema.Entity, "Expected correct UI schema entity")
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

// Test structs for embedded struct testing
type BaseEntity struct {
	ID          string `json:"id" schema:"title=Entity ID"`
	Description string `json:"description" schema:"title=Entity Description"`
	CreatedAt   string `json:"createdAt" schema:"format=date-time"`
}

type ExtendedEntity struct {
	BaseEntity `json:",inline" bson:",inline"`
	Name       string `json:"name" schema:"title=Name"`
	Value      int    `json:"value" schema:"title=Value"`
}

type DeepEmbeddedEntity struct {
	ExtendedEntity `json:",inline"`
	ExtraField     string `json:"extraField" schema:"title=Extra Field"`
}

func TestEmbeddedStructInline(t *testing.T) {
	schema := sdk.NewSchema(&ExtendedEntity{})

	// With expanded schema, no definitions and no root reference
	assert.Empty(t, schema.Definitions, "Expected empty definitions")
	assert.Empty(t, schema.Reference, "Expected no root reference")

	// Root should be object type with properties
	assert.Equal(t, sdk.SchemaTypeObject, schema.Type, "Expected object type at root")
	assert.NotNil(t, schema.Properties, "Expected properties to be present")

	props := *schema.Properties

	// Should have all properties from both BaseEntity and ExtendedEntity inlined
	// BaseEntity has: id, description, createdAt (3)
	// ExtendedEntity has: name, value (2)
	// Total: 5 properties
	assert.Len(t, props, 5, "Expected 5 properties (3 from BaseEntity + 2 from ExtendedEntity)")

	// Check BaseEntity properties are inlined
	assert.Contains(t, props, "id", "Expected 'id' property from BaseEntity")
	assert.Contains(t, props, "description", "Expected 'description' property from BaseEntity")
	assert.Contains(t, props, "createdAt", "Expected 'createdAt' property from BaseEntity")

	// Check ExtendedEntity properties
	assert.Contains(t, props, "name", "Expected 'name' property from ExtendedEntity")
	assert.Contains(t, props, "value", "Expected 'value' property from ExtendedEntity")

	// Verify types
	assert.Equal(t, sdk.SchemaTypeString, props["id"].Type, "Expected id to be string type")
	assert.Equal(t, sdk.SchemaTypeString, props["description"].Type, "Expected description to be string type")
	assert.Equal(t, sdk.SchemaTypeString, props["createdAt"].Type, "Expected createdAt to be string type")
	assert.Equal(t, sdk.SchemaTypeString, props["name"].Type, "Expected name to be string type")
	assert.Equal(t, sdk.SchemaTypeInteger, props["value"].Type, "Expected value to be integer type")

	// Verify schema decorators from embedded struct
	assert.Equal(t, "Entity ID", *props["id"].Title, "Expected id title from schema decorator")
	assert.Equal(t, sdk.SchemaFormatDateTime, *props["createdAt"].Format, "Expected createdAt format from schema decorator")

	// Verify UISchema order includes all fields in correct order
	assert.NotNil(t, schema.UISchema, "Expected UISchema to be present")
	assert.NotNil(t, schema.UISchema.Order, "Expected UISchema.Order to be present")
	order := *schema.UISchema.Order
	assert.Len(t, order, 5, "Expected 5 ordered fields")

	// BaseEntity fields should come first (from the embedded struct), then ExtendedEntity fields
	assert.Equal(t, "id", order[0], "Expected first field to be 'id' from BaseEntity")
	assert.Equal(t, "description", order[1], "Expected second field to be 'description' from BaseEntity")
	assert.Equal(t, "createdAt", order[2], "Expected third field to be 'createdAt' from BaseEntity")
	assert.Equal(t, "name", order[3], "Expected fourth field to be 'name' from ExtendedEntity")
	assert.Equal(t, "value", order[4], "Expected fifth field to be 'value' from ExtendedEntity")
}

func TestDeepEmbeddedStructInline(t *testing.T) {
	schema := sdk.NewSchema(&DeepEmbeddedEntity{})

	// With expanded schema, no definitions and no root reference
	assert.Empty(t, schema.Definitions, "Expected empty definitions")
	assert.Empty(t, schema.Reference, "Expected no root reference")

	// Root should be object type with properties
	assert.Equal(t, sdk.SchemaTypeObject, schema.Type, "Expected object type at root")
	assert.NotNil(t, schema.Properties, "Expected properties to be present")

	props := *schema.Properties

	// Should have all properties from BaseEntity, ExtendedEntity, and DeepEmbeddedEntity inlined
	// BaseEntity: id, description, createdAt (3)
	// ExtendedEntity: name, value (2)
	// DeepEmbeddedEntity: extraField (1)
	// Total: 6 properties
	assert.Len(t, props, 6, "Expected 6 properties (3 from BaseEntity + 2 from ExtendedEntity + 1 from DeepEmbeddedEntity)")

	// Check all properties are present
	assert.Contains(t, props, "id", "Expected 'id' property")
	assert.Contains(t, props, "description", "Expected 'description' property")
	assert.Contains(t, props, "createdAt", "Expected 'createdAt' property")
	assert.Contains(t, props, "name", "Expected 'name' property")
	assert.Contains(t, props, "value", "Expected 'value' property")
	assert.Contains(t, props, "extraField", "Expected 'extraField' property")

	// Verify UISchema order
	order := *schema.UISchema.Order
	assert.Len(t, order, 6, "Expected 6 ordered fields")
	assert.Equal(t, "extraField", order[5], "Expected last field to be 'extraField' from DeepEmbeddedEntity")
}

func TestMap(t *testing.T) {
	schema := sdk.NewSchema(&Map{})

	// With expanded schema, no definitions and no root reference
	assert.Empty(t, schema.Definitions, "Expected empty definitions")
	assert.Empty(t, schema.Reference, "Expected no root reference")

	// Root should be object type with properties
	assert.Equal(t, sdk.SchemaTypeObject, schema.Type, "Expected object type at root")
	assert.NotNil(t, schema.Properties, "Expected properties to be present")

	props := *schema.Properties
	assert.Len(t, props, 2, "Expected 2 properties")

	// Assert string map (map[string]string)
	stringFilterProp := props["stringFilter"]
	assert.Equal(t, sdk.SchemaTypeObject, stringFilterProp.Type, "Expected stringFilter to be object type")
	assert.NotNil(t, stringFilterProp.AdditionalProperties, "Expected stringFilter to have additionalProperties")
	assert.Equal(t, sdk.SchemaTypeString, stringFilterProp.AdditionalProperties.Type, "Expected stringFilter additionalProperties to be string type")

	// Assert any map (map[string]any)
	anyFilterProp := props["anyFilter"]
	assert.Equal(t, sdk.SchemaTypeObject, anyFilterProp.Type, "Expected anyFilter to be object type")
	assert.NotNil(t, anyFilterProp.AdditionalProperties, "Expected anyFilter to have additionalProperties")
	// For map[string]any, additionalProperties should be an empty schema (allowing any type)
	assert.Empty(t, anyFilterProp.AdditionalProperties.Type, "Expected anyFilter additionalProperties to have no type (any)")
}
