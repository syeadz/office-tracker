package domain

import "time"

// PeriodReport represents a periodic office statistics report (weekly, monthly, etc.)
type PeriodReport struct {
	Period          string      `json:"period"`
	PeriodType      string      `json:"period_type"` // "week" or "month"
	StartDate       time.Time   `json:"start_date"`
	EndDate         time.Time   `json:"end_date"`
	TotalHours      float64     `json:"total_hours"`
	TotalVisits     int64       `json:"total_visits"`
	UniqueUsers     int64       `json:"unique_users"`
	ActiveDays      int64       `json:"active_days"`
	BusiestDay      string      `json:"busiest_day"`
	BusiestDayUsers int64       `json:"busiest_day_users"`
	PeakOccupancy   int64       `json:"peak_occupancy"`
	TopUsers        []UserStats `json:"top_users"`
	GeneratedAt     time.Time   `json:"generated_at"`

	// Comparison with previous period
	HasComparison bool    `json:"has_comparison"`
	HoursChange   float64 `json:"hours_change,omitempty"`  // Percentage change
	VisitsChange  float64 `json:"visits_change,omitempty"` // Percentage change
}
