package commands_test

import (
	"testing"

	"office/internal/transport/discord/commands"

	"github.com/stretchr/testify/assert"
)

func TestHelpCommands_GetApplicationCommands(t *testing.T) {
	hc := commands.NewHelpCommands()
	cmds := hc.GetApplicationCommands()

	assert.Len(t, cmds, 1)
	assert.Equal(t, "help", cmds[0].Name)
	assert.NotEmpty(t, cmds[0].Description)
}

func TestHelpCommands_Create(t *testing.T) {
	hc := commands.NewHelpCommands()
	assert.NotNil(t, hc)

	cmds := hc.GetApplicationCommands()
	assert.NotNil(t, cmds)
	assert.Len(t, cmds, 1)
}
