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

// MockUserRepository - мок для UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByUserID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByTeamID(ctx context.Context, teamID int64) ([]domain.User, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	args := m.Called(ctx, userID, isActive)
	return args.Error(0)
}

func (m *MockUserRepository) GetByReviewerID(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PullRequest), args.Error(1)
}

func (m *MockUserRepository) DeactivateByTeamID(ctx context.Context, teamID int64) error {
	args := m.Called(ctx, teamID)
	return args.Error(0)
}

func TestUserService_SetIsActive(t *testing.T) {
	ctx := context.Background()

	t.Run("successful activation", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		user := &domain.User{
			ID:       1,
			UserID:   "u1",
			Username: "Alice",
			IsActive: false,
			TeamID:   1,
		}

		mockRepo.On("GetByUserID", ctx, "u1").Return(user, nil)
		mockRepo.On("SetIsActive", ctx, "u1", true).Return(nil)

		result, err := service.SetIsActive(ctx, "u1", true)
		assert.NoError(t, err)
		assert.True(t, result.IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		mockRepo.On("GetByUserID", ctx, "non-existent").Return(nil, errors.New("not found"))

		_, err := service.SetIsActive(ctx, "non-existent", true)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound))
		mockRepo.AssertExpectations(t)
	})

	t.Run("deactivation", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		user := &domain.User{
			ID:       1,
			UserID:   "u1",
			Username: "Alice",
			IsActive: true,
			TeamID:   1,
		}

		mockRepo.On("GetByUserID", ctx, "u1").Return(user, nil)
		mockRepo.On("SetIsActive", ctx, "u1", false).Return(nil)

		result, err := service.SetIsActive(ctx, "u1", false)
		assert.NoError(t, err)
		assert.False(t, result.IsActive)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserReviewPRs(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get review PRs", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		user := &domain.User{
			ID:       1,
			UserID:   "u1",
			Username: "Alice",
			IsActive: true,
			TeamID:   1,
		}

		prs := []domain.PullRequest{
			{
				ID:              1,
				PullRequestID:   "pr-1",
				PullRequestName: "Test PR",
				AuthorID:        2,
				StatusID:        1,
				CreatedAt:       time.Now(),
			},
		}

		mockRepo.On("GetByUserID", ctx, "u1").Return(user, nil)
		mockRepo.On("GetByReviewerID", ctx, "u1").Return(prs, nil)

		result, err := service.GetUserReviewPRs(ctx, "u1")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "pr-1", result[0].PullRequestID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		mockRepo.On("GetByUserID", ctx, "non-existent").Return(nil, errors.New("not found"))

		_, err := service.GetUserReviewPRs(ctx, "non-existent")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound))
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty review PRs", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		user := &domain.User{
			ID:       1,
			UserID:   "u1",
			Username: "Alice",
			IsActive: true,
			TeamID:   1,
		}

		mockRepo.On("GetByUserID", ctx, "u1").Return(user, nil)
		mockRepo.On("GetByReviewerID", ctx, "u1").Return([]domain.PullRequest{}, nil)

		result, err := service.GetUserReviewPRs(ctx, "u1")
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockRepo.AssertExpectations(t)
	})
}

