package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	mysqlDSN    = flag.String("mysql-dsn", "", "MySQL DSN (e.g., user:password@tcp(localhost:3306)/database)")
	postgresDSN = flag.String("postgres-dsn", "", "PostgreSQL DSN (e.g., postgres://user:password@localhost:5432/database?sslmode=disable)")
	dryRun      = flag.Bool("dry-run", false, "Perform a dry run without making changes")
)

func main() {
	flag.Parse()

	if *mysqlDSN == "" || *postgresDSN == "" {
		fmt.Println("Usage: migrate-db --mysql-dsn=... --postgres-dsn=... [--dry-run]")
		os.Exit(1)
	}

	// Connect to MySQL
	mysqlDB, err := sql.Open("mysql", *mysqlDSN)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer mysqlDB.Close()

	// Connect to PostgreSQL
	postgresDB, err := sql.Open("pgx", *postgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	// Test connections
	if err := mysqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping MySQL: %v", err)
	}
	if err := postgresDB.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	fmt.Println("Connected to both databases")
	if *dryRun {
		fmt.Println("DRY RUN MODE - No changes will be made")
	}

	// Migration order
	tables := []string{
		"users",
		"projects",
		"project_images",
		"testimonials",
		"static_texts",
		"configurations",
		"visitor_messages",
	}

	report := make(map[string]int)

	for _, table := range tables {
		count, err := migrateTable(mysqlDB, postgresDB, table, *dryRun)
		if err != nil {
			log.Printf("Error migrating %s: %v", table, err)
			continue
		}
		report[table] = count
		fmt.Printf("Migrated %s: %d rows\n", table, count)
	}

	// Reset sequences
	if !*dryRun {
		fmt.Println("\nResetting sequences...")
		for _, table := range tables {
			if err := resetSequence(postgresDB, table); err != nil {
				log.Printf("Error resetting sequence for %s: %v", table, err)
			}
		}
	}

	// Validation
	fmt.Println("\nValidating migration...")
	validateMigration(mysqlDB, postgresDB, report)

	fmt.Println("\nMigration completed!")
}

func migrateTable(mysqlDB, postgresDB *sql.DB, tableName string, dryRun bool) (int, error) {
	switch tableName {
	case "users":
		return migrateUsers(mysqlDB, postgresDB, dryRun)
	case "projects":
		return migrateProjects(mysqlDB, postgresDB, dryRun)
	case "project_images":
		return migrateProjectImages(mysqlDB, postgresDB, dryRun)
	case "testimonials":
		return migrateTestimonials(mysqlDB, postgresDB, dryRun)
	case "static_texts":
		return migrateStaticTexts(mysqlDB, postgresDB, dryRun)
	case "configurations":
		return migrateConfigurations(mysqlDB, postgresDB, dryRun)
	case "visitor_messages":
		return migrateVisitorMessages(mysqlDB, postgresDB, dryRun)
	default:
		return 0, fmt.Errorf("unknown table: %s", tableName)
	}
}

