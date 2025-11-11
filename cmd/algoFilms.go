/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type Film struct {
	LinkTitle    string
	LastMod      string
	AverageScore float64
	Path         string
}

type FilmReview struct {
	Critic   string
	Media    string
	Source   string
	SubTitle string
	Content  string
	Path     string
}

type AlgoFilm struct {
	Film
	Reviews []FilmReview
}

// algoliaFilmsCmd represents the algoliaFilms command
var algoFilmsCmd = &cobra.Command{
	Use:     "algoFilms",
	Aliases: []string{"algo-films"},
	Short:   "Bulk creation of films for input to algolia",
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Println("algo-films called")
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

		algoFilms := []AlgoFilm{}
		fmt.Println(films)
		for _, f := range films {
			reviewsFName := filepath.Join(metaCfg.HugoRoot, f.Path, "index.json")
			reviewsF, err := os.Open(reviewsFName)
			if err != nil {
				return fmt.Errorf("error opening file %s: %w", reviewsFName, err)
			}

			reviews := []FilmReview{}
			decoder := json.NewDecoder(reviewsF)
			if err := decoder.Decode(&reviews); err != nil {
				return fmt.Errorf("error decoding json at %s: %w", f.Path, err)
			}

			algoFilms = append(algoFilms, AlgoFilm{
				Film:    f,
				Reviews: reviews,
			})

			if err := reviewsF.Close(); err != nil {
				fmt.Printf("error closing reviews file %s: %s", reviewsFName, err)
			}
		}

		fmt.Println(algoFilms)
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
