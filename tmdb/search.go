package tmdb

import (
	"context"
	"fmt"
	"log"
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

func (c *Client) ShowSearch(ctx context.Context, showType string, opts *SearchOptions) (*[]Film, error) {

	if opts.Page == 0 {
		opts.Page = 1
	}

	films := []Film{}
	for {
		results, err := c.search(context.Background(), showType, opts)
		if err != nil {
			log.Fatalf("err calling search: %s", err)
		}
		films = append(films, results.Results...)
		if results.Page == results.TotalPages {
			break
		}

		opts.Page = opts.Page + 1
	}

	return &films, nil
}

func (c *Client) search(ctx context.Context, showType string, opts *SearchOptions) (*SearchResults, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/search/%s", c.BaseURL, showType), nil)
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
