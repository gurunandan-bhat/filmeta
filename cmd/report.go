/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"filmeta/guild"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var mreviews = "/mreviews"

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {

		mFName := metaCfg.HugoRoot + mreviews + "/index.json"
		filmJSONBytes, err := os.ReadFile(mFName)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", mFName, err)
		}

		films := []guild.Film{}
		if err := json.Unmarshal(filmJSONBytes, &films); err != nil {
			return fmt.Errorf("error unmarshaling list of films: %w", err)
		}

		scores := [][]string{}
		for _, film := range films {
			scores = append(scores, []string{
				film.LinkTitle,
				strconv.FormatFloat(film.AverageScore, 'f', 1, 64),
			})
		}

		w := csv.NewWriter(os.Stdout)
		if err := w.WriteAll(scores); err != nil {
			return fmt.Errorf("error writing csv file: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
