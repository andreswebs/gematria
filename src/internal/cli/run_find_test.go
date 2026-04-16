package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	gematria "github.com/andreswebs/gematria"
)

// writeTempWordList creates a temp file in t.TempDir() with the given content
// and returns its path. The file is cleaned up automatically by the test.
func writeTempWordList(t *testing.T, content string) string {
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

// --- Tracer bullet: --wordlist flag overrides GEMATRIA_WORDLIST env var ---

func TestRun_find_wordlistFlagOverridesEnv(t *testing.T) {
	// env file contains אמת (emet=441), flag file contains שלום (shalom=376)
	envFile := writeTempWordList(t, "אמת\temet\ttruth\n")
	flagFile := writeTempWordList(t, "שלום\tshalom\tpeace\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_WORDLIST": envFile})
	code := Run([]string{"--find", "376", "--wordlist", flagFile, "--output", "value"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום (flag file used, not env file)", stdout)
	}
	if strings.Contains(stdout, "אמת") {
		t.Errorf("stdout = %q, must not contain אמת (env file must not be used)", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --limit restricts results; hasMore indicator appears in line output ---

func TestRun_find_limitFlag_restrictsResults(t *testing.T) {
	// 3 entries all with value 376 (שלום); --limit 1 → 1 result + hasMore
	content := "שלום\tshalom\tpeace\nשלום\tshalom2\thello\nשלום\tshalom3\tgreeting\n"
	path := writeTempWordList(t, content)

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--limit", "1", "--output", "line"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "--limit") {
		t.Errorf("stdout = %q, want hasMore indicator mentioning --limit", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- GEMATRIA_LIMIT env var used when --limit flag absent ---

func TestRun_find_limitEnvVar_restrictsResults(t *testing.T) {
	// 2 entries both value 376; GEMATRIA_LIMIT=1 → 1 result, hasMore
	content := "שלום\tshalom\tpeace\nשלום\tshalom2\thello\n"
	path := writeTempWordList(t, content)

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_LIMIT": "1"})
	code := Run([]string{"--find", "376", "--wordlist", path, "--output", "line"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "--limit") {
		t.Errorf("stdout = %q, want hasMore indicator (GEMATRIA_LIMIT=1 applied)", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- GEMATRIA_LIMIT=invalid with --find → exit 2 ---

func TestRun_find_limitEnvVar_invalid_withFind_exit2(t *testing.T) {
	path := writeTempWordList(t, "שלום\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_LIMIT": "notanumber"})
	code := Run([]string{"--find", "376", "--wordlist", path}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 for invalid GEMATRIA_LIMIT with --find", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on config error", stdout)
	}
	if !strings.Contains(stderr, "GEMATRIA_LIMIT") {
		t.Errorf("stderr = %q, want mention of GEMATRIA_LIMIT", stderr)
	}
}

// --- GEMATRIA_LIMIT=invalid without --find → exit 0 (lazy validation) ---

func TestRun_find_limitEnvVar_invalid_withoutFind_exit0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// GEMATRIA_LIMIT is invalid, but --find is not set — env var must not be validated
	getenv := envWith(map[string]string{"GEMATRIA_LIMIT": "notanumber"})
	code := Run([]string{"--output", "value", "א"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (GEMATRIA_LIMIT not validated without --find)", code)
	}
	if stdout != "1\n" {
		t.Errorf("stdout = %q, want %q", stdout, "1\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --output value: one Hebrew word per line, no other content ---

func TestRun_find_outputValue_hebrewPerLine(t *testing.T) {
	path := writeTempWordList(t, "שלום\tshalom\tpeace\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--output", "value"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	// value format must emit just the Hebrew word, no transliteration or meaning
	if stdout != "שלום\n" {
		t.Errorf("stdout = %q, want %q (bare Hebrew word)", stdout, "שלום\n")
	}
	if strings.Contains(stdout, "shalom") {
		t.Errorf("stdout = %q, must not contain transliteration in value format", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --output line: Hebrew word with RTL marks and transliteration ---

func TestRun_find_outputLine_RTLAndTranslit(t *testing.T) {
	path := writeTempWordList(t, "שלום\tshalom\tpeace\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--output", "line"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	// Hebrew word must be RTL-wrapped
	if !strings.Contains(stdout, rtlMark+"שלום"+ltrMark) {
		t.Errorf("stdout = %q, want RTL-wrapped שלום", stdout)
	}
	// Transliteration must appear
	if !strings.Contains(stdout, "shalom") {
		t.Errorf("stdout = %q, want transliteration 'shalom'", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --output card: multi-line output with numbered entries ---

func TestRun_find_outputCard_numberedEntries(t *testing.T) {
	// Both entries match value 1 (aleph in hechrachi).
	path := writeTempWordList(t, "א\taleph\toxhead\nא\taleph2\tox\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "1", "--wordlist", path, "--output", "card"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "1.") {
		t.Errorf("stdout = %q, want entry number '1.'", stdout)
	}
	if !strings.Contains(stdout, "2.") {
		t.Errorf("stdout = %q, want entry number '2.'", stdout)
	}
	if !strings.Contains(stdout, "aleph") {
		t.Errorf("stdout = %q, want transliteration 'aleph'", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --output json with error: stderr is JSON error object, stdout empty ---

func TestRun_find_outputJSON_errorIsJSONOnStderr(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// --find without --wordlist triggers exit 2; with --output json the error must be JSON on stderr.
	// Point GEMATRIA_INDEX_LOCATION to an empty dir so auto-discovery finds nothing.
	emptyDir := t.TempDir()
	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": emptyDir})
	code := Run([]string{"--find", "376", "--output", "json"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	// stderr must be a JSON object with an "error" field
	if !strings.Contains(stderr, `"error"`) {
		t.Errorf("stderr = %q, want JSON error object with 'error' field", stderr)
	}
	// Quick validity check: starts with '{' and ends with '}\n'
	trimmed := strings.TrimSpace(stderr)
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		t.Errorf("stderr = %q, want JSON object", stderr)
	}
}

// --- --output json: hasMore field is present in lookup response ---

func TestRun_find_outputJSON_hasMoreField(t *testing.T) {
	// 2 entries both value 376; --limit 1 → hasMore=true
	content := "שלום\tshalom\tpeace\nשלום\tshalom2\thello\n"
	path := writeTempWordList(t, content)

	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--limit", "1", "--output", "json"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, `"hasMore":true`) {
		t.Errorf("stdout = %q, want JSON with hasMore:true", stdout)
	}
	if !strings.Contains(stdout, `"results"`) {
		t.Errorf("stdout = %q, want JSON with results field", stdout)
	}
}

// --- GEMATRIA_WORDLIST="" (empty string) with --find → exit 2 ---

func TestRun_find_emptyWordlistEnv_exit2(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// Point GEMATRIA_INDEX_LOCATION to an empty dir so auto-discovery finds nothing.
	emptyDir := t.TempDir()
	getenv := envWith(map[string]string{
		"GEMATRIA_WORDLIST":       "",
		"GEMATRIA_INDEX_LOCATION": emptyDir,
	})
	code := Run([]string{"--find", "376"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (empty GEMATRIA_WORDLIST treated as absent)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on config error", stdout)
	}
	if stderr == "" {
		t.Errorf("stderr empty, want error message")
	}
}

// --- --find auto-discovers .db at GEMATRIA_INDEX_LOCATION ---
//
// This is the tracer bullet: proves the auto-discovery path works end-to-end.

func TestRun_find_autoDiscoversDB(t *testing.T) {
	// Build a .db file named gematria.db inside a temp dir.
	indexDir := t.TempDir()
	dbPath := filepath.Join(indexDir, "gematria.db")
	words := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}
	_, err := gematria.WriteIndexSQLite(dbPath, words)
	if err != nil {
		t.Fatalf("WriteIndexSQLite: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	code := Run([]string{"--find", "376", "--output", "value"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (auto-discovered .db); stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום from auto-discovered .db", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --find auto-discovers .idx when no .db exists ---

func TestRun_find_autoDiscoversIdx_whenNoDB(t *testing.T) {
	indexDir := t.TempDir()
	idxPath := filepath.Join(indexDir, "gematria.idx")
	words := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}
	f, err := os.Create(idxPath)
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

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	code := Run([]string{"--find", "376", "--output", "value"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (auto-discovered .idx); stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום from auto-discovered .idx", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --find prefers .db over .idx when both exist at default location ---

func TestRun_find_autoDiscovery_prefersDBOverIdx(t *testing.T) {
	indexDir := t.TempDir()

	// .db has אמת (441); .idx has שלום (376). Looking up 441 should use the .db.
	dbWords := []gematria.Word{{Hebrew: "אמת", Transliteration: "emet", Meaning: "truth"}}
	dbPath := filepath.Join(indexDir, "gematria.db")
	if _, err := gematria.WriteIndexSQLite(dbPath, dbWords); err != nil {
		t.Fatalf("WriteIndexSQLite: %v", err)
	}

	idxWords := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}
	idxPath := filepath.Join(indexDir, "gematria.idx")
	f, err := os.Create(idxPath)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, err = gematria.WriteIndexFile(f, idxWords)
	_ = f.Close()
	if err != nil {
		t.Fatalf("WriteIndexFile: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": indexDir})
	// Look up 441 — only present in the .db, not in the .idx.
	code := Run([]string{"--find", "441", "--output", "value"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (.db preferred); stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "אמת") {
		t.Errorf("stdout = %q, want אמת (.db used, not .idx)", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --find errors with new message when no wordlist and no default index ---

func TestRun_find_noDefaultIndex_errorMentionsGematriaIndex(t *testing.T) {
	emptyDir := t.TempDir()

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": emptyDir})
	code := Run([]string{"--find", "376"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (no index found)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "gematria --index") {
		t.Errorf("stderr = %q, want error mentioning 'gematria --index'", stderr)
	}
}

// --- GEMATRIA_WORDLIST env takes precedence over auto-discovery ---

func TestRun_find_wordlistEnv_precedenceOverAutoDiscovery(t *testing.T) {
	// Index dir has אמת (441); env var points to file with שלום (376).
	indexDir := t.TempDir()
	dbWords := []gematria.Word{{Hebrew: "אמת", Transliteration: "emet", Meaning: "truth"}}
	dbPath := filepath.Join(indexDir, "gematria.db")
	if _, err := gematria.WriteIndexSQLite(dbPath, dbWords); err != nil {
		t.Fatalf("WriteIndexSQLite: %v", err)
	}

	envFile := writeTempWordList(t, "שלום\tshalom\tpeace\n")

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{
		"GEMATRIA_WORDLIST":       envFile,
		"GEMATRIA_INDEX_LOCATION": indexDir,
	})
	// Look up 376 — only in the env var file.
	code := Run([]string{"--find", "376", "--output", "value"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום (GEMATRIA_WORDLIST used, not auto-discovery)", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}
