package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config 保存服务器的所有配置参数。
type Config struct {
	Port      string
	DB_DSN    string
	RedisAddr string
	JWTSecret string
}

// Load 读取环境变量并返回配置实例。
func Load() *Config {
	// 如果存在，则加载 .env 文件（如果未找到，则默默忽略）
	_ = godotenv.Load()

	return &Config{
		Port:      getEnv("PORT", "8082"),
		DB_DSN:    getEnv("DB_DSN", "host=localhost user=postgres password=postgres dbname=qqfarm port=5432 sslmode=disable TimeZone=Asia/Shanghai"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
