---
id: wor-ekci
status: closed
deps: [wor-cg08, wor-gdda, wor-j5l7, wor-usvq, wor-ydfb]
links: []
created: 2026-04-15T04:09:47Z
type: task
priority: 1
parent: wor-jxhz
tags: [core-compute, compute-api, task]
---
# Implement Parse() and Compute() API

Implement the two high-level exported functions in the root package that form the primary compute API.

## Parse(input string) ([]Letter, error)

Resolves any input string into a []Letter sequence. Handles three input modes:
- Pure Hebrew Unicode (U+05D0–U+05EA, including sofit forms): iterate rune-by-rune, call Letter(r) for each, return InvalidCharError if any rune is outside the Hebrew range
- Pure Latin transliteration: space-split the string into tokens, call LetterByName(token) for each, return UnknownNameError (with Levenshtein suggestions) on no match
- Mixed: detect per-token whether it is Hebrew or Latin; Hebrew chars are single-rune tokens, Latin tokens are space-separated words

Position in errors: for Hebrew input, position is the byte offset of the invalid rune; for Latin input, position is the token index (0-based).

## Compute(input string, system System) (Result, error)

Calls Parse(input) to get []Letter. For each letter, looks up the value in the system's lookup table (from systems.go). Builds []LetterResult and sums Total. Returns a Result.

Special case for Atbash system: call AtbashSubstitute(r) on each letter's rune before looking up its Hechrachi value. The LetterResult.Char should reflect the original rune; the substituted rune is used only for value lookup.

ValidateSystem before computing — return InvalidSystemError if the system is unrecognized.

## Spec References

- docs/specs/code-architecture.md §3.4 (Parse, Compute signatures)
- docs/specs/requirements.md §2 (Input Handling), §3 (Gematria Systems, Atbash)
- docs/specs/cli-design.md §3.3 (Levenshtein threshold: edit distance ≤ half input length, max 2)

## Acceptance Criteria

- [ ] Parse() handles Hebrew Unicode input char-by-char; returns InvalidCharError with rune and position for non-Hebrew chars
- [ ] Parse() handles Latin transliteration names (space-separated); returns UnknownNameError with suggestions on mismatch
- [ ] Parse() handles mixed Hebrew/Latin input
- [ ] Compute() returns a Result with correct Total for all four systems using known test values (e.g., aleph=1 in Hechrachi, aleph=1 in Gadol, aleph=1 in Siduri, aleph=400 in Atbash)
- [ ] Compute() for Atbash system substitutes letters before value lookup
- [ ] Compute() returns InvalidSystemError for unrecognized system names
- [ ] Root package has zero os imports
- [ ] make build passes


## Notes

**2026-04-15T11:28:07Z**

Implemented Parse() and Compute() in src/gematria.go. Key decisions: (1) Parse uses a pre-scan to detect mode (pureHebrew/pureLatin/mixed), then dispatches to parsePureHebrew/parsePureLatin/parseMixed helpers. isHebrewRune checks U+05D0-U+05EA which exactly covers all 27 Hebrew letters. (2) Compute does NOT call AtbashSubstitute — the systemValues[Atbash] table already maps each original rune to the hechrachi value of its mirror, so no substitution needed at compute time; AtbashSubstitute is for formatter display only. (3) Multi-word aliases like 'kaf sofit' are NOT supported via space-split Latin mode (would require greedy/backtracking tokeniser); users should use the Hebrew character directly. Tests added to gematria_test.go covering all four systems, Latin/Hebrew/mixed input, InvalidSystemError, and error positioning.
