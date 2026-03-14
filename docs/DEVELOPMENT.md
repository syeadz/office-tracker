# Development Guide

This guide covers local development for Office Tracker.

Navigation:

- [README.md](README.md) — documentation index
- [../README.md](../README.md) — project overview and quick start

## Prerequisites

- Go 1.21+
- Git
- SQLite CLI (`sqlite3`)

Optional:

- Docker
- VS Code with Go extension

## Local setup

From the project root:

```bash
go mod download
cp .env.example .env
```

Minimum `.env` values:

```bash
HTTP_PORT=8080
DB_PATH=office.db
```

Run the app:

```bash
go run cmd/office/main.go
```

Health check:

```bash
curl http://localhost:8080/health
```

## Project layout

- `cmd/office/main.go`: application entrypoint
- `internal/repository`: data access
- `internal/service`: business logic
- `internal/transport/http`: HTTP handlers and routing
- `internal/transport/discord`: Discord bot
- `api/openapi.yaml`: API specification
- `api/requests.http`: API request examples

## Daily workflow

```bash
# format and static checks
go fmt ./...
go vet ./...

# unit tests
go test ./...
go test -race ./...

# integration tests (tagged)
go test -tags=integration ./test/integration/...

# all tests including integration
go test -tags=integration ./...

# optional coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### VS Code test runner and build tags

If you run tests from the VS Code Testing panel, make sure integration tests include the build tag:

- Workspace setting: `"go.testTags": "integration"`

Without that setting, tagged files in `test/integration` are excluded by Go build constraints.

## API development notes

When adding or changing endpoints:

1. Update route wiring in `internal/transport/http/server.go`.
2. Add/update handler tests under `internal/transport/http`.
3. Update `api/openapi.yaml`.
4. Update `api/requests.http` and `api/README.md`.

Use `api/requests.http` to test requests in VS Code REST Client.

## Environmental data and metrics

### Environmental data model

- Source: `POST /api/environment` (typically ESP32 every ~1 minute)
- Storage: in-memory latest reading only
- Freshness: 5 minutes
- No database persistence

If `timestamp` is omitted in `POST /api/environment`, server time is used.

### Prometheus metrics

Scrape endpoint:

```text
GET /metrics
```

Metrics currently exported:

- `office_presence_active_users`
- `office_environment_temperature_celsius` (fresh only)
- `office_esp_health_up{device_id=...}` (`1` fresh, `0` stale)
- `office_esp_health_uptime_seconds{device_id=...}` (fresh only)
- `office_esp_health_free_heap_bytes{device_id=...}` (fresh only)
- `office_esp_health_rssi_dbm{device_id=...}` (fresh only)

Important behavior:

- Stale environment data is excluded from metrics output.
- Stale ESP resource metrics are excluded, while `office_esp_health_up` becomes `0`.
- Presence metric remains available even when environment metrics are absent.

Quick local check:

```bash
# 1) publish an environment reading
curl -X POST http://localhost:8080/api/environment \
  -H 'Content-Type: application/json' \
  -d '{"temperature_c":24.8}'

# 2) verify metrics include environment gauges
curl http://localhost:8080/metrics | grep office_environment
```

### ESP32 health heartbeat endpoint

For device diagnostics and fleet visibility, use:

```text
POST /api/devices/health
GET  /api/devices/health
```

Recommended cadence is every 5-15 minutes with small fields only:

- `uptime_seconds`
- `free_heap_bytes`
- `wifi_connected`
- `rssi`
- `ip` (optional)
- `firmware_version` (optional)
- `reset_reason` (optional)

Data is kept in memory only (latest entry per `device_id`).

## Database development notes

The app auto-initializes SQLite schema on startup.

Reset local DB:

```bash
rm -f office.db office.db-wal office.db-shm
```

Inspect database:

```bash
sqlite3 office.db ".schema"
sqlite3 office.db "SELECT COUNT(*) FROM users;"
sqlite3 office.db "SELECT COUNT(*) FROM sessions;"
```

For maintenance operations, see [DATABASE_MAINTENANCE.md](DATABASE_MAINTENANCE.md).

## Discord development notes

- Keep Discord integration optional in local dev.
- Use a separate test bot token and test guild.
- Do not use production bot credentials locally.

Setup details are in [DISCORD_SETUP.md](DISCORD_SETUP.md).

## Debugging tips

- Read app logs in the terminal where `go run` is running.
- Validate endpoint behavior with `curl` or `api/requests.http`.
- Use focused test runs while iterating:

```bash
go test ./internal/transport/http -run TestUser
go test ./internal/service -run TestSession
```

## Contribution checklist

Before opening a PR:

- [ ] Code is formatted (`go fmt ./...`)
- [ ] Vet passes (`go vet ./...`)
- [ ] Tests pass (`go test ./...`)
- [ ] API docs updated (`api/openapi.yaml` and `api/README.md`)
- [ ] Relevant docs in `docs/` updated
