package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        int
	StoragePath string
}

// Load loads configuration from environment variables
// It first tries to load from .env file, then falls back to environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables only: %v", err)
	}
	cfg := &Config{}

	// Database URL
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// JWT Secret
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	// Port
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("PORT must be a valid integer: %w", err)
	}
	cfg.Port = port

	// Storage Path
	cfg.StoragePath = os.Getenv("STORAGE_PATH")
	if cfg.StoragePath == "" {
		cfg.StoragePath = "./storage"
	}

	return cfg, nil
}
