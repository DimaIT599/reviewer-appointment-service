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

// MockTeamRepository - мок для TeamRepository
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(ctx context.Context, team *domain.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockTeamRepository) GetByID(ctx context.Context, teamID int64) (*domain.Team, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockTeamRepository) GetWithUsers(ctx context.Context, teamID int64) (*domain.Team, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockTeamRepository) GetAllWithUsers(ctx context.Context) ([]domain.Team, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Team), args.Error(1)
}

func (m *MockTeamRepository) ExistsByName(ctx context.Context, teamName string) (bool, error) {
	args := m.Called(ctx, teamName)
	return args.Bool(0), args.Error(1)
}

func TestTeamService_CreateTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation with new users", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo)

		team := &domain.Team{
			Name: "backend",
			Users: []domain.User{
				{UserID: "u1", Username: "Alice", IsActive: true},
				{UserID: "u2", Username: "Bob", IsActive: true},
			},
		}

		createdTeam := &domain.Team{
			ID:        1,
			Name:      "backend",
			CreatedAt: time.Now(),
			Users: []domain.User{
				{ID: 1, UserID: "u1", Username: "Alice", IsActive: true, TeamID: 1},
				{ID: 2, UserID: "u2", Username: "Bob", IsActive: true, TeamID: 1},
			},
		}

		mockTeamRepo.On("ExistsByName", ctx, "backend").Return(false, nil)
		mockTeamRepo.On("Create", ctx, mock.AnythingOfType("*domain.Team")).Run(func(args mock.Arguments) {
			team := args.Get(1).(*domain.Team)
			team.ID = 1
			team.CreatedAt = time.Now()
		}).Return(nil)
		mockUserRepo.On("GetByUserID", ctx, "u1").Return(nil, errors.New("not found"))
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
		mockUserRepo.On("GetByUserID", ctx, "u2").Return(nil, errors.New("not found"))
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
		mockTeamRepo.On("GetWithUsers", ctx, int64(1)).Return(createdTeam, nil)

		result, err := service.CreateTeam(ctx, team)
		assert.NoError(t, err)
		assert.Equal(t, "backend", result.Name)
		assert.Len(t, result.Users, 2)
		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("team already exists", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo)

		team := &domain.Team{Name: "backend"}

		mockTeamRepo.On("ExistsByName", ctx, "backend").Return(true, nil)

		_, err := service.CreateTeam(ctx, team)
		assert.Error(t, err)
		assert.Equal(t, storage.ErrTeamExists, err)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("update existing users", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo)

		team := &domain.Team{
			Name: "backend",
			Users: []domain.User{
				{UserID: "u1", Username: "Alice Updated", IsActive: false},
			},
		}

		existingUser := &domain.User{
			ID:        1,
			UserID:    "u1",
			Username:  "Alice",
			IsActive:  true,
			TeamID:    0,
			CreatedAt: time.Now(),
		}

		createdTeam := &domain.Team{
			ID:        1,
			Name:      "backend",
			CreatedAt: time.Now(),
			Users: []domain.User{
				{ID: 1, UserID: "u1", Username: "Alice Updated", IsActive: false, TeamID: 1},
			},
		}

		mockTeamRepo.On("ExistsByName", ctx, "backend").Return(false, nil)
		mockTeamRepo.On("Create", ctx, mock.AnythingOfType("*domain.Team")).Run(func(args mock.Arguments) {
			team := args.Get(1).(*domain.Team)
			team.ID = 1
			team.CreatedAt = time.Now()
		}).Return(nil)
		mockUserRepo.On("GetByUserID", ctx, "u1").Return(existingUser, nil)
		mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
		mockTeamRepo.On("GetWithUsers", ctx, int64(1)).Return(createdTeam, nil)

		result, err := service.CreateTeam(ctx, team)
		assert.NoError(t, err)
		assert.Equal(t, "backend", result.Name)
		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestTeamService_GetTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get team", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo)

		team := &domain.Team{
			ID:   1,
			Name: "backend",
		}

		teamWithUsers := &domain.Team{
			ID:   1,
			Name: "backend",
			Users: []domain.User{
				{ID: 1, UserID: "u1", Username: "Alice", IsActive: true, TeamID: 1},
			},
		}

		mockTeamRepo.On("GetByName", ctx, "backend").Return(team, nil)
		mockTeamRepo.On("GetWithUsers", ctx, int64(1)).Return(teamWithUsers, nil)

		result, err := service.GetTeam(ctx, "backend")
		assert.NoError(t, err)
		assert.Equal(t, "backend", result.Name)
		assert.Len(t, result.Users, 1)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("team not found", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo)

		mockTeamRepo.On("GetByName", ctx, "non-existent").Return(nil, errors.New("not found"))

		_, err := service.GetTeam(ctx, "non-existent")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound))
		mockTeamRepo.AssertExpectations(t)
	})
}

func TestTeamService_DeactivateTeamUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deactivation", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo)

		team := &domain.Team{
			ID:   1,
			Name: "backend",
		}

		mockTeamRepo.On("GetByID", ctx, int64(1)).Return(team, nil)
		mockUserRepo.On("DeactivateByTeamID", ctx, int64(1)).Return(nil)

		err := service.DeactivateTeamUsers(ctx, 1)
		assert.NoError(t, err)
		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("team not found", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo)

		mockTeamRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		err := service.DeactivateTeamUsers(ctx, 999)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound))
		mockTeamRepo.AssertExpectations(t)
	})
}

