package commands

import (
	"fmt"
	"strings"

	"office/internal/logging"
	"office/internal/query"
	"office/internal/service"

	"github.com/bwmarrin/discordgo"
)

var log = logging.Component("discord.commands")

// UserCommands handles all user-related slash commands
type UserCommands struct {
	userSvc *service.UserService
}

// NewUserCommands creates a new UserCommands handler
func NewUserCommands(userSvc *service.UserService) *UserCommands {
	return &UserCommands{userSvc: userSvc}
}

// GetApplicationCommands returns the slash command definitions for user management
func (uc *UserCommands) GetApplicationCommands() []*discordgo.ApplicationCommand {
	adminPerm := int64(discordgo.PermissionAdministrator)

	return []*discordgo.ApplicationCommand{
		{
			Name:        "user-create",
			Description: "Create a new user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "member",
					Description: "Discord member to link",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
				{
					Name:        "rfid_uid",
					Description: "User's RFID card UID",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "name",
					Description: "User's display name (optional, defaults to member name)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
			},
		},
		{
			Name:        "user-get",
			Description: "Get a user by ID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "user_id",
					Description: "The user's ID",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		{
			Name:                     "user-update",
			Description:              "Update a user's information",
			DefaultMemberPermissions: &adminPerm,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "user_id",
					Description: "The user's ID",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
				{
					Name:        "name",
					Description: "New display name",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "member",
					Description: "Discord member to link (optional)",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    false,
				},
				{
					Name:        "rfid_uid",
					Description: "New RFID UID (optional)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
			},
		},
		{
			Name:                     "user-delete",
			Description:              "Delete a user",
			DefaultMemberPermissions: &adminPerm,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "user_id",
					Description: "The user's ID",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		{
			Name:        "user-list",
			Description: "List users with optional search and pagination",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "search",
					Description: "Search by user name (optional)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
				{
					Name:        "limit",
					Description: "Maximum results (default: 10, max: 50)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
				{
					Name:        "page",
					Description: "Page number (default: 1)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
				{
					Name:        "order",
					Description: "Sort order: 'asc' or 'desc' (default: asc)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Ascending", Value: "asc"},
						{Name: "Descending", Value: "desc"},
					},
				},
				{
					Name:        "sort_by",
					Description: "Sort field: 'name' or 'created_at' (default: name)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Name", Value: "name"},
						{Name: "Created At", Value: "created_at"},
					},
				},
			},
		},
	}
}

// HandleCommand routes user commands to appropriate handlers
func (uc *UserCommands) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmdName string) {
	log.Info("user command received", "command", cmdName, "user_id", interactionUserID(i), "username", interactionUsername(i))

	switch cmdName {
	case "user-create":
		uc.handleCreate(s, i)
	case "user-get":
		uc.handleGet(s, i)
	case "user-update":
		uc.handleUpdate(s, i)
	case "user-delete":
		uc.handleDelete(s, i)
	case "user-list":
		uc.handleList(s, i)
	}
}

// handleCreate creates a new user
func (uc *UserCommands) handleCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options

	name := ""
	rfidUID := ""
	discordID := ""
	defaultName := ""

	for _, opt := range opts {
		switch opt.Name {
		case "member":
			targetUser := opt.UserValue(s)
			if targetUser != nil {
				discordID = targetUser.ID

				// Prefer server-specific nickname/display name when available.
				if resolvedMember, ok := i.ApplicationCommandData().Resolved.Members[targetUser.ID]; ok {
					if strings.TrimSpace(resolvedMember.Nick) != "" {
						defaultName = strings.TrimSpace(resolvedMember.Nick)
					}
				}

				if strings.TrimSpace(defaultName) == "" {
					if strings.TrimSpace(targetUser.GlobalName) != "" {
						defaultName = strings.TrimSpace(targetUser.GlobalName)
					} else {
						defaultName = strings.TrimSpace(targetUser.Username)
					}
				}
			}
		case "name":
			name = opt.StringValue()
		case "rfid_uid":
			rfidUID = opt.StringValue()
		}
	}

	if discordID == "" {
		respondEphemeral(s, i, "❌ Member is required.")
		return
	}

	if strings.TrimSpace(name) == "" {
		name = defaultName
	}

	user, err := uc.userSvc.CreateUser(name, rfidUID, discordID)
	if err != nil {
		log.Error("failed to create user", "name", name, "rfid_uid", rfidUID, "error", err)
		respondEphemeral(s, i, "❌ Failed to create user. Please try again.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ User Created",
		Color:       0x2ECC71,
		Description: fmt.Sprintf("**%s** has been added to the system.", user.Name),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: fmt.Sprintf("%d", user.ID), Inline: true},
			{Name: "Name", Value: user.Name, Inline: true},
			{Name: "RFID UID", Value: user.RFIDUID, Inline: true},
			{Name: "Discord ID", Value: user.DiscordID, Inline: true},
		},
	}

	respondEphemeralEmbed(s, i, embed)
}

// handleGet retrieves a single user by ID
func (uc *UserCommands) handleGet(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, ok := integerOptionValue(i, "user_id")
	if !ok {
		respondEphemeral(s, i, "❌ User ID is required.")
		return
	}

	user, err := uc.userSvc.GetUserByID(userID)
	if err != nil {
		log.Error("failed to get user", "user_id", userID, "error", err)
		respondEphemeral(s, i, "❌ User not found.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "👤 User Details",
		Color: 0x3498DB,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: fmt.Sprintf("%d", user.ID), Inline: true},
			{Name: "Name", Value: user.Name, Inline: true},
			{Name: "RFID UID", Value: displayRFIDTag(user.RFIDUID), Inline: true},
			{Name: "Discord ID", Value: user.DiscordID, Inline: true},
		},
	}

	respondEphemeralEmbed(s, i, embed)
}

