package postgresql

import (
	"context"
	"reviewer-appointment-service/internal/models/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRRepo_Create(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
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

	t.Run("successful creation", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        author.ID,
			StatusID:        1, // OPEN
		}

		err := prRepo.Create(ctx, pr)
		require.NoError(t, err)
		assert.NotZero(t, pr.ID)
		assert.False(t, pr.CreatedAt.IsZero())
	})

	t.Run("duplicate PR ID", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr-1",
			PullRequestName: "Another PR",
			AuthorID:        author.ID,
			StatusID:        1,
		}

		err := prRepo.Create(ctx, pr)
		assert.Error(t, err)
	})
}

func TestPRRepo_GetByPRID(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
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

	// Создаем PR
	pr := &domain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        author.ID,
		StatusID:        1,
	}
	err = prRepo.Create(ctx, pr)
	require.NoError(t, err)

	// Назначаем ревьюверов
	err = prRepo.AddReviewer(ctx, pr.ID, reviewer1.ID)
	require.NoError(t, err)
	err = prRepo.AddReviewer(ctx, pr.ID, reviewer2.ID)
	require.NoError(t, err)

	t.Run("get PR with reviewers", func(t *testing.T) {
		found, err := prRepo.GetByPRID(ctx, "pr-1")
		require.NoError(t, err)
		assert.Equal(t, "pr-1", found.PullRequestID)
		assert.Len(t, found.Reviewers, 2)
	})

	t.Run("get non-existing PR", func(t *testing.T) {
		_, err := prRepo.GetByPRID(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestPRRepo_Update(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
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

	pr := &domain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        author.ID,
		StatusID:        1,
	}
	err = prRepo.Create(ctx, pr)
	require.NoError(t, err)

	t.Run("update PR status to merged", func(t *testing.T) {
		now := time.Now()
		pr.StatusID = 2 // MERGED
		pr.MergedAt = &now
		pr.PullRequestName = "Updated PR"

		err := prRepo.Update(ctx, pr)
		require.NoError(t, err)

		found, err := prRepo.GetByPRID(ctx, "pr-1")
		require.NoError(t, err)
		assert.Equal(t, 2, found.StatusID)
		assert.NotNil(t, found.MergedAt)
		assert.Equal(t, "Updated PR", found.PullRequestName)
	})
}

func TestPRRepo_AddReviewer(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	prRepo := NewPRRepo(storage)

	// Создаем команду и пользователей
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	userStorage := NewUserStorage(storage)
	author := &domain.User{UserID: "u1", Username: "Author", IsActive: true, TeamID: team.ID}
	reviewer := &domain.User{UserID: "u2", Username: "Reviewer", IsActive: true, TeamID: team.ID}

	err = userStorage.Create(ctx, author)
	require.NoError(t, err)
	err = userStorage.Create(ctx, reviewer)
	require.NoError(t, err)

	pr := &domain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        author.ID,
		StatusID:        1,
	}
	err = prRepo.Create(ctx, pr)
	require.NoError(t, err)

	t.Run("add reviewer", func(t *testing.T) {
		err := prRepo.AddReviewer(ctx, pr.ID, reviewer.ID)
		require.NoError(t, err)

		reviewers, err := prRepo.GetReviewers(ctx, pr.ID)
		require.NoError(t, err)
		assert.Len(t, reviewers, 1)
		assert.Equal(t, "u2", reviewers[0].UserID)
	})

	t.Run("add duplicate reviewer", func(t *testing.T) {
		err := prRepo.AddReviewer(ctx, pr.ID, reviewer.ID)
		require.NoError(t, err) // ON CONFLICT DO NOTHING, не должно быть ошибки
	})
}

func TestPRRepo_RemoveReviewer(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	prRepo := NewPRRepo(storage)

	// Создаем команду и пользователей
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	userStorage := NewUserStorage(storage)
	author := &domain.User{UserID: "u1", Username: "Author", IsActive: true, TeamID: team.ID}
	reviewer := &domain.User{UserID: "u2", Username: "Reviewer", IsActive: true, TeamID: team.ID}

	err = userStorage.Create(ctx, author)
	require.NoError(t, err)
	err = userStorage.Create(ctx, reviewer)
	require.NoError(t, err)

	pr := &domain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        author.ID,
		StatusID:        1,
	}
	err = prRepo.Create(ctx, pr)
	require.NoError(t, err)

	err = prRepo.AddReviewer(ctx, pr.ID, reviewer.ID)
	require.NoError(t, err)

	t.Run("remove reviewer", func(t *testing.T) {
		err := prRepo.RemoveReviewer(ctx, pr.ID, reviewer.ID)
		require.NoError(t, err)

		reviewers, err := prRepo.GetReviewers(ctx, pr.ID)
		require.NoError(t, err)
		assert.Len(t, reviewers, 0)
	})

	t.Run("remove non-existing reviewer", func(t *testing.T) {
		err := prRepo.RemoveReviewer(ctx, pr.ID, reviewer.ID)
		assert.Error(t, err)
	})
}

func TestPRRepo_GetReviewers(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
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

	pr := &domain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        author.ID,
		StatusID:        1,
	}
	err = prRepo.Create(ctx, pr)
	require.NoError(t, err)

	err = prRepo.AddReviewer(ctx, pr.ID, reviewer1.ID)
	require.NoError(t, err)
	err = prRepo.AddReviewer(ctx, pr.ID, reviewer2.ID)
	require.NoError(t, err)

	t.Run("get reviewers", func(t *testing.T) {
		reviewers, err := prRepo.GetReviewers(ctx, pr.ID)
		require.NoError(t, err)
		assert.Len(t, reviewers, 2)
	})

	t.Run("get reviewers for PR without reviewers", func(t *testing.T) {
		pr2 := &domain.PullRequest{
			PullRequestID:   "pr-2",
			PullRequestName: "Test PR 2",
			AuthorID:        author.ID,
			StatusID:        1,
		}
		err = prRepo.Create(ctx, pr2)
		require.NoError(t, err)

		reviewers, err := prRepo.GetReviewers(ctx, pr2.ID)
		require.NoError(t, err)
		assert.Len(t, reviewers, 0)
	})
}

func TestPRRepo_GetByReviewerID(t *testing.T) {
	storage, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()
	prRepo := NewPRRepo(storage)

	// Создаем команду и пользователей
	teamRepo := NewTeamRepo(storage)
	team := &domain.Team{Name: "backend"}
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)

	userStorage := NewUserStorage(storage)
	author := &domain.User{UserID: "u1", Username: "Author", IsActive: true, TeamID: team.ID}
	reviewer := &domain.User{UserID: "u2", Username: "Reviewer", IsActive: true, TeamID: team.ID}

	err = userStorage.Create(ctx, author)
	require.NoError(t, err)
	err = userStorage.Create(ctx, reviewer)
	require.NoError(t, err)

	pr := &domain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        author.ID,
		StatusID:        1,
	}
	err = prRepo.Create(ctx, pr)
	require.NoError(t, err)

	err = prRepo.AddReviewer(ctx, pr.ID, reviewer.ID)
	require.NoError(t, err)

	t.Run("get PRs by reviewer user_id", func(t *testing.T) {
		prs, err := prRepo.GetByReviewerID(ctx, "u2")
		require.NoError(t, err)
		assert.Len(t, prs, 1)
		assert.Equal(t, "pr-1", prs[0].PullRequestID)
	})
}

