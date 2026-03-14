//go:build integration
// +build integration

package http_integration_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"office/internal/app"
	"office/internal/repository"
	"office/internal/service"
	httptransport "office/internal/transport/http"
	"office/test/helpers"
)

func startIntegrationServer(t *testing.T, mw httptransport.MiddlewareConfig) (*httptest.Server, func()) {
	t.Helper()

	db := helpers.SetupTestDB(t)
	application := app.New(db, "0", mw)
	server := httptest.NewServer(application.HTTP.Handler())

	cleanup := func() {
		server.Close()
		_ = db.Close()
	}

	return server, cleanup
}

func startIntegrationServerWithReports(t *testing.T, mw httptransport.MiddlewareConfig, reportsEnabled bool) (*httptest.Server, func()) {
	t.Helper()

	db := helpers.SetupTestDB(t)

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	attendanceSvc := service.NewAttendanceService(userRepo, sessionRepo)
	userSvc := &service.UserService{Users: userRepo}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	reportsSvc := service.NewReportsService(statsSvc, nil, reportsEnabled)

	httpServer := httptransport.New("0", attendanceSvc, userSvc, sessionSvc, statsSvc, service.NewEnvironmentService(service.DefaultEnvironmentMaxAge), reportsSvc, mw)
	server := httptest.NewServer(httpServer.Handler())

	cleanup := func() {
		server.Close()
		_ = db.Close()
	}

	return server, cleanup
}

func doJSONRequest(t *testing.T, client *http.Client, method, url string, payload any, headers map[string]string) *http.Response {
	t.Helper()

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func decodeJSONBody[T any](t *testing.T, resp *http.Response, out *T) {
	t.Helper()
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

func readResponseBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return body
}

func TestIntegration_UserSessionLifecycleAndLeaderboard(t *testing.T) {
	server, cleanup := startIntegrationServer(t, httptransport.MiddlewareConfig{})
	defer cleanup()

	client := server.Client()

	createResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/api/users",
		map[string]any{"name": "Alice Integration", "rfid_uid": "rfid-int-001", "discord_id": "alice-1"},
		nil,
	)
	if createResp.StatusCode != http.StatusCreated {
		defer createResp.Body.Close()
		body, _ := io.ReadAll(createResp.Body)
		t.Fatalf("expected 201 for create user, got %d: %s", createResp.StatusCode, string(body))
	}

	var createdUser struct {
		ID int64 `json:"id"`
	}
	decodeJSONBody(t, createResp, &createdUser)
	if createdUser.ID == 0 {
		t.Fatalf("expected created user id")
	}

	checkInResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/api/sessions/checkin",
		map[string]any{"user_id": createdUser.ID},
		nil,
	)
	if checkInResp.StatusCode != http.StatusOK {
		defer checkInResp.Body.Close()
		body, _ := io.ReadAll(checkInResp.Body)
		t.Fatalf("expected 200 for check-in, got %d: %s", checkInResp.StatusCode, string(body))
	}
	checkInResp.Body.Close()

	openResp := doJSONRequest(t, client, http.MethodGet, server.URL+"/api/sessions/open", nil, nil)
	if openResp.StatusCode != http.StatusOK {
		defer openResp.Body.Close()
		body, _ := io.ReadAll(openResp.Body)
		t.Fatalf("expected 200 for open sessions, got %d: %s", openResp.StatusCode, string(body))
	}

	var openSessions []map[string]any
	decodeJSONBody(t, openResp, &openSessions)
	if len(openSessions) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(openSessions))
	}

	presenceResp := doJSONRequest(t, client, http.MethodGet, server.URL+"/api/presence", nil, nil)
	if presenceResp.StatusCode != http.StatusOK {
		defer presenceResp.Body.Close()
		body, _ := io.ReadAll(presenceResp.Body)
		t.Fatalf("expected 200 for presence, got %d: %s", presenceResp.StatusCode, string(body))
	}

	var presence []map[string]any
	decodeJSONBody(t, presenceResp, &presence)
	if len(presence) != 1 {
		t.Fatalf("expected 1 active presence record, got %d", len(presence))
	}

	checkOutResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/api/sessions/checkout",
		map[string]any{"user_id": createdUser.ID},
		nil,
	)
	if checkOutResp.StatusCode != http.StatusOK {
		defer checkOutResp.Body.Close()
		body, _ := io.ReadAll(checkOutResp.Body)
		t.Fatalf("expected 200 for check-out, got %d: %s", checkOutResp.StatusCode, string(body))
	}
	checkOutResp.Body.Close()

	openAfterCheckoutResp := doJSONRequest(t, client, http.MethodGet, server.URL+"/api/sessions/open", nil, nil)
	if openAfterCheckoutResp.StatusCode != http.StatusOK {
		defer openAfterCheckoutResp.Body.Close()
		body, _ := io.ReadAll(openAfterCheckoutResp.Body)
		t.Fatalf("expected 200 for open sessions after checkout, got %d: %s", openAfterCheckoutResp.StatusCode, string(body))
	}

	var openAfterCheckout []map[string]any
	decodeJSONBody(t, openAfterCheckoutResp, &openAfterCheckout)
	if len(openAfterCheckout) != 0 {
		t.Fatalf("expected 0 active sessions after checkout, got %d", len(openAfterCheckout))
	}

	leaderboardResp := doJSONRequest(t, client, http.MethodGet, server.URL+"/api/statistics/leaderboard?period=weekly&rank_by=visits", nil, nil)
	if leaderboardResp.StatusCode != http.StatusOK {
		defer leaderboardResp.Body.Close()
		body, _ := io.ReadAll(leaderboardResp.Body)
		t.Fatalf("expected 200 for leaderboard, got %d: %s", leaderboardResp.StatusCode, string(body))
	}

	var leaderboard struct {
		RankBy string `json:"rank_by"`
		Users  []struct {
			UserID int64 `json:"user_id"`
		} `json:"users"`
	}
	decodeJSONBody(t, leaderboardResp, &leaderboard)

	if leaderboard.RankBy != "visits" {
		t.Fatalf("expected rank_by=visits, got %q", leaderboard.RankBy)
	}
	if len(leaderboard.Users) == 0 {
		t.Fatalf("expected leaderboard users to include completed session")
	}
	if leaderboard.Users[0].UserID != createdUser.ID {
		t.Fatalf("expected leaderboard user_id %d, got %d", createdUser.ID, leaderboard.Users[0].UserID)
	}
}

