package test_utils_service

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_resource"
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

type Service2 struct {
}

func (h *Service2) action1(c *sdk.EndorContext[Service2Action1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello from Hybrid Service")).
		Build(), nil
}

func NewService2() sdk.EndorHybridServiceInterface {
	service2 := Service2{}
	return sdk_resource.NewHybridService[*Service2BaseModel]("resource-2", "Resource 2 (EndorHybridService)").
		WithActions(func(getSchema func() sdk.RootSchema) map[string]sdk.EndorServiceAction {
			return map[string]sdk.EndorServiceAction{
				"action-1": sdk.NewAction(
					service2.action1,
					"Test hybrid action",
				),
			}
		})
}
