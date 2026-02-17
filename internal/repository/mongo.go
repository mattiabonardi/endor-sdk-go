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

// ============================================================================
// SECTION 1: Utility Functions
// ============================================================================

// idToString converts an ID of any type (string, ObjectID) to a string representation.
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

// isIDEmpty checks if an ID is empty (nil, empty string, or empty ObjectID).
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

// cloneBsonM creates a shallow copy of a bson.M map.
func cloneBsonM(src map[string]interface{}) bson.M {
	dst := make(bson.M, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// ============================================================================
// SECTION 2: ID Strategy
// ============================================================================
// IDStrategy handles the conversion and generation of document _id fields.
// MongoDB can store _id as either primitive.ObjectID or string.
// This interface abstracts that choice, allowing repositories to work with
// string IDs internally while storing them appropriately in MongoDB.

// IDStrategy defines how _id values are handled in MongoDB.
type IDStrategy interface {
	// CreateFilter creates a MongoDB filter for finding documents by _id.
	CreateFilter(id string) (bson.M, error)

	// ToStorageFormat converts a string ID to MongoDB storage format.
	ToStorageFormat(id string) (interface{}, error)

	// FromStorageFormat converts a MongoDB _id back to string format.
	FromStorageFormat(storageID interface{}) (string, error)

	// GenerateID creates a new unique ID.
	GenerateID() interface{}
}

// ObjectIDStrategy stores _id as primitive.ObjectID in MongoDB.
// Use this when you want MongoDB's native ObjectID format.
type ObjectIDStrategy struct{}

func (s *ObjectIDStrategy) CreateFilter(id string) (bson.M, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ObjectID format: %w", err)
	}
	return bson.M{"_id": oid}, nil
}

func (s *ObjectIDStrategy) ToStorageFormat(id string) (interface{}, error) {
	return primitive.ObjectIDFromHex(id)
}

func (s *ObjectIDStrategy) FromStorageFormat(storageID interface{}) (string, error) {
	oid, ok := storageID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("expected primitive.ObjectID, got %T", storageID)
	}
	return oid.Hex(), nil
}

func (s *ObjectIDStrategy) GenerateID() interface{} {
	return primitive.NewObjectID()
}

// StringIDStrategy stores _id as string in MongoDB.
// Use this when you need custom string-based IDs.
type StringIDStrategy struct{}

func (s *StringIDStrategy) CreateFilter(id string) (bson.M, error) {
	return bson.M{"_id": id}, nil
}

func (s *StringIDStrategy) ToStorageFormat(id string) (interface{}, error) {
	return id, nil
}

func (s *StringIDStrategy) FromStorageFormat(storageID interface{}) (string, error) {
	str, ok := storageID.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", storageID)
	}
	return str, nil
}

func (s *StringIDStrategy) GenerateID() interface{} {
	// Generate a new ObjectID hex string for auto-generated string IDs
	return primitive.NewObjectID().Hex()
}

// detectIDStrategy determines the appropriate IDStrategy for a model type.
// It inspects the _id field's type: sdk.ObjectID uses ObjectIDStrategy,
// everything else uses StringIDStrategy.
// It recursively handles embedded structs with bson:",inline" tags.
func detectIDStrategy[T sdk.EntityInstanceInterface]() IDStrategy {
	var zero T
	t := reflect.TypeOf(zero)

	if t == nil {
		return &StringIDStrategy{}
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return &StringIDStrategy{}
	}

	if strategy := findIDFieldStrategy(t); strategy != nil {
		return strategy
	}

	return &StringIDStrategy{}
}

