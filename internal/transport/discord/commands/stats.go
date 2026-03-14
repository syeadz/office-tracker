// Package commands contains Discord slash command handlers for the office tracker bot.
package commands

import (
	"fmt"
	"strings"
	"time"

	"office/internal/domain"
	"office/internal/logging"
	"office/internal/service"

	"github.com/bwmarrin/discordgo"
)

var statsLog = logging.Component("discord.commands.stats")

// StatsCommands handles office stats slash command
// Supported ranges: this_week, last_week, this_month, last_30_days, custom (with from/to)
type StatsCommands struct {
	statsSvc *service.OfficeStatsService
}

// NewStatsCommands creates a new StatsCommands handler
func NewStatsCommands(statsSvc *service.OfficeStatsService) *StatsCommands {
	return &StatsCommands{statsSvc: statsSvc}
}

// GetApplicationCommands returns the slash command definitions for stats
func (sc *StatsCommands) GetApplicationCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "stats",
			Description: "Get office statistics for various time periods",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "range",
					Description: "Time range to report (default: this week)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "This Week", Value: "this_week"},
						{Name: "Last Week", Value: "last_week"},
						{Name: "This Month", Value: "this_month"},
						{Name: "Last 30 Days", Value: "last_30_days"},
						{Name: "Custom", Value: "custom"},
					},
				},
				{
					Name:        "from",
					Description: "Start date (YYYY-MM-DD) for custom range",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
				{
					Name:        "to",
					Description: "End date (YYYY-MM-DD) for custom range",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
				{
					Name:        "top",
					Description: "Top users to include (default: 10, max: 25)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
				{
					Name:        "rank_by",
					Description: "Rank leaderboard by total hours or visits (default: hours)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Total Hours", Value: "hours"},
						{Name: "Visits / Sessions", Value: "visits"},
					},
				},
				{
					Name:        "include_auto_checkout",
					Description: "Include non-RFID checkouts (including 04:00 auto-checkout) (default: false)",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    false,
				},
			},
		},
	}
}

// HandleCommand routes stats commands
func (sc *StatsCommands) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmdName string) {
	if cmdName != "stats" {
		return
	}
	statsLog.Info("stats command received", "user_id", interactionUserID(i), "username", interactionUsername(i))

	if sc.statsSvc == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Stats service not configured",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	rangeType := "this_week"
	fromStr := ""
	toStr := ""
	topLimit := int64(10)
	rankBy := "hours"
	includeNonRFIDCheckouts := false

	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "range":
			rangeType = strings.ToLower(opt.StringValue())
		case "from":
			fromStr = opt.StringValue()
		case "to":
			toStr = opt.StringValue()
		case "top":
			if opt.IntValue() > 0 {
				topLimit = opt.IntValue()
			}
		case "rank_by":
			rankBy = strings.ToLower(opt.StringValue())
		case "include_auto_checkout":
			includeNonRFIDCheckouts = opt.BoolValue()
		}
	}

	if topLimit > 25 {
		topLimit = 25
	}

	excludeAutoCheckout := !includeNonRFIDCheckouts

	var report *domain.PeriodStats
	var err error
	now := time.Now()

	switch rangeType {
	case "this_week":
		report, err = sc.statsSvc.GetWeeklyReport(rankBy, excludeAutoCheckout)
	case "last_week":
		weekday := now.Weekday()
		var daysToMonday int
		if weekday == time.Sunday {
			daysToMonday = 6
		} else {
			daysToMonday = int(weekday) - 1
		}
		lastWeekStart := now.AddDate(0, 0, -daysToMonday-7).Truncate(24 * time.Hour)
		lastWeekEnd := lastWeekStart.AddDate(0, 0, 7).Add(-1 * time.Second)
		report, err = sc.statsSvc.GetCustomReport(lastWeekStart, lastWeekEnd, int(topLimit), rankBy, excludeAutoCheckout)
	case "this_month":
		report, err = sc.statsSvc.GetMonthlyReport(now.Year(), now.Month(), rankBy, excludeAutoCheckout)
	case "last_30_days":
		from := now.AddDate(0, 0, -30)
		report, err = sc.statsSvc.GetCustomReport(from, now, int(topLimit), rankBy, excludeAutoCheckout)
	case "custom":
		if fromStr == "" || toStr == "" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Custom range requires both from and to (YYYY-MM-DD)",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		from, parseErr := time.ParseInLocation("2006-01-02", fromStr, time.Local)
		if parseErr != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Invalid from date. Use YYYY-MM-DD",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		to, parseErr := time.ParseInLocation("2006-01-02", toStr, time.Local)
		if parseErr != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Invalid to date. Use YYYY-MM-DD",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		report, err = sc.statsSvc.GetCustomReport(from, to, int(topLimit), rankBy, excludeAutoCheckout)
	default:
		// Default to this week
		report, err = sc.statsSvc.GetWeeklyReport(rankBy, excludeAutoCheckout)
	}

	if err != nil {
		statsLog.Error("failed to get stats", "range", rangeType, "err", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Failed to fetch stats.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	embed := buildStatsEmbed(report)

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		statsLog.Error("failed to respond to stats command", "err", err)
	}
}

