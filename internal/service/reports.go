package service

import (
	"fmt"
	"time"

	"office/internal/domain"
	"office/internal/logging"
)

var reportsLog = logging.Component("service.reports")

// ReportDelivery defines the interface for delivering reports to external systems
type ReportDelivery interface {
	SendPeriodReport(report *domain.PeriodReport, reportType string) error
}

// ReportsService orchestrates report generation and delivery
type ReportsService struct {
	stats               *OfficeStatsService
	delivery            ReportDelivery
	enabled             bool
	latestWeeklyReport  *domain.PeriodReport
	latestMonthlyReport *domain.PeriodReport
}

// NewReportsService creates a new reports service
func NewReportsService(stats *OfficeStatsService, delivery ReportDelivery, enabled bool) *ReportsService {
	return &ReportsService{
		stats:    stats,
		delivery: delivery,
		enabled:  enabled,
	}
}

// GenerateAndSendWeeklyReport generates and sends the weekly office report
func (r *ReportsService) GenerateAndSendWeeklyReport() error {
	if !r.enabled {
		reportsLog.Info("reports disabled, skipping weekly report")
		return nil
	}

	if r.delivery == nil {
		reportsLog.Warn("no delivery method configured, skipping weekly report")
		return nil
	}

	reportsLog.Info("generating weekly report")

	// Generate the report
	report, err := r.generateWeeklyReport()
	if err != nil {
		reportsLog.Error("failed to generate weekly report", "error", err)
		return fmt.Errorf("generate weekly report: %w", err)
	}

	// Check if there was at least one visit
	if report.TotalVisits == 0 {
		reportsLog.Info("no activity this week, skipping report send")
		r.latestWeeklyReport = report
		return nil
	}

	// Send the report
	if err := r.delivery.SendPeriodReport(report, "weekly"); err != nil {
		reportsLog.Error("failed to send weekly report", "error", err)
		return fmt.Errorf("send weekly report: %w", err)
	}

	// Cache the report for HTTP endpoint
	r.latestWeeklyReport = report

	reportsLog.Info("weekly report sent successfully",
		"period", report.Period,
		"total_hours", report.TotalHours,
		"unique_users", report.UniqueUsers,
	)

	return nil
}

// GenerateAndSendMonthlyReport generates and sends the monthly office report
func (r *ReportsService) GenerateAndSendMonthlyReport() error {
	if !r.enabled {
		reportsLog.Info("reports disabled, skipping monthly report")
		return nil
	}

	if r.delivery == nil {
		reportsLog.Warn("no delivery method configured, skipping monthly report")
		return nil
	}

	reportsLog.Info("generating monthly report")

	// Generate the report
	report, err := r.generateMonthlyReport()
	if err != nil {
		reportsLog.Error("failed to generate monthly report", "error", err)
		return fmt.Errorf("generate monthly report: %w", err)
	}

	// Check if there was at least one visit
	if report.TotalVisits == 0 {
		reportsLog.Info("no activity this month, skipping report send")
		r.latestMonthlyReport = report
		return nil
	}

	// Send the report
	if err := r.delivery.SendPeriodReport(report, "monthly"); err != nil {
		reportsLog.Error("failed to send monthly report", "error", err)
		return fmt.Errorf("send monthly report: %w", err)
	}

	// Cache the report for HTTP endpoint
	r.latestMonthlyReport = report

	reportsLog.Info("monthly report sent successfully",
		"period", report.Period,
		"total_hours", report.TotalHours,
		"unique_users", report.UniqueUsers,
	)

	return nil
}

