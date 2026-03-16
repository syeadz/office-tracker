// Package query defines filter structs for querying the database with various criteria.
// These are used by repositories and services to implement flexible querying capabilities without exposing internal details of the database schema.
package query

import "time"

// SessionFilter allows filtering session queries by various criteria.
type SessionFilter struct {
	// identity
	UserID    *int64
	NameLike  *string
	DiscordID *string

	// time
	From *time.Time
	To   *time.Time

	// presence
	ActiveOnly     bool
	Status         string  // "active", "completed", or empty for all
	CheckOutMethod *string // "rfid", "discord", "api", "auto", or nil for all

	// pagination
	Limit  int
	Offset int

	// ordering
	OrderBy string // "asc" or "desc" (default: "desc")
	SortBy  string // "check_in", "check_out", or "user_name" (default: "check_in")
}
