package routes

import (
	"go-gin-demo/handlers"
	"go-gin-demo/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	// Health check (no auth)
	r.GET("/health", handlers.HealthCheck)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Auth routes — no JWT required
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// Protected routes — JWT required
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		{
			protected.GET("/profile", handlers.GetProfile)

			// Farm routes
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
