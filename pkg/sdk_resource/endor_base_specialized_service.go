package sdk_resource

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorBaseSpecializedServiceCategory[T sdk.ResourceInstanceSpecializedInterface] struct {
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

func NewEndorBaseSpecializedServiceCategory[T sdk.ResourceInstanceSpecializedInterface](categoryID string, categoryDescription string) sdk.EndorBaseSpecializedServiceCategoryInterface {
	return &EndorBaseSpecializedServiceCategory[T]{
		ID:          categoryID,
		Description: categoryDescription,
	}
}

type EndorBaseSpecializedService[T sdk.ResourceInstanceSpecializedInterface] struct {
	Resource            string
	ResourceDescription string
	Priority            *int
	actions             map[string]sdk.EndorServiceActionInterface
	categories          map[string]sdk.EndorBaseSpecializedServiceCategoryInterface
}

func (h EndorBaseSpecializedService[T]) GetResource() string {
	return h.Resource
}

func (h EndorBaseSpecializedService[T]) GetResourceDescription() string {
	return h.ResourceDescription
}

func (h EndorBaseSpecializedService[T]) GetPriority() *int {
	return h.Priority
}

func NewEndorBaseSpecializedService[T sdk.ResourceInstanceSpecializedInterface](resource, resourceDescription string) sdk.EndorBaseSpecializedServiceInterface {
	return EndorBaseSpecializedService[T]{
		Resource:            resource,
		ResourceDescription: resourceDescription,
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
		Resource:            h.Resource,
		ResourceDescription: h.ResourceDescription,
		Priority:            h.Priority,
		Actions:             h.actions,
		ResourceSchema:      *rootSchema,
	}
}
