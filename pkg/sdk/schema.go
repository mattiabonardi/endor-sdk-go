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
	Reference            string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type                 SchemaTypeName     `json:"type,omitempty" yaml:"type,omitempty"`
	Properties           *map[string]Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Enum                 *[]string          `json:"enum,omitempty" yaml:"enum,omitempty"`
	Title                *string            `json:"title,omitempty" yaml:"title,omitempty"`
	Description          *string            `json:"description,omitempty" yaml:"description,omitempty"`
	Format               *SchemaFormatName  `json:"format,omitempty" yaml:"format,omitempty"`
	ReadOnly             *bool              `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            *bool              `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Required             []string           `json:"required,omitempty" yaml:"required,omitempty"`

	// field dimension
	MinLength *int `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength *int `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`

	UISchema *UISchema `json:"x-ui,omitempty" yaml:"x-ui,omitempty"`
}

type UISchema struct {
	Entity *string   `json:"entity,omitempty" yaml:"entity,omitempty"` // define the reference entity
	Query  *string   `json:"query,omitempty" yaml:"query,omitempty"`   // define the query to get the data of reference entity "$filter() $projection()"
	Order  *[]string `json:"order,omitempty" yaml:"order,omitempty"`   // define the order of the attributes
	Hidden *bool     `json:"hidden,omitempty" yaml:"hidden,omitempty"` // define if the property is displayable
}

type RootSchema struct {
	Schema      `json:",inline" yaml:",inline"`
	Definitions map[string]Schema `json:"$defs,omitempty" yaml:"$defs,omitempty"`
}

// SchemaTransformer is a function that transforms a schema for a specific use case.
// It creates a restrictive version of the base schema without modifying its structure.
type SchemaTransformer func(*Schema)

// Apply applies one or more SchemaTransformers to the RootSchema.
// Use case schemas are declarative restrictions of the canonical entity schema.
// They do not modify the structure, but reduce the interaction surface.
func (r *RootSchema) Apply(ts ...SchemaTransformer) *RootSchema {
	for _, t := range ts {
		t(&r.Schema)
	}
	return r
}

// modifyNestedProperty navigates to a nested property and applies a modifier function.
// Works correctly with both arrays (Items) and nested objects (map values).
func modifyNestedProperty(s *Schema, path string, modifier func(props *map[string]Schema, fieldName string)) {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return
	}

	fieldName := parts[len(parts)-1]
	parentParts := parts[:len(parts)-1]

	current := s
	for i := 0; i < len(parentParts); i++ {
		part := parentParts[i]

		if current.Properties == nil {
			return
		}

		prop, ok := (*current.Properties)[part]
		if !ok {
			return
		}

		if prop.Type == SchemaTypeArray && prop.Items != nil {
			current = prop.Items
		} else if prop.Type == SchemaTypeObject && prop.Properties != nil {
			// For the last parent part, we need to modify in place
			if i == len(parentParts)-1 {
				modifier(prop.Properties, fieldName)
				(*current.Properties)[part] = prop
				return
			}
			// Update the map entry and continue with reference
			(*current.Properties)[part] = prop
			updatedProp := (*current.Properties)[part]
			current = &updatedProp
		} else {
			return
		}
	}

	// If we reach here, current is the parent schema (e.g., array Items)
	if current.Properties != nil {
		modifier(current.Properties, fieldName)
	}
}

// modifyNestedSchema navigates to a nested schema and applies a modifier function to the schema itself.
func modifyNestedSchema(s *Schema, parentPath string, modifier func(schema *Schema)) {
	parts := strings.Split(parentPath, ".")

	current := s
	for i := 0; i < len(parts); i++ {
		part := parts[i]

		if current.Properties == nil {
			return
		}

		prop, ok := (*current.Properties)[part]
		if !ok {
			return
		}

		if prop.Type == SchemaTypeArray && prop.Items != nil {
			if i == len(parts)-1 {
				modifier(prop.Items)
				return
			}
			current = prop.Items
		} else if prop.Type == SchemaTypeObject && prop.Properties != nil {
			if i == len(parts)-1 {
				modifier(&prop)
				(*current.Properties)[part] = prop
				return
			}
			(*current.Properties)[part] = prop
			updatedProp := (*current.Properties)[part]
			current = &updatedProp
		} else {
			return
		}
	}
}

// Require marks the specified fields as required in the schema.
// Supports nested paths using dot notation (e.g., "address.city", "items.name").
// Only fields that exist in the schema properties will be added to the required list.
func Require(fields ...string) SchemaTransformer {
	return func(s *Schema) {
		if s.Properties == nil {
			return
		}

		// Group fields by their parent schema path
		topLevelRequired := make([]string, 0)
		nestedFields := make(map[string][]string) // parentPath -> fieldNames

		for _, f := range fields {
			if !strings.Contains(f, ".") {
				// Top-level field
				if _, ok := (*s.Properties)[f]; ok {
					topLevelRequired = append(topLevelRequired, f)
				}
			} else {
				// Nested field - group by parent path
				lastDot := strings.LastIndex(f, ".")
				parentPath := f[:lastDot]
				fieldName := f[lastDot+1:]
				nestedFields[parentPath] = append(nestedFields[parentPath], fieldName)
			}
		}

		if len(topLevelRequired) > 0 {
			s.Required = topLevelRequired
		}

		// Apply required to nested schemas
		for parentPath, fieldNames := range nestedFields {
			modifyNestedSchema(s, parentPath, func(schema *Schema) {
				if schema.Properties == nil {
					return
				}
				required := make([]string, 0)
				for _, fn := range fieldNames {
					if _, ok := (*schema.Properties)[fn]; ok {
						required = append(required, fn)
					}
				}
				if len(required) > 0 {
					schema.Required = append(schema.Required, required...)
				}
			})
		}
	}
}

// Forbid marks the specified fields as hidden in the schema (x-ui.hidden = true).
// Supports nested paths using dot notation (e.g., "address.city", "items.name").
// This is used when certain fields should not be displayed for a specific use case.
func Forbid(fields ...string) SchemaTransformer {
	return func(s *Schema) {
		if s.Properties == nil {
			return
		}

		for _, f := range fields {
			if !strings.Contains(f, ".") {
				// Top-level field
				if p, ok := (*s.Properties)[f]; ok {
					if p.UISchema == nil {
						p.UISchema = &UISchema{}
					}
					v := true
					p.UISchema.Hidden = &v
					(*s.Properties)[f] = p
				}
			} else {
				// Nested field
				modifyNestedProperty(s, f, func(props *map[string]Schema, fieldName string) {
					if p, ok := (*props)[fieldName]; ok {
						if p.UISchema == nil {
							p.UISchema = &UISchema{}
						}
						v := true
						p.UISchema.Hidden = &v
						(*props)[fieldName] = p
					}
				})
			}
		}
	}
}

// ReadOnly marks the specified fields as read-only in the schema.
// Supports nested paths using dot notation (e.g., "address.city", "items.name").
// Read-only fields cannot be modified by the client.
func ReadOnly(fields ...string) SchemaTransformer {
	return func(s *Schema) {
		if s.Properties == nil {
			return
		}
		for _, f := range fields {
			if !strings.Contains(f, ".") {
				// Top-level field
				if p, ok := (*s.Properties)[f]; ok {
					v := true
					p.ReadOnly = &v
					(*s.Properties)[f] = p
				}
			} else {
				// Nested field
				modifyNestedProperty(s, f, func(props *map[string]Schema, fieldName string) {
					if p, ok := (*props)[fieldName]; ok {
						v := true
						p.ReadOnly = &v
						(*props)[fieldName] = p
					}
				})
			}
		}
	}
}

// WriteOnly marks the specified fields as write-only in the schema.
// Supports nested paths using dot notation (e.g., "address.city", "items.name").
// Write-only fields are not returned in responses (e.g., passwords).
func WriteOnly(fields ...string) SchemaTransformer {
	return func(s *Schema) {
		if s.Properties == nil {
			return
		}
		for _, f := range fields {
			if !strings.Contains(f, ".") {
				// Top-level field
				if p, ok := (*s.Properties)[f]; ok {
					v := true
					p.WriteOnly = &v
					(*s.Properties)[f] = p
				}
			} else {
				// Nested field
				modifyNestedProperty(s, f, func(props *map[string]Schema, fieldName string) {
					if p, ok := (*props)[fieldName]; ok {
						v := true
						p.WriteOnly = &v
						(*props)[fieldName] = p
					}
				})
			}
		}
	}
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
		case reflect.Map:
			// Maps are represented as objects with additionalProperties
			valueType := t.Elem()
			// Check if it's map[string]any (interface{})
			if valueType.Kind() == reflect.Interface && valueType.NumMethod() == 0 {
				// For map[string]any, use empty schema to allow any value
				emptySchema := Schema{}
				schema = Schema{
					Type:                 SchemaTypeObject,
					AdditionalProperties: &emptySchema,
				}
			} else {
				// For typed maps like map[string]string, resolve the value type
				valueSchema := resolveExpandedFieldSchema(reflect.StructField{}, valueType, visited)
				schema = Schema{
					Type:                 SchemaTypeObject,
					AdditionalProperties: &valueSchema,
				}
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
		case "entity":
			s.UISchema.Entity = &v
		case "query":
			s.UISchema.Query = &v
		case "hidden":
			if v == "true" {
				trueValue := true
				s.UISchema.Hidden = &trueValue
			}
		}
	}
}
