package tmdb

import (
	"context"
	"fmt"
	"net/http"
)

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

type Credits struct {
	Id   int          `json:"id,omitempty"`
	Cast []CastMember `json:"cast,omitempty"`
	Crew []CrewMember `json:"crew,omitempty"`
}

func (c *Client) Credits(ctx context.Context, filmId int) (*Credits, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/movie/%d/credits", c.BaseURL, filmId), nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing request: %w", err)
	}

	results := Credits{}
	if err := c.sendRequest(req, &results); err != nil {
		return nil, fmt.Errorf("response error: %w", err)
	}

	return &results, nil
}
