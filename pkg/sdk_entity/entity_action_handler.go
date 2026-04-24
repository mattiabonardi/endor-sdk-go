package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewEntityActionHandler(microServiceId string, services *[]sdk.EndorHandlerInterface) sdk.EndorHandlerInterface {
	entityMethodService := EntityActionHandler{
		microServiceId: microServiceId,
		services:       services,
	}
	return NewEndorBaseHandler[*sdk.EntityAction]("entity-action", "Entity action").
		WithActions(map[string]sdk.EndorHandlerActionInterface{
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

type EntityActionHandler struct {
	microServiceId string
	services       *[]sdk.EndorHandlerInterface
}

func (h *EntityActionHandler) schema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.EntityAction{})).Build(), nil
}

func (h *EntityActionHandler) list(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityAction], error) {
	actionRepo := NewEndorHandlerActionRepository(InitRegistryCore(h.microServiceId, h.services, &c.Logger))
	entityMethods, err := actionRepo.EntityActionList(c.Session)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[[]sdk.EntityAction]().AddData(&entityMethods).AddSchema(sdk.NewSchema(&sdk.EntityAction{})).Build(), nil
}

func (h *EntityActionHandler) instance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityAction], error) {
	actionRepo := NewEndorHandlerActionRepository(InitRegistryCore(h.microServiceId, h.services, &c.Logger))
	entityAction, err := actionRepo.DictionaryActionInstance(c.Session, c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityAction]().AddData(&entityAction.entityAction).AddSchema(sdk.NewSchema(&sdk.EntityAction{})).Build(), nil
}
