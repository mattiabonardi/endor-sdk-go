package handler

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

func ValidationHandler[T any](c *sdk.EndorContext[T]) {
	if err := c.GinContext.ShouldBindJSON(&c.Payload); err != nil {
		c.BadRequest(err)
		return
	}
	c.Next()
}
