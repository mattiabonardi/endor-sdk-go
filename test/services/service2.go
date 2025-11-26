package services_test

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

type Service2BaseModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Attribute string `json:"attribute"`
}

func (h Service2BaseModel) GetID() *string {
	return &h.ID
}

func (h *Service2BaseModel) SetID(id string) {
	h.ID = id
}

type Service2Action1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Category1AdditionalSchema struct {
	AdditionalAttributeCat1 string `json:"additionalAttributeCat1"`
}

type Category2Schema struct {
	CategoryType  string `json:"categoryType" bson:"categoryType"`
	AttributeCat2 string `json:"attributeCat2"`
}

func (c Category2Schema) GetCategoryType() *string {
	return &c.CategoryType
}

func (c *Category2Schema) SetCategoryType(categoryType string) {
	c.CategoryType = categoryType
}

type Category2AdditionalSchema struct {
	AdditionalAttributeCat2 string `json:"additionalAttributeCat2"`
}

type Service2 struct {
}

func (h *Service2) action1(c *sdk.EndorContext[Service2Action1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.Info, "Hello from Hybrid Service")).
		Build(), nil
}

func NewService2() sdk.EndorHybridService {
	service2 := Service2{}

	category1AdditionalSchema, _ := sdk.NewSchema(Category1AdditionalSchema{}).ToYAML()
	category2AdditionalSchema, _ := sdk.NewSchema(Category2AdditionalSchema{}).ToYAML()

	return sdk.NewHybridService("resource-2", "Resource 2 (EndorHybridService with static categories)").
		WithBaseModel(&Service2BaseModel{}).
		WithCategories([]sdk.Category{
			{
				ID:                   "cat-1",
				Description:          "Category 1",
				AdditionalAttributes: category1AdditionalSchema,
			},
			{
				ID:                   "cat-2",
				Description:          "Category 2",
				BaseModel:            &Category2Schema{},
				AdditionalAttributes: category2AdditionalSchema,
			},
		}).
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
			return map[string]sdk.EndorServiceAction{
				"action-1": sdk.NewAction(
					service2.action1,
					"Test hybrid action",
				),
			}
		})
}