// findIDFieldStrategy recursively searches for the _id field in a struct type,
// handling inline embedded structs. Returns nil if _id field is not found.
func findIDFieldStrategy(t reflect.Type) IDStrategy {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		bsonTag := field.Tag.Get("bson")

		// Handle embedded structs with inline tag
		isInline := field.Anonymous || bsonTag == ",inline" ||
			strings.HasSuffix(bsonTag, ",inline")

		if isInline {
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct {
				if strategy := findIDFieldStrategy(embeddedType); strategy != nil {
					return strategy
				}
			}
			continue
		}

		// Check if this is the _id field
		if bsonTag == "_id" || strings.HasPrefix(bsonTag, "_id,") {
			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}

			// Check if it's sdk.ObjectID
			if fieldType.PkgPath() == "github.com/mattiabonardi/endor-sdk-go/pkg/sdk" &&
				fieldType.Name() == "ObjectID" {
				return &ObjectIDStrategy{}
			}
			return &StringIDStrategy{}
		}
	}

	return nil
}

// ============================================================================
// SECTION 3: ObjectID Field Handling
// ============================================================================
// sdk.ObjectID fields (other than _id) must be converted to primitive.ObjectID
// when storing in MongoDB and back when reading. This section handles that conversion.

// ObjectIDFieldRegistry tracks which fields in a model are of type sdk.ObjectID.
// This enables automatic conversion when reading/writing documents.
type ObjectIDFieldRegistry struct {
	fields map[string]struct{}
}

// NewObjectIDFieldRegistry creates a registry by inspecting the model's struct fields.
// It recursively handles embedded structs with bson:",inline" tags.
func NewObjectIDFieldRegistry[T any]() *ObjectIDFieldRegistry {
	registry := &ObjectIDFieldRegistry{
		fields: make(map[string]struct{}),
	}

	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		return registry
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return registry
	}

	registry.collectObjectIDFieldsWithPrefix(t, "")
	return registry
}

// collectObjectIDFieldsWithPrefix recursively collects ObjectID fields from a struct type,
// building full paths for nested structs (e.g., "lines.productId").
// Uses JSON tags for field names, except for _id which uses BSON tag.
func (r *ObjectIDFieldRegistry) collectObjectIDFieldsWithPrefix(t reflect.Type, prefix string) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		bsonTag := field.Tag.Get("bson")
		jsonTag := field.Tag.Get("json")

		// Handle embedded structs with inline tag (check both json and bson)
		isInline := field.Anonymous ||
			bsonTag == ",inline" || strings.HasSuffix(bsonTag, ",inline") ||
			jsonTag == ",inline" || strings.HasSuffix(jsonTag, ",inline")

		if isInline {
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct {
				r.collectObjectIDFieldsWithPrefix(embeddedType, prefix)
			}
			continue
		}

		// For _id field, use bson tag; for all other fields, use json tag
		var tagName string
		if bsonTag == "_id" || strings.HasPrefix(bsonTag, "_id,") {
			tagName = "_id"
		} else {
			tagName = extractTagName(jsonTag, field.Name)
		}

		if tagName == "" || tagName == "-" {
			continue
		}

		fullPath := tagName
		if prefix != "" {
			fullPath = prefix + "." + tagName
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Check if field is sdk.ObjectID
		if fieldType.PkgPath() == "github.com/mattiabonardi/endor-sdk-go/pkg/sdk" &&
			fieldType.Name() == "ObjectID" {
			r.fields[fullPath] = struct{}{}
			continue
		}

		// Get element type for slices/arrays
		elementType := fieldType
		if elementType.Kind() == reflect.Slice || elementType.Kind() == reflect.Array {
			elementType = elementType.Elem()
		}
		if elementType.Kind() == reflect.Ptr {
			elementType = elementType.Elem()
		}

		// Recurse into nested structs (including slice/array elements)
		if elementType.Kind() == reflect.Struct {
			// Skip standard library types and MongoDB types
			pkgPath := elementType.PkgPath()
			if pkgPath == "" || pkgPath == "time" || strings.HasPrefix(pkgPath, "go.mongodb.org/") {
				continue
			}
			r.collectObjectIDFieldsWithPrefix(elementType, fullPath)
		}
	}
}

// extractTagName extracts the field name from a tag (json or bson).
// Example: "fieldName,omitempty" -> "fieldName"
func extractTagName(tag, defaultName string) string {
	if tag == "" || tag == "-" {
		return defaultName
	}
	if idx := strings.Index(tag, ","); idx != -1 {
		if idx == 0 {
			return defaultName
		}
		return tag[:idx]
	}
	return tag
}

