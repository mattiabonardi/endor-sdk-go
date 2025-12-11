package sdk_resource

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewResourceActionService(microServiceId string, services *[]sdk.EndorServiceInterface) sdk.EndorServiceInterface {
	resourceMethodService := ResourceActionService{
		microServiceId: microServiceId,
		services:       services,
	}
	return NewEndorBaseService[*sdk.ResourceAction]("resource-action", "Resource action").
		WithActions(map[string]sdk.EndorServiceActionInterface{
			"schema": sdk.NewAction(
				resourceMethodService.schema,
				"Get the schema of the resource method",
			),
			"list": sdk.NewAction(
				resourceMethodService.list,
				"Search for available resources",
			),
			"instance": sdk.NewAction(
				resourceMethodService.instance,
				"Get the specified instance of resources",
			)})
}

type ResourceActionService struct {
	microServiceId string
	services       *[]sdk.EndorServiceInterface
}

func (h *ResourceActionService) schema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.ResourceAction{})).Build(), nil
}

func (h *ResourceActionService) list(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceAction], error) {
	resourceMethods, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceActionList()
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[[]sdk.ResourceAction]().AddData(&resourceMethods).AddSchema(sdk.NewSchema(&sdk.ResourceAction{})).Build(), nil
}

func (h *ResourceActionService) instance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceAction], error) {
	resourceAction, err := NewEndorServiceRepository(h.microServiceId, h.services).ActionInstance(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceAction]().AddData(&resourceAction.resourceAction).AddSchema(sdk.NewSchema(&sdk.ResourceAction{})).Build(), nil
}
