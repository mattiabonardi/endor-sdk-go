package sdk_resource

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorBaseService[T sdk.ResourceInstanceInterface] struct {
	resource            string
	resourceDescription string
	priority            *int
	actions             map[string]sdk.EndorServiceActionInterface
}

func (h EndorBaseService[T]) GetResource() string {
	return h.resource
}

func (h EndorBaseService[T]) GetResourceDescription() string {
	return h.resourceDescription
}

func (h EndorBaseService[T]) GetPriority() *int {
	return h.priority
}

func NewEndorBaseService[T sdk.ResourceInstanceInterface](resource, resourceDescription string) sdk.EndorBaseServiceInterface {
	return EndorBaseService[T]{
		resource:            resource,
		resourceDescription: resourceDescription,
	}
}

func (h EndorBaseService[T]) WithActions(
	actions map[string]sdk.EndorServiceActionInterface,
) sdk.EndorBaseServiceInterface {
	h.actions = actions
	return h
}

func (h EndorBaseService[T]) ToEndorService() sdk.EndorService {
	var baseModel T
	rootSchema := sdk.NewSchema(baseModel)
	return sdk.EndorService{
		Resource:            h.resource,
		ResourceDescription: h.resourceDescription,
		Priority:            h.priority,
		Actions:             h.actions,
		ResourceSchema:      *rootSchema,
	}
}
