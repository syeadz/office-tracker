package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIKeyMiddleware_AllowsPublicRoutesWithoutKey(t *testing.T) {
	handler := apiKeyMiddleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	paths := []string{"/health", "/", "/ui", "/ui/"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
	}
}

func TestAPIKeyMiddleware_RejectsProtectedRouteWithoutKey(t *testing.T) {
	handler := apiKeyMiddleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAPIKeyMiddleware_AllowsProtectedRouteWithValidKey(t *testing.T) {
	tests := []struct {
		name         string
		headerName   string
		headerValue  string
		expectStatus int
	}{
		{
			name:         "x-api-key header",
			headerName:   "X-API-Key",
			headerValue:  "secret",
			expectStatus: http.StatusNoContent,
		},
		{
			name:         "authorization bearer header",
			headerName:   "Authorization",
			headerValue:  "Bearer secret",
			expectStatus: http.StatusNoContent,
		},
		{
			name:         "authorization without bearer prefix",
			headerName:   "Authorization",
			headerValue:  "secret",
			expectStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := apiKeyMiddleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			req.Header.Set(tt.headerName, tt.headerValue)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectStatus, rec.Code)
		})
	}
}
