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

type Film struct {
	Id               int     `json:"id,omitempty"`
	Title            string  `json:"title,omitempty"`
	OriginalTitle    string  `json:"original_title,omitempty"`
	OriginalLanguage string  `json:"original_language,omitempty"`
	GenreIds         []int   `json:"genre_ids,omitempty"`
	Genres           []Genre `json:"genres,omitempty"`
	Overview         string  `json:"overview,omitempty"`
	BackdropPath     string  `json:"backdrop_path,omitempty"`
	PosterPath       string  `json:"poster_path,omitempty"`
	ReleaseDate      string  `json:"release_date,omitempty"`
}

type CastCredit struct {
	Name string `json:"name,omitempty"`
	Role string `json:"role,omitempty"`
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
