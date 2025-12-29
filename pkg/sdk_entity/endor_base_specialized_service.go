package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorBaseSpecializedServiceCategory[T sdk.EntityInstanceSpecializedInterface] struct {
	ID          string
	Description string
	Actions     map[string]sdk.EndorServiceActionInterface
}

func (h *EndorBaseSpecializedServiceCategory[T]) GetID() string {
	return h.ID
}

func (h *EndorBaseSpecializedServiceCategory[T]) GetDescription() string {
	return h.Description
}

func (h *EndorBaseSpecializedServiceCategory[T]) GetSchema() string {
	schema, _ := getRootSchema[T]().ToYAML()
	return schema
}

func (h *EndorBaseSpecializedServiceCategory[T]) GetActions() map[string]sdk.EndorServiceActionInterface {
	return h.Actions
}

func (h *EndorBaseSpecializedServiceCategory[T]) WithActions(actions map[string]sdk.EndorServiceActionInterface) sdk.EndorBaseSpecializedServiceCategoryInterface {
	h.Actions = actions
	return h
}

func NewEndorBaseSpecializedServiceCategory[T sdk.EntityInstanceSpecializedInterface](categoryID string, categoryDescription string) sdk.EndorBaseSpecializedServiceCategoryInterface {
	return &EndorBaseSpecializedServiceCategory[T]{
		ID:          categoryID,
		Description: categoryDescription,
	}
}

type EndorBaseSpecializedService[T sdk.EntityInstanceSpecializedInterface] struct {
	Entity            string
	EntityDescription string
	Priority          *int
	actions           map[string]sdk.EndorServiceActionInterface
	categories        map[string]sdk.EndorBaseSpecializedServiceCategoryInterface
}

func (h EndorBaseSpecializedService[T]) GetEntity() string {
	return h.Entity
}

func (h EndorBaseSpecializedService[T]) GetEntityDescription() string {
	return h.EntityDescription
}

func (h EndorBaseSpecializedService[T]) GetPriority() *int {
	return h.Priority
}

func (h EndorBaseSpecializedService[T]) GetSchema() *sdk.RootSchema {
	return getRootSchema[T]()
}

func NewEndorBaseSpecializedService[T sdk.EntityInstanceSpecializedInterface](entity, entityDescription string) sdk.EndorBaseSpecializedServiceInterface {
	return EndorBaseSpecializedService[T]{
		Entity:            entity,
		EntityDescription: entityDescription,
	}
}

func (h EndorBaseSpecializedService[T]) WithPriority(
	priority int,
) sdk.EndorBaseSpecializedServiceInterface {
	h.Priority = &priority
	return h
}

func (h EndorBaseSpecializedService[T]) WithActions(
	actions map[string]sdk.EndorServiceActionInterface,
) sdk.EndorBaseSpecializedServiceInterface {
	h.actions = actions
	return h
}

func (h EndorBaseSpecializedService[T]) WithCategories(categories []sdk.EndorBaseSpecializedServiceCategoryInterface) sdk.EndorBaseSpecializedServiceInterface {
	if h.categories == nil {
		h.categories = make(map[string]sdk.EndorBaseSpecializedServiceCategoryInterface)
	}
	for _, category := range categories {
		h.categories[category.GetID()] = category
	}
	return h
}

func (h EndorBaseSpecializedService[T]) GetCategories() []sdk.Category {
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

func (h EndorBaseSpecializedService[T]) ToEndorService() sdk.EndorService {
	// Create a new actions map to avoid modifying shared state
	actions := make(map[string]sdk.EndorServiceActionInterface)

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

	return sdk.EndorService{
		Entity:            h.Entity,
		EntityDescription: h.EntityDescription,
		Priority:          h.Priority,
		Actions:           actions,
		EntitySchema:      *rootSchema,
	}
}
