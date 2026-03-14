// Package discord provides a Discord bot implementation for office presence tracking.
//
// Structure:
//   - bot.go: Core Bot initialization and lifecycle
//   - handlers.go: Interaction routing (slash commands, buttons)
//   - commands_base.go: Setup and ping commands
//   - dashboards.go: Dashboard rendering and refresh logic
//   - commands/user.go: User management commands
//   - commands/session.go: Session querying commands
//   - commands/attendance.go: Attendance commands (checkin/checkout)
//   - commands/help.go: Help command
package discord

import (
	"sync"
	"time"

	"office/internal/app"
	"office/internal/logging"
	"office/internal/transport/discord/commands"

	"github.com/bwmarrin/discordgo"
)

const (
	// ColorGreen used for success/active states in embeds
	ColorGreen = 0x2ECC71
	// ColorRed used for errors/checked out states in embeds
	ColorRed = 0xE74C3C
	// ColorGrey used for neutral/empty states in embeds
	ColorGrey = 0x95A5A6
)

// Dashboard represents a tracked presence dashboard in a specific channel
type Dashboard struct {
	ChannelID string
	MessageID string
	GuildID   string // Cached for efficiency
}

// Bot manages Discord interactions and presence tracking.
// It handles slash commands, button interactions, and dashboard rendering.
// The bot is null-safe - all methods check if b == nil before executing.
type Bot struct {
	// Discord session and services
	session  *discordgo.Session
	services *app.Services

	// Dashboard management
	dashboards   map[string]*Dashboard // channelID -> Dashboard
	dashboardsMu sync.RWMutex
	renderChan   chan struct{} // channel to trigger dashboard renders
	stopChan     chan struct{} // channel to signal bot shutdown

	// Channel IDs for dashboards
	execChannelID      string
	communityChannelID string

	// Refresh throttling per channel
	lastRefreshTime   map[string]time.Time // Track last refresh time per channel
	lastRefreshMux    sync.RWMutex
	refreshCooldownMs int // Cooldown duration in milliseconds

	// Guild configuration
	execGuildID      string // Guild ID for exec server
	communityGuildID string // Guild ID for community server

	// Debouncing for RFID scan events
	debounceTimer   *time.Timer // Timer for debouncing scan-triggered updates
	debounceMux     sync.Mutex  // Protects debounceTimer
	debounceDelayMs int         // Debounce delay in milliseconds

	// Delegated command handlers
	userCommands        *commands.UserCommands         // User management commands
	sessionCommands     *commands.SessionCommands      // Session querying commands
	attendanceCommands  *commands.AttendanceCommands   // Attendance commands
	helpCommands        *commands.HelpCommands         // Help command
	statsCommands       *commands.StatsCommands        // Stats commands
	myStatsCommands     *commands.MyStatsCommands      // Personal stats commands
	reportsToggleCmd    *commands.ReportsToggleCommand // Reports toggle command (Admin only)
	scanHistoryCommands *commands.ScanHistoryCommands  // Scan history command (Admin only)
}

// Session returns the Discord session for use by external components
func (b *Bot) Session() *discordgo.Session {
	if b == nil {
		return nil
	}
	return b.session
}

var log = logging.Component("discord")

// New creates a new Discord bot instance with the provided token and services.
// If the token is empty, it returns nil without an error.
func New(token string, services *app.Services, execGuildID string, communityGuildID string) (*Bot, error) {
	if token == "" {
		return nil, nil
	}

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		session:             s,
		services:            services,
		dashboards:          make(map[string]*Dashboard),
		renderChan:          make(chan struct{}), // unbuffered channel for render requests
		stopChan:            make(chan struct{}),
		lastRefreshTime:     make(map[string]time.Time),
		refreshCooldownMs:   1000, // 1 second cooldown per channel
		execGuildID:         execGuildID,
		communityGuildID:    communityGuildID,
		debounceDelayMs:     1500, // 1.5 second debounce delay for scan-triggered updates
		userCommands:        commands.NewUserCommands(services.User),
		sessionCommands:     commands.NewSessionCommands(services.Session),
		attendanceCommands:  commands.NewAttendanceCommands(services.User, services.Session),
		helpCommands:        commands.NewHelpCommands(),
		statsCommands:       commands.NewStatsCommands(services.Stats),
		myStatsCommands:     commands.NewMyStatsCommands(services.Stats, services.User),
		scanHistoryCommands: commands.NewScanHistoryCommands(services.Attendance),
	}

	// Add reports toggle command if reports service is available
	if services.Reports != nil {
		bot.reportsToggleCmd = commands.NewReportsToggleCommand(services.Reports)
	}

	// Register handlers
	s.AddHandler(bot.handleInteractionCreate)
	s.AddHandler(bot.handleMessageCreate)

	return bot, nil
}

