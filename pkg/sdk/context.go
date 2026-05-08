package sdk

import (
	"github.com/gin-gonic/gin"
)

type Session struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	Development bool   `json:"development"`
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
	return ec.DIContainer.GetTranslator().T(ec.Locale, key, args)
}

// ResolveTExpr resolves t(<token>) expressions in value using the request locale.
func (ec *EndorContext[T]) ResolveTExpr(value string) string {
	return ec.DIContainer.GetTranslator().ResolveTExpr(ec.Locale, value)
}

type NoPayload struct{}
