# Gematria CLI - Code Architecture

> **Status**: Ready

This document describes the package design and code architecture for the
gematria CLI. It covers module structure, package responsibilities, public API
surface, and internal organization. It does not prescribe implementation
details — only the structural decisions and their rationale.

The CLI design it builds on lives in [cli-design.md](cli-design.md).
The requirements live in [requirements.md](requirements.md).

---

## 1. Organizing Principle

The application is organized as a **Go library with a CLI wrapper**:

- The **root package** (`github.com/andreswebs/gematria`) contains all domain
  logic — letter data, gematria computation, reverse lookup, word list
  parsing. It is importable by other Go projects.
- The **CLI** is a thin wrapper in `cmd/gematria/` that delegates to
  `internal/cli/` for flag parsing, output formatting, and orchestration.

This separation ensures the domain logic is reusable and testable
independently of the command-line interface.

---

## 2. Package Structure

Three packages total:

```
src/
  go.mod                       # module github.com/andreswebs/gematria
  gematria.go                  # Root package public API
  letters.go                   # Letter dictionary (map literals)
  systems.go                   # Gematria system lookup tables
  errors.go                    # Typed error definitions
  wordlist.go                  # WordSource interface, ParseWordList, Word type
  cmd/
    gematria/
      main.go                  # Entrypoint: cli.Run() + os.Exit()
  internal/
    cli/
      run.go                   # Run() entry point, orchestration
      config.go                # Config struct, flag parsing, env var reading
      output.go                # Formatter interface definition
      line.go                  # LineFormatter
      value.go                 # ValueFormatter
      card.go                  # CardFormatter
      json.go                  # JSONFormatter
      color.go                 # ANSI color, NO_COLOR, TTY detection
      batch.go                 # Stdin batching, --fail-early, partial success
```

File names are indicative — files may be combined or split differently during
implementation. The **package boundaries** are what matter.

### Why This Structure

- **No `pkg/` directory.** The root package is the public library. Adding
  `pkg/gematria/` would create import path stutter
  (`gematria/pkg/gematria`). The Go community has moved away from the `pkg/`
  convention.
- **No `internal/config/`.** Configuration is a CLI concern. Flag parsing,
  env var reading, and precedence resolution live in `internal/cli/`
  alongside the other CLI plumbing.
- **No CLI framework.** The flag surface is small (~8 flags, no subcommands).
  `pflag` provides GNU-style long/short flags. A framework like urfave/cli
  would fight our env var injection and TTY detection requirements without
  adding value.
- **Single flat root package.** The domain is small (27 letters, 4 systems,
  one lookup function, one parser). Sub-packages would create
  cross-dependencies for no benefit. Files provide sufficient organization.
- **Single flat `internal/cli/` package.** All CLI concerns are internal to
  this binary. No external consumer needs a clean import path into the CLI
  layer. Files organized by responsibility.

---

## 3. Root Package — Public API

The root package exposes a **two-level API**: low-level functions for single
letter operations, and high-level functions for word-level computation.

### 3.1 Letter Dictionary

Letter data for all 22 standard Hebrew letters and 5 sofit forms is stored as
**Go map literals** at the package level. Two maps:

- **Primary map** keyed by Hebrew rune — provides letter data by character.
- **Alias map** keyed by transliteration string — resolves English names to
  Hebrew runes. Built at package init time.

Map literals are compile-time type-checked, trivially small (27 entries), and
represent data that is truly static.

### 3.2 Gematria Systems

The four systems (Hechrachi, Gadol, Siduri, Atbash) are represented as
**lookup tables** — a `map[rune]int` per system. Systems are pure data (a
letter-to-number mapping), not behavior. No function types, no interfaces, no
polymorphism — just different numbers for the same letters.

### 3.3 Low-Level API

- `Letter(r rune) (Letter, error)` — Look up a letter by Hebrew character.
- `LetterByName(name string) (Letter, error)` — Look up a letter by
  transliteration alias (case-insensitive).
- `AtbashSubstitute(r rune) rune` — Return the Atbash-substituted letter.
  Separate from `Compute` because Atbash display (`--atbash` flag) is a
  presentation concern, not a computation concern.

### 3.4 High-Level API

- `Parse(input string) ([]Letter, error)` — Resolve any input (Hebrew
  characters, Latin transliterated names, or a mix) into a sequence of
  letters. Handles multi-word Latin input (e.g., `"aleph bet"` → `[א, ב]`).
  Returns typed errors with position and near-match suggestions on failure.
  Exported so consumers can resolve and inspect input without computing.
- `Compute(input string, system System) (Result, error)` — Parse input,
  compute gematria values, return a self-contained result. Calls `Parse`
  internally.
- `FindByValue(value int, source WordSource, system System, limit int) ([]Word, bool, error)`
  — Reverse lookup: find words from a `WordSource` whose gematria value
  equals the target. Returns matches, a `hasMore` indicator, and any error.

The word transliteration feature adds three more functions
(`Transliterate`, `ComputeFromLetters`, `ComputeTransliterated`), a `Scheme`
type, an `UnknownWordError`, and a `Scheme` field on `Result`. See
[transliteration.md](transliteration.md) for the complete API additions and
rationale.

### 3.5 Result Type

`Compute` returns a `Result` struct using **nested composition**:

- `Result` contains the total value, the input string, the system used, and a
  `[]LetterResult` breakdown.
- `LetterResult` embeds a full `Letter` (dictionary entry with name, meaning,
  position, all system values) plus the computed value in the selected system.

The result is **self-contained**: the CLI can render any output format (line,
value, card, json) from a single `Result` without additional lookups.

### 3.6 WordSource Interface

Reverse lookup accepts a `WordSource` interface rather than a concrete
`[]Word` slice:

