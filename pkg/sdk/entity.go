package sdk

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Category struct {
	ID          string `json:"id" bson:"id" schema:"title=t(entities.category.fields.id)"`
	Description string `json:"description" bson:"description" schema:"title=t(entities.category.fields.description)"`
	Schema      string `json:"schema" bson:"-" schema:"title=t(entities.category.fields.schema),format=yaml,readOnly=true"`
}

type HybridCategory struct {
	ID               string `json:"id" bson:"id" schema:"title=t(entities.category.fields.id),readOnly=true"`
	Description      string `json:"description" bson:"description" schema:"title=t(entities.category.fields.description),readOnly=true"`
	Schema           string `json:"schema" bson:"-" schema:"title=t(entities.category.fields.schema),format=yaml,readOnly=true"`
	AdditionalSchema string `json:"additionalSchema" bson:"additionalSchema" schema:"title=t(entities.category.fields.additional_schema),format=yaml"`
}

type DynamicCategory struct {
	ID               string `json:"id" bson:"id" schema:"title=t(entities.category.fields.id)"`
	Description      string `json:"description" bson:"description" schema:"title=t(entities.category.fields.description)"`
	AdditionalSchema string `json:"additionalSchema" bson:"additionalSchema" schema:"title=t(entities.category.fields.additional_schema),format=yaml"`
}

func (c *HybridCategory) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(c.AdditionalSchema), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Category AdditionalAttributes YAML: %w", err)
	}
	return &schema, nil
}

func (c *DynamicCategory) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(c.AdditionalSchema), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Category AdditionalAttributes YAML: %w", err)
	}
	return &schema, nil
}

type EntityInterface interface {
	GetID() any
	GetCategoryType() string
	SetCategoryType(entityType string)
	GetService() string
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
	ID          string `json:"id" bson:"_id" schema:"title=t(entities.entity.fields.id),readOnly=true"`
	Description string `json:"description" schema:"title=t(entities.entity.fields.description),readOnly=true"`
	Type        string `json:"type" schema:"title=t(entities.entity.fields.type),readOnly=true"`
	Service     string `json:"service" schema:"title=t(entities.entity.fields.service),readOnly=true" ui-schema:"entity=microservice"`
	Schema      string `json:"schema" bson:"-" schema:"title=t(entities.entity.fields.schema),format=yaml,readOnly=true"`
}

func (h *Entity) GetID() any {
	return h.ID
}

func (h *Entity) GetCategoryType() string {
	return h.Type
}

func (r *Entity) SetCategoryType(t string) {
	r.Type = t
}

func (h *Entity) GetService() string {
	return h.Service
}

func (r *Entity) SetService(service string) {
	r.Service = service
}

// #endregion

// #region Entity specialized

type EntitySpecialized struct {
	Entity     `json:",inline" bson:",inline"`
	Categories []Category `json:"categories,omitempty" schema:"title=t(entities.entity.fields.categories),readOnly=true"`
}

// #endregion

// #region Entity hybrid

type EntityHybrid struct {
	Entity           `json:",inline" bson:",inline"`
	AdditionalSchema string `json:"additionalSchema" bson:"additionalSchema" schema:"title=t(entities.entity.fields.additional_schema),format=yaml"` // YAML string, raw
}

func (h *EntityHybrid) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(h.AdditionalSchema), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EntityDefinition YAML: %w", err)
	}
	return &schema, nil
}

// #endregion

// #region Entity hybrid specialized

type EntityHybridSpecialized struct {
	EntityHybrid         `json:",inline" bson:",inline"`
	Categories           []HybridCategory  `json:"categories,omitempty" bson:"categories,omitempty" schema:"title=t(entities.entity.fields.categories),readOnly=true"`
	AdditionalCategories []DynamicCategory `json:"additionalCategories,omitempty" bson:"additionalCategories,omitempty" schema:"title=t(entities.entity.fields.additional_categories)"`
}

func (h *EntityHybridSpecialized) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(h.AdditionalSchema), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EntityDefinition YAML: %w", err)
	}
	return &schema, nil
}

// #endregion

type EntityAction struct {
	// version/entity/action
	ID          string `json:"id" schema:"title=t(entities.entity_action.fields.id)"`
	Entity      string `json:"entity" schema:"title=t(entities.entity_action.fields.entity)" ui-schema:"entity=entity"`
	Description string `json:"description" schema:"title=t(entities.entity_action.fields.description)"`
	InputSchema string `json:"inputSchema" schema:"title=t(entities.entity_action.fields.input_schema),format=yaml"`
}

func (h *EntityAction) GetID() any {
	return h.ID
}

type DynamicEntity struct {
	Id string `json:"id" bson:"_id" schema:"title=t(entities.dynamic_entity.fields.id),readOnly=true" ui-schema:"hidden=true"`
}

func (h *DynamicEntity) GetID() any {
	return h.Id
}

type DynamicEntitySpecialized struct {
	Id   string `json:"id" bson:"_id" schema:"title=t(entities.dynamic_entity.fields.id),readOnly=true" ui-schema:"hidden=true"`
	Type string `json:"type" bson:"type" schema:"title=t(entities.dynamic_entity.fields.type),readOnly=true"`
}

func (h *DynamicEntitySpecialized) GetID() any {
	return h.Id
}

func (h *DynamicEntitySpecialized) GetCategoryType() string {
	return h.Type
}

func (h *DynamicEntitySpecialized) SetCategoryType(categoryType string) {
	h.Type = categoryType
}

type EntityRepositoryInterface interface {
	List(session Session, entityType *EntityType) ([]EntityInterface, error)
	Instance(session Session, entityType *EntityType, dto ReadInstanceDTO) (*EntityInterface, error)
}
