package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// userHomeDir is a thin wrapper around os.UserHomeDir for testability.
var userHomeDir = os.UserHomeDir

// resolveIndexDir returns the directory for index files, resolved from:
// GEMATRIA_INDEX_LOCATION > XDG_DATA_HOME/gematria > ~/.local/share/gematria.
func resolveIndexDir(getenv func(string) string) (string, error) {
	if loc := getenv("GEMATRIA_INDEX_LOCATION"); loc != "" {
		return loc, nil
	}
	xdg := getenv("XDG_DATA_HOME")
	if xdg == "" {
		home, err := userHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		xdg = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(xdg, "gematria"), nil
}

// resolveIndexPath returns the full path for the default index output file.
// It resolves location from GEMATRIA_INDEX_LOCATION, then XDG_DATA_HOME/gematria,
// then ~/.local/share/gematria. The filename (without extension) is controlled
// by GEMATRIA_INDEX_NAME (default: "gematria"). The extension is ".db" for
// sqlite format and ".idx" for index format.
//
// Does NOT create the directory — the caller is responsible for that.
func resolveIndexPath(format string, getenv func(string) string) (string, error) {
	// Step 1: resolve location directory.
	location, err := resolveIndexDir(getenv)
	if err != nil {
		return "", err
	}

	// Step 2: resolve index name.
	name := getenv("GEMATRIA_INDEX_NAME")
	if name == "" {
		name = "gematria"
	}

	// Step 3: validate name — no path separators allowed.
	if strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("invalid GEMATRIA_INDEX_NAME %q: must not contain path separators", name)
	}

	// Step 4: determine extension.
	var ext string
	if format == "index" {
		ext = ".idx"
	} else {
		ext = ".db"
	}

	return filepath.Join(location, name+ext), nil
}

// discoverDefaultIndex locates an existing index file at the default location.
// It checks for <location>/<name>.db first (SQLite, preferred), then
// <location>/<name>.idx. Returns the path and true if found, or "" and false.
//
// Does NOT create directories or validate file contents.
// Invalid GEMATRIA_INDEX_NAME (path separator) → found=false, no error.
func discoverDefaultIndex(getenv func(string) string) (string, bool) {
	location, err := resolveIndexDir(getenv)
	if err != nil {
		return "", false
	}

	name := getenv("GEMATRIA_INDEX_NAME")
	if name == "" {
		name = "gematria"
	}

	// Invalid names can't match real files — skip validation, just return not found.
	for _, ext := range []string{".db", ".idx"} {
		p := filepath.Join(location, name+ext)
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}
