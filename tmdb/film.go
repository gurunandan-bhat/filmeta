package tmdb

import (
	"context"
	"fmt"
	"net/http"
)

type Genre struct {
	Id   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type FilmMeta struct {
	Film `json:"film,omitempty"`
	Cast []CastCredit      `json:"cast,omitempty"`
	Crew map[string]string `json:"crew,omitempty"`
}

func (client *Client) Film(ctx context.Context, filmID int) (FilmMeta, error) {

	filmMeta := FilmMeta{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/movie/%d", client.BaseURL, filmID), nil)
	if err != nil {
		return filmMeta, fmt.Errorf("error initializing request: %w", err)
	}

	film := Film{}
	if err := client.sendRequest(req, &film); err != nil {
		return filmMeta, fmt.Errorf("response error: %w", err)
	}

	credits, err := client.Credits(context.Background(), filmID)
	if err != nil {
		return filmMeta, fmt.Errorf("error fetching credits for %s: %s", filmMeta.Title, err)
	}

	cast := make([]CastCredit, len(credits.Cast))
	for i, member := range credits.Cast {
		cast[i] = CastCredit{
			Name: member.Name,
			Role: member.Character,
		}
	}
	crew := make(map[string]string, len(credits.Crew))
	for _, member := range credits.Crew {
		crew[member.Job] = member.Name
	}

	filmMeta = FilmMeta{
		Film: film,
		Cast: cast,
		Crew: crew,
	}

	return filmMeta, nil
}
