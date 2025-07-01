package sdk

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Init(microserviceId string, internalEndorServices *[]EndorService) {
	// load configuration
	config := LoadConfiguration()

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

	var client *mongo.Client
	ctx := context.TODO()

	if config.EndorResourceServiceEnabled {
		if config.EndorDynamicResourcesEnabled {
			// Connect to MongoDB
			clientOptions := options.Client().ApplyURI(config.EndorServiceDBUri)
			var err error
			client, err = mongo.Connect(ctx, clientOptions)
			if err != nil {
				log.Fatal("MongoDB connection error:", err)
			}

			// Ping to test connection
			err = client.Ping(ctx, nil)
			if err != nil {
				log.Fatal("MongoDB ping failed:", err)
			}
		}
		*internalEndorServices = append(*internalEndorServices, *NewResourceService(microserviceId, internalEndorServices, client, ctx, microserviceId))
		*internalEndorServices = append(*internalEndorServices, *NewResourceActionService(microserviceId, internalEndorServices, client, ctx, microserviceId))
	}

	// get all resources
	EndorServiceRepository := NewEndorServiceRepository(microserviceId, internalEndorServices, client, ctx, microserviceId)
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
