package commands

import (
	"office/internal/logging"

	"github.com/bwmarrin/discordgo"
)

var helpLog = logging.Component("discord.commands.help")

// Color constants for embeds
const (
	ColorBlue = 0x3498DB
)

// HelpCommands provides help information about the bot
type HelpCommands struct{}

// NewHelpCommands creates a new HelpCommands handler
func NewHelpCommands() *HelpCommands {
	return &HelpCommands{}
}

// GetApplicationCommands returns the slash command definitions for help
func (hc *HelpCommands) GetApplicationCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Get help about the office tracker bot and available commands",
		},
	}
}

// HandleCommand routes help commands
func (hc *HelpCommands) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmdName string) {
	if cmdName == "help" {
		hc.handleHelp(s, i)
	}
}

// handleHelp displays help information about the bot
func (hc *HelpCommands) handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	helpLog.Info("help command requested", "user_id", interactionUserID(i), "username", interactionUsername(i))

	embed := &discordgo.MessageEmbed{
		Title:       "📚 Office Tracker Bot Help",
		Description: "Track member office presence with RFID cards and Discord commands.\nUse optional parameters in brackets: `[param]`.",
		Color:       ColorBlue,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "👤 User Commands",
				Value:  "`/user-list [search] [limit] [page] [order] [sort_by]`\nList members (defaults: `limit=10`, `page=1`, `order=asc`, `sort_by=name`)\n\n`/user-get <user_id>`\nGet member details\n\n`/user-create <member> <rfid_uid> [name]`\nCreate member (admin) • `name` defaults to selected member display name\n\n`/user-update <user_id> <name> [member] [rfid_uid]`\nUpdate member (admin) • optional `member` relinks Discord account • optional `rfid_uid` updates card tag\n\n`/user-delete <user_id>`\nDelete member (admin)",
				Inline: false,
			},
			{
				Name:   "📋 Session Commands",
				Value:  "`/session-list [user] [name] [active_only] [limit] [page] [order] [sort_by]`\nList sessions with filters (defaults: `limit=10`, `page=1`, `order=desc`)\n• `user`: mention a Discord member to filter by their sessions\n\n`/session-get <session_id>`\nGet session details\n\n`/session-active`\nShow everyone currently in office",
				Inline: false,
			},
			{
				Name:   "📊 Stats Commands",
				Value:  "`/stats [range] [from] [to] [top] [rank_by] [include_auto_checkout]`\nOffice stats + leaderboard\n• `range`: `this_week` (default), `last_week`, `this_month`, `last_30_days`, `custom`\n• `rank_by`: `hours` (default) or `visits`\n• `top`: default `10`, max `25`\n• `include_auto_checkout`: include non-RFID checkouts (auto/Discord/API)\n• For `custom`, provide `from` and `to` as `YYYY-MM-DD`\n\n`/mystats [range] [from] [to] [include_auto_checkout]`\nYour personal stats + rank (default is RFID-only checkouts)",
				Inline: false,
			},
			{
				Name:   "👋 Check In/Out Commands",
				Value:  "`/checkin <member>`\nManual check-in (admin)\n\n`/checkout <member>`\nCheck out one member\n\n`/checkout-all`\nCheck out everyone currently in office\n\n`/scan-history [limit]`\nView recent RFID scan events (admin) • `limit` default `10`, max `25`",
				Inline: false,
			},
			{
				Name:   "⚙️ Dashboard Commands",
				Value:  "`/setup`\nCreate/refresh the office presence dashboard in this channel",
				Inline: false,
			},
			{
				Name:   "📈 Reports Commands",
				Value:  "`/reports-toggle <enabled>`\nEnable or disable scheduled weekly and monthly reports (admin only)",
				Inline: false,
			},
			{
				Name:   "🧪 Quick Examples",
				Value:  "`/stats range:this_month rank_by:visits top:15`\n`/stats range:custom from:2026-03-01 to:2026-03-10`\n`/user-create member:@alice rfid_uid:ABC123`",
				Inline: false,
			},
			{
				Name:   "\n💡 How It Works",
				Value:  "1. Members check in by scanning their RFID card at the office\n2. Dashboard updates automatically to show who's present\n3. Members can leave by clicking the Leave button or using commands\n4. Auto-checkout occurs at 04:00 local time for any active sessions\n5. All presence data is tracked for reports and analytics",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Commands respond with 🔒 if you lack permissions • Dashboard updates every 5 minutes",
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
