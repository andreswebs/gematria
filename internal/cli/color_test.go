package cli

import (
	"os"
	"testing"
)

func TestIsTerminal_nil(t *testing.T) {
	if IsTerminal(nil) {
		t.Error("IsTerminal(nil) = true, want false")
	}
}

func pipe(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close(); _ = w.Close() })
	return r, w
}

func TestIsTerminal_pipe(t *testing.T) {
	r, _ := pipe(t)
	if IsTerminal(r) {
		t.Error("IsTerminal(pipe) = true, want false")
	}
}

func TestUseColor(t *testing.T) {
	r, _ := pipe(t)

	cases := []struct {
		name        string
		noColorFlag bool
		noColorEnv  string
		stdout      *os.File
		want        bool
	}{
		{"flag disables color", true, "", r, false},
		{"env disables color", false, "1", r, false},
		{"env empty string is not set", false, "", r, false}, // pipe → false
		{"flag beats env", true, "1", r, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UseColor(tc.noColorFlag, tc.noColorEnv, tc.stdout)
			if got != tc.want {
				t.Errorf("UseColor(%v, %q, pipe) = %v, want %v",
					tc.noColorFlag, tc.noColorEnv, got, tc.want)
			}
		})
	}
}
