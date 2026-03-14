package dto

import "time"

// UpdateUserRequest is the DTO for updating user information
type UpdateUserRequest struct {
	Name      string `json:"name"`
	RFIDUID   string `json:"rfid_uid"`
	DiscordID string `json:"discord_id"`
}

// UpdateSessionRequest is the DTO for updating session times
type UpdateSessionRequest struct {
	CheckIn  *time.Time `json:"check_in,omitempty"`
	CheckOut *time.Time `json:"check_out,omitempty"`
}

// UserSessionActionRequest is the DTO for check-in/check-out actions
type UserSessionActionRequest struct {
	UserID int64 `json:"user_id"`
}

// DeleteResult is the response for bulk delete operations
type DeleteResult struct {
	Deleted int64 `json:"deleted"`
}
