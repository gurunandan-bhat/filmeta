package cmd

import "time"

// Film is the canonical type for decoding mreviews/index.json

type Film struct {
	LinkTitle       string    `json:"LinkTitle"`
	Language        string    `json:"Language"`
	AverageScore    float64   `json:"AverageScore"`
	URLPath         string    `json:"URLPath"`
	Path            string    `json:"Path"`
	PosterPath      string    `json:"PosterPath"`
	LocalPosterPath string    `json:"LocalPosterPath"`
	Lastmod         time.Time `json:"Lastmod"`
}

// FilmOut is both missingMeta's output and import's input. missingMeta emits a
// blank BackdropPath for every gap it reports; filling it in with the path to an
// image on disk tells import to use that image when TMDB has no backdrop of its
// own. MReviews is the film's mreviews taxonomy tag, exactly as Hugo publishes
// it -- not derived or folded, just carried through, so import can key a film's
// tag-alias file on it directly. No json tags here, so the keys are the Go
// field names.

type FilmOut struct {
	LinkTitle    string
	MReviews     string
	ID           int
	ShowType     string
	Overview     string
	BackdropPath string
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

// TermPage is the minimal shape of a Hugo mreviews term page's own
// <URLPath>/index.json (rendered by themes/guild/layouts/mreviews/term.json).
// Metadata.LinkTitle is the LinkTitle of the first review page carrying that
// tag -- the same page tag-to-title.html resolves to -- so reading it recovers
// Hugo's own answer to "what film does this tag really mean" rather than
// re-deriving it.
type TermPage struct {
	Metadata struct {
		LinkTitle string `json:"LinkTitle"`
	} `json:"Metadata"`
}
