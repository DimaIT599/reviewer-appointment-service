package services

import (
	"context"
	"fmt"
	"reviewer-appointment-service/internal/models/domain"
	"reviewer-appointment-service/internal/storage"
)

type UserService struct {
	userRepo storage.UserRepository
}

func NewUserService(userRepo storage.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: user not found", storage.ErrNotFound)
	}

	err = s.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to set is_active: %w", err)
	}

	user.IsActive = isActive
	return user, nil
}

func (s *UserService) GetUserReviewPRs(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	_, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: user not found", storage.ErrNotFound)
	}

	prs, err := s.userRepo.GetByReviewerID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get review PRs: %w", err)
	}

	return prs, nil
}
