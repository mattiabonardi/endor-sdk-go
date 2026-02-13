package repository

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test Model
type TestEntity struct {
	ID   string `bson:"_id,omitempty" json:"id,omitempty"`
	Name string `bson:"name" json:"name"`
	Age  int32  `bson:"age" json:"age"`
}

func (t *TestEntity) GetID() any {
	return t.ID
}

// Test Specialized Model for category-specific fields
type TestSpecializedEntity struct {
	ID   string `bson:"_id,omitempty" json:"id,omitempty"`
	Name string `bson:"name" json:"name"`
	Age  int32  `bson:"age" json:"age"`
	Type string `bson:"categoryType,omitempty" json:"categoryType,omitempty"`
}

func (t *TestSpecializedEntity) GetID() any {
	return t.ID
}

// Test Model with embedded struct (inline)
type TestBaseModel struct {
	ID   string `bson:"_id,omitempty" json:"id,omitempty"`
	Name string `bson:"name" json:"name"`
}

func (t *TestBaseModel) GetID() any {
	return t.ID
}

type TestEmbeddedEntity struct {
	TestBaseModel `bson:",inline" json:",inline"`
	Age           int32  `bson:"age" json:"age"`
	Email         string `bson:"email" json:"email"`
}

func (t *TestEmbeddedEntity) GetID() any {
	return t.TestBaseModel.ID
}

func (t *TestSpecializedEntity) GetCategoryType() string {
	return t.Type
}

func (t *TestSpecializedEntity) SetCategoryType(categoryType string) {
	t.Type = categoryType
}

type TestSpecializedEntityCategory struct {
	ExtraField string `bson:"extraField" json:"extraField"`
	Priority   int32  `bson:"priority" json:"priority"`
}

// Tests for ObjectIDConverter
func TestObjectIDConverter_ToFilter(t *testing.T) {
	converter := &ObjectIDConverter{}
	validID := primitive.NewObjectID().Hex()

	t.Run("valid ObjectID", func(t *testing.T) {
		filter, err := converter.ToFilter(validID)
		assert.NoError(t, err)
		assert.NotNil(t, filter)
		assert.NotNil(t, filter["_id"])
	})

	t.Run("invalid ObjectID", func(t *testing.T) {
		_, err := converter.ToFilter("invalid-id")
		assert.Error(t, err)
	})
}

func TestObjectIDConverter_GenerateNewID(t *testing.T) {
	converter := &ObjectIDConverter{}
	id := converter.GenerateNewID()
	assert.NotEmpty(t, id)
	idStr, ok := id.(string)
	assert.True(t, ok, "Generated ID should be a string")
	assert.Equal(t, 24, len(idStr))
}

// Tests for StringIDConverter
func TestStringIDConverter_ToFilter(t *testing.T) {
	converter := &StringIDConverter{}

	filter, err := converter.ToFilter("test-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-id", filter["_id"])
}

func TestStringIDConverter_FromStorageID(t *testing.T) {
	converter := &StringIDConverter{}

	t.Run("valid string ID", func(t *testing.T) {
		id, err := converter.FromStorageID("test-id")
		assert.NoError(t, err)
		assert.Equal(t, "test-id", id)
	})

	t.Run("invalid type", func(t *testing.T) {
		_, err := converter.FromStorageID(123)
		assert.Error(t, err)
	})
}

// Tests for DocumentConverter
func TestDocumentConverter_ExtractMetadata(t *testing.T) {
	converter := &DocumentConverter[*TestEntity]{}

	t.Run("with metadata", func(t *testing.T) {
		raw := bson.M{
			"_id":  "123",
			"name": "Test",
			"metadata": bson.M{
				"key1": "value1",
				"key2": 42,
			},
		}

		metadata, err := converter.ExtractMetadata(raw)
		assert.NoError(t, err)
		assert.Equal(t, "value1", metadata["key1"])
		assert.Equal(t, int32(42), metadata["key2"])
	})

	t.Run("without metadata", func(t *testing.T) {
		raw := bson.M{
			"_id":  "123",
			"name": "Test",
		}

		metadata, err := converter.ExtractMetadata(raw)
		assert.NoError(t, err)
		assert.Empty(t, metadata)
	})
}

