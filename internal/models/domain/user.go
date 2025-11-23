package domain

import "time"

type User struct {
	ID        int64     `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Username  string    `db:"username" json:"username"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	TeamID    int64     `db:"team_id" json:"team_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at,omitempty"`
}
