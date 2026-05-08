package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_i18n"
)

// EndorDIContainer is the runtime dependency-injection container passed to every handler
// action via EndorContext.DIContainer. It holds the merged repository map and the
// merged translator built from all DSL locale paths active for the current session.
type EndorDIContainer struct {
	repositories map[string]sdk.EndorRepositoryInterface
	translator   *sdk_i18n.Translator
}

func (c *EndorDIContainer) GetRepositories() map[string]sdk.EndorRepositoryInterface {
	return c.repositories
}

func (c *EndorDIContainer) GetTranslator() *sdk_i18n.Translator {
	return c.translator
}
