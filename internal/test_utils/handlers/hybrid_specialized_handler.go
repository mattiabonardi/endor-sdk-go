package test_utils_handlers

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type HybridSpecializedModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Type      string `json:"type" bson:"type" schema:"title=Type,readOnly=true"`
	Attribute string `json:"attribute" bson:"attribute"`
}

func (h HybridSpecializedModel) GetID() any {
	return h.ID
}

func (h HybridSpecializedModel) GetCategoryType() string {
	return h.Type
}

func (h *HybridSpecializedModel) SetCategoryType(categoryType string) {
	h.Type = categoryType
}

type HybridCategory1Schema struct {
	HybridSpecializedModel `json:",inline" bson:",inline"`
	AttributeCat1          string `json:"attributeCat1" bson:"attributeCat1"`
}

type HybridCategory2Schema struct {
	HybridSpecializedModel `json:",inline" bson:",inline"`
	AttributeCat2          string `json:"attributeCat2" bson:"attributeCat2"`
}

type HybridSpecializedAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type HybridSpecializedHandler struct {
}

func (h *HybridSpecializedHandler) action1(c *sdk.EndorContext[HybridSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Hybrid Specialized Handler")).
		Build(), nil
}

func NewHybridSpecializedHandler() sdk.EndorHybridSpecializedHandlerInterface {
	hybridSpecializedHandler := HybridSpecializedHandler{}

	return sdk_entity.NewEndorHybridSpecializedHandler[*HybridSpecializedModel]("hybrid-specialized-handler", "Hybrid Specialized Handler (EndorHybridSpecializedHandler)").
		WithHybridCategories(
			[]sdk.EndorHybridSpecializedHandlerCategoryInterface{
				sdk_entity.NewEndorHybridSpecializedHandlerCategory[*HybridCategory1Schema]("cat-1", "Category 1"),
				sdk_entity.NewEndorHybridSpecializedHandlerCategory[*HybridCategory2Schema]("cat-2", "Category 2"),
			},
		).
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface {
			return map[string]sdk.EndorHandlerActionInterface{
				"action-1": sdk.NewAction(
					hybridSpecializedHandler.action1,
					"Test hybrid action",
				),
			}
		})
}
