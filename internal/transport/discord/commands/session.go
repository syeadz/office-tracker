// Package commands contains Discord slash command handlers for the office tracker bot.
package commands

import (
	"fmt"
	"strings"
	"time"

	"office/internal/logging"
	"office/internal/query"
	"office/internal/repository"
	"office/internal/service"

	"github.com/bwmarrin/discordgo"
)

var sessionLog = logging.Component("discord.commands.session")

// SessionCommands handles all session-related slash commands
type SessionCommands struct {
	sessionSvc *service.SessionService
}

// NewSessionCommands creates a new SessionCommands handler
func NewSessionCommands(sessionSvc *service.SessionService) *SessionCommands {
	return &SessionCommands{sessionSvc: sessionSvc}
}

// GetApplicationCommands returns the slash command definitions for session management
func (sc *SessionCommands) GetApplicationCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "session-list",
			Description: "List office sessions with optional filters",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "user",
					Description: "Filter by Discord user (mention or select)",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    false,
				},
				{
					Name:        "name",
					Description: "Filter by user name (optional)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
				{
					Name:        "active_only",
					Description: "Show only active sessions (optional)",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    false,
				},
				{
					Name:        "limit",
					Description: "Results per page (default: 10, max: 50)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
				{
					Name:        "page",
					Description: "Page number (default: 1)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
				{
					Name:        "order",
					Description: "Sort order: 'asc' or 'desc' (default: desc)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Newest First", Value: "desc"},
						{Name: "Oldest First", Value: "asc"},
					},
				},
				{
					Name:        "sort_by",
					Description: "Sort field: 'check_in', 'check_out', or 'user_name' (default: check_in)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Check In Time", Value: "check_in"},
						{Name: "Check Out Time", Value: "check_out"},
						{Name: "User Name", Value: "user_name"},
					},
				},
			},
		},
		{
			Name:        "session-get",
			Description: "Get details of a specific session by ID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "session_id",
					Description: "The session ID",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		{
			Name:        "session-active",
			Description: "Show all currently active office sessions",
		},
	}
}

// HandleCommand routes session commands to appropriate handlers
func (sc *SessionCommands) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmdName string) {
	sessionLog.Info("session command received", "command", cmdName, "user_id", interactionUserID(i), "username", interactionUsername(i))

	switch cmdName {
	case "session-list":
		sc.handleList(s, i)
	case "session-get":
		sc.handleGet(s, i)
	case "session-active":
		sc.handleActive(s, i)
	}
}

// handleList lists sessions with optional filters and pagination
func (sc *SessionCommands) handleList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options

	filter := query.SessionFilter{}
	limit := int64(10)
	page := int64(1)

	for _, opt := range opts {
		switch opt.Name {
		case "user":
			discordID := opt.UserValue(nil).ID
			filter.DiscordID = &discordID
		case "name":
			name := opt.StringValue()
			if name != "" {
				filter.NameLike = &name
			}
		case "active_only":
			filter.ActiveOnly = opt.BoolValue()
		case "limit":
			l := opt.IntValue()
			if l > 0 && l <= 50 {
				limit = l
			}
		case "page":
			p := opt.IntValue()
			if p > 0 {
				page = p
			}
		case "order":
			order := opt.StringValue()
			if order == "asc" || order == "desc" {
				filter.OrderBy = order
			}
		case "sort_by":
			sortBy := opt.StringValue()
			if sortBy == "check_in" || sortBy == "check_out" || sortBy == "user_name" {
				filter.SortBy = sortBy
			}
		}
	}

	filter.Limit = int(limit)
	filter.Offset = int((page - 1) * limit)

	countFilter := filter
	countFilter.Limit = 0
	countFilter.Offset = 0
	totalSessions, err := sc.sessionSvc.CountSessions(countFilter)
	if err != nil {
		sessionLog.Error("failed to count sessions", "error", err)
		respondEphemeral(s, i, "❌ Failed to fetch sessions.")
		return
	}

	result, err := sc.sessionSvc.ListSessions(filter, false)
	if err != nil {
		sessionLog.Error("failed to list sessions", "error", err)
		respondEphemeral(s, i, "❌ Failed to fetch sessions.")
		return
	}

	sessions, ok := result.([]*repository.SessionWithUser)
	if !ok || len(sessions) == 0 {
		respondEphemeral(s, i, "📭 No sessions found.")
		return
	}

	// Build session list
	var sessionLines []string
	for _, sess := range sessions {
		status := "✅ Active"
		if sess.CheckOut != nil {
			status = fmt.Sprintf("❌ Checked out at <t:%d:R>", sess.CheckOut.Unix())
		}
		sessionLines = append(sessionLines, fmt.Sprintf("**%s** - Checked in <t:%d:R> - %s",
			sess.UserName, sess.CheckIn.Unix(), status))
	}

	description := strings.Join(sessionLines, "\n")
	if len(description) > 2048 {
		description = description[:2048]
		description += "\n\n*Output truncated...*"
	}

	title := "📊 Session List"
	if filter.ActiveOnly {
		title = "🟢 Active Sessions"
	}

	totalPages := int64(1)
	if limit > 0 {
		totalPages = (totalSessions + limit - 1) / limit
		if totalPages < 1 {
			totalPages = 1
		}
	}
	if page > totalPages {
		page = totalPages
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0x3498DB,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d of %d (showing %d of %d sessions) | Use HTTP API for advanced queries", page, totalPages, len(sessions), totalSessions),
		},
	}

	respondEphemeralEmbed(s, i, embed)
}

