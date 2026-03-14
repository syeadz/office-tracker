package service

import (
	"testing"
	"time"

	"office/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentService_UpdateAndGetFresh(t *testing.T) {
	now := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	svc := NewEnvironmentService(5 * time.Minute)
	svc.nowFunc = func() time.Time { return now }

	svc.Update(domain.EnvironmentReading{
		TemperatureC: 24.5,
		Timestamp:    now.Add(-2 * time.Minute),
	})

	latest, ok := svc.GetLatest()
	require.True(t, ok)
	assert.Equal(t, 24.5, latest.TemperatureC)

	fresh, ok := svc.GetFresh()
	require.True(t, ok)
	assert.Equal(t, latest.Timestamp, fresh.Timestamp)
	assert.Equal(t, 2*time.Minute, svc.Age(fresh))
}

func TestEnvironmentService_GetFreshReturnsFalseForStaleReading(t *testing.T) {
	now := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	svc := NewEnvironmentService(5 * time.Minute)
	svc.nowFunc = func() time.Time { return now }

	svc.Update(domain.EnvironmentReading{
		TemperatureC: 30.1,
		Timestamp:    now.Add(-6 * time.Minute),
	})

	fresh, ok := svc.GetFresh()
	assert.False(t, ok)
	assert.Nil(t, fresh)
}

func TestEnvironmentService_UsesCurrentTimeWhenTimestampMissing(t *testing.T) {
	now := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	svc := NewEnvironmentService(5 * time.Minute)
	svc.nowFunc = func() time.Time { return now }

	svc.Update(domain.EnvironmentReading{TemperatureC: 22.0})

	latest, ok := svc.GetLatest()
	require.True(t, ok)
	assert.Equal(t, now, latest.Timestamp)
}
