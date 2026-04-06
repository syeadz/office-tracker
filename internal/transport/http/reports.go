package http

import (
	"net/http"
	"strings"

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

	if h.reports == nil || !h.reports.IsAvailable() {
		writeErrorJSON(w, http.StatusServiceUnavailable, "reports are unavailable (missing startup guild/channel configuration)")
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
// POST /api/reports/toggle?enabled=true|false[&period=weekly|monthly|all]
func (h *ReportsHandlers) HandleToggleReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if h.reports == nil || !h.reports.IsAvailable() {
		writeErrorJSON(w, http.StatusServiceUnavailable, "reports are fully disabled (missing startup guild/channel configuration), toggle unavailable")
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

	period := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("period")))
	if period == "" {
		period = "all"
	}

	switch period {
	case "all":
		h.reports.SetEnabled(enabled)
	case "weekly":
		h.reports.SetWeeklyEnabled(enabled)
	case "monthly":
		h.reports.SetMonthlyEnabled(enabled)
	default:
		writeErrorJSON(w, http.StatusBadRequest, "invalid 'period' value, must be 'weekly', 'monthly', or 'all'")
		return
	}

	status := "disabled"
	if h.reports.IsEnabled() {
		status = "enabled"
	}

	periodStatus := "disabled"
	switch period {
	case "weekly":
		if h.reports.IsWeeklyEnabled() {
			periodStatus = "enabled"
		}
	case "monthly":
		if h.reports.IsMonthlyEnabled() {
			periodStatus = "enabled"
		}
	case "all":
		if h.reports.IsEnabled() {
			periodStatus = "enabled"
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"period":          period,
		"period_status":   periodStatus,
		"enabled":         h.reports.IsEnabled(),
		"weekly_enabled":  h.reports.IsWeeklyEnabled(),
		"monthly_enabled": h.reports.IsMonthlyEnabled(),
		"status":          status,
		"message":         "Reports " + period + " setting updated successfully",
	})
}

// HandleGetReportsStatus returns the current status of scheduled reports
// GET /api/reports/status
func (h *ReportsHandlers) HandleGetReportsStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if h.reports == nil || !h.reports.IsAvailable() {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"available":       false,
			"fully_disabled":  true,
			"enabled":         false,
			"weekly_enabled":  false,
			"monthly_enabled": false,
			"status":          "disabled",
			"message":         "Reports are fully disabled (missing startup guild/channel configuration)",
		})
		return
	}

	enabled := h.reports.IsEnabled()
	status := "disabled"
	if enabled {
		status = "enabled"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"available":       true,
		"fully_disabled":  false,
		"enabled":         enabled,
		"weekly_enabled":  h.reports.IsWeeklyEnabled(),
		"monthly_enabled": h.reports.IsMonthlyEnabled(),
		"status":          status,
		"message":         "Reports are configurable",
	})
}