// ConvertToStorage converts sdk.ObjectID string values to primitive.ObjectID
// for fields registered in this registry. This should be called before
// inserting/updating documents in MongoDB.
func (r *ObjectIDFieldRegistry) ConvertToStorage(doc bson.M) error {
	for field := range r.fields {
		val, exists := doc[field]
		if !exists || val == nil {
			continue
		}

		var hexStr string
		switch v := val.(type) {
		case string:
			hexStr = v
		case sdk.ObjectID:
			hexStr = v.String()
		default:
			continue
		}

		if hexStr == "" {
			continue
		}

		oid, err := primitive.ObjectIDFromHex(hexStr)
		if err != nil {
			return fmt.Errorf("field %s: invalid ObjectID format: %w", field, err)
		}
		doc[field] = oid
	}
	return nil
}

// ConvertFilterToStorage converts sdk.ObjectID string values to primitive.ObjectID
// in MongoDB query filters. Handles query operators like $in, $gt, $eq, etc.,
// and nested conditions like $or, $and.
func (r *ObjectIDFieldRegistry) ConvertFilterToStorage(filter bson.M) error {
	for key, val := range filter {
		if val == nil {
			continue
		}

		// Logical operators: recurse into sub-documents
		if key == "$or" || key == "$and" || key == "$nor" {
			if err := r.convertInConditions(val); err != nil {
				return err
			}
			continue
		}

		// ObjectID field: convert value recursively
		if r.IsObjectIDField(key) {
			converted, err := r.convertAny(val)
			if err != nil {
				return fmt.Errorf("field %s: %w", key, err)
			}
			filter[key] = converted
		}
	}
	return nil
}

// convertInConditions recursively processes $or/$and/$nor arrays
func (r *ObjectIDFieldRegistry) convertInConditions(val interface{}) error {
	switch conditions := val.(type) {
	case []interface{}:
		for i, cond := range conditions {
			switch c := cond.(type) {
			case bson.M:
				if err := r.ConvertFilterToStorage(c); err != nil {
					return err
				}
			case map[string]interface{}:
				bsonCond := bson.M(c)
				if err := r.ConvertFilterToStorage(bsonCond); err != nil {
					return err
				}
				conditions[i] = bsonCond
			}
		}
	case []bson.M:
		for _, cond := range conditions {
			if err := r.ConvertFilterToStorage(cond); err != nil {
				return err
			}
		}
	}
	return nil
}

// convertAny recursively converts any value that might contain ObjectID strings
func (r *ObjectIDFieldRegistry) convertAny(val interface{}) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	switch v := val.(type) {
	case string:
		if v == "" {
			return v, nil
		}
		oid, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return nil, fmt.Errorf("invalid ObjectID format: %w", err)
		}
		return oid, nil

	case sdk.ObjectID:
		if v.IsEmpty() {
			return v, nil
		}
		oid, err := primitive.ObjectIDFromHex(v.String())
		if err != nil {
			return nil, fmt.Errorf("invalid ObjectID format: %w", err)
		}
		return oid, nil

	case primitive.ObjectID:
		return v, nil

	case bson.M:
		for k, inner := range v {
			converted, err := r.convertAny(inner)
			if err != nil {
				return nil, fmt.Errorf("operator %s: %w", k, err)
			}
			v[k] = converted
		}
		return v, nil

	case map[string]interface{}:
		return r.convertAny(bson.M(v))

	case []interface{}:
		for i, item := range v {
			converted, err := r.convertAny(item)
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			v[i] = converted
		}
		return v, nil

	case []string:
		result := make([]interface{}, len(v))
		for i, s := range v {
			converted, err := r.convertAny(s)
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			result[i] = converted
		}
		return result, nil

	default:
		return val, nil
	}
}

