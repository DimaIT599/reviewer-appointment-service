package services

import (
	"context"
	"fmt"
	"reviewer-appointment-service/internal/models/domain"
	"reviewer-appointment-service/internal/storage"
)

type TeamService struct {
	teamRepo storage.TeamRepository
	userRepo storage.UserRepository
}

func NewTeamService(teamRepo storage.TeamRepository, userRepo storage.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	exists, err := s.teamRepo.ExistsByName(ctx, team.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}

	if exists {
		return nil, storage.ErrTeamExists
	}

	err = s.teamRepo.Create(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	for i := range team.Users {
		user := &team.Users[i]
		user.TeamID = team.ID

		existingUser, err := s.userRepo.GetByUserID(ctx, user.UserID)
		if err != nil {
			err = s.userRepo.Create(ctx, user)
			if err != nil {
				return nil, fmt.Errorf("failed to create user %s: %w", user.UserID, err)
			}
		} else {
			existingUser.Username = user.Username
			existingUser.IsActive = user.IsActive
			existingUser.TeamID = team.ID
			err = s.userRepo.Update(ctx, existingUser)
			if err != nil {
				return nil, fmt.Errorf("failed to update user %s: %w", user.UserID, err)
			}
			user.ID = existingUser.ID
			user.CreatedAt = existingUser.CreatedAt
		}
	}

	result, err := s.teamRepo.GetWithUsers(ctx, team.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team with users: %w", err)
	}

	return result, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("%w: team not found", storage.ErrNotFound)
	}

	teamWithUsers, err := s.teamRepo.GetWithUsers(ctx, team.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team with users: %w", err)
	}

	return teamWithUsers, nil
}

func (s *TeamService) DeactivateTeamUsers(ctx context.Context, teamID int64) error {
	_, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("%w: team not found", storage.ErrNotFound)
	}

	err = s.userRepo.DeactivateByTeamID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to deactivate team users: %w", err)
	}

	return nil
}
