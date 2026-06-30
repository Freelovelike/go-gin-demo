package handlers

import (
	"net/http"

	"go-gin-demo/dto"
	"go-gin-demo/services"

	"github.com/gin-gonic/gin"
)

// HandleAction 是所有农场操作的统一端点。
func HandleAction(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req dto.ActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1001, err.Error()))
		return
	}

	resp, err := services.ExecuteAction(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1002, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.OK(resp))
}
