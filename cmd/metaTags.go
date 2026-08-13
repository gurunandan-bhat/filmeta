/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/md5"
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

		resolved, alreadyPresent, unresolved, err := resolveTagsFromReviews(metaCfg.HugoRoot, metaDir)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "review-resolved aliases: %d created, %d already present, %d unresolved (need a real import)\n", resolved, alreadyPresent, unresolved)

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

// resolveTagsFromReviews reads every entry in mreviews/index.json and, for any
// tag with no metadata of its own, resolves the real title behind it the same
// way Hugo's tag-to-title.html partial does: via that tag's own term page,
// whose Metadata.LinkTitle is the LinkTitle of the first review page carrying
// the tag. Hugo has already computed and published that answer, so this reads
// it rather than re-deriving it -- no need to replicate Hugo's page-ordering
// rules to pick the same review Hugo would when more than one carries the tag.
//
// A tag that resolves to itself, or to a title with no metadata of its own,
// is a genuine missingMeta gap rather than an aliasing job -- reported to
// stderr and counted as unresolved rather than silently skipped.
func resolveTagsFromReviews(hugoRoot, metaDir string) (created, alreadyPresent, unresolved int, err error) {

	mreviewsPath := filepath.Join(hugoRoot, "mreviews", "index.json")
	dataBytes, err := os.ReadFile(mreviewsPath)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error reading %s: %w", mreviewsPath, err)
	}

	var films map[string]Film
	if err := json.Unmarshal(dataBytes, &films); err != nil {
		return 0, 0, 0, fmt.Errorf("error unmarshaling %s: %w", mreviewsPath, err)
	}

	for _, film := range films {
		tag := film.LinkTitle
		if tag == "" {
			continue
		}

		exists, err := metaFileExists(metaDir, tag)
		if err != nil {
			return created, alreadyPresent, unresolved, err
		}
		if exists {
			alreadyPresent++
			continue
		}

		realTitle, err := resolveRealTitle(hugoRoot, film.URLPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unresolved: %q: %v\n", tag, err)
			unresolved++
			continue
		}

		if realTitle == "" || realTitle == tag {
			fmt.Fprintf(os.Stderr, "unresolved: %q has no distinct review title to resolve to; needs a real import\n", tag)
			unresolved++
			continue
		}

		realFile := filepath.Join(metaDir, fmt.Sprintf("%x.json", md5.Sum([]byte(realTitle))))
		metaBytes, err := os.ReadFile(realFile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "unresolved: %q resolves to %q, which also has no metadata; needs a real import\n", tag, realTitle)
				unresolved++
				continue
			}
			return created, alreadyPresent, unresolved, fmt.Errorf("error reading %s: %w", realFile, err)
		}

		wrote, err := writeAliasAt(metaDir, tag, metaBytes)
		if err != nil {
			return created, alreadyPresent, unresolved, fmt.Errorf("error writing alias for %q: %w", tag, err)
		}
		if wrote {
			created++
		}
	}

	return created, alreadyPresent, unresolved, nil
}
