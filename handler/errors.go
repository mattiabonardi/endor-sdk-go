package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/models"
)

// print to http response internal server error
func ThrowInternalServerError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewDefaultResponseBuilder().AddMessage(models.NewMessage(models.Fatal, err.Error())).Build())
}

// print to http response bad request error
func ThrowBadRequest(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDefaultResponseBuilder().AddMessage(models.NewMessage(models.Fatal, err.Error())).Build())
}

// print to http response not found error
func ThrowNotFound(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusNotFound, models.NewDefaultResponseBuilder().AddMessage(models.NewMessage(models.Fatal, err.Error())).Build())
}

// print to http response unauthorized error
func ThrowUnauthorize(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewDefaultResponseBuilder().AddMessage(models.NewMessage(models.Fatal, err.Error())).Build())
}

// print to http response forbidden error
func ThrowForbidden(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusForbidden, models.NewDefaultResponseBuilder().AddMessage(models.NewMessage(models.Fatal, err.Error())).Build())
}
