/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"filmeta/tmdb"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// importCmd represents the bulkAdd command
var importCmd = &cobra.Command{
	Use:   "import input.json",
	Short: "Generate file metadata from info in data file",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), dataIsAvailable),
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Println("import called")
		outPath, err := cmd.Flags().GetString("output-dir")
		if err != nil {
			return err
		}
		outPath, err = mkAbsPath(outPath)
		if err != nil {
			return err
		}

		// Import data
		dataFile := args[0]
		jsonBytes, err := os.ReadFile(dataFile)
		if err != nil {
			return fmt.Errorf("error reading data file %s: %w", dataFile, err)
		}

		var inData []FilmOut
		if err := json.Unmarshal(jsonBytes, &inData); err != nil {
			return fmt.Errorf("error un-marshaling fcg data: %w", err)
		}

		client := tmdb.NewClient(metaCfg.TMDB.APIKey)
		posterOutPath := filepath.Join(outPath, "posters")
		if err := os.Mkdir(posterOutPath, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("error creating posters directory %s: %w", posterOutPath, err)
		}
		bdropOutPath := filepath.Join(outPath, "backdrops")
		if err := os.Mkdir(bdropOutPath, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("error creating backdrops directory %s: %w", bdropOutPath, err)
		}

		for _, film := range inData {

			filmID := film.ID
			if filmID == 0 {
				continue
			}
			showType := "movie"
			if film.ShowType == "tv" {
				showType = "tv"
			}

			tmdbFilm, err := client.Film(context.Background(), showType, filmID)
			if err != nil {
				return err
			}
			tmdbFilm.FCGTitle = film.LinkTitle
			if film.Overview != "" {
				tmdbFilm.Overview = film.Overview
			}

			fName := fmt.Sprintf("%x.json", md5.Sum([]byte(film.LinkTitle)))
			oFileName := filepath.Join(outPath, fName)
			jsonBytes, err := json.MarshalIndent(tmdbFilm, "", "\t")
			if err != nil {
				return fmt.Errorf("error marshaling film %s: %w", film.LinkTitle, err)
			}
			if err := os.WriteFile(oFileName, jsonBytes, 0644); err != nil {
				return fmt.Errorf("error writing json to file %s: %w", oFileName, err)
			}

			if err := metaModel.Save(tmdbFilm, showType); err != nil {
				// Transaction Failed - delete file
				if errDel := os.Remove(oFileName); errDel != nil {
					return fmt.Errorf("error deleting file %s after transaction rolled back with error %s: %w", oFileName, err, errDel)
				}
				return fmt.Errorf("error saving data to db: %w", err)
			}

			if tmdbFilm.PosterPath != "" {
				destPath := filepath.Join(posterOutPath, tmdbFilm.PosterPath)
				if err := client.TMDBImage(context.Background(), metaCfg.TMDB.PosterBase, tmdbFilm.PosterPath, destPath); err != nil {
					fmt.Printf("error fetching poster: %q", err)
				}
			}
			if tmdbFilm.BackdropPath != "" {
				destPath := filepath.Join(bdropOutPath, tmdbFilm.BackdropPath)
				if err := client.TMDBImage(context.Background(), metaCfg.TMDB.BackdropBase, tmdbFilm.BackdropPath, destPath); err != nil {
					fmt.Printf("error fetching backdrop: %q", err)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	importCmd.Flags().StringP("output-dir", "o", "", "Output directory to save JSON")

	if err := cobra.MarkFlagRequired(importCmd.Flags(), "output-dir"); err != nil {
		log.Fatalf("error requiring mandatory flag %s", "output-dir")
	}
}

func dataIsAvailable(cmd *cobra.Command, args []string) error {

	dataFile := args[0]
	if dataFile == "" {
		return fmt.Errorf("data file must exist and be readable")
	}

	info, err := os.Stat(dataFile)
	if err != nil {
		return fmt.Errorf("data file error: %w", err)
	}
	if info.Mode().Perm()&0444 != 0444 {
		return fmt.Errorf("data file %s is not readable", dataFile)
	}

	return nil
}

func mkAbsPath(path string) (string, error) {

	if path == "" {
		return path, nil
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				return "", fmt.Errorf("error creating output directory %s: %w", path, err)
			}
		} else {
			return "", err
		}
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("error converting %s to absolute path: %w", path, err)
	}

	return path, nil

}