func TestDocumentConverter_ToModel(t *testing.T) {
	converter := &DocumentConverter[*TestEntity]{}
	idConverter := &ObjectIDConverter{}
	testID := primitive.NewObjectID()

	raw := bson.M{
		"_id":  testID,
		"name": "TestModel",
		"age":  25,
		"metadata": bson.M{
			"extra": "data",
		},
	}

	model, err := converter.ToModel(raw, idConverter)
	assert.NoError(t, err)
	assert.Equal(t, "TestModel", model.Name)
	assert.Equal(t, int32(25), model.Age)
	assert.Equal(t, testID.Hex(), model.ID)
}

func TestDocumentConverter_ToDocument(t *testing.T) {
	converter := &DocumentConverter[*TestEntity]{}
	idConverter := &StringIDConverter{}

	testID := "test-123"
	model := &TestEntity{
		ID:   testID,
		Name: "TestDoc",
		Age:  30,
	}

	metadata := map[string]interface{}{
		"version": 1,
	}

	doc, err := converter.ToDocument(model, metadata, idConverter)
	assert.NoError(t, err)
	assert.Equal(t, "test-123", doc["_id"])
	assert.Equal(t, "TestDoc", doc["name"])
	assert.Equal(t, int32(30), doc["age"])
	assert.Equal(t, metadata, doc["metadata"])
}

func TestDocumentConverter_WithEmbeddedStruct_ToModel(t *testing.T) {
	converter := &DocumentConverter[*TestEmbeddedEntity]{}
	idConverter := &ObjectIDConverter{}
	testID := primitive.NewObjectID()

	raw := bson.M{
		"_id":   testID,
		"name":  "TestEmbedded",
		"age":   35,
		"email": "test@example.com",
		"metadata": bson.M{
			"extra": "data",
		},
	}

	model, err := converter.ToModel(raw, idConverter)
	assert.NoError(t, err)
	assert.Equal(t, "TestEmbedded", model.Name)
	assert.Equal(t, int32(35), model.Age)
	assert.Equal(t, "test@example.com", model.Email)
	assert.Equal(t, testID.Hex(), model.GetID())
}

func TestDocumentConverter_WithEmbeddedStruct_ToDocument(t *testing.T) {
	converter := &DocumentConverter[*TestEmbeddedEntity]{}
	idConverter := &StringIDConverter{}

	testID := "test-embedded-123"
	model := &TestEmbeddedEntity{
		TestBaseModel: TestBaseModel{
			ID:   testID,
			Name: "TestEmbedded",
		},
		Age:   35,
		Email: "test@example.com",
	}

	metadata := map[string]interface{}{
		"version": 1,
	}

	doc, err := converter.ToDocument(model, metadata, idConverter)
	assert.NoError(t, err)
	assert.Equal(t, "test-embedded-123", doc["_id"])
	assert.Equal(t, "TestEmbedded", doc["name"])
	assert.Equal(t, int32(35), doc["age"])
	assert.Equal(t, "test@example.com", doc["email"])
	assert.Equal(t, metadata, doc["metadata"])

	// Verify that embedded struct fields are at the top level, not nested
	_, hasNestedStruct := doc["testbasemodel"]
	assert.False(t, hasNestedStruct, "Embedded struct should be inlined, not nested")
}

func TestDocumentConverter_RoundTripWithEmbeddedStruct(t *testing.T) {
	converter := &DocumentConverter[*TestEmbeddedEntity]{}
	idConverter := &StringIDConverter{}

	testID := "round-trip-123"
	originalModel := &TestEmbeddedEntity{
		TestBaseModel: TestBaseModel{
			ID:   testID,
			Name: "RoundTrip",
		},
		Age:   40,
		Email: "roundtrip@example.com",
	}

	metadata := map[string]interface{}{
		"created": "2024-01-01",
	}

	// Convert to document
	doc, err := converter.ToDocument(originalModel, metadata, idConverter)
	assert.NoError(t, err)

	// Convert back to model
	reconstructedModel, err := converter.ToModel(doc, idConverter)
	assert.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, originalModel.GetID(), reconstructedModel.GetID())
	assert.Equal(t, originalModel.Name, reconstructedModel.Name)
	assert.Equal(t, originalModel.Age, reconstructedModel.Age)
	assert.Equal(t, originalModel.Email, reconstructedModel.Email)
}

