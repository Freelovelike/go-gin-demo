package routes

import (
	"go-gin-demo/handlers"
	"go-gin-demo/middleware"

	"github.com/gin-gonic/gin"
)

// Register 配置并注册所有 API 路由。
func Register(r *gin.Engine) {
	// 健康检查（无验证）
	r.GET("/health", handlers.HealthCheck)
	r.GET("/api/v1/time", handlers.ServerTime)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 验证路由——不需要 JWT
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// 受保护的路由——需要 JWT
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		{
			protected.GET("/profile", handlers.GetProfile)

			// 农场路由
			farm := protected.Group("/farm")
			{
				farm.GET("/config", handlers.LoadFarmConfig)
				farm.POST("/save", handlers.SaveFarm)
				farm.GET("/load", handlers.LoadFarm)
				farm.POST("/sell", handlers.SellCrop)
				farm.POST("/action", handlers.HandleAction)
			}
		}
	}
}
