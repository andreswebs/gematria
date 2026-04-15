package cli

import (
	"strings"
	"testing"
)

// --- Output format: --output value produces a bare integer ---

func TestRun_outputValue_bareInteger(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--output", "value", "א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "1\n" {
		t.Errorf("stdout = %q, want %q (bare integer for aleph)", stdout, "1\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Output format: --output line produces formatted output with letter name ---

func TestRun_outputLine_containsLetterName(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--output", "line", "א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "Aleph") {
		t.Errorf("stdout = %q, want it to contain 'Aleph'", stdout)
	}
	if !strings.Contains(stdout, "1") {
		t.Errorf("stdout = %q, want it to contain value '1'", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Output format: --output card produces multi-line table with field headers ---

func TestRun_outputCard_containsFieldHeaders(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--output", "card", "א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	for _, want := range []string{"Name", "Value", "Meaning", "Aleph", "1", "ox"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Mispar system: --mispar hechrachi (explicit) uses standard values ---

func TestRun_misparHechrachi_explicit(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// כ (kaf) = 20 in hechrachi; 11 in siduri — use explicit --mispar to verify
	code := Run([]string{"--mispar", "hechrachi", "--output", "value", "כ"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "20\n" {
		t.Errorf("stdout = %q, want %q (kaf hechrachi=20)", stdout, "20\n")
	}
}

// --- Mispar system: --mispar gadol assigns extended values to sofit letters ---

func TestRun_misparGadol_sofitExtendedValue(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// ך (kaf sofit) = 500 in gadol (vs 20 in hechrachi)
	code := Run([]string{"--mispar", "gadol", "--output", "value", "ך"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "500\n" {
		t.Errorf("stdout = %q, want %q (kaf sofit gadol=500)", stdout, "500\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Mispar system: --mispar siduri uses ordinal (positional) values ---

func TestRun_misparSiduri_ordinalValue(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// כ (kaf) = 11 in siduri (ordinal position); 20 in hechrachi
	code := Run([]string{"--mispar", "siduri", "--output", "value", "כ"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "11\n" {
		t.Errorf("stdout = %q, want %q (kaf siduri=11)", stdout, "11\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Mispar system: --mispar atbash uses cipher values (mirror letter's hechrachi value) ---

func TestRun_misparAtbash_cipherValue(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// א (aleph) atbash = 400 (value of ת tav in hechrachi)
	code := Run([]string{"--mispar", "atbash", "--output", "value", "א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "400\n" {
		t.Errorf("stdout = %q, want %q (aleph atbash=400)", stdout, "400\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- GEMATRIA_MISPAR env var: valid value is used when --mispar flag absent ---

func TestRun_gematriaSystemEnv_gadol(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_MISPAR": "gadol"})
	// ך (kaf sofit) = 500 in gadol, 20 in hechrachi
	code := Run([]string{"--output", "value", "ך"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "500\n" {
		t.Errorf("stdout = %q, want %q (GEMATRIA_MISPAR=gadol, kaf sofit=500)", stdout, "500\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- GEMATRIA_MISPAR env var: flag overrides env var ---

func TestRun_gematriaSystemFlag_overridesEnv(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// env says gadol (500 for ך), flag says hechrachi (20 for ך)
	getenv := envWith(map[string]string{"GEMATRIA_MISPAR": "gadol"})
	code := Run([]string{"--mispar", "hechrachi", "--output", "value", "ך"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "20\n" {
		t.Errorf("stdout = %q, want %q (flag hechrachi beats env gadol)", stdout, "20\n")
	}
}

// --- GEMATRIA_OUTPUT env var: used when --output flag absent ---

func TestRun_gematriaOutputEnv_value(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_OUTPUT": "value"})
	code := Run([]string{"א"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "1\n" {
		t.Errorf("stdout = %q, want %q (GEMATRIA_OUTPUT=value)", stdout, "1\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- GEMATRIA_OUTPUT env var: flag overrides env var ---

func TestRun_gematriaOutputFlag_overridesEnv(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// env says card; flag says value
	getenv := envWith(map[string]string{"GEMATRIA_OUTPUT": "card"})
	code := Run([]string{"--output", "value", "א"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "1\n" {
		t.Errorf("stdout = %q, want %q (flag value beats env card)", stdout, "1\n")
	}
}

// --- Invalid --mispar flag: exit 2, stderr lists valid values, stdout empty ---

func TestRun_invalidMispar_exit2(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--mispar", "standard", "א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 for invalid --mispar", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on config error", stdout)
	}
	if !strings.Contains(stderr, "standard") {
		t.Errorf("stderr = %q, want mention of invalid value 'standard'", stderr)
	}
	for _, valid := range []string{"hechrachi", "gadol", "siduri", "atbash"} {
		if !strings.Contains(stderr, valid) {
			t.Errorf("stderr = %q, want it to list valid value %q", stderr, valid)
		}
	}
}

// --- Latin transliteration: single name resolves to the correct letter ---

func TestRun_latinInput_singleName(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// "aleph" is a Latin alias for א; should compute the same value
	code := Run([]string{"--output", "value", "aleph"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "1\n" {
		t.Errorf("stdout = %q, want %q (aleph=1)", stdout, "1\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Latin transliteration: space-separated names resolve to the correct total ---

func TestRun_latinInput_multiWord(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// "aleph mem tav" = א(1) + מ(40) + ת(400) = 441
	code := Run([]string{"--output", "value", "aleph mem tav"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "441\n" {
		t.Errorf("stdout = %q, want %q (aleph+mem+tav=441)", stdout, "441\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- --no-color: output contains no ANSI escape sequences ---

func TestRun_noColor_noANSICodes(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--no-color", "--output", "line", "א"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if strings.Contains(stdout, "\033[") {
		t.Errorf("stdout = %q, want no ANSI escape codes with --no-color", stdout)
	}
	if !strings.Contains(stdout, "1") {
		t.Errorf("stdout = %q, want output to still contain value '1'", stdout)
	}
}

// --- Stdin batch: all lines invalid → exit 1 ---

func TestRun_stdinBatch_allInvalid_exit1(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "bad\nworse\n")

	code := Run([]string{}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 1 {
		t.Errorf("exit code = %d, want 1 (all lines failed)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty when all lines fail", stdout)
	}
	if stderr == "" {
		t.Errorf("stderr empty, want error messages")
	}
}

// --- Stdin batch: --fail-early stops on first error, remaining lines unprocessed ---

func TestRun_stdinBatch_failEarly_stopsOnFirst(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	// First line invalid, second line valid — fail-early must not process the second
	stdin := makeStdinPipe(t, "bad\nא\n")

	code := Run([]string{"--fail-early"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 1 {
		t.Errorf("exit code = %d, want 1 (--fail-early on first error)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty (no successful lines processed before first error)", stdout)
	}
	if stderr == "" {
		t.Errorf("stderr empty, want error for first bad line")
	}
	// Second line "א" must not appear in stdout (processing stopped)
	if strings.Contains(stdout, "1") {
		t.Errorf("stdout = %q, want no output from lines after first error", stdout)
	}
}