func TestIntegration_APIKeyMiddleware(t *testing.T) {
	const apiKey = "integration-secret-key"

	server, cleanup := startIntegrationServer(t, httptransport.MiddlewareConfig{
		APIKey:        apiKey,
		APIKeyEnabled: true,
	})
	defer cleanup()

	client := server.Client()

	healthResp := doJSONRequest(t, client, http.MethodGet, server.URL+"/health", nil, nil)
	if healthResp.StatusCode != http.StatusOK {
		defer healthResp.Body.Close()
		body, _ := io.ReadAll(healthResp.Body)
		t.Fatalf("expected 200 for health without API key, got %d: %s", healthResp.StatusCode, string(body))
	}
	healthResp.Body.Close()

	unauthorizedResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/api/users",
		map[string]any{"name": "Unauthorized User", "rfid_uid": "rfid-unauth-1", "discord_id": "none"},
		nil,
	)
	if unauthorizedResp.StatusCode != http.StatusUnauthorized {
		defer unauthorizedResp.Body.Close()
		body, _ := io.ReadAll(unauthorizedResp.Body)
		t.Fatalf("expected 401 without API key, got %d: %s", unauthorizedResp.StatusCode, string(body))
	}
	unauthorizedResp.Body.Close()

	authorizedResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/api/users",
		map[string]any{"name": "Authorized User", "rfid_uid": "rfid-auth-1", "discord_id": "auth-1"},
		map[string]string{"X-API-Key": apiKey},
	)
	if authorizedResp.StatusCode != http.StatusCreated {
		defer authorizedResp.Body.Close()
		body, _ := io.ReadAll(authorizedResp.Body)
		t.Fatalf("expected 201 with API key, got %d: %s", authorizedResp.StatusCode, string(body))
	}
	authorizedResp.Body.Close()

	bearerResp := doJSONRequest(
		t,
		client,
		http.MethodGet,
		server.URL+"/api/users",
		nil,
		map[string]string{"Authorization": "Bearer " + apiKey},
	)
	if bearerResp.StatusCode != http.StatusOK {
		defer bearerResp.Body.Close()
		body, _ := io.ReadAll(bearerResp.Body)
		t.Fatalf("expected 200 with bearer API key, got %d: %s", bearerResp.StatusCode, string(body))
	}

	var users []map[string]any
	decodeJSONBody(t, bearerResp, &users)
	if len(users) != 1 {
		t.Fatalf("expected 1 user from authorized list request, got %d", len(users))
	}
}

