package cli

import (
	"strings"
	"testing"
)

// --- Default scheme (academic) computes correct value via -t ---

func TestRun_transliterate_defaultScheme_academicShalom(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"-t", "--output", "value", "shalom"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "370\n" {
		t.Errorf("stdout = %q, want %q (shalom academic = שלם = 370)", stdout, "370\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Israeli scheme produces a different value for the same input ---

func TestRun_transliterate_schemeIsraeli_shalom(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"-t", "--scheme", "israeli", "--output", "value", "shalom"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "376\n" {
		t.Errorf("stdout = %q, want %q (shalom israeli = שלום = 376)", stdout, "376\n")
	}
}

// --- Schemes produce different values for the same input ---

func TestRun_transliterate_schemesDiffer(t *testing.T) {
	cases := []struct {
		input  string
		scheme string
		want   string
	}{
		{"shalom", "academic", "370\n"},
		{"shalom", "israeli", "376\n"},
		{"emet", "academic", "440\n"},
		{"emet", "israeli", "441\n"},
		{"gadol", "academic", "37\n"},
		{"gadol", "israeli", "43\n"},
	}
	for _, tc := range cases {
		t.Run(tc.input+"_"+tc.scheme, func(t *testing.T) {
			stdoutW, readStdout := pipeCapture(t)
			stderrW, _ := pipeCapture(t)
			stdin := makeStdinPipe(t, "")

			code := Run([]string{"-t", "--scheme", tc.scheme, "--output", "value", tc.input},
				stdin, stdoutW, stderrW, noenv)

			stdout := readStdout()

			if code != 0 {
				t.Errorf("exit code = %d, want 0", code)
			}
			if stdout != tc.want {
				t.Errorf("stdout = %q, want %q", stdout, tc.want)
			}
		})
	}
}

// --- GEMATRIA_SCHEME env var honored ---

func TestRun_transliterate_envSchemeHonored(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_SCHEME": "israeli"})
	code := Run([]string{"-t", "--output", "value", "shalom"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "376\n" {
		t.Errorf("stdout = %q, want 376 (env-selected israeli)", stdout)
	}
}

// --- Flag overrides env var ---

func TestRun_transliterate_flagOverridesEnv(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_SCHEME": "israeli"})
	code := Run([]string{"-t", "--scheme", "academic", "--output", "value", "shalom"},
		stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "370\n" {
		t.Errorf("stdout = %q, want 370 (flag academic should win over env israeli)", stdout)
	}
}

// --- JSON output includes scheme field when transliterated ---

func TestRun_transliterate_jsonIncludesScheme(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"-t", "--scheme", "israeli", "--output", "json", "shalom"},
		stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, `"scheme":"israeli"`) {
		t.Errorf("stdout = %q, want JSON with scheme:israeli", stdout)
	}
	if !strings.Contains(stdout, `"total":376`) {
		t.Errorf("stdout = %q, want JSON with total:376", stdout)
	}
}

// --- JSON output OMITS scheme field for non-transliterated computation ---

func TestRun_nonTransliterate_jsonOmitsScheme(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--output", "json", "שלום"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if strings.Contains(stdout, `"scheme"`) {
		t.Errorf("stdout = %q, must NOT contain 'scheme' field for non-transliterated result", stdout)
	}
}

// --- Without -t, "shalom" still errors as unknown letter alias (regression) ---

func TestRun_noTransliterate_shalomStillUnknown(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"shalom"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 1 {
		t.Errorf("exit code = %d, want 1 (UnknownNameError without -t)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if !strings.Contains(stderr, "shalom") {
		t.Errorf("stderr = %q, want it to mention 'shalom'", stderr)
	}
}

// --- Multi-token positional input produces separate computations ---

func TestRun_transliterate_multiTokenSeparateComputations(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"-t", "--scheme", "israeli", "--output", "value", "shalom", "emet"},
		stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "376\n441\n" {
		t.Errorf("stdout = %q, want %q (two computations)", stdout, "376\n441\n")
	}
}

// --- Stdin batch transliterates each line ---

func TestRun_transliterate_stdinBatch(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "shalom\nemet\ngadol\n")

	code := Run([]string{"-t", "--scheme", "israeli", "--output", "value"},
		stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 for valid stdin batch", code)
	}
	if stdout != "376\n441\n43\n" {
		t.Errorf("stdout = %q, want %q", stdout, "376\n441\n43\n")
	}
}

// --- Invalid --scheme value → exit 2, valid-list error ---

func TestRun_transliterate_invalidScheme_exit2(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"-t", "--scheme", "bogus", "shalom"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "bogus") {
		t.Errorf("stderr = %q, want it to mention 'bogus'", stderr)
	}
	if !strings.Contains(stderr, "academic") || !strings.Contains(stderr, "israeli") {
		t.Errorf("stderr = %q, want valid-list with 'academic' and 'israeli'", stderr)
	}
}

// --- Lazy GEMATRIA_SCHEME validation: bogus env var without -t is OK ---

func TestRun_transliterate_lazyEnvValidationWithoutFlag(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_SCHEME": "bogus"})
	// Without -t, GEMATRIA_SCHEME is not validated.
	code := Run([]string{"--output", "value", "א"}, stdin, stdoutW, stderrW, getenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 0 {
		t.Errorf("exit code = %d, want 0 (env not validated without -t)", code)
	}
	if stdout != "1\n" {
		t.Errorf("stdout = %q, want %q", stdout, "1\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

// --- Lazy GEMATRIA_SCHEME validation: bogus env var WITH -t errors ---

func TestRun_transliterate_lazyEnvValidationWithFlag(t *testing.T) {
	stdoutW, _ := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	getenv := envWith(map[string]string{"GEMATRIA_SCHEME": "bogus"})
	code := Run([]string{"-t", "shalom"}, stdin, stdoutW, stderrW, getenv)

	stderr := readStderr()

	if code != 2 {
		t.Errorf("exit code = %d, want 2 (env validated with -t)", code)
	}
	if !strings.Contains(stderr, "GEMATRIA_SCHEME") {
		t.Errorf("stderr = %q, want it to mention GEMATRIA_SCHEME", stderr)
	}
}

// --- Unmappable input → exit 1, UnknownWordError on stderr ---

func TestRun_transliterate_unknownWord_exit1(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// 'h3llo' contains a digit, which is unmappable in any scheme.
	code := Run([]string{"-t", "--output", "value", "h3llo"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()
	stderr := readStderr()

	if code != 1 {
		t.Errorf("exit code = %d, want 1 (UnknownWordError)", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty on error", stdout)
	}
	if stderr == "" {
		t.Errorf("stderr empty, want UnknownWordError message")
	}
}

// --- JSON UnknownWordError on stderr includes scheme field ---

func TestRun_transliterate_unknownWordJSON_includesScheme(t *testing.T) {
	stdoutW, _ := pipeCapture(t)
	stderrW, readStderr := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"-t", "--scheme", "israeli", "--output", "json", "h3llo"},
		stdin, stdoutW, stderrW, noenv)

	stderr := readStderr()

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, `"scheme":"israeli"`) {
		t.Errorf("stderr = %q, want JSON error with scheme:israeli", stderr)
	}
	if !strings.Contains(stderr, `"error"`) {
		t.Errorf("stderr = %q, want JSON error with 'error' field", stderr)
	}
}

// --- Composes orthogonally with --mispar gadol (sofit extended) ---

func TestRun_transliterate_composesWithMispar(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// shalom (israeli) under gadol: ש=300 + ל=30 + ו=6 + ם=600 = 936
	code := Run([]string{"-t", "--scheme", "israeli", "--mispar", "gadol", "--output", "value", "shalom"},
		stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "936\n" {
		t.Errorf("stdout = %q, want %q (israeli shalom under gadol)", stdout, "936\n")
	}
}

// --- Composes orthogonally with --mispar atbash ---

func TestRun_transliterate_composesWithMisparAtbash(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// shalom (israeli) שלום under atbash: ש=2, ל=20, ו=80, ם=10 → 112
	code := Run([]string{"-t", "--scheme", "israeli", "--mispar", "atbash", "--output", "value", "shalom"},
		stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "112\n" {
		t.Errorf("stdout = %q, want %q", stdout, "112\n")
	}
}

// --- Card output displays Scheme line when transliterated ---

func TestRun_transliterate_cardShowsScheme(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"-t", "--scheme", "israeli", "--output", "card", "shalom"},
		stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "Scheme: israeli") {
		t.Errorf("stdout = %q, want it to contain 'Scheme: israeli'", stdout)
	}
}

// --- Card output OMITS Scheme line for non-transliterated computation ---

func TestRun_nonTransliterate_cardOmitsScheme(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--output", "card", "שלום"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if strings.Contains(stdout, "Scheme:") {
		t.Errorf("stdout = %q, must NOT contain 'Scheme:' for non-transliterated", stdout)
	}
}

// --- Hebrew Unicode passes through with -t (transliteration is Latin-only) ---

func TestRun_transliterate_hebrewPassthrough(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	// Hebrew Unicode input is unchanged whether -t is set or not.
	code := Run([]string{"-t", "--output", "value", "שלום"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stdout != "376\n" {
		t.Errorf("stdout = %q, want 376 (Hebrew passes through)", stdout)
	}
}

// --- Mode exclusivity: -t aleph is transliterated, NOT the letter alias ---

func TestRun_transliterate_modeExclusivity(t *testing.T) {
	// gematria aleph (no -t) → 1 (letter alias for א)
	// gematria -t aleph (academic) → letter-by-letter transliteration; NOT 1
	stdoutNoFlag, readNoFlag := pipeCapture(t)
	stderrNoFlag, _ := pipeCapture(t)
	stdinNoFlag := makeStdinPipe(t, "")
	codeNoFlag := Run([]string{"--output", "value", "aleph"}, stdinNoFlag, stdoutNoFlag, stderrNoFlag, noenv)
	gotNoFlag := readNoFlag()

	if codeNoFlag != 0 {
		t.Errorf("no-flag exit = %d, want 0", codeNoFlag)
	}
	if gotNoFlag != "1\n" {
		t.Errorf("no-flag stdout = %q, want 1 (letter alias for Aleph)", gotNoFlag)
	}

	stdoutFlag, readFlag := pipeCapture(t)
	stderrFlag, _ := pipeCapture(t)
	stdinFlag := makeStdinPipe(t, "")
	codeFlag := Run([]string{"-t", "--scheme", "academic", "--output", "value", "aleph"},
		stdinFlag, stdoutFlag, stderrFlag, noenv)
	gotFlag := readFlag()

	if codeFlag != 0 {
		t.Errorf("with-flag exit = %d, want 0", codeFlag)
	}
	if gotFlag == "1\n" {
		t.Errorf("with-flag stdout = %q, must NOT equal 1 (transliteration is not letter-alias)", gotFlag)
	}
}

// --- Default scheme when -t set with no --scheme is academic ---

func TestRun_transliterate_defaultIsAcademic(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"-t", "--output", "json", "shalom"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, `"scheme":"academic"`) {
		t.Errorf("stdout = %q, want JSON to contain scheme:academic (default)", stdout)
	}
}

// --- Help text documents -t, --scheme, GEMATRIA_SCHEME, and includes examples ---

func TestRun_help_documentsTransliterationFeature(t *testing.T) {
	stdoutW, readStdout := pipeCapture(t)
	stderrW, _ := pipeCapture(t)
	stdin := makeStdinPipe(t, "")

	code := Run([]string{"--help"}, stdin, stdoutW, stderrW, noenv)

	stdout := readStdout()

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	for _, want := range []string{
		"--transliterate",
		"--scheme",
		"GEMATRIA_SCHEME",
		"academic",
		"israeli",
		"-t shalom", // example
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("--help output missing %q", want)
		}
	}
}
