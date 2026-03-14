// migrate is a one-time migration script that reads a database using the old schema
// (members + visits tables) and writes a new database using the new schema
// (users + sessions tables). The old database is opened read-only and is never modified.
//
// Usage:
//
//	go run scripts/migrate/main.go <old-db-path> <new-db-path>
//
// Schema mapping:
//
//	members.id          → users.id
//	members.name        → users.name
//	members.uid         → users.rfid_uid
//	members.discord_id  → users.discord_id
//	visits.id           → sessions.id
//	visits.member_id    → sessions.user_id
//	visits.signin_time  → sessions.check_in
//	visits.signout_time → sessions.check_out  (empty string becomes NULL)
//	(none)              → sessions.check_out_method ('rfid' by default, 'auto' for 04:00 checkouts, NULL if active)
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <old-db-path> <new-db-path>\n", os.Args[0])
		os.Exit(1)
	}

	oldPath := os.Args[1]
	newPath := os.Args[2]

	if _, err := os.Stat(oldPath); err != nil {
		log.Fatalf("old database not found: %v", err)
	}
	if _, err := os.Stat(newPath); err == nil {
		log.Fatalf("output file already exists: %s — remove it first", newPath)
	}

	newDB, err := sql.Open("sqlite", newPath)
	if err != nil {
		log.Fatalf("failed to create new database: %v", err)
	}
	defer newDB.Close()

	if err := migrate(newDB, oldPath); err != nil {
		newDB.Close()
		os.Remove(newPath)
		log.Fatalf("migration failed: %v\nPartial output file removed.", err)
	}

	fmt.Println("\nMigration completed successfully.")
}

// migrate creates the new schema in newDB, attaches the old database as a read-only
// source, transfers all rows, then detaches it.
func migrate(newDB *sql.DB, oldPath string) error {
	// Attach the old database under the alias "old" (read-only).
	if _, err := newDB.Exec(`ATTACH DATABASE 'file:` + oldPath + `?mode=ro' AS old;`); err != nil {
		return fmt.Errorf("attaching old database: %w", err)
	}
	defer newDB.Exec(`DETACH DATABASE old;`) //nolint:errcheck

	if err := validateOldSchema(newDB); err != nil {
		return err
	}

	// Pre-flight: detect users with multiple active sessions (empty signout_time).
	// The new schema enforces at most one active session per user via a unique index.
	var duplicateActive int
	if err := newDB.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT member_id FROM old.visits WHERE TRIM(signout_time) = '' GROUP BY member_id HAVING COUNT(*) > 1
		)
	`).Scan(&duplicateActive); err != nil {
		return fmt.Errorf("checking active sessions: %w", err)
	}
	if duplicateActive > 0 {
		return fmt.Errorf(
			"%d user(s) have multiple active sessions (empty signout_time). "+
				"Resolve these in the old database before migrating",
			duplicateActive,
		)
	}

	steps := []struct {
		name string
		sql  string
	}{
		{"pragma WAL", `PRAGMA journal_mode = WAL;`},
		{"pragma busy_timeout", `PRAGMA busy_timeout = 5000;`},
		{"pragma foreign_keys off", `PRAGMA foreign_keys = OFF;`},

		{"create users", `
			CREATE TABLE users (
				id           INTEGER PRIMARY KEY AUTOINCREMENT,
				name         TEXT    NOT NULL,
				rfid_uid     TEXT    UNIQUE NOT NULL,
				discord_id   TEXT,
				created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
			);`},

		{"create sessions", `
			CREATE TABLE sessions (
				id                INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id           INTEGER NOT NULL,
				check_in          DATETIME NOT NULL,
				check_out         DATETIME,
				check_out_method  TEXT,
				CHECK (check_out_method IS NULL OR check_out_method IN ('rfid', 'discord', 'api', 'auto')),
				CHECK (
					(check_out IS NULL AND check_out_method IS NULL) OR
					(check_out IS NOT NULL AND check_out_method IS NOT NULL)
				),
				FOREIGN KEY(user_id) REFERENCES users(id)
			);`},

		// members → users
		{"migrate members → users", `
			INSERT INTO users (id, name, rfid_uid, discord_id)
			SELECT id, name, uid, discord_id
			FROM old.members;`},

		// visits → sessions
		// Empty signout_time → NULL (active session).
		// Completed sessions default to check_out_method = 'rfid'.
		// Legacy 04:00 checkouts are treated as scheduler auto-checkouts.
		{"migrate visits → sessions", `
			INSERT INTO sessions (id, user_id, check_in, check_out, check_out_method)
			SELECT
				id,
				member_id,
				signin_time,
				CASE WHEN TRIM(signout_time) = '' THEN NULL ELSE signout_time END,
				CASE
					WHEN TRIM(signout_time) = '' THEN NULL
					WHEN strftime('%H:%M', substr(signout_time, 1, 19)) = '04:00' THEN 'auto'
					ELSE 'rfid'
				END
			FROM old.visits;`},

		{"index: one active session", `
			CREATE UNIQUE INDEX idx_one_active_session ON sessions(user_id) WHERE check_out IS NULL;`},
		{"index: checkout method", `
			CREATE INDEX idx_sessions_checkout_method ON sessions(check_out_method);`},

		{"pragma foreign_keys on", `PRAGMA foreign_keys = ON;`},
	}

	for _, step := range steps {
		fmt.Printf("  %-45s ... ", step.name)
		if _, err := newDB.Exec(step.sql); err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("step %q: %w", step.name, err)
		}
		fmt.Println("ok")
	}

	var userCount, sessionCount int
	if err := newDB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&userCount); err != nil {
		return err
	}
	if err := newDB.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&sessionCount); err != nil {
		return err
	}
	fmt.Printf("\nRows migrated: %d users, %d sessions\n", userCount, sessionCount)

	return nil
}

// validateOldSchema checks that the expected old-schema tables exist in the attached
// "old" database and that no new-schema tables have already been created in newDB.
func validateOldSchema(newDB *sql.DB) error {
	var name string

	for _, table := range []string{"members", "visits"} {
		err := newDB.QueryRow(
			`SELECT name FROM old.sqlite_master WHERE type='table' AND name = ?`, table,
		).Scan(&name)
		if err == sql.ErrNoRows {
			return fmt.Errorf("expected table %q not found in old database — is this the right file?", table)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
