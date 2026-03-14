package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"office/internal/api/dto"
	"office/internal/query"
	"office/internal/repository"
	"office/internal/service"
	httptransport "office/internal/transport/http"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOpenSessions(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	// Seed users and create open sessions
	user1 := helpers.SeedUser(t, db, "Alice", "UID001", "")
	user2 := helpers.SeedUser(t, db, "Bob", "UID002", "")
	helpers.SeedSession(t, db, user1.ID)
	helpers.SeedSession(t, db, user2.ID)

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/open", nil)
	w := httptest.NewRecorder()

	handler.GetOpenSessions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var sessions []map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&sessions)
	require.NoError(t, err)

	assert.Len(t, sessions, 2)
}

func TestGetOpenSessionsEmpty(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/open", nil)
	w := httptest.NewRecorder()

	handler.GetOpenSessions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var sessions []map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&sessions)
	require.NoError(t, err)

	assert.Len(t, sessions, 0)
}

func TestGetUserSessions(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	// Seed a user and create multiple sessions
	user := helpers.SeedUser(t, db, "Charlie", "UID003", "")
	sessionRepo := &repository.SessionRepo{DB: db}

	// Create first session and check out
	sessionID1 := helpers.SeedSession(t, db, user.ID)
	err := sessionRepo.CheckOut(sessionID1)
	require.NoError(t, err)

	// Create second session (still open)
	helpers.SeedSession(t, db, user.ID)

	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get sessions for existing user",
			userID:         "1",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "get sessions for non-existent user",
			userID:         "999",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "invalid user id",
			userID:         "abc",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/sessions/user/"+tt.userID, nil)
			w := httptest.NewRecorder()

			handler.GetUserSessions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var sessions []map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&sessions)
				require.NoError(t, err)

				assert.Len(t, sessions, tt.expectedCount)
			}
		})
	}
}

func TestGetPresence(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	// Seed users and create active sessions
	user1 := helpers.SeedUser(t, db, "Alice Active", "UID001", "")
	user2 := helpers.SeedUser(t, db, "Bob Active", "UID002", "")
	helpers.SeedSession(t, db, user1.ID)
	helpers.SeedSession(t, db, user2.ID)

	sessionRepo := &repository.SessionRepo{DB: db}

	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/presence", nil)
	w := httptest.NewRecorder()

	handler.GetPresence(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var presence []map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&presence)
	require.NoError(t, err)

	assert.Len(t, presence, 2)
	// Verify no check_out field in response
	for _, p := range presence {
		assert.Contains(t, p, "user_name")
		assert.Contains(t, p, "check_in")
		assert.NotContains(t, p, "check_out")
		assert.NotContains(t, p, "active")
	}
}

func TestSessionHandler_ListSessions_StatusFilters(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	activeUser := helpers.SeedUser(t, db, "Active User", "UID_STATUS_ACTIVE", "discord_active_status")
	completedUser := helpers.SeedUser(t, db, "Completed User", "UID_STATUS_COMPLETED", "discord_completed_status")

	activeSessionID := helpers.SeedSession(t, db, activeUser.ID)
	completedSessionID := helpers.SeedSession(t, db, completedUser.ID)
	require.NoError(t, sessionRepo.CheckOut(completedSessionID))

	tests := []struct {
		name           string
		query          string
		expectedID     int64
		expectedUserID int64
		expectedActive bool
	}{
		{
			name:           "status active",
			query:          "?status=active",
			expectedID:     activeSessionID,
			expectedUserID: activeUser.ID,
			expectedActive: true,
		},
		{
			name:           "status completed",
			query:          "?status=completed",
			expectedID:     completedSessionID,
			expectedUserID: completedUser.ID,
			expectedActive: false,
		},
		{
			name:           "legacy active_only true",
			query:          "?active_only=true",
			expectedID:     activeSessionID,
			expectedUserID: activeUser.ID,
			expectedActive: true,
		},
		{
			name:           "legacy active_only false",
			query:          "?active_only=false",
			expectedID:     completedSessionID,
			expectedUserID: completedUser.ID,
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/sessions"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.ListSessions(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var sessions []dto.SessionResponse
			err := json.NewDecoder(w.Body).Decode(&sessions)
			require.NoError(t, err)
			require.Len(t, sessions, 1)
			assert.Equal(t, tt.expectedID, sessions[0].ID)
			assert.Equal(t, tt.expectedUserID, sessions[0].UserID)
			assert.Equal(t, tt.expectedActive, sessions[0].Active)
		})
	}
}

