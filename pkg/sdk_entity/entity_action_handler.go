package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewEntityActionHandler(microServiceId string, module string, services *[]sdk.EndorHandlerInterface, logger *sdk.Logger) sdk.EndorHandlerInterface {
	// Use a lazy accessor so the registry core is resolved at action-call time,
	// after InitEndorHandlerRepository has been called with the correct projectLocalesFS.
	repoAccessor := func() EndorHandlerActionRepository {
		return *NewEndorHandlerActionRepository(GetRegistryCore())
	}
	entityMethodService := EntityActionHandler{
		repoAccessor: repoAccessor,
		services:     services,
	}
	return NewEndorBaseHandler[*sdk.EntityAction]("entity-action", "${t.sdk.entity_action.handler.title}").
		WithActions(map[string]sdk.EndorHandlerActionInterface{
			"schema": sdk.NewAction(
				entityMethodService.schema,
				"${t.sdk.entity_action.handler.actions.schema}",
			),
			"list": sdk.NewAction(
				entityMethodService.list,
				"${t.sdk.entity_action.handler.actions.list}",
			),
			"instance": sdk.NewAction(
				entityMethodService.instance,
				"${t.sdk.entity_action.handler.actions.instance}",
			)})
}

type EntityActionHandler struct {
	repoAccessor func() EndorHandlerActionRepository
	services     *[]sdk.EndorHandlerInterface
}

func (h *EntityActionHandler) schema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.EntityAction{})).Build(), nil
}

func (h *EntityActionHandler) list(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityAction], error) {
	repo := h.repoAccessor()
	entityMethods, err := repo.EntityActionList(c.Session)
	if err != nil {
		return nil, err
	}
	for i := range entityMethods {
		entityMethods[i].Description = c.ResolveTExpr(entityMethods[i].Description)
	}
	return sdk.NewResponseBuilder[[]sdk.EntityAction]().AddData(&entityMethods).AddSchema(sdk.NewSchema(&sdk.EntityAction{})).Build(), nil
}

func (h *EntityActionHandler) instance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityAction], error) {
	repo := h.repoAccessor()
	entityAction, err := repo.DictionaryActionInstance(c.Session, c.Payload)
	if err != nil {
		return nil, err
	}
	entityAction.entityAction.Description = c.ResolveTExpr(entityAction.entityAction.Description)
	return sdk.NewResponseBuilder[sdk.EntityAction]().AddData(&entityAction.entityAction).AddSchema(sdk.NewSchema(&sdk.EntityAction{})).Build(), nil
}
