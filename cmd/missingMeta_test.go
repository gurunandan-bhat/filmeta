package cmd

import "testing"

// A film already covered by a metadata file -- under either its tag or its
// real title, aliases included -- is not reported missing at all.
func TestFindMissingSkipsFilmsWithExistingMetadata(t *testing.T) {

	hugoRoot := t.TempDir()
	metaDir := t.TempDir()

	writeMetaFixture(t, metaDir, "Already Covered")

	films := map[string]Film{
		"key1": {LinkTitle: "Already Covered", URLPath: "/mreviews/already-covered"},
	}

	missing, err := findMissing(films, hugoRoot, metaDir)
	if err != nil {
		t.Fatalf("findMissing: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("missing = %v, want none", missing)
	}
}

// The motivating case: a review's title was hand-corrected after the
// archetype derived a mangled tag from the filename. findMissing must recover
// the real title from the tag's own term page, same as resolveTagsFromReviews
// does, and still carry the raw tag through in MReviews.
func TestFindMissingResolvesRealTitleFromTermPage(t *testing.T) {

	hugoRoot := t.TempDir()
	metaDir := t.TempDir()

	writeTermPage(t, hugoRoot, "mreviews/the-devil-s-mouth", "The Devil's Mouth")
	films := map[string]Film{
		"key1": {LinkTitle: "The Devil S Mouth", URLPath: "/mreviews/the-devil-s-mouth"},
	}

	missing, err := findMissing(films, hugoRoot, metaDir)
	if err != nil {
		t.Fatalf("findMissing: %v", err)
	}
	if len(missing) != 1 {
		t.Fatalf("missing = %v, want exactly 1", missing)
	}

	got := missing[0]
	if got.LinkTitle != "The Devil's Mouth" {
		t.Errorf("LinkTitle = %q, want the resolved real title", got.LinkTitle)
	}
	if got.MReviews != "The Devil S Mouth" {
		t.Errorf("MReviews = %q, want the raw tag", got.MReviews)
	}
}

// A film whose real title already has metadata -- just not under its tag --
// is not "missing" at all. It needs a tag alias (metaTags' job), not a fresh
// TMDB import, so findMissing must check both hashes before reporting it.
func TestFindMissingSkipsWhenRealTitleAlreadyHasMetadata(t *testing.T) {

	hugoRoot := t.TempDir()
	metaDir := t.TempDir()

	writeMetaFixture(t, metaDir, "The Devil's Mouth")
	writeTermPage(t, hugoRoot, "mreviews/the-devil-s-mouth", "The Devil's Mouth")
	films := map[string]Film{
		"key1": {LinkTitle: "The Devil S Mouth", URLPath: "/mreviews/the-devil-s-mouth"},
	}

	missing, err := findMissing(films, hugoRoot, metaDir)
	if err != nil {
		t.Fatalf("findMissing: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("missing = %v, want none -- metadata already exists under the real title", missing)
	}
}

// When the term page can't be resolved -- content not yet built, or
// malformed -- findMissing must fall back to the tag rather than fail the
// whole run, since this is a report command surfacing many films at once.
func TestFindMissingFallsBackToTagWhenTermPageUnavailable(t *testing.T) {

	hugoRoot := t.TempDir()
	metaDir := t.TempDir()
	// Deliberately no term page fixture for this URLPath.

	films := map[string]Film{
		"key1": {LinkTitle: "Some Tag", URLPath: "/mreviews/some-tag"},
	}

	missing, err := findMissing(films, hugoRoot, metaDir)
	if err != nil {
		t.Fatalf("findMissing: %v", err)
	}
	if len(missing) != 1 {
		t.Fatalf("missing = %v, want exactly 1", missing)
	}

	got := missing[0]
	if got.LinkTitle != "Some Tag" || got.MReviews != "Some Tag" {
		t.Errorf("got %+v, want both fields to fall back to the tag", got)
	}
}

// A tag whose term page resolves to itself (no title correction ever
// happened) still reports correctly: both fields carry the same tag.
func TestFindMissingHandlesSelfResolvingTag(t *testing.T) {

	hugoRoot := t.TempDir()
	metaDir := t.TempDir()

	writeTermPage(t, hugoRoot, "mreviews/foo-bar", "Foo Bar")
	films := map[string]Film{
		"key1": {LinkTitle: "Foo Bar", URLPath: "/mreviews/foo-bar"},
	}

	missing, err := findMissing(films, hugoRoot, metaDir)
	if err != nil {
		t.Fatalf("findMissing: %v", err)
	}
	if len(missing) != 1 {
		t.Fatalf("missing = %v, want exactly 1", missing)
	}

	got := missing[0]
	if got.LinkTitle != "Foo Bar" || got.MReviews != "Foo Bar" {
		t.Errorf("got %+v, want both fields set to the tag", got)
	}
}
