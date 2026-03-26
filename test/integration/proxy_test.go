package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/craftbytimi/caching-proxy/internal/cache"
	"github.com/craftbytimi/caching-proxy/internal/observability"
	"github.com/craftbytimi/caching-proxy/internal/proxy"
	"github.com/craftbytimi/caching-proxy/internal/store/memory"
	"github.com/craftbytimi/caching-proxy/internal/upstream"
)

func TestProxyIntegration_CacheHitMiss(t *testing.T) {
	// Create test upstream server
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("upstream response"))
	}))
	defer upstreamServer.Close()

	// Setup proxy components
	store := memory.New()
	upstreamClient := upstream.NewHTTPClient(upstreamServer.URL[7:], 5*time.Second)
	policy := cache.NewSimplePolicy(10 * 1024 * 1024)
	keyGen := cache.NewSimpleKeyGenerator(true)
	logger := observability.NewLogger("info")

	proxyService := proxy.NewService(store, upstreamClient, policy, keyGen, time.Minute, logger)

	// First request - should be MISS
	req1 := httptest.NewRequest("GET", "http://proxy.local/test", nil)
	w1 := httptest.NewRecorder()
	proxyService.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request status = %d, want %d", w1.Code, http.StatusOK)
	}

	if w1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("First request X-Cache = %s, want MISS", w1.Header().Get("X-Cache"))
	}

	// Second request - should be HIT
	req2 := httptest.NewRequest("GET", "http://proxy.local/test", nil)
	w2 := httptest.NewRecorder()
	proxyService.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second request status = %d, want %d", w2.Code, http.StatusOK)
	}

	if w2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("Second request X-Cache = %s, want HIT", w2.Header().Get("X-Cache"))
	}
}

func TestProxyIntegration_NonCacheableMethod(t *testing.T) {
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstreamServer.Close()

	store := memory.New()
	upstreamClient := upstream.NewHTTPClient(upstreamServer.URL[7:], 5*time.Second)
	policy := cache.NewSimplePolicy(10 * 1024 * 1024)
	keyGen := cache.NewSimpleKeyGenerator(true)
	logger := observability.NewLogger("info")

	proxyService := proxy.NewService(store, upstreamClient, policy, keyGen, time.Minute, logger)

	// POST request should bypass cache
	req := httptest.NewRequest("POST", "http://proxy.local/test", nil)
	w := httptest.NewRecorder()
	proxyService.ServeHTTP(w, req)

	if w.Header().Get("X-Cache") != "BYPASS" {
		t.Errorf("POST request X-Cache = %s, want BYPASS", w.Header().Get("X-Cache"))
	}
}

func TestProxyIntegration_TTLExpiration(t *testing.T) {
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))
	defer upstreamServer.Close()

	store := memory.New()
	upstreamClient := upstream.NewHTTPClient(upstreamServer.URL[7:], 5*time.Second)
	policy := cache.NewSimplePolicy(10 * 1024 * 1024)
	keyGen := cache.NewSimpleKeyGenerator(true)
	logger := observability.NewLogger("info")

	// Very short TTL
	proxyService := proxy.NewService(store, upstreamClient, policy, keyGen, 50*time.Millisecond, logger)

	// First request
	req1 := httptest.NewRequest("GET", "http://proxy.local/test", nil)
	w1 := httptest.NewRecorder()
	proxyService.ServeHTTP(w1, req1)

	if w1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("First request should be MISS, got %s", w1.Header().Get("X-Cache"))
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Second request after expiration
	req2 := httptest.NewRequest("GET", "http://proxy.local/test", nil)
	w2 := httptest.NewRecorder()
	proxyService.ServeHTTP(w2, req2)

	if w2.Header().Get("X-Cache") != "MISS" {
		t.Errorf("Request after expiration should be MISS, got %s", w2.Header().Get("X-Cache"))
	}
}
