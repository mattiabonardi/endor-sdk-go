package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/models"
)

func ResourceHandler(baseRoute *gin.RouterGroup, services []models.EndorService) {
	// register endpoints
	for _, s := range services {
		resourceGroup := baseRoute.Group(s.Resource)
		for key, method := range s.Methods {
			resourceGroup.POST(key, handle(s.Resource, key, method))
		}
	}
}

func handle(resource string, methodKey string, method models.EndorServiceMethodHandler[any, any]) gin.HandlerFunc {
	return func(c *gin.Context) {
		context := models.EndorServiceContext[any]{}

		// validate payload
		if method.Payload != nil {
			payload := method.Payload // Assuming method.Payload is a pointer to a new struct
			if err := c.ShouldBindJSON(payload); err != nil {
				ThrowBadRequest(c, err)
				return
			}
			context.Payload = payload // Assuming context has a Payload field
		}

		// authorize request
		if !method.Public {
			session, err := AuthorizeResource(c, resource, methodKey)
			if err != nil {
				ThrowUnauthorize(c, err)
				return
			}
			context.Session = session
		}

		// Here you might want to invoke method.Handler
		response, err := method.HandlerFunc(context)
		if err != nil {
			ThrowInternalServerError(c, err)
		}
		c.JSON(http.StatusOK, response)
	}
}
