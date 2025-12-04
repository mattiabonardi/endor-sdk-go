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
	MicroServiceId         string
	Session                Session
	Payload                T
	ResourceMetadataSchema RootSchema
	CategoryID             *string // ID della categoria per API specializzate, nil se non categorizzata

	GinContext *gin.Context
}

type NoPayload struct{}
