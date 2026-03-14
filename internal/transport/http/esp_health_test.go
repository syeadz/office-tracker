package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"office/internal/api/dto"
	"office/internal/service"
	httptransport "office/internal/transport/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestESPHealthHandler_UpsertAndList(t *testing.T) {
	healthSvc := service.NewESPHealthService(20 * time.Minute)
	handler := httptransport.NewESPHealthHandler(healthSvc)

	payload := dto.ESPHealthUpsertRequest{
		DeviceID:        "esp-lab-1",
		UptimeSeconds:   3600,
		FreeHeapBytes:   145000,
		WiFiConnected:   true,
		RSSI:            -58,
		IP:              "192.168.1.55",
		FirmwareVersion: "1.2.3",
		ResetReason:     "power_on",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	postReq := httptest.NewRequest(http.MethodPost, "/api/devices/health", bytes.NewReader(body))
	postRec := httptest.NewRecorder()
	handler.Upsert(postRec, postReq)
	assert.Equal(t, http.StatusOK, postRec.Code)

	getReq := httptest.NewRequest(http.MethodGet, "/api/devices/health", nil)
	getRec := httptest.NewRecorder()
	handler.List(getRec, getReq)
	assert.Equal(t, http.StatusOK, getRec.Code)

	var response []dto.ESPHealthStatusResponse
	err = json.NewDecoder(getRec.Body).Decode(&response)
	require.NoError(t, err)
	require.Len(t, response, 1)

	assert.Equal(t, payload.DeviceID, response[0].DeviceID)
	assert.Equal(t, payload.UptimeSeconds, response[0].UptimeSeconds)
	assert.Equal(t, payload.FreeHeapBytes, response[0].FreeHeapBytes)
	assert.Equal(t, payload.WiFiConnected, response[0].WiFiConnected)
	assert.Equal(t, payload.RSSI, response[0].RSSI)
	assert.Equal(t, payload.IP, response[0].IP)
	assert.Equal(t, payload.FirmwareVersion, response[0].FirmwareVersion)
	assert.Equal(t, payload.ResetReason, response[0].ResetReason)
	assert.True(t, response[0].Fresh)
}

func TestESPHealthHandler_RejectsNegativeValues(t *testing.T) {
	healthSvc := service.NewESPHealthService(20 * time.Minute)
	handler := httptransport.NewESPHealthHandler(healthSvc)

	payload := []byte(`{"uptime_seconds":-1,"free_heap_bytes":1000,"wifi_connected":true,"rssi":-70}`)
	req := httptest.NewRequest(http.MethodPost, "/api/devices/health", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	handler.Upsert(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
