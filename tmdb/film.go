package tmdb

import (
	"context"
	"fmt"
	"net/http"
)

type Film struct {
	Id               int     `json:"id,omitempty"`
	Title            string  `json:"title,omitempty"`
	Name             string  `json:"name,omitempty"`
	FCGTitle         string  `json:"fcg_title,omitempty"`
	OriginalTitle    string  `json:"original_title,omitempty"`
	OriginalLanguage string  `json:"original_language,omitempty"`
	GenreIds         []int   `json:"genre_ids,omitempty"`
	Genres           []Genre `json:"genres,omitempty"`
	Overview         string  `json:"overview,omitempty"`
	BackdropPath     string  `json:"backdrop_path,omitempty"`
	PosterPath       string  `json:"poster_path,omitempty"`
	ReleaseDate      string  `json:"release_date,omitempty"`
}

type Genre struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type CastMember struct {
	Id          int    `json:"id,omitempty"`
	Order       int    `json:"order"`
	Name        string `json:"name,omitempty"`
	ProfilePath string `json:"profile_path,omitempty"`
	Character   string `json:"character,omitempty"`
	Gender      int    `json:"gender,omitempty"`
}

type CrewMember struct {
	Id          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	ProfilePath string `json:"profile_path,omitempty"`
	Job         string `json:"job,omitempty"`
	Department  string `json:"department,omitempty"`
	Gender      int    `json:"gender,omitempty"`
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
	} `json:"credits"`
}

func (client *Client) Film(ctx context.Context, showType string, filmID int) (FilmWithCredits, error) {

	film := FilmWithCredits{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s/%d?append_to_response=credits", client.BaseURL, showType, filmID), nil)
	if err != nil {
		return film, fmt.Errorf("error initializing request: %w", err)
	}

	if err := client.sendRequest(req, &film); err != nil {
		return film, fmt.Errorf("response error: %w", err)
	}

	return film, nil
}
