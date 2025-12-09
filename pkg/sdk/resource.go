package sdk

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Category rappresenta una categoria con attributi dinamici specifici
type Category struct {
	ID                   string `json:"id" bson:"id" schema:"title=Category ID"`
	Description          string `json:"description" bson:"description" schema:"title=Category Description"`
	AdditionalAttributes string `json:"additionalAttributes" bson:"additionalAttributes" schema:"title=Additional category attributes schema,format=yaml"`
}

// UnmarshalAdditionalAttributes deserializza gli attributi aggiuntivi della categoria
func (c *Category) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(c.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Category AdditionalAttributes YAML: %w", err)
	}
	return &schema, nil
}

type ResourceInterface interface {
	GetID() string
	GetType() ResourceType
	AsBase() (*Resource, bool)
	AsSpecialized() (*ResourceSpecialized, bool)
	UnmarshalAdditionalAttributes() (*RootSchema, error)
}

type ResourceType string

const (
	ResourceTypeBase        ResourceType = "base"
	ResourceTypeSpecialized ResourceType = "specialized"
)

type Resource struct {
	ID                   string       `json:"id" bson:"_id" schema:"title=Id"`
	Description          string       `json:"description" schema:"title=Description"`
	Type                 ResourceType `json:"type" schema:"title=Type"`
	Service              string       `json:"service" schema:"title=Service" ui-schema:"resource=microservice"`
	AdditionalAttributes string       `json:"additionalAttributes" schema:"title=Additional attributes schema,format=yaml"` // YAML string, raw
}

func (h *Resource) GetID() string {
	return h.ID
}

func (h *Resource) GetType() ResourceType {
	return ResourceTypeBase
}

func (r *Resource) AsBase() (*Resource, bool) {
	return r, true
}

func (r *Resource) AsSpecialized() (*ResourceSpecialized, bool) {
	return nil, false
}

func (h *Resource) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(h.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ResourceDefinition YAML: %w", err)
	}
	return &schema, nil
}

type ResourceSpecialized struct {
	Resource   `json:",inline" bson:",inline"`
	Categories []Category `json:"categories,omitempty" bson:"categories,omitempty" schema:"title=Categories"`
}

func (h *ResourceSpecialized) GetID() string {
	return h.ID
}

func (h *ResourceSpecialized) GetType() ResourceType {
	return ResourceTypeSpecialized
}

func (r *ResourceSpecialized) AsBase() (*Resource, bool) {
	return &r.Resource, true
}

func (r *ResourceSpecialized) AsSpecialized() (*ResourceSpecialized, bool) {
	return r, true
}

func (h *ResourceSpecialized) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(h.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ResourceDefinition YAML: %w", err)
	}
	return &schema, nil
}

// GetCategoryByID trova una categoria per ID
func (h *ResourceSpecialized) GetCategoryByID(categoryID string) (*Category, bool) {
	for i, category := range h.Categories {
		if category.ID == categoryID {
			return &h.Categories[i], true
		}
	}
	return nil, false
}

// GetCategorySchema restituisce lo schema combinato di base e categoria specifica
func (h *ResourceSpecialized) GetCategorySchema(categoryID string) (*RootSchema, error) {
	// Schema di base della risorsa
	baseSchema, err := h.UnmarshalAdditionalAttributes()
	if err != nil {
		return nil, err
	}

	// Se non c'è categoria specificata, restituisci solo lo schema di base
	if categoryID == "" {
		return baseSchema, nil
	}

	// Trova la categoria
	category, found := h.GetCategoryByID(categoryID)
	if !found {
		return nil, fmt.Errorf("category '%s' not found", categoryID)
	}

	// Schema della categoria
	categorySchema, err := category.UnmarshalAdditionalAttributes()
	if err != nil {
		return nil, err
	}

	// Combina gli schemi (categoria estende il base)
	combined := &RootSchema{
		Schema: Schema{
			Type:       SchemaTypeObject,
			Properties: &map[string]Schema{},
		},
	}

	// Aggiungi proprietà di base
	if baseSchema.Properties != nil {
		for k, v := range *baseSchema.Properties {
			(*combined.Properties)[k] = v
		}
	}

	// Aggiungi proprietà della categoria (sovrascrivono se esistenti)
	if categorySchema.Properties != nil {
		for k, v := range *categorySchema.Properties {
			(*combined.Properties)[k] = v
		}
	}

	return combined, nil
}

type ResourceAction struct {
	// version/resource/action
	ID          string `json:"id" schema:"title=Id"`
	Resource    string `json:"resource" schema:"title=Resource" ui-schema:"resource=resource"`
	Description string `json:"description" schema:"title=Description"`
	InputSchema string `json:"inputSchema" schema:"title=Input schema,format=yaml"`
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

type DynamicResourceSpecialized struct {
	Id           string `json:"id" bson:"_id" schema:"title=Id,readOnly=true" ui-schema:"hidden=true"`
	Description  string `json:"description" bson:"description" schema:"title=Description"`
	CategoryType string `json:"categoryType" bson:"categoryType" schema:"title=Type,readOnly=true"`
}

func (h *DynamicResourceSpecialized) GetID() *string {
	return &h.Id
}

func (h *DynamicResourceSpecialized) SetID(id string) {
	h.Id = id
}

func (h *DynamicResourceSpecialized) GetCategoryType() *string {
	return &h.CategoryType
}

func (h *DynamicResourceSpecialized) SetCategoryType(categoryType string) {
	h.CategoryType = categoryType
}
