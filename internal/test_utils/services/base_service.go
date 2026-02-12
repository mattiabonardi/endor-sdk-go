package test_utils_services

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

type BaseServiceModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Attribute string `json:"attribute"`
}

func (h BaseServiceModel) GetID() string {
	return h.ID
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

func (h *BaseService) publicAction(c *sdk.EndorContext[BaseServiceAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from public action")).
		Build(), nil
}

func NewBaseServiceService() sdk.EndorBaseServiceInterface {
	baseService := BaseService{}
	return sdk_entity.NewEndorBaseService[*BaseServiceModel]("base-service", "Base Service (EndorBaseService)").
		WithActions(map[string]sdk.EndorServiceActionInterface{
			"action1": sdk.NewAction(
				baseService.action1,
				"Action 1",
			),
			"cat_1/action1": sdk.NewAction(
				baseService.action1,
				"Category 1 Action 1",
			),
			"public-action": sdk.NewConfigurableAction(
				sdk.EndorServiceActionOptions{
					Description: "Public Action",
					Public:      true,
				},
				baseService.publicAction,
			),
		})
}
