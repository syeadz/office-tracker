package http

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"office/internal/api/dto"
	"office/internal/query"
	"office/internal/repository"
	"office/internal/service"
)

type SessionHandler struct {
	sessionSvc *service.SessionService
}

// parseSessionDateQuery parses a date/time string that may be RFC3339, RFC3339Nano,
// or a bare datetime-local value from <input type="date"> / <input type="datetime-local">.
// Accepted layouts: RFC3339Nano, RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04",
// "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02".
// When inclusiveUpperBound is true the time is advanced to the last nanosecond of
// the given minute (or day for date-only inputs) so that "to" bounds are inclusive.
func parseSessionDateQuery(raw string, inclusiveUpperBound bool) (time.Time, error) {
	type layout struct {
		fmt      string
		dateOnly bool
	}
	layouts := []layout{
		{time.RFC3339Nano, false},
		{time.RFC3339, false},
		{"2006-01-02T15:04:05", false},
		{"2006-01-02T15:04", false},
		{"2006-01-02 15:04:05", false},
		{"2006-01-02 15:04", false},
		{"2006-01-02", true},
	}
	for _, l := range layouts {
		t, err := time.ParseInLocation(l.fmt, raw, time.Local)
		if err != nil {
			continue
		}
		if inclusiveUpperBound {
			if l.dateOnly {
				t = t.AddDate(0, 0, 1).Add(-time.Nanosecond)
			} else {
				t = t.Add(time.Minute).Add(-time.Nanosecond)
			}
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unrecognised date format")
}

func parseSessionCommonFilters(r *http.Request, filter *query.SessionFilter) error {
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		from, err := parseSessionDateQuery(fromStr, false)
		if err != nil {
			return fmt.Errorf("invalid from date")
		}
		filter.From = &from
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		to, err := parseSessionDateQuery(toStr, true)
		if err != nil {
			return fmt.Errorf("invalid to date")
		}
		filter.To = &to
	}

	if name := r.URL.Query().Get("name"); name != "" {
		filter.NameLike = &name
	}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid user_id")
		}
		filter.UserID = &userID
	}

	if discordID := r.URL.Query().Get("discord_id"); discordID != "" {
		filter.DiscordID = &discordID
	}

	if err := applySessionStatusQuery(filter, r); err != nil {
		return err
	}

	if err := applySessionCheckOutMethodQuery(filter, r); err != nil {
		return err
	}

	return nil
}

func applySessionStatusQuery(filter *query.SessionFilter, r *http.Request) error {
	status := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
	if status != "" {
		switch status {
		case "all":
			filter.Status = ""
		case "active", "completed":
			filter.Status = status
		default:
			return fmt.Errorf("invalid status - must be 'all', 'active', or 'completed'")
		}
		return nil
	}

	activeOnlyStr := r.URL.Query().Get("active_only")
	if activeOnlyStr == "" {
		return nil
	}

	activeOnly, err := strconv.ParseBool(activeOnlyStr)
	if err != nil {
		return fmt.Errorf("invalid active_only")
	}

	if activeOnly {
		filter.Status = "active"
	} else {
		filter.Status = "completed"
	}

	return nil
}

func applySessionCheckOutMethodQuery(filter *query.SessionFilter, r *http.Request) error {
	method := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("check_out_method")))
	if method == "" || method == "all" {
		return nil
	}

	switch method {
	case repository.CheckOutMethodRFID,
		repository.CheckOutMethodDiscord,
		repository.CheckOutMethodAPI,
		repository.CheckOutMethodAuto:
		filter.CheckOutMethod = &method
		return nil
	default:
		return fmt.Errorf("invalid check_out_method - must be 'all', 'rfid', 'discord', 'api', or 'auto'")
	}
}

func NewSessionHandler(sessionSvc *service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionSvc: sessionSvc,
	}
}

