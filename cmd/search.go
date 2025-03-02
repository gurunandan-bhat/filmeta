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
	"strconv"

	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search a film by title",
	RunE: func(cmd *cobra.Command, args []string) error {
		query, err := cmd.Flags().GetString("query")
		if err != nil {
			return err
		}
		language, err := cmd.Flags().GetString("language")
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
		opts := tmdb.SearchOptions{
			Query:        query,
			IncludeAdult: false,
			Language:     language,
			Page:         1,
		}
		year, err := cmd.Flags().GetInt("year")
		if err != nil {
			return err
		}
		if year > 0 {
			opts.Year = strconv.Itoa(year)
		}

		films, err := client.ShowSearch(context.Background(), showType, &opts)
		if err != nil {
			return err
		}

		jsonBytes, err := json.MarshalIndent(films, "", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonBytes))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	searchCmd.Flags().StringP("query", "q", "", "title to search for")
	searchCmd.Flags().StringP("language", "l", "en", "language of the output")
	searchCmd.Flags().BoolP("tv", "t", false, "search in television serials not movies")
	searchCmd.Flags().IntP("year", "y", 0, "year of release")

	cobra.MarkFlagRequired(searchCmd.Flags(), "query")
}
