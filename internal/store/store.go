package store

import (
	"context"
	"errors"
	"time"

	"github.com/craftbytimi/caching-proxy/internal/model"
)

var (
	// ErrNotFound is returned when a key is not found in the store
	ErrNotFound = errors.New("key not found in cache")

	// ErrStoreFull is returned when the store has reached capacity
	ErrStoreFull = errors.New("cache store is full")
)

// Store defines the cache storage interface
type Store interface {
	// Get retrieves a cached response by key
	Get(ctx context.Context, key string) (*model.CachedResponse, error)

	// Set stores a response with TTL
	Set(ctx context.Context, key string, response *model.CachedResponse, ttl time.Duration) error

	// Delete removes a cached response
	Delete(ctx context.Context, key string) error

	// Clear removes all cached responses
	Clear(ctx context.Context) error

	// Ping checks if the store is accessible
	Ping(ctx context.Context) error

	// Close cleans up store resources
	Close() error
}
