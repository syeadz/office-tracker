package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"office/internal/repository"
	"office/internal/service"
	httptransport "office/internal/transport/http"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsHandler_GetLeaderboard(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user1, _ := userRepo.Create("Leader1", "RFID501", "discord_leader1")
	user2, _ := userRepo.Create("Leader2", "RFID502", "discord_leader2")

	// Create sessions for user1
	for i := 0; i < 3; i++ {
		sessionID := helpers.SeedSession(t, db, user1.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	// Create session for user2
	sessionID := helpers.SeedSession(t, db, user2.ID)
	sessionRepo.CheckOut(sessionID)

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/leaderboard?rank_by=hours&limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetLeaderboard(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "hours", response["rank_by"])
	assert.Greater(t, response["count"], float64(1))
	assert.NotNil(t, response["users"])
}

func TestStatsHandler_GetLeaderboard_ByVisits(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	now := time.Now()
	from := now.Format("2006-01-02")
	to := now.Format("2006-01-02")

	user1, _ := userRepo.Create("Visits1", "RFID503", "discord_v1")
	user2, _ := userRepo.Create("Visits2", "RFID504", "discord_v2")

	// User1: 5 visits
	for i := 0; i < 5; i++ {
		sessionID := helpers.SeedSession(t, db, user1.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	// User2: 2 visits
	for i := 0; i < 2; i++ {
		sessionID := helpers.SeedSession(t, db, user2.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/leaderboard?rank_by=visits&from="+from+"&to="+to, nil)
	w := httptest.NewRecorder()

	handler.GetLeaderboard(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "visits", response["rank_by"])
}

func TestStatsHandler_GetLeaderboard_InvalidDate(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/leaderboard?from=invalid", nil)
	w := httptest.NewRecorder()

	handler.GetLeaderboard(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStatsHandler_GetWeeklyReport(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user, _ := userRepo.Create("WeeklyUser", "RFID505", "discord_weekly")

	sessionID := helpers.SeedSession(t, db, user.ID)
	sessionRepo.CheckOut(sessionID)

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/weekly", nil)
	w := httptest.NewRecorder()

	handler.GetWeeklyReport(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "week", response["period_type"])
}

func TestStatsHandler_GetMonthlyReport(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user, _ := userRepo.Create("MonthlyUser", "RFID506", "discord_monthly")

	sessionID := helpers.SeedSession(t, db, user.ID)
	sessionRepo.CheckOut(sessionID)

	now := time.Now()
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/monthly", nil)
	w := httptest.NewRecorder()

	handler.GetMonthlyReport(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "month", response["period_type"])
	assert.Contains(t, response["period"], now.Format("2006-01"))
}

func TestStatsHandler_GetMonthlyReport_InvalidMonth(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/monthly?month=13", nil)
	w := httptest.NewRecorder()

	handler.GetMonthlyReport(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStatsHandler_GetCustomReport(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user, _ := userRepo.Create("CustomUser", "RFID507", "discord_custom")

	sessionID := helpers.SeedSession(t, db, user.ID)
	sessionRepo.CheckOut(sessionID)

	now := time.Now()
	from := now.Add(-7 * 24 * time.Hour).Format("2006-01-02")
	to := now.Format("2006-01-02")

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/report?from="+from+"&to="+to, nil)
	w := httptest.NewRecorder()

	handler.GetCustomReport(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "custom", response["period_type"])
	assert.Equal(t, "hours", response["rank_by"])
}

func TestStatsHandler_GetCustomReport_RankByVisits(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user1, _ := userRepo.Create("VisitTop", "RFID509", "discord_visit_top")
	user2, _ := userRepo.Create("HourTop", "RFID510", "discord_hour_top")

	for i := 0; i < 4; i++ {
		sessionID := helpers.SeedSession(t, db, user1.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	sessionID := helpers.SeedSession(t, db, user2.ID)
	time.Sleep(1200 * time.Millisecond)
	sessionRepo.CheckOut(sessionID)

	now := time.Now()
	from := now.Add(-24 * time.Hour).Format("2006-01-02")
	to := now.Format("2006-01-02")

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/report?from="+from+"&to="+to+"&rank_by=visits", nil)
	w := httptest.NewRecorder()

	handler.GetCustomReport(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "visits", response["rank_by"])
	topUsers, ok := response["top_users"].([]interface{})
	require.True(t, ok)
	require.NotEmpty(t, topUsers)
	firstUser, ok := topUsers[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "VisitTop", firstUser["name"])
	assert.Equal(t, float64(4), firstUser["visit_count"])
}

func TestStatsHandler_GetCustomReport_MissingDates(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/report", nil)
	w := httptest.NewRecorder()

	handler.GetCustomReport(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStatsHandler_GetUserStats(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user, _ := userRepo.Create("UserStatsTest", "RFID508", "discord_ustats")

	sessionID := helpers.SeedSession(t, db, user.ID)
	sessionRepo.CheckOut(sessionID)

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/users/"+strconv.FormatInt(user.ID, 10)+"?include_auto_checkout=true", nil)
	w := httptest.NewRecorder()

	handler.GetUserStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, float64(user.ID), response["user_id"])
	assert.NotNil(t, response["total_hours"])
	assert.Equal(t, float64(1), response["visit_count"])
}

func TestStatsHandler_GetUserStats_InvalidID(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/users/invalid", nil)
	w := httptest.NewRecorder()

	handler.GetUserStats(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Edge case tests for stats handlers

func TestStatsHandler_GetLeaderboard_InvalidParams(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	tests := []struct {
		name   string
		params string
	}{
		{"invalid from", "?from=bad-date"},
		{"invalid to", "?to=not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/statistics/leaderboard"+tt.params, nil)
			w := httptest.NewRecorder()
			handler.GetLeaderboard(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestStatsHandler_GetCustomReport_InvalidDates(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	tests := []struct {
		name   string
		params string
	}{
		{"missing from", "?to=2024-01-01T00:00:00Z"},
		{"missing to", "?from=2024-01-01T00:00:00Z"},
		{"invalid from format", "?from=bad-date&to=2024-01-01T00:00:00Z"},
		{"invalid to format", "?from=2024-01-01T00:00:00Z&to=bad-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/statistics/report"+tt.params, nil)
			w := httptest.NewRecorder()
			handler.GetCustomReport(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestStatsHandler_GetUserStats_NonExistentUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/users/99999", nil)
	w := httptest.NewRecorder()

	handler.GetUserStats(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStatsHandler_GetLeaderboard_DefaultValues(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/leaderboard", nil)
	w := httptest.NewRecorder()

	handler.GetLeaderboard(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStatsHandler_GetWeeklyReport_Success(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics/weekly", nil)
	w := httptest.NewRecorder()

	handler.GetWeeklyReport(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStatsHandler_GetMonthlyReport_InvalidParams(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	handler := httptransport.NewStatsHandler(statsSvc)

	tests := []struct {
		name   string
		params string
	}{
		{"invalid year", "?year=abc"},
		{"invalid month", "?month=13"},
		{"month out of range low", "?month=0"},
		{"missing year", "?month=5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/statistics/monthly"+tt.params, nil)
			w := httptest.NewRecorder()
			handler.GetMonthlyReport(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
