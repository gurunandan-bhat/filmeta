/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/md5"
	"encoding/json"
	"filmeta/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type Film struct {
	ObjectID     string `json:"objectID"`
	LinkTitle    string
	AverageScore float64 `json:"AverageScore"`
	URLPath      string
	Genres       string
	Language     string
	Overview     string
	Cast         string
	Director     string
	Poster       string
	Reviewers    string
}

type FilmReview struct {
	Critic string
}
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

// algoliaFilmsCmd represents the algoliaFilms command
var algoFilmsCmd = &cobra.Command{
	Use:     "algoFilms",
	Aliases: []string{"algo-films"},
	Short:   "Bulk creation of films for input to algolia",
	RunE: func(cmd *cobra.Command, args []string) error {

		filmsFName := filepath.Join(metaCfg.HugoRoot, "mreviews/index.json")
		filmsF, err := os.Open(filmsFName)
		if err != nil {
			return fmt.Errorf("error opening %s: %w", filmsFName, err)
		}

		decoder := json.NewDecoder(filmsF)
		films := []Film{}
		if err := decoder.Decode(&films); err != nil {
			return fmt.Errorf("error decoding films list: %w", err)
		}
		defer func() {
			if err := filmsF.Close(); err != nil {
				fmt.Printf("error closing films file: %s: %s", filmsFName, err)
			}
		}()

		algoFilms := []Film{}
		for _, f := range films {

			objectID := fmt.Sprintf("%x", md5.Sum([]byte(f.LinkTitle)))
			assetFName := filepath.Clean(fmt.Sprintf("%s/../assets/meta/%s.json", metaCfg.HugoRoot, objectID))
			assetF, err := os.Open(assetFName)
			if err != nil {
				//				fmt.Printf("error opening asset file %s film %s: %s", assetFName, f.LinkTitle, err)
				continue
			}

			meta := Meta{}
			asstDecoder := json.NewDecoder(assetF)
			if err := asstDecoder.Decode(&meta); err != nil {
				return fmt.Errorf("error unmarshaling meta data: %w", err)
			}
			castList := make([]string, len(meta.Credits.Cast))
			for i, p := range meta.Credits.Cast {
				castList[i] = p.Name
			}
			crewList := []string{}
			for _, p := range meta.Credits.Crew {
				if p.Job == "Director" {
					crewList = append(crewList, p.Name)
				}
			}
			genreList := make([]string, len(meta.Genres))
			for i, g := range meta.Genres {
				genreList[i] = g.Name
			}

			lang, err := config.ISOLanguage(meta.Language)
			if err != nil {
				return err
			}

			f.ObjectID = objectID
			f.Language = lang
			f.Overview = meta.Overview
			f.Cast = strings.Join(castList, ", ")
			f.Director = strings.Join(crewList, ", ")
			f.Poster = meta.PosterPath
			f.Genres = strings.Join(genreList, ", ")

			revFName := filepath.Join(metaCfg.HugoRoot, f.URLPath, "index.json")
			revF, err := os.Open(revFName)
			if err != nil {
				return fmt.Errorf("error opening reviews file %s: %w", revFName, err)
			}

			revs := []FilmReview{}
			revDecoder := json.NewDecoder(revF)
			if err := revDecoder.Decode(&revs); err != nil {
				return fmt.Errorf("error decoding %s: %w", revFName, err)
			}
			critics := make([]string, len(revs))
			for i, r := range revs {
				critics[i] = r.Critic
			}
			f.Reviewers = strings.Join(critics, ", ")

			algoFilms = append(algoFilms, f)

			if err := assetF.Close(); err != nil {
				fmt.Printf("error closing metadata file %s: %s", assetFName, err)
			}

		}

		jsonBytes, err := json.Marshal(algoFilms)
		if err != nil {
			return fmt.Errorf("error marshaling json: %w", err)
		}
		fmt.Println(string(jsonBytes))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(algoFilmsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// algoliaFilmsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// algoliaFilmsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
