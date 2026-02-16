package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// cloneBsonM performs a shallow copy of a bson.M map
func cloneBsonM(src map[string]interface{}) bson.M {
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
			}
		}
	}
	return nil
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
	return primitive.NewObjectID()
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
	// For string IDs, autogenerate using ObjectID hex representation
	return primitive.NewObjectID().Hex()
}

// detectIDType uses reflection to determine if the model uses string or ObjectID for its ID
// It inspects the struct field with bson:"_id" tag to determine the appropriate storage type
func detectIDType[T sdk.EntityInstanceInterface]() string {
	var zero T
	t := reflect.TypeOf(zero)

	if t == nil {
		return "string" // default
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// For non-struct types, default to string
	if t.Kind() != reflect.Struct {
		return "string"
	}

	// Look for the _id field (BSON tag for MongoDB ID)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		bsonTag := field.Tag.Get("bson")

		// Check if this field is the _id field
		if bsonTag == "_id" || strings.HasPrefix(bsonTag, "_id,") {
			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}

			// Check if it's sdk.ObjectID
			if fieldType.PkgPath() == "github.com/mattiabonardi/endor-sdk-go/pkg/sdk" &&
				fieldType.Name() == "ObjectID" {
				return "objectid"
			}

			return "string"
		}
	}

	return "string" // default if _id field not found
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
		storageID, err := idConverter.ToStorageID(idStr)
		if err != nil {
			return nil, err
		}
		entityMap["_id"] = storageID
	}

	// Add metadata
	entityMap["metadata"] = metadata

	return entityMap, nil
}

// mongoBaseRepository contains common MongoDB operations used by both repository implementations
// This eliminates duplication while allowing each repository to maintain its specific storage format
type mongoBaseRepository[T sdk.EntityInstanceInterface] struct {
	collection     *mongo.Collection
	idConverter    IDConverter
	autoGenerateID bool
	objectIDFields map[string]struct{}
}

func newMongoBaseRepository[T sdk.EntityInstanceInterface](
	collection *mongo.Collection,
	autoGenerateID bool,
) *mongoBaseRepository[T] {
	idType := detectIDType[T]()

	var idConverter IDConverter
	if idType == "objectid" {
		idConverter = &ObjectIDConverter{}
	} else {
		idConverter = &StringIDConverter{}
	}

	return &mongoBaseRepository[T]{
		collection:     collection,
		idConverter:    idConverter,
		autoGenerateID: autoGenerateID,
		objectIDFields: getObjectIDFields[T](),
	}
}

// findOne retrieves a single document by ID
func (r *mongoBaseRepository[T]) findOne(ctx context.Context, id string) (bson.M, error) {
	filter, err := r.idConverter.ToFilter(id)
	if err != nil {
		return nil, sdk.NewBadRequestError(err)
	}

	var result bson.M
	err = r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", id))
		}
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to find entity instance: %w", err))
	}

	return result, nil
}

// find retrieves multiple documents with filter and projection
func (r *mongoBaseRepository[T]) find(ctx context.Context, filter bson.M, projection bson.M) ([]bson.M, error) {
	mongoFilter := filter
	if mongoFilter == nil {
		mongoFilter = bson.M{}
	} else {
		mongoFilter = cloneBsonM(mongoFilter)
	}

	// Convert ObjectID fields in filter to primitive.ObjectID
	if err := convertObjectIDsToStorage(mongoFilter, r.objectIDFields); err != nil {
		return nil, sdk.NewBadRequestError(err)
	}

	var opts *options.FindOptions
	if projection != nil {
		opts = options.Find().SetProjection(projection)
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to list entities: %w", err))
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to decode entities: %w", err))
	}

	return results, nil
}

// insertOne inserts a document with optional auto-generated ID
func (r *mongoBaseRepository[T]) insertOne(ctx context.Context, doc bson.M, providedID any) (string, error) {
	var idStr string

	if r.autoGenerateID {
		doc["_id"] = r.idConverter.GenerateNewID()
	} else {
		if providedID == "" {
			return "", sdk.NewBadRequestError(fmt.Errorf("ID is required when auto-generation is disabled"))
		}
		idStr, ok := providedID.(string)
		if !ok {
			return "", sdk.NewBadRequestError(fmt.Errorf("provided ID must be a string"))
		}

		// Check if exists
		_, err := r.findOne(ctx, idStr)
		if err == nil {
			return "", sdk.NewConflictError(fmt.Errorf("entity instance with id %v already exists", idStr))
		}
		var endorErr *sdk.EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			return "", err
		}

		storageID, err := r.idConverter.ToStorageID(idStr)
		if err != nil {
			return "", sdk.NewBadRequestError(err)
		}
		doc["_id"] = storageID
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", sdk.NewConflictError(fmt.Errorf("entity instance already exists: %w", err))
		}
		return "", sdk.NewInternalServerError(fmt.Errorf("failed to create entity instance: %w", err))
	}

	return idStr, nil
}

// updateOne updates a document by ID with the provided update document
func (r *mongoBaseRepository[T]) updateOne(ctx context.Context, id string, updateData bson.M) error {
	// Verify the instance exists
	_, err := r.findOne(ctx, id)
	if err != nil {
		return err
	}

	filter, err := r.idConverter.ToFilter(id)
	if err != nil {
		return sdk.NewBadRequestError(err)
	}

	if len(updateData) == 0 {
		return sdk.NewBadRequestError(fmt.Errorf("no fields to update"))
	}

	// Clone update data and convert ObjectID fields
	data := cloneBsonM(updateData)
	if err := convertObjectIDsToStorage(data, r.objectIDFields); err != nil {
		return sdk.NewBadRequestError(err)
	}

	// Perform the update with $set
	update := bson.M{"$set": data}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return sdk.NewInternalServerError(fmt.Errorf("failed to update entity instance: %w", err))
	}
	if result.MatchedCount == 0 {
		return sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", id))
	}

	return nil
}

// deleteOne deletes a document by ID
func (r *mongoBaseRepository[T]) deleteOne(ctx context.Context, id string) error {
	// Verify the instance exists
	_, err := r.findOne(ctx, id)
	if err != nil {
		return err
	}

	filter, err := r.idConverter.ToFilter(id)
	if err != nil {
		return sdk.NewBadRequestError(err)
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return sdk.NewInternalServerError(fmt.Errorf("failed to delete entity instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", id))
	}

	return nil
}