func TestSessionHandler_DeleteSessions_UsesPaginationAndOrdering(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	user := helpers.SeedUser(t, db, "Paginated Delete", "UID_DELETE_PAGE", "discord_delete_page")

	sessionID1 := helpers.SeedSession(t, db, user.ID)
	require.NoError(t, sessionRepo.CheckOut(sessionID1))
	time.Sleep(10 * time.Millisecond)

	sessionID2 := helpers.SeedSession(t, db, user.ID)
	require.NoError(t, sessionRepo.CheckOut(sessionID2))
	time.Sleep(10 * time.Millisecond)

	sessionID3 := helpers.SeedSession(t, db, user.ID)
	require.NoError(t, sessionRepo.CheckOut(sessionID3))

	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/api/sessions?user_id=%d&order=asc&limit=1&offset=1", user.ID),
		nil,
	)
	w := httptest.NewRecorder()

	handler.DeleteSessions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result dto.DeleteResult
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Deleted)

	filter := query.SessionFilter{UserID: &user.ID, OrderBy: "asc"}
	remaining, err := sessionRepo.List(filter)
	require.NoError(t, err)
	require.Len(t, remaining, 2)
	assert.Equal(t, sessionID1, remaining[0].ID)
	assert.Equal(t, sessionID3, remaining[1].ID)
}

func TestSessionHandler_CountSessions(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	userActive := helpers.SeedUser(t, db, "Count Active", "UID_COUNT_ACTIVE", "discord_count_active")
	userCompleted := helpers.SeedUser(t, db, "Count Completed", "UID_COUNT_COMPLETED", "discord_count_completed")

	helpers.SeedSession(t, db, userActive.ID)
	completedSessionID := helpers.SeedSession(t, db, userCompleted.ID)
	require.NoError(t, sessionRepo.CheckOut(completedSessionID))

	tests := []struct {
		name          string
		query         string
		expectedTotal int64
		expectedCode  int
	}{
		{name: "all sessions", query: "", expectedTotal: 2, expectedCode: http.StatusOK},
		{name: "active status", query: "?status=active", expectedTotal: 1, expectedCode: http.StatusOK},
		{name: "completed status", query: "?status=completed", expectedTotal: 1, expectedCode: http.StatusOK},
		{name: "filter by user", query: fmt.Sprintf("?user_id=%d", userCompleted.ID), expectedTotal: 1, expectedCode: http.StatusOK},
		{name: "invalid status", query: "?status=invalid", expectedTotal: 0, expectedCode: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/sessions/count"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.CountSessions(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)

			if w.Code != http.StatusOK {
				return
			}

			var response map[string]int64
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, response["total"])
		})
	}
}

func TestExportSessionsCSV(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	// Seed users and sessions
	user1 := helpers.SeedUser(t, db, "CSV User 1", "UID001", "DID001")
	user2 := helpers.SeedUser(t, db, "CSV User 2", "UID002", "DID002")
	helpers.SeedSession(t, db, user1.ID)
	helpers.SeedSession(t, db, user2.ID)
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	// Test all filters
	tests := []struct {
		name       string
		params     string
		expectUser string
	}{
		{"no filters", "", "CSV User 1"},
		{"filter by name", "name=CSV+User+2", "CSV User 2"},
		{"filter by user_id", "user_id=2", "CSV User 2"},
		{"filter by discord_id", "discord_id=DID001", "CSV User 1"},
		{"active_only true", "active_only=true", "CSV User 1"},
		{"limit 1", "limit=1", "CSV User 1"},
		{"offset 1", "offset=1", "CSV User 2"},
		{"order asc", "order=asc", "CSV User 1"},
		{"sort_by user_name", "sort_by=user_name", "CSV User 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/sessions/export"
			if tt.params != "" {
				url += "?" + tt.params
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			handler.ExportSessionsCSV(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
			assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
			assert.Contains(t, w.Body.String(), "UserName,CheckIn,CheckOut,Duration(minutes)")
			assert.Contains(t, w.Body.String(), tt.expectUser)
		})
	}

	// Test date filters
	t.Run("from/to filters", func(t *testing.T) {
		from := "?from=2000-01-01T00:00:00Z"
		to := "&to=2100-01-01T00:00:00Z"
		req := httptest.NewRequest(http.MethodGet, "/api/sessions/export"+from+to, nil)
		w := httptest.NewRecorder()
		handler.ExportSessionsCSV(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "CSV User 1")
	})
}

func TestCheckoutUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	// Seed user with an active session
	user := helpers.SeedUser(t, db, "Checkout User", "UID_CHECKOUT", "")
	sessionID := helpers.SeedSession(t, db, user.ID)

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "checkout active session",
			userID:         "1",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "checkout non-existent user",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
			checkResponse:  false,
		},
		{
			name:           "invalid user id",
			userID:         "abc",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/sessions/checkout/"+tt.userID, nil)
			w := httptest.NewRecorder()

			handler.CheckoutUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse && w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "User checked out successfully", response["message"])
				assert.NotNil(t, response["session_id"])
				assert.NotNil(t, response["user_id"])
			}
		})
	}

	// Verify the session was actually checked out
	sess, err := sessionRepo.FindByID(sessionID)
	require.NoError(t, err)
	assert.NotNil(t, sess.CheckOut)
}

