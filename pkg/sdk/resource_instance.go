package sdk

import (
	"encoding/json"
	"reflect"
)

type ResourceInstanceInterface interface {
	GetID() *string
	SetID(id string)
}

type ResourceInstanceSpecializedInterface interface {
	ResourceInstanceInterface
	GetCategoryType() *string
	SetCategoryType(categoryType string)
}

type ResourceInstance[T ResourceInstanceInterface] struct {
	This     T              `bson:",inline"`
	Metadata map[string]any `bson:"metadata,omitempty"`
}

func (d ResourceInstance[T]) MarshalJSON() ([]byte, error) {
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

func (d *ResourceInstance[T]) UnmarshalJSON(data []byte) error {
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

func (d *ResourceInstance[T]) GetID() *string {
	return d.This.GetID()
}

func (d *ResourceInstance[T]) SetID(id string) {
	d.This.SetID(id)
}

// ResourceInstanceSpecialized define the abstract model of a specialized instance
type ResourceInstanceSpecialized[T ResourceInstanceSpecializedInterface] struct {
	ResourceInstance[T]
}

func (r *ResourceInstanceSpecialized[T]) GetCategoryType() *string {
	return r.This.GetCategoryType()
}

func (r *ResourceInstanceSpecialized[T]) SetCategoryType(categoryType string) {
	r.This.SetCategoryType(categoryType)
}
