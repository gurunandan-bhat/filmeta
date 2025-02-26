/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"filmeta/config"
	"filmeta/model"
	"filmeta/tmdb"

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
		cfg, err := config.Configuration()
		if err != nil {
			return err
		}

		client := tmdb.NewClient(cfg.TMDB.APIKey)
		film, err := client.Film(context.Background(), id)
		if err != nil {
			return err
		}

		model, err := model.NewModel(cfg)
		if err != nil {
			return err
		}

		if err := model.Save(film); err != nil {
			return err
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
	cobra.MarkFlagRequired(saveCmd.Flags(), "film-id")
}
