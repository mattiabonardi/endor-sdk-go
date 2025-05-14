package main

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"github.com/mattiabonardi/endor-sdk-go/sdk/handler"
)

type Payload struct {
	Name    string `json:"name" yaml:"name"`
	Surname string `json:"surname" yaml:"surname"`
	Active  bool   `json:"active" yaml:"active"`
}

func main() {
	testService := sdk.EndorService{
		Resource: "test",
		Methods: map[string]sdk.EndorServiceMethod{
			"method": sdk.NewMethod(
				handler.ValidationHandler,
				func(ec *sdk.EndorContext[Payload]) {
					ec.End(sdk.NewResponseBuilder[Payload]().AddData(&ec.Payload))
				},
			),
		},
	}
	sdk.Init("endor-sdk-service", []sdk.EndorService{
		testService,
	})
}
