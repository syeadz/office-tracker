package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"office/internal/api/dto"
	"office/internal/service"
)

// StatsHandler handles statistics-related HTTP requests
type StatsHandler struct {
	statsSvc *service.OfficeStatsService
}

// NewStatsHandler creates a new StatsHandler
func NewStatsHandler(statsSvc *service.OfficeStatsService) *StatsHandler {
	return &StatsHandler{statsSvc: statsSvc}
}

// GetLeaderboard returns the leaderboard for a given rank_by value and date range.
// GET /api/statistics/leaderboard?rank_by=hours&limit=10&from=2026-02-01&to=2026-02-08
func (h *StatsHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	includeAutoCheckout, parseErr := parseIncludeAutoCheckout(r)
	if parseErr != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid 'include_auto_checkout' value (use true/false)")
		return
	}
	excludeAutoCheckout := !includeAutoCheckout

	rankBy := parseRankBy(r.URL.Query().Get("rank_by"))
	if rankBy == "" {
		rankBy = "hours"
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "invalid 'from' date format, use YYYY-MM-DD")
			return
		}
	} else {
		from = time.Now().Add(-7 * 24 * time.Hour)
	}

	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "invalid 'to' date format, use YYYY-MM-DD")
			return
		}
		// Set to end of day
		to = to.AddDate(0, 0, 1).Add(-1 * time.Second)
	} else {
		to = time.Now()
	}

	leaderboard, err := h.statsSvc.GetLeaderboardWithAutoCheckout(from, to, rankBy, limit, excludeAutoCheckout)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to fetch leaderboard")
		return
	}

	writeJSON(w, http.StatusOK, dto.LeaderboardResponse{
		RankBy: rankBy,
		Period: from.Format("2006-01-02") + " to " + to.Format("2006-01-02"),
		Count:  int64(len(leaderboard)),
		Users:  leaderboard,
	})
}

// GetWeeklyReport returns statistics for the current week
// GET /api/statistics/weekly
func (h *StatsHandler) GetWeeklyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	includeAutoCheckout, parseErr := parseIncludeAutoCheckout(r)
	if parseErr != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid 'include_auto_checkout' value (use true/false)")
		return
	}
	rankBy := parseRankBy(r.URL.Query().Get("rank_by"))
	if rankBy == "" {
		rankBy = "hours"
	}

	report, err := h.statsSvc.GetWeeklyReport(rankBy, !includeAutoCheckout)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to fetch weekly report")
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// GetMonthlyReport returns statistics for a specific month
// GET /api/statistics/monthly?year=2026&month=2
func (h *StatsHandler) GetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	includeAutoCheckout, parseErr := parseIncludeAutoCheckout(r)
	if parseErr != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid 'include_auto_checkout' value (use true/false)")
		return
	}
	excludeAutoCheckout := !includeAutoCheckout
	rankBy := parseRankBy(r.URL.Query().Get("rank_by"))
	if rankBy == "" {
		rankBy = "hours"
	}

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	var year int
	var month time.Month
	var err error

	// If both are provided, parse them; otherwise use current month
	if yearStr != "" || monthStr != "" {
		// Both must be provided together
		if yearStr == "" || monthStr == "" {
			writeErrorJSON(w, http.StatusBadRequest, "both 'year' and 'month' must be provided together")
			return
		}

		year, err = strconv.Atoi(yearStr)
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "invalid year format")
			return
		}

		monthInt, err := strconv.Atoi(monthStr)
		if err != nil || monthInt < 1 || monthInt > 12 {
			writeErrorJSON(w, http.StatusBadRequest, "invalid month (must be 1-12)")
			return
		}
		month = time.Month(monthInt)
	} else {
		// Default to current month/year
		now := time.Now()
		year = now.Year()
		month = now.Month()
	}

	report, err := h.statsSvc.GetMonthlyReport(year, month, rankBy, excludeAutoCheckout)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to fetch monthly report")
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// GetCustomReport returns statistics for a custom date range
// GET /api/statistics/report?from=2026-02-01&to=2026-02-08&limit=10
func (h *StatsHandler) GetCustomReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	includeAutoCheckout, err := parseIncludeAutoCheckout(r)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid 'include_auto_checkout' value (use true/false)")
		return
	}
	excludeAutoCheckout := !includeAutoCheckout
	rankBy := parseRankBy(r.URL.Query().Get("rank_by"))
	if rankBy == "" {
		rankBy = "hours"
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		writeErrorJSON(w, http.StatusBadRequest, "'from' and 'to' query parameters are required")
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid 'from' date format, use YYYY-MM-DD")
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid 'to' date format, use YYYY-MM-DD")
		return
	}

	// Set to end of day
	to = to.AddDate(0, 0, 1).Add(-1 * time.Second)

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	report, err := h.statsSvc.GetCustomReport(from, to, limit, rankBy, excludeAutoCheckout)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// GetUserStats returns statistics for a specific user over a date range
// GET /api/statistics/users/{userId}?from=2026-02-01&to=2026-02-08
func (h *StatsHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	includeAutoCheckout, err := parseIncludeAutoCheckout(r)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid 'include_auto_checkout' value (use true/false)")
		return
	}
	excludeAutoCheckout := !includeAutoCheckout

	userIDStr := r.URL.Path[len("/api/statistics/users/"):]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var from, to time.Time

	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "invalid 'from' date format, use YYYY-MM-DD")
			return
		}
	} else {
		from = time.Now().Add(-30 * 24 * time.Hour)
	}

	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "invalid 'to' date format, use YYYY-MM-DD")
			return
		}
		to = to.AddDate(0, 0, 1).Add(-1 * time.Second)
	} else {
		to = time.Now()
	}

	stats, err := h.statsSvc.GetUserStatsWithAutoCheckout(userID, from, to, excludeAutoCheckout)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			writeErrorJSON(w, http.StatusNotFound, "user not found")
			return
		}
		writeErrorJSON(w, http.StatusInternalServerError, "failed to fetch user statistics")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Helper functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func parseIncludeAutoCheckout(r *http.Request) (bool, error) {
	value := r.URL.Query().Get("include_auto_checkout")
	if value == "" {
		return false, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}

	return parsed, nil
}

func parseRankBy(value string) string {
	switch value {
	case "hours", "visits":
		return value
	default:
		return ""
	}
}
