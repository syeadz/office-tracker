package service_test

import (
	"testing"
	"time"

	"office/internal/api/dto"
	"office/internal/query"
	"office/internal/repository"
	"office/internal/service"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
)

func TestSessionService_GetSessionByID(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	// Create user and session
	user, err := userRepo.Create("GetSessionTest", "RFID999", "discord_getsession")
	assert.NoError(t, err)

	sessionID := helpers.SeedSession(t, db, user.ID)

	// Retrieve the session via service
	session, err := sessionSvc.GetSessionByID(sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, user.ID, session.UserID)
	assert.Equal(t, "GetSessionTest", session.UserName)
}

func TestSessionService_GetSessionByID_NotFound(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	// Try to get non-existent session
	session, err := sessionSvc.GetSessionByID(99999)
	assert.Error(t, err)
	assert.Nil(t, session)
}

func TestSessionService_UpdateSession(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	user, _ := userRepo.Create("UpdateSessionTest", "RFID400", "discord_update")
	sessionID := helpers.SeedSession(t, db, user.ID)

	checkIn := time.Now().Add(-2 * time.Hour)
	now := time.Now()
	err := sessionSvc.UpdateSession(sessionID, &checkIn, &now)
	assert.NoError(t, err)
}

func TestSessionService_DeleteSession(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	user, _ := userRepo.Create("DeleteSessionTest", "RFID401", "discord_delete")
	sessionID := helpers.SeedSession(t, db, user.ID)

	err := sessionSvc.DeleteSession(sessionID)
	assert.NoError(t, err)
}

func TestSessionService_DeleteSessions(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	user, _ := userRepo.Create("BulkDeleteSessionTest", "RFID402", "discord_bulk")
	sessionID1 := helpers.SeedSession(t, db, user.ID)
	sessionRepo.CheckOut(sessionID1)
	helpers.SeedSession(t, db, user.ID)

	filter := query.SessionFilter{}
	count, err := sessionSvc.DeleteSessions(filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestSessionService_CheckInUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	user, _ := userRepo.Create("CheckInTest", "RFID500", "discord_checkin")

	// First check-in should succeed
	err := sessionSvc.CheckInUser(user.ID)
	assert.NoError(t, err)

	// Second check-in should fail with ErrSessionAlreadyOpen
	err = sessionSvc.CheckInUser(user.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrSessionAlreadyOpen)
}

func TestSessionService_CheckOutUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	user, _ := userRepo.Create("CheckOutTest", "RFID501", "discord_checkout")

	// Check-out without check-in should fail
	err := sessionSvc.CheckOutUser(user.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrNoOpenSession)

	// Check in first
	err = sessionSvc.CheckInUser(user.ID)
	assert.NoError(t, err)

	// Now check-out should succeed
	err = sessionSvc.CheckOutUser(user.ID)
	assert.NoError(t, err)

	// Second check-out should fail
	err = sessionSvc.CheckOutUser(user.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrNoOpenSession)
}

func TestSessionService_ListSessions_AsDTO(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	user, _ := userRepo.Create("ListDTOTest", "RFID502", "discord_list")
	sessionID := helpers.SeedSession(t, db, user.ID)
	sessionRepo.CheckOut(sessionID)

	// Test with asDTO=true
	filter := query.SessionFilter{}
	result, err := sessionSvc.ListSessions(filter, true)
	assert.NoError(t, err)

	sessions := result.([]dto.SessionResponse)
	assert.Len(t, sessions, 1)
	assert.Equal(t, user.ID, sessions[0].UserID)
	assert.False(t, sessions[0].Active)
	assert.Equal(t, repository.CheckOutMethodRFID, sessions[0].CheckOutMethod)
}

func TestSessionService_ListSessions_ActiveOnly_AsDTO(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	user, _ := userRepo.Create("ActiveListTest", "RFID503", "discord_active")
	helpers.SeedSession(t, db, user.ID)

	// Test active sessions with DTO
	filter := query.SessionFilter{ActiveOnly: true}
	result, err := sessionSvc.ListSessions(filter, true)
	assert.NoError(t, err)

	sessions := result.([]dto.SessionResponse)
	assert.Len(t, sessions, 1)
	assert.Equal(t, user.Name, sessions[0].UserName)
	assert.NotZero(t, sessions[0].ID)
	assert.True(t, sessions[0].Active)
	assert.NotNil(t, sessions[0].CheckIn)
}
