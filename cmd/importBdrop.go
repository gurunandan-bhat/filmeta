/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"filmeta/config"
	"filmeta/model"
	"filmeta/tmdb"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// importBdropCmd represents the importBdrop command
var importBdropCmd = &cobra.Command{
	Use:     "importBdrop",
	Aliases: []string{"import-bdrop"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("importBdrop called")

		outPath, err := cmd.Flags().GetString("output-dir")
		if err != nil {
			return err
		}
		bdropOutPath := filepath.Join(outPath, "backdrops")
		bdropOutPath, err = mkAbsPath(bdropOutPath)
		if err != nil {
			return err
		}

		cfg, err := config.Configuration()
		if err != nil {
			return fmt.Errorf("error fetching configuration: %w", err)
		}
		model, err := model.NewModel(cfg)
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		qry := `SELECT IFNULL(vBackdropPath, "") FROM film where vBackdropPath is not null`
		var data []string
		if err := model.DbHandle.Select(&data, qry); err != nil {
			return fmt.Errorf("error fetching from database: %w", err)
		}

		client := tmdb.NewClient(cfg.TMDB.APIKey)
		for _, bPath := range data {
			if bPath != "" {
				if err := client.TMDBImage(context.Background(), bPath, bdropOutPath); err != nil {
					fmt.Println(err.Error())
				}
			}
			fmt.Printf("save %s, sleeping now\n", bPath)
			time.Sleep(2 * time.Second)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(importBdropCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importBdropCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importBdropCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	importBdropCmd.Flags().StringP("output-dir", "o", "", "Output directory to save JSON")
	cobra.MarkFlagRequired(importBdropCmd.Flags(), "output-dir")

}
