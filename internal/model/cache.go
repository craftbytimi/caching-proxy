package model

// CacheStatus indicates whether a request was served from cache
type CacheStatus string

const (
	// CacheHit indicates the response was served from cache
	CacheHit CacheStatus = "HIT"

	// CacheMiss indicates the response was fetched from upstream
	CacheMiss CacheStatus = "MISS"

	// CacheBypass indicates the request bypassed the cache
	CacheBypass CacheStatus = "BYPASS"
)
