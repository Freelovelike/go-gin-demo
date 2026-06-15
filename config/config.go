package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DB_DSN    string
	RedisAddr string
	JWTSecret string
}

func Load() *Config {
	// Load .env file if present (silently ignore if not found)
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
