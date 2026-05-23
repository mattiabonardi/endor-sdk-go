package examples_handlers

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type baseEntityHandlerModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Attribute string `json:"attribute" schema:"title=${t.examples.base-entity.fields.attribute}"`
}

func (h baseEntityHandlerModel) GetID() any {
	return h.ID
}

type baseEntityHandlerAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type baseEntityHandler struct {
}

func (h *baseEntityHandler) action1(c *sdk.EndorContext[baseEntityHandlerAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, c.T("examples.base-entity.messages.hello", map[string]any{}))).
		Build(), nil
}

func (h *baseEntityHandler) publicAction(c *sdk.EndorContext[baseEntityHandlerAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, c.T("examples.base-entity.messages.hello-public", map[string]any{}))).
		Build(), nil
}

func NewBaseEntityHandler() sdk.EndorBaseHandlerInterface {
	baseHandler := baseEntityHandler{}
	return sdk_entity.NewEndorBaseHandler[*baseEntityHandlerModel]("base-entity", "${t.examples.base-entity.handler.title}").
		WithExtendedDescription("${t.examples.base-entity.handler.description}").WithActions(map[string]sdk.EndorHandlerActionInterface{
		"action1": sdk.NewAction(
			baseHandler.action1,
			"${t.examples.base-entity.handler.action1}",
		),
		"public-action": sdk.NewConfigurableAction(
			sdk.EndorHandlerActionOptions{
				Description: "${t.examples.base-entity.handler.public-actions}",
				Public:      true,
			},
			baseHandler.publicAction,
		),
	})
}
