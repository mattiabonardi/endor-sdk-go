package services_test

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

type Test1Payload struct {
	String  string `json:"string"`
	Boolean bool   `json:"boolean"`
}

type Test2Payload struct {
	Array       []string                 `json:"array"`
	ObjectArray []Test2PayloadArrayIteam `json:"objectArray"`
}

type Test2PayloadArrayIteam struct {
	String string `json:"string"`
}

type GenericPayload struct {
	String string `json:"string"`
}

type Test4Payload[T any] struct {
	Value T `json:"value"`
}

type Test5EventPayload struct {
	Value string `json:"value"`
}

type Service1 struct {
}

func (h *Service1) test1(c *sdk.EndorContext[Test1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build(), nil
}

func (h *Service1) test2(c *sdk.EndorContext[Test2Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build(), nil
}

func (h *Service1) test3(c *sdk.EndorContext[Test2Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build(), nil
}

func (h *Service1) test4(c *sdk.EndorContext[Test4Payload[GenericPayload]]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build(), nil
}

func (h *Service1) test5(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	// emit event
	c.EmitEvent("test5Completed", Test5EventPayload{
		Value: "Hello World",
	})
	return sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build(), nil
}

func NewService1() sdk.EndorService {
	Service1 := Service1{}
	priority := 99
	return sdk.EndorService{
		Resource:    "test",
		Description: "Testing resource",
		Priority:    &priority,
		Methods: map[string]sdk.EndorServiceAction{
			"test1": sdk.NewAction(
				Service1.test1,
				"description 1",
			),
			"test2": sdk.NewAction(
				Service1.test2,
				"description 2",
			),
			"test3": sdk.NewAction(
				Service1.test3,
				"description 3",
			),
			"test4": sdk.NewAction(
				Service1.test4,
				"description 4",
			),
			"test5": sdk.NewConfigurableActionWithEvents(
				sdk.EndorServiceActionOptions{
					Public:          true,
					ValidatePayload: false,
				},
				Service1.test5,
				sdk.NewEventDefinition[Test5EventPayload]("test5Completed", "Emit when test 5 action is completed"),
			),
		},
	}
}
