package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimplePolicy_ShouldCache(t *testing.T) {
	policy := NewSimplePolicy(10 * 1024 * 1024) // 10MB

	tests := []struct {
		name   string
		method string
		header http.Header
		want   bool
	}{
		{
			name:   "GET request should cache",
			method: "GET",
			header: http.Header{},
			want:   true,
		},
		{
			name:   "POST request should not cache",
			method: "POST",
			header: http.Header{},
			want:   false,
		},
		{
			name:   "GET with Authorization should not cache",
			method: "GET",
			header: http.Header{"Authorization": []string{"Bearer token"}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "http://example.com/path", nil)
			req.Header = tt.header

			got := policy.ShouldCache(req)
			if got != tt.want {
				t.Errorf("ShouldCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimplePolicy_ShouldStore(t *testing.T) {
	policy := NewSimplePolicy(10 * 1024 * 1024) // 10MB

	tests := []struct {
		name       string
		statusCode int
		header     http.Header
		want       bool
	}{
		{
			name:       "200 OK should store",
			statusCode: 200,
			header:     http.Header{},
			want:       true,
		},
		{
			name:       "404 Not Found should not store",
			statusCode: 404,
			header:     http.Header{},
			want:       false,
		},
		{
			name:       "200 with Set-Cookie should not store",
			statusCode: 200,
			header:     http.Header{"Set-Cookie": []string{"session=abc"}},
			want:       false,
		},
		{
			name:       "200 with no-store should not store",
			statusCode: 200,
			header:     http.Header{"Cache-Control": []string{"no-store"}},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/", nil)
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     tt.header,
			}

			got := policy.ShouldStore(req, resp)
			if got != tt.want {
				t.Errorf("ShouldStore() = %v, want %v", got, tt.want)
			}
		})
	}
}
