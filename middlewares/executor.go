package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/errors"
	"github.com/mattiabonardi/endor-sdk-go/models"
)

func Dispatch[T any](payload T, callback models.HandlerFn[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(&payload); err != nil {
			errors.ThrowBadRequest(c, err)
			return
		}
		callback(payload, c)
	}
}

func NoPayloadDispatch(callback models.NoPayloadHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		callback(c)
	}
}
