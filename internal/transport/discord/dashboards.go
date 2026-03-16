package discord

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"office/internal/query"
	"office/internal/repository"

	"github.com/bwmarrin/discordgo"
)

// InitializeDashboards finds and registers dashboards by looking for a specific channel name in each guild
// This allows dashboards to persist across restarts without needing to configure channel IDs
func (b *Bot) InitializeDashboards(execGuildID, communityGuildID, channelName string) error {
	if channelName == "" {
		channelName = "office-tracker"
	}
	dashboards := make(map[string]*Dashboard)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Helper function to find dashboard channel by guild and register it (runs in goroutine)
	findAndRegisterDashboard := func(guildID string, serverName string) {
		defer wg.Done()

		if guildID == "" {
			log.Info("no guild ID configured for", "server", serverName)
			return
		}

		// Fetch all channels in the guild
		channels, err := b.session.GuildChannels(guildID)
		if err != nil {
			log.Warn("failed to fetch guild channels", "server", serverName, "guild_id", guildID, "error", err)
			return
		}

		// Find the 'office-tracker' channel
		var dashChannel *discordgo.Channel
		for _, ch := range channels {
			if ch.Name == channelName && ch.Type == discordgo.ChannelTypeGuildText {
				dashChannel = ch
				break
			}
		}

		if dashChannel == nil {
			log.Warn("office-tracker channel not found in guild", "server", serverName, "guild_id", guildID)
			return
		}

		// Thread-safe map write
		mu.Lock()
		dashboards[dashChannel.ID] = &Dashboard{
			ChannelID: dashChannel.ID,
			MessageID: "", // Will be fetched on first render
			GuildID:   guildID,
		}
		mu.Unlock()

		log.Info("dashboard channel found", "server", serverName, "guild_id", guildID, "channel_id", dashChannel.ID, "channel_name", channelName)

		// Store the channel IDs for later use
		switch serverName {
		case "exec":
			b.execChannelID = dashChannel.ID
		case "community":
			b.communityChannelID = dashChannel.ID
		}
	}

	// Search for dashboards in parallel
	wg.Add(2)
	go findAndRegisterDashboard(execGuildID, "exec")
	go findAndRegisterDashboard(communityGuildID, "community")
	wg.Wait()

	if len(dashboards) == 0 {
		log.Warn("no office-tracker channels found in any guild")
		return nil
	}

	b.dashboardsMu.Lock()
	for k, v := range dashboards {
		b.dashboards[k] = v
	}
	b.dashboardsMu.Unlock()

	return nil
}

// renderAllDashboards updates all registered dashboards with current office presence
func (b *Bot) renderAllDashboards() {
	b.dashboardsMu.RLock()
	dashboards := make(map[string]*Dashboard)
	for k, v := range b.dashboards {
		dashboards[k] = v
	}
	b.dashboardsMu.RUnlock()

	// Render all dashboards in parallel
	var wg sync.WaitGroup
	for channelID := range dashboards {
		wg.Add(1)
		go func(cid string) {
			defer wg.Done()
			b.renderDashboard(cid)
		}(channelID)
	}
	wg.Wait()
}

// renderDashboard updates a specific dashboard with current data
func (b *Bot) renderDashboard(channelID string) {
	// Fetch active sessions
	result, err := b.services.Session.ListSessions(query.SessionFilter{ActiveOnly: true, OrderBy: "asc", SortBy: "check_in"}, false)
	if err != nil {
		log.Error("failed to fetch active sessions", "channel_id", channelID, "error", err)
		return
	}

	// Cast result to []*SessionWithUser
	sessions, ok := result.([]*repository.SessionWithUser)
	if !ok {
		log.Error("unexpected type from ListSessions", "channel_id", channelID, "type", fmt.Sprintf("%T", result))
		return
	}

	// Build the member list
	var lines []string
	for _, session := range sessions {
		// Format: "• **Name** (time ago)"
		lines = append(lines, fmt.Sprintf("• **%s** (<t:%d:R>)", session.UserName, session.CheckIn.Unix()))
	}

	// Create embed
	footerText := fmt.Sprintf("Last update: %s", time.Now().Format("15:04:05"))
	if b.services != nil && b.services.Environment != nil {
		if reading, ok := b.services.Environment.GetLatest(); ok {
			if b.services.Environment.IsFresh(reading, 0) {
				footerText = fmt.Sprintf("%s • %.1f°C", footerText, reading.TemperatureC)
			}
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "🏢 Office Presence",
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}
	if len(lines) == 0 {
		embed.Description = "No one is currently in the office."
		embed.Color = ColorGrey
	} else {
		embed.Description = "**Currently in office:**\n" + strings.Join(lines, "\n")
		embed.Color = ColorGreen
	}

	// Get dashboard info
	b.dashboardsMu.RLock()
	dash, exists := b.dashboards[channelID]
	b.dashboardsMu.RUnlock()

	if !exists {
		log.Warn("attempted to render non-existent dashboard", "channel_id", channelID)
		return
	}

	// If MessageID is empty, retrieve the latest message sent by the bot in the channel
	if dash.MessageID == "" {
		messages, err := b.session.ChannelMessages(channelID, 50, "", "", "")
		if err != nil {
			log.Warn("failed to fetch channel messages", "channel_id", channelID, "error", err)
		} else {
			botID := b.session.State.User.ID
			for _, msg := range messages {
				if msg.Author != nil && msg.Author.ID == botID {
					dash.MessageID = msg.ID
					log.Info("dashboard message found", "channel_id", channelID, "message_id", dash.MessageID)
					break
				}
			}
		}
	}
	// Build dashboard buttons
	components := []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Refresh 🔄",
			Style:    discordgo.SecondaryButton,
			CustomID: "refresh_btn",
		},
	}
	if b.IsExecChannel(channelID) {
		components = append(components, discordgo.Button{
			Label:    "Leave 🟥",
			Style:    discordgo.DangerButton,
			CustomID: "leaving_btn",
		})
	}

	buttons := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: components},
	}

	// Edit the message
	embeds := []*discordgo.MessageEmbed{embed}
	_, err = b.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    dash.ChannelID,
		ID:         dash.MessageID,
		Embeds:     &embeds,
		Components: &buttons,
	})

	if err != nil {
		log.Warn("failed to update dashboard message", "channel_id", channelID, "error", err)
		return
	}

	log.Debug("dashboard rendered", "channel_id", channelID, "members_count", len(sessions))
}
