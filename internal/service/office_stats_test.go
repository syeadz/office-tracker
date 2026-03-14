package service_test

import (
	"testing"
	"time"

	"office/internal/repository"
	"office/internal/service"
	"office/test/helpers"

	"github.com/stretchr/testify/assert"
)

func TestOfficeStatsService_GetLeaderboard_ByHours(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	// Create users with different hours
	user1, _ := userRepo.Create("Leader1", "RFID401", "discord_leader1")
	user2, _ := userRepo.Create("Leader2", "RFID402", "discord_leader2")

	start := time.Now()

	// User1: 5 sessions
	for i := 0; i < 5; i++ {
		sessionID := helpers.SeedSession(t, db, user1.ID)
		time.Sleep(1 * time.Second) // Ensure measurable duration
		sessionRepo.CheckOut(sessionID)
	}

	// User2: 2 sessions
	for i := 0; i < 2; i++ {
		sessionID := helpers.SeedSession(t, db, user2.ID)
		time.Sleep(1 * time.Second) // Ensure measurable duration
		sessionRepo.CheckOut(sessionID)
	}

	end := time.Now()
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	leaderboard, err := statsSvc.GetLeaderboardWithAutoCheckout(start, end, "hours", 10, false)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(leaderboard), 2)

	// User1 should be first (more hours)
	assert.Greater(t, leaderboard[0].TotalHours, leaderboard[1].TotalHours)
}

func TestOfficeStatsService_GetLeaderboard_ByVisits(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user1, _ := userRepo.Create("Visitor1", "RFID403", "discord_v1")
	user2, _ := userRepo.Create("Visitor2", "RFID404", "discord_v2")

	start := time.Now()

	// User1: 7 visits
	for i := 0; i < 7; i++ {
		sessionID := helpers.SeedSession(t, db, user1.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	// User2: 3 visits
	for i := 0; i < 3; i++ {
		sessionID := helpers.SeedSession(t, db, user2.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	end := time.Now()
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	leaderboard, err := statsSvc.GetLeaderboardWithAutoCheckout(start, end, "visits", 10, false)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(leaderboard), 2)

	// User1 should be first (more visits)
	assert.Equal(t, int64(7), leaderboard[0].VisitCount)
	assert.Equal(t, int64(3), leaderboard[1].VisitCount)
}

func TestOfficeStatsService_GetLeaderboard_WithLimit(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	now := time.Now()

	// Create 5 users
	for i := 0; i < 5; i++ {
		user, _ := userRepo.Create("User"+string(rune('A'+i)), "RFID"+string(rune('A'+i)), "discord"+string(rune('a'+i)))
		sessionID := helpers.SeedSession(t, db, user.ID)
		sessionRepo.CheckOut(sessionID)
	}

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	leaderboard, err := statsSvc.GetLeaderboard(now.Add(-1*time.Hour), now.Add(1*time.Hour), "hours", 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(leaderboard))
}

func TestOfficeStatsService_GetUserStats(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}
	user, _ := userRepo.Create("StatsUser", "RFID405", "discord_stats")

	now := time.Now()
	sessionID1 := helpers.SeedSession(t, db, user.ID)
	time.Sleep(1 * time.Second) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID1)

	sessionID2 := helpers.SeedSession(t, db, user.ID)
	time.Sleep(1 * time.Second) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID2)

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	stats, err := statsSvc.GetUserStats(user.ID, now.Add(-1*time.Minute), now.Add(1*time.Hour))
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, user.ID, stats.UserID)
	assert.Equal(t, int64(2), stats.VisitCount)
	assert.Greater(t, stats.TotalHours, 0.0)
}

func TestOfficeStatsService_GetWeeklyReport(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user1, _ := userRepo.Create("WeekUser1", "RFID406", "discord_w1")
	user2, _ := userRepo.Create("WeekUser2", "RFID407", "discord_w2")

	sessionID1 := helpers.SeedSession(t, db, user1.ID)
	time.Sleep(100 * time.Millisecond) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID1)

	sessionID2 := helpers.SeedSession(t, db, user2.ID)
	time.Sleep(100 * time.Millisecond) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID2)

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	report, err := statsSvc.GetWeeklyReport("hours", false)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "week", report.PeriodType)
	assert.GreaterOrEqual(t, report.UniqueUsers, int64(2))
	assert.Equal(t, int64(2), report.TotalVisits)
}

