package test_utils_handlers

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type BaseSpecializedModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Type      string `json:"type" bson:"type" schema:"title=Type,readOnly=true"`
	Attribute string `json:"attribute" bson:"attribute"`
}

func (h *BaseSpecializedModel) GetID() any {
	return h.ID
}

func (h BaseSpecializedModel) GetCategoryType() string {
	return h.Type
}

func (h *BaseSpecializedModel) SetCategoryType(categoryType string) {
	h.Type = categoryType
}

type BaseSpecializedModelCategory1 struct {
	BaseSpecializedModel `json:",inline" bson:",inline"`
	AttributeCat1        string `json:"attributeCat1" bson:"attributeCat1"`
}

type BaseSpecializedModelCategory2 struct {
	BaseSpecializedModel `json:",inline" bson:",inline"`
	AttributeCat2        string `json:"attributeCat2" bson:"attributeCat2"`
}

type BaseSpecializedAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type BaseSpecializedHandler struct {
}

func (h *BaseSpecializedHandler) action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Hybrid Handler")).
		Build(), nil
}

func (h *BaseSpecializedHandler) category1Action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from category 1 action 1")).
		Build(), nil
}

func (h *BaseSpecializedHandler) category2Action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from category 2 action 1")).
		Build(), nil
}

func NewBaseSpecializedHandler() sdk.EndorBaseSpecializedHandlerInterface {
	baseSpecializedHandler := BaseSpecializedHandler{}

	return sdk_entity.NewEndorBaseSpecializedHandler[*BaseSpecializedModel]("base-specialized-handler", "Base Specialized Handler (EndorBaseSpecializedHandler)").
		WithCategories(
			[]sdk.EndorBaseSpecializedHandlerCategoryInterface{
				sdk_entity.NewEndorBaseSpecializedHandlerCategory[*BaseSpecializedModelCategory1]("cat-1", "Category 1").
					WithActions(map[string]sdk.EndorHandlerActionInterface{
						"action-1": sdk.NewAction(
							baseSpecializedHandler.category1Action1,
							"Action 1",
						),
					}),
				sdk_entity.NewEndorBaseSpecializedHandlerCategory[*BaseSpecializedModelCategory2]("cat-2", "Category 2").
					WithActions(map[string]sdk.EndorHandlerActionInterface{
						"action-1": sdk.NewAction(
							baseSpecializedHandler.category2Action1,
							"Action 1",
						),
					}),
			},
		).
		WithActions(map[string]sdk.EndorHandlerActionInterface{
			"action-1": sdk.NewAction(
				baseSpecializedHandler.action1,
				"Action 1",
			),
		},
		)
}
