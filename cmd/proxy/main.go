package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/craftbytimi/caching-proxy/internal/cache"
	"github.com/craftbytimi/caching-proxy/internal/config"
	"github.com/craftbytimi/caching-proxy/internal/observability"
	"github.com/craftbytimi/caching-proxy/internal/proxy"
	"github.com/craftbytimi/caching-proxy/internal/server"
	redisstore "github.com/craftbytimi/caching-proxy/internal/store/redis"
	"github.com/craftbytimi/caching-proxy/internal/upstream"
)

func main() {
	cfg, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	logger := observability.NewLogger(cfg.LogLevel)

	// Initialize Redis client
	redisClient := config.NewRedisClient(*cfg)
	cacheStore := redisstore.New(redisClient)
	defer cacheStore.Close()

	// Check Redis connection
	if err := cacheStore.Ping(context.Background()); err != nil {
		log.Fatal("could not connect to redis: ", err)
	}

	upstreamClient, err := upstream.NewHTTPClient(cfg.UpstreamURL, cfg.UpstreamTimeout)
	if err != nil {
		log.Fatal("could not configure upstream client: ", err)
	}
	defer upstreamClient.Close()

	proxyService := proxy.NewService(
		cacheStore,
		upstreamClient,
		cache.NewSimplePolicy(cfg.MaxBodySize),
		cache.NewSimpleKeyGenerator(true),
		cfg.TTL,
		logger,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", server.HealthHandler(cacheStore))
	mux.Handle("/", proxy.NewHandler(proxyService))

	handler := server.Chain(
		mux,
		server.Recovery(logger),
		server.RequestID,
		server.Logging(logger),
	)

	httpServer := server.New(server.Config{
		Port:            cfg.Port,
		Handler:         handler,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		Logger:          logger,
	})

	// Log server and Redis connection info
	logger.Info("proxy configured", "config", cfg.String())

	serverErrCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	go func() {
		serverErrCh <- httpServer.Start()
	}()

	go func() {
		signalCh <- server.WaitForShutdownSignal()
	}()

	select {
	case err := <-serverErrCh:
		if err != nil {
			log.Fatal(err)
		}
		return
	case sig := <-signalCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	if err := httpServer.Shutdown(10 * time.Second); err != nil {
		log.Fatal(err)
	}
}
