package main

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	services_test "github.com/mattiabonardi/endor-sdk-go/test/services"
)

type Payload struct {
	Name    string `json:"name" yaml:"name"`
	Surname string `json:"surname" yaml:"surname"`
	Active  bool   `json:"active" yaml:"active"`
}

func main() {
	endor, err := sdk.NewEndorInitializer().WithEndorServices(&[]sdk.EndorService{
		services_test.NewService1(),
	}).WithHybridServices(&[]sdk.EndorHybridService{
		services_test.NewService2(),
	}).Build()

	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Endor: %v", err))
	}

	endor.Init("endor-sdk-service")
}
