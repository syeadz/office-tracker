package helpers

import (
	"database/sql"
	"testing"

	"office/internal/domain"
	"office/internal/repository"
)

// CreateUserRepoForTest creates a UserRepo with a fresh in-memory DB.
// Returns the repo and the DB (so tests can query or close it).
func CreateUserRepoForTest(t *testing.T) (*repository.UserRepo, *sql.DB) {
	t.Helper()
	db := SetupTestDB(t)
	return &repository.UserRepo{DB: db}, db
}

// CreateSessionRepoForTest creates a SessionRepo with a fresh in-memory DB.
// Returns the repo and the DB (so tests can query or close it).
func CreateSessionRepoForTest(t *testing.T) (*repository.SessionRepo, *sql.DB) {
	t.Helper()
	db := SetupTestDB(t)
	return &repository.SessionRepo{DB: db}, db
}

// CreateSessionRepoWithUserForTest creates a SessionRepo and UserRepo with a seeded user.
// Returns the session repo, user repo, the created user, and the DB (so tests can query or close it).
func CreateSessionRepoWithUserForTest(t *testing.T, name, rfidUID, discordID string) (*repository.SessionRepo, *repository.UserRepo, *domain.User, *sql.DB) {
	t.Helper()
	db := SetupTestDB(t)
	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user := SeedUser(t, db, name, rfidUID, discordID)
	return sessionRepo, userRepo, user, db
}
