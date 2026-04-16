package cli

import (
	"os"
	"strings"
	"testing"
)

// makeStdinPipe creates a readable *os.File pipe pre-loaded with content.
// The write end is closed immediately so reads return EOF after the content.
func makeStdinPipe(t *testing.T, content string) *os.File {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	_, _ = w.WriteString(content)
	_ = w.Close()
	t.Cleanup(func() { _ = r.Close() })
	return r
}

// --- Tracer bullet: positional arg computes correctly ---

func TestRun_positionalArg_exit0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "1") {
		t.Errorf("stdout = %q, want it to contain the computed value '1'", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Config error: bad --output → exit 2, error on stderr ---

func TestRun_badOutput_exit2(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--output", "table", "א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 for invalid --output", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on config error", stdout)
	}
	if !strings.Contains(stderr, "table") {
		t.Errorf("stderr = %q, want it to mention invalid value 'table'", stderr)
	}
}

// --- --help → usage on stdout, exit 0 ---

func TestRun_help_exit0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--help"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 for --help", code)
	}
	if !strings.Contains(stdout, "Usage: gematria") {
		t.Errorf("stdout = %q, want usage line", stdout)
	}
	// Must include env var documentation
	if !strings.Contains(stdout, "GEMATRIA_MISPAR") {
		t.Errorf("stdout missing GEMATRIA_MISPAR env var doc: %q", stdout)
	}
	if !strings.Contains(stdout, "GEMATRIA_OUTPUT") {
		t.Errorf("stdout missing GEMATRIA_OUTPUT env var doc: %q", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty on --help", stderr)
	}
}

// --- -h is short form for --help ---

func TestRun_shortHelp_exit0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"-h"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 for -h", code)
	}
	if !strings.Contains(stdout, "Usage: gematria") {
		t.Errorf("-h: stdout = %q, want usage line", stdout)
	}
}

// --- --version → human form keeps the v prefix, exit 0 ---

func TestRun_version_exit0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--version"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 for --version", code)
	}
	if !strings.HasPrefix(stdout, "gematria v") {
		t.Errorf("stdout = %q, want it to start with 'gematria v' (human form keeps v prefix)", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty on --version", stderr)
	}
}

// --- --version --output json → JSON form strips the v prefix, exit 0 ---

func TestRun_versionJSON_exit0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--version", "--output", "json"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 for --version --output json", code)
	}
	if !strings.Contains(stdout, `"version"`) {
		t.Errorf("stdout = %q, want JSON with version field", stdout)
	}
	// The JSON value must not carry the v prefix — see writeVersion comment.
	if strings.Contains(stdout, `"version":"v`) {
		t.Errorf("stdout = %q, JSON version must not have 'v' prefix (canonical SemVer)", stdout)
	}
}

// --- Invalid Hebrew positional arg → exit 1, error on stderr, nothing on stdout ---

func TestRun_invalidHebrew_exit1(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"xyz"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 1 {
		t.Errorf("exit code = %d, want 1 for unknown name", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if stderr == "" {
		t.Errorf("stderr empty, want error message")
	}
}

// --- InvalidSystemError via env var during compute → exit 2 ---

func TestRun_invalidSystemEnv_exit2(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// GEMATRIA_MISPAR with invalid value passes parseConfig (lazy) but fails Compute
	getenv := envWith(map[string]string{"GEMATRIA_MISPAR": "nonexistent"})
	code := Run([]string{"א"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 for InvalidSystemError", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if stderr == "" {
		t.Errorf("stderr empty, want error message")
	}
}

// --- Multiple positional args → all computed, exit 0 ---

func TestRun_multipleArgs_allComputed(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"א", "ב", "ג"}, stdin, stdoutW, stderrW, noenv) // aleph=1, bet=2, gimel=3

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("stdout has %d lines, want 3: %q", len(lines), stdout)
	}
}

// --- First positional arg error → stops, returns 1 ---

func TestRun_firstArgError_exit1_noFurtherOutput(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// "bad" is invalid; "א" after it should not be processed
	code := Run([]string{"bad", "א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty (error stops further processing)", stdout)
	}
	if stderr == "" {
		t.Errorf("stderr empty, want error message")
	}
}

// --- Stdin batch: no args, non-TTY stdin → reads from pipe ---

func TestRun_stdinBatch_exit0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	// Pipe is not a TTY, so Run() must enter batch mode
	stdin := makeStdinPipe(t, "א\nב\n")

	code := Run([]string{}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 for valid stdin batch", code)
	}
	if stdout == "" {
		t.Errorf("stdout empty, want computed results")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Stdin batch partial failure → exit 4 ---

func TestRun_stdinBatch_partialFailure_exit4(t *testing.T) {
	stdoutW, _ := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "א\nbad_latin\n")

	code := Run([]string{}, stdin, stdoutW, stderrW, noenv)

	if code != 4 {
		t.Errorf("exit code = %d, want 4 for partial batch failure", code)
	}
}

// --- Reverse lookup: --find success ---

func TestRun_find_success_exit0(t *testing.T) {
	// שלום (shalom) = 376 in hechrachi
	dir := t.TempDir()
	path := dir + "/words.txt"
	if err := os.WriteFile(path, []byte("שלום\tshalom\tpeace\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want it to contain 'שלום'", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Reverse lookup: --find without wordlist → exit 2 ---

func TestRun_find_noWordlist_exit2(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "--wordlist") {
		t.Errorf("stderr = %q, want mention of --wordlist", stderr)
	}
	if !strings.Contains(stderr, "GEMATRIA_WORDLIST") {
		t.Errorf("stderr = %q, want mention of GEMATRIA_WORDLIST", stderr)
	}
}

// --- Reverse lookup: --find with non-existent file → exit 3 ---

func TestRun_find_fileNotFound_exit3(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", "/nonexistent/words.txt"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "/nonexistent/words.txt") {
		t.Errorf("stderr = %q, want path in error message", stderr)
	}
}

// --- Reverse lookup: GEMATRIA_WORDLIST env var used when --wordlist absent ---

func TestRun_find_wordlistFromEnv_exit0(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/words.txt"
	if err := os.WriteFile(path, []byte("שלום\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_WORDLIST": path})
	code := Run([]string{"--find", "376"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "שלום") {
		t.Errorf("stdout = %q, want שלום", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Reverse lookup: --find with no matching words → exit 0, no-results message ---

func TestRun_find_noResults_exit0(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/words.txt"
	if err := os.WriteFile(path, []byte("שלום\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// value 999 won't match שלום=376
	code := Run([]string{"--find", "999", "--wordlist", path}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	_ = stdout // content depends on formatter; no-results indicator acceptable
}

// --- Reverse lookup: --find --output json uses NewFormatterWithLookup context ---

func TestRun_find_outputJSON_containsValueField(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/words.txt"
	if err := os.WriteFile(path, []byte("שלום\tshalom\tpeace\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--find", "376", "--wordlist", path, "--output", "json"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, `"value":376`) {
		t.Errorf("stdout = %q, want JSON with value:376", stdout)
	}
	if !strings.Contains(stdout, `"system":"hechrachi"`) {
		t.Errorf("stdout = %q, want JSON with system field", stdout)
	}
}

// --- Output format respected for positional args ---

func TestRun_outputJSON_positionalArg(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--output", "json", "א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, `"total"`) {
		t.Errorf("stdout = %q, want JSON with 'total' field", stdout)
	}
}
