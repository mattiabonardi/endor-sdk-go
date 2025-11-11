package main

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
	services_test "github.com/mattiabonardi/endor-sdk-go/test/services"
)

type Payload struct {
	Name    string `json:"name" yaml:"name"`
	Surname string `json:"surname" yaml:"surname"`
	Active  bool   `json:"active" yaml:"active"`
}

func main() {
	sdk.NewEndorInitializer().WithEndorServices(&[]sdk.EndorService{
		services_test.NewService1(),
	}).WithHybridServices(&[]sdk.EndorHybridService{
		services_test.NewService2(),
	}).Build().Init("endor-sdk-service")
}
