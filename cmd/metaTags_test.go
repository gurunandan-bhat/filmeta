package cmd

import (
	"crypto/md5"
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
