package cli

import (
	"os"
	"strings"
	"testing"
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

	// --find without --wordlist triggers exit 2; with --output json the error must be JSON on stderr
	code := Run([]string{"--find", "376", "--output", "json"}, stdin, stdoutW, stderrW, noenv)

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

	getenv := envWith(map[string]string{"GEMATRIA_WORDLIST": ""})
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
