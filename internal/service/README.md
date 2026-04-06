# Service Layer

The service layer coordinates business logic and domain workflows.

## Services

- `AttendanceService` — RFID scan processing, check-in/out, scan history.
- `SessionService` — Session management, admin check-in/out helpers.
- `UserService` — CRUD operations for users.
- `OfficeStatsService` — Aggregates stats for leaderboards and period statistics (RFID-only checkouts by default).
- `ReportsService` — Generates weekly/monthly activity reports with trend analysis and runtime toggles.
- `SchedulerService` — Cron jobs (auto-checkout, weekly reports, monthly reports).

## Reports

### Scheduled Reports

Automated reports:

- Weekly report every **Monday at 9:00 AM**
- Monthly report every **1st day of month at 9:00 AM**

**Features:**

- Total hours, visits, unique users, active days
- Week-over-week percentage comparisons
- Top 10 contributor leaderboard
- Color-coded Discord embeds

**Configuration:**

```bash
# .env
DISCORD_REPORTS_CHANNEL_ID=<channel-id>
```

**HTTP Endpoints:**

```bash
GET /api/reports/weekly
GET /api/reports/status
POST /api/reports/toggle?enabled=true|false&period=all|weekly|monthly
```

**Design Principles:**

- Only sends reports if there was at least one visit (no spam for empty weeks)
- Interface-based delivery allows multiple channels (Discord, email, etc.)
- Graceful degradation if channel not configured
- Comprehensive logging for debugging

## Scheduled Jobs

- **Auto-checkout** — Daily at 04:00 AM (clears stuck sessions)
- **Weekly Report** — Monday at 09:00 AM (if weekly schedule enabled)
- **Monthly Report** — 1st day of month at 09:00 AM (if monthly schedule enabled)

## Implementation Details

- Uses `robfig/cron/v3` for reliable scheduling
- Services use structured logging via internal/logging
- All public methods have error handling
- Dependency injection via constructors
- Interface-based design for testability

## Testing

Run service tests:

```bash
go test ./internal/service -v
```
