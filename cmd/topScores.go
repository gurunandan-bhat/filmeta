/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
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

// topScoresCmd represents the topScores command
var topScoresCmd = &cobra.Command{
	Use:     "topScores",
	Aliases: []string{"top-scores"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("topScores called")

		fromDate, err := cmd.Flags().GetTime("from-date")
		if err != nil {
			return fmt.Errorf("error reading from date: %w", err)
		}
		toDate, err := cmd.Flags().GetTime("to-date")
		if err != nil {
			return fmt.Errorf("error reading to date: %w", err)
		}
		fmt.Println("From and To:", fromDate.Local().Format("2006-01-02 15:04:05"), toDate.Local().Format("2006-01-02 15:04:05"))

		// Now read the list of films
		films := make(map[string]Film)
		filmsFName := fmt.Sprintf("%s/mreviews/index.json", metaCfg.HugoRoot)
		jsonBytes, err := os.ReadFile(filmsFName)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", filmsFName, err)
		}
		if err := json.Unmarshal(jsonBytes, &films); err != nil {
			return fmt.Errorf("error un-marshaling list of films: %w", err)
		}

		scores := make([][]string, 0)
		for _, film := range films {

			if film.Lastmod.Before(fromDate) || film.Lastmod.After(toDate) {
				continue
			}

			scores = append(scores, []string{film.LinkTitle, film.Language, film.Lastmod.Format("2006-01-02"), fmt.Sprintf("%.1f", film.AverageScore)})
		}

		w := csv.NewWriter(os.Stdout)
		if err := w.WriteAll(scores); err != nil {
			return fmt.Errorf("error writing csv: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(topScoresCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topScoresCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// topScoresCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	now := time.Now()
	from := now.AddDate(-1, 0, 0)

	topScoresCmd.Flags().TimeP("from-date", "f", from, []string{"2006-01-02"}, "Start Date")
	topScoresCmd.Flags().TimeP("to-date", "t", now, []string{"2006-01-02"}, "End Date")

}
