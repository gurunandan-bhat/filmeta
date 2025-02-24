package tmdb

import (
	"context"
	"fmt"
	"net/http"
)

type Images struct {
	BaseUrl       string   `json:"base_url,omitempty"`
	SecureBaseUrl string   `json:"secure_base_url,omitempty"`
	BackdropSizes []string `json:"backdrop_sizes,omitempty"`
	LogoSizes     []string `json:"logo_sizes,omitempty"`
	PosterSizes   []string `json:"poster_sizes,omitempty"`
	StillSizes    []string `json:"still_sizes,omitempty"`
}

type Configuration struct {
	Images     Images   `json:"images,omitempty"`
	ChangeKeys []string `json:"change_keys,omitempty"`
}

type ConfigurationOptions struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
}

func (c *Client) GetConfiguration(ctx context.Context, options *ConfigurationOptions) (*Configuration, error) {

	limit := 100
	page := 1
	if options != nil {
		limit = options.Limit
		page = options.Page
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/faces?limit=%d&page=%d", c.BaseURL, limit, page), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	res := Configuration{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
