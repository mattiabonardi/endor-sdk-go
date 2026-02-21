package test_utils_handlers

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type BaseHandlerModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Attribute string `json:"attribute"`
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
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Base Handler")).
		Build(), nil
}

func (h *BaseHandler) publicAction(c *sdk.EndorContext[BaseHandlerAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from public action")).
		Build(), nil
}

func NewBaseHandlerHandler() sdk.EndorBaseHandlerInterface {
	baseHandler := BaseHandler{}
	return sdk_entity.NewEndorBaseHandler[*BaseHandlerModel]("base-handler", "Base Handler (EndorBaseHandler)").
		WithActions(map[string]sdk.EndorHandlerActionInterface{
			"action1": sdk.NewAction(
				baseHandler.action1,
				"Action 1",
			),
			"cat_1/action1": sdk.NewAction(
				baseHandler.action1,
				"Category 1 Action 1",
			),
			"public-action": sdk.NewConfigurableAction(
				sdk.EndorHandlerActionOptions{
					Description: "Public Action",
					Public:      true,
				},
				baseHandler.publicAction,
			),
		})
}
