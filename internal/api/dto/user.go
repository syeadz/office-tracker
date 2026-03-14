package dto

// UserResponse is the DTO for user API responses
type UserResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	RFIDUID   string `json:"rfid_uid,omitempty"`
	DiscordID string `json:"discord_id"`
}

// CreateUserRequest is the DTO for creating a new user
type CreateUserRequest struct {
	Name      string `json:"name"`
	RFIDUID   string `json:"rfid_uid"`
	DiscordID string `json:"discord_id,omitempty"`
}
