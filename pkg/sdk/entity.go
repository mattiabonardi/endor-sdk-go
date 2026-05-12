package sdk

import (
	"fmt"
	"strings"
)

type Category struct {
	ID          string `json:"id" schema:"title=t(sdk.entity.fields.category.id),readOnly=true"`
	Title       string `json:"title" schema:"title=t(sdk.entity.fields.category.title),readOnly=true"`
	Description string `json:"description" schema:"title=t(sdk.entity.fields.category.description),readOnly=true"`
	Schema      string `json:"schema" schema:"title=t(sdk.entity.fields.category.schema),format=yaml,readOnly=true"`
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
	ID          string     `json:"id" schema:"title=t(sdk.entity.fields.id),readOnly=true"`
	Title       string     `json:"title" schema:"title=t(sdk.entity.fields.title),readOnly=true"`
	Description string     `json:"description" schema:"title=t(sdk.entity.fields.description),readOnly=true"`
	Type        string     `json:"type" schema:"title=t(sdk.entity.fields.type),readOnly=true"`
	Module      string     `json:"module" schema:"title=t(sdk.entity.fields.module),readOnly=true" ui-schema:"entity=core/module"`
	Schema      string     `json:"schema" schema:"title=t(sdk.entity.fields.schema),format=yaml,readOnly=true"`
	Categories  []Category `json:"categories,omitempty" bson:"categories,omitempty" schema:"title=t(sdk.entity.fields.categories),readOnly=true"`
}

func (h *Entity) GetID() any {
	return h.ID
}

// #endregion

type EntityAction struct {
	// module/version/entity/action
	ID          string `json:"id" schema:"title=t(sdk.entity_action.fields.id)"`
	Entity      string `json:"entity" schema:"title=t(sdk.entity_action.fields.entity)" ui-schema:"core/entity"`
	Description string `json:"description" schema:"title=t(sdk.entity_action.fields.description)"`
	InputSchema string `json:"inputSchema" schema:"title=t(sdk.entity_action.fields.input_schema),format=yaml"`
}

func (h *EntityAction) GetID() any {
	return h.ID
}

type DynamicEntity struct {
	Id string `json:"id" bson:"_id" schema:"title=t(sdk.dynamic_entity.fields.id),readOnly=true" ui-schema:"hidden=true"`
}

func (h *DynamicEntity) GetID() any {
	return h.Id
}

type DynamicEntitySpecialized struct {
	Id   string `json:"id" bson:"_id" schema:"title=t(sdk.dynamic_entity.fields.id),readOnly=true" ui-schema:"hidden=true"`
	Type string `json:"type" bson:"type" schema:"title=t(sdk.dynamic_entity.fields.type),readOnly=true"`
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
	List(entityType *EntityType) ([]Entity, error)
	Instance(entityType *EntityType, dto ReadInstanceDTO) (*Entity, error)
}

// Parse entity ID <domain>/<entity>
func ParseEntityID(entityId string) (string, string, error) {
	parts := strings.Split(entityId, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid entity id (<domain>/<entity>)")
	}

	return parts[0], parts[1], nil
}

// Parse entity ID <domain>/<entity>/<category (optional)>/<action>
func ParseEntityActionID(entityActionId string) (string, string, string, string, error) {
	parts := strings.Split(entityActionId, "/")

	if len(parts) < 3 || len(parts) > 4 {
		return "", "", "", "", fmt.Errorf("invalid entity action id (<domain>/<entity>/<category optional>/<action>)")
	}

	domain := parts[0]
	entity := parts[1]
	category := ""
	action := ""

	if len(parts) == 4 {
		category = parts[2]
		action = parts[3]
	} else {
		category = ""
		action = parts[2]
	}

	return domain, entity, category, action, nil
}
