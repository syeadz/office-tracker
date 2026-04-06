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
		Name:                     "reports",
		Description:              "Check or update weekly/monthly scheduled reports (Admin only)",
		DefaultMemberPermissions: &adminPerm,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "weekly",
				Description: "Set weekly reports enabled (true) or disabled (false)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "monthly",
				Description: "Set monthly reports enabled (true) or disabled (false)",
				Required:    false,
			},
		},
	}
}

// Handle processes the reports command
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

	if c.reports == nil {
		respondEphemeral(s, i, "❌ Reports service is not configured.")
		return
	}

	weeklyProvided := false
	weeklyRequested := false
	monthlyProvided := false
	monthlyRequested := false

	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "weekly":
			weeklyProvided = true
			weeklyRequested = opt.BoolValue()
		case "monthly":
			monthlyProvided = true
			monthlyRequested = opt.BoolValue()
		}
	}

	statusOnly := !weeklyProvided && !monthlyProvided

	if !statusOnly {
		if !c.reports.IsAvailable() {
			respondEphemeral(s, i, "❌ Reports are fully disabled (missing startup guild/channel configuration), cannot toggle.")
			return
		}

		if weeklyProvided {
			c.reports.SetWeeklyEnabled(weeklyRequested)
		}
		if monthlyProvided {
			c.reports.SetMonthlyEnabled(monthlyRequested)
		}
	}

	available := c.reports.IsAvailable()
	weeklyEnabled := c.reports.IsWeeklyEnabled()
	monthlyEnabled := c.reports.IsMonthlyEnabled()
	overallEnabled := c.reports.IsEnabled()

	title := "ℹ️ Reports Status"
	description := "Current scheduled reports configuration."
	requested := "status-only"
	if weeklyProvided || monthlyProvided {
		requested = ""
		if weeklyProvided {
			requested = requested + fmt.Sprintf("weekly=%s", statusLabel(weeklyRequested))
		}
		if monthlyProvided {
			if requested != "" {
				requested = requested + ", "
			}
			requested = requested + fmt.Sprintf("monthly=%s", statusLabel(monthlyRequested))
		}
		title = "✅ Reports Settings Updated"
		description = "Applied requested weekly/monthly settings."
	}
	if !available {
		title = "⚠️ Reports Unavailable"
		description = "Reports are fully disabled (missing startup guild/channel configuration)."
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       getColorForStatus(available, overallEnabled),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Availability",
				Value:  statusLabel(available),
				Inline: true,
			},
			{
				Name:   "Overall",
				Value:  statusLabel(overallEnabled),
				Inline: true,
			},
			{
				Name:   "Weekly",
				Value:  statusLabel(weeklyEnabled),
				Inline: true,
			},
			{
				Name:   "Monthly",
				Value:  statusLabel(monthlyEnabled),
				Inline: true,
			},
			{
				Name:   "Requested",
				Value:  requested,
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
func getColorForStatus(available bool, enabled bool) int {
	if !available {
		return 0xf1c40f
	}
	if enabled {
		return 0x00ff00
	}
	return 0xff0000
}

func statusLabel(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
