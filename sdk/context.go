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

// Interface implementation methods for EndorContextInterface
// These methods implement the interfaces.EndorContextInterface contract

// GetMicroServiceId returns the identifier of the microservice handling the request.
func (c *EndorContext[T]) GetMicroServiceId() string {
	return c.MicroServiceId
}

// GetSession returns the authentication session information for the request.
func (c *EndorContext[T]) GetSession() interface{} {
	return c.Session
}

// GetPayload returns the typed payload data for the request.
func (c *EndorContext[T]) GetPayload() T {
	return c.Payload
}

// SetPayload updates the payload data for the request context.
func (c *EndorContext[T]) SetPayload(payload T) {
	c.Payload = payload
}

// GetResourceMetadataSchema returns the schema definition for the resource.
func (c *EndorContext[T]) GetResourceMetadataSchema() interface{} {
	return c.ResourceMetadataSchema
}

// GetCategoryID returns the category identifier for specialized resource operations.
func (c *EndorContext[T]) GetCategoryID() *string {
	return c.CategoryID
}

// SetCategoryID sets the category identifier for specialized resource operations.
func (c *EndorContext[T]) SetCategoryID(categoryID *string) {
	c.CategoryID = categoryID
}

// GetGinContext returns the underlying Gin HTTP context for the request.
func (c *EndorContext[T]) GetGinContext() *gin.Context {
	return c.GinContext
}
