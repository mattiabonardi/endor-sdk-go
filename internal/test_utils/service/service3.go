package test_utils_service

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_resource"
)

type Service3BaseModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Type      string `json:"type" bson:"type" schema:"title=Type,readOnly=true"`
	Attribute string `json:"attribute"`
}

func (h Service3BaseModel) GetID() *string {
	return &h.ID
}

func (h *Service3BaseModel) SetID(id string) {
	h.ID = id
}

func (h Service3BaseModel) GetCategoryType() *string {
	return &h.Type
}

func (h *Service3BaseModel) SetCategoryType(categoryType string) {
	h.Type = categoryType
}

type Category1AdditionalSchema struct {
	AdditionalAttributeCat1 string `json:"additionalAttributeCat1"`
}

type Category2Schema struct {
	AttributeCat2 string `json:"attributeCat2"`
}

type Category2AdditionalSchema struct {
	AdditionalAttributeCat2 string `json:"additionalAttributeCat2"`
}

type Service3Action1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Service3 struct {
}

func (h *Service3) action1(c *sdk.EndorContext[Service2Action1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Hybrid Service")).
		Build(), nil
}

func NewService3() sdk.EndorHybridSpecializedServiceInterface {
	service3 := Service3{}
	category1AdditionalSchema, _ := sdk.NewSchema(Category1AdditionalSchema{}).ToYAML()
	category2AdditionalSchema, _ := sdk.NewSchema(Category2AdditionalSchema{}).ToYAML()

	return sdk_resource.NewHybridSpecializedService[*Service3BaseModel]("resource-3", "Resource 3 (EndorHybridSpecializedService with static categories)").
		WithCategories(
			[]sdk.EndorHybridSpecializedServiceCategoryInterface{
				sdk_resource.NewEndorHybridSpecializedServiceCategory[*Service3BaseModel, *Category1AdditionalSchema](sdk.Category{
					ID:                   "cat-1",
					Description:          "Category 1",
					AdditionalAttributes: category1AdditionalSchema,
				}),
				sdk_resource.NewEndorHybridSpecializedServiceCategory[*Service3BaseModel, *Category2Schema](sdk.Category{
					ID:                   "cat-2",
					Description:          "Category 2",
					AdditionalAttributes: category2AdditionalSchema,
				}),
			},
		).
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
			return map[string]sdk.EndorServiceAction{
				"action-1": sdk.NewAction(
					service3.action1,
					"Test hybrid action",
				),
			}
		})
}
