// Package discord provides Discord bot integration for the Office Tracker.
// It handles slash commands, interactions, and dashboard rendering.
package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// handleSetupCommand creates a dashboard in the current channel.
// Only available in configured guild channels (exec and community).
func (b *Bot) handleSetupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Log who executed the command
	user := interactionUser(i)
	userID := "unknown"
	userName := "unknown"
	if user != nil {
		userID = user.ID
		userName = user.Username
	}
	log.Info("setup command executed", "user_id", userID, "username", userName, "guild_id", i.GuildID, "channel_id", i.ChannelID)

	// Log warning if setup is being used in an unexpected channel
	if !b.IsExecChannel(i.ChannelID) && !b.IsCommunityChannel(i.ChannelID) {
		log.Warn("setup command used in non-configured channel", "channel_id", i.ChannelID, "guild_id", i.GuildID)
	}

	// Defer response immediately
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	// Create the initial embed
	embed := &discordgo.MessageEmbed{
		Title:       "🏢 Office Presence",
		Description: "No one is currently in the office.",
		Color:       ColorGrey,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Last update: %s", "now"),
		},
	}

	// Buttons for dashboard interaction
	buttons := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Refresh 🔄",
					Style:    discordgo.SecondaryButton,
					CustomID: "refresh_btn",
				},
			},
		},
	}

	// Send the message
	msg, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: buttons,
	})
	if err != nil {
		log.Error("failed to create dashboard message", "error", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr("Failed to create dashboard"),
		})
		return
	}

	// Store the dashboard by channel ID
	b.dashboardsMu.Lock()
	b.dashboards[i.ChannelID] = &Dashboard{
		ChannelID: i.ChannelID,
		MessageID: msg.ID,
		GuildID:   i.GuildID,
	}
	b.dashboardsMu.Unlock()

	log.Info("dashboard created", "channel_id", i.ChannelID, "message_id", msg.ID, "guild_id", i.GuildID)

	// Render the dashboard immediately with current data
	b.renderDashboard(i.ChannelID)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: stringPtr("✅ Dashboard created in " + fmt.Sprintf("<#%s>", i.ChannelID)),
	})
}

// handlePingCommand responds with a pong message.
// Used for bot health checks.
func (b *Bot) handlePingCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Log who executed the command
	user := interactionUser(i)
	userID := "unknown"
	userName := "unknown"
	if user != nil {
		userID = user.ID
		userName = user.Username
	}
	log.Info("ping command executed", "user_id", userID, "username", userName, "guild_id", i.GuildID)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🏓 Pong!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
