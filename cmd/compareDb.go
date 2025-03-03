/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"filmeta/config"
	"filmeta/model"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// compareDbCmd represents the compareDb command
var compareDbCmd = &cobra.Command{
	Use:     "compareDb",
	Aliases: []string{"compare-db"},
	Short:   "Find films in content but not in db",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		inputPath := "/Users/nandan/repos/fcg/public/mreviews/index.json"
		if len(args) > 0 {
			if args[0] != "" {
				inputPath = args[0]
			}
		}

		dataBytes, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("error reading datafile: %s: %w", inputPath, err)
		}

		var films []FCGFilm
		if err := json.Unmarshal(dataBytes, &films); err != nil {
			return fmt.Errorf("error unmarshaling films")
		}

		cfg, err := config.Configuration()
		if err != nil {
			return err
		}

		model, err := model.NewModel(cfg)
		if err != nil {
			return err
		}

		notFound := []FCGFilm{}
		for _, film := range films {
			found, err := model.GetIDByTitle(film.Title)
			if err != nil {
				return fmt.Errorf("error finding title match: %w", err)
			}
			if len(found) == 0 {
				notFound = append(notFound, film)
			}
		}
		jsonBytes, err := json.MarshalIndent(notFound, "", "\t")
		if err != nil {
			return fmt.Errorf("error marshaling films not found: %w", err)
		}
		fmt.Println(string(jsonBytes))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(compareDbCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compareDbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compareDbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
