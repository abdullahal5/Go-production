package server

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go-prod-project/internal/middleware"
)

// NewRouter registers the application HTTP routes and middlewares.
func NewRouter(
	listingHandler *ListingHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Liveness & health probe
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Prometheus metrics endpoint
	mux.Handle("GET /metrics", promhttp.Handler())

	// Listing REST endpoints
	mux.HandleFunc("POST /listings", listingHandler.Create)
	mux.HandleFunc("GET /listings", listingHandler.List)
	mux.HandleFunc("GET /listings/{id}", listingHandler.GetByID)
	mux.HandleFunc("DELETE /listings/{id}", listingHandler.Delete)

	// Middleware chain: RequestId -> Prometheus Metrics -> Router
	return middleware.RequestId(middleware.Metrics(mux))
}

// LogRouter is a middleware helper for request logging if needed.
func LogRouter(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(
			r.Context(),
			"incoming request",
			"method", r.Method,
			"path", r.URL.Path,
		)

		next.ServeHTTP(w, r)
	})
}
