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

	// 数据库
	database.InitPostgres(cfg)
	database.AutoMigrate()
	database.SeedCropDefs()

	// Redis
	database.InitRedis(cfg)

	// 服务
	services.InitAuth(cfg)

	// 路由器
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	routes.Register(r)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
