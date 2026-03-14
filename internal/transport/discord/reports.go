package discord

import (
	"fmt"
	"strings"
	"time"

	"office/internal/domain"
	"office/internal/logging"

	"github.com/bwmarrin/discordgo"
)

var reportsLog = logging.Component("transport.discord.reports")

// ReportsDelivery handles Discord-specific report formatting and delivery
type ReportsDelivery struct {
	session   *discordgo.Session
	channelID string
}

// NewReportsDelivery creates a new Discord reports delivery service
func NewReportsDelivery(session *discordgo.Session, channelID string) *ReportsDelivery {
	return &ReportsDelivery{
		session:   session,
		channelID: channelID,
	}
}

// SendPeriodReport formats and sends a period report as a Discord embed
func (d *ReportsDelivery) SendPeriodReport(report *domain.PeriodReport, reportType string) error {
	if d.channelID == "" {
		return fmt.Errorf("reports channel not configured")
	}

	embed := d.formatReportEmbed(report, reportType)

	// Add "View My Stats" button
	buttonID := fmt.Sprintf("view_my_stats_%s_%s", reportType, report.Period)
	buttons := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "📊 View My Stats",
					Style:    discordgo.PrimaryButton,
					CustomID: buttonID,
				},
			},
		},
	}

	_, err := d.session.ChannelMessageSendComplex(d.channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: buttons,
	})
	if err != nil {
		return fmt.Errorf("send discord message: %w", err)
	}

	reportsLog.Info("period report sent to discord", "channel_id", d.channelID, "type", reportType)
	return nil
}

// formatReportEmbed creates a rich Discord embed for periodic reports
func (d *ReportsDelivery) formatReportEmbed(report *domain.PeriodReport, reportType string) *discordgo.MessageEmbed {
	// Build header with emoji and timestamp
	title := "📊 Weekly Office Report"
	if reportType == "monthly" {
		title = "📊 Monthly Office Report"
	}

	// Format date range
	dateRange := fmt.Sprintf("%s - %s",
		report.StartDate.Format("Jan 2"),
		report.EndDate.Format("Jan 2, 2006"),
	)

	// Build description with key metrics
	description := fmt.Sprintf("**%s**\n\n", dateRange)

	// Overview section
	overview := []string{
		fmt.Sprintf("⏱️ **Total Hours:** %.1f hrs", report.TotalHours),
		fmt.Sprintf("👥 **Unique Users:** %d", report.UniqueUsers),
		fmt.Sprintf("🚪 **Total Visits:** %d", report.TotalVisits),
		fmt.Sprintf("📅 **Active Days:** %d/7", report.ActiveDays),
		fmt.Sprintf("🔥 **Busiest Day:** %s", formatReportBusiestDay(report.BusiestDay, report.BusiestDayUsers)),
		fmt.Sprintf("🏢 **Peak Occupancy:** %d users", report.PeakOccupancy),
	}

	// Add comparison if available
	if report.HasComparison {
		if report.HoursChange != 0 {
			emoji := "📈"
			if report.HoursChange < 0 {
				emoji = "📉"
			}
			comparisonLabel := "vs Last Week"
			if reportType == "monthly" {
				comparisonLabel = "vs Last Month"
			}
			overview = append(overview, fmt.Sprintf("%s **%s:** %+.1f%%", emoji, comparisonLabel, report.HoursChange))
		}
	}

	description += strings.Join(overview, "\n")

	// Build top users leaderboard
	var leaderboard strings.Builder
	leaderboard.WriteString("\n\n**🏆 Top Contributors**\n")

	if len(report.TopUsers) == 0 {
		leaderboard.WriteString("_No activity this week_")
	} else {
		medals := []string{"🥇", "🥈", "🥉"}
		for i, user := range report.TopUsers {
			if i >= 10 { // Limit to top 10
				break
			}

			medal := ""
			if i < len(medals) {
				medal = medals[i] + " "
			} else {
				medal = fmt.Sprintf("`%2d.` ", i+1)
			}

			leaderboard.WriteString(fmt.Sprintf("%s**%s** • %.1f hrs • %d visits\n",
				medal,
				user.Name,
				user.TotalHours,
				user.VisitCount,
			))
		}
	}

	description += leaderboard.String()

	// Determine color based on activity level
	color := 0x5865F2 // Discord blurple
	if report.TotalHours >= 100 {
		color = 0x57F287 // Green - high activity
	} else if report.TotalHours >= 50 {
		color = 0xFEE75C // Yellow - moderate activity
	} else if report.TotalHours < 20 {
		color = 0xED4245 // Red - low activity
	}

	footerText := "Office Tracker • Weekly Report"
	if reportType == "monthly" {
		footerText = "Office Tracker • Monthly Report"
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Timestamp:   report.GeneratedAt.Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}

	return embed
}

func formatReportBusiestDay(day string, userCount int64) string {
	if day == "" || userCount == 0 {
		return "N/A"
	}

	return fmt.Sprintf("%s (%d users)", day, userCount)
}

// SetChannelID updates the target channel for reports
func (d *ReportsDelivery) SetChannelID(channelID string) {
	d.channelID = channelID
	reportsLog.Info("reports channel updated", "channel_id", channelID)
}
