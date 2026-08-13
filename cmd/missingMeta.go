/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

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

		var films map[string]Film
		if err := json.Unmarshal(dataBytes, &films); err != nil {
			return fmt.Errorf("error unmarshaling films slice: %w", err)
		}

		metaDir, err := cmd.Flags().GetString("meta-dir")
		if err != nil {
			return fmt.Errorf("error fetching metadata directory: %w", err)
		}
		if metaDir == "" {
			metaDir = filepath.Join(metaCfg.HugoRoot, "../assets", "meta")
		}

		missing, err := findMissing(films, metaCfg.HugoRoot, metaDir)
		if err != nil {
			return err
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
	missingMetaCmd.Flags().StringP("meta-dir", "d", "", "Metadata directory for json files (default: <hugo-root>/assets/meta)")
}

// findMissing returns a FilmOut for every film with no metadata file under
// either its tag or its real title. MReviews is always the tag exactly as
// Hugo published it in mreviews/index.json. LinkTitle is the film's real
// title where it can be recovered -- via the tag's own term page, see
// resolveRealTitle -- and falls back to the tag itself, with a warning to
// stderr, when that resolution fails (the film's review content not yet
// built, or malformed).
//
// When title and tag differ, both hashes are checked before a film counts as
// missing: metadata may already exist under the real title (a normal import
// ran before the tag was corrected, or before this diverged at all) even
// though the tag itself has no file yet. That film only needs a tag alias --
// see metaTags -- not a fresh TMDB import, so it is not reported here.
func findMissing(films map[string]Film, hugoRoot, metaDir string) ([]FilmOut, error) {

	missing := []FilmOut{}
	for _, film := range films {
		tag := film.LinkTitle

		exists, err := metaFileExists(metaDir, tag)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}

		title := tag
		if realTitle, rErr := resolveRealTitle(hugoRoot, film.URLPath); rErr != nil {
			fmt.Fprintf(os.Stderr, "warning: %q: could not resolve real title, falling back to the tag: %v\n", tag, rErr)
		} else if realTitle != "" {
			title = realTitle
		}

		if title != tag {
			exists, err := metaFileExists(metaDir, title)
			if err != nil {
				return nil, err
			}
			if exists {
				continue
			}
		}

		missing = append(missing, FilmOut{LinkTitle: title, MReviews: tag})
	}

	return missing, nil
}
