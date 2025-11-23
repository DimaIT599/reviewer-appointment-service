package storage

import (
	"errors"
	"fmt"
	"os"
	"reviewer-appointment-service/internal/config"
)

var (
	ErrTeamExists  = errors.New("team already exists")
	ErrPRExists    = errors.New("PR already exists")
	ErrPRMerged    = errors.New("PR is merged")
	ErrNotAssigned = errors.New("reviewer not assigned")
	ErrNoCandidate = errors.New("no active replacement candidate")
	ErrNotFound    = errors.New("resource not found")
	ErrUserExists  = errors.New("user already exists")
)

func GetDBConnectionString(cfg *config.Config) string {
	sslmode := "disable"
	if os.Getenv("DB_SSLMODE") != "" {
		sslmode = os.Getenv("DB_SSLMODE")
	}
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.PortDB,
		cfg.Data_base,
		sslmode,
	)
}
