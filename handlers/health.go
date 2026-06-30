package handlers

import (
	"net/http"
	"time"

	"go-gin-demo/dto"

	"github.com/gin-gonic/gin"
)

// HealthCheck 返回服务器状态。
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(gin.H{"status": "running"}))
}

// ServerTime 返回当前服务器的 Unix 时间戳（以秒为单位）。
func ServerTime(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(gin.H{"server_time": float64(time.Now().UnixMilli()) / 1000.0}))
}
