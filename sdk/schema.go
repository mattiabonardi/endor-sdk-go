package sdk

import (
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
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

type SchemaFormatName string

const (
	DateTimeFormat     SchemaFormatName = "date-time"
	DateFormat         SchemaFormatName = "date"
	TimeFormat         SchemaFormatName = "time"
	EmailFormat        SchemaFormatName = "email"
	HostnameFormat     SchemaFormatName = "hostname"
	IPv4Format         SchemaFormatName = "ipv4"
	IPv6Format         SchemaFormatName = "ipv6"
	URIFormat          SchemaFormatName = "uri"
	UUIDFormat         SchemaFormatName = "uuid"
	PasswordFormat     SchemaFormatName = "password"
	CountryCodeFormat  SchemaFormatName = "country-code"  // ISO 3166-1 alpha-2 country code
	LanguageCodeFormat SchemaFormatName = "language-code" // Language tag (e.g., en-US)
	CurrencyFormat     SchemaFormatName = "currency"      // Currency code (e.g., USD, EUR)
	YAMLFormat         SchemaFormatName = "yaml"
	JSONFormat         SchemaFormatName = "json"
)

func NewSchemaFormat(f SchemaFormatName) *SchemaFormatName {
	return &f
}

type Schema struct {
	Reference   string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type        SchemaTypeName     `json:"type,omitempty" yaml:"type,omitempty"`
	Properties  *map[string]Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	Enum        *[]string          `json:"enum,omitempty" yaml:"enum,omitempty"`
	Title       *string            `json:"title,omitempty" yaml:"title,omitempty"`
	Description *string            `json:"description,omitempty" yaml:"description,omitempty"`
	Format      *SchemaFormatName  `json:"format,omitempty" yaml:"format,omitempty"`
	ReadOnly    *bool              `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`

	UISchema *UISchema `json:"x-ui,omitempty" yaml:"x-ui,omitempty"`
}

type UISchema struct {
	Resource *string   `json:"resource,omitempty" yaml:"resource,omitempty"` // define the reference resource
	Order    *[]string `json:"order,omitempty" yaml:"order,omitempty"`       // define the order of the attributes
	Id       *string   `json:"id,omitempty" yaml:"id,omitempty"`             // define the property that refers to id
}

type RootSchema struct {
	Schema      `json:",inline" yaml:",inline"`
	Definitions map[string]Schema `json:"$defs,omitempty" yaml:"$defs,omitempty"`
}

func (h *RootSchema) ToYAML() (string, error) {
	yamlData, err := yaml.Marshal(&h)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
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

	schema.UISchema = &UISchema{
		Order: &[]string{},
	}
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

		// add field to order
		*schema.UISchema.Order = append(*schema.UISchema.Order, name)

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
	originalName := t.Name()
	if strings.Contains(originalName, "[") {
		// Extract the part before the brackets
		leftBracket := strings.LastIndex(originalName, "[")

		// Get last segment before `[`
		before := originalName[:leftBracket]
		beforeParts := strings.Split(before, "/")
		nameBefore := beforeParts[len(beforeParts)-1]

		// Extract the part inside the brackets
		rightBracket := strings.LastIndex(originalName, "]")

		inside := originalName[leftBracket+1 : rightBracket]
		insideParts := strings.FieldsFunc(inside, func(r rune) bool {
			return r == '/' || r == '.'
		})
		nameInside := insideParts[len(insideParts)-1]

		return nameBefore + "_" + nameInside
	}
	return originalName
}
