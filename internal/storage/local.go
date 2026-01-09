package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Allowed image MIME types
	allowedImageTypes = "image/jpeg,image/png,image/webp"
)

// SaveFile saves a file to the local filesystem and returns the public URL
func SaveFile(fileData []byte, originalFilename, storagePath string) (string, error) {
	// Validate image type
	contentType := http.DetectContentType(fileData)
	if !isAllowedImageType(contentType) {
		return "", fmt.Errorf("invalid image type: %s. Allowed types: %s", contentType, allowedImageTypes)
	}

	// Generate unique filename (SHA1 of file data + extension)
	hash := sha1.Sum(fileData)
	hashStr := hex.EncodeToString(hash[:])
	
	ext := filepath.Ext(originalFilename)
	if ext == "" {
		// Try to determine extension from content type
		exts, _ := mime.ExtensionsByType(contentType)
		if len(exts) > 0 {
			ext = exts[0]
		} else {
			ext = ".jpg" // default
		}
	}

	filename := hashStr + ext
	filePath := filepath.Join(storagePath, "public", "img", filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return public URL path
	return fmt.Sprintf("/storage/img/%s", filename), nil
}

// DeleteFile deletes a file from the local filesystem
func DeleteFile(url, storagePath string) error {
	// Extract filename from URL
	if !strings.HasPrefix(url, "/storage/img/") {
		return fmt.Errorf("invalid URL format: %s", url)
	}

	filename := strings.TrimPrefix(url, "/storage/img/")
	filePath := filepath.Join(storagePath, "public", "img", filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, consider it already deleted
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// isAllowedImageType checks if the content type is an allowed image type
func isAllowedImageType(contentType string) bool {
	allowed := []string{"image/jpeg", "image/png", "image/webp"}
	for _, t := range allowed {
		if contentType == t {
			return true
		}
	}
	return false
}
