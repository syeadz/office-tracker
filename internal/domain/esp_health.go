package domain

import "time"

// ESPHealthStatus contains the latest heartbeat details from an ESP32 device.
type ESPHealthStatus struct {
	DeviceID        string    `json:"device_id"`
	UptimeSeconds   int64     `json:"uptime_seconds"`
	FreeHeapBytes   int64     `json:"free_heap_bytes"`
	WiFiConnected   bool      `json:"wifi_connected"`
	RSSI            int       `json:"rssi"`
	IP              string    `json:"ip,omitempty"`
	FirmwareVersion string    `json:"firmware_version,omitempty"`
	ResetReason     string    `json:"reset_reason,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}
