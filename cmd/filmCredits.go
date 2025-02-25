/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"filmeta/config"
	"filmeta/tmdb"
	"fmt"

	"github.com/spf13/cobra"
)

// filmCreditsCmd represents the filmCredits command
var filmCreditsCmd = &cobra.Command{
	Use:   "filmCredits",
	Short: "Fetch film with credits",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("filmCredits called")

		filmID, err := cmd.Flags().GetInt("film-id")
		if err != nil {
			return err
		}

		cfg, err := config.Configuration()
		if err != nil {
			return err
		}
		client := tmdb.NewClient(cfg.TMDB.APIKey)
		film, err := client.FilmWithCredits(context.Background(), filmID)
		if err != nil {
			return err
		}

		jsonBytes, err := json.MarshalIndent(&film, "", "\t")
		if err != nil {
			return err
		}

		fmt.Println(string(jsonBytes))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(filmCreditsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// filmCreditsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// filmCreditsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	filmCreditsCmd.Flags().IntP("film-id", "i", 0, "TMDB id of film to filmCredits")
	cobra.MarkFlagRequired(filmCreditsCmd.Flags(), "film-id")
}
