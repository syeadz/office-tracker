// Package repository provides functions for interacting with the database in the Office Tracker application.
package repository

import (
	"database/sql"
	"strings"
	"time"

	"office/internal/domain"
	"office/internal/query"
)

const (
	CheckOutMethodRFID    = "rfid"
	CheckOutMethodDiscord = "discord"
	CheckOutMethodAPI     = "api"
	CheckOutMethodAuto    = "auto"
)

func normalizeCheckOutMethod(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case CheckOutMethodRFID:
		return CheckOutMethodRFID
	case CheckOutMethodDiscord:
		return CheckOutMethodDiscord
	case CheckOutMethodAuto:
		return CheckOutMethodAuto
	default:
		return CheckOutMethodAPI
	}
}

// buildCheckOutFilter returns a SQL WHERE clause fragment used by stats/leaderboard queries.
// When excludeAutoCheckout is true, only trusted RFID sign-outs are included.
// When false, all known sign-out methods are included.
func buildCheckOutFilter(excludeAutoCheckout bool) string {
	if excludeAutoCheckout {
		return "AND s.check_out_method = '" + CheckOutMethodRFID + "'"
	}

	return "AND s.check_out_method IN ('" + CheckOutMethodRFID + "','" + CheckOutMethodDiscord + "','" + CheckOutMethodAPI + "','" + CheckOutMethodAuto + "')"
}

func normalizeRankBy(rankBy string) string {
	if rankBy == "visits" {
		return "visits"
	}

	return "hours"
}

func applySessionPresenceFilter(where *[]string, filter query.SessionFilter) {
	switch filter.Status {
	case "active":
		*where = append(*where, "s.check_out IS NULL")
	case "completed":
		*where = append(*where, "s.check_out IS NOT NULL")
	default:
		if filter.ActiveOnly {
			*where = append(*where, "s.check_out IS NULL")
		}
	}
}

func buildSessionSortClause(filter query.SessionFilter) string {
	orderDir := "DESC"
	if filter.OrderBy == "asc" {
		orderDir = "ASC"
	}

	sortField := "s.check_in"
	switch filter.SortBy {
	case "check_out":
		sortField = "s.check_out"
	case "user_name":
		sortField = "u.name"
	}

	return sortField + " " + orderDir
}

// SessionWithUser represents a session enriched with user information from JOIN
type SessionWithUser struct {
	*domain.Session
	UserName string
}

type SessionRepo struct {
	DB *sql.DB
}

// GetOpenSession retrieves the ID of an open session for the specified user. If no open session is found, it returns an error.
func (r *SessionRepo) GetOpenSession(userID int64) (int64, error) {
	row := r.DB.QueryRow(`
		SELECT id FROM sessions
		WHERE user_id = ? AND check_out IS NULL
		LIMIT 1`, userID)

	var id int64
	err := row.Scan(&id)
	return id, err
}

// FindByID retrieves a session by ID with user information via JOIN.
// Returns SessionWithUser to include the user name in a single query.
func (r *SessionRepo) FindByID(id int64) (*SessionWithUser, error) {
	row := r.DB.QueryRow(`
		SELECT s.id, s.user_id, s.check_in, s.check_out, s.check_out_method, u.name
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = ?
		LIMIT 1`, id)

	var s domain.Session
	var checkOutMethod sql.NullString
	var userName string
	err := row.Scan(&s.ID, &s.UserID, &s.CheckIn, &s.CheckOut, &checkOutMethod, &userName)
	if err != nil {
		return nil, err
	}
	if checkOutMethod.Valid {
		s.CheckOutMethod = checkOutMethod.String
	}

	return &SessionWithUser{
		Session:  &s,
		UserName: userName,
	}, nil
}

// CheckIn creates a new session for the specified user with the current time as the check-in time.
// Returns an error if the operation fails.
func (r *SessionRepo) CheckIn(userID int64) error {
	_, err := r.DB.Exec(`
		INSERT INTO sessions(user_id, check_in)
		VALUES (?, ?)`, userID, time.Now())
	return err
}

