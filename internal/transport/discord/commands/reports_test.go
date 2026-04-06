package commands_test

import (
	"testing"

	"office/internal/domain"
	"office/internal/service"
	"office/internal/transport/discord/commands"

	"github.com/stretchr/testify/assert"
)

type noopReportDelivery struct{}

func (n *noopReportDelivery) SendPeriodReport(_ *domain.PeriodReport, _ string) error {
	return nil
}

func TestReportsToggleCommand_Definition(t *testing.T) {
	reportsSvc := service.NewReportsService(nil, &noopReportDelivery{}, true)
	cmd := commands.NewReportsToggleCommand(reportsSvc)

	def := cmd.Definition()
	assert.Equal(t, "reports", def.Name)
	assert.Len(t, def.Options, 2)

	assert.Equal(t, "weekly", def.Options[0].Name)
	assert.Equal(t, 0, len(def.Options[0].Choices))
	assert.False(t, def.Options[0].Required)

	assert.Equal(t, "monthly", def.Options[1].Name)
	assert.Equal(t, 0, len(def.Options[1].Choices))
	assert.False(t, def.Options[1].Required)
}
