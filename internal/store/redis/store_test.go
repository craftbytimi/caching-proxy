package redis

import "testing"

func TestStore_GetSetDelete(t *testing.T) {
	// TODO: Implement test with miniredis or real Redis
	t.Skip("Redis store tests not implemented yet")
}

func TestStore_TTLExpiration(t *testing.T) {
	// TODO: Test that entries expire after TTL
	t.Skip("TTL expiration test not implemented yet")
}

func TestStore_NotFound(t *testing.T) {
	// TODO: Test that Get returns ErrNotFound for missing keys
	t.Skip("Not found test not implemented yet")
}
