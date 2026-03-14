package service

import (
	"database/sql"
	"errors"
	"time"

	"office/internal/api/dto"
	"office/internal/query"
	"office/internal/repository"
)

type SessionService struct {
	Sessions *repository.SessionRepo
}

var (
	ErrSessionAlreadyOpen = errors.New("session already open")
	ErrNoOpenSession      = errors.New("no open session")
)

// GetSessionByID retrieves a specific session by ID and returns the domain model with user info
// Used by Discord bot and HTTP endpoints for single session lookups
func (s *SessionService) GetSessionByID(id int64) (*repository.SessionWithUser, error) {
	session, err := s.Sessions.FindByID(id)
	if err != nil {
		log.Error("error retrieving session by ID", "id", id, "err", err)
		return nil, err
	}
	return session, nil
}

// ListSessions is the unified method for retrieving sessions with optional filters.
// Set returnDTO=false to get raw repository data (for CSV export, Discord bot).
// Set returnDTO=true to get SessionResponse DTOs for HTTP APIs.
// For all sessions: empty filter.
// For user sessions: filter.UserID or filter.DiscordID.
// For date range: filter.From and filter.To.
func (s *SessionService) ListSessions(filter query.SessionFilter, asDTO bool) (interface{}, error) {
	sessions, err := s.Sessions.List(filter)
	if err != nil {
		log.Error("error retrieving sessions", "err", err)
		return nil, err
	}

	// Return raw data for CSV export, Discord bot, etc.
	if !asDTO {
		return sessions, nil
	}

	// Return SessionResponse DTOs for HTTP session APIs.
	responses := make([]dto.SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		responses = append(responses, dto.SessionResponse{
			ID:       s.ID,
			UserID:   s.UserID,
			UserName: s.UserName,
			CheckIn:  s.CheckIn,
			CheckOut: s.CheckOut,
			Active:   s.CheckOut == nil,
		})
	}
	return responses, nil
}

// CountSessions returns total sessions matching the optional filter.
func (s *SessionService) CountSessions(filter query.SessionFilter) (int64, error) {
	count, err := s.Sessions.Count(filter)
	if err != nil {
		log.Error("error counting sessions", "err", err)
		return 0, err
	}
	return count, nil
}

// UpdateSession modifies an existing session's check-in and check-out times.
func (s *SessionService) UpdateSession(id int64, checkIn, checkOut *time.Time) error {
	err := s.Sessions.Update(id, checkIn, checkOut)
	if err != nil {
		log.Error("error updating session", "id", id, "err", err)
		return err
	}
	return nil
}

// DeleteSession removes a single session by ID.
func (s *SessionService) DeleteSession(id int64) error {
	err := s.Sessions.Delete(id)
	if err != nil {
		log.Error("error deleting session", "id", id, "err", err)
		return err
	}
	return nil
}

// DeleteSessions removes sessions matching the filter (bulk delete).
// Returns the number of sessions deleted.
func (s *SessionService) DeleteSessions(filter query.SessionFilter) (int64, error) {
	count, err := s.Sessions.DeleteWithFilter(filter)
	if err != nil {
		log.Error("error deleting sessions with filter", "err", err)
		return 0, err
	}
	return count, nil
}

// CheckInUser creates a new session for a user if none is currently open.
func (s *SessionService) CheckInUser(userID int64) error {
	_, err := s.Sessions.GetOpenSession(userID)
	if err == nil {
		return ErrSessionAlreadyOpen
	}
	if !errors.Is(err, sql.ErrNoRows) {
		log.Error("error checking for open session", "user_id", userID, "err", err)
		return err
	}

	if err := s.Sessions.CheckIn(userID); err != nil {
		log.Error("error checking in user", "user_id", userID, "err", err)
		return err
	}
	// Attendance changed — notify listeners
	TriggerAttendanceChangeCallback()
	return nil
}

// CheckOutUser checks out the currently open session for a user.
func (s *SessionService) CheckOutUser(userID int64) error {
	return s.CheckOutUserWithMethod(userID, repository.CheckOutMethodAPI)
}

// CheckOutUserWithMethod checks out the currently open session for a user using the provided checkout method.
func (s *SessionService) CheckOutUserWithMethod(userID int64, method string) error {
	openSessionID, err := s.Sessions.GetOpenSession(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoOpenSession
		}
		log.Error("error fetching open session", "user_id", userID, "err", err)
		return err
	}

	if err := s.Sessions.CheckOutWithMethod(openSessionID, method); err != nil {
		log.Error("error checking out user", "user_id", userID, "session_id", openSessionID, "err", err)
		return err
	}
	// Attendance changed — notify listeners
	TriggerAttendanceChangeCallback()
	return nil
}
