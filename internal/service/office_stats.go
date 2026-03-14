package service

import (
	"fmt"
	"time"

	"office/internal/domain"
	"office/internal/repository"
)

// OfficeStatsService provides statistics and analytics for user sessions
type OfficeStatsService struct {
	Sessions *repository.SessionRepo
}

func normalizeLeaderboardRankBy(rankBy string) string {
	if rankBy == "visits" {
		return "visits"
	}

	return "hours"
}

// GetLeaderboard returns top users by specified metric over a date range
func (s *OfficeStatsService) GetLeaderboard(from, to time.Time, metric string, limit int) ([]domain.UserStats, error) {
	return s.GetLeaderboardWithAutoCheckout(from, to, metric, limit, true)
}

// GetLeaderboardWithAutoCheckout returns top users by specified metric over a date range
// with control over whether to exclude non-RFID checkouts.
// When excludeAutoCheckout=true, results are RFID-only (trusted leaderboard mode).
func (s *OfficeStatsService) GetLeaderboardWithAutoCheckout(from, to time.Time, metric string, limit int, excludeAutoCheckout bool) ([]domain.UserStats, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	metric = normalizeLeaderboardRankBy(metric)

	stats, err := s.Sessions.GetAllUserStats(from, to, metric, limit, excludeAutoCheckout)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetUserStats returns statistics for a specific user over a date range
func (s *OfficeStatsService) GetUserStats(userID int64, from, to time.Time) (*domain.UserStats, error) {
	stats, err := s.Sessions.GetUserStats(userID, from, to, true)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetUserStatsWithAutoCheckout returns statistics for a specific user over a date range
// with control over whether to exclude non-RFID checkouts.
func (s *OfficeStatsService) GetUserStatsWithAutoCheckout(userID int64, from, to time.Time, excludeAutoCheckout bool) (*domain.UserStats, error) {
	stats, err := s.Sessions.GetUserStats(userID, from, to, excludeAutoCheckout)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetPeriodStats returns aggregated statistics for a time period.
func (s *OfficeStatsService) GetPeriodStats(from, to time.Time, topLimit int, rankBy string, excludeAutoCheckout bool) (*domain.PeriodStats, error) {
	if topLimit <= 0 {
		topLimit = 10
	}
	if topLimit > 100 {
		topLimit = 100
	}

	stats, err := s.Sessions.GetPeriodStats(from, to, topLimit, rankBy, excludeAutoCheckout)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetAllUserStatsForPeriod returns all user stats for a period ordered by the metric.
func (s *OfficeStatsService) GetAllUserStatsForPeriod(from, to time.Time, metric string) ([]domain.UserStats, error) {
	return s.Sessions.GetAllUserStats(from, to, metric, 0, true)
}

// GetAllUserStatsForPeriodWithAutoCheckout returns all user stats for a period ordered by the metric
// with control over whether to exclude non-RFID checkouts.
func (s *OfficeStatsService) GetAllUserStatsForPeriodWithAutoCheckout(from, to time.Time, metric string, excludeAutoCheckout bool) ([]domain.UserStats, error) {
	return s.Sessions.GetAllUserStats(from, to, metric, 0, excludeAutoCheckout)
}

// GetWeeklyReport returns statistics for the current week.
func (s *OfficeStatsService) GetWeeklyReport(rankBy string, excludeAutoCheckout bool) (*domain.PeriodStats, error) {
	now := time.Now()

	// Calculate week start (Monday) and end (Sunday)
	weekday := now.Weekday()
	var daysToMonday int
	if weekday == time.Sunday {
		daysToMonday = 6
	} else {
		daysToMonday = int(weekday) - 1
	}

	weekStart := now.AddDate(0, 0, -daysToMonday).Truncate(24 * time.Hour)
	weekEnd := weekStart.AddDate(0, 0, 7).Add(-1 * time.Second)

	stats, err := s.Sessions.GetPeriodStats(weekStart, weekEnd, 10, rankBy, excludeAutoCheckout)
	if err != nil {
		return nil, err
	}

	// Format the period as ISO week
	year, week := weekStart.ISOWeek()
	stats.Period = fmt.Sprintf("%d-W%02d", year, week)
	stats.PeriodType = "week"

	return stats, nil
}

// GetMonthlyReport returns statistics for the specified month.
func (s *OfficeStatsService) GetMonthlyReport(year int, month time.Month, rankBy string, excludeAutoCheckout bool) (*domain.PeriodStats, error) {
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-1 * time.Second)

	stats, err := s.Sessions.GetPeriodStats(monthStart, monthEnd, 10, rankBy, excludeAutoCheckout)
	if err != nil {
		return nil, err
	}

	stats.Period = monthStart.Format("2006-01")
	stats.PeriodType = "month"

	return stats, nil
}

// GetCustomReport returns statistics for a custom date range.
func (s *OfficeStatsService) GetCustomReport(from, to time.Time, topLimit int, rankBy string, excludeAutoCheckout bool) (*domain.PeriodStats, error) {
	if topLimit <= 0 {
		topLimit = 10
	}

	// Ensure from is before to
	if from.After(to) {
		return nil, fmt.Errorf("start date must be before end date")
	}

	stats, err := s.Sessions.GetPeriodStats(from, to, topLimit, rankBy, excludeAutoCheckout)
	if err != nil {
		return nil, err
	}

	stats.Period = from.Format("2006-01-02") + " to " + to.Format("2006-01-02")
	stats.PeriodType = "custom"

	return stats, nil
}
