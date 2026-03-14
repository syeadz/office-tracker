// Package config provides configuration loading and management for the office tracker application.
// It loads configuration values from environment variables and provides defaults where necessary.
package config

import (
	"os"
	"strconv"
	"strings"

	"office/internal/logging"
)

// Config holds the configuration values for the application.
type Config struct {
	DiscordToken                string
	DiscordExecGuildID          string
	DiscordCommunityGuildID     string
	DiscordDashboardChannelName string
	DiscordReportsChannelID     string
	HTTPPort                    string
	DBPath                      string
	APIKey                      string
	CORSOrigins                 string
	CORSEnabled                 bool
	APIKeyEnabled               bool
	ReportsEnabled              bool
}

var log = logging.Component("config")

// Load reads configuration values from environment variables and returns a Config struct.
func Load() Config {
	getEnvTrim := func(key string) string {
		return strings.TrimSpace(os.Getenv(key))
	}

	cfg := Config{
		DiscordToken:                getEnvTrim("DISCORD_TOKEN"),
		DiscordExecGuildID:          getEnvTrim("DISCORD_EXEC_GUILD_ID"),
		DiscordCommunityGuildID:     getEnvTrim("DISCORD_COMMUNITY_GUILD_ID"),
		DiscordDashboardChannelName: getEnvTrim("DISCORD_DASHBOARD_CHANNEL_NAME"),
		DiscordReportsChannelID:     getEnvTrim("DISCORD_REPORTS_CHANNEL_ID"),
		HTTPPort:                    getEnvTrim("HTTP_PORT"),
		DBPath:                      getEnvTrim("DB_PATH"),
		APIKey:                      getEnvTrim("API_KEY"),
		CORSOrigins:                 getEnvTrim("CORS_ORIGINS"),
	}

	// Set defaults
	if cfg.DiscordDashboardChannelName == "" {
		cfg.DiscordDashboardChannelName = "office-tracker"
		log.Info("DISCORD_DASHBOARD_CHANNEL_NAME not set, using default", "channel_name", cfg.DiscordDashboardChannelName)
	}

	if cfg.HTTPPort == "" {
		cfg.HTTPPort = "8080"
		log.Info("HTTP_PORT not set, using default", "http_port", cfg.HTTPPort)
	} else {
		port, err := strconv.Atoi(cfg.HTTPPort)
		if err != nil || port < 1 || port > 65535 {
			log.Warn("invalid HTTP_PORT, using default", "value", cfg.HTTPPort)
			cfg.HTTPPort = "8080"
		}
	}

	if cfg.DBPath == "" {
		cfg.DBPath = "office.db"
		log.Info("DB_PATH not set, using default", "db_path", cfg.DBPath)
	}

	if cfg.DiscordToken == "" {
		log.Warn("DISCORD_TOKEN not set — bot disabled")
	} else {
		log.Info("discord bot enabled", "token_set", true)
	}

	if cfg.DiscordExecGuildID == "" {
		log.Warn("DISCORD_EXEC_GUILD_ID not set — exec guild commands disabled")
	}

	if cfg.DiscordCommunityGuildID == "" {
		log.Warn("DISCORD_COMMUNITY_GUILD_ID not set — community guild commands disabled")
	}

	if cfg.DiscordToken == "" && (cfg.DiscordExecGuildID != "" || cfg.DiscordCommunityGuildID != "") {
		log.Warn("discord guild IDs set but DISCORD_TOKEN is missing — bot will remain disabled")
	}

	// Enable CORS if CORS_ORIGINS is set
	if cfg.CORSOrigins != "" {
		cfg.CORSEnabled = true
		log.Info("CORS enabled", "origins", cfg.CORSOrigins)
		if cfg.CORSOrigins == "*" {
			log.Warn("CORS_ORIGINS is '*' — not recommended for production")
		}
	}

	// Enable API key authentication if API_KEY is set
	if cfg.APIKey != "" {
		cfg.APIKeyEnabled = true
		log.Info("API key authentication enabled")
	}

	// Enable reports if channel ID is set
	if cfg.DiscordReportsChannelID != "" {
		cfg.ReportsEnabled = true
		log.Info("weekly reports enabled", "channel_id", cfg.DiscordReportsChannelID)
	}

	log.Info(
		"configuration complete",
		"discord_token_set", cfg.DiscordToken != "",
		"discord_exec_guild_id_set", cfg.DiscordExecGuildID != "",
		"discord_community_guild_id_set", cfg.DiscordCommunityGuildID != "",
		"http_port", cfg.HTTPPort,
		"db", cfg.DBPath,
		"cors_enabled", cfg.CORSEnabled,
		"api_key_enabled", cfg.APIKeyEnabled,
		"reports_enabled", cfg.ReportsEnabled,
	)
	return cfg
}
