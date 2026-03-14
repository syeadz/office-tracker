package dto

import (
	"office/internal/domain"
	"time"
)

// LeaderboardResponse contains leaderboard data
type LeaderboardResponse struct {
	RankBy string             `json:"rank_by"`
	Period string             `json:"period"`
	Count  int64              `json:"count"`
	Users  []domain.UserStats `json:"users"`
}

// UserStatsResponse contains user statistics
type UserStatsResponse struct {
	UserID          int64      `json:"user_id"`
	Name            string     `json:"name"`
	DiscordID       string     `json:"discord_id"`
	TotalHours      float64    `json:"total_hours"`
	VisitCount      int64      `json:"visit_count"`
	ActiveDays      int64      `json:"active_days"`
	AvgDuration     float64    `json:"avg_duration"`
	BusiestDay      string     `json:"busiest_day"`
	BusiestDayHours float64    `json:"busiest_day_hours"`
	LastVisit       *time.Time `json:"last_visit,omitempty"`
	FirstVisit      *time.Time `json:"first_visit,omitempty"`
}
