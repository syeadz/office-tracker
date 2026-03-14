// Package http provides a simple HTTP server for health checks and future API endpoints.
package http

import (
	"context"
	"fmt"
	"net/http"

	"office/internal/logging"
	"office/internal/service"
)

type Server struct {
	http *http.Server
}

var log = logging.Component("http")

// New creates a new HTTP server with optional reports and in-memory environment services.
func New(port string, attSvc *service.AttendanceService, userSvc *service.UserService, sessionSvc *service.SessionService, statsSvc *service.OfficeStatsService, environmentSvc *service.EnvironmentService, reportsSvc *service.ReportsService, mwConfig MiddlewareConfig) *Server {
	mux := http.NewServeMux()
	espHealthSvc := service.NewESPHealthService(service.DefaultESPHealthMaxAge)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	// Prometheus metrics endpoint
	mux.Handle("/metrics", NewMetricsHandler(sessionSvc, environmentSvc, espHealthSvc))

	// Management UI
	mux.HandleFunc("/", UIHandler())
	mux.HandleFunc("/ui", UIHandler())

	// RFID endpoint
	mux.Handle("/api/rfid/scan", RfidHandler(attSvc))
	mux.Handle("/api/rfid/scans", ScanHistoryHandler(attSvc))

	// User endpoints
	userHandler := NewUserHandler(userSvc)
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.ListUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		case http.MethodDelete:
			userHandler.DeleteUsers(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/users/export", userHandler.ExportUsersCSV)
	mux.HandleFunc("/api/users/import", userHandler.ImportUsersCSV)
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/users" || r.URL.Path == "/api/users/" {
			switch r.Method {
			case http.MethodGet:
				userHandler.ListUsers(w, r)
			case http.MethodPost:
				userHandler.CreateUser(w, r)
			case http.MethodDelete:
				userHandler.DeleteUsers(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			switch r.Method {
			case http.MethodGet:
				userHandler.GetUser(w, r)
			case http.MethodPut:
				userHandler.UpdateUser(w, r)
			case http.MethodDelete:
				userHandler.DeleteUser(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})

	// Session endpoints
	sessionHandler := NewSessionHandler(sessionSvc)
	mux.HandleFunc("/api/presence", sessionHandler.GetPresence)
	mux.HandleFunc("/api/sessions/checkin", sessionHandler.CheckInUser)
	mux.HandleFunc("/api/sessions/checkout", sessionHandler.CheckOutUser)
	mux.HandleFunc("/api/sessions/count", sessionHandler.CountSessions)
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sessionHandler.ListSessions(w, r)
		case http.MethodDelete:
			sessionHandler.DeleteSessions(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/sessions/export", sessionHandler.ExportSessionsCSV)
	mux.HandleFunc("/api/sessions/checkout-all", sessionHandler.CheckoutAll)
	mux.HandleFunc("/api/sessions/checkout/", sessionHandler.CheckoutUser)
	mux.HandleFunc("/api/sessions/open", sessionHandler.GetOpenSessions)
	mux.HandleFunc("/api/sessions/user/", sessionHandler.GetUserSessions)
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sessions" || r.URL.Path == "/api/sessions/" {
			switch r.Method {
			case http.MethodGet:
				sessionHandler.ListSessions(w, r)
			case http.MethodDelete:
				sessionHandler.DeleteSessions(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			switch r.Method {
			case http.MethodPut:
				sessionHandler.UpdateSession(w, r)
			case http.MethodDelete:
				sessionHandler.DeleteSession(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})

	// Statistics endpoints
	statsHandler := NewStatsHandler(statsSvc)
	mux.HandleFunc("/api/statistics/leaderboard", statsHandler.GetLeaderboard)
	mux.HandleFunc("/api/statistics/weekly", statsHandler.GetWeeklyReport)
	mux.HandleFunc("/api/statistics/monthly", statsHandler.GetMonthlyReport)
	mux.HandleFunc("/api/statistics/report", statsHandler.GetCustomReport)
	mux.HandleFunc("/api/statistics/users/", statsHandler.GetUserStats)

	// Environment endpoints
	environmentHandler := NewEnvironmentHandler(environmentSvc)
	mux.HandleFunc("/api/environment", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			environmentHandler.GetLatest(w, r)
		case http.MethodPost:
			environmentHandler.UpdateLatest(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ESP32 health endpoints
	espHealthHandler := NewESPHealthHandler(espHealthSvc)
	mux.HandleFunc("/api/devices/health", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			espHealthHandler.List(w, r)
		case http.MethodPost:
			espHealthHandler.Upsert(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Reports endpoints (optional)
	if reportsSvc != nil {
		reportsHandler := NewReportsHandlers(reportsSvc)
		mux.HandleFunc("/api/reports/weekly", reportsHandler.HandleGetWeeklyReport)
		mux.HandleFunc("/api/reports/status", reportsHandler.HandleGetReportsStatus)
		mux.HandleFunc("/api/reports/toggle", reportsHandler.HandleToggleReports)
	}

	// Build middleware chain
	var middlewares []func(http.Handler) http.Handler

	// Always include logging middleware
	middlewares = append(middlewares, loggingMiddleware)

	// Add CORS middleware if enabled
	if mwConfig.CORSEnabled {
		middlewares = append(middlewares, corsMiddleware(mwConfig.CORSOrigins))
	}

	// Add API key middleware if enabled
	if mwConfig.APIKeyEnabled {
		middlewares = append(middlewares, apiKeyMiddleware(mwConfig.APIKey))
	}

	// Apply middleware chain
	handler := ChainMiddleware(mux, middlewares...)

	return &Server{
		http: &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		},
	}
}

// Start runs the HTTP server in a separate goroutine.
func (s *Server) Start() {
	go func() {
		log.Info("HTTP listening on", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start HTTP server", "error", err)
		}
	}()
}

// Stop gracefully shuts down the HTTP server with a timeout context.
func (s *Server) Stop(ctx context.Context) error {
	log.Info("HTTP server shutting down")
	return s.http.Shutdown(ctx)
}

// Handler returns the configured HTTP handler chain.
// Intended for testing and in-process integration scenarios.
func (s *Server) Handler() http.Handler {
	return s.http.Handler
}
