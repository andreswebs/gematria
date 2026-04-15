---
id: wor-sgb3
status: closed
deps: [wor-cg08, wor-q01l]
links: []
created: 2026-04-15T04:10:17Z
type: task
priority: 1
parent: wor-jxhz
tags: [core-compute, formatter-color, task]
---
# Implement Formatter interface and color management

Create two foundational CLI files: output.go (Formatter interface) and color.go (ANSI color helpers and TTY detection). These are depended on by all four formatter implementations.

## output.go — Formatter Interface

Define the Formatter interface that all four formatters implement:

  type Formatter interface {
      FormatResult(r gematria.Result) string
      FormatLookup(words []gematria.Word, hasMore bool) string
      FormatError(err error) string
  }

FormatResult: renders a single Compute result. FormatLookup: renders reverse lookup results (implementation deferred to the reverse-lookup epic task wor-g7cz; stub returning "" is acceptable here). FormatError: renders an error for stderr. JSONFormatter.FormatError returns a JSON object; the other three return plain text.

Also export a NewFormatter(output string, useColor bool) Formatter factory function that returns the correct concrete formatter.

## color.go — ANSI Color and TTY Detection

Define ANSI escape constants (bold, dim, green, reset) and a small helper for wrapping text in color sequences.

Implement UseColor(noColorFlag bool, noColorEnv string, stdout *os.File) bool:
  1. If noColorFlag is true → return false
  2. If noColorEnv is non-empty (NO_COLOR set to any value) → return false
  3. If stdout is a terminal (IsTerminal) → return true
  4. Otherwise → return false

IsTerminal(f *os.File) bool wraps golang.org/x/term IsTerminal(int(f.Fd())). Returns false if f is nil.

Color helpers should never be called when UseColor returns false — the formatters check once at construction time and store the result.

## Spec References

- docs/specs/code-architecture.md §4.3 (Formatter interface), §4.4 (Color and TTY)
- docs/specs/cli-design.md §4.3 (visual hierarchy for human formats), §9.1 (NO_COLOR, --no-color, TTY detection)

## Acceptance Criteria

- [ ] Formatter interface is defined with FormatResult, FormatLookup, FormatError methods
- [ ] NewFormatter factory returns correct formatter for each of the four output format strings
- [ ] UseColor implements three-level priority correctly (noColorFlag > NO_COLOR env > TTY)
- [ ] IsTerminal returns false for pipe file descriptors (testable with os.Pipe())
- [ ] ANSI constants are defined and used only when useColor is true
- [ ] make build passes


## Notes

**2026-04-15T11:40:17Z**

Implemented Formatter interface (output.go) and color helpers (color.go). Added golang.org/x/term dependency for IsTerminal. ANSI constants (ansiReset, ansiBold, ansiGreen) and applyColor() live in output.go — not color.go — so the unused linter doesn't fire: they are consumed by the FormatError methods of lineFormatter and cardFormatter. FormatResult and FormatLookup are stubs (return "") to be filled in by wor-jkye. jsonFormatter.FormatError is also a stub for wor-jkye. valueFormatter and cardFormatter FormatError return a plain 'Error: ...' string with optional bold ANSI wrapping. errcheck: test cleanup must use '_ = r.Close()' not bare 'r.Close()' to satisfy golangci-lint.
