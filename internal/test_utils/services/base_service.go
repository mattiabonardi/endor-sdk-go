package test_utils_services

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_resource"
)

type BaseServiceModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Attribute string `json:"attribute"`
}

func (h BaseServiceModel) GetID() *string {
	return &h.ID
}

func (h *BaseServiceModel) SetID(id string) {
	h.ID = id
}

type BaseServiceAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type BaseService struct {
}

func (h *BaseService) action1(c *sdk.EndorContext[BaseServiceAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Base Service")).
		Build(), nil
}

func NewBaseServiceService() sdk.EndorBaseServiceInterface {
	baseService := BaseService{}
	return sdk_resource.NewEndorBaseService[*BaseServiceModel]("base-service", "Base Service (EndorBaseService)").
		WithActions(map[string]sdk.EndorServiceActionInterface{
			"action-1": sdk.NewAction(
				baseService.action1,
				"Action 1",
			)})
}
