package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/craftbytimi/caching-proxy/internal/store"
)

// HealthHandler returns a health check handler
func HealthHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// Check store connectivity
		if err := store.Ping(ctx); err != nil {
			http.Error(w, fmt.Sprintf("store unhealthy: %v", err), http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}
}

// RootHandler returns a simple root handler
func RootHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "caching-proxy online")
	}
}
