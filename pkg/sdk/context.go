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

type EndorContext[T any] struct {
	MicroServiceId string
	Session        Session
	Payload        T
	CategoryType   string
	Locale         string

	GinContext *gin.Context
	Logger     Logger
}

// T translates the given key using the locale resolved from the HTTP Accept-Language header.
// Supports optional fmt.Sprintf-style args for interpolation.
func (ec *EndorContext[T]) T(key string, args ...interface{}) string {
	return sdk_i18n.T(ec.Locale, key, args...)
}

type NoPayload struct{}