// IsObjectIDField returns true if the given field name is an ObjectID field.
func (r *ObjectIDFieldRegistry) IsObjectIDField(fieldName string) bool {
	_, exists := r.fields[fieldName]
	return exists
}

// ============================================================================
// SECTION 4: Document Mapper
// ============================================================================
// DocumentMapper handles conversion between Go structs and MongoDB documents.

// DocumentMapper converts between model types and MongoDB documents.
// It handles:
// - Embedded structs with bson:",inline" tags
// - Automatic ObjectID conversion via the model's MarshalBSONValue/UnmarshalBSONValue
type DocumentMapper[T sdk.EntityInstanceInterface] struct{}

// ToModel converts a raw MongoDB document to the model type T.
// ObjectID fields are automatically converted via UnmarshalBSONValue.
func (m *DocumentMapper[T]) ToModel(doc bson.M) (T, error) {
	var model T

	entityBytes, err := bson.Marshal(doc)
	if err != nil {
		return model, fmt.Errorf("failed to marshal document: %w", err)
	}

	if err := bson.Unmarshal(entityBytes, &model); err != nil {
		return model, fmt.Errorf("failed to unmarshal to model: %w", err)
	}

	return model, nil
}

// ToDocument converts a model with metadata to a MongoDB document.
// All fields (model + metadata) are stored at root level.
// The idStrategy is used to ensure proper _id storage format.
func (m *DocumentMapper[T]) ToDocument(model T, metadata map[string]interface{}, idStrategy IDStrategy) (bson.M, error) {
	// Marshal model - ObjectID fields auto-convert via MarshalBSONValue
	entityBytes, err := bson.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model: %w", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(entityBytes, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to document: %w", err)
	}

	// Ensure _id is in correct storage format
	idStr := idToString(model.GetID())
	if idStr != "" {
		storageID, err := idStrategy.ToStorageFormat(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert _id: %w", err)
		}
		doc["_id"] = storageID
	}

	// Add metadata fields at root level (inline)
	for k, v := range metadata {
		doc[k] = v
	}

	return doc, nil
}

// ToDocumentWithoutMetadata converts a model to a MongoDB document without metadata.
// Used by StaticEntityInstanceRepository where metadata is not needed.
func (m *DocumentMapper[T]) ToDocumentWithoutMetadata(model T, idStrategy IDStrategy) (bson.M, error) {
	entityBytes, err := bson.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model: %w", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(entityBytes, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to document: %w", err)
	}

	return doc, nil
}

// ============================================================================
// SECTION 6: Base Repository
// ============================================================================
// mongoBaseRepository provides common MongoDB CRUD operations used by both
// MongoEntityInstanceRepository and MongoStaticEntityInstanceRepository.

type mongoBaseRepository[T sdk.EntityInstanceInterface] struct {
	collection     *mongo.Collection
	idStrategy     IDStrategy
	autoGenerateID bool
	objectIDFields *ObjectIDFieldRegistry
	documentMapper *DocumentMapper[T]
}

func newMongoBaseRepository[T sdk.EntityInstanceInterface](
	collection *mongo.Collection,
	autoGenerateID bool,
) *mongoBaseRepository[T] {
	return &mongoBaseRepository[T]{
		collection:     collection,
		idStrategy:     detectIDStrategy[T](),
		autoGenerateID: autoGenerateID,
		objectIDFields: NewObjectIDFieldRegistry[T](),
		documentMapper: &DocumentMapper[T]{},
	}
}

// FindByID retrieves a single document by its _id.
func (r *mongoBaseRepository[T]) FindByID(ctx context.Context, id string) (bson.M, error) {
	filter, err := r.idStrategy.CreateFilter(id)
	if err != nil {
		return nil, sdk.NewBadRequestError(err)
	}

	var result bson.M
	err = r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sdk.NewNotFoundError(fmt.Errorf("entity with id %s not found", id))
		}
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to find entity: %w", err))
	}

	return result, nil
}

