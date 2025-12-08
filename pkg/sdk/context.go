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

	GinContext *gin.Context
	Logger     Logger
}

type NoPayload struct{}
