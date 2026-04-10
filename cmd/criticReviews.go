/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type Guild struct {
	Name          string   `json:"LinkTitle,omitempty"`
	ReviewURL     string   `json:"ReviewURL,omitempty"`
	Organizations []string `json:"Organizations,omitempty"`
}
type CriticReview struct {
	Publication string
	PublishDate time.Time
}

// criticReviewsCmd represents the criticReviews command
var criticReviewsCmd = &cobra.Command{
	Use:     "criticReviews",
	Short:   "Review count of Critics",
	Aliases: []string{"critic-reviews"},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("criticReviews called")

		fromDate, err := cmd.Flags().GetTime("from-date")
		if err != nil {
			return fmt.Errorf("error reading from date: %w", err)
		}
		toDate, err := cmd.Flags().GetTime("to-date")
		if err != nil {
			return fmt.Errorf("error reading to date: %w", err)
		}
		fmt.Println("From and To:", fromDate.Local().Format("2006-01-02 15:04:05"), toDate.Local().Format("2006-01-02 15:04:05"))

		// First read all members
		guildFName := fmt.Sprintf("%s/guild/index.json", metaCfg.HugoRoot)
		jsonBytes, err := os.ReadFile(guildFName)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", guildFName, err)
		}
		var guildmembers = []Guild{}
		if err := json.Unmarshal(jsonBytes, &guildmembers); err != nil {
			return fmt.Errorf("error un-marshaling guild members: %w", err)
		}

		// result := make([]map[string]int, len(guildmembers))
		critics := make([][]string, 0)
		for _, member := range guildmembers {

			orgMap := make(map[string]int)
			for _, org := range member.Organizations {
				orgMap[strings.Trim(org, " ")] = 1
			}

			criticPath := filepath.Join(metaCfg.HugoRoot, "critics", member.ReviewURL, "index.json")
			fh, err := os.Open(criticPath)
			var reviewCount int
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("error opening file for %s: %w", member.Name, err)
				}
				critics = append(critics, []string{member.Name, member.Organizations[0], strconv.Itoa(reviewCount)})
				continue
			}
			reviewCount, err = processCritic(fh, orgMap, fromDate, toDate)
			if err != nil {
				return fmt.Errorf("error copunting reviews for %s: %w", member.Name, err)
			}
			if err := fh.Close(); err != nil {
				return fmt.Errorf("error closing %s: %w", criticPath, err)
			}

			orgs := make([]string, len(orgMap))
			i := 0
			for key := range orgMap {
				{
					orgs[i] = key
					i++
				}
			}
			critics = append(critics, []string{member.Name, strings.Join(orgs, ", "), strconv.Itoa(reviewCount)})
		}

		w := csv.NewWriter(os.Stdout)
		if err := w.WriteAll(critics); err != nil {
			return fmt.Errorf("error writing csv: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(criticReviewsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// criticReviewsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	now := time.Now()
	from := now.AddDate(-1, 0, 0)

	criticReviewsCmd.Flags().TimeP("from-date", "f", from, []string{"2006-01-02"}, "Start Date")
	criticReviewsCmd.Flags().TimeP("to-date", "t", now, []string{"2006-01-02"}, "End Date")
}

func processCritic(fh *os.File, orgMap map[string]int, fromDate, toDate time.Time) (int, error) {

	reviews := []CriticReview{}
	if err := json.NewDecoder(fh).Decode(&reviews); err != nil {
		return 0, fmt.Errorf("error decoding review: %w", err)
	}

	reviewCount := 0
	for _, review := range reviews {
		if fromDate.Before(review.PublishDate) && toDate.After(review.PublishDate) {
			reviewCount = reviewCount + 1
			if review.Publication != "" {
				orgMap[strings.Trim(review.Publication, " ")] = 1
			}
		}
	}

	return reviewCount, nil
}
