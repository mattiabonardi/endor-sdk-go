package models

import "github.com/gin-gonic/gin"

type Handler interface {
	Route(r *gin.RouterGroup)
}

type HandlerFn[T any] func(T, *gin.Context)

type NoPayloadHandler func(*gin.Context)
