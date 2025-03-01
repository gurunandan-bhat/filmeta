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

// filmCmd represents the film command
var filmCmd = &cobra.Command{
	Use:   "film",
	Short: "Fetch film info with credits",
	RunE: func(cmd *cobra.Command, args []string) error {

		filmID, err := cmd.Flags().GetInt("film-id")
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
		cfg, err := config.Configuration()
		if err != nil {
			return err
		}
		client := tmdb.NewClient(cfg.TMDB.APIKey)
		film, err := client.Film(context.Background(), showType, filmID)
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
	rootCmd.AddCommand(filmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// filmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// filmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	filmCmd.Flags().IntP("film-id", "i", 0, "TMDB id of film to display film info")
	filmCmd.Flags().BoolP("tv", "t", false, "this is a tv show")
	cobra.MarkFlagRequired(filmCmd.Flags(), "film-id")
}
