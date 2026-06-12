package handlers

import (
	"net/http"

	"go-gin-demo/dto"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns server status.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(gin.H{"status": "running"}))
}
