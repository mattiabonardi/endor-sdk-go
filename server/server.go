package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/configuration"
	"github.com/mattiabonardi/endor-sdk-go/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Init(microExecutorId string, handlers []models.Handler) {
	// load configuration
	config := configuration.Load()
	// create router
	router := gin.New()
	// create handlers
	monitoring := new(MonitoringHandler)

	// swagger
	router.StaticFS("/public/", http.Dir("public"))

	// create resources
	// monitoring
	router.GET("/readyz", monitoring.Status)
	router.GET("/livez", monitoring.Status)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// api
	api := router.Group("api").Group("v1").Group(":app")
	// register endpoints
	for _, h := range handlers {
		h.Route(api)
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

	err := configuration.InitServiceDiscovery(microExecutorId, serverAddr, routes)
	if err != nil {
		panic(err)
	}

	// start http server
	router.Run()
}
