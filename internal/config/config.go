package config

import (
	"os"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port      string
	RedisAddr string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	return Config{
		Port:      port,
		RedisAddr: redisAddr,
	}
}

func NewRedisClient(cfg Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
}
