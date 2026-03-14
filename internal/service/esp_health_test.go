package service

import (
	"testing"
	"time"

	"office/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestESPHealthService_UpdateAndGetAll(t *testing.T) {
	now := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	svc := NewESPHealthService(20 * time.Minute)
	svc.nowFunc = func() time.Time { return now }

	svc.Update(domain.ESPHealthStatus{DeviceID: "esp-a", UptimeSeconds: 1200, FreeHeapBytes: 180000, WiFiConnected: true, RSSI: -60, FirmwareVersion: "1.0.0"})
	svc.Update(domain.ESPHealthStatus{DeviceID: "esp-b", UptimeSeconds: 800, FreeHeapBytes: 160000, WiFiConnected: false, RSSI: -90, ResetReason: "power_on"})

	items := svc.GetAll()
	require.Len(t, items, 2)
	assert.Equal(t, "esp-a", items[0].DeviceID)
	assert.Equal(t, "esp-b", items[1].DeviceID)
	assert.Equal(t, now, items[0].UpdatedAt)
}

func TestESPHealthService_DefaultDeviceID(t *testing.T) {
	now := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	svc := NewESPHealthService(20 * time.Minute)
	svc.nowFunc = func() time.Time { return now }

	svc.Update(domain.ESPHealthStatus{UptimeSeconds: 10, FreeHeapBytes: 1000})
	items := svc.GetAll()
	require.Len(t, items, 1)
	assert.Equal(t, "default", items[0].DeviceID)
}

func TestESPHealthService_Freshness(t *testing.T) {
	now := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	svc := NewESPHealthService(20 * time.Minute)
	svc.nowFunc = func() time.Time { return now }

	freshStatus := domain.ESPHealthStatus{UpdatedAt: now.Add(-10 * time.Minute)}
	staleStatus := domain.ESPHealthStatus{UpdatedAt: now.Add(-30 * time.Minute)}

	assert.True(t, svc.IsFresh(freshStatus))
	assert.False(t, svc.IsFresh(staleStatus))
}
