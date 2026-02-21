package test_utils_handlers

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type HybridHandlerModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Attribute string `json:"attribute"`
}

func (h HybridHandlerModel) GetID() any {
	return h.ID
}

type HybridHandlerModelAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type HybridHandler struct {
}

func (h *HybridHandler) action1(c *sdk.EndorContext[HybridHandlerModelAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Hybrid Handler")).
		Build(), nil
}

func NewHybridHandler() sdk.EndorHybridHandlerInterface {
	hybridHandler := HybridHandler{}
	return sdk_entity.NewEndorHybridHandler[*HybridHandlerModel]("hybrid-handler", "Hybrid Handler (EndorHybridHandler)").
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface {
			return map[string]sdk.EndorHandlerActionInterface{
				"action-1": sdk.NewAction(
					hybridHandler.action1,
					"Test hybrid action",
				),
			}
		})
}
