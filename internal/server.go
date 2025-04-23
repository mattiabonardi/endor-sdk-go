package internal

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Init(microExecutorId string, services []EndorService) {
	// load configuration
	config := LoadConfiguration()
	// create router
	router := gin.New()

	// swagger
	router.StaticFS("/public/", http.Dir("public"))

	// monitoring
	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// api
	api := router.Group("api").Group(":app")
	for _, s := range services {
		resourceGroup := api.Group(s.Resource)
		for methodPath, method := range s.Methods {
			method.Register(resourceGroup, methodPath)
		}
	}

	var routes []string
	for _, routeInfo := range router.Routes() {
		if strings.Contains(routeInfo.Path, "/api/") {
			routes = append(routes, routeInfo.Path)
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

	err := InitServiceDiscovery(microExecutorId, serverAddr, routes)
	if err != nil {
		panic(err)
	}

	// start http server
	router.Run()
}
