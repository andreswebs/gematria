package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Tracer bullet: no env vars → XDG default path ---

func TestResolveIndexPath_noEnvVars_returnsXDGDefault(t *testing.T) {
	home, err := userHomeDir()
	if err != nil {
		t.Fatalf("userHomeDir: %v", err)
	}
	want := filepath.Join(home, ".local", "share", "gematria", "gematria.db")

	got, err := resolveIndexPath("sqlite", noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- GEMATRIA_INDEX_LOCATION overrides XDG_DATA_HOME ---

func TestResolveIndexPath_indexLocationEnv(t *testing.T) {
	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": "/custom/dir"})
	want := "/custom/dir/gematria.db"

	got, err := resolveIndexPath("sqlite", getenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- XDG_DATA_HOME env var is used when GEMATRIA_INDEX_LOCATION is unset ---

func TestResolveIndexPath_xdgDataHome(t *testing.T) {
	getenv := envWith(map[string]string{"XDG_DATA_HOME": "/xdg/data"})
	want := "/xdg/data/gematria/gematria.db"

	got, err := resolveIndexPath("sqlite", getenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- GEMATRIA_INDEX_NAME controls the filename ---

func TestResolveIndexPath_indexNameEnv(t *testing.T) {
	getenv := envWith(map[string]string{
		"GEMATRIA_INDEX_LOCATION": "/mydir",
		"GEMATRIA_INDEX_NAME":     "words",
	})
	want := "/mydir/words.db"

	got, err := resolveIndexPath("sqlite", getenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- format "index" → .idx extension ---

func TestResolveIndexPath_formatIndex_idxExtension(t *testing.T) {
	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": "/mydir"})
	want := "/mydir/gematria.idx"

	got, err := resolveIndexPath("index", getenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- Dots allowed in GEMATRIA_INDEX_NAME ---

func TestResolveIndexPath_dotInName_allowed(t *testing.T) {
	getenv := envWith(map[string]string{
		"GEMATRIA_INDEX_LOCATION": "/mydir",
		"GEMATRIA_INDEX_NAME":     "my.index",
	})
	want := "/mydir/my.index.db"

	got, err := resolveIndexPath("sqlite", getenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- Path separator in GEMATRIA_INDEX_NAME → error ---

func TestResolveIndexPath_forwardSlashInName_error(t *testing.T) {
	getenv := envWith(map[string]string{
		"GEMATRIA_INDEX_LOCATION": "/mydir",
		"GEMATRIA_INDEX_NAME":     "sub/index",
	})

	_, err := resolveIndexPath("sqlite", getenv)
	if err == nil {
		t.Fatal("expected error for name containing '/'")
	}
	if !strings.Contains(err.Error(), "path separators") {
		t.Errorf("error = %q, want mention of 'path separators'", err)
	}
}

func TestResolveIndexPath_backslashInName_error(t *testing.T) {
	getenv := envWith(map[string]string{
		"GEMATRIA_INDEX_LOCATION": "/mydir",
		"GEMATRIA_INDEX_NAME":     `sub\index`,
	})

	_, err := resolveIndexPath("sqlite", getenv)
	if err == nil {
		t.Fatal("expected error for name containing '\\'")
	}
	if !strings.Contains(err.Error(), "path separators") {
		t.Errorf("error = %q, want mention of 'path separators'", err)
	}
}

// --- discoverDefaultIndex: .db exists → returns (path, true) ---

func TestDiscoverDefaultIndex_dbExists_returnsPath(t *testing.T) {
	indexDir := t.TempDir()
	dbPath := indexDir + "/gematria.db"
	if err := os.WriteFile(dbPath, []byte("placeholder"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	got, found := discoverDefaultIndex(getenv)

	if !found {
		t.Fatal("found = false, want true")
	}
	if got != dbPath {
		t.Errorf("got %q, want %q", got, dbPath)
	}
}

// --- discoverDefaultIndex: only .idx exists → returns (.idx path, true) ---

func TestDiscoverDefaultIndex_idxExistsNoDb_returnsIdx(t *testing.T) {
	indexDir := t.TempDir()
	idxPath := indexDir + "/gematria.idx"
	if err := os.WriteFile(idxPath, []byte("placeholder"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	got, found := discoverDefaultIndex(getenv)

	if !found {
		t.Fatal("found = false, want true")
	}
	if got != idxPath {
		t.Errorf("got %q, want %q", got, idxPath)
	}
}

// --- discoverDefaultIndex: both .db and .idx exist → prefers .db ---

func TestDiscoverDefaultIndex_bothExist_prefersDb(t *testing.T) {
	indexDir := t.TempDir()
	dbPath := indexDir + "/gematria.db"
	idxPath := indexDir + "/gematria.idx"
	for _, p := range []string{dbPath, idxPath} {
		if err := os.WriteFile(p, []byte("placeholder"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	got, found := discoverDefaultIndex(getenv)

	if !found {
		t.Fatal("found = false, want true")
	}
	if got != dbPath {
		t.Errorf("got %q, want %q (should prefer .db over .idx)", got, dbPath)
	}
}

// --- discoverDefaultIndex: neither exists → returns ("", false) ---

func TestDiscoverDefaultIndex_neitherExists_notFound(t *testing.T) {
	indexDir := t.TempDir() // empty directory — no index files

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	got, found := discoverDefaultIndex(getenv)

	if found {
		t.Errorf("found = true, want false (got %q)", got)
	}
	if got != "" {
		t.Errorf("path = %q, want empty string", got)
	}
}

// --- discoverDefaultIndex: GEMATRIA_INDEX_NAME controls filename ---

func TestDiscoverDefaultIndex_customName(t *testing.T) {
	indexDir := t.TempDir()
	dbPath := indexDir + "/words.db"
	if err := os.WriteFile(dbPath, []byte("placeholder"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	getenv := envWith(map[string]string{
		"GEMATRIA_INDEX_LOCATION": indexDir,
		"GEMATRIA_INDEX_NAME":     "words",
	})
	got, found := discoverDefaultIndex(getenv)

	if !found {
		t.Fatal("found = false, want true")
	}
	if got != dbPath {
		t.Errorf("got %q, want %q", got, dbPath)
	}
}

// --- discoverDefaultIndex: invalid GEMATRIA_INDEX_NAME → found=false, no panic ---

func TestDiscoverDefaultIndex_invalidName_notFound(t *testing.T) {
	indexDir := t.TempDir()

	getenv := envWith(map[string]string{
		"GEMATRIA_INDEX_LOCATION": indexDir,
		"GEMATRIA_INDEX_NAME":     "sub/index", // path separator — invalid
	})
	got, found := discoverDefaultIndex(getenv)

	// Invalid names can't match real files; function must not panic or error.
	if found {
		t.Errorf("found = true, want false for invalid name (got %q)", got)
	}
}

// --- discoverDefaultIndex: XDG_DATA_HOME used when GEMATRIA_INDEX_LOCATION unset ---

func TestDiscoverDefaultIndex_xdgDataHome(t *testing.T) {
	xdgBase := t.TempDir()
	indexDir := filepath.Join(xdgBase, "gematria")
	if err := os.MkdirAll(indexDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	dbPath := filepath.Join(indexDir, "gematria.db")
	if err := os.WriteFile(dbPath, []byte("placeholder"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	getenv := envWith(map[string]string{"XDG_DATA_HOME": xdgBase})
	got, found := discoverDefaultIndex(getenv)

	if !found {
		t.Fatal("found = false, want true")
	}
	if got != dbPath {
		t.Errorf("got %q, want %q", got, dbPath)
	}
}

// --- GEMATRIA_INDEX_LOCATION takes precedence over XDG_DATA_HOME ---

func TestResolveIndexPath_indexLocationTakesPrecedenceOverXDG(t *testing.T) {
	getenv := envWith(map[string]string{
		"GEMATRIA_INDEX_LOCATION": "/override",
		"XDG_DATA_HOME":           "/xdg/data",
	})
	want := "/override/gematria.db"

	got, err := resolveIndexPath("sqlite", getenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
