package commands

import (
	"fmt"

	"office/internal/service"

	"github.com/bwmarrin/discordgo"
)

// ReportsToggleCommand provides functionality to enable/disable scheduled reports
type ReportsToggleCommand struct {
	reports *service.ReportsService
}

// NewReportsToggleCommand creates a new reports toggle command
func NewReportsToggleCommand(reports *service.ReportsService) *ReportsToggleCommand {
	return &ReportsToggleCommand{
		reports: reports,
	}
}

// Definition returns the Discord slash command definition
func (c *ReportsToggleCommand) Definition() *discordgo.ApplicationCommand {
	adminPerm := int64(discordgo.PermissionAdministrator)

	return &discordgo.ApplicationCommand{
		Name:                     "reports-toggle",
		Description:              "Enable or disable scheduled weekly and monthly reports (Admin only)",
		DefaultMemberPermissions: &adminPerm,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Enable (true) or disable (false) scheduled reports",
				Required:    true,
			},
		},
	}
}

// Handle processes the reports-toggle command
func (c *ReportsToggleCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if user has administrator permission
	if !hasAdminPermission(i.Member) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You do not have permission to use this command. Administrator role required.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Parse the enabled option
	enabled := false
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "enabled" {
			enabled = opt.BoolValue()
			break
		}
	}

	c.reports.SetEnabled(enabled)

	status := "disabled"
	emoji := "🔴"
	if enabled {
		status = "enabled"
		emoji = "✅"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Reports %s", emoji, status),
		Description: fmt.Sprintf("Scheduled weekly and monthly reports have been **%s**", status),
		Color:       getColorForStatus(enabled),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Status",
				Value:  status,
				Inline: true,
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// hasAdminPermission checks if the member has administrator permission
func hasAdminPermission(member *discordgo.Member) bool {
	if member == nil {
		return false
	}

	// Check if user has administrator permission bit
	permissions := int64(member.Permissions)
	adminPermission := int64(discordgo.PermissionAdministrator)
	return (permissions & adminPermission) == adminPermission
}

// getColorForStatus returns appropriate color for embed based on status
func getColorForStatus(enabled bool) int {
	if enabled {
		return 0x00ff00 // Green
	}
	return 0xff0000 // Red
}
