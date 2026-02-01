package sdk

import (
	"strings"
)

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

// Forbid removes the specified fields from the schema properties.
// Supports nested paths using dot notation (e.g., "address.city", "items.name").
// This is used when certain fields should not be accepted for a specific use case.
func Forbid(fields ...string) SchemaTransformer {
	return func(s *Schema) {
		if s.Properties == nil {
			return
		}

		topLevelFields := make([]string, 0)
		nestedFields := make(map[string][]string)

		for _, f := range fields {
			if !strings.Contains(f, ".") {
				topLevelFields = append(topLevelFields, f)
			} else {
				lastDot := strings.LastIndex(f, ".")
				parentPath := f[:lastDot]
				fieldName := f[lastDot+1:]
				nestedFields[parentPath] = append(nestedFields[parentPath], fieldName)
			}
		}

		// Forbid top-level fields
		for _, f := range topLevelFields {
			delete(*s.Properties, f)
		}

		// Also remove from UISchema order if present
		if s.UISchema != nil && s.UISchema.Order != nil {
			forbiddenSet := make(map[string]bool)
			for _, f := range topLevelFields {
				forbiddenSet[f] = true
			}
			newOrder := make([]string, 0)
			for _, o := range *s.UISchema.Order {
				if !forbiddenSet[o] {
					newOrder = append(newOrder, o)
				}
			}
			s.UISchema.Order = &newOrder
		}

		// Forbid nested fields
		for parentPath, fieldNames := range nestedFields {
			modifyNestedSchema(s, parentPath, func(schema *Schema) {
				if schema.Properties == nil {
					return
				}
				for _, fn := range fieldNames {
					delete(*schema.Properties, fn)
				}
				// Also remove from nested UISchema order if present
				if schema.UISchema != nil && schema.UISchema.Order != nil {
					forbiddenSet := make(map[string]bool)
					for _, fn := range fieldNames {
						forbiddenSet[fn] = true
					}
					newOrder := make([]string, 0)
					for _, o := range *schema.UISchema.Order {
						if !forbiddenSet[o] {
							newOrder = append(newOrder, o)
						}
					}
					schema.UISchema.Order = &newOrder
				}
			})
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

// ReadOnlyExcept marks all fields in the schema as read-only except for the specified fields.
// Supports nested paths using dot notation (e.g., "address.city", "items.name").
// This is useful when most fields should be read-only and only a few remain writable.
// Example: ReadOnlyExcept("name", "email", "address.city") makes all fields read-only except these three.
func ReadOnlyExcept(exceptFields ...string) SchemaTransformer {
	return func(s *Schema) {
		if s.Properties == nil {
			return
		}

		// Build a set of exceptions for quick lookup
		exceptions := make(map[string]bool)
		for _, f := range exceptFields {
			exceptions[f] = true
		}

		// Mark all fields as read-only
		markAllFieldsReadOnly(s, "", exceptions)
	}
}

// markAllFieldsReadOnly recursively marks all fields as read-only except those in the exceptions set.
// The currentPath parameter tracks the dot-notation path to the current field.
func markAllFieldsReadOnly(s *Schema, currentPath string, exceptions map[string]bool) {
	if s.Properties == nil {
		return
	}

	for key, prop := range *s.Properties {
		// Build the full path for this field
		fieldPath := key
		if currentPath != "" {
			fieldPath = currentPath + "." + key
		}

		// Only mark as read-only if not in exceptions
		if !exceptions[fieldPath] {
			v := true
			prop.ReadOnly = &v
		} else {
			// Explicitly set to false if it's an exception
			v := false
			prop.ReadOnly = &v
		}

		// Recursively handle nested objects
		if prop.Type == SchemaTypeObject && prop.Properties != nil {
			markAllFieldsReadOnly(&prop, fieldPath, exceptions)
		}

		// Recursively handle array items
		if prop.Type == SchemaTypeArray && prop.Items != nil {
			markAllFieldsReadOnly(prop.Items, fieldPath, exceptions)
		}

		(*s.Properties)[key] = prop
	}
}
