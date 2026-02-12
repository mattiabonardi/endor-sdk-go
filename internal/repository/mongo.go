package repository

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IDConverter handles ID conversion logic
type IDConverter interface {
	ToFilter(id string) (bson.M, error)
	ToStorageID(id string) (interface{}, error)
	FromStorageID(storageID interface{}) (string, error)
	GenerateNewID() string
}

// helper shallow copy
func cloneBsonM(src bson.M) bson.M {
	dst := make(bson.M, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// ObjectIDConverter for auto-generated MongoDB ObjectIDs
type ObjectIDConverter struct{}

func (c *ObjectIDConverter) ToFilter(id string) (bson.M, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ObjectID format: %w", err)
	}
	return bson.M{"_id": objectID}, nil
}

func (c *ObjectIDConverter) ToStorageID(id string) (interface{}, error) {
	return primitive.ObjectIDFromHex(id)
}

func (c *ObjectIDConverter) FromStorageID(storageID interface{}) (string, error) {
	oid, ok := storageID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("invalid _id type in database")
	}
	return oid.Hex(), nil
}

func (c *ObjectIDConverter) GenerateNewID() string {
	return primitive.NewObjectID().Hex()
}

// StringIDConverter for custom string IDs
type StringIDConverter struct{}

func (c *StringIDConverter) ToFilter(id string) (bson.M, error) {
	return bson.M{"_id": id}, nil
}

func (c *StringIDConverter) ToStorageID(id string) (interface{}, error) {
	return id, nil
}

func (c *StringIDConverter) FromStorageID(storageID interface{}) (string, error) {
	s, ok := storageID.(string)
	if !ok {
		return "", fmt.Errorf("invalid _id type in database")
	}
	return s, nil
}

func (c *StringIDConverter) GenerateNewID() string {
	return "" // No auto-generation for string IDs
}

// DocumentConverter handles BSON <-> Model conversions
//
// This converter properly supports embedded structs (struct composition) when using
// the bson:",inline" tag. Fields from embedded structs are automatically flattened
// to the top level of the BSON document during both marshaling and unmarshaling.
//
// Example:
//
//	type BaseModel struct {
//	    ID string `bson:"_id"`
//	    Name string `bson:"name"`
//	}
//
//	type ExtendedModel struct {
//	    BaseModel `bson:",inline"`  // Fields will be flattened
//	    Email string `bson:"email"`
//	}
//
// The BSON document for ExtendedModel will have fields: _id, name, email (not nested).
//
// Important notes:
//   - The _id field is preserved during ToModel conversion to support embedded structs
//     that have bson:"_id" tags, ensuring all fields in the embedded struct are properly
//     unmarshaled
//   - The ID is automatically set by BSON unmarshal based on the struct's bson:"_id" tag
//   - All other fields from embedded structs (with bson:",inline") are correctly
//     preserved during the conversion process
type DocumentConverter[T sdk.EntityInstanceInterface] struct{}

func (c *DocumentConverter[T]) ExtractMetadata(raw bson.M) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})
	if rawMeta, ok := raw["metadata"]; ok && rawMeta != nil {
		metaBytes, err := bson.Marshal(rawMeta)
		if err != nil {
			return nil, err
		}
		_ = bson.Unmarshal(metaBytes, &metadata)
	}
	return metadata, nil
}

func (c *DocumentConverter[T]) ToModel(raw bson.M, idConverter IDConverter) (T, error) {
	var model T

	// Create a clean copy excluding metadata for unmarshaling
	// We keep _id in the copy so it can be unmarshaled into embedded structs if needed
	docCopy := cloneBsonM(raw)
	delete(docCopy, "metadata")

	// Unmarshal the document - BSON will automatically populate the _id field
	// based on the struct's bson tags (e.g., bson:"_id")
	entityBytes, err := bson.Marshal(docCopy)
	if err != nil {
		return model, fmt.Errorf("failed to marshal raw entity: %w", err)
	}

	if err := bson.Unmarshal(entityBytes, &model); err != nil {
		return model, fmt.Errorf("failed to unmarshal to model: %w", err)
	}

	// The ID is automatically set by BSON unmarshal into the struct field
	// tagged with bson:"_id". No need for explicit SetID() call.

	return model, nil
}

func (c *DocumentConverter[T]) ToDocument(model T, metadata map[string]interface{}, idConverter IDConverter) (bson.M, error) {
	entityBytes, err := bson.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	var entityMap bson.M
	if err := bson.Unmarshal(entityBytes, &entityMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	// Set _id
	idPtr := model.GetID()
	if idPtr != "" {
		storageID, err := idConverter.ToStorageID(idPtr)
		if err != nil {
			return nil, err
		}
		entityMap["_id"] = storageID
	}

	// Add metadata
	entityMap["metadata"] = metadata

	return entityMap, nil
}
