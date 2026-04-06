# Discord Bot

This package provides Discord integration: dashboards and commands.

For full Discord app/bot provisioning and operational troubleshooting, see:

- [../../../docs/DISCORD_SETUP.md](../../../docs/DISCORD_SETUP.md)

## Setup summary

Required config (see [.env.example](../../.env.example)):

- `DISCORD_TOKEN`
- `DISCORD_EXEC_GUILD_ID`
- `DISCORD_COMMUNITY_GUILD_ID`
- `DISCORD_DASHBOARD_CHANNEL_NAME`

Notes:

- All commands except `setup` are restricted to the exec guild.

## Security

- Discord permission checks are applied via command registration for admin-only commands.
- Keep the bot token secret and rotate if exposed.

## Slash Commands

### General Commands

- `setup` — create a dashboard in the current channel (available in exec + community guilds)
- `ping` — bot health check
- `help` — show help information about available commands

### User Management Commands

- `user-create` — create a new user (admin only; requires member + RFID UID, optional name override)
- `user-get` — fetch a user by ID
- `user-update` — update user details (admin only; supports Discord relink and RFID tag updates)
- `user-delete` — delete a user (admin only)
- `user-list` — list users with page-based pagination (`limit`, `page`)

### Session Commands

- `session-list` — list sessions with optional filters (date range, name, etc.)
- `session-get` — get details of a specific session by ID
- `session-active` — list all currently active sessions

### Statistics Commands

- `stats` — office statistics (weekly, monthly, or custom date range). Defaults to RFID-only checkout sessions; use `include_auto_checkout` to include non-RFID checkouts (auto/Discord/API).
- `mystats` — your personal statistics (weekly, monthly, or custom date range). Uses the same `include_auto_checkout` behavior.
- `reports` — check or update scheduled reports runtime settings. Use optional booleans `weekly` and `monthly`.

### Attendance Commands (Admin)

- `checkin` — **check in a member by @ mentioning them** (admin only). Requires the member to be registered in the system.
- `checkout` — check out a member by @ mentioning them
- `checkout-all` — bulk checkout all currently checked-in members

## Dashboard Behavior

- The dashboard auto-updates on RFID scans.
- The dashboard supports manual refresh via button.
- Debounced render prevents rapid UI updates.

## Weekly Reports

Automated weekly reports sent to a configured Discord channel (if enabled):

**Configuration (in `.env`):**

```bash
DISCORD_REPORTS_CHANNEL_ID=<your-channel-id>
```

**Features:**

- Sent every Monday at 9:00 AM (local time)
- Includes: total hours, visits, unique users, active days
- Week-over-week trend analysis with percentage changes
- Top 10 contributor leaderboard with medals
- Only sends if there was at least one visit that week

**Runtime control:**

- `/reports` shows availability plus weekly/monthly enabled flags.
- `/reports weekly:true` enables weekly schedule only.
- `/reports monthly:false` disables monthly schedule only.
- `/reports weekly:true monthly:false` updates both schedules in one command.
- If reports are fully disabled at startup (missing guild/channel configuration), the command reports unavailable and does not allow toggling.

## Natural Language Presence Queries

The bot listens for messages like:

- "who's in the office"
- "anyone in the office"

When detected, it posts a presence embed in the channel and deletes it after ~5 minutes.
The embed includes a timestamp for when it was generated.
