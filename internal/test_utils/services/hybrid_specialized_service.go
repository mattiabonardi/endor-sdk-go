package test_utils_services

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_resource"
)

type HybridSpecializedModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Type      string `json:"type" bson:"type" schema:"title=Type,readOnly=true"`
	Attribute string `json:"attribute"`
}

func (h HybridSpecializedModel) GetID() *string {
	return &h.ID
}

func (h *HybridSpecializedModel) SetID(id string) {
	h.ID = id
}

func (h HybridSpecializedModel) GetCategoryType() *string {
	return &h.Type
}

func (h *HybridSpecializedModel) SetCategoryType(categoryType string) {
	h.Type = categoryType
}

type HybridCategory1AdditionalSchema struct {
	AdditionalAttributeCat1 string `json:"additionalAttributeCat1"`
}

type HybridCategory2Schema struct {
	AttributeCat2 string `json:"attributeCat2"`
}

type HybridCategory2AdditionalSchema struct {
	AdditionalAttributeCat2 string `json:"additionalAttributeCat2"`
}

type HybridSpecializedAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type HybridSpecializedService struct {
}

func (h *HybridSpecializedService) action1(c *sdk.EndorContext[HybridSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Hybrid Service")).
		Build(), nil
}

func NewHybridSpecializedService() sdk.EndorHybridSpecializedServiceInterface {
	hybridSpecializedService := HybridSpecializedService{}
	category1AdditionalSchema, _ := sdk.NewSchema(Category1AdditionalSchema{}).ToYAML()
	category2AdditionalSchema, _ := sdk.NewSchema(Category2AdditionalSchema{}).ToYAML()

	return sdk_resource.NewHybridSpecializedService[*HybridSpecializedModel]("resource-3", "Resource 3 (EndorHybridSpecializedService with static categories)").
		WithCategories(
			[]sdk.EndorHybridSpecializedServiceCategoryInterface{
				sdk_resource.NewEndorHybridSpecializedServiceCategory[*HybridSpecializedModel, *Category1AdditionalSchema](sdk.HybridCategory{
					ID:                   "cat-1",
					Description:          "Category 1",
					AdditionalAttributes: category1AdditionalSchema,
				}),
				sdk_resource.NewEndorHybridSpecializedServiceCategory[*HybridSpecializedModel, *Category2Schema](sdk.HybridCategory{
					ID:                   "cat-2",
					Description:          "Category 2",
					AdditionalAttributes: category2AdditionalSchema,
				}),
			},
		).
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface {
			return map[string]sdk.EndorServiceActionInterface{
				"action-1": sdk.NewAction(
					hybridSpecializedService.action1,
					"Test hybrid action",
				),
			}
		})
}
