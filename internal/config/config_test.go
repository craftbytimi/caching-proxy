package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("REDIS_ADDR", "")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "8080")
	}

	if cfg.RedisAddr != "localhost:6379" {
		t.Fatalf("RedisAddr = %q, want %q", cfg.RedisAddr, "localhost:6379")
	}
}

func TestLoadUsesEnvValues(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("REDIS_ADDR", "127.0.0.1:6380")

	cfg := Load()

	if cfg.Port != "3000" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "3000")
	}

	if cfg.RedisAddr != "127.0.0.1:6380" {
		t.Fatalf("RedisAddr = %q, want %q", cfg.RedisAddr, "127.0.0.1:6380")
	}
}

func TestNewRedisClientUsesConfigAddress(t *testing.T) {
	cfg := Config{
		Port:      "8080",
		RedisAddr: "cache:6379",
	}

	client := NewRedisClient(cfg)

	if client.Options().Addr != "cache:6379" {
		t.Fatalf("Addr = %q, want %q", client.Options().Addr, "cache:6379")
	}
}
