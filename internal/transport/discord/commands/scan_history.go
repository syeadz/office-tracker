package commands

import (
	"fmt"
	"strings"
	"time"

	"office/internal/logging"
	"office/internal/service"

	"github.com/bwmarrin/discordgo"
)

var scanHistoryLog = logging.Component("discord.commands.scan_history")

// ScanHistoryCommands handles the /scan-history slash command.
type ScanHistoryCommands struct {
	attendanceSvc *service.AttendanceService
}

// NewScanHistoryCommands creates a new ScanHistoryCommands handler.
func NewScanHistoryCommands(attendanceSvc *service.AttendanceService) *ScanHistoryCommands {
	return &ScanHistoryCommands{attendanceSvc: attendanceSvc}
}

// GetApplicationCommands returns the slash command definition for scan history.
func (sc *ScanHistoryCommands) GetApplicationCommands() []*discordgo.ApplicationCommand {
	adminPerm := int64(discordgo.PermissionAdministrator)
	return []*discordgo.ApplicationCommand{
		{
			Name:                     "scan-history",
			Description:              "Show recent RFID scan events",
			DefaultMemberPermissions: &adminPerm,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "limit",
					Description: "Number of recent scans to show (default: 10, max: 25)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
					MinValue:    func() *float64 { v := 1.0; return &v }(),
					MaxValue:    25,
				},
			},
		},
	}
}

// HandleCommand handles the /scan-history command.
func (sc *ScanHistoryCommands) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	scanHistoryLog.Info("scan-history command received", "user_id", interactionUserID(i))

	limit := 10
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "limit" {
			if v := int(opt.IntValue()); v > 0 && v <= 25 {
				limit = v
			}
		}
	}

	history := sc.attendanceSvc.GetScanHistory()
	if len(history) == 0 {
		respondEphemeral(s, i, "No recent scans found.")
		return
	}

	// Take the most recent `limit` entries (history is oldest-first)
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	// Build embed rows (most recent first for readability)
	var lines []string
	for idx := len(history) - 1; idx >= 0; idx-- {
		scan := history[idx]
		ts := fmt.Sprintf("<t:%d:R>", scan.Timestamp.Unix())

		var nameStr string
		if scan.Known && scan.UserName != "" {
			nameStr = "**" + scan.UserName + "**"
		} else {
			nameStr = fmt.Sprintf("Unknown (`%s`)", scan.UID)
		}

		var actionStr string
		switch strings.ToLower(scan.Action) {
		case "check_in", "checkin":
			actionStr = "✅ Check-in"
		case "check_out", "checkout":
			actionStr = "🚪 Check-out"
		default:
			if scan.Action != "" {
				actionStr = scan.Action
			} else {
				actionStr = "Scan"
			}
		}

		lines = append(lines, fmt.Sprintf("%s — %s %s", ts, actionStr, nameStr))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📡 Recent RFID Scans",
		Description: strings.Join(lines, "\n"),
		Color:       0x3498DB,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Showing %d of up to 100 recent scans", len(history)),
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
