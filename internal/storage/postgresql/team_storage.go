package postgresql

import (
	"context"
	"fmt"
	"reviewer-appointment-service/internal/models/domain"
)

type TeamRepo struct {
	storage *Storage
}

func NewTeamRepo(storage *Storage) *TeamRepo {
	return &TeamRepo{storage: storage}
}

func (r *TeamRepo) Create(ctx context.Context, team *domain.Team) error {
	const op = "repository.TeamRepo.Create"
	const query = `
        INSERT INTO pr_system.teams (name) 
        VALUES ($1) 
        RETURNING id, created_at`

	err := r.storage.DB.QueryRow(ctx, query, team.Name).Scan(&team.ID, &team.CreatedAt)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *TeamRepo) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	const op = "repository.TeamRepo.GetByName"
	const query = `
        SELECT id, name, created_at 
        FROM pr_system.teams 
        WHERE name = $1`

	var team domain.Team
	err := r.storage.DB.QueryRow(ctx, query, teamName).Scan(
		&team.ID, &team.Name, &team.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &team, nil
}

func (r *TeamRepo) GetByID(ctx context.Context, teamID int64) (*domain.Team, error) {
	const op = "repository.TeamRepo.GetByID"
	const query = `
        SELECT id, name, created_at 
        FROM pr_system.teams 
        WHERE id = $1`

	var team domain.Team
	err := r.storage.DB.QueryRow(ctx, query, teamID).Scan(
		&team.ID, &team.Name, &team.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &team, nil
}

func (r *TeamRepo) GetWithUsers(ctx context.Context, teamID int64) (*domain.Team, error) {
	const op = "repository.TeamRepo.GetWithUsers"

	teamQuery := `SELECT id, name, created_at FROM pr_system.teams WHERE id = $1`
	var team domain.Team
	err := r.storage.DB.QueryRow(ctx, teamQuery, teamID).Scan(
		&team.ID, &team.Name, &team.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	usersQuery := `
        SELECT id, user_id, username, is_active, team_id, created_at 
        FROM pr_system.users 
        WHERE team_id = $1`

	rows, err := r.storage.DB.Query(ctx, usersQuery, teamID)
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

	team.Users = users
	return &team, nil
}

func (r *TeamRepo) GetAllWithUsers(ctx context.Context) ([]domain.Team, error) {
	const op = "repository.TeamRepo.GetAllWithUsers"

	teamsQuery := `SELECT id, name, created_at FROM pr_system.teams ORDER BY name`
	rows, err := r.storage.DB.Query(ctx, teamsQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var teams []domain.Team
	for rows.Next() {
		var team domain.Team
		err := rows.Scan(&team.ID, &team.Name, &team.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		teams = append(teams, team)
	}

	for i := range teams {
		usersQuery := `
            SELECT id, user_id, username, is_active, team_id, created_at 
            FROM pr_system.users 
            WHERE team_id = $1 AND is_active = true`

		userRows, err := r.storage.DB.Query(ctx, usersQuery, teams[i].ID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		var users []domain.User
		for userRows.Next() {
			var user domain.User
			err := userRows.Scan(
				&user.ID, &user.UserID, &user.Username,
				&user.IsActive, &user.TeamID, &user.CreatedAt,
			)
			if err != nil {
				userRows.Close()
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			users = append(users, user)
		}
		userRows.Close()

		teams[i].Users = users
	}

	return teams, nil
}

func (r *TeamRepo) ExistsByName(ctx context.Context, teamName string) (bool, error) {
	const op = "repository.TeamRepo.ExistsByName"
	const query = `SELECT EXISTS(SELECT 1 FROM pr_system.teams WHERE name = $1)`

	var exists bool
	err := r.storage.DB.QueryRow(ctx, query, teamName).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return exists, nil
}
