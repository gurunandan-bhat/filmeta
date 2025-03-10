/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"filmeta/config"
	"filmeta/model"
	"filmeta/tmdb"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save -o output-dir",
	Short: "A brief description of your command",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		id := args[0]
		filmID, err := strconv.Atoi(id)
		if err != nil || filmID == 0 {
			return fmt.Errorf("arg must be a non-zero integer")
		}

		tv, err := cmd.Flags().GetBool("tv")
		if err != nil {
			return err
		}
		showType := "movie"
		if tv {
			showType = "tv"
		}

		outPath, err := cmd.Flags().GetString("output-dir")
		if err != nil {
			return err
		}
		outPath, err = mkAbsPath(outPath)
		if err != nil {
			return err
		}

		cfg, err := config.Configuration()
		if err != nil {
			return err
		}

		client := tmdb.NewClient(cfg.TMDB.APIKey)
		film, err := client.Film(context.Background(), showType, filmID)
		if err != nil {
			return err
		}

		fName, err := cmd.Flags().GetString("fcg-name")
		if err != nil {
			return fmt.Errorf("error fetching fcg-name: %w", err)
		}
		if fName == "" {
			fName = film.Title
		}
		film.FCGTitle = fName
		fName = fmt.Sprintf("%x.json", md5.Sum([]byte(fName)))
		oFile := filepath.Join(outPath, fName)

		jsonBytes, err := json.MarshalIndent(film, "", "\t")
		if err != nil {
			return fmt.Errorf("error marshaling film: %w", err)
		}
		if err := os.WriteFile(oFile, jsonBytes, 0644); err != nil {
			return fmt.Errorf("error writing json to file %s: %w", oFile, err)
		}

		model, err := model.NewModel(cfg)
		if err != nil {
			return err
		}

		if err := model.Save(film, showType); err != nil {
			// Transaction Failed - delete file
			if errDel := os.Remove(oFile); errDel != nil {
				return fmt.Errorf("error deleting file %s after transaction rolled back with error %s: %w", oFile, err, errDel)
			}
			return fmt.Errorf("error saving data to db: %w", err)
		}

		posterOutPath := filepath.Join(outPath, "posters")
		if err := os.Mkdir(posterOutPath, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("error creating directory %s: %w", posterOutPath, err)
		}
		bdropOutPath := filepath.Join(outPath, "backdrops")
		if err := os.Mkdir(bdropOutPath, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("error creating directory %s: %w", bdropOutPath, err)
		}

		if film.PosterPath != "" {
			if err := client.TMDBImage(context.Background(), film.PosterPath, posterOutPath); err != nil {
				return fmt.Errorf("error fetching image: %w", err)
			}
		}

		if film.BackdropPath != "" {
			if err := client.TMDBImage(context.Background(), film.BackdropPath, bdropOutPath); err != nil {
				return fmt.Errorf("error fetching image: %w", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// saveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// saveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	saveCmd.Flags().BoolP("tv", "t", false, "This is a television serial")
	saveCmd.Flags().StringP("output-dir", "o", "", "Output directory to save JSON")
	saveCmd.Flags().StringP("fcg-name", "f", "", "Name of film in FCG to save JSON (uses md5(Title).json if none supplied)")

	cobra.MarkFlagRequired(saveCmd.Flags(), "output-dir")
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
