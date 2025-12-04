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
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" {
			tag = t.Field(i).Name
		}
		known[tag] = struct{}{}
	}

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

// ResourceInstanceSpecialized define the abstract model of a specialized instance
type ResourceInstanceSpecialized[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct {
	This         T                      `json:",inline" bson:"this"`
	CategoryThis C                      `json:",inline" bson:"categoryThis"`
	Metadata     map[string]interface{} `json:",inline" bson:"metadata,omitempty"`
}

// GetID implementa ResourceInstanceInterface
func (r *ResourceInstanceSpecialized[T, C]) GetID() *string {
	return r.This.GetID()
}

// SetID implementa ResourceInstanceInterface
func (r *ResourceInstanceSpecialized[T, C]) SetID(id string) {
	r.This.SetID(id)
}

func (r *ResourceInstanceSpecialized[T, C]) UnmarshalJSON(data []byte) error {
	// Tutto in una map grezza
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// estrai This
	var t T
	tBytes, _ := json.Marshal(t)
	var tFields map[string]json.RawMessage
	json.Unmarshal(tBytes, &tFields)

	// prova ad applicare i campi di T
	_ = json.Unmarshal(data, &t)
	marshalledT, _ := json.Marshal(t)

	// trova i campi effettivamente valorizzati in t
	var tMap map[string]json.RawMessage
	json.Unmarshal(marshalledT, &tMap)

	for k := range tMap {
		delete(raw, k)
	}
	r.This = t

	// estrai CategoryThis
	var c C
	_ = json.Unmarshal(data, &c)
	marshalledC, _ := json.Marshal(c)

	var cMap map[string]json.RawMessage
	json.Unmarshal(marshalledC, &cMap)

	for k := range cMap {
		delete(raw, k)
	}
	r.CategoryThis = c

	// il resto dei campi → metadata
	r.Metadata = make(map[string]interface{})
	for k, v := range raw {
		var val any
		json.Unmarshal(v, &val)
		r.Metadata[k] = val
	}

	return nil
}

func (r ResourceInstanceSpecialized[T, C]) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{}

	// merge This
	tBytes, _ := json.Marshal(r.This)
	var tMap map[string]interface{}
	json.Unmarshal(tBytes, &tMap)
	for k, v := range tMap {
		result[k] = v
	}

	// merge CategoryThis
	cBytes, _ := json.Marshal(r.CategoryThis)
	var cMap map[string]interface{}
	json.Unmarshal(cBytes, &cMap)
	for k, v := range cMap {
		result[k] = v
	}

	// merge Metadata
	for k, v := range r.Metadata {
		result[k] = v
	}

	return json.Marshal(result)
}
