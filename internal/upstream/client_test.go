package upstream

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPClient_Forward(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotQuery string
	var gotHost string
	var gotBody string
	var gotConnection string
	var gotCustomHeader string

	// Create a test upstream server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotHost = r.Host
		gotBody = string(body)
		gotConnection = r.Header.Get("Connection")
		gotCustomHeader = r.Header.Get("X-Test")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("upstream response"))
	}))
	defer upstream.Close()

	client, err := NewHTTPClient(upstream.URL+"/base", 5*time.Second)
	if err != nil {
		t.Fatalf("NewHTTPClient failed: %v", err)
	}

	req := httptest.NewRequest("POST", "http://proxy.local/path?q=1", strings.NewReader("request body"))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Test", "forwarded")
	ctx := context.Background()

	resp, err := client.Forward(ctx, req)
	if err != nil {
		t.Fatalf("Forward failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("Method = %s, want %s", gotMethod, http.MethodPost)
	}

	if gotPath != "/base/path" {
		t.Errorf("Path = %s, want %s", gotPath, "/base/path")
	}

	if gotQuery != "q=1" {
		t.Errorf("Query = %s, want %s", gotQuery, "q=1")
	}

	if gotHost == "" {
		t.Error("Host should be set on the forwarded request")
	}

	if gotBody != "request body" {
		t.Errorf("Body = %q, want %q", gotBody, "request body")
	}

	if gotConnection != "" {
		t.Errorf("Connection header = %q, want empty", gotConnection)
	}

	if gotCustomHeader != "forwarded" {
		t.Errorf("X-Test = %q, want %q", gotCustomHeader, "forwarded")
	}
}

func TestNewHTTPClient_RequiresFullURL(t *testing.T) {
	_, err := NewHTTPClient("example.com", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for upstream URL without scheme and host")
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
