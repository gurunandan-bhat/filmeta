package cmd

import (
	"crypto/md5"
	"encoding/json"
	"filmeta/tagname"
	"fmt"
	"os"
	"path/filepath"
)

// writeTagAlias writes a second copy of a film's metadata under md5(tag),
// where tag is the canonical Hugo taxonomy tag for the title (package
// tagname). This lets a film's metadata be looked up by either md5(title) or
// md5(tag), with no conditional in the caller.
//
// No copy is written -- wrote is false -- when the tag equals the title
// exactly (the common case) or when the title has no ASCII tag at all (e.g.
// a title written entirely in a non-Latin script; see tagname.FromTitle).
//
// This only covers tags that are a pure fold of a known-correct title (e.g.
// after a hugo-retag canonicalization pass). It does not cover tags that
// diverge from any title-fold at all -- see resolveTagsFromReviews for that.
func writeTagAlias(metaDir, title string, jsonBytes []byte) (wrote bool, err error) {

	tag := tagname.FromTitle(title)
	if tag == "" || tag == title {
		return false, nil
	}

	return writeAliasAt(metaDir, tag, jsonBytes)
}

// writeAliasAt writes jsonBytes to md5(target).json in metaDir, so a film's
// metadata can be found under a second key regardless of how that key was
// derived. Shared by writeTagAlias (target = a folded tag) and
// resolveTagsFromReviews (target = a raw Hugo taxonomy term).
//
// Two different titles are not expected to ever collide on the same target --
// collisions are guarded on the Hugo side -- so one here is unexpected. It is
// reported to stderr rather than treated as fatal, and the new write wins.
func writeAliasAt(metaDir, target string, jsonBytes []byte) (wrote bool, err error) {

	fName := fmt.Sprintf("%x.json", md5.Sum([]byte(target)))
	destPath := filepath.Join(metaDir, fName)

	if existing, err := os.ReadFile(destPath); err == nil {
		var probe struct {
			FCGTitle string `json:"fcg_title"`
		}
		var incoming struct {
			FCGTitle string `json:"fcg_title"`
		}
		if err := json.Unmarshal(existing, &probe); err == nil && probe.FCGTitle != "" {
			if err := json.Unmarshal(jsonBytes, &incoming); err == nil && incoming.FCGTitle != "" && probe.FCGTitle != incoming.FCGTitle {
				fmt.Fprintf(os.Stderr, "warning: alias %q collides with existing entry for %q; overwriting with %q\n", target, probe.FCGTitle, incoming.FCGTitle)
			}
		}
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("error checking existing alias %s: %w", destPath, err)
	}

	if err := os.WriteFile(destPath, jsonBytes, 0644); err != nil {
		return false, fmt.Errorf("error writing alias %s: %w", destPath, err)
	}

	return true, nil
}

// metaFileExists reports whether metaDir already holds a metadata file keyed
// by md5(name).json -- shared by every command that needs to ask "does this
// film already have metadata under this name" without caring how.
func metaFileExists(metaDir, name string) (bool, error) {

	path := filepath.Join(metaDir, fmt.Sprintf("%x.json", md5.Sum([]byte(name))))
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("error checking %s: %w", path, err)
	}
}
