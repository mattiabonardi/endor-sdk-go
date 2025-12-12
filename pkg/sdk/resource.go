package sdk

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Category struct {
	ID          string `json:"id" bson:"id" schema:"title=Category ID"`
	Description string `json:"description" bson:"description" schema:"title=Category Description"`
}

type HybridCategory struct {
	ID                   string `json:"id" bson:"id" schema:"title=Category ID"`
	Description          string `json:"description" bson:"description" schema:"title=Category Description"`
	AdditionalAttributes string `json:"additionalAttributes" bson:"additionalAttributes" schema:"title=Additional category attributes schema,format=yaml"`
}

func (c *HybridCategory) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(c.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Category AdditionalAttributes YAML: %w", err)
	}
	return &schema, nil
}

type ResourceInterface interface {
	GetID() *string
	SetID(id string)
	GetCategoryType() *string
	SetCategoryType(resourceType string)
	SetService(service string)
}

type ResourceType string

const (
	ResourceTypeBase               ResourceType = "base"
	ResourceTypeBaseSpecialized    ResourceType = "base-specialized"
	ResourceTypeHybrid             ResourceType = "hybrid"
	ResourceTypeHybridSpecialized  ResourceType = "hybrid-specialized"
	ResourceTypeDynamic            ResourceType = "dynamic"
	ResourceTypeDynamicSpecialized ResourceType = "dynamic-specialized"
)

// #region Resource

type Resource struct {
	ID          string `json:"id" bson:"_id" schema:"title=Id"`
	Description string `json:"description" schema:"title=Description"`
	Type        string `json:"type" schema:"title=Type,readOnly=true"`
	Service     string `json:"service" schema:"title=Service" ui-schema:"resource=microservice"`
}

func (h *Resource) GetID() *string {
	return &h.ID
}

func (h *Resource) SetID(id string) {
	h.ID = id
}

func (h *Resource) GetCategoryType() *string {
	return &h.Type
}

func (r *Resource) SetCategoryType(t string) {
	r.Type = t
}

func (r *Resource) SetService(service string) {
	r.Service = service
}

// #endregion

// #region Resource specialized

type ResourceSpecialized struct {
	Resource   `json:",inline" bson:",inline"`
	Categories []Category `json:"categories,omitempty" bson:"categories,omitempty" schema:"title=Categories"`
}

// #endregion

// #region Resource hubrid

type ResourceHybrid struct {
	Resource             `json:",inline" bson:",inline"`
	AdditionalAttributes string `json:"additionalAttributes" schema:"title=Additional attributes schema,format=yaml"` // YAML string, raw
}

func (h *ResourceHybrid) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(h.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ResourceDefinition YAML: %w", err)
	}
	return &schema, nil
}

// #endregion

// #region Resource hybrid specialized

type ResourceHybridSpecialized struct {
	ResourceHybrid `json:",inline" bson:",inline"`
	Categories     []HybridCategory `json:"categories,omitempty" bson:"categories,omitempty" schema:"title=Categories"`
}

func (h *ResourceHybridSpecialized) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema = RootSchema{}
	err := yaml.Unmarshal([]byte(h.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ResourceDefinition YAML: %w", err)
	}
	return &schema, nil
}

// #endregion

type ResourceAction struct {
	// version/resource/action
	ID          string `json:"id" schema:"title=Id"`
	Resource    string `json:"resource" schema:"title=Resource" ui-schema:"resource=resource"`
	Description string `json:"description" schema:"title=Description"`
	InputSchema string `json:"inputSchema" schema:"title=Input schema,format=yaml"`
}

func (h *ResourceAction) GetID() *string {
	return &h.ID
}

func (h *ResourceAction) SetID(id string) {
	h.ID = id
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
