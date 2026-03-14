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

var myStatsLog = logging.Component("discord.commands.mystats")

// MyStatsCommands handles personal stats slash command
// Supported ranges: this_week, last_week, this_month, last_30_days, custom (with from/to)
type MyStatsCommands struct {
	statsSvc *service.OfficeStatsService
	userSvc  *service.UserService
}

// NewMyStatsCommands creates a new MyStatsCommands handler
func NewMyStatsCommands(statsSvc *service.OfficeStatsService, userSvc *service.UserService) *MyStatsCommands {
	return &MyStatsCommands{statsSvc: statsSvc, userSvc: userSvc}
}

// GetApplicationCommands returns the slash command definitions for mystats
func (mc *MyStatsCommands) GetApplicationCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "mystats",
			Description: "Get your personal office stats for various time periods",
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
					Name:        "include_auto_checkout",
					Description: "Include non-RFID checkouts (including 04:00 auto-checkout) (default: false)",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    false,
				},
			},
		},
	}
}

// HandleCommand routes mystats commands
func (mc *MyStatsCommands) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmdName string) {
	if cmdName != "mystats" {
		return
	}

	if mc.statsSvc == nil || mc.userSvc == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Stats service not configured",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

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

	appUser, err := mc.userSvc.GetUserByDiscordID(user.ID)
	if err != nil {
		myStatsLog.Error("failed to find user for mystats", "discord_id", user.ID, "err", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ No linked user account found.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	rangeType := "this_week"
	fromStr := ""
	toStr := ""
	includeNonRFIDCheckouts := false

	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "range":
			rangeType = strings.ToLower(opt.StringValue())
		case "from":
			fromStr = opt.StringValue()
		case "to":
			toStr = opt.StringValue()
		case "include_auto_checkout":
			includeNonRFIDCheckouts = opt.BoolValue()
		}
	}

	excludeAutoCheckout := !includeNonRFIDCheckouts

	var start time.Time
	var end time.Time
	periodLabel := ""
	now := time.Now()

	switch rangeType {
	case "this_week":
		report, err := mc.statsSvc.GetWeeklyReport("hours", excludeAutoCheckout)
		if err != nil {
			respondMyStatsError(s, i, "❌ Failed to fetch this week stats.")
			return
		}
		start = report.StartDate
		end = report.EndDate
		periodLabel = report.Period
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
		start = lastWeekStart
		end = lastWeekEnd
		year, week := lastWeekStart.ISOWeek()
		periodLabel = fmt.Sprintf("%d-W%02d", year, week)
	case "this_month":
		report, err := mc.statsSvc.GetMonthlyReport(now.Year(), now.Month(), "hours", excludeAutoCheckout)
		if err != nil {
			respondMyStatsError(s, i, "❌ Failed to fetch this month stats.")
			return
		}
		start = report.StartDate
		end = report.EndDate
		periodLabel = report.Period
	case "last_30_days":
		start = now.AddDate(0, 0, -30)
		end = now
		periodLabel = "Last 30 Days"
	case "custom":
		if fromStr == "" || toStr == "" {
			respondMyStatsError(s, i, "❌ Custom range requires both from and to (YYYY-MM-DD)")
			return
		}
		parsedFrom, err := time.ParseInLocation("2006-01-02", fromStr, time.Local)
		if err != nil {
			respondMyStatsError(s, i, "❌ Invalid from date. Use YYYY-MM-DD")
			return
		}
		parsedTo, err := time.ParseInLocation("2006-01-02", toStr, time.Local)
		if err != nil {
			respondMyStatsError(s, i, "❌ Invalid to date. Use YYYY-MM-DD")
			return
		}
		start = parsedFrom
		end = parsedTo.AddDate(0, 0, 1).Add(-1 * time.Second)
		periodLabel = fmt.Sprintf("%s to %s", start.Format("2006-01-02"), parsedTo.Format("2006-01-02"))
	default:
		// Default to this week
		report, err := mc.statsSvc.GetWeeklyReport("hours", excludeAutoCheckout)
		if err != nil {
			respondMyStatsError(s, i, "❌ Failed to fetch this week stats.")
			return
		}
		start = report.StartDate
		end = report.EndDate
		periodLabel = report.Period
	}

	userStats, err := mc.statsSvc.GetUserStatsWithAutoCheckout(appUser.ID, start, end, excludeAutoCheckout)
	if err != nil {
		myStatsLog.Error("failed to get user stats", "user_id", appUser.ID, "err", err)
		respondMyStatsError(s, i, "❌ Failed to fetch your stats.")
		return
	}

	periodStats, err := mc.statsSvc.GetPeriodStats(start, end, 0, "hours", excludeAutoCheckout)
	if err != nil {
		myStatsLog.Error("failed to get period stats", "err", err)
		respondMyStatsError(s, i, "❌ Failed to fetch report totals.")
		return
	}

	allByHours, err := mc.statsSvc.GetAllUserStatsForPeriodWithAutoCheckout(start, end, "hours", excludeAutoCheckout)
	if err != nil {
		myStatsLog.Error("failed to get hours leaderboard", "err", err)
		respondMyStatsError(s, i, "❌ Failed to fetch rankings.")
		return
	}

	allByVisits, err := mc.statsSvc.GetAllUserStatsForPeriodWithAutoCheckout(start, end, "visits", excludeAutoCheckout)
	if err != nil {
		myStatsLog.Error("failed to get visits leaderboard", "err", err)
		respondMyStatsError(s, i, "❌ Failed to fetch rankings.")
		return
	}

	hoursRank := findUserRank(allByHours, appUser.ID)
	visitsRank := findUserRank(allByVisits, appUser.ID)
	participantCount := len(allByHours)
	if len(allByVisits) > participantCount {
		participantCount = len(allByVisits)
	}

	hoursPct := 0.0
	if periodStats.TotalHours > 0 {
		hoursPct = (userStats.TotalHours / periodStats.TotalHours) * 100
	}
	visitsPct := 0.0
	if periodStats.TotalVisits > 0 {
		visitsPct = (float64(userStats.VisitCount) / float64(periodStats.TotalVisits)) * 100
	}

	embed := buildMyStatsEmbed(userStats, periodLabel, hoursPct, visitsPct, hoursRank, visitsRank, participantCount)
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		myStatsLog.Error("failed to respond to mystats command", "err", err)
	}
}

