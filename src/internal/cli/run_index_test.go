package cli

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gematria "github.com/andreswebs/gematria"
	_ "modernc.org/sqlite"
)

// writeTempWordListIndex creates a temp file for the index tests with the given
// content and returns its path. The file is cleaned up automatically by the test.
func writeTempWordListIndex(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "wordlist*.txt")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	_ = f.Close()
	return f.Name()
}

// --- Tracer bullet: no --index-output, GEMATRIA_INDEX_LOCATION set → output goes to that location ---

func TestRun_index_defaultOutput_usesIndexLocation(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\tshalom\tpeace\n")
	indexDir := t.TempDir()
	expectedDB := filepath.Join(indexDir, "gematria.db")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	code := Run([]string{"--index", "--wordlist", wordlistPath}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, expectedDB) {
		t.Errorf("stdout = %q, want mention of %q", stdout, expectedDB)
	}
	if _, err := os.Stat(expectedDB); err != nil {
		t.Errorf("expected DB at %q not found: %v", expectedDB, err)
	}
}

// --- Tracer bullet: --index dispatches to runIndex, produces a SQLite DB ---

func TestRun_index_dispatch_sqlite_tracer(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\tshalom\tpeace\n")
	dbPath := filepath.Join(t.TempDir(), "out.db")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "--wordlist", wordlistPath, "--index-output", dbPath, "--index-format", "sqlite"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	// Summary line must mention the indexed word count and the output path.
	if !strings.Contains(stdout, "1") {
		t.Errorf("stdout = %q, want mention of 1 word", stdout)
	}
	if !strings.Contains(stdout, dbPath) {
		t.Errorf("stdout = %q, want mention of output path %q", stdout, dbPath)
	}

	// Verify the DB is queryable and contains the expected data.
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open generated DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM words`).Scan(&count); err != nil {
		t.Fatalf("query words: %v", err)
	}
	if count != 1 {
		t.Errorf("words count = %d, want 1", count)
	}

	var hebrew string
	if err := db.QueryRow(`SELECT hebrew FROM words WHERE id=1`).Scan(&hebrew); err != nil {
		t.Fatalf("query hebrew: %v", err)
	}
	if hebrew != "שלום" {
		t.Errorf("hebrew = %q, want %q", hebrew, "שלום")
	}

	// The word must have values for all 4 systems.
	if err := db.QueryRow(`SELECT COUNT(*) FROM word_values WHERE word_id=1`).Scan(&count); err != nil {
		t.Fatalf("query word_values: %v", err)
	}
	if count != 4 {
		t.Errorf("word_values count = %d, want 4 (one per system)", count)
	}
}

// --- Missing --wordlist → exit 2 ---

func TestRun_index_missingWordlist_exit2(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (missing --wordlist)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "--wordlist") {
		t.Errorf("stderr = %q, want mention of --wordlist", stderr)
	}
}

// --- Non-existent wordlist file → exit 3 ---

func TestRun_index_wordlistNotFound_exit3(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "--wordlist", "/nonexistent/words.txt"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 3 {
		t.Errorf("exit code = %d, want 3 (file not found)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on file error", stdout)
	}
	if !strings.Contains(stderr, "/nonexistent/words.txt") {
		t.Errorf("stderr = %q, want path in error", stderr)
	}
}

// --- Invalid --index-format → exit 2 ---

func TestRun_index_invalidFormat_exit2(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "--wordlist", wordlistPath, "--index-format", "badformat"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (invalid --index-format)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "badformat") {
		t.Errorf("stderr = %q, want mention of invalid format value", stderr)
	}
}

// --- Default output path: <location>/gematria.db for sqlite ---

func TestRun_index_defaultOutputPath_sqlite(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")
	indexDir := t.TempDir()
	expectedDB := filepath.Join(indexDir, "gematria.db")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	code := Run([]string{"--index", "--wordlist", wordlistPath}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}

	if !strings.Contains(stdout, expectedDB) {
		t.Errorf("stdout = %q, want mention of default output path %q", stdout, expectedDB)
	}

	// The DB file must exist.
	if _, err := os.Stat(expectedDB); err != nil {
		t.Errorf("expected DB at %q not found: %v", expectedDB, err)
	}
}

// --- Default output path: <location>/gematria.idx for index format ---

func TestRun_index_defaultOutputPath_index(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")
	indexDir := t.TempDir()
	expectedIdx := filepath.Join(indexDir, "gematria.idx")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	code := Run([]string{"--index", "--wordlist", wordlistPath, "--index-format", "index"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}

	if !strings.Contains(stdout, expectedIdx) {
		t.Errorf("stdout = %q, want mention of default output path %q", stdout, expectedIdx)
	}

	// The .idx file must exist.
	if _, err := os.Stat(expectedIdx); err != nil {
		t.Errorf("expected idx at %q not found: %v", expectedIdx, err)
	}
}

// --- --index-format index: produces a valid index file readable by NewIndexWordSource ---

func TestRun_index_formatIndex_producesValidIdx(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\tshalom\tpeace\n")
	idxPath := filepath.Join(t.TempDir(), "out.idx")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "--wordlist", wordlistPath, "--index-output", idxPath, "--index-format", "index"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}
	_ = stdout

	// The .idx file must exist and start with the magic comment.
	data, err := os.ReadFile(idxPath)
	if err != nil {
		t.Fatalf("ReadFile %q: %v", idxPath, err)
	}
	if !strings.HasPrefix(string(data), "# gematria-index v1\n") {
		t.Errorf("idx file does not start with magic comment: %q", string(data)[:min(80, len(data))])
	}
	// Must contain the Hebrew word and system.
	if !strings.Contains(string(data), "שלום") {
		t.Errorf("idx file missing Hebrew word: %q", string(data))
	}
	if !strings.Contains(string(data), "hechrachi") {
		t.Errorf("idx file missing system name: %q", string(data))
	}
}

// --- --help includes Indexing: section with --index flag ---

func TestRun_index_help_exit0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--help"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 for --help", code)
	}
	for _, want := range []string{"Indexing:", "--index", "--index-output", "--index-format", "--wordlist"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q in help output", want)
		}
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty for --help", stderr)
	}
}

// --- Round-trip: index SQLite → OpenSQLiteWordSource finds the word ---

func TestRun_index_sqlite_roundTrip(t *testing.T) {
	// שלום (shalom) = 376 in hechrachi
	wordlistPath := writeTempWordListIndex(t, "שלום\tshalom\tpeace\n")
	dbPath := filepath.Join(t.TempDir(), "words.db")

	stdoutW, _ := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "--wordlist", wordlistPath, "--index-output", dbPath}, stdin, stdoutW, stderrW, noenv)
	if code != 0 {
		t.Fatalf("index exit code = %d; stderr = %q", code, readStderr())
	}

	src, err := gematria.OpenSQLiteWordSource(dbPath)
	if err != nil {
		t.Fatalf("OpenSQLiteWordSource: %v", err)
	}
	defer func() {
		if c, ok := src.(interface{ Close() error }); ok {
			_ = c.Close()
		}
	}()

	words, hasMore, err := src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false")
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	if words[0].Hebrew != "שלום" {
		t.Errorf("Hebrew = %q, want %q", words[0].Hebrew, "שלום")
	}
}

// --- Round-trip: index file → NewIndexWordSource finds the word ---

func TestRun_index_indexFile_roundTrip(t *testing.T) {
	// אמת (emet) = 441 in hechrachi
	wordlistPath := writeTempWordListIndex(t, "אמת\temet\ttruth\n")
	idxPath := filepath.Join(t.TempDir(), "words.idx")

	stdoutW, _ := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "--wordlist", wordlistPath, "--index-output", idxPath, "--index-format", "index"}, stdin, stdoutW, stderrW, noenv)
	if code != 0 {
		t.Fatalf("index exit code = %d; stderr = %q", code, readStderr())
	}

	f, err := os.Open(idxPath)
	if err != nil {
		t.Fatalf("open idx: %v", err)
	}
	defer func() { _ = f.Close() }()

	src, err := gematria.NewIndexWordSource(f)
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	words, hasMore, err := src.FindByValue(441, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false")
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	if words[0].Hebrew != "אמת" {
		t.Errorf("Hebrew = %q, want %q", words[0].Hebrew, "אמת")
	}
}

// --- Multi-word list: all words are indexed ---

func TestRun_index_multipleWords(t *testing.T) {
	content := "שלום\tshalom\tpeace\nאמת\temet\ttruth\nאור\tor\tlight\n"
	wordlistPath := writeTempWordListIndex(t, content)
	dbPath := filepath.Join(t.TempDir(), "words.db")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "--wordlist", wordlistPath, "--index-output", dbPath}, stdin, stdoutW, stderrW, noenv)
	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}

	// Summary must mention 3 words.
	if !strings.Contains(stdout, "3") {
		t.Errorf("stdout = %q, want mention of 3 words", stdout)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM words`).Scan(&count); err != nil {
		t.Fatalf("count words: %v", err)
	}
	if count != 3 {
		t.Errorf("words count = %d, want 3", count)
	}
}

