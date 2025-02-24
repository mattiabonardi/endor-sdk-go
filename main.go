package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/middlewares"
	"github.com/mattiabonardi/endor-sdk-go/models"
	"github.com/mattiabonardi/endor-sdk-go/server"
)

type TestHandler struct{}

func (h TestHandler) Route(r *gin.RouterGroup) {
	test := r.Group("test")
	test.GET("/ping", middlewares.NoPayloadDispatch(h.ping))
	test.GET("/pong", middlewares.NoPayloadDispatch(h.pong))
}

func (h TestHandler) ping(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewResponseBuilder().AddData("pong").Build())
}

func (h TestHandler) pong(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewResponseBuilder().AddData("ping").Build())
}

func main() {
	handlers := []models.Handler{}
	handlers = append(handlers, TestHandler{})
	server.Init("endor-test", handlers)
}