// RegisterSlashCommands synchronizes slash commands with Discord.
// It uses bulk overwrite per guild so stale commands are removed automatically.
func (b *Bot) RegisterSlashCommands() error {
	if b == nil {
		return nil
	}

	// Admin permission required
	adminPermission := int64(discordgo.PermissionAdministrator)

	// Build user and session commands
	userCmds := b.userCommands.GetApplicationCommands()
	sessionCmds := b.sessionCommands.GetApplicationCommands()
	attendanceCmds := b.attendanceCommands.GetApplicationCommands()
	helpCmds := b.helpCommands.GetApplicationCommands()
	statsCmds := b.statsCommands.GetApplicationCommands()
	myStatsCmds := b.myStatsCommands.GetApplicationCommands()
	scanHistoryCmds := b.scanHistoryCommands.GetApplicationCommands()

	// Prepare reports toggle command if available
	var reportsToggleCmds []*discordgo.ApplicationCommand
	if b.reportsToggleCmd != nil {
		reportsToggleCmds = []*discordgo.ApplicationCommand{b.reportsToggleCmd.Definition()}
	}

	// Build exec guild command list
	execCmds := []*discordgo.ApplicationCommand{
		{
			Name:                     "setup",
			Description:              "Create an office presence dashboard in this channel",
			DefaultMemberPermissions: &adminPermission,
		},
		{
			Name:        "ping",
			Description: "Ping the bot",
		},
	}
	execCmds = append(execCmds, userCmds...)
	execCmds = append(execCmds, sessionCmds...)
	execCmds = append(execCmds, attendanceCmds...)
	execCmds = append(execCmds, helpCmds...)
	execCmds = append(execCmds, statsCmds...)
	execCmds = append(execCmds, myStatsCmds...)
	execCmds = append(execCmds, reportsToggleCmds...)
	execCmds = append(execCmds, scanHistoryCmds...)

	// Define guild-specific commands
	guildCommands := map[string][]*discordgo.ApplicationCommand{
		b.communityGuildID: {
			{
				Name:                     "setup",
				Description:              "Create an office presence dashboard in this channel",
				DefaultMemberPermissions: &adminPermission,
			},
		},
		b.execGuildID: execCmds,
	}

	// Clear global commands for this application.
	// This prevents old global commands from showing alongside guild commands.
	if _, err := b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, "", []*discordgo.ApplicationCommand{}); err != nil {
		log.Warn("failed to clear global slash commands", "error", err)
	}

	// Synchronize per-guild commands via bulk overwrite.
	for guildID, cmds := range guildCommands {
		if guildID == "" {
			continue
		}

		if _, err := b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, guildID, cmds); err != nil {
			log.Error("failed to sync guild slash commands", "guild_id", guildID, "error", err)
			return err
		}

		log.Info("slash commands synchronized", "guild_id", guildID, "count", len(cmds))
	}

	log.Info("slash commands sync complete")
	return nil
}

// dashboardRefresher periodically updates dashboards every 5 minutes
// Also handles on-demand render requests from the renderChan
func (b *Bot) dashboardRefresher() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Info("Periodic dashboard refresh triggered")
			b.renderAllDashboards()

		case <-b.renderChan:
			// On-demand render request (e.g., from RFID scan)
			log.Info("On-demand dashboard refresh triggered")
			b.renderAllDashboards()

		case <-b.stopChan:
			return
		}
	}
}

// TriggerRender requests a debounced dashboard update for all dashboards (called when scan happens)
// Multiple rapid calls within the debounce window will result in a single render
func (b *Bot) TriggerRender() {
	b.debounceMux.Lock()
	defer b.debounceMux.Unlock()

	// Cancel any existing debounce timer
	if b.debounceTimer != nil {
		b.debounceTimer.Stop()
	}

	// Set a new timer that will trigger the render after the debounce delay
	b.debounceTimer = time.AfterFunc(time.Duration(b.debounceDelayMs)*time.Millisecond, func() {
		select {
		case b.renderChan <- struct{}{}:
		case <-b.stopChan:
		}
	})
}

// Start opens the Discord session and begins listening for events. If the bot is nil, it does nothing.
func (b *Bot) Start() error {
	if b == nil {
		return nil
	}
	log.Info("Discord bot starting")
	if err := b.session.Open(); err != nil {
		return err
	}

	// Register commands
	if err := b.RegisterSlashCommands(); err != nil {
		return err
	}

	log.Info("Discord bot ready, refreshing dashboard")
	// Immediately refresh dashboards on startup
	b.renderAllDashboards()

	// Start refresh ticker
	go b.dashboardRefresher()

	return nil
}

// FindChannelIDByName searches for a channel ID by name in the specified guild
func (b *Bot) FindChannelIDByName(guildID, channelName string) (string, error) {
	if b == nil || b.session == nil {
		return "", nil
	}

	channels, err := b.session.GuildChannels(guildID)
	if err != nil {
		return "", err
	}

	for _, ch := range channels {
		if ch.Name == channelName {
			return ch.ID, nil
		}
	}

	return "", nil
}

// IsRefreshOnCooldown checks if a channel is still on refresh cooldown
func (b *Bot) IsRefreshOnCooldown(channelID string) bool {
	b.lastRefreshMux.RLock()
	lastTime, exists := b.lastRefreshTime[channelID]
	b.lastRefreshMux.RUnlock()

	if !exists {
		return false
	}

	return time.Since(lastTime) < time.Duration(b.refreshCooldownMs)*time.Millisecond
}

// UpdateRefreshTime updates the last refresh time for a channel
func (b *Bot) UpdateRefreshTime(channelID string) {
	b.lastRefreshMux.Lock()
	defer b.lastRefreshMux.Unlock()
	b.lastRefreshTime[channelID] = time.Now()
}

// IsExecChannel returns true if the channel ID is the exec dashboard channel
func (b *Bot) IsExecChannel(channelID string) bool {
	return b.execChannelID == channelID
}

// IsCommunityChannel returns true if the channel ID is the community dashboard channel
func (b *Bot) IsCommunityChannel(channelID string) bool {
	return b.communityChannelID == channelID
}

// Stop closes the Discord session. If the bot is nil, it does nothing.
func (b *Bot) Stop() {
	if b == nil {
		return
	}
	log.Info("Discord bot stopping")
	close(b.stopChan)
	b.session.Close()
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
