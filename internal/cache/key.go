package cache

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// KeyGenerator generates cache keys from HTTP requests
type KeyGenerator interface {
	Generate(method, urlStr string, headers map[string][]string) string
}

// SimpleKeyGenerator creates cache keys from method + URL
type SimpleKeyGenerator struct {
	includeQuery bool
}

// NewSimpleKeyGenerator creates a new simple key generator
func NewSimpleKeyGenerator(includeQuery bool) *SimpleKeyGenerator {
	return &SimpleKeyGenerator{
		includeQuery: includeQuery,
	}
}

// Generate creates a cache key from the request
func (g *SimpleKeyGenerator) Generate(method, urlStr string, headers map[string][]string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// Fallback to simple hash if URL parsing fails
		return hashString(method + ":" + urlStr)
	}

	// Normalize the URL
	normalized := normalizeURL(parsedURL, g.includeQuery)

	// Create key: method:normalized_url
	key := method + ":" + normalized

	return hashString(key)
}

// normalizeURL normalizes a URL for consistent cache key generation
func normalizeURL(u *url.URL, includeQuery bool) string {
	var sb strings.Builder

	// Scheme (lowercase)
	sb.WriteString(strings.ToLower(u.Scheme))
	sb.WriteString("://")

	// Host (lowercase)
	sb.WriteString(strings.ToLower(u.Host))

	// Path
	path := u.Path
	if path == "" {
		path = "/"
	}
	sb.WriteString(path)

	// Query string (sorted for consistency)
	if includeQuery && u.RawQuery != "" {
		sb.WriteString("?")
		sb.WriteString(sortQueryString(u.RawQuery))
	}

	return sb.String()
}

// sortQueryString sorts query parameters for consistent keys
func sortQueryString(query string) string {
	values, err := url.ParseQuery(query)
	if err != nil {
		return query
	}

	// Sort keys
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build sorted query string
	var parts []string
	for _, k := range keys {
		for _, v := range values[k] {
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}

	return strings.Join(parts, "&")
}

// hashString creates a SHA-256 hash of the input
func hashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
