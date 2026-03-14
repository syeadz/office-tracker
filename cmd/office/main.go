// Package main is the entry point for the office tracker application. It initializes all components, starts the HTTP server and Discord bot, and handles graceful shutdown.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"office/internal/app"
	"office/internal/config"
	"office/internal/database"
	"office/internal/logging"
	"office/internal/service"
	"office/internal/transport/discord"
	httptransport "office/internal/transport/http"
)

func main() {
	// Setup logging
	logging.Setup()
	var log = logging.Component("main")

	log.Info("Starting office tracker application")

	// Load configuration from environment variables
	cfg := config.Load()

	// Database
	db := database.Open(cfg.DBPath)
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Error("migration failed", "err", err)
		panic(err)
	}

	// Create middleware config from environment
	mwConfig := httptransport.MiddlewareConfig{
		APIKey:        cfg.APIKey,
		CORSOrigins:   cfg.CORSOrigins,
		CORSEnabled:   cfg.CORSEnabled,
		APIKeyEnabled: cfg.APIKeyEnabled,
	}

	// Initialize application with all dependencies
	application := app.New(db, cfg.HTTPPort, mwConfig)

	// Discord bot
	bot, err := discord.New(cfg.DiscordToken, application.Services, cfg.DiscordExecGuildID, cfg.DiscordCommunityGuildID)
	if err != nil {
		log.Error("failed to create Discord bot", "err", err)
		panic(err)
	}

	// Configure reports if enabled
	var reportsService *service.ReportsService
	if bot != nil && cfg.ReportsEnabled {
		reportsDelivery := discord.NewReportsDelivery(bot.Session(), cfg.DiscordReportsChannelID)
		reportsService = service.NewReportsService(application.Services.Stats, reportsDelivery, true)
		application.Services.Reports = reportsService
		application.Services.Scheduler.SetReportsService(reportsService)
		log.Info("reports service configured")
	}

	// Start scheduler service (after reports are configured)
	if err := application.Services.Scheduler.Start(); err != nil {
		log.Error("failed to start scheduler", "err", err)
		panic(err)
	}

	if bot != nil {
		// Initialize dashboards by searching for the configured channel name in each guild
		if err := bot.InitializeDashboards(cfg.DiscordExecGuildID, cfg.DiscordCommunityGuildID, cfg.DiscordDashboardChannelName); err != nil {
			log.Error("failed to initialize dashboards", "err", err)
			panic(err)
		}
		// Wire attendance-change callback so scans and API-driven check-ins trigger dashboard refresh
		service.SetAttendanceChangeCallback(func() {
			if bot != nil {
				bot.TriggerRender()
			}
		})
		service.SetEnvironmentChangeCallback(func() {
			if bot != nil {
				bot.TriggerRender()
			}
		})

		// Initialize HTTP server with reports service
		application.HTTP = httptransport.New(cfg.HTTPPort, application.Services.Attendance, application.Services.User, application.Services.Session, application.Services.Stats, application.Services.Environment, reportsService, mwConfig)
		application.HTTP.Start()

		if err := bot.Start(); err != nil {
			log.Error("failed to start Discord bot", "err", err)
			panic(err)
		}
	}
	if bot == nil {
		// Initialize HTTP server without Discord bot
		application.HTTP = httptransport.New(cfg.HTTPPort, application.Services.Attendance, application.Services.User, application.Services.Session, application.Services.Stats, application.Services.Environment, nil, mwConfig)
		application.HTTP.Start()
	}

	log.Info("office-tracker running")

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop scheduler, HTTP and Discord bot in parallel
	done := make(chan struct{})
	go func() {
		application.Services.Scheduler.Stop(ctx)
		application.HTTP.Stop(ctx)
		close(done)
	}()

	if bot != nil {
		bot.Stop()
	}

	// Wait for HTTP to finish
	<-done
}
