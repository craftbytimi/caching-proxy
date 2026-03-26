package proxy

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/craftbytimi/caching-proxy/internal/cache"
	"github.com/craftbytimi/caching-proxy/internal/model"
	"github.com/craftbytimi/caching-proxy/internal/store"
	"github.com/craftbytimi/caching-proxy/internal/upstream"
)

// Service orchestrates the caching proxy logic
type Service struct {
	store    store.Store
	upstream upstream.Client
	policy   cache.Policy
	keyGen   cache.KeyGenerator
	ttl      time.Duration
	logger   *slog.Logger
}

// NewService creates a new proxy service
func NewService(
	store store.Store,
	upstream upstream.Client,
	policy cache.Policy,
	keyGen cache.KeyGenerator,
	ttl time.Duration,
	logger *slog.Logger,
) *Service {
	return &Service{
		store:    store,
		upstream: upstream,
		policy:   policy,
		keyGen:   keyGen,
		ttl:      ttl,
		logger:   logger,
	}
}

// ServeHTTP handles incoming proxy requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if request should use cache
	if !s.policy.ShouldCache(r) {
		s.logger.Debug("bypassing cache",
			"method", r.Method,
			"path", r.URL.Path,
		)
		s.forwardRequest(ctx, w, r, model.CacheBypass)
		return
	}

	// Generate cache key
	cacheKey := s.keyGen.Generate(r.Method, r.URL.String(), r.Header)

	// Try to get from cache
	cached, err := s.store.Get(ctx, cacheKey)
	if err == nil {
		s.logger.Info("cache hit",
			"method", r.Method,
			"path", r.URL.Path,
			"key", cacheKey,
		)
		s.serveCachedResponse(w, cached, model.CacheHit)
		return
	}

	// Cache miss - forward to upstream
	if err != store.ErrNotFound {
		s.logger.Error("cache read error",
			"error", err,
			"key", cacheKey,
		)
	}

	s.logger.Info("cache miss",
		"method", r.Method,
		"path", r.URL.Path,
		"key", cacheKey,
	)

	s.forwardAndCache(ctx, w, r, cacheKey, model.CacheMiss)
}

// forwardRequest forwards the request without caching
func (s *Service) forwardRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, status model.CacheStatus) {
	resp, err := s.upstream.Forward(ctx, r)
	if err != nil {
		s.logger.Error("upstream request failed", "error", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response
	w.Header().Set("X-Cache", string(status))
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// forwardAndCache forwards request and caches the response if appropriate
func (s *Service) forwardAndCache(ctx context.Context, w http.ResponseWriter, r *http.Request, cacheKey string, status model.CacheStatus) {
	resp, err := s.upstream.Forward(ctx, r)
	if err != nil {
		s.logger.Error("upstream request failed", "error", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed to read response body", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Check if response should be cached
	if s.policy.ShouldStore(r, resp) {
		cached := &model.CachedResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       body,
			CachedAt:   time.Now(),
		}

		if err := s.store.Set(ctx, cacheKey, cached, s.ttl); err != nil {
			s.logger.Error("failed to cache response", "error", err, "key", cacheKey)
		} else {
			s.logger.Debug("response cached", "key", cacheKey, "ttl", s.ttl)
		}
	}

	// Send response to client
	w.Header().Set("X-Cache", string(status))
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// serveCachedResponse writes a cached response to the client
func (s *Service) serveCachedResponse(w http.ResponseWriter, cached *model.CachedResponse, status model.CacheStatus) {
	w.Header().Set("X-Cache", string(status))
	copyHeaders(w.Header(), cached.Headers)
	w.WriteHeader(cached.StatusCode)
	w.Write(cached.Body)
}

// copyHeaders copies headers from src to dst
func copyHeaders(dst, src http.Header) {
	for k, values := range src {
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}
