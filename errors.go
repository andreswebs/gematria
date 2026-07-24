package gematria

import (
	"fmt"
	"strings"
)

// InvalidCharError is returned by Letter() when the rune is not a recognized
// Hebrew character or sofit form. Char and Position let the CLI report exactly
// where the invalid input occurred; Input provides surrounding context.
type InvalidCharError struct {
	Char     rune
	Position int    // byte offset of Char in the original input string
	Input    string // the full input string, for context
}

func (e *InvalidCharError) Error() string {
	return fmt.Sprintf("invalid character %q at position %d", e.Char, e.Position)
}

// UnknownNameError is returned by LetterByName() when no alias matches.
// Suggestions are pre-computed by the caller using Levenshtein distance and
// stored here so the CLI can format them without re-computing.
type UnknownNameError struct {
	Name        string
	Position    int      // 0-based index of this name in a multi-word input
	Suggestions []string // candidate alias strings, may be empty
}

func (e *UnknownNameError) Error() string {
	if len(e.Suggestions) == 0 {
		return fmt.Sprintf("unknown letter name %q", e.Name)
	}
	return fmt.Sprintf("unknown letter name %q; did you mean: %s", e.Name, strings.Join(e.Suggestions, ", "))
}

// InvalidSystemError is returned when a System string does not match any known
// constant. The Valid field is populated by ValidSystems() and is used directly
// in error messages and JSON error output.
type InvalidSystemError struct {
	Name  string
	Valid []System
}

func (e *InvalidSystemError) Error() string {
	names := make([]string, len(e.Valid))
	for i, s := range e.Valid {
		names[i] = string(s)
	}
	return fmt.Sprintf("unknown system %q; valid values: %s", e.Name, strings.Join(names, ", "))
}

// UnknownWordError is returned by Transliterate when input cannot be resolved
// to Hebrew letters under the given scheme. Suggestions is empty in v1.
type UnknownWordError struct {
	Input       string
	Scheme      Scheme
	Position    int      // token index for multi-token input
	Suggestions []string // near-match candidates, may be empty
}

func (e *UnknownWordError) Error() string {
	base := fmt.Sprintf("input %q cannot be transliterated in scheme %q", e.Input, e.Scheme)
	if len(e.Suggestions) == 0 {
		return base
	}
	return base + "; did you mean: " + strings.Join(e.Suggestions, ", ")
}

// InvalidSchemeError parallels InvalidSystemError. Returned when a Scheme
// string does not match any known constant.
type InvalidSchemeError struct {
	Name  string
	Valid []Scheme
}

func (e *InvalidSchemeError) Error() string {
	names := make([]string, len(e.Valid))
	for i, s := range e.Valid {
		names[i] = string(s)
	}
	return fmt.Sprintf("unknown scheme %q; valid values: %s", e.Name, strings.Join(names, ", "))
}

// Compile-time interface checks.
var _ error = (*InvalidCharError)(nil)
var _ error = (*UnknownNameError)(nil)
var _ error = (*InvalidSystemError)(nil)
var _ error = (*UnknownWordError)(nil)
var _ error = (*InvalidSchemeError)(nil)
