package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"office/internal/api/dto"
	"office/internal/domain"
	"office/internal/service"
	httptransport "office/internal/transport/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentHandler_UpdateLatestAndGetLatest(t *testing.T) {
	envSvc := service.NewEnvironmentService(5 * time.Minute)
	handler := httptransport.NewEnvironmentHandler(envSvc)

	timestamp := time.Now().UTC().Truncate(time.Second)
	payload := dto.EnvironmentUpsertRequest{
		TemperatureC: 25.3,
		Timestamp:    timestamp,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	postReq := httptest.NewRequest(http.MethodPost, "/api/environment", bytes.NewReader(body))
	postRec := httptest.NewRecorder()
	handler.UpdateLatest(postRec, postReq)

	assert.Equal(t, http.StatusOK, postRec.Code)

	getReq := httptest.NewRequest(http.MethodGet, "/api/environment", nil)
	getRec := httptest.NewRecorder()
	handler.GetLatest(getRec, getReq)

	assert.Equal(t, http.StatusOK, getRec.Code)

	var response dto.EnvironmentResponse
	err = json.NewDecoder(getRec.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Available)
	assert.True(t, response.Fresh)
	assert.InDelta(t, payload.TemperatureC, response.TemperatureC, 0.001)
	if assert.NotNil(t, response.Timestamp) {
		assert.Equal(t, timestamp, response.Timestamp.UTC())
	}
}

func TestEnvironmentHandler_GetLatestReturnsStaleMetadata(t *testing.T) {
	envSvc := service.NewEnvironmentService(5 * time.Minute)
	envSvc.Update(domain.EnvironmentReading{
		TemperatureC: 27.0,
		Timestamp:    time.Now().Add(-10 * time.Minute),
	})

	handler := httptransport.NewEnvironmentHandler(envSvc)
	req := httptest.NewRequest(http.MethodGet, "/api/environment", nil)
	rec := httptest.NewRecorder()
	handler.GetLatest(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.EnvironmentResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Available)
	assert.False(t, response.Fresh)
	assert.GreaterOrEqual(t, response.AgeSeconds, int64(600))
}

func TestEnvironmentHandler_GetLatestReturnsEmptyWhenNoReading(t *testing.T) {
	handler := httptransport.NewEnvironmentHandler(service.NewEnvironmentService(5 * time.Minute))
	req := httptest.NewRequest(http.MethodGet, "/api/environment", nil)
	rec := httptest.NewRecorder()
	handler.GetLatest(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.EnvironmentResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Available)
	assert.False(t, response.Fresh)
}
