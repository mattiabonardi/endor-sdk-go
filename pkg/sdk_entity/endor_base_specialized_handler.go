package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorBaseSpecializedHandlerCategory[T sdk.EntityInstanceSpecializedInterface] struct {
	ID          string
	Description string
	Actions     map[string]sdk.EndorHandlerActionInterface
}

func (h *EndorBaseSpecializedHandlerCategory[T]) GetID() string {
	return h.ID
}

func (h *EndorBaseSpecializedHandlerCategory[T]) GetDescription() string {
	return h.Description
}

func (h *EndorBaseSpecializedHandlerCategory[T]) GetSchema() string {
	schema, _ := getRootSchema[T]().ToYAML()
	return schema
}

func (h *EndorBaseSpecializedHandlerCategory[T]) GetActions() map[string]sdk.EndorHandlerActionInterface {
	return h.Actions
}

func (h *EndorBaseSpecializedHandlerCategory[T]) WithActions(actions map[string]sdk.EndorHandlerActionInterface) sdk.EndorBaseSpecializedHandlerCategoryInterface {
	h.Actions = actions
	return h
}

func NewEndorBaseSpecializedHandlerCategory[T sdk.EntityInstanceSpecializedInterface](categoryID string, categoryDescription string) sdk.EndorBaseSpecializedHandlerCategoryInterface {
	return &EndorBaseSpecializedHandlerCategory[T]{
		ID:          categoryID,
		Description: categoryDescription,
	}
}

type EndorBaseSpecializedHandler[T sdk.EntityInstanceSpecializedInterface] struct {
	Entity            string
	EntityDescription string
	Priority          *int
	actions           map[string]sdk.EndorHandlerActionInterface
	categories        map[string]sdk.EndorBaseSpecializedHandlerCategoryInterface
}

func (h EndorBaseSpecializedHandler[T]) GetEntity() string {
	return h.Entity
}

func (h EndorBaseSpecializedHandler[T]) GetEntityDescription() string {
	return h.EntityDescription
}

func (h EndorBaseSpecializedHandler[T]) GetPriority() *int {
	return h.Priority
}

func (h EndorBaseSpecializedHandler[T]) GetSchema() *sdk.RootSchema {
	return getRootSchema[T]()
}

func NewEndorBaseSpecializedHandler[T sdk.EntityInstanceSpecializedInterface](entity, entityDescription string) sdk.EndorBaseSpecializedHandlerInterface {
	return EndorBaseSpecializedHandler[T]{
		Entity:            entity,
		EntityDescription: entityDescription,
	}
}

func (h EndorBaseSpecializedHandler[T]) WithPriority(
	priority int,
) sdk.EndorBaseSpecializedHandlerInterface {
	h.Priority = &priority
	return h
}

func (h EndorBaseSpecializedHandler[T]) WithActions(
	actions map[string]sdk.EndorHandlerActionInterface,
) sdk.EndorBaseSpecializedHandlerInterface {
	h.actions = actions
	return h
}

func (h EndorBaseSpecializedHandler[T]) WithCategories(categories []sdk.EndorBaseSpecializedHandlerCategoryInterface) sdk.EndorBaseSpecializedHandlerInterface {
	if h.categories == nil {
		h.categories = make(map[string]sdk.EndorBaseSpecializedHandlerCategoryInterface)
	}
	for _, category := range categories {
		h.categories[category.GetID()] = category
	}
	return h
}

func (h EndorBaseSpecializedHandler[T]) GetCategories() []sdk.Category {
	categories := []sdk.Category{}
	for _, category := range h.categories {
		categories = append(categories, sdk.Category{
			ID:          category.GetID(),
			Description: category.GetDescription(),
			Schema:      category.GetSchema(),
		})
	}
	return categories
}

func (h EndorBaseSpecializedHandler[T]) ToEndorHandler() sdk.EndorHandler {
	// Create a new actions map to avoid modifying shared state
	actions := make(map[string]sdk.EndorHandlerActionInterface)

	// Copy existing actions
	for k, v := range h.actions {
		actions[k] = v
	}

	// check if categories are defined
	if len(h.categories) > 0 {
		// iterate over categories
		for _, category := range h.categories {
			// iterate over category actions
			if len(category.GetActions()) > 0 {
				for actionName, action := range category.GetActions() {
					actions[category.GetID()+"/"+actionName] = action
				}
			}
		}
	}

	var baseModel T
	rootSchema := sdk.NewSchema(baseModel)

	return sdk.EndorHandler{
		Entity:            h.Entity,
		EntityDescription: h.EntityDescription,
		Priority:          h.Priority,
		Actions:           actions,
		EntitySchema:      *rootSchema,
	}
}