```
type WordSource interface {
    FindByValue(value int, system System, limit int) ([]Word, bool, error)
}
```

`ParseWordList(r io.Reader)` returns an in-memory implementation backed by
the parsed word list. This is the only implementation for now.

The interface exists to allow future backends (embedded database, pre-computed
index, remote source) without changing the public API signature. See
[wordlist-backends.md](wordlist-backends.md) for details.

`ParseWordList` accepts an `io.Reader`, not a file path. The root package
never imports `os` — it is pure computation over data.

### 3.7 Typed Errors

The root package defines custom error types for each failure mode:

- `InvalidCharError` — Character outside the expected Unicode range. Carries
  the rune, its byte position in the input, and the input string.
- `UnknownNameError` — Latin input that doesn't match any transliteration
  alias. Carries the name, its position, and a list of near-match
  suggestions (Levenshtein-based, edit distance ≤ half input length, capped
  at 2).
- `InvalidSystemError` — Unrecognized gematria system name. Carries the
  invalid name and the list of valid system names.

Typed errors carry structured data so any consumer (CLI, library user, test)
can extract precise fields without parsing error messages. The CLI uses
`errors.As` to type-switch and map these to exit codes and output formats.

Exit code mapping is a CLI concern — the domain does not know about exit
codes.

---

## 4. CLI Layer — `internal/cli/`

### 4.1 Entry Point

`Run(args []string, stdin *os.File, stdout *os.File, stderr *os.File, getenv func(string) string) int`

All OS primitives are injected:

- `args` — command-line arguments (without the program name).
- `stdin`, `stdout`, `stderr` — as `*os.File` (not `io.Writer`) so TTY
  detection can call `IsTerminal()` on the file descriptor.
- `getenv` — environment variable lookup function. In production this is
  `os.Getenv`. In tests, it's a closure over a map, enabling parallel tests
  without touching the process environment.

Returns an `int` exit code. `main.go` calls `os.Exit()` with it.

### 4.2 Config Resolution

A `Config` struct holds the resolved configuration: mispar system, output
format, wordlist path, limit, atbash flag, no-color flag, fail-early flag.

Resolution follows a strict precedence order:

1. Explicit flag (`--mispar gadol`)
2. Environment variable (`GEMATRIA_MISPAR=gadol`)
3. Built-in default (`hechrachi`)

Environment variables are validated **lazily** — only when the feature they
control is actually used. A stale `GEMATRIA_WORDLIST` does not block unrelated
lookups.

Flag values for enum-type flags (`--mispar`, `--output`) require **exact
match**. No prefix matching, no fuzzy matching.

### 4.3 Formatter Interface

```
type Formatter interface {
    FormatLetter(result) string
    FormatWord(result) string
    FormatLookup(results, hasMore) string
    FormatError(err error) string
}
```

Four implementations: `LineFormatter`, `ValueFormatter`, `CardFormatter`,
`JSONFormatter`. The formatter is resolved once from the config and passed
through the orchestration. Method signatures are indicative — exact types
determined during implementation.

`FormatError` is on the interface because error format is tied to output
format: `JSONFormatter` produces structured JSON error objects on stderr,
the other three produce plain text. This was a deliberate design decision —
`--output` controls both stdout and stderr format.

### 4.4 Color and TTY

Color is controlled by three mechanisms in priority order:

1. `--no-color` flag (always disables)
2. `NO_COLOR` env var (disables if set)
3. TTY detection (enabled when stdout is a TTY, disabled when piped)

TTY detection requires `*os.File` (not `io.Writer`) on stdout and stderr.
This is why `Run()` accepts file descriptors.

### 4.5 Stdin Batching

When reading multiple lines from stdin:

- **Default behavior**: process all lines. Valid lines produce results on
  stdout, invalid lines produce per-line errors on stderr. Exit code 4 on
  partial success.
- **`--fail-early`**: stop on first error, exit immediately with the
  appropriate error code (1, 2, or 3).

### 4.6 No-Args Behavior

When stdin is a TTY and no arguments are given, print a short usage hint to
stderr and exit 0. When stdin is piped, read from it as normal.

---

## 5. `cmd/gematria/main.go`

Ultra-thin. Three to five lines:

```go
func main() {
    code := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getenv)
    os.Exit(code)
}
```

`main()` is the one function you can't unit test in Go. Everything testable
lives in `internal/cli/` and the root package.

---

## 6. Dependency Flow

```
cmd/gematria/main.go
    → internal/cli/        (flag parsing, formatting, orchestration)
        → gematria          (domain: letters, computation, lookup)
```

The root package has **no dependencies** on the CLI layer. The CLI layer
depends on the root package. `main.go` depends on the CLI layer only.

No circular dependencies. No upward imports.

---

## 7. Testing Strategy

### Root Package

- Pure unit tests. No I/O, no filesystem, no environment.
- Letter dictionary: verify all 27 entries, alias resolution, error cases.
- Computation: table-driven tests across all 4 systems.
- Reverse lookup: test with `ParseWordList` over `strings.NewReader()`.
- Typed errors: verify fields (position, suggestions) are populated
  correctly.

### `internal/cli/`

- Integration-style tests using `Run()` with injected args, fake stdin/stdout
  (pipes), and a `getenv` closure.
- Formatter tests: give each formatter a domain result, assert the output
  string.
- Batch tests: multi-line stdin with valid and invalid lines, verify
  stdout/stderr separation and exit codes.
- TTY behavior can be tested by passing pipe file descriptors (not a TTY) and
  verifying color is absent.

### `cmd/gematria/`

- Not directly unit tested (it's 3 lines).
- Covered by end-to-end shell tests (shelltestrunner or similar) that invoke
  the built binary.
