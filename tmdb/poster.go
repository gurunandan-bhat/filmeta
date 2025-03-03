package tmdb

import (
	"context"
	"filmeta/config"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func (c *Client) Poster(ctx context.Context, posterPath, outDir string) error {

	cfg, err := config.Configuration()
	if err != nil {
		return fmt.Errorf("error fetching configuration: %w", err)
	}
	imgBase := cfg.TMDB.ImgBase
	imgURL := fmt.Sprintf("%s%s", imgBase, posterPath)
	req, err := http.NewRequestWithContext(ctx, "GET", imgURL, nil)
	if err != nil {
		return fmt.Errorf("error initializing request: %w", err)
	}

	fPath := filepath.Join(outDir, filepath.Base(posterPath))
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
