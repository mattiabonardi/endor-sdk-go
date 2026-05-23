package main

import (
	"embed"

	examples_handlers "github.com/mattiabonardi/endor-sdk-go/internal/examples/handlers"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_server"
)

type Payload struct {
	Name    string `json:"name" yaml:"name"`
	Surname string `json:"surname" yaml:"surname"`
	Active  bool   `json:"active" yaml:"active"`
}

//go:embed locales
var locales embed.FS

func main() {
	sdk_server.NewEndorInitializer().WithEndorHandlers(&[]sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
		examples_handlers.NewBaseSpecializedEntityHandler(),
		examples_handlers.NewHybridEntityHandler(),
		examples_handlers.NewHybridSpecializedEntityHandler(),
	}).WithLocalesFS(locales).Build().Init("sdk")
}
