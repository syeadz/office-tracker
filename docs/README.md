# Documentation Index

Use this file as the entry point for project documentation.

## Start here

- [../README.md](../README.md) — project overview, quick start, high-level usage
- [DEVELOPMENT.md](DEVELOPMENT.md) — local development workflow and testing
- [DEPLOYMENT.md](DEPLOYMENT.md) — production deployment options and operations

## Feature-specific guides

- [DISCORD_SETUP.md](DISCORD_SETUP.md) — Discord app/bot setup and operational troubleshooting
- [DATABASE_MAINTENANCE.md](DATABASE_MAINTENANCE.md) — SQLite maintenance, backup, and recovery

## API reference

- [../api/README.md](../api/README.md) — API behavior and endpoint summary
- [../api/openapi.yaml](../api/openapi.yaml) — OpenAPI specification (source of truth)
- [../api/requests.http](../api/requests.http) — runnable request examples

## Package-level docs

Keep package-level READMEs focused on implementation notes for contributors.
Avoid duplicating full setup flows or endpoint lists that already live in `docs/` or `api/`.

- [../cmd/office/README.md](../cmd/office/README.md) — entrypoint responsibilities
- [../internal/service/README.md](../internal/service/README.md) — business logic layer notes
- [../internal/repository/README.md](../internal/repository/README.md) — data access layer notes
- [../internal/transport/http/README.md](../internal/transport/http/README.md) — HTTP transport behavior and UI notes
- [../internal/transport/discord/README.md](../internal/transport/discord/README.md) — Discord transport behavior and command surface
