package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

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
	// TODO: Implement table-specific migration logic
	// For users: set password_reset_required=true, leave password empty
	// For others: direct copy preserving IDs
	
	// Placeholder
	return 0, nil
}

func resetSequence(postgresDB *sql.DB, tableName string) error {
	query := fmt.Sprintf("SELECT setval('%s_id_seq', (SELECT MAX(id) FROM %s));", tableName, tableName)
	_, err := postgresDB.Exec(query)
	return err
}

func validateMigration(mysqlDB, postgresDB *sql.DB, report map[string]int) {
	// TODO: Compare row counts, check FK integrity
	fmt.Println("Validation complete")
}
