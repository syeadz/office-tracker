package service

import (
	"sync"
	"time"

	"office/internal/domain"
)

const DefaultEnvironmentMaxAge = 5 * time.Minute

// EnvironmentService keeps the latest environmental reading in memory.
type EnvironmentService struct {
	mu      sync.RWMutex
	latest  *domain.EnvironmentReading
	maxAge  time.Duration
	nowFunc func() time.Time
}

// NewEnvironmentService creates an in-memory environment service.
func NewEnvironmentService(maxAge time.Duration) *EnvironmentService {
	if maxAge <= 0 {
		maxAge = DefaultEnvironmentMaxAge
	}

	return &EnvironmentService{
		maxAge:  maxAge,
		nowFunc: time.Now,
	}
}

// Update stores the latest environmental reading in memory.
func (s *EnvironmentService) Update(reading domain.EnvironmentReading) {
	if reading.Timestamp.IsZero() {
		reading.Timestamp = s.nowFunc()
	}

	s.mu.Lock()
	s.latest = cloneEnvironmentReading(reading)
	s.mu.Unlock()

	TriggerEnvironmentChangeCallback()
}

// GetLatest returns the latest reading if one exists.
func (s *EnvironmentService) GetLatest() (*domain.EnvironmentReading, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.latest == nil {
		return nil, false
	}

	return cloneEnvironmentReading(*s.latest), true
}

// GetFresh returns the latest reading only if it is still within the configured freshness window.
func (s *EnvironmentService) GetFresh() (*domain.EnvironmentReading, bool) {
	return s.GetFreshWithin(s.maxAge)
}

// GetFreshWithin returns the latest reading only if it is within the provided freshness window.
func (s *EnvironmentService) GetFreshWithin(maxAge time.Duration) (*domain.EnvironmentReading, bool) {
	reading, ok := s.GetLatest()
	if !ok {
		return nil, false
	}

	if !s.IsFresh(reading, maxAge) {
		return nil, false
	}

	return reading, true
}

// IsFresh reports whether a reading should still be considered usable.
func (s *EnvironmentService) IsFresh(reading *domain.EnvironmentReading, maxAge time.Duration) bool {
	if reading == nil || reading.Timestamp.IsZero() {
		return false
	}
	if maxAge <= 0 {
		maxAge = s.maxAge
	}

	return s.nowFunc().Sub(reading.Timestamp) <= maxAge
}

// Age returns how old the reading is. Future timestamps are clamped to zero.
func (s *EnvironmentService) Age(reading *domain.EnvironmentReading) time.Duration {
	if reading == nil || reading.Timestamp.IsZero() {
		return 0
	}

	age := s.nowFunc().Sub(reading.Timestamp)
	if age < 0 {
		return 0
	}

	return age
}

func cloneEnvironmentReading(reading domain.EnvironmentReading) *domain.EnvironmentReading {
	return &domain.EnvironmentReading{
		TemperatureC: reading.TemperatureC,
		Timestamp:    reading.Timestamp,
	}
}
