package services_test

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"github.com/mattiabonardi/endor-sdk-go/sdk/handler"
)

type Service2 struct {
}

func (h *Service2) test1(c *sdk.EndorContext[Test1Payload]) {
	c.End(sdk.NewResponseBuilder[any]().AddMessage(sdk.NewMessage(sdk.Info, "Hello World")).Build())
}

func NewService2() sdk.EndorService {
	Service2 := Service2{}
	return sdk.EndorService{
		Resource: "test2",
		Apps:     []string{"app1", "app2"},
		Methods: map[string]sdk.EndorServiceMethod{
			"test1": sdk.NewMethod(
				handler.AuthorizationHandler,
				Service2.test1,
			),
		},
	}
}