// handleUpdate updates a user's information
func (uc *UserCommands) handleUpdate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options

	userID := int64(0)
	name := ""
	selectedDiscordID := ""
	selectedRFIDUID := ""
	hasSelectedMember := false
	hasSelectedRFIDUID := false

	for _, opt := range opts {
		switch opt.Name {
		case "user_id":
			userID = opt.IntValue()
		case "name":
			name = opt.StringValue()
		case "member":
			targetUser := opt.UserValue(s)
			if targetUser != nil {
				selectedDiscordID = targetUser.ID
				hasSelectedMember = true
			}
		case "rfid_uid":
			selectedRFIDUID = strings.TrimSpace(opt.StringValue())
			hasSelectedRFIDUID = true
		}
	}

	currentUser, err := uc.userSvc.GetUserByID(userID)
	if err != nil {
		log.Error("failed to get user for update", "user_id", userID, "error", err)
		respondEphemeral(s, i, "❌ Failed to update user. User may not exist.")
		return
	}

	discordID := currentUser.DiscordID
	if hasSelectedMember {
		discordID = selectedDiscordID
	}

	rfidUID := currentUser.RFIDUID
	if hasSelectedRFIDUID {
		rfidUID = selectedRFIDUID
	}

	user, err := uc.userSvc.UpdateUser(userID, name, rfidUID, discordID)
	if err != nil {
		log.Error("failed to update user", "user_id", userID, "error", err)
		respondEphemeral(s, i, "❌ Failed to update user. User may not exist.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ User Updated",
		Color:       0x2ECC71,
		Description: fmt.Sprintf("**%s** has been updated.", user.Name),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ID", Value: fmt.Sprintf("%d", user.ID), Inline: true},
			{Name: "Name", Value: user.Name, Inline: true},
			{Name: "RFID UID", Value: displayRFIDTag(user.RFIDUID), Inline: true},
			{Name: "Discord ID", Value: user.DiscordID, Inline: true},
		},
	}

	respondEphemeralEmbed(s, i, embed)
}

// handleDelete deletes a user
func (uc *UserCommands) handleDelete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, ok := integerOptionValue(i, "user_id")
	if !ok {
		respondEphemeral(s, i, "❌ User ID is required.")
		return
	}

	// Get user info before deletion for response
	user, err := uc.userSvc.GetUserByID(userID)
	if err != nil {
		log.Error("failed to get user", "user_id", userID, "error", err)
		respondEphemeral(s, i, "❌ User not found.")
		return
	}

	err = uc.userSvc.DeleteUser(userID)
	if err != nil {
		log.Error("failed to delete user", "user_id", userID, "error", err)
		respondEphemeral(s, i, "❌ Failed to delete user.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ User Deleted",
		Color:       0xE74C3C,
		Description: fmt.Sprintf("**%s** has been removed from the system.", user.Name),
	}

	respondEphemeralEmbed(s, i, embed)
}

// handleList lists users with optional search and pagination
func (uc *UserCommands) handleList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options

	filter := query.UserFilter{}
	limit := int64(10)
	page := int64(1)

	for _, opt := range opts {
		switch opt.Name {
		case "search":
			search := opt.StringValue()
			if search != "" {
				filter.NameLike = &search
			}
		case "limit":
			l := opt.IntValue()
			if l > 0 && l <= 50 {
				limit = l
			}
		case "page":
			p := opt.IntValue()
			if p > 0 {
				page = p
			}
		case "order":
			order := opt.StringValue()
			if order == "asc" || order == "desc" {
				filter.OrderBy = order
			}
		case "sort_by":
			sortBy := opt.StringValue()
			if sortBy == "name" || sortBy == "created_at" {
				filter.SortBy = sortBy
			}
		}
	}

	filter.Limit = int(limit)
	filter.Offset = int((page - 1) * limit)

	countFilter := filter
	countFilter.Limit = 0
	countFilter.Offset = 0
	totalUsers, err := uc.userSvc.CountUsers(countFilter)
	if err != nil {
		log.Error("failed to count users", "error", err)
		respondEphemeral(s, i, "❌ Failed to fetch users.")
		return
	}

	users, err := uc.userSvc.ListUsers(filter)
	if err != nil {
		log.Error("failed to list users", "error", err)
		respondEphemeral(s, i, "❌ Failed to fetch users.")
		return
	}

	if len(users) == 0 {
		respondEphemeral(s, i, "📭 No users found.")
		return
	}

	// Build user list
	var userLines []string
	for _, user := range users {
		discordIDStr := user.DiscordID
		if discordIDStr == "" {
			discordIDStr = "Not linked"
		}
		userLines = append(userLines, fmt.Sprintf("**%s** (ID: %d) - RFID: %s - Discord: %s", user.Name, user.ID, displayRFIDTag(user.RFIDUID), discordIDStr))
	}

	description := strings.Join(userLines, "\n")
	if len(description) > 2048 {
		description = description[:2048]
		description += "\n\n*Output truncated...*"
	}

	title := "👥 User List"
	if filter.NameLike != nil && *filter.NameLike != "" {
		title += fmt.Sprintf(" (search: \"%s\")", *filter.NameLike)
	}

	totalPages := int64(1)
	if limit > 0 {
		totalPages = (totalUsers + limit - 1) / limit
		if totalPages < 1 {
			totalPages = 1
		}
	}
	if page > totalPages {
		page = totalPages
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0x3498DB,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d of %d (showing %d of %d users)", page, totalPages, len(users), totalUsers),
		},
	}

	respondEphemeralEmbed(s, i, embed)
}