func buildStatsEmbed(report *domain.PeriodStats) *discordgo.MessageEmbed {
	rankLabel := leaderboardRankLabel(report.RankBy)
	if report.RankBy == "" {
		rankLabel = leaderboardRankLabel("hours")
	}

	fields := []*discordgo.MessageEmbedField{
		{Name: "Total Hours", Value: fmt.Sprintf("%.1f", report.TotalHours), Inline: true},
		{Name: "Total Visits", Value: fmt.Sprintf("%d", report.TotalVisits), Inline: true},
		{Name: "Unique Users", Value: fmt.Sprintf("%d", report.UniqueUsers), Inline: true},
		{Name: "Active Days", Value: fmt.Sprintf("%d", report.ActiveDays), Inline: true},
		{Name: "Busiest Day", Value: formatOfficeBusiestDay(report.BusiestDay, report.BusiestDayUsers), Inline: true},
		{Name: "Peak Occupancy", Value: fmt.Sprintf("%d users", report.PeakOccupancy), Inline: true},
		{Name: "Avg Hours / User", Value: fmt.Sprintf("%.2f", report.AveragePerUser), Inline: true},
		{Name: "Leaderboard Rank By", Value: rankLabel, Inline: true},
	}

	if len(report.TopUsers) > 0 {
		leaderboard := "```\n"
		for i, user := range report.TopUsers {
			rank := i + 1
			leaderboard += formatLeaderboardEntry(rank, user, report.RankBy)
		}
		leaderboard += "```\n"
		fields = append(fields, &discordgo.MessageEmbedField{Name: fmt.Sprintf("Leaderboard (%s)", rankLabel), Value: leaderboard})
	}

	return &discordgo.MessageEmbed{
		Title:       "📊 Office Stats",
		Description: fmt.Sprintf("Period: %s", report.Period),
		Color:       0x3498DB,
		Fields:      fields,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
}

func leaderboardRankLabel(rankBy string) string {
	if rankBy == "visits" {
		return "Visits / Sessions"
	}

	return "Total Hours"
}

func formatLeaderboardEntry(rank int, user domain.UserStats, rankBy string) string {
	if rankBy == "visits" {
		return fmt.Sprintf("%d. %s - %d visits (%.1f hrs)\n", rank, user.Name, user.VisitCount, user.TotalHours)
	}

	return fmt.Sprintf("%d. %s - %.1f hrs (%d visits)\n", rank, user.Name, user.TotalHours, user.VisitCount)
}

func formatOfficeBusiestDay(day string, userCount int64) string {
	if day == "" || userCount == 0 {
		return "N/A"
	}

	return fmt.Sprintf("%s (%d users)", day, userCount)
}
