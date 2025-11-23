package postgresql

import (
	"context"
	"fmt"
	"reviewer-appointment-service/internal/models/domain"
)

type StatsRepo struct {
	storage *Storage
}

func NewStatsRepo(storage *Storage) *StatsRepo {
	return &StatsRepo{storage: storage}
}

func (r *StatsRepo) GetTotalPRs(ctx context.Context) (int, error) {
	const op = "repository.StatsRepo.GetTotalPRs"
	const query = `SELECT COUNT(*) FROM pr_system.pull_requests`

	var count int
	err := r.storage.DB.QueryRow(ctx, query).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return count, nil
}

func (r *StatsRepo) GetTotalUsers(ctx context.Context) (int, error) {
	const op = "repository.StatsRepo.GetTotalUsers"
	const query = `SELECT COUNT(*) FROM pr_system.users`

	var count int
	err := r.storage.DB.QueryRow(ctx, query).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return count, nil
}

func (r *StatsRepo) GetActiveUsers(ctx context.Context) (int, error) {
	const op = "repository.StatsRepo.GetActiveUsers"
	const query = `SELECT COUNT(*) FROM pr_system.users WHERE is_active = true`

	var count int
	err := r.storage.DB.QueryRow(ctx, query).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return count, nil
}

func (r *StatsRepo) GetPRsByStatus(ctx context.Context) (map[string]int, error) {
	const op = "repository.StatsRepo.GetPRsByStatus"
	const query = `
        SELECT s.name, COUNT(pr.id) 
        FROM pr_system.statuses s 
        LEFT JOIN pr_system.pull_requests pr ON s.id = pr.status_id 
        GROUP BY s.id, s.name`

	rows, err := r.storage.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var statusName string
		var count int
		err := rows.Scan(&statusName, &count)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		result[statusName] = count
	}

	return result, nil
}

func (r *StatsRepo) GetTopReviewers(ctx context.Context, limit int) ([]domain.ReviewerStats, error) {
	const op = "repository.StatsRepo.GetTopReviewers"
	const query = `
        SELECT 
            u.user_id, 
            u.username, 
            COUNT(prr.id) as review_count
        FROM pr_system.users u
        JOIN pr_system.pr_reviewers prr ON u.id = prr.reviewer_id
        WHERE u.is_active = true
        GROUP BY u.id, u.user_id, u.username
        ORDER BY review_count DESC
        LIMIT $1`

	rows, err := r.storage.DB.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var stats []domain.ReviewerStats
	for rows.Next() {
		var stat domain.ReviewerStats
		err := rows.Scan(&stat.UserID, &stat.Username, &stat.ReviewCount)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}
