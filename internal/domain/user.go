package domain

import "time"

// User represents a user in the system.
// It contains the user's ID, name, RFID UID, Discord ID, and the time they were created.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	RFIDUID   string    `json:"rfid_uid"`
	DiscordID string    `json:"discord_id"`
	CreatedAt time.Time `json:"created_at"`
}
