package commands_test

import (
	"testing"

	"office/internal/transport/discord/commands"

	"github.com/stretchr/testify/assert"
)

func TestStatsCommands_GetApplicationCommands(t *testing.T) {
	sc := commands.NewStatsCommands(nil)
	cmds := sc.GetApplicationCommands()

	assert.Len(t, cmds, 1)
	assert.Equal(t, "stats", cmds[0].Name)
	assert.NotEmpty(t, cmds[0].Description)
	assert.GreaterOrEqual(t, len(cmds[0].Options), 3)

	var rankByFound bool
	for _, opt := range cmds[0].Options {
		if opt.Name == "rank_by" {
			rankByFound = true
			assert.Len(t, opt.Choices, 2)
			assert.Equal(t, "hours", opt.Choices[0].Value)
			assert.Equal(t, "visits", opt.Choices[1].Value)
		}
	}

	assert.True(t, rankByFound)
}
