package upstream

import (
	"context"
	"net/http"
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
	upstreamURL string
	client      *http.Client
}

// NewHTTPClient creates a new HTTP client for upstream requests
func NewHTTPClient(upstreamURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		upstreamURL: upstreamURL,
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Forward sends a request to the upstream server
func (c *HTTPClient) Forward(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Clone the request
	upstreamReq := req.Clone(ctx)

	// Update the URL to point to upstream
	upstreamReq.URL.Scheme = "http" // TODO: Parse from upstreamURL
	upstreamReq.URL.Host = c.upstreamURL
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
