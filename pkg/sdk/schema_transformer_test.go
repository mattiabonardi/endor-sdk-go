package sdk_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
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

	t.Run("ReadOnlyExcept marks all fields read-only except specified ones", func(t *testing.T) {
		schema := sdk.NewSchema(Order{})

		schema.Apply(sdk.ReadOnlyExcept("type", "notes", "receivedQty"))

		props := *schema.Properties

		// Exception fields should NOT be read-only (explicitly set to false)
		assert.NotNil(t, props["type"].ReadOnly)
		assert.False(t, *props["type"].ReadOnly)
		assert.NotNil(t, props["notes"].ReadOnly)
		assert.False(t, *props["notes"].ReadOnly)
		assert.NotNil(t, props["receivedQty"].ReadOnly)
		assert.False(t, *props["receivedQty"].ReadOnly)

		// All other fields should be read-only
		assert.NotNil(t, props["id"].ReadOnly)
		assert.True(t, *props["id"].ReadOnly)
		assert.NotNil(t, props["status"].ReadOnly)
		assert.True(t, *props["status"].ReadOnly)
		assert.NotNil(t, props["date"].ReadOnly)
		assert.True(t, *props["date"].ReadOnly)
		assert.NotNil(t, props["warehouseId"].ReadOnly)
		assert.True(t, *props["warehouseId"].ReadOnly)
	})

	t.Run("ReadOnlyExcept with nested fields", func(t *testing.T) {
		schema := sdk.NewSchema(OrderWithItems{})

		schema.Apply(sdk.ReadOnlyExcept("type", "warehouse.name", "items.quantity"))

		props := *schema.Properties

		// Top-level exception should be writable
		assert.NotNil(t, props["type"].ReadOnly)
		assert.False(t, *props["type"].ReadOnly)

		// Top-level non-exception should be read-only
		assert.NotNil(t, props["id"].ReadOnly)
		assert.True(t, *props["id"].ReadOnly)
		assert.NotNil(t, props["status"].ReadOnly)
		assert.True(t, *props["status"].ReadOnly)

		// Nested object exception should be writable
		warehouseProps := *props["warehouse"].Properties
		assert.NotNil(t, warehouseProps["name"].ReadOnly)
		assert.False(t, *warehouseProps["name"].ReadOnly)

		// Nested object non-exception should be read-only
		assert.NotNil(t, warehouseProps["id"].ReadOnly)
		assert.True(t, *warehouseProps["id"].ReadOnly)
		assert.NotNil(t, warehouseProps["address"].ReadOnly)
		assert.True(t, *warehouseProps["address"].ReadOnly)

		// Nested array items exception should be writable
		itemsProps := *props["items"].Items.Properties
		assert.NotNil(t, itemsProps["quantity"].ReadOnly)
		assert.False(t, *itemsProps["quantity"].ReadOnly)

		// Nested array items non-exception should be read-only
		assert.NotNil(t, itemsProps["productId"].ReadOnly)
		assert.True(t, *itemsProps["productId"].ReadOnly)
		assert.NotNil(t, itemsProps["productName"].ReadOnly)
		assert.True(t, *itemsProps["productName"].ReadOnly)
		assert.NotNil(t, itemsProps["price"].ReadOnly)
		assert.True(t, *itemsProps["price"].ReadOnly)
		assert.NotNil(t, itemsProps["notes"].ReadOnly)
		assert.True(t, *itemsProps["notes"].ReadOnly)
	})

	t.Run("ReadOnlyExcept with no exceptions marks all read-only", func(t *testing.T) {
		schema := sdk.NewSchema(Order{})

		schema.Apply(sdk.ReadOnlyExcept())

		props := *schema.Properties

		// All fields should be read-only
		assert.NotNil(t, props["id"].ReadOnly)
		assert.True(t, *props["id"].ReadOnly)
		assert.NotNil(t, props["type"].ReadOnly)
		assert.True(t, *props["type"].ReadOnly)
		assert.NotNil(t, props["status"].ReadOnly)
		assert.True(t, *props["status"].ReadOnly)
		assert.NotNil(t, props["notes"].ReadOnly)
		assert.True(t, *props["notes"].ReadOnly)
	})

	t.Run("ReadOnlyExcept use case: mostly read-only schema", func(t *testing.T) {
		// Real-world scenario: 17 out of 20 fields should be read-only
		schema := sdk.NewSchema(Order{}).Apply(
			sdk.ReadOnlyExcept("status", "notes", "receivedQty"),
		)

		props := *schema.Properties

		// Only 3 writable fields
		assert.False(t, *props["status"].ReadOnly)
		assert.False(t, *props["notes"].ReadOnly)
		assert.False(t, *props["receivedQty"].ReadOnly)

		// All others read-only
		assert.True(t, *props["id"].ReadOnly)
		assert.True(t, *props["type"].ReadOnly)
		assert.True(t, *props["date"].ReadOnly)
		assert.True(t, *props["warehouseId"].ReadOnly)
	})
}