// handleGet retrieves a specific session by ID
func (sc *SessionCommands) handleGet(s *discordgo.Session, i *discordgo.InteractionCreate) {
	sessionID, ok := integerOptionValue(i, "session_id")
	if !ok {
		respondEphemeral(s, i, "❌ Session ID is required.")
		return
	}

	// Use the service method to maintain layered architecture
	sess, err := sc.sessionSvc.GetSessionByID(sessionID)
	if err != nil {
		sessionLog.Error("failed to get session", "session_id", sessionID, "error", err)
		respondEphemeral(s, i, "❌ Session not found.")
		return
	}

	status := "✅ Active"
	checkOutTime := "Still checked in"
	if sess.CheckOut != nil {
		status = "❌ Checked Out"
		checkOutTime = fmt.Sprintf("<t:%d:f> (<t:%d:R>)", sess.CheckOut.Unix(), sess.CheckOut.Unix())
	}

	duration := ""
	if sess.CheckOut != nil {
		d := sess.CheckOut.Sub(sess.CheckIn)
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		duration = fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		d := time.Since(sess.CheckIn)
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		duration = fmt.Sprintf("%dh %dm (ongoing)", hours, minutes)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📋 Session #%d", sess.ID),
		Color:       0x3498DB,
		Description: fmt.Sprintf("**%s** %s", sess.UserName, status),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User ID", Value: fmt.Sprintf("%d", sess.UserID), Inline: true},
			{Name: "Session ID", Value: fmt.Sprintf("%d", sess.ID), Inline: true},
			{Name: "Status", Value: status, Inline: true},
			{Name: "Checked In", Value: fmt.Sprintf("<t:%d:f> (<t:%d:R>)", sess.CheckIn.Unix(), sess.CheckIn.Unix()), Inline: false},
			{Name: "Checked Out", Value: checkOutTime, Inline: false},
			{Name: "Duration", Value: duration, Inline: true},
		},
	}

	respondEphemeralEmbed(s, i, embed)
}

// handleActive shows all currently active sessions
func (sc *SessionCommands) handleActive(s *discordgo.Session, i *discordgo.InteractionCreate) {
	filter := query.SessionFilter{
		ActiveOnly: true,
		Limit:      50, // Get a good chunk of active sessions
	}

	result, err := sc.sessionSvc.ListSessions(filter, false)
	if err != nil {
		sessionLog.Error("failed to list active sessions", "error", err)
		respondEphemeral(s, i, "❌ Failed to fetch sessions.")
		return
	}

	sessions, ok := result.([]*repository.SessionWithUser)
	if !ok || len(sessions) == 0 {
		respondEphemeral(s, i, "🟢 No one is in the office right now.")
		return
	}

	// Build active list
	var sessionLines []string
	for _, sess := range sessions {
		d := time.Since(sess.CheckIn)
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		sessionLines = append(sessionLines, fmt.Sprintf("• **%s** - In office for %dh %dm", sess.UserName, hours, minutes))
	}

	description := strings.Join(sessionLines, "\n")
	if len(description) > 2048 {
		description = description[:2048]
		description += "\n\n*List truncated...*"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🟢 Currently in Office (%d)", len(sessions)),
		Description: description,
		Color:       0x2ECC71,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Last updated: %s", time.Now().Format("15:04:05")),
		},
	}

	respondEphemeralEmbed(s, i, embed)
}
