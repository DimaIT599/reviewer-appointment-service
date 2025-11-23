package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"reviewer-appointment-service/internal"
	"reviewer-appointment-service/internal/config"
	"reviewer-appointment-service/internal/storage/postgresql"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config.yaml"
	}

	log.Printf("Loading config from: %s (CONFIG_PATH env: %s)", configPath, os.Getenv("CONFIG_PATH"))
	cfg := config.MustConfig(configPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("Connecting to database: host=%s, port=%d, database=%s, user=%s",
		cfg.Host, cfg.PortDB, cfg.Data_base, cfg.User)

	storage, err := waitForDatabase(ctx, cfg, 30*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v\n"+
			"Please ensure PostgreSQL is running and accessible.\n"+
			"You can use: docker-compose up -d (to start PostgreSQL)\n"+
			"Or set environment variables: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME", err)
	}
	defer storage.Close()

	log.Println("Successfully connected to database")

	port := fmt.Sprintf("%d", cfg.Server.PortServer)
	server := internal.NewServer(port, storage)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Run(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")
}

func createDatabaseIfNotExists(cfg *config.Config) error {
	adminCfg := *cfg
	adminCfg.Data_base = "postgres"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adminStorage, err := postgresql.NewStorage(&adminCfg, ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to admin database: %w", err)
	}
	defer adminStorage.Close()

	var exists bool
	err = adminStorage.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.Data_base).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = adminStorage.DB.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", cfg.Data_base))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database '%s' created successfully", cfg.Data_base)
	}

	return nil
}

func waitForDatabase(ctx context.Context, cfg *config.Config, timeout time.Duration) (*postgresql.Storage, error) {
	deadline := time.Now().Add(timeout)
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for database connection")
		}

		storage, err := postgresql.NewStorage(cfg, ctx)
		if err == nil {
			return storage, nil
		}

		if i == 0 {
			if createErr := createDatabaseIfNotExists(cfg); createErr != nil {
				log.Printf("Warning: failed to create database if not exists: %v", createErr)
			}
		}

		log.Printf("Database connection attempt %d/%d failed: %v. Retrying in %v...", i+1, maxRetries, err, retryDelay)
		time.Sleep(retryDelay)

		if i < 5 {
			retryDelay = 2 * time.Second
		} else if i < 10 {
			retryDelay = 3 * time.Second
		} else {
			retryDelay = 5 * time.Second
		}
	}

	return nil, fmt.Errorf("max retries exceeded")
}
