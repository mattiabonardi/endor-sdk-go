package main

import (
	test_utils_handlers "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/handlers"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_server"
)

type Payload struct {
	Name    string `json:"name" yaml:"name"`
	Surname string `json:"surname" yaml:"surname"`
	Active  bool   `json:"active" yaml:"active"`
}

func main() {
	sdk_server.NewEndorInitializer().WithEndorHandlers(&[]sdk.EndorHandlerInterface{
		test_utils_handlers.NewBaseHandlerHandler(),
		test_utils_handlers.NewBaseSpecializedHandler(),
		test_utils_handlers.NewHybridHandler(),
		test_utils_handlers.NewHybridSpecializedHandler(),
	}).Build().Init("endor-sdk-service")
}
