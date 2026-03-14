// Package domain contains the business objects for the Office Tracker application.
package domain

import "time"

// Session represents a user's session in the system.
// It contains the session's ID, the user's ID, the check-in time, and an optional check-out time.
// If the user has not checked out, the CheckOut field will be nil.
type Session struct {
	ID       int64      `json:"id"`
	UserID   int64      `json:"user_id"`
	CheckIn  time.Time  `json:"check_in"`
	CheckOut *time.Time `json:"check_out,omitempty"`
}
