package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, name string, mode os.FileMode) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(`[]`), mode); err != nil {
		t.Fatalf("writing fixture %s: %v", path, err)
	}
	// WriteFile applies the process umask, so set the mode explicitly.
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("chmod %s: %v", path, err)
	}

	return path
}

// The original check was `info.Mode().Perm()&0444 != 0444`, which demanded the
// file be readable by group and world. A private file the invoking user owns is
// perfectly readable and must be accepted.
func TestDataIsAvailableAcceptsPrivateFile(t *testing.T) {

	for _, mode := range []os.FileMode{0600, 0400, 0640, 0644, 0666} {
		path := writeFile(t, "data.json", mode)
		if err := dataIsAvailable(nil, []string{path}); err != nil {
			t.Errorf("mode %04o: unexpected error: %v", mode, err)
		}
	}
}

func TestDataIsAvailableRejectsUnreadable(t *testing.T) {

	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses file permissions")
	}

	path := writeFile(t, "data.json", 0000)
	if err := dataIsAvailable(nil, []string{path}); err == nil {
		t.Error("expected an error for a file with no read permission")
	}
}

func TestDataIsAvailableRejectsMissingAndDirectory(t *testing.T) {

	dir := t.TempDir()

	if err := dataIsAvailable(nil, []string{filepath.Join(dir, "absent.json")}); err == nil {
		t.Error("expected an error for a missing file")
	}
	if err := dataIsAvailable(nil, []string{dir}); err == nil {
		t.Error("expected an error for a directory")
	}
	if err := dataIsAvailable(nil, []string{""}); err == nil {
		t.Error("expected an error for an empty path")
	}
}
