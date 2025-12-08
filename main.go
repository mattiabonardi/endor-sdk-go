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
	sdk_server.NewEndorInitializer().WithEndorServices(&[]sdk.EndorServiceInterface{
		test_utils_service.NewService1(),
		test_utils_service.NewService2(),
		test_utils_service.NewService3(),
	}).Build().Init("endor-sdk-service")
}
