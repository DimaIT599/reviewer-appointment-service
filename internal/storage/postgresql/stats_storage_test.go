package postgresql

import (
	"context"
	"reviewer-appointment-service/internal/models/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsRepo_GetTotalPRs(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	statsRepo := NewStatsRepo(storage)
	prRepo := NewPRRepo(storage)

	// Создаем команду и пользователя
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	userStorage := NewUserStorage(storage)
	author := &domain.User{UserID: "u1", Username: "Author", IsActive: true, TeamID: team.ID}
	err = userStorage.Create(ctx, author)
	require.NoError(t, err)

	t.Run("empty database", func(t *testing.T) {
		count, err := statsRepo.GetTotalPRs(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("with PRs", func(t *testing.T) {
		prs := []*domain.PullRequest{
			{PullRequestID: "pr-1", PullRequestName: "PR 1", AuthorID: author.ID, StatusID: 1},
			{PullRequestID: "pr-2", PullRequestName: "PR 2", AuthorID: author.ID, StatusID: 1},
			{PullRequestID: "pr-3", PullRequestName: "PR 3", AuthorID: author.ID, StatusID: 2},
		}

		for _, pr := range prs {
			err := prRepo.Create(ctx, pr)
			require.NoError(t, err)
		}

		count, err := statsRepo.GetTotalPRs(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})
}

func TestStatsRepo_GetTotalUsers(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	statsRepo := NewStatsRepo(storage)
	userStorage := NewUserStorage(storage)

	// Создаем команду
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	t.Run("empty database", func(t *testing.T) {
		count, err := statsRepo.GetTotalUsers(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("with users", func(t *testing.T) {
		users := []*domain.User{
			{UserID: "u1", Username: "Alice", IsActive: true, TeamID: team.ID},
			{UserID: "u2", Username: "Bob", IsActive: false, TeamID: team.ID},
			{UserID: "u3", Username: "Charlie", IsActive: true, TeamID: team.ID},
		}

		for _, user := range users {
			err := userStorage.Create(ctx, user)
			require.NoError(t, err)
		}

		count, err := statsRepo.GetTotalUsers(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})
}

func TestStatsRepo_GetActiveUsers(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	statsRepo := NewStatsRepo(storage)
	userStorage := NewUserStorage(storage)

	// Создаем команду
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	users := []*domain.User{
		{UserID: "u1", Username: "Alice", IsActive: true, TeamID: team.ID},
		{UserID: "u2", Username: "Bob", IsActive: false, TeamID: team.ID},
		{UserID: "u3", Username: "Charlie", IsActive: true, TeamID: team.ID},
	}

	for _, user := range users {
		err := userStorage.Create(ctx, user)
		require.NoError(t, err)
	}

	t.Run("get active users count", func(t *testing.T) {
		count, err := statsRepo.GetActiveUsers(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

func TestStatsRepo_GetPRsByStatus(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	statsRepo := NewStatsRepo(storage)
	prRepo := NewPRRepo(storage)

	// Создаем команду и пользователя
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	userStorage := NewUserStorage(storage)
	author := &domain.User{UserID: "u1", Username: "Author", IsActive: true, TeamID: team.ID}
	err = userStorage.Create(ctx, author)
	require.NoError(t, err)

	prs := []*domain.PullRequest{
		{PullRequestID: "pr-1", PullRequestName: "PR 1", AuthorID: author.ID, StatusID: 1}, // OPEN
		{PullRequestID: "pr-2", PullRequestName: "PR 2", AuthorID: author.ID, StatusID: 1}, // OPEN
		{PullRequestID: "pr-3", PullRequestName: "PR 3", AuthorID: author.ID, StatusID: 2}, // MERGED
	}

	for _, pr := range prs {
		err := prRepo.Create(ctx, pr)
		require.NoError(t, err)
	}

	t.Run("get PRs by status", func(t *testing.T) {
		stats, err := statsRepo.GetPRsByStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, stats["OPEN"])
		assert.Equal(t, 1, stats["MERGED"])
	})
}

func TestStatsRepo_GetTopReviewers(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	statsRepo := NewStatsRepo(storage)
	prRepo := NewPRRepo(storage)

	// Создаем команду и пользователей
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	userStorage := NewUserStorage(storage)
	author := &domain.User{UserID: "u1", Username: "Author", IsActive: true, TeamID: team.ID}
	reviewer1 := &domain.User{UserID: "u2", Username: "Reviewer1", IsActive: true, TeamID: team.ID}
	reviewer2 := &domain.User{UserID: "u3", Username: "Reviewer2", IsActive: true, TeamID: team.ID}

	err = userStorage.Create(ctx, author)
	require.NoError(t, err)
	err = userStorage.Create(ctx, reviewer1)
	require.NoError(t, err)
	err = userStorage.Create(ctx, reviewer2)
	require.NoError(t, err)

	// Создаем PRs и назначаем ревьюверов
	pr1 := &domain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "PR 1",
		AuthorID:        author.ID,
		StatusID:        1,
	}
	err = prRepo.Create(ctx, pr1)
	require.NoError(t, err)
	err = prRepo.AddReviewer(ctx, pr1.ID, reviewer1.ID)
	require.NoError(t, err)
	err = prRepo.AddReviewer(ctx, pr1.ID, reviewer2.ID)
	require.NoError(t, err)

	pr2 := &domain.PullRequest{
		PullRequestID:   "pr-2",
		PullRequestName: "PR 2",
		AuthorID:        author.ID,
		StatusID:        1,
	}
	err = prRepo.Create(ctx, pr2)
	require.NoError(t, err)
	err = prRepo.AddReviewer(ctx, pr2.ID, reviewer1.ID)
	require.NoError(t, err)

	t.Run("get top reviewers", func(t *testing.T) {
		stats, err := statsRepo.GetTopReviewers(ctx, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(stats), 2)

		// Reviewer1 должен быть первым (2 назначения)
		assert.Equal(t, "u2", stats[0].UserID)
		assert.Equal(t, 2, stats[0].ReviewCount)
	})
}

