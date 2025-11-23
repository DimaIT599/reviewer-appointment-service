package services

import (
	"context"
	"errors"
	"reviewer-appointment-service/internal/models/domain"
	"reviewer-appointment-service/internal/storage"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPRRepository - мок для PRRepository
type MockPRRepository struct {
	mock.Mock
}

func (m *MockPRRepository) Create(ctx context.Context, pr *domain.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRRepository) Update(ctx context.Context, pr *domain.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRRepository) GetByPRID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *MockPRRepository) GetByReviewerID(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	args := m.Called(ctx, reviewerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PullRequest), args.Error(1)
}

func (m *MockPRRepository) GetOpenPRsByUserIDs(ctx context.Context, userIDs []string) ([]domain.PullRequest, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PullRequest), args.Error(1)
}

func (m *MockPRRepository) AddReviewer(ctx context.Context, prID int64, reviewerID int64) error {
	args := m.Called(ctx, prID, reviewerID)
	return args.Error(0)
}

func (m *MockPRRepository) RemoveReviewer(ctx context.Context, prID int64, reviewerID int64) error {
	args := m.Called(ctx, prID, reviewerID)
	return args.Error(0)
}

func (m *MockPRRepository) GetReviewers(ctx context.Context, prID int64) ([]domain.User, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func TestPRService_CreatePR(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation with reviewers", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		author := &domain.User{
			ID:       1,
			UserID:   "u1",
			Username: "Author",
			IsActive: true,
			TeamID:   1,
		}

		team := &domain.Team{
			ID:   1,
			Name: "backend",
		}

		candidates := []domain.User{
			{ID: 2, UserID: "u2", Username: "Reviewer1", IsActive: true, TeamID: 1},
			{ID: 3, UserID: "u3", Username: "Reviewer2", IsActive: true, TeamID: 1},
		}

		createdPR := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        1,
			StatusID:        StatusOpenID,
			CreatedAt:       time.Now(),
			Reviewers: []domain.User{
				{ID: 2, UserID: "u2", Username: "Reviewer1", IsActive: true, TeamID: 1},
				{ID: 3, UserID: "u3", Username: "Reviewer2", IsActive: true, TeamID: 1},
			},
		}

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(nil, errors.New("not found")).Once()
		mockUserRepo.On("GetByUserID", ctx, "u1").Return(author, nil).Once()
		mockTeamRepo.On("GetByID", ctx, int64(1)).Return(team, nil).Once()
		mockPRRepo.On("Create", ctx, mock.AnythingOfType("*domain.PullRequest")).Run(func(args mock.Arguments) {
			pr := args.Get(1).(*domain.PullRequest)
			pr.ID = 1
			pr.CreatedAt = time.Now()
		}).Return(nil).Once()
		mockUserRepo.On("GetByTeamID", ctx, int64(1)).Return(candidates, nil).Once()
		mockPRRepo.On("AddReviewer", ctx, int64(1), int64(2)).Return(nil).Once()
		mockPRRepo.On("AddReviewer", ctx, int64(1), int64(3)).Return(nil).Once()
		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(createdPR, nil).Once()

		result, err := service.CreatePR(ctx, "pr-1", "Test PR", "u1")
		assert.NoError(t, err)
		assert.Equal(t, "pr-1", result.PullRequestID)
		assert.Len(t, result.Reviewers, 2)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("PR already exists", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		existingPR := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Existing PR",
			AuthorID:        1,
			StatusID:        StatusOpenID,
		}

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(existingPR, nil).Once()

		_, err := service.CreatePR(ctx, "pr-1", "Test PR", "u1")
		assert.Error(t, err)
		assert.Equal(t, storage.ErrPRExists, err)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("author not found", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(nil, errors.New("not found")).Once()
		mockUserRepo.On("GetByUserID", ctx, "non-existent").Return(nil, errors.New("not found")).Once()

		_, err := service.CreatePR(ctx, "pr-1", "Test PR", "non-existent")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound))
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("less than 2 reviewers available", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		author := &domain.User{
			ID:       1,
			UserID:   "u1",
			Username: "Author",
			IsActive: true,
			TeamID:   1,
		}

		team := &domain.Team{
			ID:   1,
			Name: "backend",
		}

		candidates := []domain.User{
			{ID: 2, UserID: "u2", Username: "Reviewer1", IsActive: true, TeamID: 1},
		}

		createdPR := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        1,
			StatusID:        StatusOpenID,
			CreatedAt:       time.Now(),
			Reviewers: []domain.User{
				{ID: 2, UserID: "u2", Username: "Reviewer1", IsActive: true, TeamID: 1},
			},
		}

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(nil, errors.New("not found")).Once()
		mockUserRepo.On("GetByUserID", ctx, "u1").Return(author, nil).Once()
		mockTeamRepo.On("GetByID", ctx, int64(1)).Return(team, nil).Once()
		mockPRRepo.On("Create", ctx, mock.AnythingOfType("*domain.PullRequest")).Run(func(args mock.Arguments) {
			pr := args.Get(1).(*domain.PullRequest)
			pr.ID = 1
			pr.CreatedAt = time.Now()
		}).Return(nil).Once()
		mockUserRepo.On("GetByTeamID", ctx, int64(1)).Return(candidates, nil).Once()
		mockPRRepo.On("AddReviewer", ctx, int64(1), int64(2)).Return(nil).Once()
		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(createdPR, nil).Once()

		result, err := service.CreatePR(ctx, "pr-1", "Test PR", "u1")
		assert.NoError(t, err)
		assert.Len(t, result.Reviewers, 1)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})
}

