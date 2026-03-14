# Discord Bot Setup Guide

Complete guide to setting up Discord integration for Office Tracker.

Navigation:

- [README.md](README.md) — documentation index
- [../README.md](../README.md) — project overview and quick start

## Overview

The Discord bot provides:

- Slash commands for check-in/checkout
- User statistics and leaderboards
- Automated dashboards with live presence
- Weekly and monthly activity reports
- Role-based access control

## Prerequisites

- Discord account
- Server admin permissions
- Discord Developer Portal access

## Step 1: Create Discord Application

1. Visit Discord Developer Portal: <https://discord.com/developers/applications>
2. Click **New Application**
3. Enter name: `Office Tracker`
4. Click **Create**

## Step 2: Configure Bot

### Create Bot User

1. Go to **Bot** tab in left sidebar
2. Click **Add Bot**
3. Confirm by clicking **Yes, do it!**

### Bot Settings

Configure these settings:

- **Public Bot**: OFF (recommended)
- **Require OAuth2 Code Grant**: OFF
- **Presence Intent**: ON
- **Server Members Intent**: ON (required for @mentions)
- **Message Content Intent**: OFF (not needed)

### Get Bot Token

1. In **Bot** tab, click **Reset Token**
2. Copy the token (you'll need this for `.env`)
3. ⚠️ **Keep this secret!** Never commit to git.

## Step 3: Set Bot Permissions

### Calculate Permissions

Go to **Bot** tab and enable these permissions:

**General Permissions:**

- Read Messages/View Channels

**Text Permissions:**

- Send Messages
- Send Messages in Threads
- Embed Links
- Attach Files
- Read Message History
- Mention Everyone (optional, for reports)
- Use Slash Commands

**Permissions Integer:** `277025770496`

## Step 4: Invite Bot to Server

### Generate Invite Link

1. Go to **OAuth2** > **URL Generator**
2. Select scopes:
   - `bot`
   - `applications.commands`
3. Select permissions (same as above)
4. Copy generated URL
5. Open URL in browser and select your server
6. Click **Authorize**

## Step 5: Get Guild IDs

### Executive Board Guild ID

For the main server with executive board members:

1. Enable Discord Developer Mode:
   - Open Discord Settings
   - Go to **App Settings** > **Advanced**
   - Enable **Developer Mode**
2. Right-click your server icon
3. Click **Copy Server ID**
4. This is your `DISCORD_EXEC_GUILD_ID`

### Community Guild ID

If you have a separate community server:

1. Right-click community server icon
2. Click **Copy Server ID**
3. This is your `DISCORD_COMMUNITY_GUILD_ID`

*Note: Can be the same as EXEC_GUILD_ID if using one server.*

## Step 6: Configure Channels

### Dashboard Channel

Create a dedicated channel for the live dashboard:

1. Create text channel: `#office-tracker`
2. Set permissions:
   - Bot: Send Messages, Embed Links, Attach Files
   - @everyone: Read Messages (no send)
3. Use channel name in config: `DISCORD_DASHBOARD_CHANNEL_NAME=office-tracker`

### Reports Channel

For weekly and monthly reports (optional):

1. Create channel: `#office-reports`
2. Right-click channel name
3. Click **Copy Channel ID**
4. This is your `DISCORD_REPORTS_CHANNEL_ID`

## Step 7: Configure Application

Update your `.env` file:

```bash
# Required for Discord integration
DISCORD_TOKEN=your_bot_token_here
DISCORD_EXEC_GUILD_ID=123456789012345678
DISCORD_COMMUNITY_GUILD_ID=123456789012345678

# Dashboard configuration
DISCORD_DASHBOARD_CHANNEL_NAME=office-tracker

# Optional: Weekly and monthly reports
DISCORD_REPORTS_CHANNEL_ID=987654321098765432
```

## Step 8: Start Application

```bash
# Development
go run cmd/office/main.go

# Production (systemd)
sudo systemctl restart office-tracker
```

Check logs for successful connection:

```text
Discord bot connected successfully
Registered slash commands for guild: 123456789012345678
```

## Step 9: Test Commands

### Available Commands

Try these in your Discord server:

```text
/help                    - Show all commands
/checkin @member        - Check in a member (admin)
/checkout @member       - Check out a member
/checkout-all           - Check out all active members (admin)
/scan-history [limit]   - View recent RFID scan events (admin)
/mystats                - View your weekly stats
/stats                  - View server leaderboard
/session-list [user]    - List sessions, optionally filtered by @user
```

### Command Permissions

Some commands are admin-only by default (`/setup`, `/checkin`, `/checkout-all`, `/scan-history`, `/reports-toggle`, `/user-create`, `/user-update`, `/user-delete`).

For all other commands, you can further customize permissions:

1. Go to **Server Settings** > **Integrations**
2. Click on **Office Tracker**
3. Configure command permissions:
   - `/checkin` - Restrict to Officers role
   - `/checkout` - Everyone
   - `/mystats` - Everyone
   - `/stats` - Everyone

## Features Configuration

### Automated Dashboards

The bot automatically updates the dashboard channel with:

- Currently present members
- Total hours today
- Active sessions count
- Leaderboard

**Update Frequency:** Every 5 minutes (hardcoded)

**To disable:** Remove `DISCORD_DASHBOARD_CHANNEL_NAME` from `.env`

### Automated Reports

The bot automatically sends periodic activity reports to keep everyone informed about office usage.

#### Weekly Reports

Sent every Monday at 9 AM:

- Total weekly hours
- Most active members
- Comparison with previous week
- Participation stats
- Busiest day (date + user count)
- Peak occupancy

#### Monthly Reports

Sent on the 1st of each month at 9 AM:

- Total monthly hours
- Top contributors
- Comparison with previous month
- Activity trends
- Busiest day (date + user count)
- Peak occupancy

**To enable reports:** Set `DISCORD_REPORTS_CHANNEL_ID` in `.env`

**To disable reports:** Leave `DISCORD_REPORTS_CHANNEL_ID` empty

**Note:** Reports are only sent if there was activity during the period (no empty reports).

### Custom Schedule

To change report schedules, modify [internal/service/scheduler.go](../internal/service/scheduler.go):

```go
// Weekly: Every Monday at 9 AM
s.cron.AddFunc("0 9 * * 1", s.WeeklyReportJob)

// Monthly: 1st of month at 9 AM
s.cron.AddFunc("0 9 1 * *", s.MonthlyReportJob)

// Examples:
// Every Friday at 5 PM: "0 17 * * 5"
// Every day at noon: "0 12 * * *"
// 15th of month at 2 PM: "0 14 15 * *"
```

## Troubleshooting

### Bot Appears Offline

**Check:**

- Token is correct in `.env`
- Bot token hasn't been regenerated
- Application is running: `systemctl status office-tracker`

**Fix:**

```bash
# Verify token
grep DISCORD_TOKEN /opt/office-tracker/.env

# Restart service
sudo systemctl restart office-tracker

# Check logs
sudo journalctl -u office-tracker -n 50
```

### Commands Don't Appear

**Check:**

- Bot has `applications.commands` scope
- Guild ID is correct
- Bot is in the server

**Fix:**

```bash
# Restart application to re-register commands
sudo systemctl restart office-tracker

# Wait 5 minutes, then try:
# Type / in Discord and look for bot commands
```

### "Unknown Interaction" Error

**Cause:** Bot received command but couldn't respond in time (3 seconds)

**Fix:**

- Check application logs for errors
- Verify database isn't locked
- Ensure bot has channel permissions

### Dashboard Not Updating

**Check:**

- Channel name matches exactly (case-sensitive)
- Bot has permissions in channel
- Channel exists in correct guild

**Verify:**

```bash
# Check logs for dashboard updates
sudo journalctl -u office-tracker | grep -i dashboard
```

### Missing Permissions Error

**Fix:**

1. Remove bot from server
2. Generate new invite link with correct permissions
3. Re-invite bot
4. Restart application

### Reports Not Sending

**Check:**

- `DISCORD_REPORTS_CHANNEL_ID` is set
- Channel ID is correct (not channel name)
- Bot has permissions in reports channel
- It's Monday morning

**Test manually:**

```bash
# Check if reports are scheduled
sudo journalctl -u office-tracker | grep -i "weekly report"
```

## Advanced Configuration

### Multiple Servers

To use bot in multiple servers:

1. Enable **Public Bot** in developer portal
2. Set `DISCORD_COMMUNITY_GUILD_ID` to additional server
3. Commands register in both servers
4. Dashboard only works in one server (first match)

### Custom Bot Avatar

1. Go to Discord Developer Portal
2. Open your application
3. Go to **Bot** tab
4. Click bot avatar to upload image
5. Recommended: 512x512 PNG

### Rich Presence

Update bot status in `internal/transport/discord/bot.go`:

```go
bot.UpdateStatusComplex(discordgo.UpdateStatusData{
    Status: "online",
    Activities: []*discordgo.Activity{{
        Name: "Office Attendance",
        Type: discordgo.ActivityTypeWatching,
    }},
})
```

### Custom Embed Colors

Modify embed colors in `internal/transport/discord/dashboards.go`:

```go
embed := &discordgo.MessageEmbed{
    Color: 0x00ff00, // Green (hex color code)
    Title: "Office Tracker",
}
```

## Security Best Practices

- [ ] Keep bot token secret
- [ ] Use `.env` file (never commit to git)
- [ ] Disable **Public Bot** if not needed
- [ ] Only enable required intents
- [ ] Restrict sensitive commands to specific roles
- [ ] Regularly rotate bot token
- [ ] Monitor bot activity logs
- [ ] Set up alerts for unauthorized access

## Rate Limits

Discord imposes rate limits:

- **Commands**: 5 per 5 seconds per user
- **Messages**: 5 per 5 seconds per channel
- **Embeds**: 10 per message

The application handles these automatically with backoff.

## Slash Command Reference

### /checkin

**Description:** Check in a member to the office

**Options:**

- `member` (required) - User to check in

**Permissions:** Officers only (configure in server settings)

**Example:**

```text
/checkin @JohnDoe
```

### /checkout

**Description:** Check out a member from the office

**Options:** `member` (required)

**Permissions:** Everyone

**Example:**

```text
/checkout @JohnDoe
```

### /mystats

**Description:** View your personal attendance statistics (hours, visits, busiest day, avg duration, first/last visit, leaderboard rank)

**Options:**

- `range` (optional)
- `from` / `to` (optional, for custom range)
- `include_auto_checkout` (optional) — include non-RFID checkouts (auto/Discord/API). Default is RFID-only checkout sessions.

**Permissions:** Everyone

**Example:**

```text
/mystats
```

### /stats

**Description:** View the server-wide attendance leaderboard, including busiest day and peak occupancy

**Options:**

- `range` (optional)
- `from` / `to` (optional, for custom range)
- `top` (optional)
- `rank_by` (optional)
- `include_auto_checkout` (optional) — include non-RFID checkouts (auto/Discord/API). Default is RFID-only checkout sessions.

**Permissions:** Everyone

**Example:**

```text
/stats
```

### /help

**Description:** Display help information and available commands

**Options:** None

**Permissions:** Everyone

**Example:**

```text
/help
```

## Useful Resources

- Discord Developer Portal: <https://discord.com/developers/applications>
- Discord.js Guide: <https://discordjs.guide/>
- Discord Bot Permissions Calculator: <https://discordapi.com/permissions.html>
- DiscordGo Documentation: <https://pkg.go.dev/github.com/bwmarrin/discordgo>

## Migration Notes

If migrating from webhook-based system:

1. Keep old webhook running during transition
2. Test bot commands thoroughly
3. Update user documentation
4. Migrate historical data if needed
5. Disable webhooks after successful cutover
