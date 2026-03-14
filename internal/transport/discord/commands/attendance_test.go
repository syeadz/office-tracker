package commands_test

import (
	"testing"

	"office/internal/query"
	"office/internal/repository"
	"office/internal/service"
	"office/internal/transport/discord/commands"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttendanceCommands_GetApplicationCommands(t *testing.T) {
	ac := commands.NewAttendanceCommands(nil, nil)
	cmds := ac.GetApplicationCommands()

	assert.Len(t, cmds, 3)
	assert.Equal(t, "checkin", cmds[0].Name)
	assert.Equal(t, "checkout", cmds[1].Name)
	assert.Equal(t, "checkout-all", cmds[2].Name)
}

func TestAttendanceCommands_NoNilHandling(t *testing.T) {
	// Verify the command handler can be created and doesn't panic
	ac := commands.NewAttendanceCommands(nil, nil)
	assert.NotNil(t, ac)

	// Verify GetApplicationCommands doesn't panic
	cmds := ac.GetApplicationCommands()
	assert.NotNil(t, cmds)
}

func TestAttendanceCommands_WithServices(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	ac := commands.NewAttendanceCommands(userSvc, sessionSvc)
	assert.NotNil(t, ac)

	// Verify commands can be retrieved
	cmds := ac.GetApplicationCommands()
	assert.Len(t, cmds, 3)

	// Verify command names
	names := []string{cmds[0].Name, cmds[1].Name, cmds[2].Name}
	assert.Contains(t, names, "checkin")
	assert.Contains(t, names, "checkout")
	assert.Contains(t, names, "checkout-all")
}

func TestAttendanceCommands_CheckoutAll_GetSessions(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	sessionSvc := &service.SessionService{Sessions: sessionRepo}

	// Create users with active sessions
	user1, _ := userRepo.Create("Alice", "RFID001", "discord1")
	user2, _ := userRepo.Create("Bob", "RFID002", "discord2")

	helpers.SeedSession(t, db, user1.ID)
	helpers.SeedSession(t, db, user2.ID)

	// Verify active sessions exist
	filter := query.SessionFilter{ActiveOnly: true}
	result, err := sessionSvc.ListSessions(filter, false)
	require.NoError(t, err)

	sessions := result.([]*repository.SessionWithUser)
	assert.Len(t, sessions, 2)
	// Verify both users are in the sessions
	userIDs := map[int64]bool{sessions[0].UserID: true, sessions[1].UserID: true}
	assert.True(t, userIDs[user1.ID])
	assert.True(t, userIDs[user2.ID])
}
