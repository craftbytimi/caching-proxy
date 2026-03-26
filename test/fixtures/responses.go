package fixtures

import (
	"net/http"
	"time"

	"github.com/craftbytimi/caching-proxy/internal/model"
)

// SampleCachedResponse returns a sample cached response for testing
func SampleCachedResponse() *model.CachedResponse {
	return &model.CachedResponse{
		StatusCode: http.StatusOK,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body:     []byte(`{"message": "test response"}`),
		CachedAt: time.Now(),
	}
}

// LargeCachedResponse returns a large cached response for testing size limits
func LargeCachedResponse(sizeInMB int) *model.CachedResponse {
	body := make([]byte, sizeInMB*1024*1024)
	for i := range body {
		body[i] = 'A'
	}

	return &model.CachedResponse{
		StatusCode: http.StatusOK,
		Headers: map[string][]string{
			"Content-Type": {"text/plain"},
		},
		Body:     body,
		CachedAt: time.Now(),
	}
}
