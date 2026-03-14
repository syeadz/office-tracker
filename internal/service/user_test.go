package service_test

import (
	"testing"

	"office/internal/query"
	"office/internal/repository"
	"office/internal/service"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	user, err := userSvc.CreateUser("Charlie", "RFID789", "discord3")
	assert.NoError(t, err)
	assert.Equal(t, "Charlie", user.Name)
}

func TestUserService_GetUserByID(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	seeded := helpers.SeedUser(t, db, "Dana", "RFID101", "discord4")

	user, err := userSvc.GetUserByID(seeded.ID)
	assert.NoError(t, err)
	assert.Equal(t, seeded.Name, user.Name)
}

func TestUserService_GetUserByDiscordID(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}
	seeded := helpers.SeedUser(t, db, "Frank", "RFID999", "discord_frank_123")

	user, err := userSvc.GetUserByDiscordID("discord_frank_123")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, seeded.Name, user.Name)
	assert.Equal(t, "discord_frank_123", user.DiscordID)
}

func TestUserService_GetUserByDiscordID_NotFound(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	user, err := userSvc.GetUserByDiscordID("nonexistent_discord_id")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserService_UpdateUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	user, _ := userSvc.CreateUser("Original", "RFID300", "original_discord")

	updated, err := userSvc.UpdateUser(user.ID, "Updated", "new_rfid", "new_discord")
	assert.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, "new_discord", updated.DiscordID)

	// Verify RFID was updated by fetching from repository
	domainUser, err := userRepo.FindByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "new_rfid", domainUser.RFIDUID)
}

func TestUserService_DeleteUser(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	user, _ := userSvc.CreateUser("ToDelete", "RFID301", "discord_delete")

	err := userSvc.DeleteUser(user.ID)
	assert.NoError(t, err)
}

func TestUserService_DeleteUsers(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	userSvc.CreateUser("Test_A", "RFID302", "discord1")
	userSvc.CreateUser("Test_B", "RFID303", "discord2")
	userSvc.CreateUser("Keep", "RFID304", "discord3")

	nameLike := "Test"
	count, err := userSvc.DeleteUsers(query.UserFilter{NameLike: &nameLike})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	_, err := userSvc.GetUserByID(99999)
	assert.Error(t, err)
}

func TestUserService_ListUsers_EmptyDatabase(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	users, err := userSvc.ListUsers(query.UserFilter{})
	assert.NoError(t, err)
	assert.Empty(t, users)
}

func TestUserService_ListUsers_WithFilters(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	userSvc.CreateUser("Alice", "RFID400", "discord_alice")
	userSvc.CreateUser("Bob", "RFID401", "discord_bob")
	userSvc.CreateUser("Charlie", "RFID402", "discord_charlie")

	// Test name search
	nameLike := "Ali"
	users, err := userSvc.ListUsers(query.UserFilter{NameLike: &nameLike})
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].Name)

	// Test limit
	users, err = userSvc.ListUsers(query.UserFilter{Limit: 2})
	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserService_ListUsersRaw(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	userSvc.CreateUser("RawUser", "RFID500", "discord_raw")

	users, err := userSvc.ListUsersRaw(query.UserFilter{})
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	// Verify it returns domain.User with RFID
	assert.Equal(t, "RFID500", users[0].RFIDUID)
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Users: userRepo}

	// Deleting non-existent user shouldn't cause crash
	// (SQLite doesn't error on DELETE with no matches)
	err := userSvc.DeleteUser(99999)
	assert.NoError(t, err)
}
