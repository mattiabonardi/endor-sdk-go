package repository

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test Model
type TestResource struct {
	ID   *string `bson:"_id,omitempty" json:"id,omitempty"`
	Name string  `bson:"name" json:"name"`
	Age  int32   `bson:"age" json:"age"`
}

func (t *TestResource) GetID() *string {
	return t.ID
}

func (t *TestResource) SetID(id string) {
	t.ID = &id
}

// Test Specialized Model for category-specific fields
type TestSpecializedResource struct {
	ID   *string `bson:"_id,omitempty" json:"id,omitempty"`
	Name string  `bson:"name" json:"name"`
	Age  int32   `bson:"age" json:"age"`
	Type *string `bson:"categoryType,omitempty" json:"categoryType,omitempty"`
}

func (t *TestSpecializedResource) GetID() *string {
	return t.ID
}

func (t *TestSpecializedResource) SetID(id string) {
	t.ID = &id
}

func (t *TestSpecializedResource) GetCategoryType() *string {
	return t.Type
}

func (t *TestSpecializedResource) SetCategoryType(categoryType string) {
	t.Type = &categoryType
}

type TestSpecializedResourceCategory struct {
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
	assert.Equal(t, 24, len(id))
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
	converter := &DocumentConverter[*TestResource]{}

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
	converter := &DocumentConverter[*TestResource]{}
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
	assert.Equal(t, testID.Hex(), *model.ID)
}

