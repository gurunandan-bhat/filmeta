package tmdb

import (
	"context"
	"log"
)

func (c *Client) FilmSearch(ctx context.Context, opts *SearchOptions) (*[]Film, error) {

	if opts.Page == 0 {
		opts.Page = 1
	}

	films := []Film{}
	for {
		results, err := c.Search(context.Background(), opts)
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
