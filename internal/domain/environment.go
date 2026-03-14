package domain

import "time"

// EnvironmentReading contains the latest office temperature reading.
type EnvironmentReading struct {
	TemperatureC float64   `json:"temperature_c"`
	Timestamp    time.Time `json:"timestamp"`
}
