/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/spf13/cobra"
)

// algoTemplateCmd represents the algoTemplate command
var algoTemplateCmd = &cobra.Command{
	Use:     "algoTemplate",
	Aliases: []string{"algo-template"},
	Short:   "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) != 1 {
			return fmt.Errorf("incorrect number of argumants, want 1, found %d", len(args))
		}
		jsonFilePath := args[0]
		fmt.Println("algo template called with", jsonFilePath)

		jsonStr, err := os.ReadFile(jsonFilePath)
		if err != nil {
			// 1. Check if the file does not exist
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("Error: The file '%s' does not exist.\n", jsonFilePath)
			}

			// 2. Check if the error is due to permission denied
			if errors.Is(err, os.ErrPermission) {
				return fmt.Errorf("Error: Permission denied. You cannot read '%s'.\n", jsonFilePath)
			}

			// 3. Handle any other unexpected errors (e.g., hardware issues)
			return fmt.Errorf("An unexpected error occurred: %v\n", err)
		}

		// Success: The file exists and we have permissions to read it
		// fmt.Println("File read successfully Content:")
		// fmt.Println(string(jsonStr))

		result := []search.SearchResult{}
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			return fmt.Errorf("failed to de-serialize search result: %w", err)
		}
		firstIndexResult := result[0]
		fmt.Println(firstIndexResult.SearchResponse)

		// resp := result[0].SearchResponse
		// fmt.Println(*resp.NbHits, *resp.NbPages, *resp.Page, *resp.HitsPerPage)

		// for _, hit := range resp.GetHits() {

		// 	mHit := hit.GetHighlightResult()
		// 	for key, hilightResult := range mHit {
		// 		fmt.Println(key)
		// 		fmt.Println(hilightResult.HighlightResultOption.GetFullyHighlighted())
		// 		fmt.Printf("%s\n", strings.Repeat("*", 80))
		// 	}
		// 	fmt.Printf("%s\n", strings.Repeat("-", 80))
		// }

		return nil
	},
}

func init() {
	rootCmd.AddCommand(algoTemplateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// algoTemplateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// algoTemplateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
