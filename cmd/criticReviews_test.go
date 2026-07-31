package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The missing-index branch used to render member.Organizations[0] directly,
// which panics for a member whose guild entry names no organization. All 56
// members currently have one, so the crash was latent rather than active.
func TestJoinOrgsHandlesNoOrganizations(t *testing.T) {

	member := Guild{Name: "Nameless Critic"}

	orgMap := make(map[string]int)
	for _, org := range member.Organizations {
		orgMap[org] = 1
	}

	if got := joinOrgs(orgMap); got != "" {
		t.Errorf("joinOrgs on an empty set = %q, want the empty string", got)
	}
}

// Organizations come from a map, so without sorting the CSV column would vary
// between runs for anyone with more than one.
func TestJoinOrgsIsSortedAndDeduplicated(t *testing.T) {

	orgMap := map[string]int{
		"The Hollywood Reporter India": 1,
		"Film Companion":               1,
		"Baradwaj Rangan":              1,
	}

	want := "Baradwaj Rangan, Film Companion, The Hollywood Reporter India"
	for range 20 {
		if got := joinOrgs(orgMap); got != want {
			t.Fatalf("joinOrgs = %q, want %q", got, want)
		}
	}
}

func TestProcessCriticCountsAndCollectsPublications(t *testing.T) {

	reviews := map[string]CriticReview{
		"a": {Publication: "Film Companion", PublishDate: mustParse(t, "2026-06-13T19:36:41+05:30")},
		"b": {Publication: "  Mint  ", PublishDate: mustParse(t, "2026-06-01T00:00:00+05:30")},
		"c": {Publication: "", PublishDate: mustParse(t, "2026-06-30T23:59:59+05:30")},
		"d": {Publication: "Out Of Range", PublishDate: mustParse(t, "2026-07-01T00:00:00+05:30")},
	}

	path := filepath.Join(t.TempDir(), "index.json")
	body, err := json.Marshal(reviews)
	if err != nil {
		t.Fatalf("marshaling fixture: %v", err)
	}
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	fh, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer func() { _ = fh.Close() }()

	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 0, 30)

	orgMap := map[string]int{"Existing Org": 1}
	count, err := processCritic(fh, orgMap, start, end)
	if err != nil {
		t.Fatalf("processCritic: %v", err)
	}

	if count != 3 {
		t.Errorf("count = %d, want 3 (the fourth review falls outside the window)", count)
	}
	// "  Mint  " must be trimmed, and the blank publication must not appear.
	if got, want := joinOrgs(orgMap), "Existing Org, Film Companion, Mint"; got != want {
		t.Errorf("joinOrgs = %q, want %q", got, want)
	}
}
