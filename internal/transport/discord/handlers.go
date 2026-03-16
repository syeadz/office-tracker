package discord

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"office/internal/query"
	"office/internal/repository"
	"office/internal/service"

	"github.com/bwmarrin/discordgo"
)

// handleInteractionCreate routes Discord interactions to appropriate handlers.
// Handles both slash commands and message component (button) interactions.
func (b *Bot) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleButtonClick(s, i)
	}
}

func interactionUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i == nil {
		return nil
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User
	}
	return i.User
}

// handleMessageCreate responds to natural language office presence queries.
func (b *Bot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if b == nil || b.services == nil {
		return
	}
	if m == nil || m.Author == nil {
		return
	}
	if m.Author.Bot {
		return
	}
	if m.GuildID == "" {
		return
	}
	if m.GuildID != b.execGuildID && m.GuildID != b.communityGuildID {
		return
	}

	if !isOfficePresenceQuery(m.Content) {
		return
	}

	result, err := b.services.Session.ListSessions(query.SessionFilter{ActiveOnly: true}, false)
	if err != nil {
		log.Error("failed to fetch active sessions for presence query", "err", err)
		return
	}

	sessions, ok := result.([]*repository.SessionWithUser)
	if !ok {
		log.Error("unexpected type from ListSessions", "type", fmt.Sprintf("%T", result))
		return
	}

	lines := make([]string, 0, len(sessions))
	for _, session := range sessions {
		lines = append(lines, "• **"+session.UserName+"** (<t:"+strconv.FormatInt(session.CheckIn.Unix(), 10)+":R>)")
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🏢 Office Presence",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       ColorGrey,
		Description: "No one is currently in the office.",
	}

	if len(lines) > 0 {
		embed.Description = "**Currently in office:**\n" + strings.Join(lines, "\n")
		embed.Color = ColorGreen
	}

	msg, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		log.Error("failed to send presence embed", "err", err)
		return
	}

	go func(channelID, messageID string) {
		time.Sleep(5 * time.Minute)
		_ = s.ChannelMessageDelete(channelID, messageID)
	}(m.ChannelID, msg.ID)
}

func isOfficePresenceQuery(content string) bool {
	text := strings.ToLower(strings.TrimSpace(content))
	if text == "" {
		return false
	}

	queries := []string{
		"who's in the office",
		"whos in the office",
		"who is in the office",
		"anyone in the office",
		"anyone in office",
		"anyone in the office",
		"who's in office",
		"whos in office",
		"who is in office",
	}

	for _, q := range queries {
		if strings.Contains(text, q) {
			return true
		}
	}

	return false
}

// handleCommand processes slash commands and routes them to handlers.
// Routing:
//   - "setup", "ping": Handled directly
//   - "user-*": Delegated to userCommands.HandleCommand()
//   - "session-*": Delegated to sessionCommands.HandleCommand()
//   - "checkin", "checkout", "checkout-all": Delegated to attendanceCommands.HandleCommand()
//   - "help": Delegated to helpCommands.HandleCommand()
func (b *Bot) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := i.ApplicationCommandData()

	if cmd.Name != "setup" && i.GuildID != b.execGuildID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Commands are only available in the exec server.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	switch cmd.Name {
	case "setup":
		b.handleSetupCommand(s, i)
	case "ping":
		b.handlePingCommand(s, i)
	case "help":
		b.helpCommands.HandleCommand(s, i, cmd.Name)
	case "stats":
		b.statsCommands.HandleCommand(s, i, cmd.Name)
	case "mystats":
		b.myStatsCommands.HandleCommand(s, i, cmd.Name)
	case "reports-toggle":
		if b.reportsToggleCmd != nil {
			b.reportsToggleCmd.Handle(s, i)
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Reports feature is not available",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	case "scan-history":
		b.scanHistoryCommands.HandleCommand(s, i)
	case "checkin", "checkout", "checkout-all":
		b.attendanceCommands.HandleCommand(s, i, cmd.Name)
	default:
		// Route to user commands if it starts with "user-"
		if len(cmd.Name) > 5 && cmd.Name[:5] == "user-" {
			b.userCommands.HandleCommand(s, i, cmd.Name)
		}
		// Route to session commands if it starts with "session-"
		if len(cmd.Name) > 8 && cmd.Name[:8] == "session-" {
			b.sessionCommands.HandleCommand(s, i, cmd.Name)
		}
	}
}