func TestDocumentConverter_ToDocument(t *testing.T) {
	converter := &DocumentConverter[*TestResource]{}
	idConverter := &StringIDConverter{}

	testID := "test-123"
	model := &TestResource{
		ID:   &testID,
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

// Tests for SpecializedDocumentConverter
func TestSpecializedDocumentConverter_ExtractMetadata(t *testing.T) {
	converter := &SpecializedDocumentConverter[*TestSpecializedResource, TestSpecializedResourceCategory]{}

	t.Run("with metadata as bson.M", func(t *testing.T) {
		raw := bson.M{
			"_id":  "123",
			"name": "Test",
			"metadata": bson.M{
				"key1": "value1",
				"key2": int32(42),
			},
		}

		metadata, err := converter.ExtractMetadata(raw)
		assert.NoError(t, err)
		assert.Equal(t, "value1", metadata["key1"])
		assert.Equal(t, int32(42), metadata["key2"])
	})

	t.Run("with metadata as map[string]interface{}", func(t *testing.T) {
		raw := bson.M{
			"_id":  "123",
			"name": "Test",
			"metadata": map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		}

		metadata, err := converter.ExtractMetadata(raw)
		assert.NoError(t, err)
		assert.Equal(t, "value1", metadata["key1"])
		assert.Equal(t, 42, metadata["key2"])
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

	t.Run("with nil metadata", func(t *testing.T) {
		raw := bson.M{
			"_id":      "123",
			"name":     "Test",
			"metadata": nil,
		}

		metadata, err := converter.ExtractMetadata(raw)
		assert.NoError(t, err)
		assert.Empty(t, metadata)
	})
}

func TestSpecializedDocumentConverter_ToSpecialized(t *testing.T) {
	converter := &SpecializedDocumentConverter[*TestSpecializedResource, TestSpecializedResourceCategory]{}
	idConverter := &StringIDConverter{}

	t.Run("complete specialized resource", func(t *testing.T) {
		raw := bson.M{
			"_id":          "test-123",
			"name":         "TestResource",
			"age":          int32(25),
			"categoryType": "premium",
			"extraField":   "extra-data",
			"priority":     int32(10),
			"metadata": bson.M{
				"version":   1,
				"createdBy": "system",
			},
		}

		specialized, err := converter.ToSpecialized(raw, idConverter)
		assert.NoError(t, err)
		assert.NotNil(t, specialized)

		// Test This (base resource)
		assert.Equal(t, "test-123", *specialized.This.ID)
		assert.Equal(t, "TestResource", specialized.This.Name)

		// Test CategoryThis (specialized fields)
		assert.NotNil(t, specialized.This.Type)
		if specialized.This.Type != nil {
			assert.Equal(t, "premium", *specialized.This.Type)
		}
		assert.Equal(t, "extra-data", specialized.CategoryThis.ExtraField)

		// Test Metadata
		assert.Equal(t, 1, specialized.Metadata["version"])
		assert.Equal(t, "system", specialized.Metadata["createdBy"])
	})

	t.Run("with ObjectID", func(t *testing.T) {
		idConverter := &ObjectIDConverter{}
		testObjectID := primitive.NewObjectID()

		raw := bson.M{
			"_id":        testObjectID,
			"name":       "TestResource",
			"age":        int32(30),
			"extraField": "test-data",
			"priority":   int32(5),
		}

		specialized, err := converter.ToSpecialized(raw, idConverter)
		assert.NoError(t, err)
		assert.Equal(t, testObjectID.Hex(), *specialized.This.ID)
		assert.Equal(t, "TestResource", specialized.This.Name)
		assert.Equal(t, "test-data", specialized.CategoryThis.ExtraField)
	})

	t.Run("with empty metadata", func(t *testing.T) {
		raw := bson.M{
			"_id":        "test-456",
			"name":       "MinimalResource",
			"age":        int32(20),
			"extraField": "minimal",
			"priority":   int32(1),
		}

		specialized, err := converter.ToSpecialized(raw, idConverter)
		assert.NoError(t, err)
		assert.Empty(t, specialized.Metadata)
	})

	t.Run("invalid ObjectID should return error", func(t *testing.T) {
		idConverter := &ObjectIDConverter{}
		raw := bson.M{
			"_id":        "invalid-object-id",
			"name":       "TestResource",
			"age":        int32(25),
			"extraField": "test",
			"priority":   int32(1),
		}

		_, err := converter.ToSpecialized(raw, idConverter)
		assert.Error(t, err)
	})
}

func TestSpecializedDocumentConverter_ToDocument(t *testing.T) {
	converter := &SpecializedDocumentConverter[*TestSpecializedResource, TestSpecializedResourceCategory]{}
	idConverter := &StringIDConverter{}

	t.Run("complete specialized resource to document", func(t *testing.T) {
		testID := "test-789"
		categoryType := "premium"

		specialized := sdk.ResourceInstanceSpecialized[*TestSpecializedResource, TestSpecializedResourceCategory]{
			This: &TestSpecializedResource{
				ID:   &testID,
				Name: "TestDoc",
				Age:  30,
				Type: &categoryType,
			},
			CategoryThis: TestSpecializedResourceCategory{
				ExtraField: "extra-value",
				Priority:   10,
			},
			Metadata: map[string]interface{}{
				"version":    2,
				"lastUpdate": "2023-12-06",
			},
		}

		doc, err := converter.ToDocument(specialized, idConverter)
		assert.NoError(t, err)

		// Test merged fields
		assert.Equal(t, "test-789", doc["_id"])
		assert.Equal(t, "TestDoc", doc["name"])
		assert.Equal(t, int32(30), doc["age"])
		assert.Equal(t, "premium", doc["categoryType"])
		assert.Equal(t, "extra-value", doc["extraField"])
		assert.Equal(t, int32(10), doc["priority"])

		// Test metadata
		metadata := doc["metadata"].(map[string]interface{})
		assert.Equal(t, 2, metadata["version"])
		assert.Equal(t, "2023-12-06", metadata["lastUpdate"])
	})

	t.Run("with ObjectID converter", func(t *testing.T) {
		idConverter := &ObjectIDConverter{}
		testID := primitive.NewObjectID().Hex()

		specialized := sdk.ResourceInstanceSpecialized[*TestSpecializedResource, TestSpecializedResourceCategory]{
			This: &TestSpecializedResource{
				ID:   &testID,
				Name: "ObjectIDTest",
				Age:  25,
			},
			CategoryThis: TestSpecializedResourceCategory{
				ExtraField: "object-id-test",
				Priority:   5,
			},
		}

		doc, err := converter.ToDocument(specialized, idConverter)
		assert.NoError(t, err)

		// Verify ObjectID conversion
		objectID, ok := doc["_id"].(primitive.ObjectID)
		assert.True(t, ok)
		assert.Equal(t, testID, objectID.Hex())
	})

	t.Run("without metadata", func(t *testing.T) {
		testID := "test-no-meta"

		specialized := sdk.ResourceInstanceSpecialized[*TestSpecializedResource, TestSpecializedResourceCategory]{
			This: &TestSpecializedResource{
				ID:   &testID,
				Name: "NoMeta",
				Age:  20,
			},
			CategoryThis: TestSpecializedResourceCategory{
				ExtraField: "no-meta-test",
				Priority:   1,
			},
			Metadata: nil,
		}

		doc, err := converter.ToDocument(specialized, idConverter)
		assert.NoError(t, err)
		assert.Nil(t, doc["metadata"])
	})

	t.Run("CategoryThis fields override This fields", func(t *testing.T) {
		testID := "test-override"
		categoryType := "premium"

		specialized := sdk.ResourceInstanceSpecialized[*TestSpecializedResource, TestSpecializedResourceCategory]{
			This: &TestSpecializedResource{
				ID:   &testID,
				Name: "OriginalName",
				Age:  30,
				Type: &categoryType,
			},
			CategoryThis: TestSpecializedResourceCategory{
				ExtraField: "override-test",
				Priority:   10,
			},
		}

		doc, err := converter.ToDocument(specialized, idConverter)
		assert.NoError(t, err)

		// Verify fields from both This and CategoryThis
		assert.Equal(t, "test-override", doc["_id"])
		assert.Equal(t, "OriginalName", doc["name"])        // from This
		assert.Equal(t, int32(30), doc["age"])              // from This
		assert.Equal(t, "premium", doc["categoryType"])     // from This
		assert.Equal(t, "override-test", doc["extraField"]) // from CategoryThis
		assert.Equal(t, int32(10), doc["priority"])         // from CategoryThis
	})

	t.Run("invalid ObjectID should return error", func(t *testing.T) {
		idConverter := &ObjectIDConverter{}
		invalidID := "invalid-object-id"

		specialized := sdk.ResourceInstanceSpecialized[*TestSpecializedResource, TestSpecializedResourceCategory]{
			This: &TestSpecializedResource{
				ID:   &invalidID,
				Name: "InvalidID",
				Age:  25,
			},
			CategoryThis: TestSpecializedResourceCategory{
				ExtraField: "invalid-test",
				Priority:   1,
			},
		}

		_, err := converter.ToDocument(specialized, idConverter)
		assert.Error(t, err)
	})
}
