package sdk

import (
	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_i18n"
)

type Session struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	Development bool   `json:"development"`
}

// EndorDIContainerInterface is the dependency-injection contract exposed to handler code.
// It provides type-safe access to registered repositories without creating a circular
// dependency between the sdk and sdk_entity packages.
// Use the generic helpers GetRepositoryFromContainer and GetHandlerFromContainer
// (defined in sdk_entity) for compile-time type assertions.
type EndorDIContainerInterface interface {
	GetRepositoryByName(name string) (EndorRepositoryInterface, bool)
	GetHandlerByName(name string) (EndorHandlerInterface, bool)
}

type EndorContext[T any] struct {
	MicroServiceId string
	Session        Session
	Payload        T
	CategoryType   string
	Locale         string

	// DIContainer gives handler code access to all registered handlers and repositories
	// for the current session (production or per-user development overlay).
	DIContainer EndorDIContainerInterface

	GinContext *gin.Context
	Logger     Logger
}

// T translates the given key using named placeholder interpolation {{key}}.
func (ec *EndorContext[T]) T(key string, args map[string]any) string {
	return sdk_i18n.T(ec.Locale, key, args)
}

type NoPayload struct{}
