package dto

import "time"

// ESPHealthUpsertRequest is the payload sent by an ESP32 heartbeat.
type ESPHealthUpsertRequest struct {
	DeviceID        string `json:"device_id,omitempty"`
	UptimeSeconds   int64  `json:"uptime_seconds"`
	FreeHeapBytes   int64  `json:"free_heap_bytes"`
	WiFiConnected   bool   `json:"wifi_connected"`
	RSSI            int    `json:"rssi"`
	IP              string `json:"ip,omitempty"`
	FirmwareVersion string `json:"firmware_version,omitempty"`
	ResetReason     string `json:"reset_reason,omitempty"`
}

// ESPHealthStatusResponse is the read model for a device health record.
type ESPHealthStatusResponse struct {
	DeviceID        string    `json:"device_id"`
	UptimeSeconds   int64     `json:"uptime_seconds"`
	FreeHeapBytes   int64     `json:"free_heap_bytes"`
	WiFiConnected   bool      `json:"wifi_connected"`
	RSSI            int       `json:"rssi"`
	IP              string    `json:"ip,omitempty"`
	FirmwareVersion string    `json:"firmware_version,omitempty"`
	ResetReason     string    `json:"reset_reason,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
	Fresh           bool      `json:"fresh"`
	AgeSeconds      int64     `json:"age_seconds"`
}
