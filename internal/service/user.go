package service

import (
	"office/internal/api/dto"
	"office/internal/domain"
	"office/internal/query"
	"office/internal/repository"
)

type UserService struct {
	Users *repository.UserRepo
}

// CreateUser creates a new user and returns the DTO representation
func (s *UserService) CreateUser(name, rfidUID, discordID string) (*dto.UserResponse, error) {
	user, err := s.Users.Create(name, rfidUID, discordID)
	if err != nil {
		log.Error("error creating user", "err", err)
		return nil, err
	}

	return toUserDTO(user), nil
}

// GetUserByID retrieves a user by ID and returns the DTO representation
func (s *UserService) GetUserByID(id int64) (*dto.UserResponse, error) {
	user, err := s.Users.FindByID(id)
	if err != nil {
		log.Error("error retrieving user by ID", "id", id, "err", err)
		return nil, err
	}

	return toUserDTO(user), nil
}

// GetUserByDiscordID retrieves a user by Discord ID and returns the domain model
// Used by the Discord bot to find users for checkout operations
func (s *UserService) GetUserByDiscordID(discordID string) (*domain.User, error) {
	user, err := s.Users.FindByDiscordID(discordID)
	if err != nil {
		log.Error("error retrieving user by discord id", "discord_id", discordID, "err", err)
		return nil, err
	}
	return user, nil
}

// ListUsers retrieves users matching the optional filter and returns DTO representations
// Pass empty filter to get all users.
// Set filter.NameLike for search by name.
// Set filter.Limit/Offset for pagination.
func (s *UserService) ListUsers(filter query.UserFilter) ([]*dto.UserResponse, error) {
	users, err := s.Users.List(filter)
	if err != nil {
		log.Error("error listing users", "err", err)
		return nil, err
	}

	dtos := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		dtos[i] = toUserDTO(user)
	}
	return dtos, nil
}

// UpdateUser modifies an existing user's information and returns the DTO representation.
func (s *UserService) UpdateUser(id int64, name, rfidUID, discordID string) (*dto.UserResponse, error) {
	user, err := s.Users.Update(id, name, rfidUID, discordID)
	if err != nil {
		log.Error("error updating user", "id", id, "err", err)
		return nil, err
	}
	return toUserDTO(user), nil
}

// CountUsers returns total users matching the optional filter.
func (s *UserService) CountUsers(filter query.UserFilter) (int64, error) {
	count, err := s.Users.Count(filter)
	if err != nil {
		log.Error("error counting users", "err", err)
		return 0, err
	}
	return count, nil
}

// DeleteUser removes a single user by ID.
func (s *UserService) DeleteUser(id int64) error {
	err := s.Users.Delete(id)
	if err != nil {
		log.Error("error deleting user", "id", id, "err", err)
		return err
	}
	return nil
}

// DeleteUsers removes users matching the filter (bulk delete).
// Returns the number of users deleted.
func (s *UserService) DeleteUsers(filter query.UserFilter) (int64, error) {
	count, err := s.Users.DeleteWithFilter(filter)
	if err != nil {
		log.Error("error deleting users with filter", "err", err)
		return 0, err
	}
	return count, nil
}

// ListUsersRaw retrieves users matching the optional filter and returns domain models
// Use this for CSV export, Discord bot, or other internal purposes that need full user data including RFID
func (s *UserService) ListUsersRaw(filter query.UserFilter) ([]*domain.User, error) {
	users, err := s.Users.List(filter)
	if err != nil {
		log.Error("error listing users", "err", err)
		return nil, err
	}
	return users, nil
}

// toUserDTO converts a domain.User to dto.UserResponse
func toUserDTO(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		RFIDUID:   user.RFIDUID,
		DiscordID: user.DiscordID,
	}
}