func respondMyStatsError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func buildMyStatsEmbed(userStats *domain.UserStats, periodLabel string, hoursPct float64, visitsPct float64, hoursRank int, visitsRank int, participantCount int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "📈 My Stats",
		Description: fmt.Sprintf("Period: %s", periodLabel),
		Color:       0x3498DB,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Hours", Value: fmt.Sprintf("%.1f hrs (%.1f%%)", userStats.TotalHours, hoursPct), Inline: true},
			{Name: "Visits", Value: fmt.Sprintf("%d (%.1f%%)", userStats.VisitCount, visitsPct), Inline: true},
			{Name: "Active Days", Value: fmt.Sprintf("%d", userStats.ActiveDays), Inline: true},
			{Name: "Busiest Day", Value: formatPersonalBusiestDay(userStats.BusiestDay, userStats.BusiestDayHours), Inline: true},
			{Name: "Avg Duration", Value: fmt.Sprintf("%.1f hrs", userStats.AvgDuration), Inline: true},
			{Name: "First Visit", Value: formatOptionalVisitTime(userStats.FirstVisit), Inline: true},
			{Name: "Last Visit", Value: formatOptionalVisitTime(userStats.LastVisit), Inline: true},
			{Name: "Rank (Hours)", Value: formatRank(hoursRank, participantCount), Inline: true},
			{Name: "Rank (Visits)", Value: formatRank(visitsRank, participantCount), Inline: true},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func findUserRank(stats []domain.UserStats, userID int64) int {
	for i, s := range stats {
		if s.UserID == userID {
			return i + 1
		}
	}
	return 0
}

func formatRank(rank int, total int) string {
	if rank == 0 || total == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d / %d", rank, total)
}

func formatOptionalVisitTime(value *time.Time) string {
	if value == nil {
		return "N/A"
	}

	return value.Local().Format("2006-01-02 15:04")
}

func formatPersonalBusiestDay(day string, hours float64) string {
	if day == "" || hours <= 0 {
		return "N/A"
	}

	return fmt.Sprintf("%s (%.1f hrs)", day, hours)
}
