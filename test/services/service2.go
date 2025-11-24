package services_test

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

type Service2BaseModel struct {
	ID        string `json:"id" bson:"_id"`
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

type Category1Schema struct {
	AdditionalAttributeCat1 string `json:"additionalAttributeCat1"`
}

type Category2Schema struct {
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

	category1Schema, _ := sdk.NewSchema(Category1Schema{}).ToYAML()
	category2Schema, _ := sdk.NewSchema(Category2Schema{}).ToYAML()

	return sdk.NewHybridService("resource-2", "Resource 2 (EndorHybridService with static categories)").
		WithBaseModel(&Service2BaseModel{}).
		WithCategories([]sdk.Category{
			{
				ID:                   "cat-1",
				Description:          "Category 1",
				AdditionalAttributes: category1Schema,
			},
			{
				ID:                   "cat-2",
				Description:          "Category 2",
				AdditionalAttributes: category2Schema,
			},
		}).
		WithActions(func() map[string]sdk.EndorServiceAction {
			return map[string]sdk.EndorServiceAction{
				"action-1": sdk.NewAction(
					service2.action1,
					"Test hybrid action",
				),
			}
		})
}
