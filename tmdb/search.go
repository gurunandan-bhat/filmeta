package tmdb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type SearchResults struct {
	Results      []Film `json:"results,omitempty"`
	Page         int    `json:"page,omitempty"`
	TotalPages   int    `json:"total_pages,omitempty"`
	TotalResults int    `json:"total_results,omitempty"`
}

type SearchOptions struct {
	Query        string `schema:"query,required"`
	IncludeAdult bool   `schema:"include_adult,default:false"`
	Language     string `schema:"language,omitempty"`
	Page         int    `schema:"page,default:1"`
	Region       string `schema:"region,omitempty"`
	Year         string `schema:"year,omitempty"`
}

func (c *Client) Search(ctx context.Context, opts *SearchOptions) (*SearchResults, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/search/movie", c.BaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing request: %w", err)
	}

	v := url.Values{}
	if err := encoder.Encode(opts, v); err != nil {
		return nil, fmt.Errorf("error encoding query params: %w", err)
	}

	req.URL.RawQuery = v.Encode()
	results := SearchResults{}
	if err := c.sendRequest(req, &results); err != nil {
		return nil, fmt.Errorf("response error: %w", err)
	}

	return &results, nil
}