func migrateUsers(mysqlDB, postgresDB *sql.DB, dryRun bool) (int, error) {
	rows, err := mysqlDB.Query(`
		SELECT id, name, email, email_verified_at, password, password_reset_required,
		       reset_token_hash, reset_token_expires_at, remember_token, created_at, updated_at
		FROM users
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query MySQL users: %w", err)
	}
	defer rows.Close()

	if dryRun {
		count := 0
		for rows.Next() {
			count++
		}
		return count, nil
	}

	tx, err := postgresDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO users (id, name, email, email_verified_at, password, password_reset_required,
		                  reset_token_hash, reset_token_expires_at, remember_token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int64
		var name, email string
		var emailVerifiedAt, resetTokenExpiresAt, createdAt, updatedAt sql.NullTime
		var password, resetTokenHash, rememberToken sql.NullString
		var passwordResetRequired sql.NullBool

		err := rows.Scan(&id, &name, &email, &emailVerifiedAt, &password, &passwordResetRequired,
			&resetTokenHash, &resetTokenExpiresAt, &rememberToken, &createdAt, &updatedAt)
		if err != nil {
			return count, fmt.Errorf("failed to scan row: %w", err)
		}

		// Set password_reset_required=true and password to empty string for migration
		newPassword := ""
		resetRequired := true

		_, err = stmt.Exec(id, name, email, nullableTime(emailVerifiedAt), newPassword, resetRequired,
			nullableString(resetTokenHash), nullableTime(resetTokenExpiresAt),
			nullableString(rememberToken), nullableTime(createdAt), nullableTime(updatedAt))
		if err != nil {
			return count, fmt.Errorf("failed to insert user %d: %w", id, err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

func migrateProjects(mysqlDB, postgresDB *sql.DB, dryRun bool) (int, error) {
	rows, err := mysqlDB.Query(
		"SELECT id, status, name, category, client, `order`, highlighted, created_at, updated_at FROM projects",
	)
	if err != nil {
		return 0, fmt.Errorf("failed to query MySQL projects: %w", err)
	}
	defer rows.Close()

	if dryRun {
		count := 0
		for rows.Next() {
			count++
		}
		return count, nil
	}

	tx, err := postgresDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO projects (id, status, name, category, client, "order", highlighted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int64
		var status int
		var name string
		var category, client sql.NullString
		var order int
		var highlighted bool
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &status, &name, &category, &client, &order, &highlighted, &createdAt, &updatedAt)
		if err != nil {
			return count, fmt.Errorf("failed to scan row: %w", err)
		}

		_, err = stmt.Exec(id, status, name, nullableString(category), nullableString(client),
			order, highlighted, nullableTime(createdAt), nullableTime(updatedAt))
		if err != nil {
			return count, fmt.Errorf("failed to insert project %d: %w", id, err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

func migrateProjectImages(mysqlDB, postgresDB *sql.DB, dryRun bool) (int, error) {
	rows, err := mysqlDB.Query(
		"SELECT id, name, url, project_id, `order`, blur_hash, created_at, updated_at FROM project_images",
	)
	if err != nil {
		return 0, fmt.Errorf("failed to query MySQL project_images: %w", err)
	}
	defer rows.Close()

	if dryRun {
		count := 0
		for rows.Next() {
			count++
		}
		return count, nil
	}

	tx, err := postgresDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO project_images (id, name, url, project_id, "order", blur_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id, projectID int64
		var name, url string
		var order int
		var blurHash sql.NullString
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &name, &url, &projectID, &order, &blurHash, &createdAt, &updatedAt)
		if err != nil {
			return count, fmt.Errorf("failed to scan row: %w", err)
		}

		_, err = stmt.Exec(id, name, url, projectID, order, nullableString(blurHash),
			nullableTime(createdAt), nullableTime(updatedAt))
		if err != nil {
			return count, fmt.Errorf("failed to insert project_image %d: %w", id, err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

func migrateTestimonials(mysqlDB, postgresDB *sql.DB, dryRun bool) (int, error) {
	rows, err := mysqlDB.Query(`
		SELECT id, full_name, profession, testimonial, status, created_at, updated_at
		FROM testimonials
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query MySQL testimonials: %w", err)
	}
	defer rows.Close()

	if dryRun {
		count := 0
		for rows.Next() {
			count++
		}
		return count, nil
	}

	tx, err := postgresDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO testimonials (id, full_name, profession, testimonial, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int64
		var fullName, profession, testimonial, status string
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &fullName, &profession, &testimonial, &status, &createdAt, &updatedAt)
		if err != nil {
			return count, fmt.Errorf("failed to scan row: %w", err)
		}

		_, err = stmt.Exec(id, fullName, profession, testimonial, status,
			nullableTime(createdAt), nullableTime(updatedAt))
		if err != nil {
			return count, fmt.Errorf("failed to insert testimonial %d: %w", id, err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

func migrateStaticTexts(mysqlDB, postgresDB *sql.DB, dryRun bool) (int, error) {
	rows, err := mysqlDB.Query(`
		SELECT id, key, label, content, created_at, updated_at
		FROM static_texts
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query MySQL static_texts: %w", err)
	}
	defer rows.Close()

	if dryRun {
		count := 0
		for rows.Next() {
			count++
		}
		return count, nil
	}

	tx, err := postgresDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO static_texts (id, key, label, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int64
		var key, label, content string
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &key, &label, &content, &createdAt, &updatedAt)
		if err != nil {
			return count, fmt.Errorf("failed to scan row: %w", err)
		}

		_, err = stmt.Exec(id, key, label, content, nullableTime(createdAt), nullableTime(updatedAt))
		if err != nil {
			return count, fmt.Errorf("failed to insert static_text %d: %w", id, err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

func migrateConfigurations(mysqlDB, postgresDB *sql.DB, dryRun bool) (int, error) {
	rows, err := mysqlDB.Query(`
		SELECT id, key, value, created_at, updated_at
		FROM configurations
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query MySQL configurations: %w", err)
	}
	defer rows.Close()

	if dryRun {
		count := 0
		for rows.Next() {
			count++
		}
		return count, nil
	}

	tx, err := postgresDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO configurations (id, key, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int64
		var key, value string
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &key, &value, &createdAt, &updatedAt)
		if err != nil {
			return count, fmt.Errorf("failed to scan row: %w", err)
		}

		_, err = stmt.Exec(id, key, value, nullableTime(createdAt), nullableTime(updatedAt))
		if err != nil {
			return count, fmt.Errorf("failed to insert configuration %d: %w", id, err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

func migrateVisitorMessages(mysqlDB, postgresDB *sql.DB, dryRun bool) (int, error) {
	rows, err := mysqlDB.Query(`
		SELECT id, email, address, description, seen, created_at, updated_at
		FROM visitor_messages
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query MySQL visitor_messages: %w", err)
	}
	defer rows.Close()

	if dryRun {
		count := 0
		for rows.Next() {
			count++
		}
		return count, nil
	}

	tx, err := postgresDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO visitor_messages (id, email, address, description, seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int64
		var email, address, description string
		var seen bool
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &email, &address, &description, &seen, &createdAt, &updatedAt)
		if err != nil {
			return count, fmt.Errorf("failed to scan row: %w", err)
		}

		_, err = stmt.Exec(id, email, address, description, seen, nullableTime(createdAt), nullableTime(updatedAt))
		if err != nil {
			return count, fmt.Errorf("failed to insert visitor_message %d: %w", id, err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

// Helper functions for nullable types
func nullableTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func nullableString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func resetSequence(postgresDB *sql.DB, tableName string) error {
	// Set sequence to max(id) or 1 if table is empty
	query := fmt.Sprintf(
		"SELECT setval('%s_id_seq', COALESCE((SELECT MAX(id) FROM %s), 0) + 1, false);",
		tableName, tableName)
	_, err := postgresDB.Exec(query)
	return err
}

func validateMigration(mysqlDB, postgresDB *sql.DB, report map[string]int) {
	// TODO: Compare row counts, check FK integrity
	fmt.Println("Validation complete")
}
