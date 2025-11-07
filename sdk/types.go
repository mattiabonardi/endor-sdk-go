package sdk

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ObjectID è un wrapper per primitive.ObjectID che mantiene lo stesso comportamento con MongoDB
// ma serializza/deserializza in JSON come stringa esadecimale.
type ObjectID primitive.ObjectID

// MarshalJSON serializza l'ObjectID come stringa esadecimale.
func (oid ObjectID) MarshalJSON() ([]byte, error) {
	return json.Marshal(primitive.ObjectID(oid).Hex())
}

// UnmarshalJSON deserializza una stringa esadecimale in ObjectID.
func (oid *ObjectID) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	if hexStr == "" {
		*oid = ObjectID(primitive.NilObjectID)
		return nil
	}
	primitiveOID, err := primitive.ObjectIDFromHex(hexStr)
	if err != nil {
		return err
	}
	*oid = ObjectID(primitiveOID)
	return nil
}

// Hex restituisce la rappresentazione esadecimale dell’ObjectID.
func (oid ObjectID) Hex() string {
	return primitive.ObjectID(oid).Hex()
}

// IsZero ritorna true se l’ObjectID è vuoto.
func (oid ObjectID) IsZero() bool {
	return primitive.ObjectID(oid).IsZero()
}

// ToPrimitive converte il tipo custom in primitive.ObjectID.
func (oid ObjectID) ToPrimitive() primitive.ObjectID {
	return primitive.ObjectID(oid)
}

// NewObjectID genera un nuovo ObjectID.
func NewObjectID() ObjectID {
	return ObjectID(primitive.NewObjectID())
}

// MarshalBSONValue implementa bson.ValueMarshaler per compatibilità nativa con MongoDB.
func (oid ObjectID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(primitive.ObjectID(oid))
}

// UnmarshalBSONValue implementa bson.ValueUnmarshaler per compatibilità nativa con MongoDB.
func (oid *ObjectID) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	var primitiveOID primitive.ObjectID
	if err := bson.UnmarshalValue(t, data, &primitiveOID); err != nil {
		return err
	}
	*oid = ObjectID(primitiveOID)
	return nil
}
