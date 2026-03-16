// Package service provides business logic for the application.
package service

import (
	"context"
	"log/slog"
	"time"

	"office/internal/query"
	"office/internal/repository"

	"github.com/robfig/cron/v3"
)

// SchedulerService manages scheduled jobs like auto-checkouts and reports.
type SchedulerService struct {
	cron     *cron.Cron
	sessions *repository.SessionRepo
	reports  *ReportsService
}

// NewSchedulerService creates a new scheduler service.
func NewSchedulerService(sessions *repository.SessionRepo) *SchedulerService {
	return &SchedulerService{
		cron:     cron.New(cron.WithLocation(time.Local)),
		sessions: sessions,
	}
}

// SetReportsService sets the reports service for scheduled report generation
func (s *SchedulerService) SetReportsService(reports *ReportsService) {
	s.reports = reports
}

// Start initializes and starts all scheduled jobs.
func (s *SchedulerService) Start() error {
	slog.Info("starting scheduler service")

	// Auto-checkout at 4 AM every day (local timezone, standard cron format: minute hour day month dayofweek)
	_, err := s.cron.AddFunc("0 4 * * *", s.AutoCheckoutJob)
	if err != nil {
		slog.Error("failed to schedule auto-checkout job", "error", err)
		return err
	}

	// Weekly report every Monday at 9 AM
	if s.reports != nil && s.reports.IsEnabled() {
		_, err = s.cron.AddFunc("0 9 * * 1", s.WeeklyReportJob)
		if err != nil {
			slog.Error("failed to schedule weekly report job", "error", err)
			return err
		}
		slog.Info("weekly report job scheduled for Mondays at 09:00 local time")

		// Monthly report on first day of month at 9 AM
		_, err = s.cron.AddFunc("0 9 1 * *", s.MonthlyReportJob)
		if err != nil {
			slog.Error("failed to schedule monthly report job", "error", err)
			return err
		}
		slog.Info("monthly report job scheduled for 1st of month at 09:00 local time")
	}

	s.cron.Start()
	slog.Info("scheduler service started with auto-checkout job at 04:00 local time")
	return nil
}

// Stop gracefully stops the scheduler.
func (s *SchedulerService) Stop(ctx context.Context) {
	slog.Info("stopping scheduler service")
	<-s.cron.Stop().Done()
}

// WeeklyReportJob generates and sends the weekly office report
func (s *SchedulerService) WeeklyReportJob() {
	slog.Info("running weekly report job")

	if s.reports == nil {
		slog.Warn("reports service not configured")
		return
	}

	if err := s.reports.GenerateAndSendWeeklyReport(); err != nil {
		slog.Error("weekly report job failed", "error", err)
	}
}

// MonthlyReportJob generates and sends the monthly office report
func (s *SchedulerService) MonthlyReportJob() {
	slog.Info("running monthly report job")

	if s.reports == nil {
		slog.Warn("reports service not configured")
		return
	}

	if err := s.reports.GenerateAndSendMonthlyReport(); err != nil {
		slog.Error("monthly report job failed", "error", err)
	}
}

// AutoCheckoutJob checks out all open sessions.
// This runs daily at 4 AM to ensure no one is accidentally checked in overnight.
func (s *SchedulerService) AutoCheckoutJob() {
	slog.Info("running auto-checkout job")

	// Get all open sessions
	filter := query.SessionFilter{ActiveOnly: true, OrderBy: "asc", SortBy: "check_in"}

	sessions, err := s.sessions.List(filter)
	if err != nil {
		slog.Error("failed to fetch open sessions for auto-checkout", "error", err)
		return
	}

	checkedOut := 0
	for _, session := range sessions {
		err := s.sessions.CheckOutWithMethod(session.ID, repository.CheckOutMethodAuto)
		if err != nil {
			slog.Error("failed to auto-checkout session",
				"session_id", session.ID,
				"user_id", session.UserID,
				"error", err)
			continue
		}
		checkedOut++
	}

	slog.Info("auto-checkout job completed",
		"sessions_checked_out", checkedOut,
		"total_open_sessions", len(sessions))
}
