package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MonitoringHandler struct{}

func (h MonitoringHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
