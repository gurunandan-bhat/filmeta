package tmdb

import (
	"context"
	"fmt"
	"net/http"
)

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

type FilmWithCredits struct {
	Film
	Credits struct {
		Cast []CastMember `json:"cast,omitempty"`
		Crew []CrewMember `json:"crew,omitempty"`
	} `json:"credits,omitempty"`
}

func (client *Client) FilmWithCredits(ctx context.Context, filmID int) (FilmWithCredits, error) {

	film := FilmWithCredits{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/movie/%d?append_to_response=credits", client.BaseURL, filmID), nil)
	if err != nil {
		return film, fmt.Errorf("error initializing request: %w", err)
	}

	if err := client.sendRequest(req, &film); err != nil {
		return film, fmt.Errorf("response error: %w", err)
	}

	return film, nil
}
