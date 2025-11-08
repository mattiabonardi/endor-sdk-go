package sdk

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Init(microserviceId string, internalEndorServices *[]EndorService) {
	// load configuration
	config := GetConfig()

	// define runtime configuration
	config.EndorDynamicResourceDBName = microserviceId

	// create router
	router := gin.New()

	// monitoring
	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if config.EndorResourceServiceEnabled {
		*internalEndorServices = append(*internalEndorServices, *NewResourceService(microserviceId, internalEndorServices, microserviceId))
		*internalEndorServices = append(*internalEndorServices, *NewResourceActionService(microserviceId, internalEndorServices, microserviceId))
	}

	// get all resources
	EndorServiceRepository := NewEndorServiceRepository(microserviceId, internalEndorServices, microserviceId)
	resources, err := EndorServiceRepository.EndorServiceList()
	if err != nil {
		log.Fatal(err)
	}

	router.NoRoute(func(c *gin.Context) {
		// find the resource in path /api/{version}/{resource}/{method}
		pathSegments := strings.Split(c.Request.URL.Path, "/")
		if len(pathSegments) > 4 {
			resource := pathSegments[3]
			action := pathSegments[4]
			endorRepositoryDictionary, err := EndorServiceRepository.Instance(ReadInstanceDTO{
				Id: resource,
			})
			if err == nil {
				if method, ok := endorRepositoryDictionary.EndorService.Methods[action]; ok {
					method.CreateHTTPCallback(microserviceId)(c)
					return
				}
			}
		}
		response := NewDefaultResponseBuilder()
		response.AddMessage(NewMessage(Fatal, "404 page not found (uri: "+c.Request.RequestURI+", method: "+c.Request.Method+")"))
		c.JSON(http.StatusNotFound, response.Build())
	})

	err = InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), resources)
	if err != nil {
		log.Fatal(err)
	}
	swaggerPath, err := CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), resources, "/api")
	if err != nil {
		log.Fatal(err)
	}

	// swagger
	router.StaticFS("/swagger", http.Dir(swaggerPath))

	// start http server
	router.Run()
}