func TestCheckoutAll(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	// Seed multiple users with active sessions
	user1 := helpers.SeedUser(t, db, "User1", "UID1", "")
	user2 := helpers.SeedUser(t, db, "User2", "UID2", "")
	user3 := helpers.SeedUser(t, db, "User3", "UID3", "")

	helpers.SeedSession(t, db, user1.ID)
	helpers.SeedSession(t, db, user2.ID)
	sessionID3 := helpers.SeedSession(t, db, user3.ID)

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/checkout-all", nil)
	w := httptest.NewRecorder()

	handler.CheckoutAll(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, float64(3), response["count"])
	assert.Len(t, response["sessions"], 3)

	// Verify all sessions were checked out
	sess, err := sessionRepo.FindByID(sessionID3)
	require.NoError(t, err)
	assert.NotNil(t, sess.CheckOut)
}

func TestCheckoutAllEmpty(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/checkout-all", nil)
	w := httptest.NewRecorder()

	handler.CheckoutAll(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, float64(0), response["count"])
	assert.Equal(t, "No active sessions to check out", response["message"])
}

// Edge case tests for session handlers

func TestSessionHandler_CheckInUser_AlreadyCheckedIn(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	user := helpers.SeedUser(t, db, "DoubleCheckIn", "RFID800", "discord_double")
	helpers.SeedSession(t, db, user.ID)

	reqBody := dto.UserSessionActionRequest{UserID: user.ID}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/checkin", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CheckInUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "already checked in")
}

func TestSessionHandler_CheckOutUser_NoActiveSession(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	user := helpers.SeedUser(t, db, "NoSession", "RFID801", "discord_nosession")

	reqBody := dto.UserSessionActionRequest{UserID: user.ID}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/checkout", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CheckOutUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "no open session")
}

func TestSessionHandler_UpdateSession_InvalidJSON(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/sessions/1", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.UpdateSession(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_UpdateSession_NoFields(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	reqBody := dto.UpdateSessionRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/sessions/1", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.UpdateSession(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_ListSessions_InvalidFilters(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	tests := []struct {
		name   string
		params string
	}{
		{"invalid from date", "?from=not-a-date"},
		{"invalid to date", "?to=invalid"},
		{"invalid user_id", "?user_id=abc"},
		{"invalid status", "?status=stale"},
		{"invalid active_only", "?active_only=maybe"},
		{"invalid limit", "?limit=xyz"},
		{"negative limit", "?limit=-5"},
		{"invalid offset", "?offset=abc"},
		{"invalid order", "?order=random"},
		{"invalid sort_by", "?sort_by=invalid_field"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/sessions"+tt.params, nil)
			w := httptest.NewRecorder()
			handler.ListSessions(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestSessionHandler_ExportCSV_EmptyResult(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/export", nil)
	w := httptest.NewRecorder()

	handler.ExportSessionsCSV(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "UserName,CheckIn,CheckOut,Duration(minutes)")
}

func TestSessionHandler_DeleteSession_InvalidID(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/invalid", nil)
	w := httptest.NewRecorder()

	handler.DeleteSession(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_CheckInCheckOut_InvalidJSON(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/checkin", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	handler.CheckInUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/sessions/checkout", bytes.NewReader([]byte("not json")))
	w = httptest.NewRecorder()
	handler.CheckOutUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_CheckInCheckOut_MissingUserID(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	reqBody := dto.UserSessionActionRequest{UserID: 0}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/checkin", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.CheckInUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/sessions/checkout", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.CheckOutUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_UpdateSession_InvalidSessionID(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	handler := httptransport.NewSessionHandler(sessionSvc)

	reqBody := dto.UpdateSessionRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/sessions/abc", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.UpdateSession(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
