package sdk

import (
	"encoding/json"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/yaml.v3"
)

type Resource struct {
	ID                   string `json:"id" bson:"_id" schema:"title=Id"`
	Description          string `json:"description" schema:"title=Description"`
	Service              string `json:"service" schema:"title=Service"`
	AdditionalAttributes string `json:"additionalAttributes" schema:"title=AdditionalAttributes,format=yaml"` // YAML string, raw
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

type ResourceInstanceInterface[ID comparable] interface {
	GetID() *ID
}

type ResourceInstance[ID comparable, T ResourceInstanceInterface[ID]] struct {
	This     T              `bson:",inline"`
	Metadata map[string]any `bson:"metadata,omitempty"`
}

func (d ResourceInstance[ID, T]) MarshalJSON() ([]byte, error) {
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

func (d *ResourceInstance[ID, T]) UnmarshalJSON(data []byte) error {
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

func (d *ResourceInstance[ID, T]) GetID() *ID {
	return d.This.GetID()
}

type DynamicResource struct {
	Id          primitive.ObjectID `json:"id" bson:"_id" schema:"title=Id,readOnly=true" ui-schema:"hidden=true"`
	Description string             `json:"description" bson:"description" schema:"title=Description"`
}

func (h DynamicResource) GetID() *primitive.ObjectID {
	return &h.Id
}

func (h *DynamicResource) SetID(id primitive.ObjectID) {
	h.Id = id
}

// MarshalJSON implements custom JSON marshaling for DynamicResource
func (h DynamicResource) MarshalJSON() ([]byte, error) {
	// Create a temporary struct to avoid recursion
	temp := struct {
		Id          string `json:"id"`
		Description string `json:"description"`
	}{
		Id:          h.Id.Hex(),
		Description: h.Description,
	}
	return json.Marshal(temp)
}
