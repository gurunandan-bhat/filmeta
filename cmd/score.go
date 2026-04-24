/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var film, critic string
var defaultScoreFile = "/../data/freescores.json"

// scoreCmd represents the score command
var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "A brief description of your command",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {

		var err error
		critic, err = cmd.Flags().GetString("critic")
		if err != nil {
			return fmt.Errorf("error reading critic flag: %w", err)
		}
		film, err = cmd.Flags().GetString("film")
		if err != nil {
			return fmt.Errorf("error reading film flag: %w", err)
		}
		if err := validate(critic, film); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		score, parseErr := strconv.ParseFloat(strings.TrimSpace(args[0]), 64)
		if parseErr != nil || score == 0.0 {
			return fmt.Errorf("score must be a non-zero float: got %s instead", args[0])
		}

		outPath, err := cmd.Flags().GetString("outPath")
		if err != nil {
			return fmt.Errorf("error reading outPath flag: %w", err)
		}
		if outPath == "" {
			outPath = defaultScoreFile
		}

		outPath = filepath.Join(metaCfg.HugoRoot, outPath)
		filmScores := FreeScores(make(map[string]Scores))

		jsonBytes, err := os.ReadFile(outPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error reading score file: %w", err)
		}
		if len(jsonBytes) == 0 {
			jsonBytes = []byte("{}")
		}

		if err := json.Unmarshal(jsonBytes, &filmScores); err != nil {
			return fmt.Errorf("error unmarshaling scores: %w", err)
		}

		if err := filmScores.update(film, critic, score); err != nil {
			return err
		}

		jsonBytes, err = json.MarshalIndent(filmScores, "", "\t")
		if err != nil {
			return fmt.Errorf("error marshaling scores: %w", err)
		}

		if err := os.WriteFile(outPath, jsonBytes, 0644); err != nil {
			return fmt.Errorf("error writing score file: %w", err)
		}

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
	scoreCmd.Flags().StringP("critic", "a", "", "reviewer assigning score")
	scoreCmd.Flags().StringP("film", "f", "", "film to assign score")
	scoreCmd.Flags().StringP("outPath", "o", "", "score data file")

	if err := cobra.MarkFlagRequired(scoreCmd.Flags(), "critic"); err != nil {
		panic(fmt.Sprintf("error requiring mandatory flag critic: %v", err))
	}
	if err := cobra.MarkFlagRequired(scoreCmd.Flags(), "film"); err != nil {
		panic(fmt.Sprintf("error requiring mandatory flag film: %v", err))
	}
}

func validate(critic, film string) error {

	// 1. Check if author exists
	// 2. Check that we have at least one review
	// 3. Check that author has not reviewed film
	criticPath := entityExists(critic, "/critics")
	if criticPath == "" {
		return fmt.Errorf("no author matched %s", critic)
	}
	filmPath := entityExists(film, "/mreviews")
	if filmPath == "" {
		return fmt.Errorf("no film matched %s", film)
	}

	reviewPath := entityExists(film, criticPath)
	if reviewPath != "" {
		return fmt.Errorf("%s has reviewed %s", critic, film)
	}

	return nil
}

func entityExists(e, path string) string {

	jsonPath := fmt.Sprintf("%s%s/index.json", metaCfg.HugoRoot, path)
	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("Error reading %s: %v", jsonPath, err)
		return ""
	}
	entities := map[string]Entity{}
	if err := json.Unmarshal(jsonBytes, &entities); err != nil {
		fmt.Printf("Error unmarshaling %s: %v", jsonPath, err)
		return ""
	}

	eHash := md5.Sum([]byte(e))
	entity, ok := entities[hex.EncodeToString(eHash[:])]
	if !ok {
		return ""
	}

	return entity.Path
}

func (s FreeScores) update(film, critic string, score float64) error {

	scores, ok := s[film]
	if !ok {
		s[film] = Scores{critic: score}
		return nil
	}
	scores[critic] = score
	s[film] = scores

	return nil
}
