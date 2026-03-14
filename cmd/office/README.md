# cmd/office

This is the application entry point.

## Responsibilities

- Load configuration
- Initialize database and migrations
- Wire dependencies (`app.New`)
- Start scheduler, HTTP server, and Discord bot
- Setup Discord commands
- Handle graceful shutdown

## Files

- `main.go` — main process bootstrap
