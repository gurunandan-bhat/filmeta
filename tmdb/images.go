package tmdb

import (
	"context"
	"fmt"
	"net/http"
)

type Image struct {
	FilePath     string  `json:"file_path,omitempty"`
	Height       int     `json:"height,omitempty"`
	Width        int     `json:"width,omitempty"`
	AspectRation float64 `json:"aspect_ration,omitempty"`
	Iso6391      string  `json:"iso_639_1,omitempty"`
	VoteAverage  float64 `json:"vote_average,omitempty"`
	VoteCount    int     `json:"vote_count,omitempty"`
}

type FilmImages struct {
	Id        int     `json:"id,omitempty"`
	Backdrops []Image `json:"backdrops,omitempty"`
	Logos     []Image `json:"logos,omitempty"`
	Posters   []Image `json:"posters,omitempty"`
}

func (c *Client) FilmImages(ctx context.Context, filmId int) (*FilmImages, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/movie/%d/images", c.BaseURL, filmId), nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing request: %w", err)
	}

	results := FilmImages{}
	if err := c.sendRequest(req, &results); err != nil {
		return nil, fmt.Errorf("response error: %w", err)
	}

	return &results, nil
}
