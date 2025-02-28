package models

type SchemaTypeName string

const (
	StringType  SchemaTypeName = "string"
	NumberType  SchemaTypeName = "number"
	BooleanType SchemaTypeName = "boolean"
	ObjectType  SchemaTypeName = "object"
	ArrayType   SchemaTypeName = "array"
)

type Schema struct {
	Type       SchemaTypeName    `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
}
