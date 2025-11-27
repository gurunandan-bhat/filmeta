package guild

import "time"

type Film struct {
	LinkTitle    string    `json:"LinkTitle,omitempty"`
	Lastmod      time.Time `json:"Lastmod,omitempty"`
	URLPath      string    `json:"URLPath,omitempty"`
	PosterPath   string    `json:"PosterPath,omitempty"`
	AverageScore float64   `json:"AverageScore,omitempty"`
}

type ReviewParams struct {
	Critics   []string `json:"critics,omitempty"`
	ImagePath string   `json:"img,omitempty"`
	Media     string   `json:"media,omitempty"`
	Score     float64  `json:"score,omitempty"`
	Source    string   `json:"source,omitempty"`
	SubTitle  string   `json:"subtitle,omitempty"`
	Opening   string   `json:"opening,omitempty"`
}

type Review struct {
	LinkTitle string       `json:"LinkTitle,omitempty"`
	Lastmod   time.Time    `json:"Lastmod,omitempty"`
	Path      string       `json:"Path,omitempty"`
	Params    ReviewParams `json:"params,omitempty"`
}
