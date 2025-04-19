package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/configuration"
	"github.com/mattiabonardi/endor-sdk-go/handler"
	"github.com/mattiabonardi/endor-sdk-go/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Init(microExecutorId string, services []models.EndorService) {
	// load configuration
	config := configuration.LoadConfiguration()
	// create router
	router := gin.New()
	// create handlers
	monitoring := new(handler.MonitoringHandler)

	// swagger
	router.StaticFS("/public/", http.Dir("public"))

	// monitoring
	router.GET("/readyz", monitoring.Status)
	router.GET("/livez", monitoring.Status)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// api
	api := router.Group("api").Group(":app")
	// register endpoints
	for _, s := range services {
		resourceHandler := api.Group(s.Resource)
		for key, method := range s.Methods {
			resourceHandler.POST(key, handle(s.Resource, key, method))
		}
	}

	var routes []string
	for _, routeInfo := range router.Routes() {
		if strings.Contains(routeInfo.Path, "/api/") {
			routes = append(routes, routeInfo.Path)
		}
	}

	router.NoRoute(func(c *gin.Context) {
		response := models.NewDefaultResponseBuilder()
		response.AddMessage(models.NewMessage(models.Fatal, "404 page not found (uri: "+c.Request.RequestURI+", method: "+c.Request.Method+")"))
		c.JSON(http.StatusNotFound, response.Build())
	})

	// service discovery configuration
	var serverAddr string
	if config.ServerDNS != "" {
		serverAddr = config.ServerDNS // Dereference pointer
	} else {
		serverAddr = fmt.Sprintf("http://localhost:%s", config.ServerPort)
	}

	err := handler.InitServiceDiscovery(microExecutorId, serverAddr, routes)
	if err != nil {
		panic(err)
	}

	// start http server
	router.Run()
}

func handle(resource string, methodKey string, method models.EndorServiceMethodHandler[any, any]) gin.HandlerFunc {
	return func(c *gin.Context) {
		context := models.EndorServiceContext[any]{}

		// validate payload
		if method.Payload != nil {
			payload := method.Payload // Assuming method.Payload is a pointer to a new struct
			if err := c.ShouldBindJSON(payload); err != nil {
				handler.ThrowBadRequest(c, err)
				return
			}
			context.Payload = payload // Assuming context has a Payload field
		}

		// authorize request
		if !method.Public {
			session, err := handler.AuthorizeResource(c, resource, methodKey)
			if err != nil {
				handler.ThrowUnauthorize(c, err)
				return
			}
			context.Session = session
		}

		// Here you might want to invoke method.Handler
		response, err := method.HandlerFunc(context)
		if err != nil {
			handler.ThrowInternalServerError(c, err)
		}
		c.JSON(http.StatusOK, response)
	}
}