// CheckOut updates the specified session with the current time as the check-out time.
// Returns an error if the operation fails.
func (r *SessionRepo) CheckOut(sessionID int64) error {
	return r.CheckOutWithMethod(sessionID, CheckOutMethodRFID)
}

// CheckOutWithMethod updates the specified session with the current time and checkout method.
// Returns an error if the operation fails.
func (r *SessionRepo) CheckOutWithMethod(sessionID int64, method string) error {
	method = normalizeCheckOutMethod(method)

	_, err := r.DB.Exec(`
		UPDATE sessions SET check_out = ?, check_out_method = ?
		WHERE id = ?`, time.Now(), method, sessionID)
	return err
}

// List retrieves sessions matching the filter, joining users for name and discord_id.
// Returns SessionWithUser to avoid N+1 queries.
func (r *SessionRepo) List(filter query.SessionFilter) ([]*SessionWithUser, error) {
	// Build query dynamically based on filter
	q := `SELECT s.id, s.user_id, s.check_in, s.check_out, s.check_out_method, u.name
          FROM sessions s
          JOIN users u ON s.user_id = u.id`

	var where []string
	var args []interface{}

	if filter.UserID != nil {
		where = append(where, "s.user_id = ?")
		args = append(args, *filter.UserID)
	}
	if filter.NameLike != nil {
		where = append(where, "u.name LIKE ?")
		args = append(args, "%"+*filter.NameLike+"%")
	}
	if filter.DiscordID != nil {
		where = append(where, "u.discord_id = ?")
		args = append(args, *filter.DiscordID)
	}
	if filter.CheckOutMethod != nil {
		where = append(where, "s.check_out_method = ?")
		args = append(args, *filter.CheckOutMethod)
	}
	if filter.From != nil {
		where = append(where, "s.check_in >= ?")
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		where = append(where, "s.check_in <= ?")
		args = append(args, *filter.To)
	}
	applySessionPresenceFilter(&where, filter)

	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}

	q += " ORDER BY " + buildSessionSortClause(filter)

	if filter.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		q += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*SessionWithUser
	for rows.Next() {
		var s domain.Session
		var checkOutMethod sql.NullString
		var userName string
		if err := rows.Scan(&s.ID, &s.UserID, &s.CheckIn, &s.CheckOut, &checkOutMethod, &userName); err != nil {
			return nil, err
		}
		if checkOutMethod.Valid {
			s.CheckOutMethod = checkOutMethod.String
		}
		sessions = append(sessions, &SessionWithUser{
			Session:  &s,
			UserName: userName,
		})
	}
	return sessions, nil
}

// Count returns total sessions matching the filter (ignores limit/offset).
func (r *SessionRepo) Count(filter query.SessionFilter) (int64, error) {
	q := `SELECT COUNT(*)
		FROM sessions s
		JOIN users u ON s.user_id = u.id`

	var where []string
	var args []interface{}

	if filter.UserID != nil {
		where = append(where, "s.user_id = ?")
		args = append(args, *filter.UserID)
	}
	if filter.NameLike != nil {
		where = append(where, "u.name LIKE ?")
		args = append(args, "%"+*filter.NameLike+"%")
	}
	if filter.DiscordID != nil {
		where = append(where, "u.discord_id = ?")
		args = append(args, *filter.DiscordID)
	}
	if filter.CheckOutMethod != nil {
		where = append(where, "s.check_out_method = ?")
		args = append(args, *filter.CheckOutMethod)
	}
	if filter.From != nil {
		where = append(where, "s.check_in >= ?")
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		where = append(where, "s.check_in <= ?")
		args = append(args, *filter.To)
	}
	applySessionPresenceFilter(&where, filter)

	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}

	var count int64
	if err := r.DB.QueryRow(q, args...).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// Update modifies an existing session's check-in and check-out times.