func TestPRService_MergePR(t *testing.T) {
	ctx := context.Background()

	t.Run("successful merge", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		pr := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        1,
			StatusID:        StatusOpenID,
			CreatedAt:       time.Now(),
		}

		mergedPR := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        1,
			StatusID:        StatusMergedID,
			CreatedAt:       time.Now(),
		}
		now := time.Now()
		mergedPR.MergedAt = &now

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(pr, nil).Once()
		mockPRRepo.On("Update", ctx, mock.AnythingOfType("*domain.PullRequest")).Return(nil).Once()
		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(mergedPR, nil).Once()

		result, err := service.MergePR(ctx, "pr-1")
		assert.NoError(t, err)
		assert.Equal(t, StatusMergedID, result.StatusID)
		assert.NotNil(t, result.MergedAt)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("idempotent merge - already merged", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		now := time.Now()
		mergedPR := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        1,
			StatusID:        StatusMergedID,
			MergedAt:        &now,
			CreatedAt:       time.Now(),
		}

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(mergedPR, nil).Once()

		result, err := service.MergePR(ctx, "pr-1")
		assert.NoError(t, err)
		assert.Equal(t, StatusMergedID, result.StatusID)
		// Update не должен вызываться
		mockPRRepo.AssertNotCalled(t, "Update", ctx, mock.Anything)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("PR not found", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		mockPRRepo.On("GetByPRID", ctx, "non-existent").Return(nil, errors.New("not found")).Once()

		_, err := service.MergePR(ctx, "non-existent")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound))
		mockPRRepo.AssertExpectations(t)
	})
}

