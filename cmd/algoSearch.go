/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/spf13/cobra"
)

// algoSearchCmd represents the algoSearch command
var algoSearchCmd = &cobra.Command{
	Use:     "algoSearch <search-term>",
	Aliases: []string{"algo-search"},
	Short:   "Search Algolia FCG Reviews index",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return errors.New("a search term is required")
		}

		term := args[0]
		fmt.Println("Term: ", term)

		appID := metaCfg.Algolia.AppID
		apiKey := metaCfg.Algolia.SearchKey
		indexName := "fcg_reviews"

		client, err := search.NewClient(appID, apiKey)
		if err != nil {
			panic(err)
		}

		// Search for 'test'
		searchResp, err := client.Search(
			client.NewApiSearchRequest(
				search.NewEmptySearchMethodParams().SetRequests(
					[]search.SearchQuery{
						*search.SearchForHitsAsSearchQuery(
							search.NewEmptySearchForHits().SetIndexName(indexName).SetQuery(term),
						),
					},
				),
			),
		)
		if err != nil {
			return fmt.Errorf("error searching for %s: %w", term, err)
		}

		jsonBytes, err := json.MarshalIndent(searchResp.Results, "", "\t")
		if err != nil {
			return fmt.Errorf("error marshaling search results: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(algoSearchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// algoSearchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// algoSearchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
