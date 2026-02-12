package test_utils_services

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type BaseSpecializedModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Type      string `json:"type" bson:"type" schema:"title=Type,readOnly=true"`
	Attribute string `json:"attribute" bson:"attribute"`
}

func (h *BaseSpecializedModel) GetID() string {
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

type BaseSpecializedService struct {
}

func (h *BaseSpecializedService) action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Hybrid Service")).
		Build(), nil
}

func (h *BaseSpecializedService) category1Action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from category 1 action 1")).
		Build(), nil
}

func (h *BaseSpecializedService) category2Action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from category 2 action 1")).
		Build(), nil
}

func NewBaseSpecializedService() sdk.EndorBaseSpecializedServiceInterface {
	baseSpecializedService := BaseSpecializedService{}

	return sdk_entity.NewEndorBaseSpecializedService[*BaseSpecializedModel]("base-specialized-service", "Base Specialized Service (EndorBaseSpecializedService)").
		WithCategories(
			[]sdk.EndorBaseSpecializedServiceCategoryInterface{
				sdk_entity.NewEndorBaseSpecializedServiceCategory[*BaseSpecializedModelCategory1]("cat-1", "Category 1").
					WithActions(map[string]sdk.EndorServiceActionInterface{
						"action-1": sdk.NewAction(
							baseSpecializedService.category1Action1,
							"Action 1",
						),
					}),
				sdk_entity.NewEndorBaseSpecializedServiceCategory[*BaseSpecializedModelCategory2]("cat-2", "Category 2").
					WithActions(map[string]sdk.EndorServiceActionInterface{
						"action-1": sdk.NewAction(
							baseSpecializedService.category2Action1,
							"Action 1",
						),
					}),
			},
		).
		WithActions(map[string]sdk.EndorServiceActionInterface{
			"action-1": sdk.NewAction(
				baseSpecializedService.action1,
				"Action 1",
			),
		},
		)
}
