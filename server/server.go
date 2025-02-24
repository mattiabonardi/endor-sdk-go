package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Init(handlers []models.Handler) {
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
}
