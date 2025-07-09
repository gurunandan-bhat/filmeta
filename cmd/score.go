/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// scoreCmd represents the score command
var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "A brief description of your command",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {

		author, err := cmd.Flags().GetString("author")
		if err != nil {
			return fmt.Errorf("author not supplied: %w", err)
		}
		if err := entityExists(author, "critics"); err != nil {
			return err
		}

		film, err := cmd.Flags().GetString("film")
		if err != nil {
			return fmt.Errorf("film not supplied: %w", err)
		}
		if err := entityExists(film, "mreviews"); err != nil {
			return err
		}

		// We need to chack that athe film has not been reviewed by author

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		score, err := strconv.ParseFloat(strings.TrimSpace(args[0]), 64)
		if err != nil || score == 0.0 {
			return fmt.Errorf("score must be a non-zero float: got %s instead", args[0])
		}
		fmt.Println("score called with ", score)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	scoreCmd.Flags().StringP("author", "a", "", "reviewer assigning score")
	scoreCmd.Flags().StringP("film", "f", "", "film to assign score")

	cobra.MarkFlagRequired(scoreCmd.Flags(), "author")
	cobra.MarkFlagRequired(scoreCmd.Flags(), "film")
}

type Entity struct {
	LinkTitle string `json:"LinkTitle,omitempty"`
	Path      string `json:"Path,omitempty"`
}

func entityExists(e, kind string) error {

	jsonBytes, err := os.ReadFile(fmt.Sprintf("%s/%s/index.json", metaCfg.HugoRoot, kind))
	if err != nil {
		return fmt.Errorf("error reading critics json: %w", err)
	}
	entities := []Entity{}
	if err := json.Unmarshal(jsonBytes, &entities); err != nil {
		return fmt.Errorf("error unmarshaling authors: %w", err)
	}

	matched := false
	for _, a := range entities {
		if e == a.LinkTitle {
			matched = true
			break
		}
	}
	if !matched {
		return fmt.Errorf("%s not found in list of %s", e, kind)
	}

	return nil
}
