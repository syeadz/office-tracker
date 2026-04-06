package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"office/internal/domain"
	"office/internal/service"
	httptransport "office/internal/transport/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportsHandler_StatusFullyDisabledWhenUnavailable(t *testing.T) {
	handler := httptransport.NewReportsHandlers(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/reports/status", nil)
	rec := httptest.NewRecorder()

	handler.HandleGetReportsStatus(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var payload map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&payload)
	require.NoError(t, err)

	assert.Equal(t, false, payload["available"])
	assert.Equal(t, true, payload["fully_disabled"])
	assert.Equal(t, false, payload["enabled"])
	assert.Equal(t, false, payload["weekly_enabled"])
	assert.Equal(t, false, payload["monthly_enabled"])
}

func TestReportsHandler_TogglePeriod(t *testing.T) {
	reportsSvc := service.NewReportsService(nil, &noopReportDelivery{}, true)
	handler := httptransport.NewReportsHandlers(reportsSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/reports/toggle?period=weekly&enabled=false", nil)
	rec := httptest.NewRecorder()

	handler.HandleToggleReports(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, reportsSvc.IsWeeklyEnabled())
	assert.True(t, reportsSvc.IsMonthlyEnabled())

	statusReq := httptest.NewRequest(http.MethodGet, "/api/reports/status", nil)
	statusRec := httptest.NewRecorder()
	handler.HandleGetReportsStatus(statusRec, statusReq)

	assert.Equal(t, http.StatusOK, statusRec.Code)

	var payload map[string]interface{}
	err := json.NewDecoder(statusRec.Body).Decode(&payload)
	require.NoError(t, err)

	assert.Equal(t, true, payload["available"])
	assert.Equal(t, false, payload["weekly_enabled"])
	assert.Equal(t, true, payload["monthly_enabled"])
}

type noopReportDelivery struct{}

func (n *noopReportDelivery) SendPeriodReport(_ *domain.PeriodReport, _ string) error {
	return nil
}
