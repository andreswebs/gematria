package cli

import (
	"fmt"

	gematria "github.com/andreswebs/gematria"
)

// ANSI escape sequences for terminal styling.
// Used by formatters that have confirmed useColor is true.
const (
	ansiReset = "\033[0m"
	ansiBold  = "\033[1m"
	ansiGreen = "\033[32m"
)

// RTL/LTR Unicode marks for correct terminal rendering of Hebrew text.
const (
	rtlMark = "\u200F" // RIGHT-TO-LEFT MARK
	ltrMark = "\u200E" // LEFT-TO-RIGHT MARK
)

// applyColor wraps s with the given ANSI code and the reset sequence.
func applyColor(code, s string) string {
	return code + s + ansiReset
}

// wrapHebrew wraps a Hebrew string with RTL/LTR marks for correct
// terminal bidi rendering and copy-paste behavior.
func wrapHebrew(s string) string {
	return rtlMark + s + ltrMark
}

// Formatter renders gematria results, lookup results, and errors into
// strings for stdout and stderr. Each output format (line, value, card, json)
// has its own implementation; the factory NewFormatter selects the right one.
type Formatter interface {
	// FormatResult renders a single Compute result for stdout.
	FormatResult(r gematria.Result) string
	// FormatLookup renders reverse-lookup results for stdout.
	// hasMore indicates that additional matches exist beyond those returned.
	FormatLookup(words []gematria.Word, hasMore bool) string
	// FormatError renders an error for stderr.
	// JSONFormatter produces a JSON object; other formatters produce plain text.
	FormatError(err error) string
}

// NewFormatter returns the Formatter for the named output format.
// output must be one of "line", "value", "card", "json".
// Unrecognized values fall back to the line formatter.
// showAtbash controls whether human-facing formats append Atbash substitutions.
func NewFormatter(output string, useColor bool, showAtbash bool) Formatter {
	return NewFormatterWithLookup(output, useColor, showAtbash, 0, "")
}

// NewFormatterWithLookup is like NewFormatter but also sets the reverse-lookup
// context (target value and system) used by card and json formatters when
// rendering FormatLookup output.
func NewFormatterWithLookup(output string, useColor bool, showAtbash bool, findValue int, findSystem gematria.System) Formatter {
	switch output {
	case "value":
		return &valueFormatter{}
	case "card":
		return &cardFormatter{useColor: useColor, showAtbash: showAtbash, findValue: findValue, findSystem: findSystem}
	case "json":
		return &jsonFormatter{findValue: findValue, findSystem: findSystem}
	default: // "line" and any unrecognized value
		return &lineFormatter{useColor: useColor, showAtbash: showAtbash}
	}
}

// formatErrorPlain produces "Error: <message>" with optional ANSI bold.
func formatErrorPlain(err error, useColor bool) string {
	msg := fmt.Sprintf("Error: %s", err.Error())
	if useColor {
		return applyColor(ansiBold, msg)
	}
	return msg
}
