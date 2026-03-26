package httputil

import (
	"net/http"
)

// WriteJSON writes a JSON response with proper headers
func WriteJSON(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
}
