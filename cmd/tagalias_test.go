package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// When the tag equals the title -- the common case -- no alias file is
// written at all.
func TestWriteTagAliasSkipsWhenTagMatchesTitle(t *testing.T) {

	dir := t.TempDir()
	wrote, err := writeTagAlias(dir, "Node", []byte(`{"fcg_title":"Node"}`))
	if err != nil {
		t.Fatalf("writeTagAlias: %v", err)
	}
	if wrote {
		t.Fatal("wrote = true, want false when tag == title")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has %d entries, want 0", len(entries))
	}
}

// A title with no ASCII-foldable character (tagname.FromTitle returns "")
// must not produce a file keyed on the empty string.
func TestWriteTagAliasSkipsWhenTagIsEmpty(t *testing.T) {

	dir := t.TempDir()
	wrote, err := writeTagAlias(dir, "日本語", []byte(`{"fcg_title":"日本語"}`))
	if err != nil {
		t.Fatalf("writeTagAlias: %v", err)
	}
	if wrote {
		t.Fatal("wrote = true, want false when the title has no ASCII tag")
	}
}

// The common motivating case: "Q&A: What's Next?" tags to "Q A What s Next",
// so a second copy must land at md5(tag).json alongside the title-keyed one.
func TestWriteTagAliasCreatesCopyWhenTitleAndTagDiffer(t *testing.T) {

	dir := t.TempDir()
	title := "Q&A: What's Next?"
	body := fmt.Appendf(nil, `{"fcg_title":%q}`, title)
	wrote, err := writeTagAlias(dir, title, body)
	if err != nil {
		t.Fatalf("writeTagAlias: %v", err)
	}
	if !wrote {
		t.Fatal("wrote = false, want true when title and tag differ")
	}

	wantName := fmt.Sprintf("%x.json", md5.Sum([]byte("Q A What s Next")))
	got, err := os.ReadFile(filepath.Join(dir, wantName))
	if err != nil {
		t.Fatalf("reading alias file %s: %v", wantName, err)
	}
	if string(got) != string(body) {
		t.Errorf("alias content = %s, want %s", got, body)
	}
}

// Two different titles are not expected to collide on the same tag -- that is
// guarded on the Hugo side -- but if it happens the newer write must win
// rather than the command failing.
func TestWriteTagAliasOverwritesOnCollision(t *testing.T) {

	dir := t.TempDir()

	first := []byte(`{"fcg_title":"Café"}`)
	if _, err := writeTagAlias(dir, "Café", first); err != nil {
		t.Fatalf("first write: %v", err)
	}

	second := []byte(`{"fcg_title":"Cafe!"}`)
	wrote, err := writeTagAlias(dir, "Cafe!", second)
	if err != nil {
		t.Fatalf("second write: %v", err)
	}
	if !wrote {
		t.Fatal("wrote = false on collision, want true (overwrite)")
	}

	wantName := fmt.Sprintf("%x.json", md5.Sum([]byte("Cafe")))
	got, err := os.ReadFile(filepath.Join(dir, wantName))
	if err != nil {
		t.Fatalf("reading alias file %s: %v", wantName, err)
	}

	var probe struct {
		FCGTitle string `json:"fcg_title"`
	}
	if err := json.Unmarshal(got, &probe); err != nil {
		t.Fatalf("unmarshaling alias content: %v", err)
	}
	if probe.FCGTitle != "Cafe!" {
		t.Errorf("fcg_title = %q, want %q (the second write should win)", probe.FCGTitle, "Cafe!")
	}
}
