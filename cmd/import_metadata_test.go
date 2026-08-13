package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// When MReviews is empty (an older-format input file, or a hand-written one
// that never set it), only the title-keyed file is written -- the same
// behaviour import always had.
func TestWriteFilmMetadataWritesOnlyTitleWhenMReviewsIsEmpty(t *testing.T) {

	dir := t.TempDir()
	film := FilmOut{LinkTitle: "Solo Title"}
	body := []byte(`{"fcg_title":"Solo Title"}`)

	if err := writeFilmMetadata(dir, film, body); err != nil {
		t.Fatalf("writeFilmMetadata: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("dir has %d entries, want 1", len(entries))
	}

	titleName := fmt.Sprintf("%x.json", md5.Sum([]byte("Solo Title")))
	if entries[0].Name() != titleName {
		t.Errorf("file = %s, want %s", entries[0].Name(), titleName)
	}
}

// When MReviews equals LinkTitle -- the common case -- only one file is
// written, not two identical copies under the same hash.
func TestWriteFilmMetadataWritesOnlyOneFileWhenTagMatchesTitle(t *testing.T) {

	dir := t.TempDir()
	film := FilmOut{LinkTitle: "Same Both Ways", MReviews: "Same Both Ways"}
	body := []byte(`{"fcg_title":"Same Both Ways"}`)

	if err := writeFilmMetadata(dir, film, body); err != nil {
		t.Fatalf("writeFilmMetadata: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("dir has %d entries, want 1", len(entries))
	}
}

// The motivating case: title and tag differ, so two byte-identical copies
// are written, one keyed by each.
func TestWriteFilmMetadataWritesBothFilesWhenTagDiffers(t *testing.T) {

	dir := t.TempDir()
	film := FilmOut{LinkTitle: "The Devil's Mouth", MReviews: "The Devil S Mouth"}
	body := []byte(`{"fcg_title":"The Devil's Mouth","overview":"spooky"}`)

	if err := writeFilmMetadata(dir, film, body); err != nil {
		t.Fatalf("writeFilmMetadata: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("dir has %d entries, want 2", len(entries))
	}

	for _, name := range []string{
		fmt.Sprintf("%x.json", md5.Sum([]byte("The Devil's Mouth"))),
		fmt.Sprintf("%x.json", md5.Sum([]byte("The Devil S Mouth"))),
	} {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		if string(got) != string(body) {
			t.Errorf("%s content = %s, want byte-identical copy %s", name, got, body)
		}
	}
}

// Re-running writeFilmMetadata for the same film (a re-import) must not
// spuriously warn about a collision with itself.
func TestWriteFilmMetadataIsIdempotent(t *testing.T) {

	dir := t.TempDir()
	film := FilmOut{LinkTitle: "Repeatable", MReviews: "Repeatable Tag"}
	body := []byte(`{"fcg_title":"Repeatable"}`)

	if err := writeFilmMetadata(dir, film, body); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := writeFilmMetadata(dir, film, body); err != nil {
		t.Fatalf("second write: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("dir has %d entries after two identical writes, want 2", len(entries))
	}
}

// A sanity check that the title file really is addressable by title alone,
// matching the md5(LinkTitle) convention every other command relies on.
func TestWriteFilmMetadataTitleFileIsFcgTitleAddressed(t *testing.T) {

	dir := t.TempDir()
	film := FilmOut{LinkTitle: "Findable Film", MReviews: "Findable Tag"}
	body := []byte(`{"fcg_title":"Findable Film"}`)

	if err := writeFilmMetadata(dir, film, body); err != nil {
		t.Fatalf("writeFilmMetadata: %v", err)
	}

	titleName := fmt.Sprintf("%x.json", md5.Sum([]byte("Findable Film")))
	got, err := os.ReadFile(filepath.Join(dir, titleName))
	if err != nil {
		t.Fatalf("reading title file: %v", err)
	}

	var probe struct {
		FCGTitle string `json:"fcg_title"`
	}
	if err := json.Unmarshal(got, &probe); err != nil {
		t.Fatalf("unmarshaling: %v", err)
	}
	if probe.FCGTitle != "Findable Film" {
		t.Errorf("fcg_title = %q, want %q", probe.FCGTitle, "Findable Film")
	}
}
