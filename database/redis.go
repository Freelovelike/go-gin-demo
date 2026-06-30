package database

import (
	"context"
	"log"

	"go-gin-demo/config"

	"github.com/redis/go-redis/v9"
)

// Redis 是全局的 Redis 客户端实例。
var Redis *redis.Client

// InitRedis 初始化并连接 Redis 服务器。
func InitRedis(cfg *config.Config) {
	Redis = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "",
		DB:       0,
	})

	if err := Redis.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	log.Println("Redis connected")
}
