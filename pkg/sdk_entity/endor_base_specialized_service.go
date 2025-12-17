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

func NewEndorBaseSpecializedService[T sdk.EntityInstanceSpecializedInterface](entity, entityDescription string) sdk.EndorBaseSpecializedServiceInterface {
	return EndorBaseSpecializedService[T]{
		Entity:            entity,
		EntityDescription: entityDescription,
	}
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

func (h EndorBaseSpecializedService[T]) ToEndorService() sdk.EndorService {
	// check if categories are defined
	if len(h.categories) > 0 {
		// iterate over categories
		for _, category := range h.categories {
			// iterate over category actions
			if len(category.GetActions()) > 0 {
				for actionName, action := range category.GetActions() {
					h.actions[category.GetID()+"/"+actionName] = action
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
		Actions:           h.actions,
		EntitySchema:      *rootSchema,
	}
}
