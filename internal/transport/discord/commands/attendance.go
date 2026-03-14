package commands

import (
	"errors"
	"fmt"

	"office/internal/logging"
	"office/internal/query"
	"office/internal/repository"
	"office/internal/service"

	"github.com/bwmarrin/discordgo"
)

var attendanceLog = logging.Component("discord.commands.attendance")

// Color constants for embeds
const (
	ColorGreen = 0x2ECC71
)

// AttendanceCommands handles attendance operations (check-in/check-out).
type AttendanceCommands struct {
	userSvc    *service.UserService
	sessionSvc *service.SessionService
}

// NewAttendanceCommands creates a new AttendanceCommands handler.
func NewAttendanceCommands(userSvc *service.UserService, sessionSvc *service.SessionService) *AttendanceCommands {
	return &AttendanceCommands{
		userSvc:    userSvc,
		sessionSvc: sessionSvc,
	}
}

// GetApplicationCommands returns the slash command definitions for attendance operations.
func (ac *AttendanceCommands) GetApplicationCommands() []*discordgo.ApplicationCommand {
	adminPerm := int64(discordgo.PermissionAdministrator)

	return []*discordgo.ApplicationCommand{
		{
			Name:                     "checkin",
			Description:              "Check in a member (start a session)",
			DefaultMemberPermissions: &adminPerm,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "member",
					Description: "The member to check in",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
		},
		{
			Name:        "checkout",
			Description: "Check out a member",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "member",
					Description: "The member to check out",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
		},
		{
			Name:        "checkout-all",
			Description: "Check out all currently checked-in members",
		},
	}
}

// HandleCommand routes attendance commands to appropriate handlers.
func (ac *AttendanceCommands) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmdName string) {
	attendanceLog.Info("checkout command received", "command", cmdName, "user_id", interactionUserID(i), "username", interactionUsername(i))

	switch cmdName {
	case "checkin":
		ac.handleCheckin(s, i)
	case "checkout":
		ac.handleCheckout(s, i)
	case "checkout-all":
		ac.handleCheckoutAll(s, i)
	}
}

func selectedMemberDiscordID(s *discordgo.Session, i *discordgo.InteractionCreate) (string, bool) {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name != "member" {
			continue
		}
		targetUser := opt.UserValue(s)
		if targetUser == nil || targetUser.ID == "" {
			return "", false
		}
		return targetUser.ID, true
	}
	return "", false
}

// handleCheckin checks in a specific member by Discord ID
func (ac *AttendanceCommands) handleCheckin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	targetDiscordID, ok := selectedMemberDiscordID(s, i)
	if !ok {
		respondEphemeral(s, i, "❌ Invalid member specified.")
		return
	}

	// Find user by Discord ID
	user, err := ac.userSvc.GetUserByDiscordID(targetDiscordID)
	if err != nil {
		attendanceLog.Error("user not found", "discord_id", targetDiscordID, "error", err)
		respondEphemeral(s, i, "❌ User not found in system.")
		return
	}

	// Check them in
	err = ac.sessionSvc.CheckInUser(user.ID)
	if err != nil {
		if errors.Is(err, service.ErrSessionAlreadyOpen) {
			respondEphemeral(s, i, fmt.Sprintf("❌ **%s** is already checked in.", user.Name))
			return
		}
		attendanceLog.Error("failed to check in", "user_id", user.ID, "error", err)
		respondEphemeral(s, i, "❌ Failed to check in member. Please try again.")
		return
	}

	attendanceLog.Info("member checked in", "user_id", user.ID, "user_name", user.Name)

	respondEphemeralEmbed(s, i, &discordgo.MessageEmbed{
		Title:       "✅ Member Checked In",
		Description: fmt.Sprintf("**%s** has been checked in.", user.Name),
		Color:       ColorGreen,
	})
}

// handleCheckout checks out a specific member by Discord ID.
func (ac *AttendanceCommands) handleCheckout(s *discordgo.Session, i *discordgo.InteractionCreate) {
	targetDiscordID, ok := selectedMemberDiscordID(s, i)
	if !ok {
		respondEphemeral(s, i, "❌ Invalid member specified.")
		return
	}

	// Find user by Discord ID
	user, err := ac.userSvc.GetUserByDiscordID(targetDiscordID)
	if err != nil {
		attendanceLog.Error("user not found", "discord_id", targetDiscordID, "error", err)
		respondEphemeral(s, i, "❌ User not found in system.")
		return
	}

	// Check them out using the service layer
	err = ac.sessionSvc.CheckOutUserWithMethod(user.ID, repository.CheckOutMethodDiscord)
	if err != nil {
		if errors.Is(err, service.ErrNoOpenSession) {
			attendanceLog.Error("no open session found", "user_id", user.ID, "error", err)
			respondEphemeral(s, i, fmt.Sprintf("❌ **%s** has no active session.", user.Name))
			return
		}
		attendanceLog.Error("failed to check out", "user_id", user.ID, "error", err)
		respondEphemeral(s, i, "❌ Failed to check out member. Please try again.")
		return
	}

	attendanceLog.Info("member checked out", "user_id", user.ID, "user_name", user.Name)

	respondEphemeralEmbed(s, i, &discordgo.MessageEmbed{
		Title:       "✅ Member Checked Out",
		Description: fmt.Sprintf("**%s** has been checked out.", user.Name),
		Color:       ColorGreen,
	})
}

// handleCheckoutAll checks out all currently active sessions
func (ac *AttendanceCommands) handleCheckoutAll(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get all active sessions
	filter := query.SessionFilter{ActiveOnly: true}
	result, err := ac.sessionSvc.ListSessions(filter, false)
	if err != nil {
		attendanceLog.Error("failed to get active sessions", "error", err)
		respondEphemeral(s, i, "❌ Failed to retrieve active sessions.")
		return
	}

	sessions, ok := result.([]*repository.SessionWithUser)
	if !ok || len(sessions) == 0 {
		respondEphemeral(s, i, "📭 No active sessions to check out.")
		return
	}

	// Check out all active sessions using the service layer
	checkedOutCount := 0
	failedCount := 0
	var checkedOutUsers []string

	for _, sess := range sessions {
		err := ac.sessionSvc.CheckOutUserWithMethod(sess.UserID, repository.CheckOutMethodDiscord)
		if err != nil {
			attendanceLog.Error("failed to check out user", "user_name", sess.UserName, "error", err)
			failedCount++
			continue
		}
		checkedOutCount++
		checkedOutUsers = append(checkedOutUsers, sess.UserName)
	}

	description := fmt.Sprintf("Checked out **%d** member(s)", checkedOutCount)
	if failedCount > 0 {
		description += fmt.Sprintf(" (**%d** failed)", failedCount)
	}
	if len(checkedOutUsers) > 0 {
		description += ":\n"
		for _, name := range checkedOutUsers {
			description += fmt.Sprintf("• %s\n", name)
		}
	}

	respondEphemeralEmbed(s, i, &discordgo.MessageEmbed{
		Title:       "✅ Bulk Checkout",
		Description: description,
		Color:       ColorGreen,
	})

	attendanceLog.Info("bulk checkout completed", "count", checkedOutCount)
}
