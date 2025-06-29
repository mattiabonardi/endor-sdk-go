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

func Init(microserviceId string, internalEndorResources *[]EndorResource) {
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
		// Connect to MongoDB
		clientOptions := options.Client().ApplyURI(config.EndorResourceDBUri)
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
		*internalEndorResources = append(*internalEndorResources, *NewResourceService(microserviceId, internalEndorResources, client, ctx, microserviceId))
		*internalEndorResources = append(*internalEndorResources, *NewResourceActionService(microserviceId, internalEndorResources, client, ctx, microserviceId))
	}

	// get all resources
	endorResourceRepository := NewEndorResourceRepository(microserviceId, internalEndorResources, client, ctx, microserviceId)
	resources, err := endorResourceRepository.EndorResourceList()
	if err != nil {
		log.Fatal(err)
	}

	router.NoRoute(func(c *gin.Context) {
		// find the resource in path /api/{version}/{resource}/{method}
		pathSegments := strings.Split(c.Request.URL.Path, "/")
		if len(pathSegments) > 4 {
			resource := pathSegments[3]
			action := pathSegments[4]
			endorRepositoryDictionary, err := endorResourceRepository.Instance(ReadInstanceDTO{
				Id: resource,
			})
			if err == nil {
				if method, ok := endorRepositoryDictionary.endorResource.Methods[action]; ok {
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
