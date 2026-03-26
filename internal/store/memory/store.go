package memory

import (
	"context"
	"sync"
	"time"

	"github.com/craftbytimi/caching-proxy/internal/model"
	"github.com/craftbytimi/caching-proxy/internal/store"
)

type entry struct {
	response  *model.CachedResponse
	expiresAt time.Time
}

// Store implements an in-memory cache store
type Store struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

// New creates a new in-memory store
func New() *Store {
	return &Store{
		entries: make(map[string]*entry),
	}
}

// Get retrieves a cached response by key
func (s *Store) Get(ctx context.Context, key string) (*model.CachedResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, exists := s.entries[key]
	if !exists {
		return nil, store.ErrNotFound
	}

	if time.Now().After(e.expiresAt) {
		return nil, store.ErrNotFound
	}

	return e.response, nil
}

// Set stores a response with TTL
func (s *Store) Set(ctx context.Context, key string, response *model.CachedResponse, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[key] = &entry{
		response:  response,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a cached response
func (s *Store) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, key)
	return nil
}

// Clear removes all cached responses
func (s *Store) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = make(map[string]*entry)
	return nil
}

// Ping checks if the store is accessible
func (s *Store) Ping(ctx context.Context) error {
	return nil
}

// Close cleans up store resources
func (s *Store) Close() error {
	return nil
}
