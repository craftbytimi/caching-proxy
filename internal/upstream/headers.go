package upstream

import "net/http"

// Hop-by-hop headers that should not be forwarded
var hopByHopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

// CleanRequestHeaders removes hop-by-hop headers from request
func CleanRequestHeaders(header http.Header) {
	for _, h := range hopByHopHeaders {
		header.Del(h)
	}
}

// CleanResponseHeaders removes hop-by-hop headers from response
func CleanResponseHeaders(header http.Header) {
	for _, h := range hopByHopHeaders {
		header.Del(h)
	}
}

// CopyHeaders copies headers from src to dst
func CopyHeaders(dst, src http.Header) {
	for k, values := range src {
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}
