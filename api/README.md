# Office Tracker API Documentation

This folder contains API specs and request examples for the HTTP server.

Navigation:

- [../docs/README.md](../docs/README.md) — documentation index
- [../README.md](../README.md) — project overview and quick start

## Contents

- [openapi.yaml](openapi.yaml): OpenAPI 3.0 specification
- [requests.http](requests.http): request examples for REST Client

## Runtime configuration

Common environment variables:

- `HTTP_PORT` (default: `8080`)
- `DB_PATH` (default: `office.db`)
- `API_KEY` (optional, enables API key auth)
- `CORS_ORIGINS` (optional, enables CORS)

Example:

```bash
export HTTP_PORT=8080
export DB_PATH=office.db
export API_KEY=your-secret-api-key
export CORS_ORIGINS=http://localhost:3000,https://example.com
```

## Authentication

If `API_KEY` is set, all endpoints except `GET /health`, `GET /`, and `GET /ui` require auth.

Supported headers:

- `X-API-Key: your-api-key`
- `Authorization: Bearer your-api-key`

Examples:

```bash
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/users
```

```bash
curl -H "Authorization: Bearer your-api-key" \
  http://localhost:8080/api/users
```

## Quick start

Run locally:

```bash
go run cmd/office/main.go
```

API base URL: `http://localhost:8080`

Basic checks:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/users
```

Create a user:

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","rfid_uid":"ABC123"}'
```

## Endpoint summary

### Health

- `GET /health`
- `GET /metrics` (Prometheus format)

### Users

- `GET /api/users`
- `POST /api/users`
- `DELETE /api/users` (bulk delete with filters)
- `GET /api/users/{id}`
- `PUT /api/users/{id}`
- `DELETE /api/users/{id}`
- `GET /api/users/export`
- `POST /api/users/import`

### Sessions and presence

- `GET /api/presence`
- `GET /api/sessions`
- `DELETE /api/sessions` (bulk delete with filters)
- `PUT /api/sessions/{id}`
- `DELETE /api/sessions/{id}`
- `GET /api/sessions/export`
- `GET /api/sessions/open`
- `GET /api/sessions/user/{userId}`
- `POST /api/sessions/checkin`
- `POST /api/sessions/checkout`
- `POST /api/sessions/checkout/{userId}`
- `POST /api/sessions/checkout-all`

Session payload note:

- Session objects returned by `GET /api/sessions` and `GET /api/sessions/user/{userId}` include `check_out_method` for completed sessions (`rfid`, `discord`, `api`, `auto`).
- Session list/count/export and bulk delete endpoints support `check_out_method=rfid|discord|api|auto` (or `all`).
- CSV exports from `GET /api/sessions/export` include a `CheckOutMethod` column.

### Statistics

- `GET /api/statistics/leaderboard`
- `GET /api/statistics/weekly`
- `GET /api/statistics/monthly`
- `GET /api/statistics/report`
- `GET /api/statistics/users/{id}`

Default behavior for statistics endpoints is **RFID-only checkouts** for trusted leaderboards.
Set `include_auto_checkout=true` to include non-RFID checkouts (auto/Discord/API/manual).

### Environment

- `GET /api/environment`
- `POST /api/environment`

#### Environmental data behavior

- Environmental readings are stored **in memory only** (no database persistence).
- Only the **latest** reading is kept.
- Freshness window is **5 minutes**.
- If `POST /api/environment` omits `timestamp`, the server uses current time.
- Dashboard and metrics use the reading only when it is fresh.

Recommended ingestion pattern (ESP32):

- Send one reading every ~60 seconds.
- Prefer omitting `timestamp` unless device time is trusted/synchronized.

Example payload:

```json
{
  "temperature_c": 24.8
}
```

### Device health (ESP32 heartbeats)

- `GET /api/devices/health`
- `POST /api/devices/health`

Use this endpoint for low-frequency heartbeat data from one or more ESP32 devices
(recommended every 5-15 minutes).

Recommended small fields:

- `uptime_seconds`
- `free_heap_bytes`
- `wifi_connected`
- `rssi`
- `ip` (optional)
- `firmware_version` (optional)
- `reset_reason` (optional)

Notes:

- Data is stored in memory only (latest record per device ID).
- If `device_id` is omitted, the server stores under `default`.

### Metrics (Prometheus)

- `GET /metrics`

This endpoint exposes Prometheus-format metrics for scraping.

Currently exposed:

- `office_presence_active_users` — active checked-in users.
- `office_environment_temperature_celsius` — latest fresh temperature.
- `office_esp_health_up{device_id=...}` — ESP heartbeat freshness status (`1` fresh, `0` stale).
- `office_esp_health_uptime_seconds{device_id=...}` — ESP uptime (fresh only).
- `office_esp_health_free_heap_bytes{device_id=...}` — ESP free heap (fresh only).
- `office_esp_health_rssi_dbm{device_id=...}` — ESP Wi-Fi RSSI (fresh only).

Stale-data policy:

- If environmental data is stale (>5 minutes), the temperature metric is **not emitted**.
- If ESP heartbeat data is stale (>20 minutes), ESP resource metrics are **not emitted** and `office_esp_health_up` is `0`.
- Presence metric is still emitted.

Auth note:

- If `API_KEY` is enabled, `/metrics` also requires API key auth.

### Reports

- `GET /api/reports/weekly` (available when reports service is enabled)
- `GET /api/reports/status` (check if reports are enabled)
- `POST /api/reports/toggle?enabled=true|false` (enable/disable reports at runtime)

### RFID attendance

- `POST /api/rfid/scan`
- `GET /api/rfid/scans`
- `DELETE /api/rfid/scans`

## Response notes

- Most endpoints return JSON.
- `GET /health` returns plain text: `ok`.
- Most API handler validation and error responses return JSON in the form `{ "error": "..." }`.
- Middleware or non-API responses may still return plain text in some cases.

## Keeping docs in sync

When endpoints change:

1. Update [openapi.yaml](openapi.yaml).
2. Update [requests.http](requests.http).
3. Update this file's endpoint summary.
