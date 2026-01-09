package db

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/stdlib"
)

// Run runs all pending migrations
func Run() error {
	if Pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}

	// Get underlying *sql.DB from pgxpool
	sqlDB := stdlib.OpenDB(*Pool.Config().ConnConfig)
	defer sqlDB.Close()

	// Create postgres driver instance
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get migrations directory path (relative to project root)
	migrationsPath := filepath.Join("migrations")
	if !filepath.IsAbs(migrationsPath) {
		// If relative, try to resolve from current working directory
		// In production, migrations will be in ./migrations relative to the binary
		migrationsPath = filepath.Join(".", "migrations")
	}

	// Create migrate instance using file source
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// GetVersion returns the current migration version
func GetVersion() (uint, bool, error) {
	if Pool == nil {
		return 0, false, fmt.Errorf("database connection pool is not initialized")
	}

	sqlDB := stdlib.OpenDB(*Pool.Config().ConnConfig)
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get migrations directory path
	migrationsPath := filepath.Join("migrations")
	if !filepath.IsAbs(migrationsPath) {
		migrationsPath = filepath.Join(".", "migrations")
	}

	// Create migrate instance using file source
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
