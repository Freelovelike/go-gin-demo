package handlers

import (
	"net/http"
	"time"

	"go-gin-demo/dto"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns server status.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(gin.H{"status": "running"}))
}

func ServerTime(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(gin.H{"server_time": float64(time.Now().UnixMilli()) / 1000.0}))
}
