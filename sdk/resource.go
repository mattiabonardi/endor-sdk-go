package sdk

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

type Resource struct {
	ID                   string `json:"id" bson:"_id" schema:"title=Id"`
	Description          string `json:"description" schema:"title=Description"`
	Service              string `json:"service" schema:"title=Service" ui-schema:"resource=microservice"`
	AdditionalAttributes string `json:"additionalAttributes" schema:"title=Additional attributes schema,format=yaml"` // YAML string, raw
}

func (h *Resource) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var def map[string]Schema
	err := yaml.Unmarshal([]byte(h.AdditionalAttributes), &def)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ResourceDefinition YAML: %w", err)
	}
	return &RootSchema{
		Schema: Schema{
			Type:       ObjectType,
			Properties: &def,
		},
	}, nil
}

type ResourceAction struct {
	// version/resource/action
	ID          string `json:"id" schema:"title=Id"`
	Resource    string `json:"resource" schema:"title=Resource" ui-schema:"resource=resource"`
	Description string `json:"description" schema:"title=Description"`
	InputSchema string `json:"inputSchema" schema:"title=Input schema,format=yaml"`
}

type ResourceInstanceInterface interface {
	GetID() *string
	SetID(id string)
}

type ResourceInstance[T ResourceInstanceInterface] struct {
	This     T              `bson:",inline"`
	Metadata map[string]any `bson:"metadata,omitempty"`
}

// ToSchema genera lo schema della risorsa, includendo lo schema dei metadata
func (d *ResourceInstance[T]) ToSchema(metadataSchema *Schema) *RootSchema {
	t := reflect.TypeOf(d.This)
	// Se T è un puntatore, ottieni il tipo sottostante
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	rootSchema := NewSchema(reflect.New(t).Interface())
	if metadataSchema != nil && metadataSchema.Properties != nil {
		if rootSchema.Schema.Properties == nil {
			props := make(map[string]Schema)
			rootSchema.Schema.Properties = &props
		}
		for k, v := range *metadataSchema.Properties {
			(*rootSchema.Schema.Properties)[k] = v
		}
	}
	return rootSchema
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

type DynamicResource struct {
	Id          string `json:"id" bson:"_id" schema:"title=Id,readOnly=true" ui-schema:"hidden=true"`
	Description string `json:"description" bson:"description" schema:"title=Description"`
}

func (h *DynamicResource) GetID() *string {
	return &h.Id
}

func (h *DynamicResource) SetID(id string) {
	h.Id = id
}
