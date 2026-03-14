# Database Maintenance Guide

This guide covers SQLite maintenance tasks for Office Tracker.

Navigation:

- [README.md](README.md) — documentation index
- [../README.md](../README.md) — project overview and quick start

## Overview

The application uses SQLite with WAL (Write-Ahead Logging) mode enabled.
Regular maintenance keeps the database compact and responsive.

## Database location

Default database path: `office.db` (configured via `DB_PATH`).

Related files:

- `office.db` - Main database file
- `office.db-wal` - Write-Ahead Log file
- `office.db-shm` - Shared memory file used by WAL

## Routine maintenance tasks

### 1. WAL checkpoint (merge WAL into main database)

WAL accumulates writes and should be merged periodically.

When to run:

- After high-volume operations (bulk imports or deletions)
- Before backups
- When `office.db-wal` becomes large (for example, over 10 MB)

Command:

```bash
sqlite3 office.db "PRAGMA wal_checkpoint(TRUNCATE);"
```

What it does:

- Moves WAL data into the main database
- Truncates the WAL file to zero bytes
- Usually safe while the app is running (brief lock window)

Checkpoint modes:

- `PASSIVE` - Least intrusive, does not wait for readers
- `FULL` - Waits for readers and checkpoints everything possible
- `RESTART` - Like `FULL`, then resets WAL for reuse
- `TRUNCATE` - Like `RESTART`, then truncates WAL to zero bytes

### 2. VACUUM (reclaim deleted space)

SQLite does not automatically shrink file size after deletes.

When to run:

- After deleting many rows
- When file size is much larger than active data
- Periodically (monthly or quarterly)

Command:

```bash
sqlite3 office.db "VACUUM;"
```

What it does:

- Rebuilds the database file
- Reclaims unused space
- Defragments pages

Important:

- Requires free disk space roughly equal to database size
- Prefer running during maintenance windows
- Can take minutes on larger databases

Safer online alternative:

```bash
sqlite3 office.db "PRAGMA auto_vacuum = INCREMENTAL;"
sqlite3 office.db "PRAGMA incremental_vacuum;"
```

### 3. ANALYZE (update query planner statistics)

`ANALYZE` helps SQLite choose better query plans.

When to run:

- After large imports or deletions
- When query performance degrades
- Periodically (monthly)

Command:

```bash
sqlite3 office.db "ANALYZE;"
```

### 4. Integrity checks

Use integrity checks to detect corruption.

When to run:

- After crashes or abrupt shutdowns
- Before major upgrades
- During scheduled health checks

Quick check:

```bash
sqlite3 office.db "PRAGMA quick_check;"
```

Full check:

```bash
sqlite3 office.db "PRAGMA integrity_check;"
```

## Backup procedures

### Method 1: Online backup (recommended)

```bash
sqlite3 office.db "PRAGMA wal_checkpoint(TRUNCATE);"
sqlite3 office.db ".backup office-backup-$(date +%Y%m%d).db"
```

### Method 2: File copy (app stopped)

```bash
cp office.db office-backup-$(date +%Y%m%d).db
cp office.db-wal office-backup-$(date +%Y%m%d).db-wal 2>/dev/null || true
cp office.db-shm office-backup-$(date +%Y%m%d).db-shm 2>/dev/null || true
```

### Method 3: Automated daily backup (cron)

```bash
0 3 * * * /usr/local/bin/office-db-backup.sh
```

## Database statistics

### View database size

```bash
ls -lh office.db
du -h office.db*
```

### View table row counts

```bash
sqlite3 office.db << EOF
SELECT name, COUNT(*) AS row_count
FROM sqlite_master
WHERE type='table' AND name IN ('users', 'sessions')
GROUP BY name;
EOF
```

### View WAL settings

```bash
sqlite3 office.db "PRAGMA wal_autocheckpoint;"
sqlite3 office.db "PRAGMA journal_mode;"
sqlite3 office.db "PRAGMA journal_size_limit;"
```

## Performance tuning

Current startup settings:

```sql
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;
PRAGMA foreign_keys = ON;
```

Optional read-heavy tuning:

```bash
sqlite3 office.db "PRAGMA cache_size = -64000;"
sqlite3 office.db "PRAGMA temp_store = MEMORY;"
```

Optional write-heavy tuning:

```bash
sqlite3 office.db "PRAGMA synchronous = NORMAL;"
sqlite3 office.db "PRAGMA wal_autocheckpoint = 1000;"
```

## Monitoring queries

### Active sessions count

```bash
sqlite3 office.db "SELECT COUNT(*) FROM sessions WHERE check_out IS NULL;"
```

### Summary statistics

```bash
sqlite3 office.db << EOF
SELECT
  (SELECT COUNT(*) FROM users) AS total_users,
  (SELECT COUNT(*) FROM sessions) AS total_sessions,
  (SELECT COUNT(*) FROM sessions WHERE check_out IS NULL) AS active_sessions;
EOF
```

### Recent activity (last 24 hours)

```bash
sqlite3 office.db << EOF
SELECT u.name, s.check_in, s.check_out
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.check_in >= datetime('now', '-1 day')
ORDER BY s.check_in DESC
LIMIT 20;
EOF
```

## Recommended schedule

### Daily

- Backup database
- Run WAL checkpoint

### Weekly

- Review database and WAL file sizes
- Check error logs

### Monthly

- Run `ANALYZE`
- Review row counts and slow queries

### Quarterly

- Run `PRAGMA integrity_check`
- Run `VACUUM` if many deletions occurred
- Purge old backups

### Before major updates

- Take a full backup
- Run integrity check
- Test restore process

## Troubleshooting

### "database is locked" errors

Common causes:

- Long-running queries
- Too many concurrent writers
- Stuck transactions

Helpful commands:

```bash
sqlite3 office.db "PRAGMA wal_checkpoint;"
sqlite3 office.db "PRAGMA busy_timeout = 10000;"
```

### WAL file keeps growing

```bash
ls -lh office.db-wal
sqlite3 office.db "PRAGMA wal_checkpoint(RESTART);"
```

### Slow queries

```bash
sqlite3 office.db "SELECT * FROM sqlite_stat1 LIMIT 1;"
sqlite3 office.db ".indexes sessions"
```

If `sqlite_stat1` is empty, run `ANALYZE`.

### Corruption recovery

1. Stop the application.
2. Copy database files for safety.
3. Attempt dump-and-restore.
4. Validate recovered file with `PRAGMA integrity_check`.

Example:

```bash
cp office.db office.db.corrupt
cp office.db-wal office.db-wal.corrupt
sqlite3 office.db.corrupt ".dump" | sqlite3 office.db.recovered
sqlite3 office.db.recovered "PRAGMA integrity_check;"
```

## Command reference

```bash
sqlite3 office.db "PRAGMA wal_checkpoint(TRUNCATE);"
sqlite3 office.db "VACUUM;"
sqlite3 office.db "ANALYZE;"
sqlite3 office.db "PRAGMA integrity_check;"
sqlite3 office.db ".backup backup.db"
sqlite3 office.db ".schema"
sqlite3 office.db ".dump" > backup.sql
sqlite3 new.db < backup.sql
```

## Additional resources

- SQLite PRAGMA statements: <https://www.sqlite.org/pragma.html>
- SQLite WAL mode: <https://www.sqlite.org/wal.html>
- SQLite VACUUM: <https://www.sqlite.org/lang_vacuum.html>
- SQLite optimization overview: <https://www.sqlite.org/optoverview.html>
