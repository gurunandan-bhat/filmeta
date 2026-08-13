package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func writeMetaFixture(t *testing.T, dir, title string) {
	t.Helper()

	fName := fmt.Sprintf("%x.json", md5.Sum([]byte(title)))
	body := fmt.Sprintf(`{"fcg_title":%q}`, title)
	if err := os.WriteFile(filepath.Join(dir, fName), []byte(body), 0644); err != nil {
		t.Fatalf("writing fixture for %q: %v", title, err)
	}
}

// A directory of already-imported metadata files should end up with an alias
// copy for every title whose canonical tag differs, none for the rest, and
// non-.json entries left untouched.
func TestBackfillTagAliasesScansDirectory(t *testing.T) {

	dir := t.TempDir()

	writeMetaFixture(t, dir, "Node")              // tag == title, no alias
	writeMetaFixture(t, dir, "Q&A: What's Next?") // tag differs, gets an alias
	writeMetaFixture(t, dir, "日本語")               // no ASCII tag, no alias

	// A file that isn't a metadata blob at all must be left alone.
	if err := os.WriteFile(filepath.Join(dir, "posters"), []byte("not json"), 0644); err != nil {
		t.Fatalf("writing non-json entry: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "posters-dir"), 0755); err != nil {
		t.Fatalf("making subdir: %v", err)
	}

	created, unchanged, noTitle, err := backfillTagAliases(dir)
	if err != nil {
		t.Fatalf("backfillTagAliases: %v", err)
	}
	if created != 1 {
		t.Errorf("created = %d, want 1", created)
	}
	if unchanged != 2 {
		t.Errorf("unchanged = %d, want 2", unchanged)
	}
	if noTitle != 0 {
		t.Errorf("noTitle = %d, want 0", noTitle)
	}

	aliasName := fmt.Sprintf("%x.json", md5.Sum([]byte("Q A What s Next")))
	if _, err := os.Stat(filepath.Join(dir, aliasName)); err != nil {
		t.Errorf("expected alias file %s: %v", aliasName, err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	// 3 fixtures + 1 alias + 1 non-json file + 1 subdir = 6 entries.
	if len(entries) != 6 {
		t.Errorf("dir has %d entries, want 6", len(entries))
	}
}

// A metadata file with no fcg_title (malformed or from an older format) is
// counted and skipped rather than failing the whole run.
func TestBackfillTagAliasesSkipsFilesWithNoTitle(t *testing.T) {

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "orphan.json"), []byte(`{"overview":"no title here"}`), 0644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	created, unchanged, noTitle, err := backfillTagAliases(dir)
	if err != nil {
		t.Fatalf("backfillTagAliases: %v", err)
	}
	if created != 0 || unchanged != 0 {
		t.Errorf("created=%d unchanged=%d, want 0, 0", created, unchanged)
	}
	if noTitle != 1 {
		t.Errorf("noTitle = %d, want 1", noTitle)
	}
}

// Re-running the backfill must be idempotent: no new files ever appear, even
// though the alias copy itself carries the same fcg_title as the original and
// so is recomputed (and rewritten to itself, unchanged) on every pass.
func TestBackfillTagAliasesIsIdempotent(t *testing.T) {

	dir := t.TempDir()
	writeMetaFixture(t, dir, "Q&A: What's Next?")

	if _, _, _, err := backfillTagAliases(dir); err != nil {
		t.Fatalf("first run: %v", err)
	}

	entriesAfterFirst, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entriesAfterFirst) != 2 {
		t.Fatalf("dir has %d entries after first run, want 2 (original + alias)", len(entriesAfterFirst))
	}

	created, unchanged, _, err := backfillTagAliases(dir)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	// Both files -- the original and the alias -- carry the same fcg_title,
	// so both are recomputed as needing the alias write on every pass; the
	// write just lands on the same bytes.
	if created != 2 {
		t.Errorf("second run created = %d, want 2 (both files recompute the same alias)", created)
	}
	if unchanged != 0 {
		t.Errorf("second run unchanged = %d, want 0", unchanged)
	}

	entriesAfterSecond, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entriesAfterSecond) != len(entriesAfterFirst) {
		t.Errorf("entry count changed from %d to %d on re-run", len(entriesAfterFirst), len(entriesAfterSecond))
	}
}

func writeMreviewsIndex(t *testing.T, hugoRoot string, films map[string]Film) {
	t.Helper()

	body, err := json.Marshal(films)
	if err != nil {
		t.Fatalf("marshaling mreviews index: %v", err)
	}
	dir := filepath.Join(hugoRoot, "mreviews")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("making mreviews dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.json"), body, 0644); err != nil {
		t.Fatalf("writing mreviews index: %v", err)
	}
}

func writeTermPage(t *testing.T, hugoRoot, urlPath, linkTitle string) {
	t.Helper()

	dir := filepath.Join(hugoRoot, urlPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("making term page dir: %v", err)
	}
	body := fmt.Sprintf(`{"Metadata":{"LinkTitle":%q}}`, linkTitle)
	if err := os.WriteFile(filepath.Join(dir, "index.json"), []byte(body), 0644); err != nil {
		t.Fatalf("writing term page: %v", err)
	}
}