// --- Default output: directory is auto-created when it does not exist ---

func TestRun_index_defaultOutput_autoCreatesDirectory(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")
	// Point to a subdirectory that does not yet exist.
	indexDir := filepath.Join(t.TempDir(), "new", "subdir")
	expectedDB := filepath.Join(indexDir, "gematria.db")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	code := Run([]string{"--index", "--wordlist", wordlistPath}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}
	_ = stdout
	if _, err := os.Stat(expectedDB); err != nil {
		t.Errorf("expected DB at %q not found (directory not auto-created?): %v", expectedDB, err)
	}
}

// --- Invalid GEMATRIA_INDEX_NAME (path separator) → exit 2 ---

func TestRun_index_invalidIndexName_exit2(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{
		"GEMATRIA_INDEX_LOCATION": t.TempDir(),
		"GEMATRIA_INDEX_NAME":     "sub/index",
	})
	code := Run([]string{"--index", "--wordlist", wordlistPath}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (invalid GEMATRIA_INDEX_NAME); stderr = %q", code, stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "path separators") {
		t.Errorf("stderr = %q, want mention of 'path separators'", stderr)
	}
}

// --- --index-output flag bypasses GEMATRIA_INDEX_LOCATION entirely ---

func TestRun_index_outputFlagBypassesEnvVars(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")
	explicitOut := filepath.Join(t.TempDir(), "explicit.db")
	envDir := t.TempDir() // would be used if --index-output were absent

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": envDir})
	code := Run([]string{"--index", "--wordlist", wordlistPath, "--index-output", explicitOut}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	// Output must be at the explicit path, not inside envDir.
	if _, err := os.Stat(explicitOut); err != nil {
		t.Errorf("expected DB at explicit path %q not found: %v", explicitOut, err)
	}
	if !strings.Contains(stdout, explicitOut) {
		t.Errorf("stdout = %q, want mention of explicit output path %q", stdout, explicitOut)
	}
	// Nothing should have been written inside envDir.
	entries, _ := os.ReadDir(envDir)
	if len(entries) != 0 {
		t.Errorf("envDir %q contains unexpected files: %v", envDir, entries)
	}
}

// --- Bad --index-output path → exit 3 ---

func TestRun_index_badOutputPath_exit3(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// Point --index-output at a directory that does not exist.
	code := Run([]string{"--index", "--wordlist", wordlistPath, "--index-output", "/nonexistent/dir/out.db"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 3 {
		t.Errorf("exit code = %d, want 3 (bad output path); stderr = %q", code, stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on file error", stdout)
	}
	// MkdirAll fails first — check for the directory in the error.
	if !strings.Contains(stderr, "/nonexistent/dir") {
		t.Errorf("stderr = %q, want directory path in error message", stderr)
	}
}

// --- Conflict: --index and --find are mutually exclusive → exit 2 ---

func TestRun_index_conflict_findMutuallyExclusive(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "--find", "376", "--wordlist", wordlistPath}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (--index and --find mutually exclusive)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "--index") || !strings.Contains(stderr, "--find") {
		t.Errorf("stderr = %q, want mention of both --index and --find", stderr)
	}
}

// --- Conflict: --index and --transliterate are mutually exclusive → exit 2 ---

func TestRun_index_conflict_transliterateMutuallyExclusive(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "-t", "--wordlist", wordlistPath}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (--index and -t mutually exclusive)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "--index") {
		t.Errorf("stderr = %q, want mention of --index", stderr)
	}
}

