package cli

import "testing"

// TestExitCodes_conformance drives cli.Run through at least one scenario per
// declared exit code and asserts that (a) the observed code matches the
// expected code and (b) every observed code is declared in ExitCodes, per
// docs/adr/0001-exit-code-taxonomy.md.
func TestExitCodes_conformance(t *testing.T) {
	// A non-existent wordlist path for the I/O-error scenario. It must not end
	// in .db/.idx so it resolves to the in-memory backend, whose os.Open fails.
	missingWordlist := "/nonexistent/gematria-exitcodes-test/words.txt"

	tests := []struct {
		name   string
		args   []string
		stdin  string
		getenv func(string) string
		want   int
	}{
		{
			name: "success/positional arg",
			args: []string{"א"},
			want: exitOK,
		},
		{
			name:  "partial batch/one ok one bad",
			stdin: "א\nbad_latin\n",
			want:  exitPartialBatch,
		},
		{
			name: "usage/invalid --mispar flag",
			args: []string{"--mispar", "bogus", "א"},
			want: exitUsage,
		},
		{
			name: "usage/--find without wordlist",
			args: []string{"--find", "376"},
			// Point index discovery at an empty dir so nothing is found.
			getenv: envWith(map[string]string{"GEMATRIA_INDEX_LOCATION": t.TempDir()}),
			want:   exitUsage,
		},
		{
			name: "data/invalid input",
			args: []string{"xyz"},
			want: exitDataErr,
		},
		{
			name:  "data/all batch lines failed",
			stdin: "bad_latin\nalso_bad\n",
			want:  exitDataErr,
		},
		{
			name: "io/wordlist not found",
			args: []string{"--find", "376", "--wordlist", missingWordlist},
			want: exitIOErr,
		},
		{
			name:   "config/invalid GEMATRIA_MISPAR at compute time",
			args:   []string{"א"},
			getenv: envWith(map[string]string{"GEMATRIA_MISPAR": "nonexistent"}),
			want:   exitConfigErr,
		},
		{
			name:   "config/invalid GEMATRIA_LIMIT",
			args:   []string{"--find", "376", "--wordlist", missingWordlist},
			getenv: envWith(map[string]string{"GEMATRIA_LIMIT": "notanumber"}),
			want:   exitConfigErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdoutW, _ := pipeCapture(t)
			stderrW, _ := pipeCapture(t)
			stdin := makeStdinPipe(t, tt.stdin)

			getenv := tt.getenv
			if getenv == nil {
				getenv = noenv
			}

			code := Run(tt.args, stdin, stdoutW, stderrW, getenv)

			if code != tt.want {
				t.Errorf("exit code = %d, want %d", code, tt.want)
			}
			if _, ok := ExitCodes[code]; !ok {
				t.Errorf("exit code %d is not declared in ExitCodes", code)
			}
		})
	}
}

// TestExitCodes_registryComplete guards the registry against drift: it must
// declare exactly the codes the tool documents, and no failure code outside
// the fleet taxonomy's mandatory range.
func TestExitCodes_registryComplete(t *testing.T) {
	want := []int{exitOK, exitPartialBatch, exitUsage, exitDataErr, exitIOErr, exitConfigErr}
	if len(ExitCodes) != len(want) {
		t.Errorf("ExitCodes has %d entries, want %d", len(ExitCodes), len(want))
	}
	for _, c := range want {
		if _, ok := ExitCodes[c]; !ok {
			t.Errorf("ExitCodes missing declared code %d", c)
		}
	}
}
