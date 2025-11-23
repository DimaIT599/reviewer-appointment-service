package services

import (
	"context"
	"fmt"
	"math/rand"
	"reviewer-appointment-service/internal/models/domain"
	"reviewer-appointment-service/internal/storage"
	"time"
)

const (
	StatusOpenID   = 1
	StatusMergedID = 2
	MaxReviewers   = 2
)

type PRService struct {
	prRepo   storage.PRRepository
	userRepo storage.UserRepository
	teamRepo storage.TeamRepository
}

func NewPRService(prRepo storage.PRRepository, userRepo storage.UserRepository, teamRepo storage.TeamRepository) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

func (s *PRService) CreatePR(ctx context.Context, prID, prName, authorUserID string) (*domain.PullRequest, error) {
	existingPR, err := s.prRepo.GetByPRID(ctx, prID)
	if err == nil && existingPR != nil {
		return nil, storage.ErrPRExists
	}

	author, err := s.userRepo.GetByUserID(ctx, authorUserID)
	if err != nil {
		return nil, fmt.Errorf("%w: author not found", storage.ErrNotFound)
	}

	_, err = s.teamRepo.GetByID(ctx, author.TeamID)
	if err != nil {
		return nil, fmt.Errorf("%w: author team not found", storage.ErrNotFound)
	}

	pr := &domain.PullRequest{
		PullRequestID:   prID,
		PullRequestName: prName,
		AuthorID:        author.ID,
		StatusID:        StatusOpenID,
	}

	err = s.prRepo.Create(ctx, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	candidates, err := s.getActiveTeamMembers(ctx, author.TeamID, author.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	reviewersCount := MaxReviewers
	if len(candidates) < MaxReviewers {
		reviewersCount = len(candidates)
	}

	selectedReviewers := s.selectRandomReviewers(candidates, reviewersCount)

	for _, reviewer := range selectedReviewers {
		err = s.prRepo.AddReviewer(ctx, pr.ID, reviewer.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to add reviewer: %w", err)
		}
	}

	result, err := s.prRepo.GetByPRID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created PR: %w", err)
	}

	return result, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetByPRID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("%w: PR not found", storage.ErrNotFound)
	}

	if pr.StatusID == StatusMergedID {
		return pr, nil
	}

	now := time.Now()
	pr.StatusID = StatusMergedID
	pr.MergedAt = &now

	err = s.prRepo.Update(ctx, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	result, err := s.prRepo.GetByPRID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merged PR: %w", err)
	}

	return result, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (string, error) {
	pr, err := s.prRepo.GetByPRID(ctx, prID)
	if err != nil {
		return "", fmt.Errorf("%w: PR not found", storage.ErrNotFound)
	}

	if pr.StatusID == StatusMergedID {
		return "", storage.ErrPRMerged
	}

	oldReviewer, err := s.userRepo.GetByUserID(ctx, oldUserID)
	if err != nil {
		return "", fmt.Errorf("%w: old reviewer not found", storage.ErrNotFound)
	}

	reviewers, err := s.prRepo.GetReviewers(ctx, pr.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get reviewers: %w", err)
	}

	isAssigned := false
	for _, reviewer := range reviewers {
		if reviewer.UserID == oldUserID {
			isAssigned = true
			break
		}
	}

	if !isAssigned {
		return "", storage.ErrNotAssigned
	}

	_, err = s.teamRepo.GetByID(ctx, oldReviewer.TeamID)
	if err != nil {
		return "", fmt.Errorf("%w: reviewer team not found", storage.ErrNotFound)
	}

	excludeIDs := make(map[int64]bool)
	excludeIDs[pr.AuthorID] = true
	excludeIDs[oldReviewer.ID] = true
	for _, reviewer := range reviewers {
		excludeIDs[reviewer.ID] = true
	}

	candidates, err := s.getActiveTeamMembersExcluding(ctx, oldReviewer.TeamID, excludeIDs)
	if err != nil {
		return "", fmt.Errorf("failed to get replacement candidates: %w", err)
	}

	if len(candidates) == 0 {
		return "", storage.ErrNoCandidate
	}

	newReviewer := s.selectRandomReviewers(candidates, 1)[0]

	err = s.prRepo.RemoveReviewer(ctx, pr.ID, oldReviewer.ID)
	if err != nil {
		return "", fmt.Errorf("failed to remove old reviewer: %w", err)
	}

	err = s.prRepo.AddReviewer(ctx, pr.ID, newReviewer.ID)
	if err != nil {
		return "", fmt.Errorf("failed to add new reviewer: %w", err)
	}

	return newReviewer.UserID, nil
}

func (s *PRService) getActiveTeamMembers(ctx context.Context, teamID, excludeUserID int64) ([]domain.User, error) {
	allUsers, err := s.userRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	var activeUsers []domain.User
	for _, user := range allUsers {
		if user.IsActive && user.ID != excludeUserID {
			activeUsers = append(activeUsers, user)
		}
	}

	return activeUsers, nil
}

func (s *PRService) getActiveTeamMembersExcluding(ctx context.Context, teamID int64, excludeIDs map[int64]bool) ([]domain.User, error) {
	allUsers, err := s.userRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	var activeUsers []domain.User
	for _, user := range allUsers {
		if user.IsActive && !excludeIDs[user.ID] {
			activeUsers = append(activeUsers, user)
		}
	}

	return activeUsers, nil
}

func (s *PRService) GetPR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	return s.prRepo.GetByPRID(ctx, prID)
}

func (s *PRService) selectRandomReviewers(candidates []domain.User, count int) []domain.User {
	if count <= 0 || len(candidates) == 0 {
		return []domain.User{}
	}

	if count >= len(candidates) {
		return candidates
	}

	shuffled := make([]domain.User, len(candidates))
	copy(shuffled, candidates)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:count]
}
