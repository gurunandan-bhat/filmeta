/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

// dedupMetaCmd represents the dedupMeta command
var dedupMetaCmd = &cobra.Command{
	Use:     "dedupMeta",
	Aliases: []string{"dedup-meta"},
	Short:   "Deletes any duplicate metadata files in '/assets/meta'",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("dedupMeta called")

		md := goldmark.New(
			goldmark.WithExtensions(
				&frontmatter.Extender{},
			),
		)
		ctx := parser.NewContext()

		var data struct {
			Title    string   `toml:"title"`
			MReviews []string `toml:"mreviews"`
		}

		// Collect all json filenames in "/assets/meta"
		metaPath := filepath.Clean(fmt.Sprintf("%s/../assets/meta/*.json", metaCfg.HugoRoot))
		metaFPaths, err := filepath.Glob(metaPath)
		if err != nil {
			return err
		}
		metaMap := make(map[string]struct{})
		for _, p := range metaFPaths {
			metaMap[filepath.Base(p)] = struct{}{}
		}

		reviewsPath := filepath.Clean(fmt.Sprintf("%s/../content/reviews/*.md", metaCfg.HugoRoot))
		reviews, err := filepath.Glob(reviewsPath)
		if err != nil {
			return err
		}
		for _, fName := range reviews {
			review, err := os.ReadFile(fName)
			if err != nil {
				return err
			}
			if err := md.Convert(review, io.Discard, parser.WithContext(ctx)); err != nil {
				return err
			}

			fm := frontmatter.Get(ctx)
			if fm == nil {
				return fmt.Errorf("No front matter found in %s", fName)
			}
			if err := fm.Decode(&data); err != nil {
				return err
			}

			titleHash := md5.Sum([]byte(data.Title))
			mTagHash := md5.Sum([]byte(data.MReviews[0]))

			titleFName := fmt.Sprintf("%s.json", hex.EncodeToString(titleHash[:]))
			mTagFName := fmt.Sprintf("%s.json", hex.EncodeToString(mTagHash[:]))

			if _, ok := metaMap[titleFName]; ok {
				delete(metaMap, titleFName)
				fmt.Println("Removing ", titleFName)
			}

			if _, ok := metaMap[mTagFName]; ok {
				delete(metaMap, mTagFName)
				fmt.Println("Removing ", mTagFName)
			}
		}
		for fName := range metaMap {
			toDelete := filepath.Clean(fmt.Sprintf("%s/../assets/meta/%s", metaCfg.HugoRoot, fName))
			if err := os.Remove(toDelete); err != nil {
				return err
			}
		}

		fmt.Println(len(metaMap))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dedupMetaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dedupMetaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dedupMetaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
