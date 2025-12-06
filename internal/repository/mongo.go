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

	// Remove metadata and _id before converting
	docCopy := cloneBsonM(raw)
	delete(docCopy, "metadata")
	delete(docCopy, "_id")

	resourceBytes, err := bson.Marshal(docCopy)
	if err != nil {
		return model, fmt.Errorf("failed to marshal raw resource: %w", err)
	}

	if err := bson.Unmarshal(resourceBytes, &model); err != nil {
		return model, fmt.Errorf("failed to unmarshal to model: %w", err)
	}

	// Set ID
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

// SpecializedDocumentConverter handles conversions for specialized resources
type SpecializedDocumentConverter[T sdk.ResourceInstanceInterface, C sdk.ResourceInstanceSpecializedInterface] struct{}

func (c *SpecializedDocumentConverter[T, C]) ExtractMetadata(raw bson.M) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})
	if rawMeta, ok := raw["metadata"]; ok && rawMeta != nil {
		switch m := rawMeta.(type) {
		case bson.M:
			for k, v := range m {
				metadata[k] = v
			}
		case map[string]interface{}:
			for k, v := range m {
				metadata[k] = v
			}
		default:
			if b, err := bson.Marshal(m); err == nil {
				_ = bson.Unmarshal(b, &metadata)
			}
		}
	}
	return metadata, nil
}

func (c *SpecializedDocumentConverter[T, C]) ToSpecialized(raw bson.M, idConverter IDConverter) (*sdk.ResourceInstanceSpecialized[T, C], error) {
	// Extract metadata
	metadata, err := c.ExtractMetadata(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Create a copy without metadata
	docNoMeta := cloneBsonM(raw)
	delete(docNoMeta, "metadata")

	// Convert to This (base resource)
	var baseThis T
	{
		b, err := bson.Marshal(docNoMeta)
		if err != nil {
			return nil, fmt.Errorf("marshal docNoMeta: %w", err)
		}
		if err := bson.Unmarshal(b, &baseThis); err != nil {
			return nil, fmt.Errorf("unmarshal into This: %w", err)
		}

		// Set ID from _id field
		if rawId, ok := raw["_id"]; ok {
			idStr, err := idConverter.FromStorageID(rawId)
			if err != nil {
				return nil, err
			}
			if idPtr := baseThis.GetID(); idPtr == nil || *idPtr == "" {
				baseThis.SetID(idStr)
			}
		}
	}

	// Convert to CategoryThis (specialized fields)
	var categoryThis C
	{
		b, err := bson.Marshal(docNoMeta)
		if err != nil {
			return nil, fmt.Errorf("marshal docNoMeta for CategoryThis: %w", err)
		}
		if err := bson.Unmarshal(b, &categoryThis); err != nil {
			return nil, fmt.Errorf("unmarshal into CategoryThis: %w", err)
		}
	}

	return &sdk.ResourceInstanceSpecialized[T, C]{
		This:         baseThis,
		CategoryThis: categoryThis,
		Metadata:     metadata,
	}, nil
}

func (c *SpecializedDocumentConverter[T, C]) ToDocument(data sdk.ResourceInstanceSpecialized[T, C], idConverter IDConverter) (bson.M, error) {
	doc := bson.M{}

	// Merge This fields
	{
		bt, err := bson.Marshal(data.This)
		if err != nil {
			return nil, fmt.Errorf("marshal This: %w", err)
		}
		var tMap bson.M
		if err := bson.Unmarshal(bt, &tMap); err != nil {
			return nil, fmt.Errorf("unmarshal This->map: %w", err)
		}
		for k, v := range tMap {
			doc[k] = v
		}
	}

	// Merge CategoryThis fields (may override This fields)
	{
		bc, err := bson.Marshal(data.CategoryThis)
		if err != nil {
			return nil, fmt.Errorf("marshal CategoryThis: %w", err)
		}
		var cMap bson.M
		if err := bson.Unmarshal(bc, &cMap); err != nil {
			return nil, fmt.Errorf("unmarshal CategoryThis->map: %w", err)
		}
		for k, v := range cMap {
			doc[k] = v
		}
	}

	// Ensure _id is set correctly
	idPtr := data.This.GetID()
	if idPtr != nil && *idPtr != "" {
		storageID, err := idConverter.ToStorageID(*idPtr)
		if err != nil {
			return nil, err
		}
		doc["_id"] = storageID
	}

	// Add metadata as subdocument
	if data.Metadata != nil {
		doc["metadata"] = data.Metadata
	}

	return doc, nil
}
