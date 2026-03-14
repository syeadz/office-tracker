# Office Tracker

A Go service for tracking office attendance with RFID scans,
plus Discord automation and a lightweight admin UI.

## Quick Start

```bash
# From project root
cp .env.example .env
# edit .env with your settings

go run cmd/office/main.go
```

The server will start on `HTTP_PORT` (default: 8080).
The web management UI is available at <http://localhost:8080/>
or <http://localhost:8080/ui>

## What This App Does

- **Attendance Tracking** — Check-in/check-out via RFID scans or Discord commands
- **REST API** — HTTP endpoints for attendance, analytics, and management
- **Web Dashboard** — Management UI with statistics, leaderboards, and analytics
- **Discord Integration** — Slash commands, presence queries, and automated dashboards
- **Analytics** — Weekly and monthly reports (optional), leaderboards, and historical data
- **Trusted Leaderboards** — Default stats use RFID sign-outs only; optional include flag can include non-RFID checkouts
- **Auto-Checkout** — Daily scheduled cleanup at 04:00 AM

## Features

### Core Features

- ✅ RFID scan integration
- ✅ Discord /checkin, /checkout, /mystats slash commands
- ✅ Weekly and monthly activity reports (Discord embeds)
- ✅ HTTP API for automation
- ✅ Web UI for management
- ✅ Leaderboards and user statistics
- ✅ Session history and filtering

### Channels

- **RFID** — Physical badge readers
- **Discord** — Bot commands and dashboards
- **HTTP** — REST API and web UI

## Configuration

Set these in `.env` (see [.env.example](.env.example) for full reference):

```bash
# Required
HTTP_PORT=8080
DB_PATH=office.db

# Optional: Discord integration
DISCORD_TOKEN=<your-token>
DISCORD_EXEC_GUILD_ID=<guild-id>
DISCORD_COMMUNITY_GUILD_ID=<guild-id>
DISCORD_DASHBOARD_CHANNEL_NAME=office-tracker

# Optional: Weekly and monthly reports
DISCORD_REPORTS_CHANNEL_ID=<channel-id>

# Optional: Security
API_KEY=<your-api-key>
CORS_ORIGINS=http://localhost:3000
```

## Usage

### Discord Commands

```text
/checkin @member             # Check in someone
/checkout @member            # Check out someone
/mystats                      # View your weekly stats
/stats                        # View office leaderboard
/help                         # Show available commands
```

### HTTP API

**Get presence:**

```bash
curl http://localhost:8080/api/presence
```

**Check in user:**

```bash
curl -X POST http://localhost:8080/api/sessions/checkin \
  -H "Content-Type: application/json" \
  -d '{"user_id": "123"}'
```

**View weekly report:**

```bash
curl http://localhost:8080/api/reports/weekly
```

**Publish environmental reading (in-memory):**

```bash
curl -X POST http://localhost:8080/api/environment \
    -H "Content-Type: application/json" \
    -d '{"temperature_c":24.8}'
```

**Prometheus metrics:**

```bash
curl http://localhost:8080/metrics
```

**ESP32 heartbeat (health):**

```bash
curl -X POST http://localhost:8080/api/devices/health \
    -H "Content-Type: application/json" \
    -d '{"device_id":"esp-lab-1","uptime_seconds":3670,"free_heap_bytes":153248,"wifi_connected":true,"rssi":-61}'
```

See [api/README.md](api/README.md) for complete API reference.

### Environmental data behavior

- In-memory only (no DB persistence)
- Stores latest reading only
- Freshness window: 5 minutes
- Stale readings are excluded from dashboard usage and environment metrics

### Web UI

Visit <http://localhost:8080/ui> for:

- Real-time dashboards
- User and session management
- Statistics and analytics
- CSV import/export

## Architecture

This project follows **layered clean architecture**:

```text
Transport Layer (HTTP, Discord)
    ↓
Service Layer (Business Logic)
    ↓
Repository Layer (Data Access)
    ↓
Database (SQLite)
```

Key patterns:

- Dependency injection via constructors
- Interface-based services for testability
- Structured logging throughout
- Graceful error handling

### Folder Structure

```text
internal/
├── api/        # DTOs and request/response models
├── app/        # Dependency injection container
├── config/     # Configuration from environment
├── database/   # Schema and migrations
├── domain/     # Domain models (User, Session, Report)
├── logging/    # Structured logging setup
├── query/      # Query filters
├── repository/ # Data access layer
├── service/    # Business logic
└── transport/  # HTTP and Discord handlers
```

See folder-specific READMEs for detailed documentation.

## Documentation

- **[docs/README.md](docs/README.md)** — central documentation index
- **[api/README.md](api/README.md)** — API reference (endpoints, auth, examples)

Use the docs index for the full map (development, deployment, Discord setup, DB maintenance, and package-level implementation docs).

## Database

SQLite database with automatic migrations on startup. Configure path via `DB_PATH`.

```bash
# For development
DB_PATH=office.db

# For production
DB_PATH=/var/lib/office-tracker/office.db
```

## Running Tests

```bash
# Default test suite
go test ./...

# Include integration-tagged tests
go test -tags=integration ./...
```

For full test workflow (race, focused runs, coverage, VS Code tag config), see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

## Docker

```bash
docker run --rm -p 8080:8080 --env-file .env office-tracker:local
```

## Notes

- Discord commands (except `setup`) are restricted to the exec guild.

## Security

- If `API_KEY` is set, all HTTP endpoints (except `/health`) require auth.
- Avoid `CORS_ORIGINS=*` in production.
- Discord admin-only commands are enforced via Discord permissions.
