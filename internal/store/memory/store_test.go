package memory

import (
	"context"
	"testing"
	"time"

	"github.com/craftbytimi/caching-proxy/internal/model"
	"github.com/craftbytimi/caching-proxy/internal/store"
)

func TestMemoryStore_GetSetDelete(t *testing.T) {
	s := New()
	ctx := context.Background()

	key := "test-key"
	response := &model.CachedResponse{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/plain"}},
		Body:       []byte("test body"),
		CachedAt:   time.Now(),
	}

	// Test Set
	err := s.Set(ctx, key, response, time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get
	retrieved, err := s.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.StatusCode != response.StatusCode {
		t.Errorf("StatusCode = %d, want %d", retrieved.StatusCode, response.StatusCode)
	}

	// Test Delete
	err = s.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = s.Get(ctx, key)
	if err != store.ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryStore_TTLExpiration(t *testing.T) {
	s := New()
	ctx := context.Background()

	key := "test-key"
	response := &model.CachedResponse{
		StatusCode: 200,
		Body:       []byte("test"),
	}

	// Set with very short TTL
	err := s.Set(ctx, key, response, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should not find expired entry
	_, err = s.Get(ctx, key)
	if err != store.ErrNotFound {
		t.Errorf("Expected ErrNotFound for expired entry, got %v", err)
	}
}
