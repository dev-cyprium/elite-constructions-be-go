package storage

import (
	"encoding/base64"
	"fmt"
	"image"
	"os"

	"github.com/buckket/go-blurhash"
	_ "image/jpeg"
	_ "image/png"
)

// GenerateBlurHash generates a blurhash from an image file and returns it as a data URL
func GenerateBlurHash(filePath string) (string, error) {
	// Open and decode the image
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Generate blurhash (components: 4x4 is a good balance)
	hash, err := blurhash.Encode(4, 4, img)
	if err != nil {
		return "", fmt.Errorf("failed to generate blurhash: %w", err)
	}

	// Convert to data URL format: data:image/svg+xml;base64,<base64-encoded-svg>
	// For blurhash, we'll use a simple data URL format
	// Actually, blurhash is just a string, so we'll return it as-is
	// But the spec says "data URL format", so let's encode it as base64 in a data URL
	encoded := base64.StdEncoding.EncodeToString([]byte(hash))
	return fmt.Sprintf("data:text/plain;base64,%s", encoded), nil
}
