package http

import (
	"net/http"
	"strings"
	"time"
)

// MiddlewareConfig holds configuration for HTTP middlewares
type MiddlewareConfig struct {
	APIKey        string
	CORSOrigins   string
	CORSEnabled   bool
	APIKeyEnabled bool
}

// loggingMiddleware logs HTTP requests with method, path, and duration
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info("http request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// If allowedOrigins is "*", allow all origins
			if allowedOrigins == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if allowedOrigins != "" {
				// Check if the request origin is in the allowed list
				origins := strings.Split(allowedOrigins, ",")
				for _, o := range origins {
					o = strings.TrimSpace(o)
					if o == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// apiKeyMiddleware validates API key from request headers
func apiKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip API key check for public endpoints
			if isPublicNoAuthPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Check for API key in header
			requestKey := r.Header.Get("X-API-Key")
			if requestKey == "" {
				requestKey = r.Header.Get("Authorization")
				// Remove "Bearer " prefix if present
				if after, ok := strings.CutPrefix(requestKey, "Bearer "); ok {
					requestKey = after
				}
			}

			if requestKey != apiKey {
				log.Warn("unauthorized request - invalid API key", "path", r.URL.Path, "remote_addr", r.RemoteAddr)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isPublicNoAuthPath(path string) bool {
	return path == "/health" || path == "/" || path == "/ui" || strings.HasPrefix(path, "/ui/")
}

// ChainMiddleware applies multiple middlewares in order
func ChainMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
