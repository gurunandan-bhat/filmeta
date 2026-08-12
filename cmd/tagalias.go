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
// Two different titles are not expected to ever collide on the same tag --
// that is guarded on the Hugo side -- so a collision here is unexpected. It
// is reported to stderr rather than treated as fatal, and the new write wins.
func writeTagAlias(metaDir, title string, jsonBytes []byte) (wrote bool, err error) {

	tag := tagname.FromTitle(title)
	if tag == "" || tag == title {
		return false, nil
	}

	fName := fmt.Sprintf("%x.json", md5.Sum([]byte(tag)))
	destPath := filepath.Join(metaDir, fName)

	if existing, err := os.ReadFile(destPath); err == nil {
		var probe struct {
			FCGTitle string `json:"fcg_title"`
		}
		if err := json.Unmarshal(existing, &probe); err == nil && probe.FCGTitle != "" && probe.FCGTitle != title {
			fmt.Fprintf(os.Stderr, "warning: tag %q for %q collides with existing entry for %q; overwriting\n", tag, title, probe.FCGTitle)
		}
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("error checking existing tag alias %s: %w", destPath, err)
	}

	if err := os.WriteFile(destPath, jsonBytes, 0644); err != nil {
		return false, fmt.Errorf("error writing tag alias %s: %w", destPath, err)
	}

	return true, nil
}
