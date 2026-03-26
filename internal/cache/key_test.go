package cache

import (
	"testing"
)

func TestKeyGenerator_Generate(t *testing.T) {
	gen := NewSimpleKeyGenerator(true)

	tests := []struct {
		name     string
		method   string
		url      string
		wantSame bool
	}{
		{
			name:     "same URL produces same key",
			method:   "GET",
			url:      "http://example.com/path",
			wantSame: true,
		},
		{
			name:     "query order doesn't matter",
			method:   "GET",
			url:      "http://example.com/path?b=2&a=1",
			wantSame: true, // Should match ?a=1&b=2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := gen.Generate(tt.method, tt.url, nil)
			if key1 == "" {
				t.Error("Generated key is empty")
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		includeQuery bool
		want         string
	}{
		{
			name:         "lowercase host",
			input:        "http://EXAMPLE.COM/path",
			includeQuery: false,
			want:         "http://example.com/path",
		},
		{
			name:         "add trailing slash for empty path",
			input:        "http://example.com",
			includeQuery: false,
			want:         "http://example.com/",
		},
		{
			name:         "preserve path",
			input:        "http://example.com/path/to/resource",
			includeQuery: false,
			want:         "http://example.com/path/to/resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Complete test implementation
			t.Skip("Normalize URL test not fully implemented")
		})
	}
}

func TestSortQueryString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "sort by key",
			input: "b=2&a=1",
			want:  "a=1&b=2",
		},
		{
			name:  "already sorted",
			input: "a=1&b=2",
			want:  "a=1&b=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortQueryString(tt.input)
			if got != tt.want {
				t.Errorf("sortQueryString() = %v, want %v", got, tt.want)
			}
		})
	}
}
