package tmdb

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func (c *Client) TMDBImage(ctx context.Context, baseURL, imgPath, outDir string) error {

	imgURL := fmt.Sprintf("%s%s", baseURL, imgPath)
	req, err := http.NewRequestWithContext(ctx, "GET", imgURL, nil)
	if err != nil {
		return fmt.Errorf("error initializing request: %w", err)
	}

	fPath := filepath.Join(outDir, filepath.Base(imgPath))
	file, err := os.Create(fPath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", fPath, err)
	}
	defer file.Close()

	if err := c.sendRequest(req, file); err != nil {
		return fmt.Errorf("response error: %w", err)
	}

	return nil
}
