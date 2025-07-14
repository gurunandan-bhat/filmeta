/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var film, critic string
var defaultOutDir = "/../data/freescores/"

type FilmScore struct {
	Critic string
	Score  float64
}

// scoreCmd represents the score command
var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "A brief description of your command",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {

		critic, _ = cmd.Flags().GetString("critic")
		film, _ = cmd.Flags().GetString("film")
		if err := validate(critic, film); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		score, err := strconv.ParseFloat(strings.TrimSpace(args[0]), 64)
		if err != nil || score == 0.0 {
			return fmt.Errorf("score must be a non-zero float: got %s instead", args[0])
		}

		h := md5.New()
		io.WriteString(h, film)
		outDir, _ := cmd.Flags().GetString("outDir")
		if outDir == "" {
			outDir = defaultOutDir
		}

		outDir = filepath.Clean(metaCfg.HugoRoot + outDir)
		scoreFName := filepath.Join(outDir, fmt.Sprintf("%x.json", h.Sum(nil)))
		scores := make(map[string]float64)

		scoreFile, err := os.OpenFile(scoreFName, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer scoreFile.Close()

		jsonBytes, err := io.ReadAll(scoreFile)
		if err != nil {
			return err
		}
		if len(jsonBytes) > 0 {
			if err := json.Unmarshal(jsonBytes, &scores); err != nil {
				return err
			}
		}

		scores[critic] = score
		jsonBytes, err = json.MarshalIndent(&scores, "", "\t")
		if err != nil {
			return err
		}
		fmt.Println(scores)

		if err := scoreFile.Truncate(0); err != nil {
			return fmt.Errorf("error truncating file: %w", err)
		}
		if _, err := scoreFile.Seek(0, 0); err != nil {
			return err
		}

		_, err = scoreFile.WriteString(string(jsonBytes))
		if err != nil {
			return err
		}

		fmt.Println("Filepath:", scoreFName)
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
	scoreCmd.Flags().StringP("critic", "c", "", "reviewer assigning score")
	scoreCmd.Flags().StringP("film", "f", "", "film to assign score")
	scoreCmd.Flags().StringP("outDir", "o", "", "score data dir")

	cobra.MarkFlagRequired(scoreCmd.Flags(), "critic")
	cobra.MarkFlagRequired(scoreCmd.Flags(), "film")
}

type Entity struct {
	LinkTitle string `json:"LinkTitle,omitempty"`
	Path      string `json:"Path,omitempty"`
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

	var ePath string
	jsonBytes, err := os.ReadFile(fmt.Sprintf("%s%s/index.json", metaCfg.HugoRoot, path))
	if err != nil {
		return ePath
	}
	entities := []Entity{}
	if err := json.Unmarshal(jsonBytes, &entities); err != nil {
		return ePath
	}

	for _, a := range entities {
		if e == a.LinkTitle {
			ePath = a.Path
			break
		}
	}

	return ePath
}
