package helpers

import (
	"database/sql"
	"testing"

	"office/internal/domain"
	"office/internal/repository"
	"office/internal/service"
)

// CreateAttendanceServiceForTest creates an AttendanceService wired to a fresh in-memory DB with optional seeded users.
// Returns the service and the DB (so tests can query or close it).
func CreateAttendanceServiceForTest(t *testing.T, seedUsers ...*domain.User) (*service.AttendanceService, *sql.DB) {
	t.Helper()
	db := SetupTestDB(t)

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	// Seed users if provided
	for _, u := range seedUsers {
		_, err := userRepo.Create(u.Name, u.RFIDUID, u.DiscordID)
		if err != nil {
			t.Fatal(err)
		}
	}

	attSvc := service.NewAttendanceService(userRepo, sessionRepo)
	return attSvc, db
}

// CreateSessionServiceForTest creates a SessionService with a fresh in-memory DB.
// Returns the service and the DB (so tests can query or close it).
func CreateSessionServiceForTest(t *testing.T, db *sql.DB) *service.SessionService {
	t.Helper()
	sessionRepo := &repository.SessionRepo{DB: db}
	return &service.SessionService{Sessions: sessionRepo}
}

// CreateUserServiceForTest creates a UserService with a fresh in-memory DB.
// Returns the service and the DB (so tests can query or close it).
func CreateUserServiceForTest(t *testing.T) (*service.UserService, *sql.DB) {
	t.Helper()
	db := SetupTestDB(t)
	userRepo := &repository.UserRepo{DB: db}
	return &service.UserService{Users: userRepo}, db
}
