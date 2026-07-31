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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// criticReviewsCmd represents the criticReviews command
var criticReviewsCmd = &cobra.Command{
	Use:     "criticReviews",
	Short:   "Review count of Critics",
	Aliases: []string{"critic-reviews"},
	RunE: func(cmd *cobra.Command, args []string) error {
		start, end, err := dateRange(cmd)
		if err != nil {
			return err
		}
		// stdout carries the CSV, so progress goes to stderr
		fmt.Fprintln(os.Stderr, "From and To:", describeRange(start, end))

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
				if org = strings.TrimSpace(org); org != "" {
					orgMap[org] = 1
				}
			}

			criticPath := filepath.Join(metaCfg.HugoRoot, "critics", member.ReviewURL, "index.json")
			fh, err := os.Open(criticPath)
			var reviewCount int
			switch {
			case err == nil:
				reviewCount, err = processCritic(fh, orgMap, start, end)
				if err != nil {
					return fmt.Errorf("error counting reviews for %s: %w", member.Name, err)
				}
				if err := fh.Close(); err != nil {
					return fmt.Errorf("error closing %s: %w", criticPath, err)
				}
			case errors.Is(err, os.ErrNotExist):
				// The member has no review index yet, so they keep a zero count
				// and whatever organizations their guild entry names.
			default:
				return fmt.Errorf("error opening file for %s: %w", member.Name, err)
			}

			critics = append(critics, []string{member.Name, joinOrgs(orgMap), strconv.Itoa(reviewCount)})
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
	addDateRangeFlags(criticReviewsCmd)
}

// joinOrgs renders the organization set as a comma-separated string. Map
// iteration order is random, so the names are sorted to keep the report
// reproducible between runs.
func joinOrgs(orgMap map[string]int) string {

	orgs := make([]string, 0, len(orgMap))
	for org := range orgMap {
		orgs = append(orgs, org)
	}
	slices.Sort(orgs)

	return strings.Join(orgs, ", ")
}

func processCritic(fh *os.File, orgMap map[string]int, start, end time.Time) (int, error) {

	reviews := make(map[string]CriticReview, 0)
	if err := json.NewDecoder(fh).Decode(&reviews); err != nil {
		return 0, fmt.Errorf("error decoding review: %w", err)
	}

	reviewCount := 0
	for _, review := range reviews {
		if inRange(review.PublishDate, start, end) {
			reviewCount = reviewCount + 1
			if pub := strings.TrimSpace(review.Publication); pub != "" {
				orgMap[pub] = 1
			}
		}
	}

	return reviewCount, nil
}