func TestDocumentConverter_EmbeddedStructWithNestedID(t *testing.T) {
	// This test verifies that embedded structs where the ID is in the embedded struct
	// are handled correctly through the full round-trip
	converter := &DocumentConverter[*TestEmbeddedEntity]{}
	idConverter := &StringIDConverter{}

	// Simulate a raw document from MongoDB
	rawDoc := bson.M{
		"_id":   "embedded-id-123",
		"name":  "EmbeddedTest",
		"age":   int32(25),
		"email": "embedded@test.com",
		"metadata": bson.M{
			"version": 1,
		},
	}

	// Convert to model
	model, err := converter.ToModel(rawDoc, idConverter)
	assert.NoError(t, err)

	// The ID should be set correctly even though it's in the embedded struct
	assert.NotNil(t, model.GetID())
	assert.Equal(t, "embedded-id-123", model.GetID())
	assert.Equal(t, "EmbeddedTest", model.Name)
	assert.Equal(t, int32(25), model.Age)
	assert.Equal(t, "embedded@test.com", model.Email)
}

// Test Model similar to HybridSpecializedModel with non-pointer ID
type TestBaseWithStringID struct {
	ID        string `bson:"_id" json:"id"`
	Type      string `bson:"type" json:"type"`
	Attribute string `bson:"attribute" json:"attribute"`
}

func (t *TestBaseWithStringID) GetID() any {
	return t.ID
}

type TestExtendedWithStringID struct {
	TestBaseWithStringID `bson:",inline" json:",inline"`
	ExtraField           string `bson:"extraField" json:"extraField"`
}

func TestDocumentConverter_EmbeddedStructWithNonPointerID_ToModel(t *testing.T) {
	// This test reproduces the issue with HybridSpecializedModel
	converter := &DocumentConverter[*TestExtendedWithStringID]{}
	idConverter := &StringIDConverter{}

	rawDoc := bson.M{
		"_id":        "test-id-123",
		"type":       "test-type",
		"attribute":  "base-attribute",
		"extraField": "extra-value",
		"metadata": bson.M{
			"version": 1,
		},
	}

	model, err := converter.ToModel(rawDoc, idConverter)
	assert.NoError(t, err)

	// All fields should be populated correctly
	assert.Equal(t, "test-id-123", model.GetID())
	assert.Equal(t, "test-type", model.Type)
	assert.Equal(t, "base-attribute", model.Attribute, "Embedded struct attribute should be preserved")
	assert.Equal(t, "extra-value", model.ExtraField)
}

func TestDocumentConverter_EmbeddedStructWithNonPointerID_ToDocument(t *testing.T) {
	converter := &DocumentConverter[*TestExtendedWithStringID]{}
	idConverter := &StringIDConverter{}

	model := &TestExtendedWithStringID{
		TestBaseWithStringID: TestBaseWithStringID{
			ID:        "test-id-456",
			Type:      "test-type",
			Attribute: "base-attribute",
		},
		ExtraField: "extra-value",
	}

	metadata := map[string]interface{}{
		"version": 2,
	}

	doc, err := converter.ToDocument(model, metadata, idConverter)
	assert.NoError(t, err)

	assert.Equal(t, "test-id-456", doc["_id"])
	assert.Equal(t, "test-type", doc["type"])
	assert.Equal(t, "base-attribute", doc["attribute"])
	assert.Equal(t, "extra-value", doc["extraField"])
	assert.Equal(t, metadata, doc["metadata"])
}

func TestDocumentConverter_EmbeddedStructWithNonPointerID_RoundTrip(t *testing.T) {
	converter := &DocumentConverter[*TestExtendedWithStringID]{}
	idConverter := &StringIDConverter{}

	original := &TestExtendedWithStringID{
		TestBaseWithStringID: TestBaseWithStringID{
			ID:        "round-trip-789",
			Type:      "round-type",
			Attribute: "round-attribute",
		},
		ExtraField: "round-extra",
	}

	metadata := map[string]interface{}{
		"version": 3,
	}

	// To document
	doc, err := converter.ToDocument(original, metadata, idConverter)
	assert.NoError(t, err)

	// Back to model
	reconstructed, err := converter.ToModel(doc, idConverter)
	assert.NoError(t, err)

	assert.Equal(t, original.GetID(), reconstructed.GetID())
	assert.Equal(t, original.Type, reconstructed.Type)
	assert.Equal(t, original.Attribute, reconstructed.Attribute, "Embedded struct attribute should survive round trip")
	assert.Equal(t, original.ExtraField, reconstructed.ExtraField)
}

