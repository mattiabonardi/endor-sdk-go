package test_utils_services

type BaseSpecializedModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Type      string `json:"type" bson:"type" schema:"title=Type,readOnly=true"`
	Attribute string `json:"attribute"`
}

func (h BaseSpecializedModel) GetID() *string {
	return &h.ID
}

func (h *BaseSpecializedModel) SetID(id string) {
	h.ID = id
}

func (h BaseSpecializedModel) GetCategoryType() *string {
	return &h.Type
}

func (h *BaseSpecializedModel) SetCategoryType(categoryType string) {
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

type BaseSpecializedAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

/*type BaseSpecializedService struct {
}

func (h *BaseSpecializedService) action1(c *sdk.EndorContext[BaseSpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Hybrid Service")).
		Build(), nil
}

func NewBaseSpecializedService() sdk.EndorHybridSpecializedServiceInterface {
	baseSpecializedService := BaseSpecializedService{}
	category1AdditionalSchema, _ := sdk.NewSchema(Category1AdditionalSchema{}).ToYAML()
	category2AdditionalSchema, _ := sdk.NewSchema(Category2AdditionalSchema{}).ToYAML()

	return sdk_resource.NewHybridSpecializedService[*BaseSpecializedModel]("resource-3", "Resource 3 (EndorHybridSpecializedService with static categories)").
		WithCategories(
			[]sdk.EndorHybridSpecializedServiceCategoryInterface{
				sdk_resource.NewEndorHybridSpecializedServiceCategory[*BaseSpecializedModel, *Category1AdditionalSchema](sdk.HybridCategory{
					ID:                   "cat-1",
					Description:          "Category 1",
					AdditionalAttributes: category1AdditionalSchema,
				}),
				sdk_resource.NewEndorHybridSpecializedServiceCategory[*BaseSpecializedModel, *Category2Schema](sdk.HybridCategory{
					ID:                   "cat-2",
					Description:          "Category 2",
					AdditionalAttributes: category2AdditionalSchema,
				}),
			},
		).
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface {
			return map[string]sdk.EndorServiceActionInterface{
				"action-1": sdk.NewAction(
					baseSpecializedService.action1,
					"Test hybrid action",
				),
			}
		})
}
*/
