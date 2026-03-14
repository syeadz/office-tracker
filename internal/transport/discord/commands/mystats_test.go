package commands_test

import (
	"testing"

	"office/internal/transport/discord/commands"

	"github.com/stretchr/testify/assert"
)

func TestMyStatsCommands_GetApplicationCommands(t *testing.T) {
	mc := commands.NewMyStatsCommands(nil, nil)
	cmds := mc.GetApplicationCommands()

	assert.Len(t, cmds, 1)
	assert.Equal(t, "mystats", cmds[0].Name)
	assert.NotEmpty(t, cmds[0].Description)
	assert.GreaterOrEqual(t, len(cmds[0].Options), 1)
}