func (r *SessionRepo) Update(id int64, checkIn, checkOut *time.Time) error {
	q := `UPDATE sessions SET `
	var sets []string
	var args []interface{}

	if checkIn != nil {
		sets = append(sets, "check_in = ?")
		args = append(args, *checkIn)
	}
	if checkOut != nil {
		sets = append(sets, "check_out = ?")
		args = append(args, *checkOut)
		sets = append(sets, "check_out_method = ?")
		args = append(args, CheckOutMethodAPI)
	}

	if len(sets) == 0 {
		return nil // Nothing to update
	}

	q += strings.Join(sets, ", ") + " WHERE id = ?"
	args = append(args, id)

	_, err := r.DB.Exec(q, args...)
	return err
}

// Delete removes a single session by ID.
func (r *SessionRepo) Delete(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// DeleteWithFilter removes sessions matching the filter (bulk delete).
// Returns the number of sessions deleted.
func (r *SessionRepo) DeleteWithFilter(filter query.SessionFilter) (int64, error) {
	q := `DELETE FROM sessions WHERE id IN (
		SELECT s.id FROM sessions s
		JOIN users u ON s.user_id = u.id`

	var where []string
	var args []interface{}

	if filter.UserID != nil {
		where = append(where, "s.user_id = ?")
		args = append(args, *filter.UserID)
	}
	if filter.NameLike != nil {
		where = append(where, "u.name LIKE ?")
		args = append(args, "%"+*filter.NameLike+"%")
	}
	if filter.DiscordID != nil {
		where = append(where, "u.discord_id = ?")
		args = append(args, *filter.DiscordID)
	}
	if filter.CheckOutMethod != nil {
		where = append(where, "s.check_out_method = ?")
		args = append(args, *filter.CheckOutMethod)
	}
	if filter.From != nil {
		where = append(where, "s.check_in >= ?")
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		where = append(where, "s.check_in <= ?")
		args = append(args, *filter.To)
	}
	applySessionPresenceFilter(&where, filter)

	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}

	q += " ORDER BY " + buildSessionSortClause(filter)

	if filter.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		q += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	q += ")"

	result, err := r.DB.Exec(q, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetUserStats retrieves aggregated statistics for a user over a date range
func (r *SessionRepo) GetUserStats(userID int64, from, to time.Time, excludeAutoCheckout bool) (*domain.UserStats, error) {
	checkOutFilter := buildCheckOutFilter(excludeAutoCheckout)

	row := r.DB.QueryRow(`
		SELECT 
			u.id,
			u.name,
			u.discord_id,
			COALESCE(SUM((strftime('%s', substr(s.check_out, 1, 19)) - strftime('%s', substr(s.check_in, 1, 19))) / 3600.0), 0.0) as total_hours,
			COUNT(s.id) as visit_count,
			COUNT(DISTINCT substr(s.check_in, 1, 10)) as active_days,
			COALESCE(AVG((strftime('%s', substr(s.check_out, 1, 19)) - strftime('%s', substr(s.check_in, 1, 19))) / 3600.0), 0.0) as avg_duration_hours,
			MAX(s.check_out) as last_visit,
			MIN(s.check_in) as first_visit
		FROM users u
		LEFT JOIN sessions s ON u.id = s.user_id AND s.check_in >= ? AND s.check_in <= ? AND s.check_out IS NOT NULL `+checkOutFilter+`
		WHERE u.id = ?
		GROUP BY u.id`, from, to, userID)

	var stats domain.UserStats
	var lastVisit, firstVisit sql.NullString

	err := row.Scan(
		&stats.UserID,
		&stats.Name,
		&stats.DiscordID,
		&stats.TotalHours,
		&stats.VisitCount,
		&stats.ActiveDays,
		&stats.AvgDuration,
		&lastVisit,
		&firstVisit,
	)

	if err != nil {
		return nil, err
	}

	// Parse timestamps from strings
	if lastVisit.Valid {
		if t, err := time.Parse(time.RFC3339Nano, lastVisit.String); err == nil {
			stats.LastVisit = &t
		} else if t, err := time.Parse("2006-01-02 15:04:05", lastVisit.String); err == nil {
			stats.LastVisit = &t
		}
	}
	if firstVisit.Valid {
		if t, err := time.Parse(time.RFC3339Nano, firstVisit.String); err == nil {
			stats.FirstVisit = &t
		} else if t, err := time.Parse("2006-01-02 15:04:05", firstVisit.String); err == nil {
			stats.FirstVisit = &t
		}
	}

	busiestDayRow := r.DB.QueryRow(`
		SELECT
			substr(s.check_in, 1, 10) as busiest_day,
			COALESCE(SUM((strftime('%s', substr(s.check_out, 1, 19)) - strftime('%s', substr(s.check_in, 1, 19))) / 3600.0), 0.0) as busiest_day_hours
		FROM sessions s
		WHERE s.user_id = ? AND s.check_in >= ? AND s.check_in <= ? AND s.check_out IS NOT NULL `+checkOutFilter+`
		GROUP BY substr(s.check_in, 1, 10)
		ORDER BY busiest_day_hours DESC, busiest_day DESC
		LIMIT 1`, userID, from, to)

	var busiestDay sql.NullString
	var busiestDayHours sql.NullFloat64
	err = busiestDayRow.Scan(&busiestDay, &busiestDayHours)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if busiestDay.Valid {
		stats.BusiestDay = busiestDay.String
	}
	if busiestDayHours.Valid {
		stats.BusiestDayHours = busiestDayHours.Float64
	}

	return &stats, nil
}

// GetAllUserStats retrieves aggregated statistics for all users over a date range,
// optionally limited to top N users by a metric (hours or visits)
func (r *SessionRepo) GetAllUserStats(from, to time.Time, orderBy string, limit int, excludeAutoCheckout bool) ([]domain.UserStats, error) {
	checkOutFilter := buildCheckOutFilter(excludeAutoCheckout)

	// Validate orderBy to prevent SQL injection
	orderBy = normalizeRankBy(orderBy)

	orderBySQL := "total_hours DESC"
	if orderBy == "visits" {
		orderBySQL = "visit_count DESC"
	}

	query := `
		SELECT 
			u.id,
			u.name,
			u.discord_id,
			COALESCE(SUM((strftime('%s', substr(s.check_out, 1, 19)) - strftime('%s', substr(s.check_in, 1, 19))) / 3600.0), 0.0) as total_hours,
			COUNT(s.id) as visit_count,
			COUNT(DISTINCT substr(s.check_in, 1, 10)) as active_days,
			COALESCE(AVG((strftime('%s', substr(s.check_out, 1, 19)) - strftime('%s', substr(s.check_in, 1, 19))) / 3600.0), 0.0) as avg_duration_hours,
			MAX(s.check_out) as last_visit,
			MIN(s.check_in) as first_visit
		FROM users u
		LEFT JOIN sessions s ON u.id = s.user_id AND s.check_in >= ? AND s.check_in <= ? AND s.check_out IS NOT NULL ` + checkOutFilter + `
		GROUP BY u.id
		ORDER BY ` + orderBySQL

	if limit > 0 {
		query += ` LIMIT ?`
	}

	var rows *sql.Rows
	var err error
	if limit > 0 {
		rows, err = r.DB.Query(query, from, to, limit)
	} else {
		rows, err = r.DB.Query(query, from, to)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []domain.UserStats
	for rows.Next() {
		var s domain.UserStats
		var lastVisit, firstVisit sql.NullString

		err := rows.Scan(
			&s.UserID,
			&s.Name,
			&s.DiscordID,
			&s.TotalHours,
			&s.VisitCount,
			&s.ActiveDays,
			&s.AvgDuration,
			&lastVisit,
			&firstVisit,
		)

		if err != nil {
			return nil, err
		}

		// Parse timestamps from strings
		if lastVisit.Valid {
			if t, err := time.Parse(time.RFC3339Nano, lastVisit.String); err == nil {
				s.LastVisit = &t
			} else if t, err := time.Parse("2006-01-02 15:04:05", lastVisit.String); err == nil {
				s.LastVisit = &t
			}
		}
		if firstVisit.Valid {
			if t, err := time.Parse(time.RFC3339Nano, firstVisit.String); err == nil {
				s.FirstVisit = &t
			} else if t, err := time.Parse("2006-01-02 15:04:05", firstVisit.String); err == nil {
				s.FirstVisit = &t
			}
		}

		stats = append(stats, s)
	}

	return stats, rows.Err()
}

// GetPeriodStats retrieves aggregated statistics for a time period with leaderboard ranking control.
func (r *SessionRepo) GetPeriodStats(from, to time.Time, topLimit int, rankBy string, excludeAutoCheckout bool) (*domain.PeriodStats, error) {
	checkOutFilter := buildCheckOutFilter(excludeAutoCheckout)
	rankBy = normalizeRankBy(rankBy)

	// Get overall period stats
	row := r.DB.QueryRow(`
		SELECT 
			COALESCE(SUM((strftime('%s', substr(s.check_out, 1, 19)) - strftime('%s', substr(s.check_in, 1, 19))) / 3600.0), 0.0) as total_hours,
			COUNT(s.id) as total_visits,
			COUNT(DISTINCT substr(s.check_in, 1, 10)) as active_days,
			COUNT(DISTINCT s.user_id) as unique_users
		FROM sessions s
		WHERE s.check_in >= ? AND s.check_in <= ? AND s.check_out IS NOT NULL `+checkOutFilter, from, to)

	var totalHours float64
	var totalVisits, activeDays, uniqueUsers int64

	err := row.Scan(&totalHours, &totalVisits, &activeDays, &uniqueUsers)
	if err != nil {
		return nil, err
	}

	busiestDayRow := r.DB.QueryRow(`
		SELECT
			substr(s.check_in, 1, 10) as busiest_day,
			COUNT(DISTINCT s.user_id) as busiest_day_users
		FROM sessions s
		WHERE s.check_in >= ? AND s.check_in <= ? AND s.check_out IS NOT NULL `+checkOutFilter+`
		GROUP BY substr(s.check_in, 1, 10)
		ORDER BY busiest_day_users DESC, busiest_day DESC
		LIMIT 1`, from, to)

	var busiestDay sql.NullString
	var busiestDayUsers sql.NullInt64
	err = busiestDayRow.Scan(&busiestDay, &busiestDayUsers)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	peakOccupancyRow := r.DB.QueryRow(`
		WITH events AS (
			SELECT s.check_in as event_time, 1 as delta, 1 as event_order
			FROM sessions s
			WHERE s.check_in >= ? AND s.check_in <= ? AND s.check_out IS NOT NULL `+checkOutFilter+`

			UNION ALL

			SELECT s.check_out as event_time, -1 as delta, 0 as event_order
			FROM sessions s
			WHERE s.check_in >= ? AND s.check_in <= ? AND s.check_out IS NOT NULL `+checkOutFilter+`
		),
		running AS (
			SELECT
				SUM(delta) OVER (
					ORDER BY event_time, event_order
					ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
				) as occupancy
			FROM events
		)
		SELECT COALESCE(MAX(occupancy), 0) FROM running`, from, to, from, to)

	var peakOccupancy int64
	err = peakOccupancyRow.Scan(&peakOccupancy)
	if err != nil {
		return nil, err
	}

	// Get top users for this period
	topUsers, err := r.GetAllUserStats(from, to, rankBy, topLimit, excludeAutoCheckout)
	if err != nil {
		return nil, err
	}

	avgPerUser := 0.0
	if uniqueUsers > 0 {
		avgPerUser = totalHours / float64(uniqueUsers)
	}

	periodType := "custom"
	periodStr := from.Format("2006-01-02") + " to " + to.Format("2006-01-02")

	return &domain.PeriodStats{
		Period:          periodStr,
		PeriodType:      periodType,
		RankBy:          rankBy,
		StartDate:       from,
		EndDate:         to,
		TotalHours:      totalHours,
		TotalVisits:     totalVisits,
		ActiveDays:      activeDays,
		UniqueUsers:     uniqueUsers,
		BusiestDay:      busiestDay.String,
		BusiestDayUsers: busiestDayUsers.Int64,
		PeakOccupancy:   peakOccupancy,
		AveragePerUser:  avgPerUser,
		TopUsers:        topUsers,
	}, nil
}
