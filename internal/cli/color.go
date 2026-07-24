package cli

import (
	"os"

	"golang.org/x/term"
)

// IsTerminal reports whether f is connected to a terminal device.
// Returns false if f is nil.
func IsTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// UseColor reports whether ANSI color output should be used.
//
// Priority (highest to lowest):
//  1. noColorFlag — if true, always return false
//  2. noColorEnv  — if non-empty (NO_COLOR set to any value), return false
//  3. TTY detection — return true only when stdout is a terminal
func UseColor(noColorFlag bool, noColorEnv string, stdout *os.File) bool {
	if noColorFlag {
		return false
	}
	if noColorEnv != "" {
		return false
	}
	return IsTerminal(stdout)
}
