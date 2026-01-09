package db

import (
	"context"
	"fmt"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool holds the PostgreSQL connection pool
var Pool *pgxpool.Pool

// Connect initializes the PostgreSQL connection pool
func Connect(cfg *config.Config) error {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	Pool = pool
	return nil
}

// Close closes the connection pool
func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
