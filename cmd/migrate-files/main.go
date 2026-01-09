package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	sourceDir  = flag.String("source-dir", "", "Source directory containing existing files")
	targetDir  = flag.String("target-dir", "", "Target directory (./storage/public)")
	postgresDSN = flag.String("postgres-dsn", "", "PostgreSQL DSN")
)

func main() {
	flag.Parse()

	if *sourceDir == "" || *targetDir == "" || *postgresDSN == "" {
		fmt.Println("Usage: migrate-files --source-dir=... --target-dir=... --postgres-dsn=...")
		os.Exit(1)
	}

	// Connect to PostgreSQL
	db, err := sql.Open("pgx", *postgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Query project_images.url from DB
	rows, err := db.Query("SELECT url FROM project_images")
	if err != nil {
		log.Fatalf("Failed to query project_images: %v", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Printf("Error scanning URL: %v", err)
			continue
		}
		urls = append(urls, url)
	}

	fmt.Printf("Found %d image URLs in database\n", len(urls))

	// Process URLs
	copied := 0
	missing := []string{}
	skipped := []string{}

	for _, url := range urls {
		// Parse URL
		filename, err := parseURL(url)
		if err != nil {
			skipped = append(skipped, url)
			fmt.Printf("Skipping invalid URL: %s\n", url)
			continue
		}

		// Check if external URL
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			skipped = append(skipped, url)
			fmt.Printf("Skipping external URL: %s\n", url)
			continue
		}

		// Find source file
		sourcePath := filepath.Join(*sourceDir, "img", filename)
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			missing = append(missing, url)
			fmt.Printf("Missing file: %s\n", sourcePath)
			continue
		}

		// Copy file
		targetPath := filepath.Join(*targetDir, "img", filename)
		if err := copyFile(sourcePath, targetPath); err != nil {
			log.Printf("Error copying %s: %v", sourcePath, err)
			continue
		}

		copied++
	}

	// Print report
	fmt.Println("\n=== Migration Report ===")
	fmt.Printf("Copied: %d files\n", copied)
	fmt.Printf("Missing: %d files\n", len(missing))
	fmt.Printf("Skipped: %d files\n", len(skipped))

	if len(missing) > 0 {
		fmt.Println("\nMissing files:")
		for _, url := range missing {
			fmt.Printf("  - %s\n", url)
		}
	}
}

func parseURL(url string) (string, error) {
	// Pattern: /storage/img/filename.jpg -> filename.jpg
	if strings.HasPrefix(url, "/storage/img/") {
		return strings.TrimPrefix(url, "/storage/img/"), nil
	}

	// Pattern: ../public/img/file.jpg -> resolve relative
	if strings.Contains(url, "../") || strings.Contains(url, "./") {
		// Extract filename from path
		return filepath.Base(url), nil
	}

	// Try to extract filename
	filename := filepath.Base(url)
	if filename != "" && filename != "." {
		return filename, nil
	}

	return "", fmt.Errorf("could not parse URL: %s", url)
}

func copyFile(src, dst string) error {
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Create destination
	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	// Copy
	_, err = io.Copy(dest, source)
	return err
}
