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
//   - The ID is then explicitly set via SetID() to ensure consistency regardless of
//     the struct's BSON tags
//   - All other fields from embedded structs (with bson:",inline") are correctly
//     preserved during the conversion process
type DocumentConverter[T sdk.ResourceInstanceInterface] struct{}

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
	// We keep _id in the copy so it can be unmarshaled into embedded structs if needed,
	// but we'll explicitly set it afterwards via SetID() to ensure consistency
	docCopy := cloneBsonM(raw)
	delete(docCopy, "metadata")

	// First, unmarshal with _id present (for embedded structs that might have bson:"_id")
	resourceBytes, err := bson.Marshal(docCopy)
	if err != nil {
		return model, fmt.Errorf("failed to marshal raw resource: %w", err)
	}

	if err := bson.Unmarshal(resourceBytes, &model); err != nil {
		return model, fmt.Errorf("failed to unmarshal to model: %w", err)
	}

	// Always explicitly set ID via SetID() to ensure it's properly set
	// regardless of whether the model struct has bson:"_id" tag or not
	if rawID, ok := raw["_id"]; ok {
		idStr, err := idConverter.FromStorageID(rawID)
		if err != nil {
			return model, err
		}
		model.SetID(idStr)
	}

	return model, nil
}

func (c *DocumentConverter[T]) ToDocument(model T, metadata map[string]interface{}, idConverter IDConverter) (bson.M, error) {
	resourceBytes, err := bson.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %w", err)
	}

	var resourceMap bson.M
	if err := bson.Unmarshal(resourceBytes, &resourceMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource: %w", err)
	}

	// Set _id
	idPtr := model.GetID()
	if idPtr != nil && *idPtr != "" {
		storageID, err := idConverter.ToStorageID(*idPtr)
		if err != nil {
			return nil, err
		}
		resourceMap["_id"] = storageID
	}

	// Add metadata
	resourceMap["metadata"] = metadata

	return resourceMap, nil
}
