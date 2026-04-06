/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// createPostsCmd represents the createPosts command
var createPostsCmd = &cobra.Command{
	Use:     "createPosts",
	Short:   "Create test posts from IMDB IDs",
	Aliases: []string{"create-posts"},
	RunE: func(cmd *cobra.Command, args []string) error {

		idsFile, err := cmd.Flags().GetString("ids-file")
		if err != nil {
			return fmt.Errorf("error fetching file with TMDB IDs: %w", err)
		}
		fmt.Println("createPosts called with", idsFile)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(createPostsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createPostsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	createPostsCmd.Flags().StringP("ids-file", "i", "", "File listing TMDB IDs one per line")
	if err := cobra.MarkFlagRequired(createPostsCmd.Flags(), "ids-file"); err != nil {
		log.Fatalf("error requiring mandatory flag %s", "ids-file")
	}

}