// Find retrieves documents matching the filter with optional projection.
func (r *mongoBaseRepository[T]) Find(ctx context.Context, filter bson.M, projection bson.M) ([]bson.M, error) {
	mongoFilter := filter
	if mongoFilter == nil {
		mongoFilter = bson.M{}
	} else {
		mongoFilter = cloneBsonM(mongoFilter)
	}

	// Convert ObjectID fields in filter to primitive.ObjectID
	if err := r.objectIDFields.ConvertFilterToStorage(mongoFilter); err != nil {
		return nil, sdk.NewBadRequestError(err)
	}

	var opts *options.FindOptions
	if projection != nil {
		opts = options.Find().SetProjection(projection)
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to find entities: %w", err))
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to decode entities: %w", err))
	}

	return results, nil
}

// Insert creates a new document. If autoGenerateID is true, a new ID is generated.
// Otherwise, providedID must be non-empty.
func (r *mongoBaseRepository[T]) Insert(ctx context.Context, doc bson.M, providedID any) (string, error) {
	var idStr string

	if r.autoGenerateID {
		doc["_id"] = r.idStrategy.GenerateID()
	} else {
		if isIDEmpty(providedID) {
			return "", sdk.NewBadRequestError(fmt.Errorf("ID is required when auto-generation is disabled"))
		}
		idStr = idToString(providedID)

		// Check for existing document
		_, err := r.FindByID(ctx, idStr)
		if err == nil {
			return "", sdk.NewConflictError(fmt.Errorf("entity with id %s already exists", idStr))
		}
		var endorErr *sdk.EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			return "", err
		}

		storageID, err := r.idStrategy.ToStorageFormat(idStr)
		if err != nil {
			return "", sdk.NewBadRequestError(err)
		}
		doc["_id"] = storageID
	}

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", sdk.NewConflictError(fmt.Errorf("entity already exists: %w", err))
		}
		return "", sdk.NewInternalServerError(fmt.Errorf("failed to create entity: %w", err))
	}

	// Return the string ID
	if r.autoGenerateID {
		idStr, _ = r.idStrategy.FromStorageFormat(result.InsertedID)
	}
	return idStr, nil
}

// Update modifies an existing document by its _id.
func (r *mongoBaseRepository[T]) Update(ctx context.Context, id string, updateData bson.M) error {
	// Verify existence
	if _, err := r.FindByID(ctx, id); err != nil {
		return err
	}

	filter, err := r.idStrategy.CreateFilter(id)
	if err != nil {
		return sdk.NewBadRequestError(err)
	}

	if len(updateData) == 0 {
		return sdk.NewBadRequestError(fmt.Errorf("no fields to update"))
	}

	// Clone and convert ObjectID fields
	data := cloneBsonM(updateData)
	if err := r.objectIDFields.ConvertToStorage(data); err != nil {
		return sdk.NewBadRequestError(err)
	}

	update := bson.M{"$set": data}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return sdk.NewInternalServerError(fmt.Errorf("failed to update entity: %w", err))
	}
	if result.MatchedCount == 0 {
		return sdk.NewNotFoundError(fmt.Errorf("entity with id %s not found", id))
	}

	return nil
}

// Delete removes a document by its _id.
func (r *mongoBaseRepository[T]) Delete(ctx context.Context, id string) error {
	// Verify existence
	if _, err := r.FindByID(ctx, id); err != nil {
		return err
	}

	filter, err := r.idStrategy.CreateFilter(id)
	if err != nil {
		return sdk.NewBadRequestError(err)
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return sdk.NewInternalServerError(fmt.Errorf("failed to delete entity: %w", err))
	}
	if result.DeletedCount == 0 {
		return sdk.NewNotFoundError(fmt.Errorf("entity with id %s not found", id))
	}

	return nil
}

// GetIDStrategy returns the ID strategy used by this repository.
func (r *mongoBaseRepository[T]) GetIDStrategy() IDStrategy {
	return r.idStrategy
}

// GetDocumentMapper returns the document mapper used by this repository.
func (r *mongoBaseRepository[T]) GetDocumentMapper() *DocumentMapper[T] {
	return r.documentMapper
}
