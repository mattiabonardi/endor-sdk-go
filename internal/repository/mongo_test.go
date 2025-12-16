package repository

import (
	"testing"

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
