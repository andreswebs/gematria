---
id: wor-jxhz
status: closed
deps: []
links: []
created: 2026-04-15T04:09:20Z
type: epic
priority: 1
tags: [core-compute, epic]
---
# Core Gematria Compute

Implement the primary feature of the gematria CLI: given Hebrew input (from positional args or stdin), compute its numeric gematria value using a chosen system, and render the result in the chosen output format.

This epic covers all code not provided by the letter-dictionary or reverse-lookup epics:

## Root Package Additions

- LetterResult and Result types (Compute return value, self-contained for all output formats)
- Parse(input string) ([]Letter, error) — resolves Hebrew Unicode, Latin transliteration, and mixed input into a letter sequence; returns typed errors with position and suggestions
- Compute(input string, system System) (Result, error) — calls Parse, applies system value tables, returns Result

## CLI Layer (internal/cli/)

- config.go — Config struct; flag parsing for --mispar/-m, --output/-o, --no-color, --atbash, --fail-early, --version, --help; env var resolution (GEMATRIA_MISPAR, GEMATRIA_OUTPUT) with flag > env > default precedence; strict enum validation with valid-values list on error
- output.go — Formatter interface: FormatResult(Result), FormatLookup([]Word, bool), FormatError(error)
- color.go — ANSI constants/helpers; UseColor(noColorFlag, noColorEnv bool, stdout *os.File) bool; IsTerminal wrapper; priority: --no-color > NO_COLOR > TTY detection
- line.go — LineFormatter: single-line "word = value (breakdown)" with RTL marks; supports --atbash display
- value.go — ValueFormatter: bare integer only
- card.go — CardFormatter: multi-line per-letter table with name, value, position, meaning, system; supports --atbash column
- json.go — JSONFormatter: structured JSON on stdout; structured JSON errors on stderr; no RTL marks, no ANSI
- batch.go — stdin batching: process all lines by default (exit 4 on partial success); --fail-early stops on first error; per-line errors include line number
- run.go (compute path) — orchestration: no-args+TTY usage hint, positional args, stdin delegation; maps typed errors to exit codes; --help and --version handling

## Documentation

- AGENTS.md — agent workflow guide: recommended flags, output format, error handling, exit codes, copy-paste examples

## Spec References

- docs/specs/requirements.md §2 (Input Handling), §3 (Systems), §4 (Output Formats), §5 (RTL Output), §8 (Error Handling), §9 (Env Vars)
- docs/specs/cli-design.md §1-§6, §8-§9
- docs/specs/code-architecture.md §3.4 (High-Level API), §3.5 (Result Type), §4 (CLI Layer), §7 (Testing)

## Task List

1. Define LetterResult and Result types
2. Implement Parse() and Compute() API
3. Implement CLI Config struct and flag parsing
4. Implement Formatter interface and color management
5. Implement four compute output formatters (line, value, card, json)
6. Implement stdin batching
7. Orchestrate compute path in Run()
8. Write AGENTS.md
9. Unit tests for Parse() and Compute()
10. CLI integration tests for compute path

