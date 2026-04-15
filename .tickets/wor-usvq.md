---
id: wor-usvq
status: closed
deps: [wor-gdda]
links: []
created: 2026-04-15T04:01:54Z
type: task
priority: 1
parent: wor-g7f2
tags: [letter-dictionary, typed-errors, task]
---
# Define typed error types

Add src/errors.go with three exported error structs: InvalidCharError, UnknownNameError, and InvalidSystemError. Each carries structured fields so the CLI can extract precise data for both human and JSON error formatting without parsing error strings.

## Spec References
- docs/specs/requirements.md §8 (error messages identify invalid character and position)
- docs/specs/code-architecture.md §3.7 (typed errors with structured fields, CLI uses errors.As)
- docs/specs/cli-design.md §4.5 (structured error fields for JSON output: error, position, invalid_input, suggestions)

## Implementation Details

### InvalidCharError

Returned by Letter() when the rune is not a recognized Hebrew character or sofit form. Carries the precise rune and its byte offset so the CLI can report position to the user (or agent via JSON).

    type InvalidCharError struct {
        Char     rune
        Position int    // byte offset of Char in the original input string
        Input    string // the full input string, for context
    }

    func (e *InvalidCharError) Error() string {
        return fmt.Sprintf("invalid character %q at position %d", e.Char, e.Position)
    }

### UnknownNameError

Returned by LetterByName() when no alias matches. Suggestions are pre-computed by the caller (gematria.go) using Levenshtein distance and stored here — the CLI layer formats them without re-computing.

    type UnknownNameError struct {
        Name        string
        Position    int      // index of this name in a multi-word input, 0-based
        Suggestions []string // candidate alias strings, may be empty
    }

    func (e *UnknownNameError) Error() string {
        if len(e.Suggestions) == 0 {
            return fmt.Sprintf("unknown letter name %q", e.Name)
        }
        return fmt.Sprintf("unknown letter name %q; did you mean: %s", e.Name, strings.Join(e.Suggestions, ", "))
    }

### InvalidSystemError

Returned when a System string does not match any known constant. The Valid field is populated by ValidSystems() from systems.go and is used directly in the error message and in JSON error output.

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

### Compile-Time Interface Checks

Add blank-identifier checks at the bottom of the file to guarantee all three implement error:

    var _ error = (*InvalidCharError)(nil)
    var _ error = (*UnknownNameError)(nil)
    var _ error = (*InvalidSystemError)(nil)

## Acceptance Criteria

- [ ] `InvalidCharError` has Char (rune), Position (int), Input (string) fields — all exported
- [ ] `UnknownNameError` has Name (string), Position (int), Suggestions ([]string) fields — all exported
- [ ] `InvalidSystemError` has Name (string), Valid ([]System) fields — all exported
- [ ] All three implement `error` via `Error() string` method on pointer receiver
- [ ] Compile-time interface checks `var _ error = (*T)(nil)` present for all three types
- [ ] `Error()` messages are human-readable and include the invalid value
- [ ] `UnknownNameError.Error()` includes suggestions when Suggestions is non-empty
- [ ] File imports only `fmt` and `strings` from the standard library; no `os` import
- [ ] `go build ./...` from src/ passes


## Notes

**2026-04-15T11:14:24Z**

Implemented src/errors.go with InvalidCharError, UnknownNameError, and InvalidSystemError. All three have exported fields, pointer-receiver Error() methods, and compile-time interface checks (var _ error = (*T)(nil)). Tests live in src/errors_test.go covering Error() message content, field accessibility, and errors.As compatibility. No imports beyond fmt and strings. make build passes clean.
