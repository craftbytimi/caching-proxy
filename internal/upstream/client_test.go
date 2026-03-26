package upstream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClient_Forward(t *testing.T) {
	// Create a test upstream server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("upstream response"))
	}))
	defer upstream.Close()

	client := NewHTTPClient(upstream.URL[7:], 5*time.Second) // Remove http://

	req := httptest.NewRequest("GET", "http://proxy.local/path", nil)
	ctx := context.Background()

	resp, err := client.Forward(ctx, req)
	if err != nil {
		t.Fatalf("Forward failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCleanRequestHeaders(t *testing.T) {
	header := http.Header{
		"Content-Type": []string{"application/json"},
		"Connection":   []string{"keep-alive"},
		"User-Agent":   []string{"test"},
	}

	CleanRequestHeaders(header)

	if header.Get("Connection") != "" {
		t.Error("Connection header should be removed")
	}

	if header.Get("Content-Type") == "" {
		t.Error("Content-Type header should be preserved")
	}
}