// handleButtonClick processes button interactions from dashboard messages.
// Supported buttons:
//   - "refresh_btn": Manually refresh dashboard (with cooldown)
//   - "leaving_btn": Check out current user from office
//   - "view_my_stats_*": Show personal stats for report period
func (b *Bot) handleButtonClick(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	switch customID {
	case "refresh_btn":
		// Log who clicked the refresh button
		user := interactionUser(i)
		userID := "unknown"
		userName := "unknown"
		if user != nil {
			userID = user.ID
			userName = user.Username
		}
		log.Info("refresh button clicked", "user_id", userID, "username", userName, "channel_id", i.ChannelID, "guild_id", i.GuildID)

		// Check if the channel is on cooldown
		if b.IsRefreshOnCooldown(i.ChannelID) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⏱️ Please wait before refreshing again",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		// Update refresh time
		b.UpdateRefreshTime(i.ChannelID)
		// Render dashboard in background (non-blocking)
		go b.renderDashboard(i.ChannelID)

	case "leaving_btn":
		// Get user's Discord ID from interaction
		user := interactionUser(i)
		if user == nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Unable to identify user.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		userDiscordID := user.ID
		userName := user.Username
		log.Info("leaving button clicked", "user_id", userDiscordID, "username", userName, "channel_id", i.ChannelID, "guild_id", i.GuildID)

		// Acknowledge silently — no ephemeral reply will be shown
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})

		// Look up user by Discord ID and check them out
		go func() {
			// Find user by Discord ID
			user, err := b.services.User.GetUserByDiscordID(userDiscordID)
			if err != nil {
				log.Error("failed to find user by discord id", "discord_id", userDiscordID, "error", err)
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "❌ Could not find your user account. Please contact an admin.",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				return
			}

			// Check out the session
			err = b.services.Session.CheckOutUserWithMethod(user.ID, repository.CheckOutMethodDiscord)
			if err != nil {
				log.Error("failed to check out session", "user_id", user.ID, "error", err)
				message := "❌ Failed to check out. Please try again."
				if errors.Is(err, service.ErrNoOpenSession) {
					message = "❌ No active session found. You may not be checked in."
				}
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: message,
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				return
			}

			// Trigger dashboard update in background
			b.TriggerRender()
		}()

	default:
		// Check if it's a view_my_stats button
		if len(customID) > 13 && customID[:13] == "view_my_stats" {
			b.handleViewMyStatsButton(s, i, customID)
		}
	}
}

