# Internal Architecture

This folder contains the core application logic, organized by responsibility.

## Package Overview

- `app/` — Dependency wiring and app initialization.
- `config/` — Environment configuration loader and defaults.
- `database/` — SQLite open + migrations.
- `domain/` — Core business models (User, Session, Stats).
- `logging/` — Structured logging helpers.
- `query/` — Query filter objects.
- `repository/` — Data access layer (SQL/SQLite).
- `service/` — Business logic and orchestration.
- `transport/` — External interfaces (HTTP and Discord).
- `api/` — DTOs for API requests/responses.

## Data Flow (High Level)

```
HTTP/Discord -> transport -> service -> repository -> database
```

## See Also

- HTTP + Discord transport: [internal/transport/README.md](transport/README.md)
- Services: [internal/service/README.md](service/README.md)
- Repositories: [internal/repository/README.md](repository/README.md)
