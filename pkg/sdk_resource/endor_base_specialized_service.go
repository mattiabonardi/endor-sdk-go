package sdk_resource

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorSpecializedServiceCategory[T sdk.ResourceInstanceSpecializedInterface, C any] struct {
	Category sdk.Category
}

func (h *EndorSpecializedServiceCategory[T, C]) GetID() string {
	return h.Category.ID
}

func NewEndorSpecializedServiceCategory[T sdk.ResourceInstanceSpecializedInterface, C any](category sdk.Category) sdk.EndorBaseSpecializedServiceCategoryInterface {
	return &EndorSpecializedServiceCategory[T, C]{
		Category: category,
	}
}

type EndorbaseSpecializedService[T sdk.ResourceInstanceSpecializedInterface] struct {
	Resource            string
	ResourceDescription string
	Priority            *int
	actions             map[string]sdk.EndorServiceActionInterface
	categories          map[string]sdk.EndorBaseSpecializedServiceCategoryInterface
}

func (h EndorbaseSpecializedService[T]) GetResource() string {
	return h.Resource
}

func (h EndorbaseSpecializedService[T]) GetResourceDescription() string {
	return h.ResourceDescription
}

func (h EndorbaseSpecializedService[T]) GetPriority() *int {
	return h.Priority
}

func NewEndorBaseSpecializedService[T sdk.ResourceInstanceSpecializedInterface](resource, resourceDescription string) sdk.EndorBaseSpecializedServiceInterface {
	return EndorbaseSpecializedService[T]{
		Resource:            resource,
		ResourceDescription: resourceDescription,
	}
}

func (h EndorbaseSpecializedService[T]) WithActions(
	actions map[string]sdk.EndorServiceActionInterface,
) sdk.EndorBaseSpecializedServiceInterface {
	h.actions = actions
	return h
}

func (h EndorbaseSpecializedService[T]) WithCategories(categories []sdk.EndorBaseSpecializedServiceCategoryInterface) sdk.EndorBaseSpecializedServiceInterface {
	if h.categories == nil {
		h.categories = make(map[string]sdk.EndorBaseSpecializedServiceCategoryInterface)
	}
	for _, category := range categories {
		h.categories[category.GetID()] = category
	}
	return h
}

func (h EndorbaseSpecializedService[T]) ToEndorService() sdk.EndorService {
	// check if categories are defined
	if len(h.categories) > 0 {
		// iterate over categories
		/*for _, category := range h.categories {
			//TODO: add methods
		}*/
	}

	return sdk.EndorService{
		Resource:            h.Resource,
		ResourceDescription: h.ResourceDescription,
		Priority:            h.Priority,
		Actions:             h.actions,
	}
}