// handleViewMyStatsButton processes "View My Stats" button clicks from reports
// CustomID format: "view_my_stats_<reportType>_<period>"
// Examples: "view_my_stats_weekly_2026-W10", "view_my_stats_monthly_2026-02"
func (b *Bot) handleViewMyStatsButton(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Get user
	user := interactionUser(i)
	if user == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Unable to identify user.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Defer response
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	// Process in background
	go func() {
		// Look up user
		appUser, err := b.services.User.GetUserByDiscordID(user.ID)
		if err != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("❌ Could not find your user account. Please contact an admin."),
			})
			return
		}

		// Parse customID to get report type and period
		// Format: view_my_stats_<type>_<period>
		parts := strings.Split(customID, "_")
		if len(parts) < 4 {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("❌ Invalid button data."),
			})
			return
		}

		reportType := parts[3] // "weekly" or "monthly"
		period := strings.Join(parts[4:], "_")

		// Calculate date range based on report type and period
		var start, end time.Time
		var periodLabel string

		switch reportType {
		case "weekly":
			// Parse ISO week format: 2026-W10
			var year, week int
			_, err := fmt.Sscanf(period, "%d-W%d", &year, &week)
			if err != nil {
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: stringPtr("❌ Invalid week format."),
				})
				return
			}

			// Calculate date range for that ISO week.
			// Jan 4 is always in ISO week 1, so use it to find the Monday of week 1.
			jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.Local)
			weekdayJan4 := int(jan4.Weekday())
			if weekdayJan4 == 0 {
				weekdayJan4 = 7 // Sunday = 7
			}
			start = jan4.AddDate(0, 0, (1-weekdayJan4)+(week-1)*7)
			end = start.AddDate(0, 0, 7).Add(-1 * time.Second)
			periodLabel = period

		case "monthly":
			// Parse month format: 2026-02
			var year, month int
			_, err := fmt.Sscanf(period, "%d-%d", &year, &month)
			if err != nil {
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: stringPtr("❌ Invalid month format."),
				})
				return
			}

			start = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
			end = start.AddDate(0, 1, 0).Add(-1 * time.Second)
			periodLabel = period
		default:
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("❌ Unknown report type."),
			})
			return
		}

		// Fetch user stats for that period
		userStats, err := b.services.Stats.GetUserStatsWithAutoCheckout(appUser.ID, start, end, true)
		if err != nil {
			log.Error("failed to get user stats", "user_id", appUser.ID, "error", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("❌ Failed to fetch your stats."),
			})
			return
		}

		// Fetch period stats for context
		periodStats, err := b.services.Stats.GetPeriodStats(start, end, 0, "hours", true)
		if err != nil {
			log.Error("failed to get period stats", "error", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("❌ Failed to fetch report totals."),
			})
			return
		}

		// Fetch leaderboards for ranking
		allByHours, err := b.services.Stats.GetAllUserStatsForPeriodWithAutoCheckout(start, end, "hours", true)
		if err != nil {
			log.Error("failed to get hours leaderboard", "error", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("❌ Failed to fetch rankings."),
			})
			return
		}
		allByVisits, err := b.services.Stats.GetAllUserStatsForPeriodWithAutoCheckout(start, end, "visits", true)
		if err != nil {
			log.Error("failed to get visits leaderboard", "error", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("❌ Failed to fetch rankings."),
			})
			return
		}

		// Compute ranks
		hoursRank, visitsRank := 0, 0
		for idx, u := range allByHours {
			if u.UserID == appUser.ID {
				hoursRank = idx + 1
				break
			}
		}
		for idx, u := range allByVisits {
			if u.UserID == appUser.ID {
				visitsRank = idx + 1
				break
			}
		}
		participantCount := len(allByHours)
		if len(allByVisits) > participantCount {
			participantCount = len(allByVisits)
		}

		// Compute percentages
		hoursPct := 0.0
		if periodStats.TotalHours > 0 {
			hoursPct = (userStats.TotalHours / periodStats.TotalHours) * 100
		}
		visitsPct := 0.0
		if periodStats.TotalVisits > 0 {
			visitsPct = (float64(userStats.VisitCount) / float64(periodStats.TotalVisits)) * 100
		}

		// Format helper values
		rankStr := func(rank, total int) string {
			if rank == 0 || total == 0 {
				return "N/A"
			}
			return fmt.Sprintf("%d / %d", rank, total)
		}
		busiestDayStr := "N/A"
		if userStats.BusiestDay != "" && userStats.BusiestDayHours > 0 {
			busiestDayStr = fmt.Sprintf("%s (%.1f hrs)", userStats.BusiestDay, userStats.BusiestDayHours)
		}
		firstVisitStr, lastVisitStr := "N/A", "N/A"
		if userStats.FirstVisit != nil {
			firstVisitStr = userStats.FirstVisit.Local().Format("2006-01-02 15:04")
		}
		if userStats.LastVisit != nil {
			lastVisitStr = userStats.LastVisit.Local().Format("2006-01-02 15:04")
		}

		embed := &discordgo.MessageEmbed{
			Title:       "📈 My Stats",
			Description: fmt.Sprintf("Period: %s", periodLabel),
			Color:       0x3498DB,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Hours", Value: fmt.Sprintf("%.1f hrs (%.1f%%)", userStats.TotalHours, hoursPct), Inline: true},
				{Name: "Visits", Value: fmt.Sprintf("%d (%.1f%%)", userStats.VisitCount, visitsPct), Inline: true},
				{Name: "Active Days", Value: fmt.Sprintf("%d", userStats.ActiveDays), Inline: true},
				{Name: "Busiest Day", Value: busiestDayStr, Inline: true},
				{Name: "Avg Duration", Value: fmt.Sprintf("%.1f hrs", userStats.AvgDuration), Inline: true},
				{Name: "First Visit", Value: firstVisitStr, Inline: true},
				{Name: "Last Visit", Value: lastVisitStr, Inline: true},
				{Name: "Rank (Hours)", Value: rankStr(hoursRank, participantCount), Inline: true},
				{Name: "Rank (Visits)", Value: rankStr(visitsRank, participantCount), Inline: true},
			},
			Timestamp: end.Format(time.RFC3339),
		}

		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
	}()
}
