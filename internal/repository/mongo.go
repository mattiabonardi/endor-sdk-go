package repository

import (
	"fmt"
	"reflect"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IDConverter handles ID conversion logic
type IDConverter interface {
	ToFilter(id any) (bson.M, error)
	ToStorageID(id any) (interface{}, error)
	FromStorageID(storageID interface{}) (any, error)
	GenerateNewID() any
}

// idToString converts an ID of any type (string, ObjectID) to a string
func idToString(id any) string {
	if id == nil {
		return ""
	}
	switch v := id.(type) {
	case string:
		return v
	case sdk.ObjectID:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// isIDEmpty checks if an ID is empty (nil, empty string, or empty ObjectID)
func isIDEmpty(id any) bool {
	if id == nil {
		return true
	}
	switch v := id.(type) {
	case string:
		return v == ""
	case sdk.ObjectID:
		return v == ""
	default:
		return false
	}
}

// helper shallow copy
func cloneBsonM(src bson.M) bson.M {
	dst := make(bson.M, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// getObjectIDFields returns a map of BSON field names that are of type sdk.ObjectID in the model
// This uses reflection to analyze the struct tags and field types
func getObjectIDFields[T any]() map[string]struct{} {
	result := make(map[string]struct{})
	var zero T
	t := reflect.TypeOf(zero)

	if t == nil {
		return result
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return result
	}

	// Recursively collect ObjectID fields, including those from embedded structs
	var collectObjectIDFields func(reflect.Type)
	collectObjectIDFields = func(t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Check if this is an embedded/anonymous field with "inline" tag
			bsonTag := field.Tag.Get("bson")
			isInline := field.Anonymous || bsonTag == ",inline" ||
				(len(bsonTag) >= 7 && bsonTag[len(bsonTag)-7:] == ",inline")

			if isInline && field.Type.Kind() == reflect.Struct {
				// Recursively collect fields from embedded struct
				collectObjectIDFields(field.Type)
			} else if isInline && field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
				// Handle pointer to embedded struct
				collectObjectIDFields(field.Type.Elem())
			} else {
				// Check if field is of type sdk.ObjectID
				fieldType := field.Type
				if fieldType.Kind() == reflect.Ptr {
					fieldType = fieldType.Elem()
				}

				// Check if it's sdk.ObjectID type
				if fieldType.PkgPath() == "github.com/mattiabonardi/endor-sdk-go/pkg/sdk" &&
					fieldType.Name() == "ObjectID" {
					// Extract BSON tag name
					tagName := ""
					if bsonTag != "" && bsonTag != "-" {
						// Remove options like ",omitempty" from tag
						for commaIdx := 0; commaIdx < len(bsonTag); commaIdx++ {
							if bsonTag[commaIdx] == ',' {
								tagName = bsonTag[:commaIdx]
								break
							}
						}
						if tagName == "" {
							tagName = bsonTag
						}
					} else {
						tagName = field.Name
					}

					if tagName != "" && tagName != "-" {
						result[tagName] = struct{}{}
					}
				}
			}
		}
	}

	collectObjectIDFields(t)
	return result
}

// convertObjectIDsToStorage converts sdk.ObjectID fields in a BSON document to primitive.ObjectID
// This is called before storing data in MongoDB
func convertObjectIDsToStorage(doc bson.M, objectIDFields map[string]struct{}) error {
	for field := range objectIDFields {
		if val, exists := doc[field]; exists && val != nil {
			// Convert string to ObjectID
			switch v := val.(type) {
			case string:
				if v != "" {
					oid, err := primitive.ObjectIDFromHex(v)
					if err != nil {
						return fmt.Errorf("field %s: invalid ObjectID format: %w", field, err)
					}
					doc[field] = oid
				}
			case sdk.ObjectID:
				if v != "" {
					oid, err := primitive.ObjectIDFromHex(v.String())
					if err != nil {
						return fmt.Errorf("field %s: invalid ObjectID format: %w", field, err)
					}
					doc[field] = oid
				}
			// primitive.ObjectID is already in the correct format
			case primitive.ObjectID:
				// Already correct, do nothing
			}
		}
	}
	return nil
}

// convertObjectIDsFromStorage converts primitive.ObjectID fields in a BSON document to strings
// This is called after reading data from MongoDB, so sdk.ObjectID fields can be properly unmarshaled
func convertObjectIDsFromStorage(doc bson.M, objectIDFields map[string]struct{}) {
	for field := range objectIDFields {
		if val, exists := doc[field]; exists && val != nil {
			// Convert primitive.ObjectID to string
			if oid, ok := val.(primitive.ObjectID); ok {
				doc[field] = oid.Hex()
			}
		}
	}
}

// ObjectIDConverter for auto-generated MongoDB ObjectIDs
type ObjectIDConverter struct{}

func (c *ObjectIDConverter) ToFilter(id any) (bson.M, error) {
	idStr := idToString(id)
	objectID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid ObjectID format: %w", err)
	}
	return bson.M{"_id": objectID}, nil
}

func (c *ObjectIDConverter) ToStorageID(id any) (interface{}, error) {
	idStr := idToString(id)
	return primitive.ObjectIDFromHex(idStr)
}

func (c *ObjectIDConverter) FromStorageID(storageID interface{}) (any, error) {
	oid, ok := storageID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("invalid _id type in database")
	}
	return oid.Hex(), nil
}

func (c *ObjectIDConverter) GenerateNewID() any {
	return primitive.NewObjectID().Hex()
}

// StringIDConverter for custom string IDs
type StringIDConverter struct{}

func (c *StringIDConverter) ToFilter(id any) (bson.M, error) {
	idStr := idToString(id)
	return bson.M{"_id": idStr}, nil
}

func (c *StringIDConverter) ToStorageID(id any) (interface{}, error) {
	return idToString(id), nil
}

func (c *StringIDConverter) FromStorageID(storageID interface{}) (any, error) {
	s, ok := storageID.(string)
	if !ok {
		return "", fmt.Errorf("invalid _id type in database")
	}
	return s, nil
}

func (c *StringIDConverter) GenerateNewID() any {
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
	// The ObjectID type has MarshalBSONValue/UnmarshalBSONValue methods that handle
	// automatic conversion between primitive.ObjectID and ObjectID
	entityBytes, err := bson.Marshal(docCopy)
	if err != nil {
		return model, fmt.Errorf("failed to marshal raw entity: %w", err)
	}

	if err := bson.Unmarshal(entityBytes, &model); err != nil {
		return model, fmt.Errorf("failed to unmarshal to model: %w", err)
	}

	return model, nil
}

func (c *DocumentConverter[T]) ToDocument(model T, metadata map[string]interface{}, idConverter IDConverter) (bson.M, error) {
	// Marshal the model - ObjectID fields will be automatically converted to primitive.ObjectID
	// via their MarshalBSONValue method
	entityBytes, err := bson.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	var entityMap bson.M
	if err := bson.Unmarshal(entityBytes, &entityMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	// Always ensure _id is properly set using the IDConverter
	// This ensures consistency with the repository's ID handling strategy
	idPtr := model.GetID()
	idStr := idToString(idPtr)
	if idStr != "" {
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
