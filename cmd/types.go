package cmd

import "time"

// Film is the canonical type for decoding mreviews/index.json

type Film struct {
	LinkTitle       string    `json:"LinkTitle"`
	AverageScore    float64   `json:"AverageScore"`
	URLPath         string    `json:"URLPath"`
	Path            string    `json:"Path"`
	PosterPath      string    `json:"PosterPath"`
	LocalPosterPath string    `json:"LocalPosterPath"`
	Lastmod         time.Time `json:"Lastmod"`
}

type FilmOut struct {
	LinkTitle string
	ID        int
	ShowType  string
	Overview  string
}

// Hugo content post

type PostFormat struct {
	Title    string    `toml:"title"`
	Date     time.Time `toml:"date"`
	Draft    bool      `toml:"draft"`
	Cast     []string  `toml:"cast"`
	Genres   []string  `toml:"genres"`
	Director []string  `toml:"director"`
	Language []string  `toml:"language"`
}

// Guild and critics

type Guild struct {
	Name          string   `json:"LinkTitle,omitempty"`
	ReviewURL     string   `json:"ReviewURL,omitempty"`
	Organizations []string `json:"Organizations,omitempty"`
	Path          string   `json:"Path,omitempty"`
}

type CriticReview struct {
	Publication string
	PublishDate time.Time
}

// Reviews and scoring

type Scores map[string]float64

type FreeScores map[string]Scores

// Utility

type Entity struct {
	LinkTitle string `json:"LinkTitle,omitempty"`
	Path      string `json:"Path,omitempty"`
}
