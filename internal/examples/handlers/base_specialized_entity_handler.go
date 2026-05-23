package examples_handlers

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type BaseSpecializedEntityModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Type      string `json:"type" bson:"type" schema:"title=Type,readOnly=true"`
	Attribute string `json:"attribute" bson:"attribute"`
}

func (h *BaseSpecializedEntityModel) GetID() any {
	return h.ID
}

func (h BaseSpecializedEntityModel) GetCategoryType() string {
	return h.Type
}

func (h *BaseSpecializedEntityModel) SetCategoryType(categoryType string) {
	h.Type = categoryType
}

type BaseSpecializedModelCategory1 struct {
	BaseSpecializedEntityModel `json:",inline" bson:",inline"`
	AttributeCat1              string `json:"attributeCat1" bson:"attributeCat1"`
}

type BaseSpecializedModelCategory2 struct {
	BaseSpecializedEntityModel `json:",inline" bson:",inline"`
	AttributeCat2              string `json:"attributeCat2" bson:"attributeCat2"`
}

type BaseSpecializedAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type BaseSpecializedEntityHandler struct {
}

func (h *BaseSpecializedEntityHandler) action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, c.T("examples.base-specialized-entity.messages.hello", nil))).
		Build(), nil
}

func (h *BaseSpecializedEntityHandler) category1Action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, c.T("examples.base-specialized-entity.messages.hello-cat1", nil))).
		Build(), nil
}

func (h *BaseSpecializedEntityHandler) category2Action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, c.T("examples.base-specialized-entity.messages.hello-cat2", nil))).
		Build(), nil
}

func NewBaseSpecializedEntityHandler() sdk.EndorBaseSpecializedHandlerInterface {
	baseSpecializedHandler := BaseSpecializedEntityHandler{}

	return sdk_entity.NewEndorBaseSpecializedHandler[*BaseSpecializedEntityModel]("base-specialized-entity", "${t.examples.base-specialized-entity.handler.title}").
		WithCategories(
			[]sdk.EndorBaseSpecializedHandlerCategoryInterface{
				sdk_entity.NewEndorBaseSpecializedHandlerCategory[*BaseSpecializedModelCategory1]("cat-1", "${t.examples.base-specialized-entity.categories.cat-1.title}").
					WithExtendedDescription("${t.examples.base-specialized-entity.categories.cat-1.description}").
					WithActions(map[string]sdk.EndorHandlerActionInterface{
						"action-1": sdk.NewAction(
							baseSpecializedHandler.category1Action1,
							"${t.examples.base-specialized-entity.categories.cat-1.action1}",
						),
					}),
				sdk_entity.NewEndorBaseSpecializedHandlerCategory[*BaseSpecializedModelCategory2]("cat-2", "${t.examples.base-specialized-entity.categories.cat-2.title}").
					WithActions(map[string]sdk.EndorHandlerActionInterface{
						"action-1": sdk.NewAction(
							baseSpecializedHandler.category2Action1,
							"${t.examples.base-specialized-entity.categories.cat-2.action1}",
						),
					}),
			},
		).
		WithActions(map[string]sdk.EndorHandlerActionInterface{
			"action-1": sdk.NewAction(
				baseSpecializedHandler.action1,
				"${t.examples.base-specialized-entity.handler.action1}",
			),
		})
}
