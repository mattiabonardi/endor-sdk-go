package services_test

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"github.com/mattiabonardi/endor-sdk-go/sdk/handler"
)

type Test1Payload struct {
	String  string `json:"string" yaml:"string"`
	Boolean bool   `json:"boolean" yaml:"boolean"`
}

type Test2Payload struct {
	Array       []string                 `json:"array" yaml:"array"`
	ObjectArray []Test2PayloadArrayIteam `json:"objectArray" yaml:"objectArray"`
}

type Test2PayloadArrayIteam struct {
	String string `json:"string" yaml:"string"`
}

type GenericPayload struct {
	String string `json:"string" yaml:"string"`
}

type Test4Payload[T any] struct {
	Value T `json:"value" yaml:"value"`
}

type Service1 struct {
}

func (h *Service1) test1(c *sdk.EndorContext[Test1Payload]) {
	c.End(sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build())
}

func (h *Service1) test2() func(c *sdk.EndorContext[Test2Payload]) {
	return func(c *sdk.EndorContext[Test2Payload]) {
		c.End(sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build())
	}
}

func (h *Service1) test3() func(c *sdk.EndorContext[Test2Payload]) {
	return func(c *sdk.EndorContext[Test2Payload]) {
		c.End(sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build())
	}
}

func (h *Service1) test4() func(c *sdk.EndorContext[Test4Payload[GenericPayload]]) {
	return func(c *sdk.EndorContext[Test4Payload[GenericPayload]]) {
		c.End(sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build())
	}
}

func NewService1() sdk.EndorService {
	Service1 := Service1{}
	return sdk.EndorService{
		Resource: "test",
		Methods: map[string]sdk.EndorServiceMethod{
			"test1": sdk.NewMethod(
				handler.AuthorizationHandler,
				Service1.test1,
			),
			"test2": sdk.NewMethod(
				handler.AuthorizationHandler,
				Service1.test2(),
			),
			"test3": sdk.NewMethod(
				handler.AuthorizationHandler,
				Service1.test3(),
			),
			"test4": sdk.NewMethod(
				handler.AuthorizationHandler,
				Service1.test4(),
			),
		},
	}
}
