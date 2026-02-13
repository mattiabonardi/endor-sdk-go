package sdk

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// ObjectID is a custom type based on string that represents a MongoDB ObjectID
// It maintains all string properties but is automatically converted to primitive.ObjectID
// in MongoDB repositories during read/write operations.
type ObjectID string

// String returns the string representation of the ObjectID
func (id ObjectID) String() string {
	return string(id)
}

// IsEmpty returns true if the ObjectID is empty
func (id ObjectID) IsEmpty() bool {
	return string(id) == ""
}

// ToPrimitiveObjectID converts the ObjectID to a primitive.ObjectID
// Returns an error if the string is not a valid ObjectID hex string
func (id ObjectID) ToPrimitiveObjectID() (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(string(id))
}

// NewObjectID creates a new ObjectID from a primitive.ObjectID
func NewObjectID(oid primitive.ObjectID) ObjectID {
	return ObjectID(oid.Hex())
}

// NewObjectIDFromString creates a new ObjectID from a string
// Returns an error if the string is not a valid ObjectID hex string
func NewObjectIDFromString(s string) (ObjectID, error) {
	// Validate that it's a valid ObjectID hex string
	_, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		return "", err
	}
	return ObjectID(s), nil
}

// GenerateObjectID generates a new ObjectID
func GenerateObjectID() ObjectID {
	return ObjectID(primitive.NewObjectID().Hex())
}

// MarshalBSONValue implements the bsoncodec.ValueMarshaler interface
// This allows ObjectID to be automatically marshaled as primitive.ObjectID in BSON
func (id ObjectID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if id.IsEmpty() {
		return bsontype.Null, nil, nil
	}

	oid, err := primitive.ObjectIDFromHex(string(id))
	if err != nil {
		return bsontype.Null, nil, err
	}

	return bsontype.ObjectID, bsoncore.AppendObjectID(nil, oid), nil
}

// UnmarshalBSONValue implements the bsoncodec.ValueUnmarshaler interface
// This allows ObjectID to be automatically unmarshaled from primitive.ObjectID in BSON
func (id *ObjectID) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bsontype.Null {
		*id = ""
		return nil
	}

	if t != bsontype.ObjectID {
		return ErrInvalidBSONType
	}

	oid, _, ok := bsoncore.ReadObjectID(data)
	if !ok {
		return ErrInvalidBSONType
	}

	*id = ObjectID(oid.Hex())
	return nil
}

// ErrInvalidBSONType is returned when unmarshaling ObjectID from an invalid BSON type
var ErrInvalidBSONType = fmt.Errorf("invalid BSON type for ObjectID")
