package cache

import (
	"net/http"
	"strings"
)

// Policy determines cacheability of requests and responses
type Policy interface {
	// ShouldCache determines if a request should check the cache
	ShouldCache(req *http.Request) bool

	// ShouldStore determines if a response should be stored
	ShouldStore(req *http.Request, resp *http.Response) bool
}

// SimplePolicy implements basic caching rules
type SimplePolicy struct {
	maxBodySize int64 // Maximum body size to cache in bytes
}

// NewSimplePolicy creates a new simple policy
func NewSimplePolicy(maxBodySize int64) *SimplePolicy {
	return &SimplePolicy{
		maxBodySize: maxBodySize,
	}
}

// ShouldCache determines if a request should check the cache
func (p *SimplePolicy) ShouldCache(req *http.Request) bool {
	// Only cache GET requests
	if req.Method != http.MethodGet {
		return false
	}

	// Don't cache requests with Authorization header
	if req.Header.Get("Authorization") != "" {
		return false
	}

	// Don't cache requests with sensitive cookies
	if hasSensitiveCookies(req) {
		return false
	}

	return true
}

// ShouldStore determines if a response should be stored
func (p *SimplePolicy) ShouldStore(req *http.Request, resp *http.Response) bool {
	// Only store 200 OK responses
	if resp.StatusCode != http.StatusOK {
		return false
	}

	// Check Content-Length if present
	if resp.ContentLength > 0 && resp.ContentLength > p.maxBodySize {
		return false
	}

	// Don't cache responses with Set-Cookie header
	if resp.Header.Get("Set-Cookie") != "" {
		return false
	}

	// Respect Cache-Control: no-store
	cacheControl := resp.Header.Get("Cache-Control")
	if strings.Contains(strings.ToLower(cacheControl), "no-store") {
		return false
	}

	return true
}

// hasSensitiveCookies checks if the request has sensitive cookies
func hasSensitiveCookies(req *http.Request) bool {
	cookies := req.Cookies()
	for _, cookie := range cookies {
		name := strings.ToLower(cookie.Name)
		// Skip common session cookies
		if strings.Contains(name, "session") ||
			strings.Contains(name, "auth") ||
			strings.Contains(name, "token") {
			return true
		}
	}
	return false
}