// The motivating case: a review's title was hand-corrected after the archetype
// derived a mangled tag from the filename, so the live mreviews tag no longer
// matches any fold of a known title. The real title is recovered via the
// term page Hugo already published, exactly mirroring tag-to-title.html.
func TestResolveTagsFromReviewsCreatesAliasForDivergedTitle(t *testing.T) {

	hugoRoot := t.TempDir()
	metaDir := t.TempDir()

	writeMetaFixture(t, metaDir, "The Devil's Mouth")
	writeMreviewsIndex(t, hugoRoot, map[string]Film{
		"key1": {LinkTitle: "The Devil S Mouth", URLPath: "/mreviews/the-devil-s-mouth"},
	})
	writeTermPage(t, hugoRoot, "mreviews/the-devil-s-mouth", "The Devil's Mouth")

	created, alreadyPresent, unresolved, err := resolveTagsFromReviews(hugoRoot, metaDir)
	if err != nil {
		t.Fatalf("resolveTagsFromReviews: %v", err)
	}
	if created != 1 {
		t.Errorf("created = %d, want 1", created)
	}
	if alreadyPresent != 0 || unresolved != 0 {
		t.Errorf("alreadyPresent=%d unresolved=%d, want 0, 0", alreadyPresent, unresolved)
	}

	aliasName := fmt.Sprintf("%x.json", md5.Sum([]byte("The Devil S Mouth")))
	got, err := os.ReadFile(filepath.Join(metaDir, aliasName))
	if err != nil {
		t.Fatalf("reading alias file %s: %v", aliasName, err)
	}

	var probe struct {
		FCGTitle string `json:"fcg_title"`
	}
	if err := json.Unmarshal(got, &probe); err != nil {
		t.Fatalf("unmarshaling alias content: %v", err)
	}
	if probe.FCGTitle != "The Devil's Mouth" {
		t.Errorf("alias fcg_title = %q, want the real title copied verbatim", probe.FCGTitle)
	}
}

// A tag that already has its own metadata file needs no resolution at all --
// the term page is never even consulted.
func TestResolveTagsFromReviewsSkipsTagsWithExistingMetadata(t *testing.T) {

	hugoRoot := t.TempDir()
	metaDir := t.TempDir()

	writeMetaFixture(t, metaDir, "Already Covered")
	writeMreviewsIndex(t, hugoRoot, map[string]Film{
		"key1": {LinkTitle: "Already Covered", URLPath: "/mreviews/already-covered"},
	})
	// Deliberately no term page fixture: resolving this tag must not require one.

	created, alreadyPresent, unresolved, err := resolveTagsFromReviews(hugoRoot, metaDir)
	if err != nil {
		t.Fatalf("resolveTagsFromReviews: %v", err)
	}
	if alreadyPresent != 1 {
		t.Errorf("alreadyPresent = %d, want 1", alreadyPresent)
	}
	if created != 0 || unresolved != 0 {
		t.Errorf("created=%d unresolved=%d, want 0, 0", created, unresolved)
	}
}

// When a tag resolves to itself -- no review's title was ever corrected away
// from the archetype-derived tag -- there is nothing to alias to; it is a
// genuine missingMeta gap and must be reported, not silently skipped.
func TestResolveTagsFromReviewsReportsSelfResolvingTagAsUnresolved(t *testing.T) {

	hugoRoot := t.TempDir()
	metaDir := t.TempDir()

	writeMreviewsIndex(t, hugoRoot, map[string]Film{
		"key1": {LinkTitle: "Foo Bar", URLPath: "/mreviews/foo-bar"},
	})
	writeTermPage(t, hugoRoot, "mreviews/foo-bar", "Foo Bar")

	created, alreadyPresent, unresolved, err := resolveTagsFromReviews(hugoRoot, metaDir)
	if err != nil {
		t.Fatalf("resolveTagsFromReviews: %v", err)
	}
	if unresolved != 1 {
		t.Errorf("unresolved = %d, want 1", unresolved)
	}
	if created != 0 || alreadyPresent != 0 {
		t.Errorf("created=%d alreadyPresent=%d, want 0, 0", created, alreadyPresent)
	}
}

// When the resolved real title itself has no metadata either, that is also a
// genuine missingMeta gap -- reported as unresolved rather than aliased to
// nothing.
func TestResolveTagsFromReviewsReportsMissingResolvedTitleAsUnresolved(t *testing.T) {

	hugoRoot := t.TempDir()
	metaDir := t.TempDir()

	writeMreviewsIndex(t, hugoRoot, map[string]Film{
		"key1": {LinkTitle: "Messy Tag", URLPath: "/mreviews/messy-tag"},
	})
	writeTermPage(t, hugoRoot, "mreviews/messy-tag", "The Real Title")
	// No metadata fixture for "The Real Title" either.

	created, alreadyPresent, unresolved, err := resolveTagsFromReviews(hugoRoot, metaDir)
	if err != nil {
		t.Fatalf("resolveTagsFromReviews: %v", err)
	}
	if unresolved != 1 {
		t.Errorf("unresolved = %d, want 1", unresolved)
	}
	if created != 0 || alreadyPresent != 0 {
		t.Errorf("created=%d alreadyPresent=%d, want 0, 0", created, alreadyPresent)
	}
}
