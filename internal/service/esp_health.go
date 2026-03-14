package service

import (
	"sort"
	"sync"
	"time"

	"office/internal/domain"
)

const DefaultESPHealthMaxAge = 20 * time.Minute

// ESPHealthService stores latest heartbeat information from ESP32 devices in memory.
type ESPHealthService struct {
	mu      sync.RWMutex
	latest  map[string]domain.ESPHealthStatus
	maxAge  time.Duration
	nowFunc func() time.Time
}

// NewESPHealthService creates a new in-memory ESP health service.
func NewESPHealthService(maxAge time.Duration) *ESPHealthService {
	if maxAge <= 0 {
		maxAge = DefaultESPHealthMaxAge
	}

	return &ESPHealthService{
		latest:  make(map[string]domain.ESPHealthStatus),
		maxAge:  maxAge,
		nowFunc: time.Now,
	}
}

// Update stores the latest status for one ESP32 device.
func (s *ESPHealthService) Update(status domain.ESPHealthStatus) {
	if status.DeviceID == "" {
		status.DeviceID = "default"
	}
	status.UpdatedAt = s.nowFunc()

	s.mu.Lock()
	s.latest[status.DeviceID] = status
	s.mu.Unlock()
}

// GetAll returns the latest status for all devices sorted by device ID.
func (s *ESPHealthService) GetAll() []domain.ESPHealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]domain.ESPHealthStatus, 0, len(s.latest))
	for _, v := range s.latest {
		items = append(items, v)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].DeviceID < items[j].DeviceID
	})

	return items
}

// IsFresh reports whether the device status is still fresh.
func (s *ESPHealthService) IsFresh(status domain.ESPHealthStatus) bool {
	return s.nowFunc().Sub(status.UpdatedAt) <= s.maxAge
}

// Age returns status age in seconds, clamped at zero.
func (s *ESPHealthService) Age(status domain.ESPHealthStatus) time.Duration {
	age := s.nowFunc().Sub(status.UpdatedAt)
	if age < 0 {
		return 0
	}
	return age
}
