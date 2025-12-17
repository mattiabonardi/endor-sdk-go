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

type EntityInterface interface {
	GetID() string
	SetID(id string)
	GetCategoryType() string
	SetCategoryType(entityType string)
	SetService(service string)
}

type EntityType string

const (
	EntityTypeBase               EntityType = "base"
	EntityTypeBaseSpecialized    EntityType = "base-specialized"
	EntityTypeHybrid             EntityType = "hybrid"
	EntityTypeHybridSpecialized  EntityType = "hybrid-specialized"
	EntityTypeDynamic            EntityType = "dynamic"
	EntityTypeDynamicSpecialized EntityType = "dynamic-specialized"
)

// #region Entity

type Entity struct {
	ID          string `json:"id" bson:"_id" schema:"title=Id"`
	Description string `json:"description" schema:"title=Description"`
	Type        string `json:"type" schema:"title=Type,readOnly=true"`
	Service     string `json:"service" schema:"title=Service" ui-schema:"entity=microservice"`
}

func (h *Entity) GetID() string {
	return h.ID
}

func (h *Entity) SetID(id string) {
	h.ID = id
}

func (h *Entity) GetCategoryType() string {
	return h.Type
}

func (r *Entity) SetCategoryType(t string) {
	r.Type = t
}

func (r *Entity) SetService(service string) {
	r.Service = service
}

// #endregion

// #region Entity specialized

type EntitySpecialized struct {
	Entity     `json:",inline" bson:",inline"`
	Categories []Category `json:"categories,omitempty" bson:"categories,omitempty" schema:"title=Categories"`
}

// #endregion

// #region Entity hubrid

type EntityHybrid struct {
	Entity               `json:",inline" bson:",inline"`
	AdditionalAttributes string `json:"additionalAttributes" schema:"title=Additional attributes schema,format=yaml"` // YAML string, raw
}

func (h *EntityHybrid) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(h.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EntityDefinition YAML: %w", err)
	}
	return &schema, nil
}

// #endregion

// #region Entity hybrid specialized

type EntityHybridSpecialized struct {
	EntityHybrid `json:",inline" bson:",inline"`
	Categories   []HybridCategory `json:"categories,omitempty" bson:"categories,omitempty" schema:"title=Categories"`
}

func (h *EntityHybridSpecialized) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema = RootSchema{}
	err := yaml.Unmarshal([]byte(h.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EntityDefinition YAML: %w", err)
	}
	return &schema, nil
}

// #endregion

type EntityAction struct {
	// version/entity/action
	ID          string `json:"id" schema:"title=Id"`
	Entity      string `json:"entity" schema:"title=Entity" ui-schema:"entity=entity"`
	Description string `json:"description" schema:"title=Description"`
	InputSchema string `json:"inputSchema" schema:"title=Input schema,format=yaml"`
}

func (h *EntityAction) GetID() string {
	return h.ID
}

func (h *EntityAction) SetID(id string) {
	h.ID = id
}

type DynamicEntity struct {
	Id          string `json:"id" bson:"_id" schema:"title=Id,readOnly=true" ui-schema:"hidden=true"`
	Description string `json:"description" bson:"description" schema:"title=Description"`
}

func (h *DynamicEntity) GetID() string {
	return h.Id
}

func (h *DynamicEntity) SetID(id string) {
	h.Id = id
}

type DynamicEntitySpecialized struct {
	Id          string `json:"id" bson:"_id" schema:"title=Id,readOnly=true" ui-schema:"hidden=true"`
	Description string `json:"description" bson:"description" schema:"title=Description"`
	Type        string `json:"type" bson:"type" schema:"title=Type,readOnly=true"`
}

func (h *DynamicEntitySpecialized) GetID() string {
	return h.Id
}

func (h *DynamicEntitySpecialized) SetID(id string) {
	h.Id = id
}

func (h *DynamicEntitySpecialized) GetCategoryType() string {
	return h.Type
}

func (h *DynamicEntitySpecialized) SetCategoryType(categoryType string) {
	h.Type = categoryType
}
