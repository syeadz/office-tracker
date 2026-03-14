package http

import (
	"net/http"

	"office/internal/logging"
	"office/internal/service"
)

var httpLogger = logging.Component("http")

// ReportsHandlers provides HTTP endpoints for reports data
type ReportsHandlers struct {
	reports *service.ReportsService
}

// NewReportsHandlers creates a new reports handlers instance
func NewReportsHandlers(reports *service.ReportsService) *ReportsHandlers {
	return &ReportsHandlers{
		reports: reports,
	}
}

// HandleGetWeeklyReport returns the latest weekly report data
func (h *ReportsHandlers) HandleGetWeeklyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	report, err := h.reports.GetLatestWeeklyReport()
	if err != nil {
		httpLogger.Error("failed to get weekly report", "error", err)
		writeErrorJSON(w, http.StatusNotFound, "no report available")
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// HandleToggleReports enables or disables scheduled reports
// POST /api/reports/toggle?enabled=true|false
func (h *ReportsHandlers) HandleToggleReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	enabledStr := r.URL.Query().Get("enabled")
	if enabledStr == "" {
		writeErrorJSON(w, http.StatusBadRequest, "missing 'enabled' query parameter (true/false)")
		return
	}

	var enabled bool
	switch enabledStr {
	case "true":
		enabled = true
	case "false":
		enabled = false
	default:
		writeErrorJSON(w, http.StatusBadRequest, "invalid 'enabled' value, must be 'true' or 'false'")
		return
	}

	h.reports.SetEnabled(enabled)

	status := "disabled"
	if enabled {
		status = "enabled"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"status":  status,
		"message": "Reports " + status + " successfully",
	})
}

// HandleGetReportsStatus returns the current status of scheduled reports
// GET /api/reports/status
func (h *ReportsHandlers) HandleGetReportsStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	enabled := h.reports.IsEnabled()
	status := "disabled"
	if enabled {
		status = "enabled"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": enabled,
		"status":  status,
	})
}
