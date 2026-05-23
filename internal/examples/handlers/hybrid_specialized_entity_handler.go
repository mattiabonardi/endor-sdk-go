package examples_handlers

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type HybridSpecializedEntityModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Type      string `json:"type" bson:"type" schema:"title=Type,readOnly=true"`
	Attribute string `json:"attribute" bson:"attribute"`
}

func (h HybridSpecializedEntityModel) GetID() any {
	return h.ID
}

func (h HybridSpecializedEntityModel) GetCategoryType() string {
	return h.Type
}

func (h *HybridSpecializedEntityModel) SetCategoryType(categoryType string) {
	h.Type = categoryType
}

type HybridCategory1Schema struct {
	HybridSpecializedEntityModel `json:",inline" bson:",inline"`
	AttributeCat1                string `json:"attributeCat1" bson:"attributeCat1"`
}

type HybridCategory2Schema struct {
	HybridSpecializedEntityModel `json:",inline" bson:",inline"`
	AttributeCat2                string `json:"attributeCat2" bson:"attributeCat2"`
}

type HybridSpecializedAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type HybridSpecializedEntityHandler struct {
}

func (h *HybridSpecializedEntityHandler) action1(c *sdk.EndorContext[HybridSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, c.T("examples.hybrid-specialized-entity.messages.hello", nil))).
		Build(), nil
}

func NewHybridSpecializedEntityHandler() sdk.EndorHybridSpecializedHandlerInterface {
	hybridSpecializedHandler := HybridSpecializedEntityHandler{}

	return sdk_entity.NewEndorHybridSpecializedHandler[*HybridSpecializedEntityModel]("hybrid-specialized-entity", "${t.examples.hybrid-specialized-entity.handler.title}").
		WithHybridCategories(
			[]sdk.EndorHybridSpecializedHandlerCategoryInterface{
				sdk_entity.NewEndorHybridSpecializedHandlerCategory[*HybridCategory1Schema]("cat-1", "${t.examples.hybrid-specialized-entity.categories.cat-1.title}"),
				sdk_entity.NewEndorHybridSpecializedHandlerCategory[*HybridCategory2Schema]("cat-2", "${t.examples.hybrid-specialized-entity.categories.cat-2.title}"),
			},
		).
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorHandlerActionInterface {
			return map[string]sdk.EndorHandlerActionInterface{
				"action-1": sdk.NewAction(
					hybridSpecializedHandler.action1,
					"${t.examples.hybrid-specialized-entity.handler.action1}",
				),
			}
		})
}
