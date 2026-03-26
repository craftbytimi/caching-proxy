package config

import (
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port            string
	UpstreamURL     string
	TTL             time.Duration
	RedisAddr       string
	RedisDB         int
	RedisPassword   string
	MaxBodySize     int64
	UpstreamTimeout time.Duration
	LogLevel        string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	upstreamURL := os.Getenv("UPSTREAM_URL")

	ttlStr := os.Getenv("CACHE_TTL")
	ttl := 60 * time.Second
	if ttlStr != "" {
		if val, err := strconv.Atoi(ttlStr); err == nil {
			ttl = time.Duration(val) * time.Second
		}
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisDBStr := os.Getenv("REDIS_DB")
	redisDB := 0
	if redisDBStr != "" {
		if val, err := strconv.Atoi(redisDBStr); err == nil {
			redisDB = val
		}
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	maxBodyStr := os.Getenv("MAX_BODY_SIZE")
	maxBodySize := int64(10 * 1024 * 1024) // 10MB default
	if maxBodyStr != "" {
		if val, err := strconv.ParseInt(maxBodyStr, 10, 64); err == nil {
			maxBodySize = val * 1024 * 1024
		}
	}

	timeoutStr := os.Getenv("UPSTREAM_TIMEOUT")
	timeout := 10 * time.Second
	if timeoutStr != "" {
		if val, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(val) * time.Second
		}
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return Config{
		Port:            port,
		UpstreamURL:     upstreamURL,
		TTL:             ttl,
		RedisAddr:       redisAddr,
		RedisDB:         redisDB,
		RedisPassword:   redisPassword,
		MaxBodySize:     maxBodySize,
		UpstreamTimeout: timeout,
		LogLevel:        logLevel,
	}
}

func NewRedisClient(cfg Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		DB:       cfg.RedisDB,
		Password: cfg.RedisPassword,
	})
}
