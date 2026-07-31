package cmd

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func writeImage(t *testing.T, dir, name string, body []byte) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatalf("writing fixture %s: %v", path, err)
	}

	return path
}

// The site's hand-added backdrops are named for the md5 of their own contents,
// and referenced with the leading slash TMDB's own paths carry, because Hugo
// resolves the asset as "meta/backdrops" + backdrop_path.
func TestCopyLocalBackdropIsContentAddressed(t *testing.T) {

	dir := t.TempDir()
	destDir := t.TempDir()

	body := []byte("not really a jpeg, but the bytes are what get hashed")
	src := writeImage(t, dir, "poster-from-the-publicist.jpg", body)

	got, err := copyLocalBackdrop(src, destDir)
	if err != nil {
		t.Fatalf("copyLocalBackdrop: %v", err)
	}

	want := fmt.Sprintf("/%x.jpg", md5.Sum(body))
	if got != want {
		t.Errorf("backdrop_path = %q, want %q", got, want)
	}

	copied, err := os.ReadFile(filepath.Join(destDir, filepath.Base(got)))
	if err != nil {
		t.Fatalf("reading the copy: %v", err)
	}
	if string(copied) != string(body) {
		t.Error("the copied file does not match the source bytes")
	}
}

// Content addressing has to make a repeated import idempotent: the same image
// must land on the same name rather than accumulating copies.
func TestCopyLocalBackdropRepeatsCleanly(t *testing.T) {

	dir := t.TempDir()
	destDir := t.TempDir()

	body := []byte("the same bytes twice")
	first := writeImage(t, dir, "one.jpg", body)
	second := writeImage(t, dir, "a-different-name.jpg", body)

	pathA, err := copyLocalBackdrop(first, destDir)
	if err != nil {
		t.Fatalf("first copy: %v", err)
	}
	pathB, err := copyLocalBackdrop(second, destDir)
	if err != nil {
		t.Fatalf("second copy: %v", err)
	}

	if pathA != pathB {
		t.Errorf("same bytes gave %q then %q", pathA, pathB)
	}

	entries, err := os.ReadDir(destDir)
	if err != nil {
		t.Fatalf("reading dest dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("dest dir holds %d files, want 1", len(entries))
	}
}

// An upper-case extension would otherwise produce a name inconsistent with
// every backdrop already in the site.
func TestCopyLocalBackdropNormalisesExtension(t *testing.T) {

	dir := t.TempDir()
	destDir := t.TempDir()

	src := writeImage(t, dir, "SHOUTY.JPG", []byte("bytes"))

	got, err := copyLocalBackdrop(src, destDir)
	if err != nil {
		t.Fatalf("copyLocalBackdrop: %v", err)
	}
	if filepath.Ext(got) != ".jpg" {
		t.Errorf("extension = %q, want %q", filepath.Ext(got), ".jpg")
	}
}

// A path the user typed by hand is the likeliest thing to be wrong, so these
// are hard errors rather than a warning that scrolls past in a bulk run.
func TestCopyLocalBackdropRejectsBadInput(t *testing.T) {

	dir := t.TempDir()
	destDir := t.TempDir()

	cases := []struct {
		name string
		src  string
	}{
		{"missing file", filepath.Join(dir, "absent.jpg")},
		{"no extension", writeImage(t, dir, "backdrop", []byte("bytes"))},
		{"empty file", writeImage(t, dir, "empty.jpg", nil)},
		{"a directory", dir},
	}

	for _, c := range cases {
		if _, err := copyLocalBackdrop(c.src, destDir); err == nil {
			t.Errorf("%s: expected an error", c.name)
		}
	}
}
