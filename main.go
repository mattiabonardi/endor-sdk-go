package main

import (
	test_utils_service "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/service"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_server"
)

type Payload struct {
	Name    string `json:"name" yaml:"name"`
	Surname string `json:"surname" yaml:"surname"`
	Active  bool   `json:"active" yaml:"active"`
}

func main() {
	sdk_server.NewEndorInitializer().WithEndorServices(&[]sdk.EndorService{
		test_utils_service.NewService1(),
	}).WithHybridServices(&[]sdk.EndorHybridService{
		test_utils_service.NewService2(),
	}).Build().Init("endor-sdk-service")
}