// CheckInUser handles POST /api/sessions/checkin
func (h *SessionHandler) CheckInUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req dto.UserSessionActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == 0 {
		writeErrorJSON(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if err := h.sessionSvc.CheckInUser(req.UserID); err != nil {
		if errors.Is(err, service.ErrSessionAlreadyOpen) {
			writeErrorJSON(w, http.StatusBadRequest, "user already checked in")
			return
		}
		log.Error("failed to check in user", "user_id", req.UserID, "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to check in user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "checked-in"})
}

// CheckOutUser handles POST /api/sessions/checkout
func (h *SessionHandler) CheckOutUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req dto.UserSessionActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.UserID == 0 {
		writeErrorJSON(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if err := h.sessionSvc.CheckOutUser(req.UserID); err != nil {
		if errors.Is(err, service.ErrNoOpenSession) {
			writeErrorJSON(w, http.StatusBadRequest, "user has no open session")
			return
		}
		log.Error("failed to check out user", "user_id", req.UserID, "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to check out user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "checked-out"})
}

// GetOpenSessions handles GET /api/sessions/open
func (h *SessionHandler) GetOpenSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := query.SessionFilter{ActiveOnly: true, OrderBy: "asc", SortBy: "check_in"}
	result, err := h.sessionSvc.ListSessions(filter, false)
	if err != nil {
		log.Error("failed to get open sessions", "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to get open sessions")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetUserSessions handles GET /api/sessions/user/{userId}
func (h *SessionHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract user ID from path
	userIDStr := r.URL.Path[len("/api/sessions/user/"):]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	filter := query.SessionFilter{UserID: &userID}
	result, err := h.sessionSvc.ListSessions(filter, false)
	if err != nil {
		log.Error("failed to get user sessions", "userID", userID, "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to get user sessions")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateSession handles PUT /api/sessions/{id}
func (h *SessionHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract ID from path
	idStr := r.URL.Path[len("/api/sessions/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid session id")
		return
	}

	var req dto.UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("invalid request body", "err", err)
		writeErrorJSON(w, http.StatusBadRequest, "invalid request")
		return
	}

	// At least one field must be provided
	if req.CheckIn == nil && req.CheckOut == nil {
		writeErrorJSON(w, http.StatusBadRequest, "at least one field (check_in or check_out) is required")
		return
	}

	err = h.sessionSvc.UpdateSession(id, req.CheckIn, req.CheckOut)
	if err != nil {
		log.Error("failed to update session", "id", id, "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to update session")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteSession handles DELETE /api/sessions/{id}
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract ID from path
	idStr := r.URL.Path[len("/api/sessions/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid session id")
		return
	}

	err = h.sessionSvc.DeleteSession(id)
	if err != nil {
		log.Error("failed to delete session", "id", id, "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to delete session")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPresence handles GET /api/presence - returns currently active users
func (h *SessionHandler) GetPresence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := query.SessionFilter{ActiveOnly: true, OrderBy: "asc", SortBy: "check_in"}
	result, err := h.sessionSvc.ListSessions(filter, false)
	if err != nil {
		log.Error("failed to get active sessions", "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to get active sessions")
		return
	}

	sessions, ok := result.([]*repository.SessionWithUser)
	if !ok {
		log.Error("unexpected presence payload type")
		writeErrorJSON(w, http.StatusInternalServerError, "failed to get active sessions")
		return
	}

	responses := make([]dto.PresenceResponse, 0, len(sessions))
	for _, session := range sessions {
		responses = append(responses, dto.PresenceResponse{
			UserName: session.UserName,
			CheckIn:  session.CheckIn,
		})
	}

	writeJSON(w, http.StatusOK, responses)
}

// CountSessions handles GET /api/sessions/count with optional filters
func (h *SessionHandler) CountSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := query.SessionFilter{}
	if err := parseSessionCommonFilters(r, &filter); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	count, err := h.sessionSvc.CountSessions(filter)
	if err != nil {
		log.Error("failed to count sessions", "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to count sessions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{"total": count})
}

// ListSessions handles GET /api/sessions with query parameters
// Supports: ?from=&to=&name=&limit=&offset=&order=&sort_by=
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := query.SessionFilter{}
	if err := parseSessionCommonFilters(r, &filter); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			writeErrorJSON(w, http.StatusBadRequest, "invalid limit")
			return
		}
		filter.Limit = limit
	} else {
		filter.Limit = 100 // default
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			writeErrorJSON(w, http.StatusBadRequest, "invalid offset")
			return
		}
		filter.Offset = offset
	}

	if order := r.URL.Query().Get("order"); order != "" {
		order = strings.ToLower(order)
		if order == "asc" || order == "desc" {
			filter.OrderBy = order
		} else {
			writeErrorJSON(w, http.StatusBadRequest, "invalid order - must be 'asc' or 'desc'")
			return
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		sortBy = strings.ToLower(sortBy)
		if sortBy == "check_in" || sortBy == "check_out" || sortBy == "user_name" {
			filter.SortBy = sortBy
		} else {
			writeErrorJSON(w, http.StatusBadRequest, "invalid sort_by - must be 'check_in', 'check_out', or 'user_name'")
			return
		}
	}

	result, err := h.sessionSvc.ListSessions(filter, true)
	if err != nil {
		log.Error("failed to list sessions", "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteSessions handles DELETE /api/sessions (bulk delete with filters)
func (h *SessionHandler) DeleteSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := query.SessionFilter{}

	if err := parseSessionCommonFilters(r, &filter); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			writeErrorJSON(w, http.StatusBadRequest, "invalid limit")
			return
		}
		filter.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			writeErrorJSON(w, http.StatusBadRequest, "invalid offset")
			return
		}
		filter.Offset = offset
	}

	if order := r.URL.Query().Get("order"); order != "" {
		order = strings.ToLower(order)
		if order == "asc" || order == "desc" {
			filter.OrderBy = order
		} else {
			writeErrorJSON(w, http.StatusBadRequest, "invalid order - must be 'asc' or 'desc'")
			return
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		sortBy = strings.ToLower(sortBy)
		if sortBy == "check_in" || sortBy == "check_out" || sortBy == "user_name" {
			filter.SortBy = sortBy
		} else {
			writeErrorJSON(w, http.StatusBadRequest, "invalid sort_by - must be 'check_in', 'check_out', or 'user_name'")
			return
		}
	}

	// Safety check - require at least one filter to prevent accidental delete all
	if filter.From == nil && filter.To == nil && filter.NameLike == nil && filter.UserID == nil && filter.DiscordID == nil && filter.Status == "" && filter.CheckOutMethod == nil {
		writeErrorJSON(w, http.StatusBadRequest, "filter required for bulk delete (e.g., ?from=date&to=date)")
		return
	}

	count, err := h.sessionSvc.DeleteSessions(filter)
	if err != nil {
		log.Error("failed to delete sessions", "err", err)
		writeErrorJSON(w, http.StatusInternalServerError, "failed to delete sessions")
		return
	}

	writeJSON(w, http.StatusOK, dto.DeleteResult{Deleted: count})
}

// ExportSessionsCSV handles GET /api/sessions/export - exports sessions as CSV
func (h *SessionHandler) ExportSessionsCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use the same filter parsing as ListSessions.
	// Keep CSV default ordering stable and intuitive for exports (oldest first).
	filter := query.SessionFilter{OrderBy: "asc"}

	if err := parseSessionCommonFilters(r, &filter); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		filter.Limit = limit
	} else {
		filter.Limit = 100 // default
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		filter.Offset = offset
	}

	if order := r.URL.Query().Get("order"); order != "" {
		order = strings.ToLower(order)
		if order == "asc" || order == "desc" {
			filter.OrderBy = order
		} else {
			http.Error(w, "invalid order - must be 'asc' or 'desc'", http.StatusBadRequest)
			return
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		sortBy = strings.ToLower(sortBy)
		if sortBy == "check_in" || sortBy == "check_out" || sortBy == "user_name" {
			filter.SortBy = sortBy
		} else {
			http.Error(w, "invalid sort_by - must be 'check_in', 'check_out', or 'user_name'", http.StatusBadRequest)
			return
		}
	}

	result, err := h.sessionSvc.ListSessions(filter, false)
	if err != nil {
		log.Error("failed to get sessions", "err", err)
		http.Error(w, "failed to get sessions", http.StatusInternalServerError)
		return
	}

	sessions := result.([]*repository.SessionWithUser)

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=sessions.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"UserName", "CheckIn", "CheckOut", "CheckOutMethod", "Duration(minutes)"})

	// Write data
	for _, s := range sessions {
		checkInStr := s.CheckIn.Format(time.RFC3339)
		checkOutStr := ""
		duration := ""

		if s.CheckOut != nil {
			checkOutStr = s.CheckOut.Format(time.RFC3339)
			durationMinutes := s.CheckOut.Sub(s.CheckIn).Minutes()
			duration = fmt.Sprintf("%.2f", durationMinutes)
		}

		writer.Write([]string{s.UserName, checkInStr, checkOutStr, s.CheckOutMethod, duration})
	}
}

// CheckoutAll handles POST /api/sessions/checkout-all - checks out all active sessions
func (h *SessionHandler) CheckoutAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all active sessions
	filter := query.SessionFilter{ActiveOnly: true, OrderBy: "asc", SortBy: "check_in"}
	result, err := h.sessionSvc.ListSessions(filter, false)
	if err != nil {
		log.Error("failed to get active sessions", "err", err)
		http.Error(w, "failed to get active sessions", http.StatusInternalServerError)
		return
	}

	sessions := result.([]*repository.SessionWithUser)
	if len(sessions) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "No active sessions to check out",
			"count":    0,
			"sessions": []string{},
		})
		return
	}

	// Check out all active sessions
	checkedOutCount := 0
	var checkedOutUsers []string

	for _, sess := range sessions {
		err := h.sessionSvc.CheckOutUser(sess.UserID)
		if err != nil {
			log.Error("failed to check out user via service", "user_id", sess.UserID, "error", err)
			continue
		}
		checkedOutCount++
		checkedOutUsers = append(checkedOutUsers, sess.UserName)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  fmt.Sprintf("Successfully checked out %d session(s)", checkedOutCount),
		"count":    checkedOutCount,
		"sessions": checkedOutUsers,
	})

	log.Info("bulk checkout completed", "count", checkedOutCount)
}

// CheckoutUser handles POST /api/sessions/checkout/{userId} - checks out a specific user
func (h *SessionHandler) CheckoutUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from path
	userIDStr := r.URL.Path[len("/api/sessions/checkout/"):]

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Find the open session for the user (to get session_id)
	sessionRepo := h.sessionSvc.Sessions
	openSessionID, err := sessionRepo.GetOpenSession(userID)
	if err != nil {
		log.Error("failed to find open session for checkout", "user_id", userID, "error", err)
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "No active session found for this user",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to check out user",
			})
		}
		return
	}

	err = h.sessionSvc.CheckOutUser(userID)
	if err != nil {
		log.Error("failed to check out user via service", "user_id", userID, "error", err)
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrNoOpenSession) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "No active session found for this user",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to check out user",
			})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "User checked out successfully",
		"session_id": openSessionID,
		"user_id":    userID,
	})

	log.Info("user checked out", "user_id", userID, "session_id", openSessionID)
}
