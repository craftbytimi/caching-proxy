package config

import (
	"flag"
	"fmt"
	"net/url"
	"time"
)

// ParseFlags parses command-line flags and merges with config
func ParseFlags() (*Config, error) {
	var (
		port            = flag.String("port", "", "HTTP server port")
		upstreamURL     = flag.String("upstream", "", "Upstream server URL")
		ttl             = flag.Int("ttl", 0, "Cache TTL in seconds")
		redisAddr       = flag.String("redis", "", "Redis address")
		redisDB         = flag.Int("redis-db", 0, "Redis database number")
		redisPassword   = flag.String("redis-password", "", "Redis password")
		maxBodySize     = flag.Int("max-body", 0, "Max cacheable body size in MB")
		upstreamTimeout = flag.Int("timeout", 0, "Upstream timeout in seconds")
		logLevel        = flag.String("log-level", "", "Log level (debug|info|warn|error)")
	)

	flag.Parse()

	// Start with environment-based config
	cfg := Load()

	// Override with flags if provided
	if *port != "" {
		cfg.Port = *port
	}
	if *upstreamURL != "" {
		cfg.UpstreamURL = *upstreamURL
	}
	if *ttl > 0 {
		cfg.TTL = time.Duration(*ttl) * time.Second
	}
	if *redisAddr != "" {
		cfg.RedisAddr = *redisAddr
	}
	if *redisDB > 0 {
		cfg.RedisDB = *redisDB
	}
	if *redisPassword != "" {
		cfg.RedisPassword = *redisPassword
	}
	if *maxBodySize > 0 {
		cfg.MaxBodySize = int64(*maxBodySize) * 1024 * 1024 // Convert MB to bytes
	}
	if *upstreamTimeout > 0 {
		cfg.UpstreamTimeout = time.Duration(*upstreamTimeout) * time.Second
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("port is required")
	}

	if c.UpstreamURL == "" {
		return fmt.Errorf("upstream URL is required")
	}

	// Validate upstream URL format
	parsedURL, err := url.Parse(c.UpstreamURL)
	if err != nil {
		return fmt.Errorf("invalid upstream URL: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("upstream URL must include scheme and host")
	}

	if c.TTL <= 0 {
		return fmt.Errorf("TTL must be positive")
	}

	if c.MaxBodySize <= 0 {
		return fmt.Errorf("max body size must be positive")
	}

	return nil
}

// String returns a string representation of the config
func (c *Config) String() string {
	return fmt.Sprintf(
		"port=%s upstream=%s ttl=%s redis=%s max_body=%dMB timeout=%s log_level=%s",
		c.Port,
		c.UpstreamURL,
		c.TTL,
		c.RedisAddr,
		c.MaxBodySize/(1024*1024),
		c.UpstreamTimeout,
		c.LogLevel,
	)
}
