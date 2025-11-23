package postgresql

import (
	"context"
	"reviewer-appointment-service/internal/models/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamRepo_Create(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	teamRepo := NewTeamRepo(storage)

	t.Run("successful creation", func(t *testing.T) {
		team := &domain.Team{Name: "backend"}
		err := teamRepo.Create(ctx, team)
		require.NoError(t, err)
		assert.NotZero(t, team.ID)
		assert.False(t, team.CreatedAt.IsZero())
	})

	t.Run("duplicate team name", func(t *testing.T) {
		team := &domain.Team{Name: "backend"}
		err := teamRepo.Create(ctx, team)
		assert.Error(t, err)
	})
}

func TestTeamRepo_GetByName(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	teamRepo := NewTeamRepo(storage)

	// Создаем команду
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	t.Run("get existing team", func(t *testing.T) {
		found, err := teamRepo.GetByName(ctx, "backend")
		require.NoError(t, err)
		assert.Equal(t, "backend", found.Name)
		assert.Equal(t, team.ID, found.ID)
	})

	t.Run("get non-existing team", func(t *testing.T) {
		_, err := teamRepo.GetByName(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestTeamRepo_GetByID(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	teamRepo := NewTeamRepo(storage)

	// Создаем команду
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	t.Run("get existing team by ID", func(t *testing.T) {
		found, err := teamRepo.GetByID(ctx, team.ID)
		require.NoError(t, err)
		assert.Equal(t, "backend", found.Name)
		assert.Equal(t, team.ID, found.ID)
	})

	t.Run("get non-existing team by ID", func(t *testing.T) {
		_, err := teamRepo.GetByID(ctx, 99999)
		assert.Error(t, err)
	})
}

func TestTeamRepo_GetWithUsers(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	teamRepo := NewTeamRepo(storage)
	userStorage := NewUserStorage(storage)

	// Создаем команду
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	// Создаем пользователей
	users := []*domain.User{
		{UserID: "u1", Username: "Alice", IsActive: true, TeamID: team.ID},
		{UserID: "u2", Username: "Bob", IsActive: false, TeamID: team.ID},
		{UserID: "u3", Username: "Charlie", IsActive: true, TeamID: team.ID},
	}

	for _, user := range users {
		err := userStorage.Create(ctx, user)
		require.NoError(t, err)
	}

	t.Run("get team with users", func(t *testing.T) {
		found, err := teamRepo.GetWithUsers(ctx, team.ID)
		require.NoError(t, err)
		assert.Equal(t, "backend", found.Name)
		assert.Len(t, found.Users, 3) // Все пользователи, включая неактивных
	})
}

func TestTeamRepo_GetAllWithUsers(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	teamRepo := NewTeamRepo(storage)
	userStorage := NewUserStorage(storage)

	// Создаем несколько команд
	teams := []*domain.Team{
		{Name: "backend"},
		{Name: "frontend"},
	}

	for _, team := range teams {
		err := teamRepo.Create(ctx, team)
		require.NoError(t, err)

		// Добавляем пользователей в каждую команду
		user := &domain.User{
			UserID:   "u-" + team.Name,
			Username: "User " + team.Name,
			IsActive: true,
			TeamID:   team.ID,
		}
		err = userStorage.Create(ctx, user)
		require.NoError(t, err)
	}

	t.Run("get all teams with users", func(t *testing.T) {
		found, err := teamRepo.GetAllWithUsers(ctx)
		require.NoError(t, err)
		assert.Len(t, found, 2)
		for _, team := range found {
			assert.Greater(t, len(team.Users), 0)
		}
	})
}

func TestTeamRepo_ExistsByName(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	teamRepo := NewTeamRepo(storage)

	// Создаем команду
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	t.Run("team exists", func(t *testing.T) {
		exists, err := teamRepo.ExistsByName(ctx, "backend")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("team does not exist", func(t *testing.T) {
		exists, err := teamRepo.ExistsByName(ctx, "non-existent")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

