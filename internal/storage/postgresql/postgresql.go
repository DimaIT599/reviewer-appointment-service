package postgresql

import (
	"context"
	"fmt"
	"log"
	"reviewer-appointment-service/internal/config"
	"reviewer-appointment-service/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	DB *pgxpool.Pool
}

func NewStorage(cfg *config.Config, ctx context.Context) (*Storage, error) {
	const op = "storage.postgresql.NewStorage"

	connectString := storage.GetDBConnectionString(cfg)
	log.Printf("Connection string: %s", connectString)

	poolConfig, err := pgxpool.ParseConfig(connectString)
	if err != nil {
		return nil, fmt.Errorf("%s parse config error: %w", op, err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%s create connection pool error: %w", op, err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s ping database error: %w", op, err)
	}

	return &Storage{DB: pool}, nil
}

func (s *Storage) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
}
