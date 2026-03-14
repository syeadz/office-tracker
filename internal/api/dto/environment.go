package dto

import "time"

// EnvironmentUpsertRequest contains the temperature payload sent by the ESP32.
type EnvironmentUpsertRequest struct {
	TemperatureC float64   `json:"temperature_c"`
	Timestamp    time.Time `json:"timestamp"`
}

// EnvironmentResponse contains the latest environmental reading and freshness metadata.
type EnvironmentResponse struct {
	Available    bool       `json:"available"`
	Fresh        bool       `json:"fresh"`
	TemperatureC float64    `json:"temperature_c"`
	Timestamp    *time.Time `json:"timestamp,omitempty"`
	AgeSeconds   int64      `json:"age_seconds"`
}
