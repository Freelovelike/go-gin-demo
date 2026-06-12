package database

import (
	"context"
	"log"

	"go-gin-demo/config"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

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
