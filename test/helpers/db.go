// Package helpers provides utility functions for setting up test databases and seeding data for unit tests.
package helpers

import (
	"database/sql"
	"testing"

	"office/internal/domain"
	"office/internal/repository"

	_ "modernc.org/sqlite"
)

// SetupTestDB creates a fresh in-memory SQLite DB with the required tables.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper() // marks this function as a helper so errors point to the calling test
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	queries := []string{
		`PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000; PRAGMA foreign_keys = ON;`,

		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			rfid_uid TEXT UNIQUE NOT NULL,
			discord_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			check_in DATETIME NOT NULL,
			check_out DATETIME,
			check_out_method TEXT,
			CHECK (check_out_method IS NULL OR check_out_method IN ('rfid', 'discord', 'api', 'auto')),
			CHECK ((check_out IS NULL AND check_out_method IS NULL) OR (check_out IS NOT NULL AND check_out_method IS NOT NULL)),
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_one_active_session ON sessions(user_id) WHERE check_out IS NULL;`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_checkout_method ON sessions(check_out_method);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}

	return db
}

// SeedUser inserts a user into the DB and returns the domain object.
func SeedUser(t *testing.T, db *sql.DB, name, rfidUID, discordID string) *domain.User {
	t.Helper()
	userRepo := &repository.UserRepo{DB: db}
	u, err := userRepo.Create(name, rfidUID, discordID)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// SeedSession inserts a checked-in session for a given user.
func SeedSession(t *testing.T, db *sql.DB, userID int64) int64 {
	t.Helper()
	sessionRepo := &repository.SessionRepo{DB: db}
	err := sessionRepo.CheckIn(userID)
	if err != nil {
		t.Fatal(err)
	}

	sessionID, err := sessionRepo.GetOpenSession(userID)
	if err != nil {
		t.Fatal(err)
	}
	return sessionID
}
