/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// metaTagsCmd represents the metaTags command
var metaTagsCmd = &cobra.Command{
	Use:     "metaTags",
	Aliases: []string{"meta-tags"},
	Short:   "Back-fill tag-keyed alias copies of existing film metadata",
	RunE: func(cmd *cobra.Command, args []string) error {

		metaDir, err := cmd.Flags().GetString("meta-dir")
		if err != nil {
			return fmt.Errorf("error fetching metadata directory: %w", err)
		}
		if metaDir == "" {
			metaDir = filepath.Join(metaCfg.HugoRoot, "../assets", "meta")
		}

		created, unchanged, noTitle, err := backfillTagAliases(metaDir)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "tag aliases: %d created, %d unchanged (title already canonical), %d skipped (no fcg_title)\n", created, unchanged, noTitle)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(metaTagsCmd)

	metaTagsCmd.Flags().StringP("meta-dir", "d", "", "Metadata directory for json files (default: <hugo-root>/assets/meta)")
}

// backfillTagAliases scans every *.json file in metaDir and, for each one,
// writes a tag-keyed alias copy via writeTagAlias using the file's own
// fcg_title. Re-running it is a no-op for films already aliased: writeTagAlias
// writes the tag-copy file to itself unchanged.
func backfillTagAliases(metaDir string) (created, unchanged, noTitle int, err error) {

	entries, err := os.ReadDir(metaDir)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error reading metadata directory %s: %w", metaDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(metaDir, entry.Name())
		jsonBytes, err := os.ReadFile(path)
		if err != nil {
			return created, unchanged, noTitle, fmt.Errorf("error reading %s: %w", path, err)
		}

		var probe struct {
			FCGTitle string `json:"fcg_title"`
		}
		if err := json.Unmarshal(jsonBytes, &probe); err != nil {
			return created, unchanged, noTitle, fmt.Errorf("error unmarshaling %s: %w", path, err)
		}
		if probe.FCGTitle == "" {
			fmt.Fprintf(os.Stderr, "skipping %s: no fcg_title\n", entry.Name())
			noTitle++
			continue
		}

		wrote, err := writeTagAlias(metaDir, probe.FCGTitle, jsonBytes)
		if err != nil {
			return created, unchanged, noTitle, fmt.Errorf("error writing tag alias for %s: %w", probe.FCGTitle, err)
		}
		if wrote {
			created++
		} else {
			unchanged++
		}
	}

	return created, unchanged, noTitle, nil
}
