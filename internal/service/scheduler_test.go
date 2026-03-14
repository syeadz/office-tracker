package service_test

import (
	"context"
	"testing"
	"time"

	"office/internal/query"
	"office/internal/repository"
	"office/internal/service"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
)

func TestSchedulerService_AutoCheckoutJob(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	// Create test users
	user1, _ := userRepo.Create("User1", "RFID1", "discord1")
	user2, _ := userRepo.Create("User2", "RFID2", "discord2")

	// Create open sessions
	_ = sessionRepo.CheckIn(user1.ID)
	_ = sessionRepo.CheckIn(user2.ID)

	// Verify sessions are open
	openFilter := query.SessionFilter{ActiveOnly: true}
	sessions, err := sessionRepo.List(openFilter)
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)

	// Run auto-checkout job manually
	scheduler := service.NewSchedulerService(sessionRepo)
	scheduler.AutoCheckoutJob()

	// Verify all sessions are now closed
	sessions, err = sessionRepo.List(openFilter)
	assert.NoError(t, err)
	assert.Len(t, sessions, 0)
}

func TestSchedulerService_AutoCheckoutJob_NoOpenSessions(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	// Create a user
	user, _ := userRepo.Create("User", "RFID", "discord")

	// Create and immediately checkout a session
	sessionRepo.CheckIn(user.ID)
	sessionID, _ := sessionRepo.GetOpenSession(user.ID)
	sessionRepo.CheckOut(sessionID)

	// Run auto-checkout job (should not error)
	scheduler := service.NewSchedulerService(sessionRepo)
	scheduler.AutoCheckoutJob()

	// Verify no open sessions
	openFilter := query.SessionFilter{ActiveOnly: true}
	sessions, err := sessionRepo.List(openFilter)
	assert.NoError(t, err)
	assert.Len(t, sessions, 0)
}

func TestSchedulerService_Start_Stop(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	scheduler := service.NewSchedulerService(sessionRepo)

	// Start scheduler
	err := scheduler.Start()
	assert.NoError(t, err)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	scheduler.Stop(ctx)

	// Should complete without error
}
