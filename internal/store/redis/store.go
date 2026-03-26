package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/craftbytimi/caching-proxy/internal/model"
	"github.com/craftbytimi/caching-proxy/internal/store"
	"github.com/redis/go-redis/v9"
)

// Store implements the store.Store interface using Redis
type Store struct {
	client *redis.Client
}

// New creates a new Redis store
func New(client *redis.Client) *Store {
	return &Store{
		client: client,
	}
}

// Get retrieves a cached response by key
func (s *Store) Get(ctx context.Context, key string) (*model.CachedResponse, error) {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	var response model.CachedResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Set stores a response with TTL
func (s *Store) Set(ctx context.Context, key string, response *model.CachedResponse, ttl time.Duration) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a cached response
func (s *Store) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// Clear removes all cached responses
func (s *Store) Clear(ctx context.Context) error {
	return s.client.FlushDB(ctx).Err()
}

// Ping checks if the store is accessible
func (s *Store) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Close cleans up store resources
func (s *Store) Close() error {
	return s.client.Close()
}
