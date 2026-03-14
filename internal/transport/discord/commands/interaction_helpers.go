package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// interactionUser extracts the user from an interaction, handling both slash commands and message components
func interactionUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i == nil {
		return nil
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User
	}
	return i.User
}

// interactionUserID extracts the user ID from an interaction
func interactionUserID(i *discordgo.InteractionCreate) string {
	user := interactionUser(i)
	if user == nil {
		return "unknown"
	}
	return user.ID
}

// interactionUsername extracts the username from an interaction
func interactionUsername(i *discordgo.InteractionCreate) string {
	user := interactionUser(i)
	if user == nil {
		return "unknown"
	}
	return user.Username
}

func displayRFIDTag(rfidUID string) string {
	trimmed := strings.TrimSpace(rfidUID)
	if trimmed == "" {
		return "N/A"
	}
	return trimmed
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func respondEphemeralEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func integerOptionValue(i *discordgo.InteractionCreate, optionName string) (int64, bool) {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == optionName {
			return opt.IntValue(), true
		}
	}
	return 0, false
}
