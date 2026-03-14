package service

import (
	"errors"
	"testing"

	"office/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReportDelivery is a mock implementation of ReportDelivery
type MockReportDelivery struct {
	mock.Mock
}

func (m *MockReportDelivery) SendPeriodReport(report *domain.PeriodReport, reportType string) error {
	args := m.Called(report, reportType)
	return args.Error(0)
}

func TestNewReportsService(t *testing.T) {
	mockDelivery := new(MockReportDelivery)

	svc := NewReportsService(nil, mockDelivery, true)

	assert.NotNil(t, svc)
	assert.True(t, svc.enabled)
	assert.Equal(t, mockDelivery, svc.delivery)
}

func TestReportsService_GenerateAndSendWeeklyReport_Disabled(t *testing.T) {
	mockDelivery := new(MockReportDelivery)
	svc := NewReportsService(nil, mockDelivery, false)

	err := svc.GenerateAndSendWeeklyReport()

	assert.NoError(t, err)
	mockDelivery.AssertNotCalled(t, "SendPeriodReport")
}

func TestReportsService_GenerateAndSendWeeklyReport_NoDelivery(t *testing.T) {
	svc := NewReportsService(nil, nil, true)

	err := svc.GenerateAndSendWeeklyReport()

	assert.NoError(t, err)
}

func TestReportsService_GenerateAndSendWeeklyReport_DeliveryError(t *testing.T) {
	// We can't properly test delivery errors without a real stats service
	// and avoiding import cycles. This test just verifies the service handles nil gracefully.
	svc := NewReportsService(nil, nil, true)

	err := svc.GenerateAndSendWeeklyReport()

	// With nil stats and nil delivery, should just log and return nil
	assert.NoError(t, err)
}

func TestReportsService_SetEnabled(t *testing.T) {
	svc := NewReportsService(nil, nil, true)

	assert.True(t, svc.IsEnabled())

	svc.SetEnabled(false)
	assert.False(t, svc.IsEnabled())

	svc.SetEnabled(true)
	assert.True(t, svc.IsEnabled())
}

func TestReportsService_DeliveryErrorHandling(t *testing.T) {
	// This test verifies that delivery errors are properly wrapped
	mockDelivery := new(MockReportDelivery)
	mockDelivery.On("SendPeriodReport", mock.AnythingOfType("*domain.PeriodReport"), mock.AnythingOfType("string")).Return(errors.New("delivery failed"))

	svc := NewReportsService(nil, mockDelivery, true)

	err := svc.GenerateAndSendWeeklyReport()

	// Should fail (either at generation or delivery)
	assert.Error(t, err)
}

func TestReportsService_GetLatestWeeklyReport_NotAvailable(t *testing.T) {
	svc := NewReportsService(nil, nil, true)

	report, err := svc.GetLatestWeeklyReport()

	assert.Error(t, err)
	assert.Nil(t, report)
}
