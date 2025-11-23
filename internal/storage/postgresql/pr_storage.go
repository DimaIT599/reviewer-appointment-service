package postgresql

import (
	"context"
	"fmt"
	"reviewer-appointment-service/internal/models/domain"
	"strings"
)

type PRRepo struct {
	storage *Storage
}

func NewPRRepo(storage *Storage) *PRRepo {
	return &PRRepo{storage: storage}
}

func (r *PRRepo) Create(ctx context.Context, pr *domain.PullRequest) error {
	const op = "repository.PRRepo.Create"
	const query = `
        INSERT INTO pr_system.pull_requests (pull_request_id, pull_request_name, author_id, status_id) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id, created_at`

	err := r.storage.DB.QueryRow(
		ctx, query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.StatusID,
	).Scan(&pr.ID, &pr.CreatedAt)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *PRRepo) Update(ctx context.Context, pr *domain.PullRequest) error {
	const op = "repository.PRRepo.Update"
	const query = `
        UPDATE pr_system.pull_requests 
        SET pull_request_name = $1, status_id = $2, merged_at = $3 
        WHERE pull_request_id = $4 
        RETURNING id, author_id, created_at`

	err := r.storage.DB.QueryRow(
		ctx, query, pr.PullRequestName, pr.StatusID, pr.MergedAt, pr.PullRequestID,
	).Scan(&pr.ID, &pr.AuthorID, &pr.CreatedAt)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *PRRepo) GetByPRID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	const op = "repository.PRRepo.GetByPRID"
	const query = `
        SELECT id, pull_request_id, pull_request_name, author_id, status_id, merged_at, created_at 
        FROM pr_system.pull_requests 
        WHERE pull_request_id = $1`

	var pr domain.PullRequest
	err := r.storage.DB.QueryRow(ctx, query, prID).Scan(
		&pr.ID, &pr.PullRequestID, &pr.PullRequestName,
		&pr.AuthorID, &pr.StatusID, &pr.MergedAt, &pr.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	reviewers, err := r.GetReviewers(ctx, pr.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get reviewers: %w", op, err)
	}
	pr.Reviewers = reviewers

	return &pr, nil
}

func (r *PRRepo) GetByReviewerID(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	const op = "repository.PRRepo.GetByReviewerID"
	const query = `
        SELECT 
            pr.id, pr.pull_request_id, pr.pull_request_name, 
            pr.author_id, pr.status_id, pr.merged_at, pr.created_at
        FROM pr_system.pull_requests pr
        JOIN pr_system.pr_reviewers prr ON pr.id = prr.pr_id
        JOIN pr_system.users u ON prr.reviewer_id = u.id
        WHERE u.user_id = $1`

	rows, err := r.storage.DB.Query(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		err := rows.Scan(
			&pr.ID, &pr.PullRequestID, &pr.PullRequestName,
			&pr.AuthorID, &pr.StatusID, &pr.MergedAt, &pr.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

func (r *PRRepo) GetOpenPRsByUserIDs(ctx context.Context, userIDs []string) ([]domain.PullRequest, error) {
	const op = "repository.PRRepo.GetOpenPRsByUserIDs"

	if len(userIDs) == 0 {
		return []domain.PullRequest{}, nil
	}

	// Создаем плейсхолдеры для IN clause
	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
        SELECT 
            pr.id, pr.pull_request_id, pr.pull_request_name, 
            pr.author_id, pr.status_id, pr.merged_at, pr.created_at,
            u.id, u.user_id, u.username, u.is_active, u.team_id, u.created_at
        FROM pr_system.pull_requests pr
        JOIN pr_system.users u ON pr.author_id = u.id
        WHERE u.user_id IN (%s) AND pr.status_id = 1`, // status_id = 1 для открытых PR
		strings.Join(placeholders, ","))

	rows, err := r.storage.DB.Query(ctx, query, args...)
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

func (r *PRRepo) AddReviewer(ctx context.Context, prID int64, reviewerID int64) error {
	const op = "repository.PRRepo.AddReviewer"
	const query = `
        INSERT INTO pr_system.pr_reviewers (pr_id, reviewer_id) 
        VALUES ($1, $2) 
        ON CONFLICT (pr_id, reviewer_id) DO NOTHING`

	_, err := r.storage.DB.Exec(ctx, query, prID, reviewerID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *PRRepo) RemoveReviewer(ctx context.Context, prID int64, reviewerID int64) error {
	const op = "repository.PRRepo.RemoveReviewer"
	const query = `
        DELETE FROM pr_system.pr_reviewers 
        WHERE pr_id = $1 AND reviewer_id = $2`

	result, err := r.storage.DB.Exec(ctx, query, prID, reviewerID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: reviewer not found for this PR", op)
	}

	return nil
}

func (r *PRRepo) GetReviewers(ctx context.Context, prID int64) ([]domain.User, error) {
	const op = "repository.PRRepo.GetReviewers"
	const query = `
        SELECT u.id, u.user_id, u.username, u.is_active, u.team_id, u.created_at
        FROM pr_system.pr_reviewers prr
        JOIN pr_system.users u ON prr.reviewer_id = u.id
        WHERE prr.pr_id = $1`

	rows, err := r.storage.DB.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var reviewers []domain.User
	for rows.Next() {
		var reviewer domain.User
		err := rows.Scan(
			&reviewer.ID, &reviewer.UserID, &reviewer.Username,
			&reviewer.IsActive, &reviewer.TeamID, &reviewer.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		reviewers = append(reviewers, reviewer)
	}

	return reviewers, nil
}
