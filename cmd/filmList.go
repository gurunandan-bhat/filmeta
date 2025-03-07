/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type Review struct {
	Author string    `json:"author,omitempty"`
	Date   time.Time `json:"date,omitempty"`
	Film   string    `json:"film,omitempty"`
	Link   string    `json:"link,omitempty"`
	Media  string    `json:"media,omitempty"`
	Path   string    `json:"path,omitempty"`
}

type Film struct {
	Film    string   `json:"film,omitempty"`
	Reviews []Review `json:"reviews,omitempty"`
}

// filmListCmd represents the filmList command
var filmListCmd = &cobra.Command{
	Use:     "filmList",
	Aliases: []string{"film-list"},
	Args:    cobra.MaximumNArgs(1),
	Short:   "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("film List called")
		source := ""
		if len(args) == 0 {
			source = "/Users/nandan/repos/fcg/public/reviews/index.json"
		}
		jsonBytes, err := os.ReadFile(source)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", source, err)
		}
		var films []Film
		if err := json.Unmarshal(jsonBytes, &films); err != nil {
			return fmt.Errorf("error un-marshaling films: %w", err)
		}

		w := csv.NewWriter(os.Stdout)
		for _, film := range films {
			for _, review := range film.Reviews {
				row := []string{review.Author, review.Film, review.Date.Format("Mon Jan 2 15:04:05 MST 2006")}
				if err := w.Write(row); err != nil {
					return fmt.Errorf("error writing %v: %w", row, err)
				}
			}
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(filmListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// filmListCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// filmListCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
