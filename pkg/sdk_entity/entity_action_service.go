package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewEntityActionService(microServiceId string, services *[]sdk.EndorServiceInterface) sdk.EndorServiceInterface {
	entityMethodService := EntityActionService{
		microServiceId: microServiceId,
		services:       services,
	}
	return NewEndorBaseService[*sdk.EntityAction]("entity-action", "Entity action").
		WithActions(map[string]sdk.EndorServiceActionInterface{
			"schema": sdk.NewAction(
				entityMethodService.schema,
				"Get the schema of the entity method",
			),
			"list": sdk.NewAction(
				entityMethodService.list,
				"Search for available entities",
			),
			"instance": sdk.NewAction(
				entityMethodService.instance,
				"Get the specified instance of entities",
			)})
}

type EntityActionService struct {
	microServiceId string
	services       *[]sdk.EndorServiceInterface
}

func (h *EntityActionService) schema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.EntityAction{})).Build(), nil
}

func (h *EntityActionService) list(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityAction], error) {
	entityMethods, err := NewEndorServiceRepository(h.microServiceId, h.services).EntityActionList()
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[[]sdk.EntityAction]().AddData(&entityMethods).AddSchema(sdk.NewSchema(&sdk.EntityAction{})).Build(), nil
}

func (h *EntityActionService) instance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityAction], error) {
	entityAction, err := NewEndorServiceRepository(h.microServiceId, h.services).ActionInstance(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityAction]().AddData(&entityAction.entityAction).AddSchema(sdk.NewSchema(&sdk.EntityAction{})).Build(), nil
}
