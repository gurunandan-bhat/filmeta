package tmdb

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func (c *Client) TMDBImage(ctx context.Context, baseURL, imgURI, destPath string) (err error) {

	imgURL := fmt.Sprintf("%s%s", baseURL, imgURI)
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
		return fmt.Errorf("tmdbImage: error creating directory %s: %w", destDir, err)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", destPath, err)
	}
	defer func() {
		if closeErr := outFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing file %s: %w", destPath, closeErr)
		}
	}()

	if err := c.sendRequest(req, outFile); err != nil {
		return fmt.Errorf("response error for tmdb image request: %w", err)
	}

	return nil
}
