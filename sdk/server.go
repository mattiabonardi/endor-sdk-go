package sdk

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Init(microserviceId string, services []EndorService) {
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
		services = append(services, NewResourceService(microserviceId, services, client, ctx, microserviceId))
	}

	registry := GetInternalServiceRegistry()
	registry.Init(client, ctx, microserviceId)
	internalServices := registry.GetInternalServices()
	apiBasePath := "/api"
	for _, service := range services {
		registry.RegisterService(&internalServices, service, microserviceId)
	}

	// load dynamic services
	serviceMap := registry.GetServices()
	for _, service := range serviceMap {
		services = append(services, service.instance)
	}

	router.NoRoute(func(c *gin.Context) {
		def := registry.GetService(c.Request.URL.Path)
		if def != nil {
			def.callback(c)
		} else {
			response := NewDefaultResponseBuilder()
			response.AddMessage(NewMessage(Fatal, "404 page not found (uri: "+c.Request.RequestURI+", method: "+c.Request.Method+")"))
			c.JSON(http.StatusNotFound, response.Build())
		}
	})

	err := InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), services)
	if err != nil {
		log.Fatal(err)
	}
	swaggerPath, err := CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), services, apiBasePath)
	if err != nil {
		log.Fatal(err)
	}

	// swagger
	router.StaticFS("/swagger", http.Dir(swaggerPath))

	// start http server
	router.Run()
}
