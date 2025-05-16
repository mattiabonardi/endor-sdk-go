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
	Reference  string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type       SchemaTypeName     `json:"type,omitempty" yaml:"type,omitempty"`
	Properties *map[string]Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items      *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	Enum       *[]string          `json:"enum,omitempty" yaml:"enum,omitempty"`
}

type RootSchema struct {
	Schema
	Definitions map[string]Schema `json:"$defs,omitempty" yaml:"$defs,omitempty"`
}

func NewSchema(model any) *RootSchema {
	t := reflect.TypeOf(model)
	return NewSchemaByType(t)
}

func NewSchemaByType(t reflect.Type) *RootSchema {
	defs := make(map[string]Schema)
	refSchema := buildSchemaWithDefs(t, defs)

	return &RootSchema{
		Schema:      refSchema,
		Definitions: defs,
	}
}

func buildSchemaWithDefs(t reflect.Type, defs map[string]Schema) Schema {
	// Dereference pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	typeName := getTypeName(t)
	if _, ok := defs[typeName]; ok {
		return Schema{
			Reference: "#/$defs/" + typeName,
		}
	}

	if t.Kind() != reflect.Struct {
		panic("buildSchemaWithDefs: input must be a struct or pointer to struct")
	}

	schema := Schema{
		Type:       ObjectType,
		Properties: &map[string]Schema{},
	}

	// Prevent infinite recursion
	defs[typeName] = schema

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
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

		(*schema.Properties)[name] = resolveFieldSchema(field.Type, defs)
	}

	// Update the schema now that it's fully constructed
	defs[typeName] = schema

	return Schema{
		Reference: "#/$defs/" + typeName,
	}
}

func resolveFieldSchema(t reflect.Type, defs map[string]Schema) Schema {
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
		itemSchema := resolveFieldSchema(t.Elem(), defs)
		return Schema{
			Type:  ArrayType,
			Items: &itemSchema,
		}
	case reflect.Struct:
		return buildSchemaWithDefs(t, defs)
	default:
		return Schema{Type: StringType}
	}
}

func getTypeName(t reflect.Type) string {
	// Dereference pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
