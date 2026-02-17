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

// Tests for ObjectIDStrategy
func TestObjectIDStrategy_CreateFilter(t *testing.T) {
	strategy := &ObjectIDStrategy{}
	validID := primitive.NewObjectID().Hex()

	t.Run("valid ObjectID", func(t *testing.T) {
		filter, err := strategy.CreateFilter(validID)
		assert.NoError(t, err)
		assert.NotNil(t, filter)
		assert.NotNil(t, filter["_id"])
	})

	t.Run("invalid ObjectID", func(t *testing.T) {
		_, err := strategy.CreateFilter("invalid-id")
		assert.Error(t, err)
	})
}

func TestObjectIDStrategy_GenerateID(t *testing.T) {
	strategy := &ObjectIDStrategy{}
	id := strategy.GenerateID()
	assert.NotEmpty(t, id)
	oid, ok := id.(primitive.ObjectID)
	assert.True(t, ok, "Generated ID should be primitive.ObjectID")
	assert.False(t, oid.IsZero(), "Generated ID should not be zero")
}

// Tests for StringIDStrategy
func TestStringIDStrategy_CreateFilter(t *testing.T) {
	strategy := &StringIDStrategy{}

	filter, err := strategy.CreateFilter("test-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-id", filter["_id"])
}

func TestStringIDStrategy_FromStorageFormat(t *testing.T) {
	strategy := &StringIDStrategy{}

	t.Run("valid string ID", func(t *testing.T) {
		id, err := strategy.FromStorageFormat("test-id")
		assert.NoError(t, err)
		assert.Equal(t, "test-id", id)
	})

	t.Run("invalid type", func(t *testing.T) {
		_, err := strategy.FromStorageFormat(123)
		assert.Error(t, err)
	})
}

// Tests for DocumentMapper
func TestDocumentMapper_ToModel(t *testing.T) {
	mapper := &DocumentMapper[*TestEntity]{}
	testID := primitive.NewObjectID()

	raw := bson.M{
		"_id":  testID,
		"name": "TestModel",
		"age":  25,
	}

	model, err := mapper.ToModel(raw)
	assert.NoError(t, err)
	assert.Equal(t, "TestModel", model.Name)
	assert.Equal(t, int32(25), model.Age)
	assert.Equal(t, testID.Hex(), model.ID)
}

func TestDocumentMapper_ToDocument(t *testing.T) {
	mapper := &DocumentMapper[*TestEntity]{}
	idStrategy := &StringIDStrategy{}

	testID := "test-123"
	model := &TestEntity{
		ID:   testID,
		Name: "TestDoc",
		Age:  30,
	}

	metadata := map[string]interface{}{
		"version": 1,
	}

	doc, err := mapper.ToDocument(model, metadata, idStrategy)
	assert.NoError(t, err)
	assert.Equal(t, "test-123", doc["_id"])
	assert.Equal(t, "TestDoc", doc["name"])
	assert.Equal(t, int32(30), doc["age"])
	// Metadata fields are now inline at root level
	assert.Equal(t, 1, doc["version"])
}

func TestDocumentMapper_WithEmbeddedStruct_ToModel(t *testing.T) {
	mapper := &DocumentMapper[*TestEmbeddedEntity]{}
	testID := primitive.NewObjectID()

	raw := bson.M{
		"_id":   testID,
		"name":  "TestEmbedded",
		"age":   35,
		"email": "test@example.com",
	}

	model, err := mapper.ToModel(raw)
	assert.NoError(t, err)
	assert.Equal(t, "TestEmbedded", model.Name)
	assert.Equal(t, int32(35), model.Age)
	assert.Equal(t, "test@example.com", model.Email)
	assert.Equal(t, testID.Hex(), model.GetID())
}

func TestDocumentMapper_WithEmbeddedStruct_ToDocument(t *testing.T) {
	mapper := &DocumentMapper[*TestEmbeddedEntity]{}
	idStrategy := &StringIDStrategy{}

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

	doc, err := mapper.ToDocument(model, metadata, idStrategy)
	assert.NoError(t, err)
	assert.Equal(t, "test-embedded-123", doc["_id"])
	assert.Equal(t, "TestEmbedded", doc["name"])
	assert.Equal(t, int32(35), doc["age"])
	assert.Equal(t, "test@example.com", doc["email"])
	// Metadata fields are now inline at root level
	assert.Equal(t, 1, doc["version"])

	// Verify that embedded struct fields are at the top level, not nested
	_, hasNestedStruct := doc["testbasemodel"]
	assert.False(t, hasNestedStruct, "Embedded struct should be inlined, not nested")
}

