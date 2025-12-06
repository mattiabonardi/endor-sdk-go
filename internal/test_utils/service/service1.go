package test_utils_service

import "github.com/mattiabonardi/endor-sdk-go/pkg/sdk"

type Service1Action1Payload struct {
	Greet string `json:"greet"`
}

type Service1 struct {
}

func (h *Service1) action1(c *sdk.EndorContext[Service1Action1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello World")).Build(), nil
}

func (h *Service1) cat1_action1(c *sdk.EndorContext[Service1Action1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Hello World")).Build(), nil
}

func (h *Service1) publicAction(c *sdk.EndorContext[Service1Action1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "Public Hello World")).Build(), nil
}

func NewService1() sdk.EndorService {
	service := Service1{}
	return sdk.EndorService{
		Resource:    "resource-1",
		Description: "Resource 1 (EndorService with categories)",
		Methods: map[string]sdk.EndorServiceAction{
			"action1": sdk.NewAction(
				service.action1,
				"Action 1",
			),
			"cat_1/action1": sdk.NewAction(
				service.cat1_action1,
				"Action specified for category",
			),
			"public-action": sdk.NewConfigurableAction(
				sdk.EndorServiceActionOptions{
					Description:     "Public Action that doesn't require authentication",
					Public:          true,
					ValidatePayload: true,
				},
				service.publicAction,
			),
		},
	}
}