func TestPRService_ReassignReviewer(t *testing.T) {
	ctx := context.Background()

	t.Run("successful reassignment", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		pr := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        1,
			StatusID:        StatusOpenID,
			CreatedAt:       time.Now(),
		}

		oldReviewer := &domain.User{
			ID:       2,
			UserID:   "u2",
			Username: "Old Reviewer",
			IsActive: true,
			TeamID:   1,
		}

		team := &domain.Team{
			ID:   1,
			Name: "backend",
		}

		reviewers := []domain.User{
			{ID: 2, UserID: "u2", Username: "Old Reviewer", IsActive: true, TeamID: 1},
		}

		candidates := []domain.User{
			{ID: 3, UserID: "u3", Username: "New Reviewer", IsActive: true, TeamID: 1},
		}

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(pr, nil).Once()
		mockUserRepo.On("GetByUserID", ctx, "u2").Return(oldReviewer, nil).Once()
		mockPRRepo.On("GetReviewers", ctx, int64(1)).Return(reviewers, nil).Once()
		mockTeamRepo.On("GetByID", ctx, int64(1)).Return(team, nil).Once()
		mockUserRepo.On("GetByTeamID", ctx, int64(1)).Return(candidates, nil).Once()
		mockPRRepo.On("RemoveReviewer", ctx, int64(1), int64(2)).Return(nil).Once()
		mockPRRepo.On("AddReviewer", ctx, int64(1), int64(3)).Return(nil).Once()

		newUserID, err := service.ReassignReviewer(ctx, "pr-1", "u2")
		assert.NoError(t, err)
		assert.Equal(t, "u3", newUserID)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("PR not found", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		mockPRRepo.On("GetByPRID", ctx, "non-existent").Return(nil, errors.New("not found")).Once()

		_, err := service.ReassignReviewer(ctx, "non-existent", "u2")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound))
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("PR already merged", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		now := time.Now()
		mergedPR := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        1,
			StatusID:        StatusMergedID,
			MergedAt:        &now,
			CreatedAt:       time.Now(),
		}

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(mergedPR, nil).Once()

		_, err := service.ReassignReviewer(ctx, "pr-1", "u2")
		assert.Error(t, err)
		assert.Equal(t, storage.ErrPRMerged, err)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("reviewer not assigned", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		pr := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        1,
			StatusID:        StatusOpenID,
			CreatedAt:       time.Now(),
		}

		oldReviewer := &domain.User{
			ID:       2,
			UserID:   "u2",
			Username: "Old Reviewer",
			IsActive: true,
			TeamID:   1,
		}

		reviewers := []domain.User{
			{ID: 3, UserID: "u3", Username: "Other Reviewer", IsActive: true, TeamID: 1},
		}

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(pr, nil).Once()
		mockUserRepo.On("GetByUserID", ctx, "u2").Return(oldReviewer, nil).Once()
		mockPRRepo.On("GetReviewers", ctx, int64(1)).Return(reviewers, nil).Once()

		_, err := service.ReassignReviewer(ctx, "pr-1", "u2")
		assert.Error(t, err)
		assert.Equal(t, storage.ErrNotAssigned, err)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("no replacement candidate", func(t *testing.T) {
		mockPRRepo := new(MockPRRepository)
		mockUserRepo := new(MockUserRepository)
		mockTeamRepo := new(MockTeamRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

		pr := &domain.PullRequest{
			ID:              1,
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        1,
			StatusID:        StatusOpenID,
			CreatedAt:       time.Now(),
		}

		oldReviewer := &domain.User{
			ID:       2,
			UserID:   "u2",
			Username: "Old Reviewer",
			IsActive: true,
			TeamID:   1,
		}

		team := &domain.Team{
			ID:   1,
			Name: "backend",
		}

		reviewers := []domain.User{
			{ID: 2, UserID: "u2", Username: "Old Reviewer", IsActive: true, TeamID: 1},
		}

		mockPRRepo.On("GetByPRID", ctx, "pr-1").Return(pr, nil).Once()
		mockUserRepo.On("GetByUserID", ctx, "u2").Return(oldReviewer, nil).Once()
		mockPRRepo.On("GetReviewers", ctx, int64(1)).Return(reviewers, nil).Once()
		mockTeamRepo.On("GetByID", ctx, int64(1)).Return(team, nil).Once()
		mockUserRepo.On("GetByTeamID", ctx, int64(1)).Return([]domain.User{}, nil).Once()

		_, err := service.ReassignReviewer(ctx, "pr-1", "u2")
		assert.Error(t, err)
		assert.Equal(t, storage.ErrNoCandidate, err)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})
}

