package handlers

import (
	"net/http"

	"go-gin-demo/dto"
	"go-gin-demo/services"

	"github.com/gin-gonic/gin"
)

// SaveFarm receives the full game save from the client.
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

func LoadFarmConfig(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(services.LoadFarmConfig()))
}

// LoadFarm returns the full game save for the authenticated user.
func LoadFarm(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	resp, err := services.LoadFarm(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Fail(4001, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.OK(resp))
}

// SellCrop sells inventory items server-side.
func SellCrop(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req dto.SellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1001, err.Error()))
		return
	}

	resp, err := services.SellCrop(userID, req.CropID, req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1002, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.OK(resp))
}