func TestDocumentConverter_EmbeddedStructWithObjectID_RoundTrip(t *testing.T) {
	// This test simulates the HybridSpecializedModel scenario with ObjectID converter
	converter := &DocumentConverter[*TestExtendedWithStringID]{}
	idConverter := &ObjectIDConverter{}

	testID := primitive.NewObjectID()
	original := &TestExtendedWithStringID{
		TestBaseWithStringID: TestBaseWithStringID{
			ID:        testID.Hex(),
			Type:      "hybrid-type",
			Attribute: "hybrid-attribute", // This field was reported as problematic
		},
		ExtraField: "hybrid-extra",
	}

	metadata := map[string]interface{}{
		"version": 1,
	}

	// To document (simulates database write)
	doc, err := converter.ToDocument(original, metadata, idConverter)
	assert.NoError(t, err)

	// Verify the document structure
	assert.Equal(t, testID, doc["_id"], "ID should be converted to ObjectID")
	assert.Equal(t, "hybrid-type", doc["type"])
	assert.Equal(t, "hybrid-attribute", doc["attribute"], "Attribute from embedded struct should be in document")
	assert.Equal(t, "hybrid-extra", doc["extraField"])

	// From document (simulates database read)
	reconstructed, err := converter.ToModel(doc, idConverter)
	assert.NoError(t, err)

	// All fields should be preserved through the round trip
	assert.Equal(t, original.ID, reconstructed.GetID())
	assert.Equal(t, original.Type, reconstructed.Type)
	assert.Equal(t, original.Attribute, reconstructed.Attribute, "Attribute from embedded struct must be preserved")
	assert.Equal(t, original.ExtraField, reconstructed.ExtraField)
}

// TestModelWithObjectID demonstrates using ObjectID in entity models
type TestModelWithObjectID struct {
	ID         sdk.ObjectID `bson:"_id" json:"id"`
	SupplierID sdk.ObjectID `bson:"supplierId" json:"supplierId"`
	ProductID  sdk.ObjectID `bson:"productId" json:"productId"`
	Name       string       `bson:"name" json:"name"`
	Quantity   int          `bson:"quantity" json:"quantity"`
}

func (t *TestModelWithObjectID) GetID() any {
	return t.ID
}

// TestObjectIDFieldDetection verifies that ObjectID fields are correctly identified
func TestObjectIDFieldDetection(t *testing.T) {
	fields := getObjectIDFields[*TestModelWithObjectID]()

	// Should identify all ObjectID fields
	assert.Contains(t, fields, "_id", "Should detect _id as ObjectID field")
	assert.Contains(t, fields, "supplierId", "Should detect supplierId as ObjectID field")
	assert.Contains(t, fields, "productId", "Should detect productID as ObjectID field")

	// Should not include non-ObjectID fields
	assert.NotContains(t, fields, "name", "Should not detect name as ObjectID field")
	assert.NotContains(t, fields, "quantity", "Should not detect quantity as ObjectID field")

	// Should have exactly 3 ObjectID fields
	assert.Equal(t, 3, len(fields), "Should detect exactly 3 ObjectID fields")
}

