package examples_handlers

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type BaseHandlerModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Attribute string `json:"attribute" schema:"title=${t.examples.base-handler.fields.attribute}"`
}

func (h BaseHandlerModel) GetID() any {
	return h.ID
}

type BaseHandlerAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type BaseHandler struct {
}

func (h *BaseHandler) action1(c *sdk.EndorContext[BaseHandlerAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, c.T("examples.base-handler.messages.hello", map[string]any{}))).
		Build(), nil
}

func (h *BaseHandler) publicAction(c *sdk.EndorContext[BaseHandlerAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, c.T("examples.base-handler.messages.hello-public", map[string]any{}))).
		Build(), nil
}

func NewBaseHandlerHandler() sdk.EndorBaseHandlerInterface {
	baseHandler := BaseHandler{}
	return sdk_entity.NewEndorBaseHandler[*BaseHandlerModel]("base-handler", "${t.examples.base-handler.handler.title}").
		WithExtendedDescription("${t.examples.base-handler.handler.description}").WithActions(map[string]sdk.EndorHandlerActionInterface{
		"action1": sdk.NewAction(
			baseHandler.action1,
			"${t.examples.base-handler.handler.action1}",
		),
		"public-action": sdk.NewConfigurableAction(
			sdk.EndorHandlerActionOptions{
				Description: "${t.examples.base-handler.handler.public-actions}",
				Public:      true,
			},
			baseHandler.publicAction,
		),
	})
}
