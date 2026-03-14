# Transport Layer

This folder exposes the system to the outside world. It currently supports HTTP and Discord.

## HTTP (`internal/transport/http`)

- Serves REST API endpoints and management UI.
- Endpoints are registered in `server.go`.
- Management UI HTML/JS is in `static.go`.

Key endpoints:
- `/health`
- `/api/users`, `/api/sessions`, `/api/presence`
- `/api/statistics/*`

See: [internal/transport/http/README.md](http/README.md)

## Discord (`internal/transport/discord`)

- Slash commands for user/session/admin operations.
- Dashboard message rendering + button interactions.

Key files:
- `bot.go` — lifecycle + command registration
- `handlers.go` — routing commands/buttons
- `commands/` — command handlers

See: [internal/transport/discord/README.md](discord/README.md)

## Notes

- All commands except `setup` are restricted to the exec guild.
- Admin permissions are enforced for sensitive commands where configured.
