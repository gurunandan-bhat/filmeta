/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type FilmOut struct {
	LinkTitle string
	ID        int
	ShowType  string
	Overview  string
}

// missingMetaCmd represents the missingMeta command
var missingMetaCmd = &cobra.Command{
	Use:     "missingMeta",
	Aliases: []string{"missing-meta"},
	Short:   "Find films in content but not in db",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		inputPath := filepath.Join(metaCfg.HugoRoot, "mreviews/index.json")

		if len(args) > 0 {
			if args[0] != "" {
				inputPath = args[0]
			}
		}

		dataBytes, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("error reading datafile: %s: %w", inputPath, err)
		}

		var films []FilmOut
		if err := json.Unmarshal(dataBytes, &films); err != nil {
			return fmt.Errorf("error unmarshaling films slice: %w", err)
		}

		metaDir, err := cmd.Flags().GetString("meta-dir")
		if err != nil {
			return fmt.Errorf("error fetching metadata directory: %w", err)
		}
		if metaDir == "" {
			return errors.New("metadata directory cannot be empty")
		}

		missing := []FilmOut{}
		for _, film := range films {
			title := film.LinkTitle
			meta := fmt.Sprintf("%s/%x.json", metaDir, md5.Sum([]byte(title)))

			_, err := os.Stat(meta)
			if err != nil {
				if os.IsNotExist(err) {
					missing = append(missing, FilmOut{LinkTitle: title})
					continue
				}
				return fmt.Errorf("error finding file %s for %s: %w", meta, title, err)
			}
		}
		outStr := ""
		if len(missing) > 0 {
			missingOut, err := json.MarshalIndent(missing, "", "\t")
			if err != nil {
				return fmt.Errorf("error unmarshaling missing films: %w", err)
			}
			outStr = string(missingOut)
		}
		fmt.Println(outStr)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(missingMetaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// missingMetaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// missingMetaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	missingMetaCmd.Flags().StringP("meta-dir", "d", "/home/nandan/repos/guild/assets/meta", "Metadata directory for json files")
}