func TestIntegration_UsersCSVImportExportRoundTrip(t *testing.T) {
	server, cleanup := startIntegrationServer(t, httptransport.MiddlewareConfig{})
	defer cleanup()

	client := server.Client()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileWriter, err := writer.CreateFormFile("file", "users.csv")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	_, err = fileWriter.Write([]byte("Name,RFID_UID,DiscordID\nAlice CSV,rfid-csv-1,alice-csv\nBob CSV,rfid-csv-2,bob-csv\n"))
	if err != nil {
		t.Fatalf("failed to write CSV content: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/users/import", &body)
	if err != nil {
		t.Fatalf("failed to build import request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	importResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("import request failed: %v", err)
	}
	if importResp.StatusCode != http.StatusOK {
		body := readResponseBody(t, importResp)
		t.Fatalf("expected 200 for import, got %d: %s", importResp.StatusCode, string(body))
	}

	var importResult struct {
		Imported int `json:"imported"`
		Failed   int `json:"failed"`
	}
	decodeJSONBody(t, importResp, &importResult)
	if importResult.Imported != 2 || importResult.Failed != 0 {
		t.Fatalf("expected import result {imported:2, failed:0}, got %+v", importResult)
	}

	exportResp := doJSONRequest(t, client, http.MethodGet, server.URL+"/api/users/export", nil, nil)
	if exportResp.StatusCode != http.StatusOK {
		body := readResponseBody(t, exportResp)
		t.Fatalf("expected 200 for export, got %d: %s", exportResp.StatusCode, string(body))
	}

	data := readResponseBody(t, exportResp)
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse exported CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 CSV rows (header + 2 users), got %d", len(records))
	}

	if records[0][0] != "Name" || records[0][1] != "RFID_UID" {
		t.Fatalf("unexpected CSV header: %v", records[0])
	}
}

func TestIntegration_ReportsToggleAndStatus(t *testing.T) {
	server, cleanup := startIntegrationServerWithReports(t, httptransport.MiddlewareConfig{}, true)
	defer cleanup()

	client := server.Client()

	statusResp := doJSONRequest(t, client, http.MethodGet, server.URL+"/api/reports/status", nil, nil)
	if statusResp.StatusCode != http.StatusOK {
		body := readResponseBody(t, statusResp)
		t.Fatalf("expected 200 for reports status, got %d: %s", statusResp.StatusCode, string(body))
	}

	var statusPayload struct {
		Enabled bool   `json:"enabled"`
		Status  string `json:"status"`
	}
	decodeJSONBody(t, statusResp, &statusPayload)
	if !statusPayload.Enabled || statusPayload.Status != "enabled" {
		t.Fatalf("expected reports initially enabled, got %+v", statusPayload)
	}

	toggleOffResp := doJSONRequest(t, client, http.MethodPost, server.URL+"/api/reports/toggle?enabled=false", nil, nil)
	if toggleOffResp.StatusCode != http.StatusOK {
		body := readResponseBody(t, toggleOffResp)
		t.Fatalf("expected 200 for reports toggle off, got %d: %s", toggleOffResp.StatusCode, string(body))
	}
	toggleOffResp.Body.Close()

	statusAfterOffResp := doJSONRequest(t, client, http.MethodGet, server.URL+"/api/reports/status", nil, nil)
	if statusAfterOffResp.StatusCode != http.StatusOK {
		body := readResponseBody(t, statusAfterOffResp)
		t.Fatalf("expected 200 for reports status after toggle off, got %d: %s", statusAfterOffResp.StatusCode, string(body))
	}
	decodeJSONBody(t, statusAfterOffResp, &statusPayload)
	if statusPayload.Enabled || statusPayload.Status != "disabled" {
		t.Fatalf("expected reports disabled after toggle off, got %+v", statusPayload)
	}

	toggleInvalidResp := doJSONRequest(t, client, http.MethodPost, server.URL+"/api/reports/toggle?enabled=maybe", nil, nil)
	if toggleInvalidResp.StatusCode != http.StatusBadRequest {
		body := readResponseBody(t, toggleInvalidResp)
		t.Fatalf("expected 400 for invalid toggle value, got %d: %s", toggleInvalidResp.StatusCode, string(body))
	}
	toggleInvalidResp.Body.Close()
}

func TestIntegration_CORSAndAPIKeyMiddlewareTogether(t *testing.T) {
	const (
		apiKey = "integration-cors-api-key"
		origin = "http://localhost:3000"
	)

	server, cleanup := startIntegrationServer(t, httptransport.MiddlewareConfig{
		APIKey:        apiKey,
		APIKeyEnabled: true,
		CORSEnabled:   true,
		CORSOrigins:   origin,
	})
	defer cleanup()

	client := server.Client()

	preflightReq, err := http.NewRequest(http.MethodOptions, server.URL+"/api/users", nil)
	if err != nil {
		t.Fatalf("failed to build preflight request: %v", err)
	}
	preflightReq.Header.Set("Origin", origin)
	preflightReq.Header.Set("Access-Control-Request-Method", "GET")

	preflightResp, err := client.Do(preflightReq)
	if err != nil {
		t.Fatalf("preflight request failed: %v", err)
	}
	if preflightResp.StatusCode != http.StatusNoContent {
		body := readResponseBody(t, preflightResp)
		t.Fatalf("expected 204 for preflight, got %d: %s", preflightResp.StatusCode, string(body))
	}
	if got := preflightResp.Header.Get("Access-Control-Allow-Origin"); got != origin {
		preflightResp.Body.Close()
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", origin, got)
	}
	preflightResp.Body.Close()

	unauthResp := doJSONRequest(
		t,
		client,
		http.MethodGet,
		server.URL+"/api/users",
		nil,
		map[string]string{"Origin": origin},
	)
	if unauthResp.StatusCode != http.StatusUnauthorized {
		body := readResponseBody(t, unauthResp)
		t.Fatalf("expected 401 for unauthorized request, got %d: %s", unauthResp.StatusCode, string(body))
	}
	if got := unauthResp.Header.Get("Access-Control-Allow-Origin"); got != origin {
		unauthResp.Body.Close()
		t.Fatalf("expected CORS header on unauthorized response, got %q", got)
	}
	unauthResp.Body.Close()

	authResp := doJSONRequest(
		t,
		client,
		http.MethodGet,
		server.URL+"/api/users",
		nil,
		map[string]string{"Origin": origin, "X-API-Key": apiKey},
	)
	if authResp.StatusCode != http.StatusOK {
		body := readResponseBody(t, authResp)
		t.Fatalf("expected 200 for authorized request, got %d: %s", authResp.StatusCode, string(body))
	}
	if got := authResp.Header.Get("Access-Control-Allow-Origin"); got != origin {
		authResp.Body.Close()
		t.Fatalf("expected CORS header on authorized response, got %q", got)
	}
	authResp.Body.Close()
}