// TestConvertObjectIDsToStorage verifies conversion from ObjectID to primitive.ObjectID
func TestConvertObjectIDsToStorage(t *testing.T) {
	oid1 := primitive.NewObjectID()
	oid2 := primitive.NewObjectID()
	oid3 := primitive.NewObjectID()

	doc := bson.M{
		"_id":        oid1.Hex(),
		"supplierId": oid2.Hex(),
		"productId":  oid3.Hex(),
		"name":       "Test Product",
		"quantity":   10,
	}

	fields := map[string]struct{}{
		"_id":        {},
		"supplierId": {},
		"productId":  {},
	}

	err := convertObjectIDsToStorage(doc, fields)
	assert.NoError(t, err, "Should convert without error")

	// Verify conversion to primitive.ObjectID
	assert.IsType(t, primitive.ObjectID{}, doc["_id"], "_id should be primitive.ObjectID")
	assert.IsType(t, primitive.ObjectID{}, doc["supplierId"], "supplierId should be primitive.ObjectID")
	assert.IsType(t, primitive.ObjectID{}, doc["productId"], "productId should be primitive.ObjectID")

	// Verify correct ObjectID values
	assert.Equal(t, oid1, doc["_id"], "_id should match original")
	assert.Equal(t, oid2, doc["supplierId"], "supplierId should match original")
	assert.Equal(t, oid3, doc["productId"], "productId should match original")

	// Verify non-ObjectID fields unchanged
	assert.Equal(t, "Test Product", doc["name"], "name should remain unchanged")
	assert.Equal(t, 10, doc["quantity"], "quantity should remain unchanged")
}

// TestConvertObjectIDsFromStorage verifies conversion from primitive.ObjectID to string
func TestConvertObjectIDsFromStorage(t *testing.T) {
	oid1 := primitive.NewObjectID()
	oid2 := primitive.NewObjectID()
	oid3 := primitive.NewObjectID()

	doc := bson.M{
		"_id":        oid1,
		"supplierId": oid2,
		"productId":  oid3,
		"name":       "Test Product",
		"quantity":   10,
	}

	fields := map[string]struct{}{
		"_id":        {},
		"supplierId": {},
		"productId":  {},
	}

	convertObjectIDsFromStorage(doc, fields)

	// Verify conversion to string
	assert.IsType(t, "", doc["_id"], "_id should be string")
	assert.IsType(t, "", doc["supplierId"], "supplierId should be string")
	assert.IsType(t, "", doc["productId"], "productId should be string")

	// Verify correct hex values
	assert.Equal(t, oid1.Hex(), doc["_id"], "_id should match hex representation")
	assert.Equal(t, oid2.Hex(), doc["supplierId"], "supplierId should match hex representation")
	assert.Equal(t, oid3.Hex(), doc["productId"], "productId should match hex representation")

	// Verify non-ObjectID fields unchanged
	assert.Equal(t, "Test Product", doc["name"], "name should remain unchanged")
	assert.Equal(t, 10, doc["quantity"], "quantity should remain unchanged")
}

// TestDocumentConverterWithObjectID verifies full round-trip conversion
func TestDocumentConverterWithObjectID(t *testing.T) {
	converter := &DocumentConverter[*TestModelWithObjectID]{}
	idConverter := &ObjectIDConverter{}

	// Create test model
	originalModel := &TestModelWithObjectID{
		ID:         sdk.GenerateObjectID(),
		SupplierID: sdk.GenerateObjectID(),
		ProductID:  sdk.GenerateObjectID(),
		Name:       "Test Widget",
		Quantity:   42,
	}

	// Convert to document
	doc, err := converter.ToDocument(originalModel, map[string]interface{}{}, idConverter)
	assert.NoError(t, err, "ToDocument should not error")

	// Verify ObjectID fields are primitive.ObjectID in document
	assert.IsType(t, primitive.ObjectID{}, doc["_id"], "_id should be primitive.ObjectID in document")
	assert.IsType(t, primitive.ObjectID{}, doc["supplierId"], "supplierId should be primitive.ObjectID in document")
	assert.IsType(t, primitive.ObjectID{}, doc["productId"], "productId should be primitive.ObjectID in document")

	// Convert back to model
	reconstructedModel, err := converter.ToModel(doc, idConverter)
	assert.NoError(t, err, "ToModel should not error")

	// Verify all fields match
	assert.Equal(t, originalModel.ID, reconstructedModel.ID, "ID should match")
	assert.Equal(t, originalModel.SupplierID, reconstructedModel.SupplierID, "SupplierID should match")
	assert.Equal(t, originalModel.ProductID, reconstructedModel.ProductID, "ProductID should match")
	assert.Equal(t, originalModel.Name, reconstructedModel.Name, "Name should match")
	assert.Equal(t, originalModel.Quantity, reconstructedModel.Quantity, "Quantity should match")
}

