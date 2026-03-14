package domain

import "time"

// UserStats contains aggregated statistics for a user over a period
type UserStats struct {
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

// PeriodStats contains aggregated statistics for a time period
type PeriodStats struct {
	Period          string      `json:"period"`
	PeriodType      string      `json:"period_type"`
	RankBy          string      `json:"rank_by"`
	StartDate       time.Time   `json:"start_date"`
	EndDate         time.Time   `json:"end_date"`
	TotalHours      float64     `json:"total_hours"`
	TotalVisits     int64       `json:"total_visits"`
	ActiveDays      int64       `json:"active_days"`
	UniqueUsers     int64       `json:"unique_users"`
	BusiestDay      string      `json:"busiest_day"`
	BusiestDayUsers int64       `json:"busiest_day_users"`
	PeakOccupancy   int64       `json:"peak_occupancy"`
	AveragePerUser  float64     `json:"average_per_user"`
	TopUsers        []UserStats `json:"top_users"`
}
