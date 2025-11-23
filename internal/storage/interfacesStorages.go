package storage

import (
	"context"
	"reviewer-appointment-service/internal/models/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	GetByUserID(ctx context.Context, userID string) (*domain.User, error)
	GetByTeamID(ctx context.Context, teamID int64) ([]domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	GetByReviewerID(ctx context.Context, userID string) ([]domain.PullRequest, error)
	DeactivateByTeamID(ctx context.Context, teamID int64) error
}

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	GetByName(ctx context.Context, teamName string) (*domain.Team, error)
	GetByID(ctx context.Context, teamID int64) (*domain.Team, error)
	GetWithUsers(ctx context.Context, teamID int64) (*domain.Team, error)
	GetAllWithUsers(ctx context.Context) ([]domain.Team, error)
	ExistsByName(ctx context.Context, teamName string) (bool, error)
}

type PRRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	Update(ctx context.Context, pr *domain.PullRequest) error
	GetByPRID(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetByReviewerID(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
	GetOpenPRsByUserIDs(ctx context.Context, userIDs []string) ([]domain.PullRequest, error)
	AddReviewer(ctx context.Context, prID int64, reviewerID int64) error
	RemoveReviewer(ctx context.Context, prID int64, reviewerID int64) error
	GetReviewers(ctx context.Context, prID int64) ([]domain.User, error)
}

type StatsRepository interface {
	GetTotalPRs(ctx context.Context) (int, error)
	GetTotalUsers(ctx context.Context) (int, error)
	GetActiveUsers(ctx context.Context) (int, error)
	GetPRsByStatus(ctx context.Context) (map[string]int, error)
	GetTopReviewers(ctx context.Context, limit int) ([]domain.ReviewerStats, error)
}
