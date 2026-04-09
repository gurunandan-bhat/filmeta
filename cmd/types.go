package cmd

import "time"

// Film metadata building blocks

type Genre struct {
	ID   int `json:"id"`
	Name string
}

type Person struct {
	Name string `json:"name"`
	Job  string `json:"job"`
}

type Credits struct {
	Cast []Person `json:"cast"`
	Crew []Person `json:"crew"`
}

type Meta struct {
	Language   string  `json:"original_language"`
	Overview   string  `json:"overview"`
	PosterPath string  `json:"poster_path"`
	Genres     []Genre `json:"genres"`
	Credits    Credits `json:"credits"`
}

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

// Algolia index representations

type FilmIndex struct {
	ObjectID        string `json:"objectID"`
	LinkTitle       string
	AverageScore    float64 `json:"AverageScore"`
	URLPath         string
	Genres          string
	Language        string
	Overview        string
	Cast            string
	Director        string
	Poster          string
	LocalPosterPath string `json:"LocalPosterPath"`
	Reviewers       string
}

type FilmReview struct {
	Critic string
}

type FilmOut struct {
	LinkTitle string
	ID        int
	ShowType  string
	Overview  string
}

type FCGFilm struct {
	ReviewCount int       `json:"count,omitempty"`
	Title       string    `json:"title,omitempty"`
	ShowType    string    `json:"show,omitempty"`
	TMDBID      int       `json:"id,omitempty"`
	ReviewDate  time.Time `json:"date,omitempty"`
}

type Data struct {
	Metadata []FCGFilm `json:"metadata,omitempty"`
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

type Critic struct {
	Name        string `json:"LinkTitle,omitempty"`
	ReviewCount int    `json:"ReviewCount,omitempty"`
	Lastmod     string `json:"Lastmod,omitempty"`
}

type CriticMap map[string]Critic

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
