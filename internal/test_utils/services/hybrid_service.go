package test_utils_services

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_resource"
)

type HybridServiceModel struct {
	ID        string `json:"id" bson:"_id" schema:"title=Id,readOnly=true"`
	Attribute string `json:"attribute"`
}

func (h HybridServiceModel) GetID() *string {
	return &h.ID
}

func (h *HybridServiceModel) SetID(id string) {
	h.ID = id
}

type HybridServiceModelAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type HybridService struct {
}

func (h *HybridService) action1(c *sdk.EndorContext[HybridServiceModelAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Hybrid Service")).
		Build(), nil
}

func NewHybridService() sdk.EndorHybridServiceInterface {
	hybridService := HybridService{}
	return sdk_resource.NewEndorHybridService[*HybridServiceModel]("hybrid-service", "Hybrid Service (EndorHybridService)").
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceActionInterface {
			return map[string]sdk.EndorServiceActionInterface{
				"action-1": sdk.NewAction(
					hybridService.action1,
					"Test hybrid action",
				),
			}
		})
}
