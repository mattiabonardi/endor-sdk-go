package sdk

import "github.com/gin-gonic/gin"

type Session struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type EndorContext[T any] struct {
	MicroServiceId string
	Session        Session
	Payload        T

	GinContext *gin.Context
}

type NoPayload struct{}
