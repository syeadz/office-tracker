// Package dto defines Data Transfer Objects for API responses and requests.
// These structs are used to decouple the internal domain models from the external API representation,
// allowing for flexibility and security in how data is exposed through the API.
package dto

import "time"

// SessionResponse is the DTO for session API responses
type SessionResponse struct {
	ID       int64      `json:"id"`
	UserID   int64      `json:"user_id"`
	UserName string     `json:"user_name"`
	CheckIn  time.Time  `json:"check_in"`
	CheckOut *time.Time `json:"check_out,omitempty"`
	Active   bool       `json:"active"`
}

// PresenceResponse is the DTO for active presence API responses
// Excludes CheckOut since active sessions cannot have a checkout time
type PresenceResponse struct {
	UserName string    `json:"user_name"`
	CheckIn  time.Time `json:"check_in"`
}

// ScanResult is the DTO for RFID scan responses
type ScanResult struct {
	User   string `json:"user"`
	Action string `json:"action"`
}

// ScanRequest is the DTO for RFID scan requests
type ScanRequest struct {
	UID string `json:"uid"`
}

// ScanLog is the DTO for complete scan history entries
type ScanLog struct {
	UID       string    `json:"uid"`
	Timestamp time.Time `json:"timestamp"`
	UserName  string    `json:"user_name,omitempty"`
	Known     bool      `json:"known"`
	Action    string    `json:"action,omitempty"`
}
