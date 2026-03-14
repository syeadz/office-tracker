package repository_test

import (
	"database/sql"
	"testing"
	"time"

	"office/internal/query"
	"office/internal/repository"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
)

func TestSessionRepo_CheckIn(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("Frank", "RFID006", "discord_frank")

	err := sessionRepo.CheckIn(user.ID)
	assert.NoError(t, err)

	// Verify session was created
	sessionID, err := sessionRepo.GetOpenSession(user.ID)
	assert.NoError(t, err)
	assert.Greater(t, sessionID, int64(0))
}

func TestSessionRepo_GetOpenSession(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("Grace", "RFID007", "discord_grace")
	helpers.SeedSession(t, db, user.ID)

	sessionID, err := sessionRepo.GetOpenSession(user.ID)
	assert.NoError(t, err)
	assert.Greater(t, sessionID, int64(0))
}

func TestSessionRepo_GetOpenSession_NotFound(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	_, err := sessionRepo.GetOpenSession(999)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestSessionRepo_FindByID(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("TestUser", "RFID_TEST", "discord_test")
	sessionID := helpers.SeedSession(t, db, user.ID)

	sess, err := sessionRepo.FindByID(sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, sess)
	assert.Equal(t, sessionID, sess.ID)
	assert.Equal(t, user.ID, sess.UserID)
	assert.Equal(t, "TestUser", sess.UserName)
	assert.NotNil(t, sess.CheckIn)
	assert.Nil(t, sess.CheckOut)
}

func TestSessionRepo_FindByID_NotFound(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	sess, err := sessionRepo.FindByID(9999)
	assert.Error(t, err)
	assert.Nil(t, sess)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestSessionRepo_FindByID_WithCheckOut(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("CheckedOut", "RFID_CHECKOUT", "discord_checkout")
	sessionID := helpers.SeedSession(t, db, user.ID)

	// Check out the session
	err := sessionRepo.CheckOut(sessionID)
	assert.NoError(t, err)

	// Retrieve and verify
	sess, err := sessionRepo.FindByID(sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, sess)
	assert.Equal(t, sessionID, sess.ID)
	assert.NotNil(t, sess.CheckOut)
}

func TestSessionRepo_CheckOut(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("Henry", "RFID008", "discord_henry")
	sessionID := helpers.SeedSession(t, db, user.ID)

	err := sessionRepo.CheckOut(sessionID)
	assert.NoError(t, err)

	// Verify no open session exists
	_, err = sessionRepo.GetOpenSession(user.ID)
	assert.Error(t, err)
}

func TestSessionRepo_AllOpen(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user1, _ := userRepo.Create("Ivy", "RFID009", "discord_ivy")
	user2, _ := userRepo.Create("Jack", "RFID010", "discord_jack")

	helpers.SeedSession(t, db, user1.ID)
	helpers.SeedSession(t, db, user2.ID)

	filter := query.SessionFilter{ActiveOnly: true}
	sessions, err := sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
}

func TestSessionRepo_AllOpen_Empty(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	filter := query.SessionFilter{ActiveOnly: true}
	sessions, err := sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSessionRepo_AllOpen_IgnoresClosed(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("Kate", "RFID011", "discord_kate")
	sessionID := helpers.SeedSession(t, db, user.ID)

	// Check out the session
	sessionRepo.CheckOut(sessionID)

	// List with ActiveOnly should return empty
	filter := query.SessionFilter{ActiveOnly: true}
	sessions, err := sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSessionRepo_AllForUser(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("Leo", "RFID012", "discord_leo")

	// Create and check in
	sessionID := helpers.SeedSession(t, db, user.ID)

	filter := query.SessionFilter{UserID: &user.ID}
	sessions, err := sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, sessionID, sessions[0].ID)
	assert.Equal(t, user.ID, sessions[0].UserID)

	// Check out and verify it's still in the list
	sessionRepo.CheckOut(sessionID)
	sessions, err = sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.NotNil(t, sessions[0].CheckOut)
}

func TestSessionRepo_AllForUser_Multiple(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("Megan", "RFID013", "discord_megan")

	// Create multiple sessions
	sessionID1 := helpers.SeedSession(t, db, user.ID)
	sessionRepo.CheckOut(sessionID1)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	helpers.SeedSession(t, db, user.ID)

	filter := query.SessionFilter{UserID: &user.ID}
	sessions, err := sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
}

func TestSessionRepo_AllForUser_Empty(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userID := int64(999)
	filter := query.SessionFilter{UserID: &userID}
	sessions, err := sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSessionRepo_List_ActiveOnly(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user1, _ := userRepo.Create("ListActiveUser1", "RFID020", "discord_active1")
	user2, _ := userRepo.Create("ListActiveUser2", "RFID021", "discord_active2")

	helpers.SeedSession(t, db, user1.ID) // active
	sessionID := helpers.SeedSession(t, db, user2.ID)
	sessionRepo.CheckOut(sessionID) // closed

	filter := query.SessionFilter{ActiveOnly: true}
	sessions, err := sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, user1.ID, sessions[0].UserID)
}

func TestSessionRepo_List_DateRange(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("DateRangeUser", "RFID022", "discord_daterange")

	// Create session in the past
	helpers.SeedSession(t, db, user.ID)
	time.Sleep(10 * time.Millisecond)

	from := time.Now().Add(-1 * time.Minute)
	to := time.Now().Add(1 * time.Minute)

	filter := query.SessionFilter{From: &from, To: &to}
	sessions, err := sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(sessions), 1)
}

func TestSessionRepo_List_NameLike(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("SearchableUser", "RFID023", "discord_searchable")
	helpers.SeedSession(t, db, user.ID)

	name := "Searchable"
	filter := query.SessionFilter{NameLike: &name}
	sessions, err := sessionRepo.List(filter)
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
}

func TestSessionRepo_Update(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("UpdateTest", "RFID200", "discord_update")
	sessionID := helpers.SeedSession(t, db, user.ID)

	now := time.Now()
	err := sessionRepo.Update(sessionID, &now, nil)
	assert.NoError(t, err)
}

func TestSessionRepo_Delete(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("DeleteTest", "RFID201", "discord_delete")
	sessionID := helpers.SeedSession(t, db, user.ID)

	err := sessionRepo.Delete(sessionID)
	assert.NoError(t, err)

	sessions, err := sessionRepo.List(query.SessionFilter{})
	assert.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSessionRepo_DeleteWithFilter(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("BulkDeleteTest", "RFID202", "discord_bulk")

	oldTime := time.Now().AddDate(0, -1, 0)
	newTime := time.Now()

	db.Exec(`INSERT INTO sessions(user_id, check_in, check_out, check_out_method) VALUES (?, ?, ?, ?)`, user.ID, oldTime, oldTime.Add(time.Hour), repository.CheckOutMethodRFID)
	db.Exec(`INSERT INTO sessions(user_id, check_in) VALUES (?, ?)`, user.ID, newTime)

	filter := query.SessionFilter{To: &oldTime}
	count, err := sessionRepo.DeleteWithFilter(filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	remaining, err := sessionRepo.List(query.SessionFilter{})
	assert.NoError(t, err)
	assert.Len(t, remaining, 1)
}

func TestSessionRepo_GetUserStats(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("StatsUser", "RFID300", "discord_stats")

	// Create sessions with known durations
	now := time.Now()
	sessionID1 := helpers.SeedSession(t, db, user.ID)
	time.Sleep(1 * time.Second) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID1)

	sessionID2 := helpers.SeedSession(t, db, user.ID)
	time.Sleep(1 * time.Second) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID2)

	// Get stats
	stats, err := sessionRepo.GetUserStats(user.ID, now.Add(-1*time.Hour), now.Add(1*time.Hour), false)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, user.ID, stats.UserID)
	assert.Equal(t, "StatsUser", stats.Name)
	assert.Equal(t, int64(2), stats.VisitCount)
	assert.Greater(t, stats.TotalHours, 0.0)
	assert.GreaterOrEqual(t, stats.ActiveDays, int64(1))
	assert.Greater(t, stats.AvgDuration, 0.0)
	assert.NotEmpty(t, stats.BusiestDay)
	assert.Greater(t, stats.BusiestDayHours, 0.0)
}

func TestSessionRepo_GetUserStats_NoSessions(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("NoSessionUser", "RFID301", "discord_nosession")

	now := time.Now()
	stats, err := sessionRepo.GetUserStats(user.ID, now.Add(-1*time.Hour), now.Add(1*time.Hour), false)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.VisitCount)
	assert.Equal(t, 0.0, stats.TotalHours)
	assert.Equal(t, "", stats.BusiestDay)
	assert.Equal(t, 0.0, stats.BusiestDayHours)
}

