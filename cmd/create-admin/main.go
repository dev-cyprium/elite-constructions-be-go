package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/auth"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/config"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/term"
)

func main() {
	var name, email, password string
	flag.StringVar(&name, "name", "", "User name (required)")
	flag.StringVar(&email, "email", "", "User email (required)")
	flag.StringVar(&password, "password", "", "User password (required, or will prompt if not provided)")
	flag.Parse()

	// Validate required flags
	if name == "" {
		log.Fatal("Error: --name is required")
	}
	if email == "" {
		log.Fatal("Error: --email is required")
	}

	// If password not provided, prompt for it
	if password == "" {
		fmt.Print("Enter password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Error reading password: %v", err)
		}
		password = string(passwordBytes)
		fmt.Println() // New line after password input
	}

	if len(password) < 8 {
		log.Fatal("Error: password must be at least 8 characters")
	}

	// Load configuration
	// For this command, we only need DATABASE_URL, but config.Load requires JWT_SECRET
	// So we'll set a dummy JWT_SECRET if it's not set
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "dummy-secret-for-create-admin-command-only-min-32-chars-long")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	if err := db.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Hash password with Argon2id
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create sqlc queries instance
	queries := sqlc.New(db.Pool)
	ctx := context.Background()

	// Check if user with email already exists
	_, err = queries.GetUserByEmail(ctx, email)
	if err == nil {
		log.Fatalf("Error: user with email %s already exists", email)
	}
	// If error is not "no rows", it's a real error we should report
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Fatalf("Error checking for existing user: %v", err)
	}

	// Create user
	user, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		Name:                  name,
		Email:                 email,
		Password:              hashedPassword,
		PasswordResetRequired: pgtype.Bool{Bool: false, Valid: true},
	})
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Admin user created successfully!\n")
	fmt.Printf("ID: %d\n", user.ID)
	fmt.Printf("Name: %s\n", user.Name)
	fmt.Printf("Email: %s\n", user.Email)
}
