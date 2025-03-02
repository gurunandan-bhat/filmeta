/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"filmeta/config"
	"filmeta/model"
	"filmeta/tmdb"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {

		id, err := cmd.Flags().GetInt("film-id")
		if err != nil {
			return err
		}

		tv, err := cmd.Flags().GetBool("tv")
		if err != nil {
			return err
		}
		showType := "movie"
		if tv {
			showType = "tv"
		}

		outPath, err := cmd.Flags().GetString("output")
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
		film, err := client.Film(context.Background(), showType, id)
		if err != nil {
			return err
		}

		fName := fmt.Sprintf("%s-%d.json", showType, film.Id)
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
	saveCmd.Flags().IntP("film-id", "i", 0, "TMDB id of film to save")
	saveCmd.Flags().BoolP("tv", "t", false, "This is a television serial")
	saveCmd.Flags().StringP("output", "o", "", "Save data as JSON in output file")

	cobra.MarkFlagRequired(saveCmd.Flags(), "film-id")
}

func mkAbsPath(path string) (string, error) {

	if path == "" {
		return path, nil
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path, 0755); err != nil {
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
