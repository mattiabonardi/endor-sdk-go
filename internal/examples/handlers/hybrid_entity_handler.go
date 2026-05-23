package examples_handlers

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type HybridEntityModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Attribute string `json:"attribute"`
}

func (h HybridEntityModel) GetID() any {
	return h.ID
}

type HybridHandlerModelAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type HybridEntityHandler struct {
}

func (h *HybridEntityHandler) action1(c *sdk.EndorContext[HybridHandlerModelAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, c.T("examples.hybrid-entity.messages.hello", nil))).
		Build(), nil
}

func NewHybridEntityHandler() sdk.EndorHybridHandlerInterface {
	hybridHandler := HybridEntityHandler{}
	return sdk_entity.NewEndorHybridHandler[*HybridEntityModel]("hybrid-entity", "${t.examples.hybrid-entity.handler.title}").
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface {
			return map[string]sdk.EndorHandlerActionInterface{
				"action-1": sdk.NewAction(
					hybridHandler.action1,
					"${t.examples.hybrid-entity.handler.action1}",
				),
			}
		})
}
