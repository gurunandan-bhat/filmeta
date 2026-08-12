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
	"strings"

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

			// Whatever TMDB itself returned, captured before a local backdrop
			// can overwrite it, so the fetch below only ever asks TMDB for a
			// path TMDB actually gave us.
			tmdbBackdrop := tmdbFilm.BackdropPath

			// A film with no backdrop at TMDB can name a local image instead.
			// It is copied in beside the fetched ones and recorded in the
			// metadata, so this has to happen before the json is written.
			if tmdbBackdrop == "" && film.BackdropPath != "" {
				backdropPath, err := copyLocalBackdrop(film.BackdropPath, bdropOutPath)
				if err != nil {
					return fmt.Errorf("error using local backdrop for %s: %w", film.LinkTitle, err)
				}
				tmdbFilm.BackdropPath = backdropPath
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
			if _, err := writeTagAlias(outPath, film.LinkTitle, jsonBytes); err != nil {
				return fmt.Errorf("error writing tag alias for %s: %w", film.LinkTitle, err)
			}

			if tmdbFilm.PosterPath != "" {
				destPath := filepath.Join(posterOutPath, tmdbFilm.PosterPath)
				if err := client.TMDBImage(context.Background(), metaCfg.TMDB.PosterBase, tmdbFilm.PosterPath, destPath); err != nil {
					fmt.Printf("error fetching poster: %q", err)
				}
			}
			if tmdbBackdrop != "" {
				destPath := filepath.Join(bdropOutPath, tmdbBackdrop)
				if err := client.TMDBImage(context.Background(), metaCfg.TMDB.BackdropBase, tmdbBackdrop, destPath); err != nil {
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

// dataIsAvailable checks that the input file exists and can actually be read by
// this process. Permission bits alone cannot answer that -- whether a mode is
// sufficient depends on ownership, group membership and any ACLs -- so the file
// is opened rather than having its mode inspected.
func dataIsAvailable(cmd *cobra.Command, args []string) error {

	dataFile := args[0]
	if dataFile == "" {
		return fmt.Errorf("data file must exist and be readable")
	}

	fh, err := os.Open(dataFile)
	if err != nil {
		return fmt.Errorf("data file error: %w", err)
	}
	defer func() { _ = fh.Close() }()

	info, err := fh.Stat()
	if err != nil {
		return fmt.Errorf("data file error: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("data file %s is a directory", dataFile)
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

// copyLocalBackdrop copies an image from srcPath into destDir and returns the
// value to record as backdrop_path.
//
// The copy is named for the md5 of its own contents, keeping the convention the
// hand-added backdrops in the site already follow: content addressing means
// re-importing the same film rewrites the same bytes to the same name rather
// than accumulating copies, and it cannot collide with the paths TMDB hands out.
// The returned path carries the leading slash that TMDB's own values have,
// because Hugo resolves the asset as "meta/backdrops" + backdrop_path.
func copyLocalBackdrop(srcPath, destDir string) (string, error) {

	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext == "" {
		return "", fmt.Errorf("backdrop %s has no file extension", srcPath)
	}

	imgBytes, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("error reading backdrop %s: %w", srcPath, err)
	}
	if len(imgBytes) == 0 {
		return "", fmt.Errorf("backdrop %s is empty", srcPath)
	}

	fName := fmt.Sprintf("%x%s", md5.Sum(imgBytes), ext)
	destPath := filepath.Join(destDir, fName)
	if err := os.WriteFile(destPath, imgBytes, 0644); err != nil {
		return "", fmt.Errorf("error writing backdrop to %s: %w", destPath, err)
	}

	return "/" + fName, nil
}
