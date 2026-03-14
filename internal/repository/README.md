# Repository Layer

The repository layer encapsulates database access and SQL queries.

## Responsibilities

- CRUD operations for users and sessions.
- Statistics aggregation for leaderboards.
- Filtering (date range, user, active-only).

## Notes

- Uses SQLite via `database/sql`.
- Stats queries default to RFID-only checkouts for trusted leaderboards.
- `include_auto_checkout=true` includes non-RFID checkout methods (auto/Discord/API).
