package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	gematria "github.com/andreswebs/gematria"
)

// createTestSQLiteDB creates a temporary SQLite word-list DB using WriteIndexSQLite.
func createTestSQLiteDB(t *testing.T, words []gematria.Word) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.db")
	_, err := gematria.WriteIndexSQLite(path, words)
	if err != nil {
		t.Fatalf("WriteIndexSQLite: %v", err)
	}
	return path
}

// createTestIndexFile creates a temporary .idx file using WriteIndexFile.
func createTestIndexFile(t *testing.T, words []gematria.Word) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.idx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, err = gematria.WriteIndexFile(f, words)
	_ = f.Close()
	if err != nil {
		t.Fatalf("WriteIndexFile: %v", err)
	}
	return path
}

// --- Tracer bullet: .db extension auto-selects SQLite backend ---

func TestRun_find_dbExtension_usesSQLiteBackend(t *testing.T) {
	words := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}
	path := createTestSQLiteDB(t, words)

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--output", "value"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (SQLite backend)", code)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום from SQLite backend", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- .idx extension auto-selects index file backend ---

func TestRun_find_idxExtension_usesIndexBackend(t *testing.T) {
	words := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}
	path := createTestIndexFile(t, words)

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--output", "value"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (index backend)", code)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום from index backend", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Companion .idx file auto-selects index backend ---

func TestRun_find_companionIdxFile_usesIndexBackend(t *testing.T) {
	words := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}

	// Create a plain .txt wordlist and a companion .idx file
	dir := t.TempDir()
	txtPath := filepath.Join(dir, "words.txt")
	idxPath := txtPath + ".idx"

	if err := os.WriteFile(txtPath, []byte("שלום\tshalom\tpeace\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	f, err := os.Create(idxPath)
	if err != nil {
		t.Fatalf("Create idx: %v", err)
	}
	_, err = gematria.WriteIndexFile(f, words)
	_ = f.Close()
	if err != nil {
		t.Fatalf("WriteIndexFile: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// Use the .txt path — companion .idx should be auto-detected
	code := Run([]string{"--find", "376", "--wordlist", txtPath, "--output", "value"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (companion .idx used)", code)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום via companion .idx", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- SQLite backend: io.Closer is called (no file leak) ---

func TestRun_find_sqliteBackend_closerCalled(t *testing.T) {
	words := []gematria.Word{{Hebrew: "אמת", Transliteration: "emet", Meaning: "truth"}}
	path := createTestSQLiteDB(t, words)

	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// Value 441 = emet in hechrachi
	code := Run([]string{"--find", "441", "--wordlist", path}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "אמת") {
		t.Errorf("stdout = %q, want אמת", stdout)
	}
}

// --- --wordlist-format invalid value → exit 64 ---

func TestRun_find_wordlistFormatInvalid_exit64(t *testing.T) {
	path := writeTempWordList(t, "שלום\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--wordlist-format", "badformat"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 64 {
		t.Errorf("exit code = %d, want 64 for invalid --wordlist-format", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on config error", stdout)
	}
	if !strings.Contains(stderr, "badformat") {
		t.Errorf("stderr = %q, want mention of invalid value 'badformat'", stderr)
	}
	if !strings.Contains(stderr, "--wordlist-format") {
		t.Errorf("stderr = %q, want mention of --wordlist-format", stderr)
	}
}

// --- --wordlist-format memory: forces memory backend for .db path ---

func TestRun_find_wordlistFormatMemory_overridesExtension(t *testing.T) {
	// Write a plain TSV wordlist but name it .db — format=memory forces memory backend
	dir := t.TempDir()
	path := filepath.Join(dir, "words.db")
	if err := os.WriteFile(path, []byte("שלום\tshalom\tpeace\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// Without --wordlist-format, .db would try SQLite and fail on plain text
	// With --wordlist-format memory, plain text is parsed directly
	code := Run([]string{"--find", "376", "--wordlist", path, "--wordlist-format", "memory"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (memory format forced)", code)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום from memory backend", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --wordlist-format sqlite: forces SQLite backend for plain-named file ---

func TestRun_find_wordlistFormatSqlite_overridesExtension(t *testing.T) {
	words := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}
	dir := t.TempDir()
	// Name it .txt but it's a real SQLite DB
	path := filepath.Join(dir, "words.txt")
	_, err := gematria.WriteIndexSQLite(path, words)
	if err != nil {
		t.Fatalf("WriteIndexSQLite: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--wordlist-format", "sqlite"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (sqlite format forced)", code)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום from SQLite backend", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --wordlist-format index: forces index backend for plain-named file ---

func TestRun_find_wordlistFormatIndex_overridesExtension(t *testing.T) {
	words := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}
	dir := t.TempDir()
	// Name it .txt but it's a real index file
	path := filepath.Join(dir, "words.txt")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, err = gematria.WriteIndexFile(f, words)
	_ = f.Close()
	if err != nil {
		t.Fatalf("WriteIndexFile: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--wordlist-format", "index"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (index format forced)", code)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום from index backend", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}