// generateWeeklyReport creates a weekly report with statistics and insights
func (r *ReportsService) generateWeeklyReport() (*domain.PeriodReport, error) {
	if r.stats == nil {
		return nil, fmt.Errorf("stats service not configured")
	}

	// Get previous week stats (the completed week we're reporting on)
	now := time.Now()
	weekday := now.Weekday()
	var daysToMonday int
	if weekday == time.Sunday {
		daysToMonday = 6
	} else {
		daysToMonday = int(weekday) - 1
	}

	// Last week = the week that just completed
	lastWeekStart := now.AddDate(0, 0, -daysToMonday-7).Truncate(24 * time.Hour)
	lastWeekEnd := lastWeekStart.AddDate(0, 0, 7).Add(-1 * time.Second)

	currentStats, err := r.stats.GetPeriodStats(lastWeekStart, lastWeekEnd, 10, "hours", true)
	if err != nil {
		return nil, fmt.Errorf("get previous week stats: %w", err)
	}

	// Get week before last for comparison
	twoWeeksAgoStart := lastWeekStart.AddDate(0, 0, -7)
	twoWeeksAgoEnd := twoWeeksAgoStart.AddDate(0, 0, 7).Add(-1 * time.Second)

	previousStats, err := r.stats.GetPeriodStats(twoWeeksAgoStart, twoWeeksAgoEnd, 10, "hours", true)
	if err != nil {
		reportsLog.Warn("could not get comparison week stats", "error", err)
		// Continue without comparison data
		previousStats = nil
	}

	// Build the weekly report
	year, week := lastWeekStart.ISOWeek()
	report := &domain.PeriodReport{
		Period:          fmt.Sprintf("%d-W%02d", year, week),
		PeriodType:      "week",
		StartDate:       currentStats.StartDate,
		EndDate:         currentStats.EndDate,
		TotalHours:      currentStats.TotalHours,
		TotalVisits:     currentStats.TotalVisits,
		UniqueUsers:     currentStats.UniqueUsers,
		ActiveDays:      currentStats.ActiveDays,
		BusiestDay:      currentStats.BusiestDay,
		BusiestDayUsers: currentStats.BusiestDayUsers,
		PeakOccupancy:   currentStats.PeakOccupancy,
		TopUsers:        currentStats.TopUsers,
		GeneratedAt:     time.Now(),
	}

	// Calculate comparisons if we have previous week data
	if previousStats != nil && previousStats.TotalHours > 0 {
		report.HoursChange = ((currentStats.TotalHours - previousStats.TotalHours) / previousStats.TotalHours) * 100
		report.HasComparison = true

		if previousStats.TotalVisits > 0 {
			report.VisitsChange = float64(currentStats.TotalVisits-previousStats.TotalVisits) / float64(previousStats.TotalVisits) * 100
		}
	}

	return report, nil
}

// generateMonthlyReport creates a monthly report with statistics and insights
func (r *ReportsService) generateMonthlyReport() (*domain.PeriodReport, error) {
	if r.stats == nil {
		return nil, fmt.Errorf("stats service not configured")
	}

	// Get previous month stats (the completed month we're reporting on)
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	currentStats, err := r.stats.GetMonthlyReport(lastMonth.Year(), lastMonth.Month(), "hours", true)
	if err != nil {
		return nil, fmt.Errorf("get previous month stats: %w", err)
	}

	// Get month before last for comparison
	twoMonthsAgo := now.AddDate(0, -2, 0)
	previousStats, err := r.stats.GetMonthlyReport(twoMonthsAgo.Year(), twoMonthsAgo.Month(), "hours", true)
	if err != nil {
		reportsLog.Warn("could not get comparison month stats", "error", err)
		// Continue without comparison data
		previousStats = nil
	}

	// Build the monthly report
	report := &domain.PeriodReport{
		Period:          currentStats.Period,
		PeriodType:      "month",
		StartDate:       currentStats.StartDate,
		EndDate:         currentStats.EndDate,
		TotalHours:      currentStats.TotalHours,
		TotalVisits:     currentStats.TotalVisits,
		UniqueUsers:     currentStats.UniqueUsers,
		ActiveDays:      currentStats.ActiveDays,
		BusiestDay:      currentStats.BusiestDay,
		BusiestDayUsers: currentStats.BusiestDayUsers,
		PeakOccupancy:   currentStats.PeakOccupancy,
		TopUsers:        currentStats.TopUsers,
		GeneratedAt:     time.Now(),
	}

	// Calculate comparisons if we have previous month data
	if previousStats != nil && previousStats.TotalHours > 0 {
		report.HoursChange = ((currentStats.TotalHours - previousStats.TotalHours) / previousStats.TotalHours) * 100
		report.HasComparison = true

		if previousStats.TotalVisits > 0 {
			report.VisitsChange = float64(currentStats.TotalVisits-previousStats.TotalVisits) / float64(previousStats.TotalVisits) * 100
		}
	}

	return report, nil
}

// SetEnabled enables or disables report generation
func (r *ReportsService) SetEnabled(enabled bool) {
	r.enabled = enabled
	reportsLog.Info("reports enabled status changed", "enabled", enabled)
}

// IsEnabled returns whether reports are currently enabled
func (r *ReportsService) IsEnabled() bool {
	return r.enabled
}

// GetLatestWeeklyReport returns the most recently generated weekly report
func (r *ReportsService) GetLatestWeeklyReport() (*domain.PeriodReport, error) {
	if r.latestWeeklyReport == nil {
		return nil, fmt.Errorf("no weekly report available")
	}
	return r.latestWeeklyReport, nil
}

// GetLatestMonthlyReport returns the most recently generated monthly report
func (r *ReportsService) GetLatestMonthlyReport() (*domain.PeriodReport, error) {
	if r.latestMonthlyReport == nil {
		return nil, fmt.Errorf("no monthly report available")
	}
	return r.latestMonthlyReport, nil
}
