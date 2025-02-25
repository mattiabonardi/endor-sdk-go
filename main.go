package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/middlewares"
	"github.com/mattiabonardi/endor-sdk-go/models"
	"github.com/mattiabonardi/endor-sdk-go/server"
)

type TestHandler1 struct{}

func (h TestHandler1) Route(r *gin.RouterGroup) {
	test := r.Group("test1")
	test.GET("/ping", middlewares.NoPayloadDispatch(h.ping))
	test.GET("/pong", middlewares.NoPayloadDispatch(h.pong))
}

func (h TestHandler1) ping(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewResponseBuilder().AddData("pong").Build())
}

func (h TestHandler1) pong(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewResponseBuilder().AddData("ping").Build())
}

type TestHandler2 struct{}

func (h TestHandler2) Route(r *gin.RouterGroup) {
	test := r.Group("test2")
	test.GET("/ping", middlewares.NoPayloadDispatch(h.ping))
	test.GET("/pong", middlewares.NoPayloadDispatch(h.pong))
}

func (h TestHandler2) ping(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewResponseBuilder().AddData("pong").Build())
}

func (h TestHandler2) pong(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewResponseBuilder().AddData("ping").Build())
}

func main() {
	handlers := []models.Handler{}
	handlers = append(handlers, TestHandler1{})
	handlers = append(handlers, TestHandler2{})
	server.Init("endor-test", handlers)
}
