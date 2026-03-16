# HTTP Transport

This package serves the REST API and the management UI.

For canonical API endpoint/reference documentation, use:

- [../../../api/README.md](../../../api/README.md)
- [../../../api/openapi.yaml](../../../api/openapi.yaml)

## Management UI

The web UI is served from `static.go` at `/` and `/ui` and provides:

- Dashboard + live presence
- User management (create/delete, CSV import/export)
- Sessions and scans (including sign-out method visibility for completed sessions)
- Office stats
- Leaderboards (weekly/monthly/custom, hours or visits)
- Member stats with rankings, percentages, and recent session history
- Admin check-in/out from the UI

## Scope of this document

This file intentionally focuses on transport-layer behavior and UI notes.
Endpoint-level reference and request examples are maintained in `api/` to avoid drift.

## Response Format

- Successful API responses are JSON (except `/health`, which returns plain text).
- Most handler-level API errors now return JSON in the shape `{ "error": "..." }` for consistency.
- Some middleware-generated errors may still be plain text.

## Security

- If `API_KEY` is set, all HTTP endpoints except `/health`, `/`, and `/ui` require authentication.
- The management UI page is public, but its API calls use the same API key header.
