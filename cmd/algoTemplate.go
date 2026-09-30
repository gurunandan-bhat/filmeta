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
	"github.com/kylelemons/godebug/pretty"
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

		type IDXOut struct {
			Page     int32
			NumPages int32
			Hits     []map[string]any
		}
		idxOut := make([]IDXOut, 0)
		for _, r := range result {
			sResp := r.SearchResponse
			hits := sResp.Hits
			page := sResp.Page
			numPages := sResp.NbPages
			allHits := make([]map[string]any, 0)
			for _, hit := range hits {

				hitMap := make(map[string]any)
				hitMap["LocalPosterPath"] = hit.AdditionalProperties["LocalPosterPath"]
				hitMap["URLPath"] = hit.AdditionalProperties["URLPath"]

				hHit := hit.GetHighlightResult()
				for key, val := range hHit {
					hitMap[key] = val.HighlightResultOption.Value
				}

				allHits = append(allHits, hitMap)
			}
			idxOut = append(idxOut, IDXOut{
				Page:     *page,
				NumPages: *numPages,
				Hits:     allHits,
			})
		}

		pretty.Print(idxOut)

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
