package postgresql

import (
	"context"
	"reviewer-appointment-service/internal/models/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStorage_Create(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	userStorage := NewUserStorage(storage)

	// Создаем команду для пользователя
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "test-team"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		user := &domain.User{
			UserID:   "u1",
			Username:  "Alice",
			IsActive: true,
			TeamID:   team.ID,
		}

		err := userStorage.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.False(t, user.CreatedAt.IsZero())
	})

	t.Run("duplicate user_id", func(t *testing.T) {
		user := &domain.User{
			UserID:   "u1",
			Username:  "Bob",
			IsActive: true,
			TeamID:   team.ID,
		}

		err := userStorage.Create(ctx, user)
		assert.Error(t, err)
	})
}

func TestUserStorage_GetByUserID(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	userStorage := NewUserStorage(storage)

	// Создаем команду и пользователя
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "test-team"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	user := &domain.User{
		UserID:   "u1",
		Username: "Alice",
		IsActive: true,
		TeamID:   team.ID,
	}
	err = userStorage.Create(ctx, user)
	require.NoError(t, err)

	t.Run("get existing user", func(t *testing.T) {
		found, err := userStorage.GetByUserID(ctx, "u1")
		require.NoError(t, err)
		assert.Equal(t, "u1", found.UserID)
		assert.Equal(t, "Alice", found.Username)
		assert.True(t, found.IsActive)
		assert.Equal(t, team.ID, found.TeamID)
	})

	t.Run("get non-existing user", func(t *testing.T) {
		_, err := userStorage.GetByUserID(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestUserStorage_GetByTeamID(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	userStorage := NewUserStorage(storage)

	// Создаем команду
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "test-team"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	// Создаем несколько пользователей
	users := []*domain.User{
		{UserID: "u1", Username: "Alice", IsActive: true, TeamID: team.ID},
		{UserID: "u2", Username: "Bob", IsActive: false, TeamID: team.ID},
		{UserID: "u3", Username: "Charlie", IsActive: true, TeamID: team.ID},
	}

	for _, user := range users {
		err := userStorage.Create(ctx, user)
		require.NoError(t, err)
	}

	t.Run("get all users from team", func(t *testing.T) {
		found, err := userStorage.GetByTeamID(ctx, team.ID)
		require.NoError(t, err)
		assert.Len(t, found, 3) // Все пользователи, включая неактивных
	})
}

func TestUserStorage_SetIsActive(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	userStorage := NewUserStorage(storage)

	// Создаем команду и пользователя
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "test-team"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	user := &domain.User{
		UserID:   "u1",
		Username: "Alice",
		IsActive: true,
		TeamID:   team.ID,
	}
	err = userStorage.Create(ctx, user)
	require.NoError(t, err)

	t.Run("deactivate user", func(t *testing.T) {
		err := userStorage.SetIsActive(ctx, "u1", false)
		require.NoError(t, err)

		found, err := userStorage.GetByUserID(ctx, "u1")
		require.NoError(t, err)
		assert.False(t, found.IsActive)
	})

	t.Run("activate user", func(t *testing.T) {
		err := userStorage.SetIsActive(ctx, "u1", true)
		require.NoError(t, err)

		found, err := userStorage.GetByUserID(ctx, "u1")
		require.NoError(t, err)
		assert.True(t, found.IsActive)
	})

	t.Run("non-existing user", func(t *testing.T) {
		err := userStorage.SetIsActive(ctx, "non-existent", false)
		assert.Error(t, err)
	})
}

func TestUserStorage_Update(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	userStorage := NewUserStorage(storage)

	// Создаем команду и пользователя
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "test-team"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	user := &domain.User{
		UserID:   "u1",
		Username: "Alice",
		IsActive: true,
		TeamID:   team.ID,
	}
	err = userStorage.Create(ctx, user)
	require.NoError(t, err)

	t.Run("update user", func(t *testing.T) {
		user.Username = "Alice Updated"
		user.IsActive = false
		err := userStorage.Update(ctx, user)
		require.NoError(t, err)

		found, err := userStorage.GetByUserID(ctx, "u1")
		require.NoError(t, err)
		assert.Equal(t, "Alice Updated", found.Username)
		assert.False(t, found.IsActive)
	})
}

func TestUserStorage_DeactivateByTeamID(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	userStorage := NewUserStorage(storage)

	// Создаем команду и пользователей
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "test-team"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	users := []*domain.User{
		{UserID: "u1", Username: "Alice", IsActive: true, TeamID: team.ID},
		{UserID: "u2", Username: "Bob", IsActive: true, TeamID: team.ID},
	}

	for _, user := range users {
		err := userStorage.Create(ctx, user)
		require.NoError(t, err)
	}

	t.Run("deactivate all users in team", func(t *testing.T) {
		err := userStorage.DeactivateByTeamID(ctx, team.ID)
		require.NoError(t, err)

		found, err := userStorage.GetByTeamID(ctx, team.ID)
		require.NoError(t, err)
		for _, user := range found {
			assert.False(t, user.IsActive)
		}
	})
}

func TestUserStorage_GetByReviewerID(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	userStorage := NewUserStorage(storage)
	prRepo := NewPRRepo(storage)

	// Создаем команду и пользователей
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "test-team"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	author := &domain.User{UserID: "u1", Username: "Author", IsActive: true, TeamID: team.ID}
	reviewer := &domain.User{UserID: "u2", Username: "Reviewer", IsActive: true, TeamID: team.ID}
	err = userStorage.Create(ctx, author)
	require.NoError(t, err)
	err = userStorage.Create(ctx, reviewer)
	require.NoError(t, err)

	// Создаем PR
	pr := &domain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        author.ID,
		StatusID:        1, // OPEN
	}
	err = prRepo.Create(ctx, pr)
	require.NoError(t, err)

	// Назначаем ревьювера
	err = prRepo.AddReviewer(ctx, pr.ID, reviewer.ID)
	require.NoError(t, err)

	t.Run("get PRs by reviewer", func(t *testing.T) {
		prs, err := userStorage.GetByReviewerID(ctx, "u2")
		require.NoError(t, err)
		assert.Len(t, prs, 1)
		assert.Equal(t, "pr-1", prs[0].PullRequestID)
	})
}

