package handlers

import (
	"net/http"

	"go-gin-demo/dto"
	"go-gin-demo/services"

	"github.com/gin-gonic/gin"
)

// Register 处理用户注册。
func Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1001, err.Error()))
		return
	}

	resp, err := services.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1002, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.OK(resp))
}

// Login 处理用户登录。
func Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Fail(1001, err.Error()))
		return
	}

	resp, err := services.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.Fail(2001, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.OK(resp))
}

// GetProfile 返回当前用户的信息（需要 JWT）。
func GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	user, err := services.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Fail(4001, "user not found"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(user))
}