func TestSessionRepo_GetAllUserStats(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user1, _ := userRepo.Create("User1", "RFID302", "discord_user1")
	user2, _ := userRepo.Create("User2", "RFID303", "discord_user2")
	user3, _ := userRepo.Create("User3", "RFID304", "discord_user3")

	// Create sessions
	now := time.Now()
	for i := 0; i < 3; i++ {
		sessionID := helpers.SeedSession(t, db, user1.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	for i := 0; i < 2; i++ {
		sessionID := helpers.SeedSession(t, db, user2.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	sessionID := helpers.SeedSession(t, db, user3.ID)
	sessionRepo.CheckOut(sessionID)

	// Get all stats ordered by hours
	stats, err := sessionRepo.GetAllUserStats(now.Add(-1*time.Hour), now.Add(1*time.Hour), "hours", 10, false)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(stats), 3)

	// First user should have more hours
	assert.Greater(t, stats[0].VisitCount, stats[2].VisitCount)
}

func TestSessionRepo_GetAllUserStats_OrderByVisits(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user1, _ := userRepo.Create("VisitsUser1", "RFID305", "discord_v1")
	user2, _ := userRepo.Create("VisitsUser2", "RFID306", "discord_v2")

	now := time.Now()

	// User1: 5 visits
	for i := 0; i < 5; i++ {
		sessionID := helpers.SeedSession(t, db, user1.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	// User2: 2 visits
	for i := 0; i < 2; i++ {
		sessionID := helpers.SeedSession(t, db, user2.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	stats, err := sessionRepo.GetAllUserStats(now.Add(-1*time.Hour), now.Add(1*time.Hour), "visits", 10, false)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(stats), 2)

	// First should have 5 visits
	assert.Equal(t, int64(5), stats[0].VisitCount)
}

func TestSessionRepo_GetPeriodStats(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user1, _ := userRepo.Create("PeriodUser1", "RFID307", "discord_p1")
	user2, _ := userRepo.Create("PeriodUser2", "RFID308", "discord_p2")

	now := time.Now()
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	// Create sessions
	sessionID1 := helpers.SeedSession(t, db, user1.ID)
	time.Sleep(1 * time.Second) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID1)

	sessionID2 := helpers.SeedSession(t, db, user2.ID)
	time.Sleep(1 * time.Second) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID2)

	stats, err := sessionRepo.GetPeriodStats(from, to, 10, "hours", false)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "hours", stats.RankBy)
	assert.Equal(t, int64(2), stats.UniqueUsers)
	assert.Equal(t, int64(2), stats.TotalVisits)
	assert.Greater(t, stats.TotalHours, 0.0)
	assert.Greater(t, stats.AveragePerUser, 0.0)
	assert.Len(t, stats.TopUsers, 2)
	assert.NotEmpty(t, stats.BusiestDay)
	assert.GreaterOrEqual(t, stats.BusiestDayUsers, int64(1))
	assert.GreaterOrEqual(t, stats.PeakOccupancy, int64(1))
}

func TestSessionRepo_GetPeriodStatsRankedByVisits(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user1, _ := userRepo.Create("PeriodVisits1", "RFID309", "discord_period_v1")
	user2, _ := userRepo.Create("PeriodVisits2", "RFID310", "discord_period_v2")

	now := time.Now()
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	for i := 0; i < 3; i++ {
		sessionID := helpers.SeedSession(t, db, user1.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	sessionID := helpers.SeedSession(t, db, user2.ID)
	time.Sleep(1200 * time.Millisecond)
	sessionRepo.CheckOut(sessionID)

	stats, err := sessionRepo.GetPeriodStats(from, to, 10, "visits", false)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "visits", stats.RankBy)
	assert.GreaterOrEqual(t, len(stats.TopUsers), 2)
	assert.Equal(t, user1.ID, stats.TopUsers[0].UserID)
	assert.Equal(t, int64(3), stats.TopUsers[0].VisitCount)
}

func TestSessionRepo_GetPeriodStats_BusiestDayAndPeakOccupancy(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user1, _ := userRepo.Create("OfficePeak1", "RFID311", "discord_peak_1")
	user2, _ := userRepo.Create("OfficePeak2", "RFID312", "discord_peak_2")
	user3, _ := userRepo.Create("OfficePeak3", "RFID313", "discord_peak_3")

	day1Start := time.Date(2026, time.January, 10, 9, 0, 0, 0, time.UTC)
	day2Start := time.Date(2026, time.January, 11, 9, 0, 0, 0, time.UTC)

	_, err := db.Exec(`INSERT INTO sessions(user_id, check_in, check_out, check_out_method) VALUES (?, ?, ?, ?)`, user1.ID, day1Start, day1Start.Add(90*time.Minute), repository.CheckOutMethodRFID)
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO sessions(user_id, check_in, check_out, check_out_method) VALUES (?, ?, ?, ?)`, user2.ID, day1Start.Add(15*time.Minute), day1Start.Add(2*time.Hour), repository.CheckOutMethodRFID)
	assert.NoError(t, err)

	_, err = db.Exec(`INSERT INTO sessions(user_id, check_in, check_out, check_out_method) VALUES (?, ?, ?, ?)`, user1.ID, day2Start, day2Start.Add(time.Hour), repository.CheckOutMethodRFID)
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO sessions(user_id, check_in, check_out, check_out_method) VALUES (?, ?, ?, ?)`, user2.ID, day2Start.Add(5*time.Minute), day2Start.Add(65*time.Minute), repository.CheckOutMethodRFID)
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO sessions(user_id, check_in, check_out, check_out_method) VALUES (?, ?, ?, ?)`, user3.ID, day2Start.Add(10*time.Minute), day2Start.Add(70*time.Minute), repository.CheckOutMethodRFID)
	assert.NoError(t, err)

	from := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 11, 23, 59, 59, 0, time.UTC)

	stats, err := sessionRepo.GetPeriodStats(from, to, 10, "hours", false)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "2026-01-11", stats.BusiestDay)
	assert.Equal(t, int64(3), stats.BusiestDayUsers)
	assert.Equal(t, int64(3), stats.PeakOccupancy)
}

func TestSessionRepo_GetUserStats_BusiestDay(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	user, _ := userRepo.Create("BusiestDayUser", "RFID314", "discord_busiest_day")

	jan10 := time.Date(2026, time.January, 10, 9, 0, 0, 0, time.UTC)
	jan11 := time.Date(2026, time.January, 11, 9, 0, 0, 0, time.UTC)
	jan13 := time.Date(2026, time.January, 13, 9, 0, 0, 0, time.UTC)

	_, err := db.Exec(`INSERT INTO sessions(user_id, check_in, check_out, check_out_method) VALUES (?, ?, ?, ?)`, user.ID, jan10, jan10.Add(time.Hour), repository.CheckOutMethodRFID)
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO sessions(user_id, check_in, check_out, check_out_method) VALUES (?, ?, ?, ?)`, user.ID, jan11, jan11.Add(2*time.Hour), repository.CheckOutMethodRFID)
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO sessions(user_id, check_in, check_out, check_out_method) VALUES (?, ?, ?, ?)`, user.ID, jan13, jan13.Add(3*time.Hour), repository.CheckOutMethodRFID)
	assert.NoError(t, err)

	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 31, 23, 59, 59, 0, time.UTC)

	stats, err := sessionRepo.GetUserStats(user.ID, from, to, false)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "2026-01-13", stats.BusiestDay)
	assert.InEpsilon(t, 3.0, stats.BusiestDayHours, 0.001)
}

func TestSessionRepo_GetPeriodStats_DefaultsToRFIDOnlyCheckouts(t *testing.T) {
	sessionRepo, db := helpers.CreateSessionRepoForTest(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	rfidUser, _ := userRepo.Create("RFIDOnlyUser", "RFID500", "discord_rfid_only")
	discordUser, _ := userRepo.Create("DiscordCheckoutUser", "RFID501", "discord_checkout")

	now := time.Now()
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	rfidSessionID := helpers.SeedSession(t, db, rfidUser.ID)
	assert.NoError(t, sessionRepo.CheckOutWithMethod(rfidSessionID, repository.CheckOutMethodRFID))

	discordSessionID := helpers.SeedSession(t, db, discordUser.ID)
	assert.NoError(t, sessionRepo.CheckOutWithMethod(discordSessionID, repository.CheckOutMethodDiscord))

	rfidOnlyStats, err := sessionRepo.GetPeriodStats(from, to, 10, "hours", true)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rfidOnlyStats.TotalVisits)
	assert.Equal(t, int64(1), rfidOnlyStats.UniqueUsers)

	allStats, err := sessionRepo.GetPeriodStats(from, to, 10, "hours", false)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), allStats.TotalVisits)
	assert.Equal(t, int64(2), allStats.UniqueUsers)
}
