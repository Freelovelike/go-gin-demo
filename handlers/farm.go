package handlers

import (
	"net/http"

	"go-gin-demo/dto"
	"go-gin-demo/services"

	"github.com/gin-gonic/gin"
)

// SaveFarm 从客户端接收完整的游戏存档。
func SaveFarm(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req dto.SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1001, err.Error()))
		return
	}

	if err := services.SaveFarm(userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Fail(5001, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{"saved": true}))
}

// LoadFarmConfig 响应客户端获取农场配置信息（作物、肥料等）的请求。
func LoadFarmConfig(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(services.LoadFarmConfig()))
}

// LoadFarm 返回经过验证的用户的完整游戏存档。
func LoadFarm(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	resp, err := services.LoadFarm(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Fail(4001, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.OK(resp))
}

// SellCrop 在服务器端出售库存物品。
func SellCrop(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req dto.SellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1001, err.Error()))
		return
	}

	resp, err := services.SellCrop(userID, *req.CropID, req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1002, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.OK(resp))
}
