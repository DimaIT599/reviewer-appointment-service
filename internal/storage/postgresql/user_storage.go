package postgresql

import (
	"context"
	"fmt"
	"reviewer-appointment-service/internal/models/domain"
)

type UserStorage struct {
	storage *Storage
}

func NewUserStorage(storage *Storage) *UserStorage {
	return &UserStorage{storage: storage}
}

func (r *UserStorage) Create(ctx context.Context, user *domain.User) error {
	const op = "storage.postgresql.UserStorage.Create"

	query := `
		INSERT INTO pr_system.users (user_id, username, team_id, is_active) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at`

	err := r.storage.DB.QueryRow(
		ctx, query, user.UserID, user.Username, user.TeamID, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *UserStorage) Update(ctx context.Context, user *domain.User) error {
	const op = "storage.postgresql.UserStorage.Update"

	query := `
		UPDATE pr_system.users 
		SET username = $1, team_id = $2, is_active = $3 
		WHERE user_id = $4 
		RETURNING id, created_at`

	err := r.storage.DB.QueryRow(
		ctx, query, user.Username, user.TeamID, user.IsActive, user.UserID,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *UserStorage) GetByUserID(ctx context.Context, userID string) (*domain.User, error) {
	const op = "storage.postgresql.UserStorage.GetByUserID"

	query := `
		SELECT id, user_id, username, is_active, team_id, created_at 
		FROM pr_system.users 
		WHERE user_id = $1`

	var user domain.User
	err := r.storage.DB.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.UserID, &user.Username, &user.IsActive, &user.TeamID, &user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

func (r *UserStorage) GetByTeamID(ctx context.Context, teamID int64) ([]domain.User, error) {
	const op = "storage.postgresql.UserStorage.GetByTeamID"

	query := `
	SELECT id, user_id, username, is_active, team_id, created_at 
	FROM pr_system.users 
	WHERE team_id = $1`

	rows, err := r.storage.DB.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID, &user.UserID, &user.Username,
			&user.IsActive, &user.TeamID, &user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserStorage) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	const op = "storage.postgresql.UserStorage.SetIsActive"

	query := `
		UPDATE pr_system.users 
		SET is_active = $1 
		WHERE user_id = $2`

	result, err := r.storage.DB.Exec(ctx, query, isActive, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: user not found", op)
	}

	return nil
}

func (r *UserStorage) GetByReviewerID(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	const op = "storage.postgresql.UserStorage.GetByReviewerID"

	query := `
		SELECT 
			pr.id, pr.pull_request_id, pr.pull_request_name, 
			pr.author_id, pr.status_id, pr.merged_at, pr.created_at,
			u.id, u.user_id, u.username, u.is_active, u.team_id, u.created_at
		FROM pr_system.pull_requests pr
		JOIN pr_system.pr_reviewers prr ON pr.id = prr.pr_id
		JOIN pr_system.users u ON pr.author_id = u.id
		JOIN pr_system.users reviewer ON prr.reviewer_id = reviewer.id
		WHERE reviewer.user_id = $1`

	rows, err := r.storage.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		var author domain.User
		err := rows.Scan(
			&pr.ID, &pr.PullRequestID, &pr.PullRequestName,
			&pr.AuthorID, &pr.StatusID, &pr.MergedAt, &pr.CreatedAt,
			&author.ID, &author.UserID, &author.Username, &author.IsActive, &author.TeamID, &author.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		pr.Author = &author
		prs = append(prs, pr)
	}

	return prs, nil
}

func (r *UserStorage) DeactivateByTeamID(ctx context.Context, teamID int64) error {
	const op = "storage.postgresql.UserStorage.DeactivateByTeamID"

	query := `
		UPDATE pr_system.users 
		SET is_active = false 
		WHERE team_id = $1`

	_, err := r.storage.DB.Exec(ctx, query, teamID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
