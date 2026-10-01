package cmd

import (
	"context"
	"fmt"
	"html/template"
	"maps"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/kylelemons/godebug/pretty"
	"github.com/spf13/cobra"
)

var (
	searchPage  int32
	searchIndex string
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search <search-term>",
	Short: "Search an Algolia index and display results",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if searchPage < 1 {
			return fmt.Errorf("page must be at least 1, got %d", searchPage)
		}

		client, err := search.NewClient(metaCfg.Algolia.AppID, metaCfg.Algolia.SearchKey)
		if err != nil {
			return fmt.Errorf("creating algolia client: %w", err)
		}

		// Algolia pages are 0-indexed; the flag is 1-based.
		resp, err := searchSingleIndex(cmd.Context(), client, searchIndex, args[0], searchPage-1)
		if err != nil {
			return err
		}

		pretty.Print(buildPage(resp))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().Int32VarP(&searchPage, "page", "p", 1, "Search results page to retrieve (1-based)")
	searchCmd.Flags().StringVarP(&searchIndex, "index", "i", "fcg-index", "Algolia index to search")
}

// searchSingleIndex runs one query; page is Algolia's 0-based page number.
func searchSingleIndex(ctx context.Context, client *search.APIClient, index, term string, page int32) (*search.SearchResponse, error) {
	params := search.SearchParamsObjectAsSearchParams(
		search.NewSearchParamsObject().SetQuery(term).SetPage(page),
	)
	request := client.NewApiSearchSingleIndexRequest(index).WithSearchParams(params)

	resp, err := client.SearchSingleIndex(request, search.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("searching %q for %q: %w", index, term, err)
	}
	return resp, nil
}

// buildPage flattens a response into the shape the templates consume.
func buildPage(resp *search.SearchResponse) IDXOut {
	hits := make([]map[string]any, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		hits = append(hits, hitToMap(hit))
	}

	return IDXOut{
		Page:        resp.GetPage() + 1,
		NumPages:    resp.GetNbPages(),
		NbHits:      resp.GetNbHits(),
		HitsPerPage: resp.GetHitsPerPage(),
		Hits:        hits,
	}
}

// hitToMap starts from every retrievable attribute (searchable or not, all of
// which the SDK leaves in AdditionalProperties) and overlays the highlighted
// text for the attributes that have it. Highlights are template.HTML so the
// <em> tags survive html/template. That is only safe while the index data is
// trusted, since Algolia does not escape field text unless escapeHTML is set.
func hitToMap(hit search.Hit) map[string]any {
	m := make(map[string]any, len(hit.AdditionalProperties)+1)
	maps.Copy(m, hit.AdditionalProperties)
	m["ObjectID"] = hit.ObjectID

	for key, hr := range hit.GetHighlightResult() {
		if hr.HighlightResultOption != nil {
			m[key] = template.HTML(hr.HighlightResultOption.Value)
		}
	}
	return m
}
