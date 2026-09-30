/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/spf13/cobra"
)

// newSearchCmd represents the newSearch command
var newSearchCmd = &cobra.Command{
	Use:     "newSearch <search-term>",
	Aliases: []string{"new-search"},
	Short:   "Search Algolia FCG Reviews index",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return errors.New("a search term is required")
		}

		term := args[0]
		appID := metaCfg.Algolia.AppID
		apiKey := metaCfg.Algolia.SearchKey
		indexName := "fcg-index"

		client, err := search.NewClient(appID, apiKey)
		if err != nil {
			panic(err)
		}

		paramsObj := search.NewSearchParamsObject().
			SetQuery(term).
			SetPage(3)
		params := search.SearchParamsObjectAsSearchParams(paramsObj)

		request := client.NewApiSearchSingleIndexRequest(indexName).
			WithSearchParams(params)

		searchResp, err := client.SearchSingleIndex(request)
		if err != nil {
			return fmt.Errorf("error searching for %s: %w", term, err)
		}

		jsonBytes, err := json.MarshalIndent(searchResp, "", "\t")
		if err != nil {
			return fmt.Errorf("error marshaling search results: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newSearchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newSearchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newSearchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
