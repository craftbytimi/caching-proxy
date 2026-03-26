package model

import "net/http"

// RequestInfo holds information about an incoming request
type RequestInfo struct {
	Method string
	URL    string
	Header http.Header
}
