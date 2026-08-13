package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// resolveRealTitle reads a Hugo mreviews term page's own <urlPath>/index.json
// (rendered by themes/guild/layouts/mreviews/term.json) and returns
// Metadata.LinkTitle -- the LinkTitle of the first review page carrying that
// tag, i.e. the same title Hugo's tag-to-title.html partial resolves the tag
// to. Hugo has already computed and published that answer, so this reads it
// rather than re-deriving it; see TermPage.
func resolveRealTitle(hugoRoot, urlPath string) (string, error) {

	termPath := filepath.Join(hugoRoot, urlPath, "index.json")
	termBytes, err := os.ReadFile(termPath)
	if err != nil {
		return "", fmt.Errorf("error reading term page %s: %w", termPath, err)
	}

	var term TermPage
	if err := json.Unmarshal(termBytes, &term); err != nil {
		return "", fmt.Errorf("error unmarshaling term page %s: %w", termPath, err)
	}

	return term.Metadata.LinkTitle, nil
}
