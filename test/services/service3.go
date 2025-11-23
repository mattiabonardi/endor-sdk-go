package services_test

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

type Service3Action1Payload struct {
	Greet string `json:"greet"`
}

type Service3 struct {
}

func (h *Service3) action1(c *sdk.EndorContext[Service3Action1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build(), nil
}

func (h *Service3) cat1_action1(c *sdk.EndorContext[Service3Action1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build(), nil
}

func NewService3() sdk.EndorService {
	service := Service3{}
	return sdk.EndorService{
		Resource:    "resource-3",
		Description: "Resource 3 (EndorService with categories)",
		Methods: map[string]sdk.EndorServiceAction{
			"action1": sdk.NewAction(
				service.action1,
				"Action 1",
			),
			"cat_1/action1": sdk.NewAction(
				service.cat1_action1,
				"Action specified for category",
			),
		},
	}
}
