package models

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Session struct {
	Id    string `json:"id"`
	User  string `json:"user"`
	Email string `json:"email"`
	App   string `json:"app"`
}

type EndorContext struct {
	Index    int
	Handlers []EndorHandlerFunc
	Session  Session
	Data     map[string]interface{}

	// bridge with gin framework
	GinContext *gin.Context
}

// Continue to next middleware
func (c *EndorContext) Next() {
	c.Index++
	for c.Index < len(c.Handlers) {
		c.Handlers[c.Index](c)
		return
	}
}

// feedback functions
func (c *EndorContext) End(obj interface{}) {
	c.GinContext.JSON(200, obj)
}
func (c *EndorContext) InternalServerError(err error) {
	c.GinContext.AbortWithStatusJSON(http.StatusInternalServerError, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())).Build())
}
