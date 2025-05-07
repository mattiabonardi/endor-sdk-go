package sdk

import (
	"reflect"
	"strings"
)

type SchemaTypeName string

const (
	StringType  SchemaTypeName = "string"
	IntegerType SchemaTypeName = "integer"
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

func NewSchema(model any) *Schema {
	t := reflect.TypeOf(model)

	// Dereference pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("NewSchema: input must be a struct or pointer to struct")
	}

	schema := Schema{
		Type:       ObjectType,
		Properties: map[string]Schema{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // skip unexported fields
			continue
		}

		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = field.Name
		}
		if name == "-" {
			continue
		}

		schema.Properties[name] = newFieldSchema(field.Type)
	}

	return &schema
}

func newFieldSchema(t reflect.Type) Schema {
	// Dereference pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.PkgPath() == "go.mongodb.org/mongo-driver/bson/primitive" && t.Name() == "ObjectID" {
		return Schema{Type: StringType}
	}

	switch t.Kind() {
	case reflect.String:
		return Schema{Type: StringType}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Schema{Type: IntegerType}
	case reflect.Float32, reflect.Float64:
		return Schema{Type: NumberType}
	case reflect.Bool:
		return Schema{Type: BooleanType}
	case reflect.Slice, reflect.Array:
		itemSchema := newFieldSchema(t.Elem())
		return Schema{
			Type:  ArrayType,
			Items: &itemSchema,
		}
	case reflect.Struct:
		s := NewSchema(reflect.New(t).Interface())
		return *s
	default:
		return Schema{Type: StringType}
	}
}
