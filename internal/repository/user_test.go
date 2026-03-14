package repository_test

import (
	"testing"

	"office/internal/query"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
)

func TestUserRepo_Create(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	user, err := userRepo.Create("Alice", "RFID001", "discord_alice")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, "RFID001", user.RFIDUID)
	assert.Equal(t, "discord_alice", user.DiscordID)
	assert.Greater(t, user.ID, int64(0))
}

func TestUserRepo_FindByID(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	created, _ := userRepo.Create("Bob", "RFID002", "discord_bob")

	user, err := userRepo.FindByID(created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, "Bob", user.Name)
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	user, err := userRepo.FindByID(999)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepo_FindByRFID(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	userRepo.Create("Charlie", "RFID003", "discord_charlie")

	user, err := userRepo.FindByRFID("RFID003")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "Charlie", user.Name)
	assert.Equal(t, "RFID003", user.RFIDUID)
}

func TestUserRepo_FindByRFID_NotFound(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	user, err := userRepo.FindByRFID("UNKNOWN")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepo_FindByDiscordID(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	userRepo.Create("Discord User", "RFID_DISCORD", "discord_user_123")

	user, err := userRepo.FindByDiscordID("discord_user_123")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "Discord User", user.Name)
	assert.Equal(t, "discord_user_123", user.DiscordID)
}

func TestUserRepo_FindByDiscordID_NotFound(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	user, err := userRepo.FindByDiscordID("unknown_discord_id")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepo_All_Empty(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	filter := query.UserFilter{}
	users, err := userRepo.List(filter)
	assert.NoError(t, err)
	assert.Empty(t, users)
}

func TestUserRepo_Update(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	user, _ := userRepo.Create("Original", "RFID100", "original_discord")

	updated, err := userRepo.Update(user.ID, "Updated", "new_rfid", "new_discord")
	assert.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, "new_discord", updated.DiscordID)
	assert.Equal(t, "new_rfid", updated.RFIDUID)
}

func TestUserRepo_Delete(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	user, _ := userRepo.Create("ToDelete", "RFID101", "discord_delete")

	err := userRepo.Delete(user.ID)
	assert.NoError(t, err)

	_, err = userRepo.FindByID(user.ID)
	assert.Error(t, err)
}

func TestUserRepo_DeleteWithFilter(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	userRepo.Create("Test_User_A", "RFID102", "discord1")
	userRepo.Create("Test_User_B", "RFID103", "discord2")
	userRepo.Create("Other_User", "RFID104", "discord3")

	nameLike := "Test_User"
	filter := query.UserFilter{NameLike: &nameLike}
	count, err := userRepo.DeleteWithFilter(filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	remaining, err := userRepo.List(query.UserFilter{})
	assert.NoError(t, err)
	assert.Len(t, remaining, 1)
	assert.Equal(t, "Other_User", remaining[0].Name)
}

func TestUserRepo_Create_DuplicateRFID(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	// Create first user
	_, err := userRepo.Create("User1", "DUPLICATE_RFID", "discord1")
	assert.NoError(t, err)

	// Attempt to create user with duplicate RFID
	_, err = userRepo.Create("User2", "DUPLICATE_RFID", "discord2")
	assert.Error(t, err)
}

func TestUserRepo_List_WithPagination(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	// Create 5 users
	for i := 1; i <= 5; i++ {
		userRepo.Create("User"+string(rune('A'+i-1)), "RFID"+string(rune('0'+i)), "discord"+string(rune('0'+i)))
	}

	// Test limit
	filter := query.UserFilter{Limit: 2}
	users, err := userRepo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	// Test offset
	filter = query.UserFilter{Limit: 2, Offset: 2}
	users, err = userRepo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "UserC", users[0].Name)
}

func TestUserRepo_List_WithSorting(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	userRepo.Create("Zebra", "RFID1", "discord1")
	userRepo.Create("Alpha", "RFID2", "discord2")
	userRepo.Create("Beta", "RFID3", "discord3")

	// Test ascending order (default)
	filter := query.UserFilter{OrderBy: "asc"}
	users, err := userRepo.List(filter)
	assert.NoError(t, err)
	assert.Equal(t, "Alpha", users[0].Name)
	assert.Equal(t, "Zebra", users[len(users)-1].Name)

	// Test descending order
	filter = query.UserFilter{OrderBy: "desc"}
	users, err = userRepo.List(filter)
	assert.NoError(t, err)
	assert.Equal(t, "Zebra", users[0].Name)
	assert.Equal(t, "Alpha", users[len(users)-1].Name)
}

func TestUserRepo_Update_NonExistentUser(t *testing.T) {
	userRepo, db := helpers.CreateUserRepoForTest(t)
	defer db.Close()

	// Update operation on non-existent user should fail when trying to find it
	user, err := userRepo.Update(99999, "Updated", "RFID", "discord")
	assert.Error(t, err)
	assert.Nil(t, user)
}
