// Package service contains the business logic for the Office Tracker application.
package service

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"office/internal/api/dto"
	"office/internal/logging"
	"office/internal/repository"
)

type AttendanceService struct {
	Users              *repository.UserRepo
	Sessions           *repository.SessionRepo
	scanHistory        []dto.ScanLog
	scanHistoryMux     sync.RWMutex
	lastScanTime       map[int64]time.Time // Track last scan time per user
	lastScanMux        sync.RWMutex
	scanCooldownMillis int // Cooldown duration in milliseconds
}

var log = logging.Component("service")

// Scan processes a scan of an RFID card. It checks if the card belongs to a known user and whether they have an open session.
// If the user has no open session, it creates one (check-in). If the user has an open session, it closes it (check-out).
// It returns a ScanResult indicating the user's name and the action taken (check-in or check-out), or an error if the card is unknown.
// Respects per-user cooldown to prevent duplicate scans within a short timeframe.
func (s *AttendanceService) Scan(uid string) (*dto.ScanResult, error) {
	user, err := s.Users.FindByRFID(uid)
	if err != nil {
		log.Error("unknown card scanned", "uid", uid)
		s.recordScan(uid, "", false, "")
		return nil, errors.New("unknown card")
	}

	// Check if user is on cooldown
	if s.IsOnCooldown(user.ID) {
		log.Warn("scan rejected - user on cooldown", "user", user.Name, "user_id", user.ID)
		return nil, errors.New("too many scans too quickly")
	}

	openSessionID, err := s.Sessions.GetOpenSession(user.ID)

	// No open session → check in
	if errors.Is(err, sql.ErrNoRows) {
		log.Info("check in", "user", user.Name)
		if err := s.Sessions.CheckIn(user.ID); err != nil {
			log.Error("db error checking in", "err", err)
			return nil, errors.New("database error")
		}
		// Update cooldown after successful scan
		s.updateLastScanTime(user.ID)
		// Record scan asynchronously (non-blocking)
		go s.recordScan(uid, user.Name, true, "check-in")
		TriggerAttendanceChangeCallback()
		return &dto.ScanResult{User: user.Name, Action: "check-in"}, nil
	} else if err != nil {
		log.Error("db error fetching open session", "err", err)
		return nil, errors.New("database error")
	}

	// Otherwise check out
	log.Info("check out", "user", user.Name)
	if err := s.Sessions.CheckOutWithMethod(openSessionID, repository.CheckOutMethodRFID); err != nil {
		log.Error("db error checking out", "err", err)
		return nil, errors.New("database error")
	}
	// Update cooldown after successful scan
	s.updateLastScanTime(user.ID)
	// Record scan asynchronously (non-blocking)
	go s.recordScan(uid, user.Name, true, "check-out")
	TriggerAttendanceChangeCallback()
	return &dto.ScanResult{User: user.Name, Action: "check-out"}, nil
}

// recordScan stores a scan log entry with UID, timestamp, user info, and action
func (s *AttendanceService) recordScan(uid, userName string, known bool, action string) {
	s.scanHistoryMux.Lock()
	defer s.scanHistoryMux.Unlock()

	// Add new entry
	s.scanHistory = append(s.scanHistory, dto.ScanLog{
		UID:       uid,
		Timestamp: time.Now(),
		UserName:  userName,
		Known:     known,
		Action:    action,
	})

	// Keep only the last 100 entries
	if len(s.scanHistory) > 100 {
		s.scanHistory = s.scanHistory[len(s.scanHistory)-100:]
	}
}

// GetScanHistory returns the list of all recent scans (both known and unknown)
func (s *AttendanceService) GetScanHistory() []dto.ScanLog {
	s.scanHistoryMux.RLock()
	defer s.scanHistoryMux.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]dto.ScanLog, len(s.scanHistory))
	copy(result, s.scanHistory)
	return result
}

// ClearScanHistory clears the scan history
func (s *AttendanceService) ClearScanHistory() {
	s.scanHistoryMux.Lock()
	defer s.scanHistoryMux.Unlock()
	s.scanHistory = nil
}

// NewAttendanceService creates a new AttendanceService with default cooldown of 500ms per user
func NewAttendanceService(users *repository.UserRepo, sessions *repository.SessionRepo) *AttendanceService {
	return &AttendanceService{
		Users:              users,
		Sessions:           sessions,
		lastScanTime:       make(map[int64]time.Time),
		scanCooldownMillis: 500, // 500ms cooldown per user
	}
}

// IsOnCooldown checks if a user is still on scan cooldown
func (s *AttendanceService) IsOnCooldown(userID int64) bool {
	s.lastScanMux.RLock()
	lastTime, exists := s.lastScanTime[userID]
	s.lastScanMux.RUnlock()

	if !exists {
		return false
	}

	return time.Since(lastTime) < time.Duration(s.scanCooldownMillis)*time.Millisecond
}

// updateLastScanTime updates the last scan time for a user
func (s *AttendanceService) updateLastScanTime(userID int64) {
	s.lastScanMux.Lock()
	defer s.lastScanMux.Unlock()
	s.lastScanTime[userID] = time.Now()
}
