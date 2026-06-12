package main

import (
	"log"

	"go-gin-demo/config"
	"go-gin-demo/database"
	"go-gin-demo/middleware"
	"go-gin-demo/routes"
	"go-gin-demo/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Database
	database.InitPostgres(cfg)
	database.AutoMigrate()
	database.SeedCropDefs()

	// Redis
	database.InitRedis(cfg)

	// Services
	services.InitAuth(cfg)

	// Router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	routes.Register(r)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
