package main

import (
	test_utils_services "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/services"
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
		test_utils_services.NewBaseServiceService(),
		test_utils_services.NewBaseSpecializedService(),
		test_utils_services.NewHybridService(),
		test_utils_services.NewHybridSpecializedService(),
	}).Build().Init("endor-sdk-service")
}
