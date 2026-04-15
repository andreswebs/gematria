package cli

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	gematria "github.com/andreswebs/gematria"
)

// pipeCapture creates an os.Pipe pair and returns a function to read all output.
func pipeCapture(t *testing.T) (w *os.File, read func() string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	read = func() string {
		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		return buf.String()
	}
	return w, read
}

func makeScanner(lines ...string) *bufio.Scanner {
	return bufio.NewScanner(strings.NewReader(strings.Join(lines, "\n")))
}

func successCompute(input string) (gematria.Result, error) {
	return gematria.Compute(input, gematria.Hechrachi)
}

// alwaysError returns an InvalidCharError for every input.
func alwaysError(_ string) (gematria.Result, error) {
	return gematria.Result{}, &gematria.InvalidCharError{Char: 'x', Position: 0}
}

// sysError returns an InvalidSystemError for every input.
func sysError(_ string) (gematria.Result, error) {
	return gematria.Result{}, &gematria.InvalidSystemError{Name: "bad", Valid: gematria.ValidSystems()}
}

// --- Tracer bullet: single successful line ---

func TestProcessBatch_allSuccess_returns0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)

	formatter := NewFormatter("value", false, false)
	scanner := makeScanner("א")

	code := processBatch(scanner, successCompute, formatter, stdoutW, stderrW, false)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "1") {
		t.Errorf("stdout = %q, want it to contain the computed value", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Single line fails → exit 1 ---

func TestProcessBatch_allErrors_returns1(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)

	formatter := NewFormatter("line", false, false)
	scanner := makeScanner("not-hebrew")

	code := processBatch(scanner, alwaysError, formatter, stdoutW, stderrW, false)

	stdout := readStdout()
	stderr := readStderr()

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on all-error", stdout)
	}
	if stderr == "" {
		t.Errorf("stderr empty, want error message")
	}
}

// --- Mixed success and failure → exit 4 ---

func TestProcessBatch_mixed_returns4(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)

	formatter := NewFormatter("value", false, false)

	// Line 1: valid Hebrew "א" → succeeds
	// Line 2: "bad" → fails
	callCount := 0
	mixed := func(input string) (gematria.Result, error) {
		callCount++
		if callCount == 1 {
			return successCompute(input)
		}
		return alwaysError(input)
	}

	scanner := makeScanner("א", "bad")
	code := processBatch(scanner, mixed, formatter, stdoutW, stderrW, false)

	stdout := readStdout()
	stderr := readStderr()

	if code != 4 {
		t.Errorf("exit code = %d, want 4 (partial success)", code)
	}
	if stdout == "" {
		t.Errorf("stdout empty, want successful result on stdout")
	}
	if stderr == "" {
		t.Errorf("stderr empty, want error on stderr")
	}
}

// --- Empty input → exit 0 ---

func TestProcessBatch_emptyInput_returns0(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)

	formatter := NewFormatter("value", false, false)
	scanner := makeScanner() // empty

	code := processBatch(scanner, successCompute, formatter, stdoutW, stderrW, false)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 for empty input", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- failEarly stops on first error ---

func TestProcessBatch_failEarly_stopsOnFirstError(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)

	formatter := NewFormatter("value", false, false)

	processedLines := 0
	counted := func(input string) (gematria.Result, error) {
		processedLines++
		return alwaysError(input)
	}

	// 3 lines; with failEarly, should stop after the first error
	scanner := makeScanner("bad1", "bad2", "bad3")
	code := processBatch(scanner, counted, formatter, stdoutW, stderrW, true)

	readStdout()

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if processedLines != 1 {
		t.Errorf("processed %d lines, want 1 (failEarly should stop)", processedLines)
	}
}

// --- failEarly with InvalidSystemError → exit 2 ---

func TestProcessBatch_failEarly_invalidSystem_returns2(t *testing.T) {
	stdoutW, _ := pipeCapture(t)
	stderrW, _ := pipeCapture(t)

	formatter := NewFormatter("line", false, false)
	scanner := makeScanner("anything")

	code := processBatch(scanner, sysError, formatter, stdoutW, stderrW, true)

	if code != 2 {
		t.Errorf("exit code = %d, want 2 for InvalidSystemError", code)
	}
}

// --- Per-line error includes 1-based line number (plain text) ---

func TestProcessBatch_errorIncludesLineNumber_plainText(t *testing.T) {
	stdoutW, _ := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)

	formatter := NewFormatter("line", false, false)
	// Put valid input on line 1, invalid on line 2, invalid on line 3
	callCount := 0
	compute := func(input string) (gematria.Result, error) {
		callCount++
		if callCount == 1 {
			return successCompute(input)
		}
		return alwaysError(input)
	}
	scanner := makeScanner("א", "bad", "bad")
	_ = processBatch(scanner, compute, formatter, stdoutW, stderrW, false)

	stderr := readStderr()

	if !strings.Contains(stderr, "line 2") {
		t.Errorf("stderr = %q, want it to contain 'line 2'", stderr)
	}
	if !strings.Contains(stderr, "line 3") {
		t.Errorf("stderr = %q, want it to contain 'line 3'", stderr)
	}
}

// --- Per-line error includes line number in JSON format ---

func TestProcessBatch_errorIncludesLineNumber_json(t *testing.T) {
	stdoutW, _ := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)

	formatter := NewFormatter("json", false, false)
	scanner := makeScanner("bad")

	_ = processBatch(scanner, alwaysError, formatter, stdoutW, stderrW, false)

	stderr := readStderr()

	if !strings.Contains(stderr, `"line":1`) {
		t.Errorf("stderr = %q, want JSON to contain \"line\":1", stderr)
	}
}

// --- Multiple successful lines all written to stdout ---

func TestProcessBatch_multipleSuccess_allWritten(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)

	formatter := NewFormatter("value", false, false)
	scanner := makeScanner("א", "ב", "ג") // aleph=1, bet=2, gimel=3

	code := processBatch(scanner, successCompute, formatter, stdoutW, stderrW, false)
	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	// Each result on its own line
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("stdout has %d lines, want 3: %q", len(lines), stdout)
	}
}
