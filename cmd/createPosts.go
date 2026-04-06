/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createPostsCmd represents the createPosts command
var createPostsCmd = &cobra.Command{
	Use:     "createPosts",
	Short:   "Create test posts from IMDB IDs",
	Aliases: []string{"create-posts"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("createPosts called")
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
	// createPostsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
