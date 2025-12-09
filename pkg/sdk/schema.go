package sdk

import (
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type SchemaTypeName string

const (
	SchemaTypeString  SchemaTypeName = "string"
	SchemaTypeInteger SchemaTypeName = "integer"
	SchemaTypeNumber  SchemaTypeName = "number"
	SchemaTypeBoolean SchemaTypeName = "boolean"
	SchemaTypeObject  SchemaTypeName = "object"
	SchemaTypeArray   SchemaTypeName = "array"
)

type SchemaFormatName string

const (
	SchemaFormatDateTime     SchemaFormatName = "date-time"
	SchemaFormatDate         SchemaFormatName = "date"
	SchemaFormatTime         SchemaFormatName = "time"
	SchemaFormatEmail        SchemaFormatName = "email"
	SchemaFormatHostname     SchemaFormatName = "hostname"
	SchemaFormatIPv4         SchemaFormatName = "ipv4"
	SchemaFormatIPv6         SchemaFormatName = "ipv6"
	SchemaFormatURI          SchemaFormatName = "uri"
	SchemaFormatUUID         SchemaFormatName = "uuid"
	SchemaFormatPassword     SchemaFormatName = "password"
	SchemaFormatCountryCode  SchemaFormatName = "country-code"  // ISO 3166-1 alpha-2 country code
	SchemaFormatLanguageCode SchemaFormatName = "language-code" // Language tag (e.g., en-US)
	SchemaFormatCurrency     SchemaFormatName = "currency"      // Currency code (e.g., USD, EUR)
	SchemaFormatYAML         SchemaFormatName = "yaml"
	SchemaFormatJSON         SchemaFormatName = "json"
	SchemaFormatAsset        SchemaFormatName = "asset"
	SchemaFormatImageAsset   SchemaFormatName = "image-asset"
	SchemaFormatAudioAsset   SchemaFormatName = "audio-asset"
	SchemaFormatVideoAsset   SchemaFormatName = "video-asset"
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
	WriteOnly   *bool              `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`

	// field dimension
	MinLength *int `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength *int `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`

	UISchema *UISchema `json:"x-ui,omitempty" yaml:"x-ui,omitempty"`
}

type UISchema struct {
	Resource *string   `json:"resource,omitempty" yaml:"resource,omitempty"` // define the reference resource
	Order    *[]string `json:"order,omitempty" yaml:"order,omitempty"`       // define the order of the attributes
	Hidden   *bool     `json:"hidden,omitempty" yaml:"hidden,omitempty"`     // define if the property is displayable
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
	// New implementation that expands schemas directly
	visited := make(map[string]bool)
	expandedSchema := buildExpandedSchema(t, visited)

	return &RootSchema{
		Schema:      expandedSchema,
		Definitions: make(map[string]Schema), // Keep empty but present
	}
}

func ResolveGenericSchema[T any]() *RootSchema {
	var zeroT T
	tType := reflect.TypeOf(zeroT)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}
	// convert type to schema
	if tType != nil && tType != reflect.TypeOf(NoPayload{}) {
		return NewSchemaByType(tType)
	}
	return nil
}

func buildExpandedSchema(t reflect.Type, visited map[string]bool) Schema {
	// Dereference pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	typeName := getTypeName(t)

	// Check for infinite recursion - if we've seen this type before, return a simple string schema
	if visited[typeName] {
		return Schema{
			Type:        SchemaTypeString,
			Description: &[]string{"Recursive reference to " + typeName}[0],
		}
	}

	if t.Kind() != reflect.Struct {
		panic("buildExpandedSchema: input must be a struct or pointer to struct")
	}

	// Mark this type as visited
	visited[typeName] = true

	schema := Schema{
		Type:       SchemaTypeObject,
		Properties: &map[string]Schema{},
	}

	schema.UISchema = &UISchema{
		Order: &[]string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		jsonTagParts := strings.Split(jsonTag, ",")
		name := jsonTagParts[0]

		// Check if this is an embedded (anonymous) field with inline tag
		isInline := field.Anonymous && hasInlineTag(jsonTagParts)

		if isInline {
			// Inline the embedded struct's properties
			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct {
				embeddedSchema := buildExpandedSchema(fieldType, visited)
				if embeddedSchema.Properties != nil {
					for k, v := range *embeddedSchema.Properties {
						(*schema.Properties)[k] = v
					}
				}
				// Add embedded struct's field order
				if embeddedSchema.UISchema != nil && embeddedSchema.UISchema.Order != nil {
					*schema.UISchema.Order = append(*schema.UISchema.Order, *embeddedSchema.UISchema.Order...)
				}
			}
			continue
		}

		if name == "" {
			name = field.Name
		}
		if name == "-" {
			continue
		}

		// add field to order
		*schema.UISchema.Order = append(*schema.UISchema.Order, name)

		(*schema.Properties)[name] = resolveExpandedFieldSchema(field, field.Type, visited)
	}

	// Unmark this type as we're done processing it
	visited[typeName] = false

	return schema
}

func resolveExpandedFieldSchema(f reflect.StructField, t reflect.Type, visited map[string]bool) Schema {
	// Dereference pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var schema Schema

	// Handle special types
	if t.PkgPath() == "go.mongodb.org/mongo-driver/bson/primitive" && t.Name() == "ObjectID" {
		schema = Schema{Type: SchemaTypeString}
	} else if t.PkgPath() == "go.mongodb.org/mongo-driver/bson/primitive" && t.Name() == "DateTime" {
		schema = Schema{Type: SchemaTypeString, Format: NewSchemaFormat(SchemaFormatDateTime)}
	} else {
		// Handle built-in kinds
		switch t.Kind() {
		case reflect.String:
			schema = Schema{Type: SchemaTypeString}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			schema = Schema{Type: SchemaTypeInteger}
		case reflect.Float32, reflect.Float64:
			schema = Schema{Type: SchemaTypeNumber}
		case reflect.Bool:
			schema = Schema{Type: SchemaTypeBoolean}
		case reflect.Slice, reflect.Array:
			// Don't recurse with the same field – array element doesn't have tags
			itemSchema := resolveExpandedFieldSchema(reflect.StructField{}, t.Elem(), visited)
			schema = Schema{
				Type:  SchemaTypeArray,
				Items: &itemSchema,
			}
		case reflect.Struct:
			schema = buildExpandedSchema(t, visited)
		default:
			schema = Schema{Type: SchemaTypeString}
		}
	}

	if tag := f.Tag.Get("schema"); tag != "" {
		props := parseSchemaTag(tag)
		applySchemaDecorators(&schema, props)
	}
	if tag := f.Tag.Get("ui-schema"); tag != "" {
		props := parseSchemaTag(tag)
		applyUISchemaDecorators(&schema, props)
	}

	return schema
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

// hasInlineTag checks if the json tag parts contain "inline" option
func hasInlineTag(tagParts []string) bool {
	for _, part := range tagParts[1:] { // Skip the first part (field name)
		if strings.TrimSpace(part) == "inline" {
			return true
		}
	}
	return false
}

func parseSchemaTag(tag string) map[string]string {
	parts := strings.Split(tag, ",")
	props := make(map[string]string)
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			props[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return props
}

func applySchemaDecorators(s *Schema, props map[string]string) {
	for key, val := range props {
		v := val

		switch key {
		// metadata
		case "description":
			s.Description = &v
		case "title":
			s.Title = &v
		case "format":
			f := SchemaFormatName(v)
			s.Format = &f

		// permissions
		case "readOnly":
			if v == "true" {
				boolean := true
				s.ReadOnly = &boolean
			}
		case "writeOnly":
			if v == "true" {
				boolean := true
				s.WriteOnly = &boolean
			}

		// field size
		case "maxLength":
			if i, err := strconv.Atoi(v); err == nil {
				s.MaxLength = &i
			}
		case "minLength":
			if i, err := strconv.Atoi(v); err == nil {
				s.MinLength = &i
			}
		}
	}
}

func applyUISchemaDecorators(s *Schema, props map[string]string) {
	if s.UISchema == nil {
		s.UISchema = &UISchema{}
	}
	for key, val := range props {
		v := val
		switch key {
		case "resource":
			s.UISchema.Resource = &v
		case "hidden":
			if v == "true" {
				trueValue := true
				s.UISchema.Hidden = &trueValue
			}
		}
	}
}
