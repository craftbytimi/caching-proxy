package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/craftbytimi/caching-proxy/internal/config"
)

func main() {
	cfg := config.Load()

	// Initialize Redis client
	redisClient := config.NewRedisClient(cfg)
	defer redisClient.Close()

	// Check Redis connection
	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal("could not connect to redis: ", err)
	}

	// Basic health check endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "caching-proxy online")
	})

	// Health check for Redis connectivity
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		err := redisClient.Ping(r.Context()).Err()
		if err != nil {
			http.Error(w, "redis is down", http.StatusServiceUnavailable)
			return
		}

		fmt.Fprintln(w, "ok")
	})

	// Log server and Redis connection info
	log.Println("server running on port " + cfg.Port)
	log.Println("redis connected at " + cfg.RedisAddr)

	err = http.ListenAndServe(":"+cfg.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
