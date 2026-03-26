package proxy

import "net/http"

// Handler wraps the proxy service as an http.Handler
type Handler struct {
	service *Service
}

// NewHandler creates a new proxy handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.service.ServeHTTP(w, r)
}
