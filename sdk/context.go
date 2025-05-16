package sdk

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Session struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	App      string `json:"app"`
}

type EndorContext[T any] struct {
	Index    int
	Handlers []EndorHandlerFunc[T]
	Session  Session
	Payload  T
	Data     map[string]interface{}

	// bridge with gin framework
	GinContext *gin.Context
}

type NoPayload struct{}

// Continue to next middleware
func (c *EndorContext[T]) Next() {
	c.Index++
	for c.Index < len(c.Handlers) {
		c.Handlers[c.Index](c)
		return
	}
}

// feedback functions
func (c *EndorContext[T]) End(obj interface{}) {
	c.GinContext.JSON(http.StatusOK, obj)
}
func (c *EndorContext[T]) InternalServerError(err error) {
	c.GinContext.AbortWithStatusJSON(http.StatusInternalServerError, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())).Build())
}
func (c *EndorContext[T]) BadRequest(err error) {
	c.GinContext.AbortWithStatusJSON(http.StatusBadRequest, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())).Build())
}
func (c *EndorContext[T]) NotFound(err error) {
	c.GinContext.AbortWithStatusJSON(http.StatusNotFound, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())).Build())
}
func (c *EndorContext[T]) Unauthorize(err error) {
	c.GinContext.AbortWithStatusJSON(http.StatusUnauthorized, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())).Build())
}
func (c *EndorContext[T]) Forbidden(err error) {
	c.GinContext.AbortWithStatusJSON(http.StatusForbidden, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())).Build())
}
