package sdk

import (
	"encoding/json"
	"reflect"
)

type EntityInstanceInterface interface {
	GetID() string
}

type EntityInstanceSpecializedInterface interface {
	EntityInstanceInterface
	GetCategoryType() string
	SetCategoryType(categoryType string)
}

type EntityInstance[T EntityInstanceInterface] struct {
	This     T              `bson:",inline"`
	Metadata map[string]any `bson:"metadata,omitempty"`
}

func (d EntityInstance[T]) MarshalJSON() ([]byte, error) {
	base, err := json.Marshal(d.This)
	if err != nil {
		return nil, err
	}

	var baseMap map[string]any
	if err := json.Unmarshal(base, &baseMap); err != nil {
		return nil, err
	}

	for k, v := range d.Metadata {
		if _, exists := baseMap[k]; !exists {
			baseMap[k] = v
		}
	}

	return json.Marshal(baseMap)
}

func (d *EntityInstance[T]) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var temp T
	t := reflect.TypeOf(temp)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	known := map[string]struct{}{}
	// Recursively collect field names, including those from embedded structs
	var collectFields func(reflect.Type)
	collectFields = func(t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Check if this is an embedded/anonymous field with "inline" tag
			jsonTag := field.Tag.Get("json")
			isInline := field.Anonymous || jsonTag == ",inline" ||
				(len(jsonTag) >= 7 && jsonTag[len(jsonTag)-7:] == ",inline")

			if isInline && field.Type.Kind() == reflect.Struct {
				// Recursively collect fields from embedded struct
				collectFields(field.Type)
			} else if isInline && field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct {
				// Handle pointer to embedded struct
				collectFields(field.Type.Elem())
			} else {
				// Regular field: extract JSON tag name
				tag := field.Tag.Get("json")
				if tag == "" {
					tag = field.Name
				} else {
					// Remove options like ",omitempty", ",inline" from tag
					for commaIdx := 0; commaIdx < len(tag); commaIdx++ {
						if tag[commaIdx] == ',' {
							tag = tag[:commaIdx]
							break
						}
					}
				}
				if tag != "" && tag != "-" {
					known[tag] = struct{}{}
				}
			}
		}
	}
	collectFields(t)

	baseMap := make(map[string]any)
	metaMap := make(map[string]any)

	for k, v := range raw {
		if _, isKnown := known[k]; isKnown {
			baseMap[k] = v
		} else {
			metaMap[k] = v
		}
	}

	baseBytes, err := json.Marshal(baseMap)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(baseBytes, &d.This); err != nil {
		return err
	}

	d.Metadata = metaMap
	return nil
}

func (d *EntityInstance[T]) GetID() string {
	return d.This.GetID()
}

// PartialEntityInstance represents a partial update to an EntityInstance
// This is used for update operations where only some fields need to be modified
// Similar to EntityInstance but uses map[string]any for This to allow partial updates
type PartialEntityInstance[T EntityInstanceInterface] struct {
	This     map[string]any `bson:",inline"`
	Metadata map[string]any `bson:"metadata,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling for PartialEntityInstance
// It separates flat JSON into This (known entity fields) and Metadata (unknown fields)
func (p *PartialEntityInstance[T]) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var temp T
	t := reflect.TypeOf(temp)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	known := map[string]struct{}{}
	// Recursively collect field names, including those from embedded structs
	var collectFields func(reflect.Type)
	collectFields = func(t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Check if this is an embedded/anonymous field with "inline" tag
			jsonTag := field.Tag.Get("json")
			isInline := field.Anonymous || jsonTag == ",inline" ||
				(len(jsonTag) >= 7 && jsonTag[len(jsonTag)-7:] == ",inline")

			if isInline && field.Type.Kind() == reflect.Struct {
				// Recursively collect fields from embedded struct
				collectFields(field.Type)
			} else if isInline && field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
				// Handle pointer to embedded struct
				collectFields(field.Type.Elem())
			} else {
				// Regular field: extract JSON tag name
				tag := field.Tag.Get("json")
				if tag == "" {
					tag = field.Name
				} else {
					// Remove options like ",omitempty", ",inline" from tag
					for commaIdx := 0; commaIdx < len(tag); commaIdx++ {
						if tag[commaIdx] == ',' {
							tag = tag[:commaIdx]
							break
						}
					}
				}
				if tag != "" && tag != "-" {
					known[tag] = struct{}{}
				}
			}
		}
	}
	collectFields(t)

	thisMap := make(map[string]any)
	metaMap := make(map[string]any)

	for k, v := range raw {
		if _, isKnown := known[k]; isKnown {
			thisMap[k] = v
		} else {
			metaMap[k] = v
		}
	}

	p.This = thisMap
	p.Metadata = metaMap
	return nil
}

// EntityInstanceSpecialized define the abstract model of a specialized instance
type EntityInstanceSpecialized[T EntityInstanceSpecializedInterface] struct {
	EntityInstance[T]
}

func (r *EntityInstanceSpecialized[T]) GetCategoryType() string {
	return r.This.GetCategoryType()
}

func (r *EntityInstanceSpecialized[T]) SetCategoryType(categoryType string) {
	r.This.SetCategoryType(categoryType)
}