// --- Conflict: --index-output without --index → exit 2 ---

func TestRun_index_conflict_indexOutputWithoutIndex(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index-output", "foo.db"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (--index-output requires --index)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "--index-output") {
		t.Errorf("stderr = %q, want mention of --index-output", stderr)
	}
}

// --- Conflict: --index-format without --index → exit 2 ---

func TestRun_index_conflict_indexFormatWithoutIndex(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index-format", "index"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (--index-format requires --index)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "--index-format") {
		t.Errorf("stderr = %q, want mention of --index-format", stderr)
	}
}

// --- Conflict: --index with positional argument → exit 2 ---

func TestRun_index_conflict_positionalArgs(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--index", "--wordlist", wordlistPath, "shalom"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (--index does not accept positional args)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "--index") {
		t.Errorf("stderr = %q, want mention of --index", stderr)
	}
}

// --- GEMATRIA_WORDLIST resolves the word list when --wordlist is absent ---

func TestRun_index_envWordlist_resolution(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\tshalom\tpeace\n")
	indexDir := t.TempDir()
	expectedDB := filepath.Join(indexDir, "gematria.db")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{
		"GEMATRIA_WORDLIST":       wordlistPath,
		"GEMATRIA_INDEX_LOCATION": indexDir,
	})
	// No --wordlist flag; relies on GEMATRIA_WORDLIST.
	code := Run([]string{"--index"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, expectedDB) {
		t.Errorf("stdout = %q, want mention of %q", stdout, expectedDB)
	}
	if _, err := os.Stat(expectedDB); err != nil {
		t.Errorf("expected DB at %q not found: %v", expectedDB, err)
	}
}

// --- Old subcommand syntax regression: "index" as Latin input → exit 1, NOT 0 ---

func TestRun_index_oldSubcommandSyntax_regression(t *testing.T) {
	wordlistPath := writeTempWordListIndex(t, "שלום\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// Old syntax: "index" is now treated as an unknown Latin letter name.
	code := Run([]string{"index", "--wordlist", wordlistPath}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	// "index" is not a valid letter name → exit 1 (input error), not 0.
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (old subcommand treated as unknown input); stderr = %q stdout = %q", code, stderr, stdout)
	}
}