func TestOfficeStatsService_GetMonthlyReport(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user, _ := userRepo.Create("MonthUser", "RFID408", "discord_month")

	sessionID := helpers.SeedSession(t, db, user.ID)
	time.Sleep(100 * time.Millisecond) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID)

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	now := time.Now()
	report, err := statsSvc.GetMonthlyReport(now.Year(), now.Month(), "hours", false)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "month", report.PeriodType)
	assert.GreaterOrEqual(t, report.UniqueUsers, int64(1))
}

func TestOfficeStatsService_GetCustomReport(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user, _ := userRepo.Create("CustomUser", "RFID409", "discord_custom")

	sessionID := helpers.SeedSession(t, db, user.ID)
	sessionRepo.CheckOut(sessionID)

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	now := time.Now()
	from := now.Add(-2 * time.Hour)
	to := now.Add(1 * time.Hour)

	report, err := statsSvc.GetCustomReport(from, to, 10, "hours", false)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "custom", report.PeriodType)
	assert.GreaterOrEqual(t, report.UniqueUsers, int64(1))
}

func TestOfficeStatsService_GetCustomReport_InvalidDateRange(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	sessionRepo := &repository.SessionRepo{DB: db}
	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	now := time.Now()
	from := now.Add(1 * time.Hour)
	to := now.Add(-1 * time.Hour)

	report, err := statsSvc.GetCustomReport(from, to, 10, "hours", false)
	assert.Error(t, err)
	assert.Nil(t, report)
}

func TestOfficeStatsService_GetPeriodStats(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user1, _ := userRepo.Create("PeriodUser1", "RFID410", "discord_p1")
	user2, _ := userRepo.Create("PeriodUser2", "RFID411", "discord_p2")

	now := time.Now()
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	sessionID1 := helpers.SeedSession(t, db, user1.ID)
	time.Sleep(1 * time.Second) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID1)

	sessionID2 := helpers.SeedSession(t, db, user2.ID)
	time.Sleep(1 * time.Second) // Ensure measurable duration
	sessionRepo.CheckOut(sessionID2)

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	report, err := statsSvc.GetPeriodStats(from, to, 10, "hours", false)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "hours", report.RankBy)
	assert.Equal(t, int64(2), report.UniqueUsers)
	assert.Equal(t, int64(2), report.TotalVisits)
	assert.Greater(t, report.AveragePerUser, 0.0)
}

func TestOfficeStatsService_GetPeriodStatsRankedByVisits(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	sessionRepo := &repository.SessionRepo{DB: db}

	user1, _ := userRepo.Create("VisitLeader", "RFID412", "discord_visit_leader")
	user2, _ := userRepo.Create("HourLeader", "RFID413", "discord_hour_leader")

	now := time.Now()
	from := now.Add(-2 * time.Hour)
	to := now.Add(2 * time.Hour)

	for i := 0; i < 4; i++ {
		sessionID := helpers.SeedSession(t, db, user1.ID)
		sessionRepo.CheckOut(sessionID)
		time.Sleep(10 * time.Millisecond)
	}

	sessionID := helpers.SeedSession(t, db, user2.ID)
	time.Sleep(1200 * time.Millisecond)
	sessionRepo.CheckOut(sessionID)

	statsSvc := &service.OfficeStatsService{Sessions: sessionRepo}

	report, err := statsSvc.GetPeriodStats(from, to, 10, "visits", false)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "visits", report.RankBy)
	assert.GreaterOrEqual(t, len(report.TopUsers), 2)
	assert.Equal(t, user1.ID, report.TopUsers[0].UserID)
	assert.Equal(t, int64(4), report.TopUsers[0].VisitCount)
}
