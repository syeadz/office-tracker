package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"office/internal/domain"
	"office/internal/repository"
	"office/internal/service"
	httptransport "office/internal/transport/http"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandler_ExposesPresenceAndFreshEnvironment(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	envSvc := service.NewEnvironmentService(5 * time.Minute)

	user, err := userRepo.Create("MetricsUser", "RFIDM001", "discord_metrics")
	require.NoError(t, err)
	err = sessionSvc.CheckInUser(user.ID)
	require.NoError(t, err)

	envSvc.Update(domain.EnvironmentReading{TemperatureC: 24.7, Timestamp: time.Now().Add(-1 * time.Minute)})

	h := httptransport.NewMetricsHandler(sessionSvc, envSvc, nil)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "office_presence_active_users")
	assert.Contains(t, body, "office_environment_temperature_celsius")
	assert.NotContains(t, body, "office_environment_humidity_percent")
	assert.NotContains(t, body, "office_environment_pressure_hpa")
}

func TestMetricsHandler_DoesNotExposeStaleEnvironment(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	envSvc := service.NewEnvironmentService(5 * time.Minute)
	envSvc.Update(domain.EnvironmentReading{TemperatureC: 28.1, Timestamp: time.Now().Add(-10 * time.Minute)})

	h := httptransport.NewMetricsHandler(sessionSvc, envSvc, nil)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "office_presence_active_users")
	assert.NotContains(t, body, "office_environment_temperature_celsius")
	assert.NotContains(t, body, "office_environment_humidity_percent")
	assert.NotContains(t, body, "office_environment_pressure_hpa")
}

func TestMetricsEndpoint_ExposedByServer(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	attendanceSvc := service.NewAttendanceService(userRepo, sessionRepo)
	userSvc := &service.UserService{Users: userRepo}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}
	envSvc := service.NewEnvironmentService(5 * time.Minute)

	srv := httptransport.New("0", attendanceSvc, userSvc, sessionSvc, statsSvc, envSvc, nil, httptransport.MiddlewareConfig{})
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	body := string(bodyBytes)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, strings.Contains(body, "office_presence_active_users"))
}

func TestMetricsHandler_ExposesESPHealthMetricsWhenFresh(t *testing.T) {
	espSvc := service.NewESPHealthService(20 * time.Minute)
	espSvc.Update(domain.ESPHealthStatus{
		DeviceID:      "esp-lab-1",
		UptimeSeconds: 600,
		FreeHeapBytes: 140000,
		WiFiConnected: true,
		RSSI:          -58,
	})

	h := httptransport.NewMetricsHandler(nil, nil, espSvc)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "office_esp_health_up")
	assert.Contains(t, body, "device_id=\"esp-lab-1\"")
	assert.Contains(t, body, "office_esp_health_up{device_id=\"esp-lab-1\"} 1")
	assert.Contains(t, body, "office_esp_health_uptime_seconds")
	assert.Contains(t, body, "office_esp_health_free_heap_bytes")
	assert.Contains(t, body, "office_esp_health_rssi_dbm")
}

func TestMetricsHandler_StaleESPHealthOnlyEmitsUpZero(t *testing.T) {
	espSvc := service.NewESPHealthService(1 * time.Millisecond)
	espSvc.Update(domain.ESPHealthStatus{
		DeviceID:      "esp-lab-2",
		UptimeSeconds: 100,
		FreeHeapBytes: 130000,
		WiFiConnected: true,
		RSSI:          -70,
	})
	time.Sleep(10 * time.Millisecond)

	h := httptransport.NewMetricsHandler(nil, nil, espSvc)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "office_esp_health_up{device_id=\"esp-lab-2\"} 0")
	assert.NotContains(t, body, "office_esp_health_uptime_seconds{device_id=\"esp-lab-2\"}")
	assert.NotContains(t, body, "office_esp_health_free_heap_bytes{device_id=\"esp-lab-2\"}")
	assert.NotContains(t, body, "office_esp_health_rssi_dbm{device_id=\"esp-lab-2\"}")
}
