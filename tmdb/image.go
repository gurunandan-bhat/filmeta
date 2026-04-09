package tmdb

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func (c *Client) TMDBImage(ctx context.Context, baseURL, imgPath, destPath string) error {

	imgURL := fmt.Sprintf("%s%s", baseURL, imgPath)
	req, err := http.NewRequestWithContext(ctx, "GET", imgURL, nil)
	if err != nil {
		return fmt.Errorf("error initializing request: %w", err)
	}

	// os.Create fails if the parent directory does not exist
	// so check if it needs to be created and create it. This is
	// not really required (see calling code) but it might if
	// called from elsewhere
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", destDir, err)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", destPath, err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Fatalf("error closing file %s: %v", destPath, err)
		}
	}()

	if err := c.sendRequest(req, outFile); err != nil {
		return fmt.Errorf("response error for tmdb image request: %w", err)
	}

	return nil
}