// TestObjectIDWithEmbeddedStructs verifies ObjectID works with embedded structs
type TestBaseModelWithObjectID struct {
	ID   sdk.ObjectID `bson:"_id" json:"id"`
	Name string       `bson:"name" json:"name"`
}

func (t *TestBaseModelWithObjectID) GetID() any {
	return t.ID
}

type TestExtendedModelWithObjectID struct {
	TestBaseModelWithObjectID `bson:",inline" json:",inline"`
	CategoryID                sdk.ObjectID `bson:"categoryId" json:"categoryId"`
	Price                     float64      `bson:"price" json:"price"`
}

func TestObjectIDWithEmbeddedStructs(t *testing.T) {
	fields := getObjectIDFields[*TestExtendedModelWithObjectID]()

	// Should identify ObjectID fields from both base and extended structs
	assert.Contains(t, fields, "_id", "Should detect _id from embedded struct")
	assert.Contains(t, fields, "categoryId", "Should detect categoryId from extended struct")
	assert.Equal(t, 2, len(fields), "Should detect exactly 2 ObjectID fields")

	// Test round-trip conversion
	converter := &DocumentConverter[*TestExtendedModelWithObjectID]{}
	idConverter := &ObjectIDConverter{}

	originalModel := &TestExtendedModelWithObjectID{
		TestBaseModelWithObjectID: TestBaseModelWithObjectID{
			ID:   sdk.GenerateObjectID(),
			Name: "Base Product",
		},
		CategoryID: sdk.GenerateObjectID(),
		Price:      99.99,
	}

	// Convert to document
	doc, err := converter.ToDocument(originalModel, map[string]interface{}{}, idConverter)
	assert.NoError(t, err)

	// Convert back to model
	reconstructedModel, err := converter.ToModel(doc, idConverter)
	assert.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, originalModel.ID, reconstructedModel.ID)
	assert.Equal(t, originalModel.Name, reconstructedModel.Name)
	assert.Equal(t, originalModel.CategoryID, reconstructedModel.CategoryID)
	assert.Equal(t, originalModel.Price, reconstructedModel.Price)
}

// TestMixedIDTypes verifies handling of both ObjectID and string fields
type TestMixedIDModel struct {
	ID       sdk.ObjectID `bson:"_id" json:"id"`
	RefID    sdk.ObjectID `bson:"refId" json:"refId"`
	LegacyID string       `bson:"legacyId" json:"legacyId"`
	Name     string       `bson:"name" json:"name"`
}

func (t *TestMixedIDModel) GetID() any {
	return t.ID
}

func TestMixedIDTypes(t *testing.T) {
	fields := getObjectIDFields[*TestMixedIDModel]()

	// Should only identify ObjectID fields, not string fields
	assert.Contains(t, fields, "_id")
	assert.Contains(t, fields, "refId")
	assert.NotContains(t, fields, "legacyId", "String field should not be detected as ObjectID")
	assert.NotContains(t, fields, "name")
	assert.Equal(t, 2, len(fields))

	// Test conversion
	converter := &DocumentConverter[*TestMixedIDModel]{}
	idConverter := &ObjectIDConverter{}

	originalModel := &TestMixedIDModel{
		ID:       sdk.GenerateObjectID(),
		RefID:    sdk.GenerateObjectID(),
		LegacyID: "legacy-string-id-123",
		Name:     "Mixed Model",
	}

	doc, err := converter.ToDocument(originalModel, map[string]interface{}{}, idConverter)
	assert.NoError(t, err)

	// Verify ObjectID fields are primitive.ObjectID
	assert.IsType(t, primitive.ObjectID{}, doc["_id"])
	assert.IsType(t, primitive.ObjectID{}, doc["refId"])

	// Verify string fields remain strings
	assert.IsType(t, "", doc["legacyId"])
	assert.Equal(t, "legacy-string-id-123", doc["legacyId"])
	assert.Equal(t, "Mixed Model", doc["name"])

	// Round-trip
	reconstructedModel, err := converter.ToModel(doc, idConverter)
	assert.NoError(t, err)
	assert.Equal(t, originalModel.LegacyID, reconstructedModel.LegacyID)
}