func TestDocumentMapper_RoundTripWithEmbeddedStruct(t *testing.T) {
	mapper := &DocumentMapper[*TestEmbeddedEntity]{}
	idStrategy := &StringIDStrategy{}

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
	doc, err := mapper.ToDocument(originalModel, metadata, idStrategy)
	assert.NoError(t, err)

	// Convert back to model
	reconstructedModel, err := mapper.ToModel(doc)
	assert.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, originalModel.GetID(), reconstructedModel.GetID())
	assert.Equal(t, originalModel.Name, reconstructedModel.Name)
	assert.Equal(t, originalModel.Age, reconstructedModel.Age)
	assert.Equal(t, originalModel.Email, reconstructedModel.Email)
}

func TestDocumentMapper_EmbeddedStructWithNestedID(t *testing.T) {
	// This test verifies that embedded structs where the ID is in the embedded struct
	// are handled correctly through the full round-trip
	mapper := &DocumentMapper[*TestEmbeddedEntity]{}

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
	model, err := mapper.ToModel(rawDoc)
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

func TestDocumentMapper_EmbeddedStructWithNonPointerID_ToModel(t *testing.T) {
	// This test reproduces the issue with HybridSpecializedModel
	mapper := &DocumentMapper[*TestExtendedWithStringID]{}

	rawDoc := bson.M{
		"_id":        "test-id-123",
		"type":       "test-type",
		"attribute":  "base-attribute",
		"extraField": "extra-value",
		"metadata": bson.M{
			"version": 1,
		},
	}

	model, err := mapper.ToModel(rawDoc)
	assert.NoError(t, err)

	// All fields should be populated correctly
	assert.Equal(t, "test-id-123", model.GetID())
	assert.Equal(t, "test-type", model.Type)
	assert.Equal(t, "base-attribute", model.Attribute, "Embedded struct attribute should be preserved")
	assert.Equal(t, "extra-value", model.ExtraField)
}

func TestDocumentMapper_EmbeddedStructWithNonPointerID_ToDocument(t *testing.T) {
	mapper := &DocumentMapper[*TestExtendedWithStringID]{}
	idStrategy := &StringIDStrategy{}

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

	doc, err := mapper.ToDocument(model, metadata, idStrategy)
	assert.NoError(t, err)

	assert.Equal(t, "test-id-456", doc["_id"])
	assert.Equal(t, "test-type", doc["type"])
	assert.Equal(t, "base-attribute", doc["attribute"])
	assert.Equal(t, "extra-value", doc["extraField"])
	// Metadata fields are now inline at root level
	assert.Equal(t, 2, doc["version"])
}

func TestDocumentMapper_EmbeddedStructWithNonPointerID_RoundTrip(t *testing.T) {
	mapper := &DocumentMapper[*TestExtendedWithStringID]{}
	idStrategy := &StringIDStrategy{}

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
	doc, err := mapper.ToDocument(original, metadata, idStrategy)
	assert.NoError(t, err)

	// Back to model
	reconstructed, err := mapper.ToModel(doc)
	assert.NoError(t, err)

	assert.Equal(t, original.GetID(), reconstructed.GetID())
	assert.Equal(t, original.Type, reconstructed.Type)
	assert.Equal(t, original.Attribute, reconstructed.Attribute, "Embedded struct attribute should survive round trip")
	assert.Equal(t, original.ExtraField, reconstructed.ExtraField)
}

func TestDocumentMapper_EmbeddedStructWithObjectID_RoundTrip(t *testing.T) {
	// This test simulates the HybridSpecializedModel scenario with ObjectID strategy
	mapper := &DocumentMapper[*TestExtendedWithStringID]{}
	idStrategy := &ObjectIDStrategy{}

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
	doc, err := mapper.ToDocument(original, metadata, idStrategy)
	assert.NoError(t, err)

	// Verify the document structure
	assert.Equal(t, testID, doc["_id"], "ID should be converted to ObjectID")
	assert.Equal(t, "hybrid-type", doc["type"])
	assert.Equal(t, "hybrid-attribute", doc["attribute"], "Attribute from embedded struct should be in document")
	assert.Equal(t, "hybrid-extra", doc["extraField"])

	// From document (simulates database read)
	reconstructed, err := mapper.ToModel(doc)
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
	registry := NewObjectIDFieldRegistry[*TestModelWithObjectID]()

	// Should identify all ObjectID fields
	assert.True(t, registry.IsObjectIDField("_id"), "Should detect _id as ObjectID field")
	assert.True(t, registry.IsObjectIDField("supplierId"), "Should detect supplierId as ObjectID field")
	assert.True(t, registry.IsObjectIDField("productId"), "Should detect productID as ObjectID field")

	// Should not include non-ObjectID fields
	assert.False(t, registry.IsObjectIDField("name"), "Should not detect name as ObjectID field")
	assert.False(t, registry.IsObjectIDField("quantity"), "Should not detect quantity as ObjectID field")
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

	registry := NewObjectIDFieldRegistry[*TestModelWithObjectID]()
	err := registry.ConvertToStorage(doc)
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

// TestDocumentMapperWithObjectID verifies full round-trip conversion
func TestDocumentMapperWithObjectID(t *testing.T) {
	mapper := &DocumentMapper[*TestModelWithObjectID]{}
	idStrategy := &ObjectIDStrategy{}

	// Create test model
	originalModel := &TestModelWithObjectID{
		ID:         sdk.GenerateObjectID(),
		SupplierID: sdk.GenerateObjectID(),
		ProductID:  sdk.GenerateObjectID(),
		Name:       "Test Widget",
		Quantity:   42,
	}

	// Convert to document
	doc, err := mapper.ToDocument(originalModel, map[string]interface{}{}, idStrategy)
	assert.NoError(t, err, "ToDocument should not error")

	// Verify ObjectID fields are primitive.ObjectID in document
	assert.IsType(t, primitive.ObjectID{}, doc["_id"], "_id should be primitive.ObjectID in document")
	assert.IsType(t, primitive.ObjectID{}, doc["supplierId"], "supplierId should be primitive.ObjectID in document")
	assert.IsType(t, primitive.ObjectID{}, doc["productId"], "productId should be primitive.ObjectID in document")

	// Convert back to model
	reconstructedModel, err := mapper.ToModel(doc)
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
	registry := NewObjectIDFieldRegistry[*TestExtendedModelWithObjectID]()

	// Should identify ObjectID fields from both base and extended structs
	assert.True(t, registry.IsObjectIDField("_id"), "Should detect _id from embedded struct")
	assert.True(t, registry.IsObjectIDField("categoryId"), "Should detect categoryId from extended struct")

	// Test round-trip conversion
	mapper := &DocumentMapper[*TestExtendedModelWithObjectID]{}
	idStrategy := &ObjectIDStrategy{}

	originalModel := &TestExtendedModelWithObjectID{
		TestBaseModelWithObjectID: TestBaseModelWithObjectID{
			ID:   sdk.GenerateObjectID(),
			Name: "Base Product",
		},
		CategoryID: sdk.GenerateObjectID(),
		Price:      99.99,
	}

	// Convert to document
	doc, err := mapper.ToDocument(originalModel, map[string]interface{}{}, idStrategy)
	assert.NoError(t, err)

	// Convert back to model
	reconstructedModel, err := mapper.ToModel(doc)
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
	registry := NewObjectIDFieldRegistry[*TestMixedIDModel]()

	// Should only identify ObjectID fields, not string fields
	assert.True(t, registry.IsObjectIDField("_id"))
	assert.True(t, registry.IsObjectIDField("refId"))
	assert.False(t, registry.IsObjectIDField("legacyId"), "String field should not be detected as ObjectID")
	assert.False(t, registry.IsObjectIDField("name"))

	// Test conversion
	mapper := &DocumentMapper[*TestMixedIDModel]{}
	idStrategy := &ObjectIDStrategy{}

	originalModel := &TestMixedIDModel{
		ID:       sdk.GenerateObjectID(),
		RefID:    sdk.GenerateObjectID(),
		LegacyID: "legacy-string-id-123",
		Name:     "Mixed Model",
	}

	doc, err := mapper.ToDocument(originalModel, map[string]interface{}{}, idStrategy)
	assert.NoError(t, err)

	// Verify ObjectID fields are primitive.ObjectID
	assert.IsType(t, primitive.ObjectID{}, doc["_id"])
	assert.IsType(t, primitive.ObjectID{}, doc["refId"])

	// Verify string fields remain strings
	assert.IsType(t, "", doc["legacyId"])
	assert.Equal(t, "legacy-string-id-123", doc["legacyId"])
	assert.Equal(t, "Mixed Model", doc["name"])

	// Round-trip
	reconstructedModel, err := mapper.ToModel(doc)
	assert.NoError(t, err)
	assert.Equal(t, originalModel.LegacyID, reconstructedModel.LegacyID)
}

// Test Model without json tag on one field (uses Go field name for that field)
type TestEntityNoBsonTag struct {
	ID          string `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description"` // No json tag - should use field name "Description"
}

func (t *TestEntityNoBsonTag) GetID() any {
	return t.ID
}

// Test model with ObjectID fields for filter conversion
type TestEntityWithObjectIDFields struct {
	ID       string        `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string        `bson:"name" json:"name"`
	UserID   sdk.ObjectID  `bson:"userId" json:"userId"`
	OwnerID  sdk.ObjectID  `bson:"ownerId" json:"ownerId"`
	ParentID *sdk.ObjectID `bson:"parentId,omitempty" json:"parentId,omitempty"`
}

func (t *TestEntityWithObjectIDFields) GetID() any {
	return t.ID
}

func TestObjectIDFieldRegistry_ConvertFilterToStorage(t *testing.T) {
	registry := NewObjectIDFieldRegistry[*TestEntityWithObjectIDFields]()

	// Verify ObjectID fields are registered
	assert.True(t, registry.IsObjectIDField("userId"))
	assert.True(t, registry.IsObjectIDField("ownerId"))
	assert.True(t, registry.IsObjectIDField("parentId"))
	assert.False(t, registry.IsObjectIDField("name"))

	t.Run("simple equality filter", func(t *testing.T) {
		userIDStr := "507f1f77bcf86cd799439011"
		filter := bson.M{
			"userId": userIDStr,
			"name":   "test",
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		// userId should be converted to primitive.ObjectID
		userIDObj, ok := filter["userId"].(primitive.ObjectID)
		assert.True(t, ok, "userId should be primitive.ObjectID")
		assert.Equal(t, userIDStr, userIDObj.Hex())

		// name should remain as string
		assert.Equal(t, "test", filter["name"])
	})

	t.Run("$in operator", func(t *testing.T) {
		userID1 := "507f1f77bcf86cd799439011"
		userID2 := "507f1f77bcf86cd799439012"

		filter := bson.M{
			"userId": bson.M{
				"$in": []interface{}{userID1, userID2},
			},
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		// Check that $in array elements are converted
		inClause := filter["userId"].(bson.M)["$in"].([]interface{})
		assert.Len(t, inClause, 2)

		oid1, ok := inClause[0].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, userID1, oid1.Hex())

		oid2, ok := inClause[1].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, userID2, oid2.Hex())
	})

	t.Run("$gt operator", func(t *testing.T) {
		userIDStr := "507f1f77bcf86cd799439011"

		filter := bson.M{
			"userId": bson.M{
				"$gt": userIDStr,
			},
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		gtValue := filter["userId"].(bson.M)["$gt"]
		oid, ok := gtValue.(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, userIDStr, oid.Hex())
	})

	t.Run("multiple operators", func(t *testing.T) {
		gtID := "507f1f77bcf86cd799439011"
		ltID := "607f1f77bcf86cd799439099"

		filter := bson.M{
			"userId": bson.M{
				"$gt": gtID,
				"$lt": ltID,
			},
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		userIDClause := filter["userId"].(bson.M)

		gtOID, ok := userIDClause["$gt"].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, gtID, gtOID.Hex())

		ltOID, ok := userIDClause["$lt"].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, ltID, ltOID.Hex())
	})

	t.Run("$or operator with ObjectID fields", func(t *testing.T) {
		userID1 := "507f1f77bcf86cd799439011"
		ownerID1 := "507f1f77bcf86cd799439012"

		filter := bson.M{
			"$or": []interface{}{
				bson.M{"userId": userID1},
				bson.M{"ownerId": ownerID1},
			},
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		orClauses := filter["$or"].([]interface{})
		assert.Len(t, orClauses, 2)

		// Check first condition
		cond1 := orClauses[0].(bson.M)
		userOID, ok := cond1["userId"].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, userID1, userOID.Hex())

		// Check second condition
		cond2 := orClauses[1].(bson.M)
		ownerOID, ok := cond2["ownerId"].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, ownerID1, ownerOID.Hex())
	})

	t.Run("$and operator with mixed fields", func(t *testing.T) {
		userIDStr := "507f1f77bcf86cd799439011"

		filter := bson.M{
			"$and": []interface{}{
				bson.M{"userId": userIDStr},
				bson.M{"name": "test"},
			},
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		andClauses := filter["$and"].([]interface{})
		assert.Len(t, andClauses, 2)

		// Check ObjectID field is converted
		cond1 := andClauses[0].(bson.M)
		userOID, ok := cond1["userId"].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, userIDStr, userOID.Hex())

		// Check string field remains as string
		cond2 := andClauses[1].(bson.M)
		assert.Equal(t, "test", cond2["name"])
	})

	t.Run("$nin operator with string array", func(t *testing.T) {
		userID1 := "507f1f77bcf86cd799439011"
		userID2 := "507f1f77bcf86cd799439012"

		filter := bson.M{
			"userId": bson.M{
				"$nin": []string{userID1, userID2},
			},
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		ninClause := filter["userId"].(bson.M)["$nin"].([]interface{})
		assert.Len(t, ninClause, 2)

		oid1, ok := ninClause[0].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, userID1, oid1.Hex())

		oid2, ok := ninClause[1].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, userID2, oid2.Hex())
	})

	t.Run("complex nested query", func(t *testing.T) {
		userID1 := "507f1f77bcf86cd799439011"
		userID2 := "507f1f77bcf86cd799439012"
		ownerID := "607f1f77bcf86cd799439099"

		filter := bson.M{
			"$or": []interface{}{
				bson.M{
					"userId": bson.M{
						"$in": []interface{}{userID1, userID2},
					},
				},
				bson.M{
					"$and": []interface{}{
						bson.M{"ownerId": ownerID},
						bson.M{"name": "test"},
					},
				},
			},
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		// Verify the structure is maintained and ObjectIDs are converted
		orClauses := filter["$or"].([]interface{})
		assert.Len(t, orClauses, 2)

		// First OR clause: userId $in
		firstClause := orClauses[0].(bson.M)
		userIDInClause := firstClause["userId"].(bson.M)["$in"].([]interface{})
		assert.Len(t, userIDInClause, 2)

		// Second OR clause: $and with ownerId and name
		secondClause := orClauses[1].(bson.M)
		andClauses := secondClause["$and"].([]interface{})
		assert.Len(t, andClauses, 2)

		ownerIDCond := andClauses[0].(bson.M)
		ownerOID, ok := ownerIDCond["ownerId"].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, ownerID, ownerOID.Hex())
	})

	t.Run("invalid ObjectID format", func(t *testing.T) {
		filter := bson.M{
			"userId": "invalid-objectid",
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ObjectID format")
	})

	t.Run("nil and empty values", func(t *testing.T) {
		filter := bson.M{
			"userId":  nil,
			"ownerId": "",
			"name":    "test",
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		// nil and empty should be preserved
		assert.Nil(t, filter["userId"])
		assert.Equal(t, "", filter["ownerId"])
		assert.Equal(t, "test", filter["name"])
	})

	t.Run("sdk.ObjectID type", func(t *testing.T) {
		userIDStr := "507f1f77bcf86cd799439011"
		userID := sdk.ObjectID(userIDStr)

		filter := bson.M{
			"userId": userID,
		}

		err := registry.ConvertFilterToStorage(filter)
		assert.NoError(t, err)

		oid, ok := filter["userId"].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, userIDStr, oid.Hex())
	})
}
