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

func Init(microExecutorId string, services []EndorService) {
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

	if config.EndorResourceServiceEnabled {
		// Connect to MongoDB
		ctx := context.TODO()
		clientOptions := options.Client().ApplyURI(config.EndorResourceDBUri)
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatal("MongoDB connection error:", err)
		}

		// Ping to test connection
		err = client.Ping(ctx, nil)
		if err != nil {
			log.Fatal("MongoDB ping failed:", err)
		}
		services = append(services, NewResourceService(microExecutorId, services, client, ctx, microExecutorId))
	}

	// api
	api := router.Group("api")
	for _, s := range services {
		var versionGroup *gin.RouterGroup
		if s.Version != "" {
			versionGroup = api.Group(s.Version)
		} else {
			versionGroup = api.Group("v1")
		}
		resourceGroup := versionGroup.Group(s.Resource)
		for methodPath, method := range s.Methods {
			method.Register(resourceGroup, methodPath)
		}
	}

	router.NoRoute(func(c *gin.Context) {
		response := NewDefaultResponseBuilder()
		response.AddMessage(NewMessage(Fatal, "404 page not found (uri: "+c.Request.RequestURI+", method: "+c.Request.Method+")"))
		c.JSON(http.StatusNotFound, response.Build())
	})

	// service discovery configuration
	var serverAddr string
	if config.ServerDNS != "" {
		serverAddr = config.ServerDNS // Dereference pointer
	} else {
		serverAddr = fmt.Sprintf("http://localhost:%s", config.ServerPort)
	}

	err := InitializeApiGatewayConfiguration(microExecutorId, serverAddr, services)
	if err != nil {
		log.Fatal(err)
	}
	swaggerPath, err := CreateSwaggerConfiguration(microExecutorId, serverAddr, services, api.BasePath())
	if err != nil {
		log.Fatal(err)
	}

	// swagger
	router.StaticFS("/swagger", http.Dir(swaggerPath))

	// start http server
	router.Run()
}
