package upstream

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client forwards requests to upstream servers
type Client interface {
	// Forward sends a request to the upstream and returns the response
	Forward(ctx context.Context, req *http.Request) (*http.Response, error)

	// Close cleans up client resources
	Close() error
}

// HTTPClient implements the Client interface using net/http
type HTTPClient struct {
	upstreamURL *url.URL
	client      *http.Client
}

// NewHTTPClient creates a new HTTP client for upstream requests
func NewHTTPClient(upstreamURL string, timeout time.Duration) (*HTTPClient, error) {
	parsedURL, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("parse upstream URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("upstream URL must include scheme and host")
	}

	return &HTTPClient{
		upstreamURL: parsedURL,
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}, nil
}

// Forward sends a request to the upstream server
func (c *HTTPClient) Forward(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Clone the request
	upstreamReq := req.Clone(ctx)

	// Update the URL to point to upstream while preserving request path and query.
	upstreamURL := *c.upstreamURL
	upstreamURL.Path = joinURLPath(c.upstreamURL.Path, req.URL.Path)
	upstreamURL.RawQuery = req.URL.RawQuery
	upstreamURL.Fragment = ""

	upstreamReq.URL = &upstreamURL
	upstreamReq.Host = c.upstreamURL.Host
	upstreamReq.RequestURI = ""

	// Clean hop-by-hop headers
	CleanRequestHeaders(upstreamReq.Header)

	// Forward the request
	return c.client.Do(upstreamReq)
}

// Close cleans up client resources
func (c *HTTPClient) Close() error {
	c.client.CloseIdleConnections()
	return nil
}

func joinURLPath(basePath, requestPath string) string {
	switch {
	case basePath == "" && requestPath == "":
		return "/"
	case basePath == "" || basePath == "/":
		if requestPath == "" {
			return "/"
		}
		if strings.HasPrefix(requestPath, "/") {
			return requestPath
		}
		return "/" + requestPath
	case requestPath == "" || requestPath == "/":
		return strings.TrimRight(basePath, "/") + "/"
	default:
		return strings.TrimRight(basePath, "/") + "/" + strings.TrimLeft(requestPath, "/")
	}
}
