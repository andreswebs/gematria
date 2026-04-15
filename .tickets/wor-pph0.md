---
id: wor-pph0
status: closed
deps: [wor-ekci]
links: []
created: 2026-04-15T04:11:34Z
type: task
priority: 2
parent: wor-jxhz
tags: [core-compute, compute-unit-tests, task]
---
# Unit tests for Parse() and Compute()

Write table-driven unit tests in the root package (src/) covering Parse() and Compute(). No I/O, no filesystem, no environment — pure computation tests.

## Parse() Tests (src/parse_test.go or src/gematria_test.go)

- Hebrew Unicode input: single letter (all 27 entries), multi-char word; verify correct letter sequence
- Latin transliteration: single name, multiple space-separated names (e.g., "aleph mem tav" → [א, מ, ת]); case-insensitive variants
- Mixed input: Hebrew char + Latin name in same string
- InvalidCharError: verify rune field, position field, and that errors.As extracts the type correctly
- UnknownNameError: verify name field, position field, suggestions field (e.g., "shen" → suggestion "shin")
- Levenshtein threshold: input with edit distance > 2 or > half-length gets empty suggestions

## Compute() Tests (same file or separate)

- Hechrachi system: aleph=1, bet=2, ..., tav=400; verify word total (e.g., "אמת" = 441)
- Gadol system: final kaf=500, final mem=600, final nun=700, final pe=800, final tsadi=900
- Siduri system: aleph=1, bet=2, ..., tav=22 (ordinal by position)
- Atbash system: aleph substituted for tav gets value 400; tav substituted for aleph gets value 1
- InvalidSystemError for unknown system name; errors.As extracts Name and ValidSystems fields
- Result.Total matches sum of Result.Letters values
- Result.System matches the system passed to Compute

## Testing Style

Use table-driven subtests (t.Run). Assert typed errors with errors.As not string comparison. Avoid testing implementation details — test behavior through the public API only.

## Spec References

- docs/specs/code-architecture.md §7 (Testing Strategy: pure unit tests, table-driven, errors.As)
- docs/specs/requirements.md §1 (27 letters), §3 (all 4 systems with specific values)
- docs/specs/cli-design.md §3.3 (Levenshtein threshold)

## Acceptance Criteria

- [ ] Parse() tested with Hebrew input, Latin input, mixed input
- [ ] Parse() error cases use errors.As to assert error type and fields (not string match)
- [ ] Levenshtein suggestion threshold tested: close typos get suggestions, garbage input does not
- [ ] Compute() tested for all four systems with at least one known value each
- [ ] Compute() Atbash test confirms substitution before value lookup
- [ ] Result.Total == sum of Letters[i].Value for all test cases
- [ ] All tests are table-driven using t.Run
- [ ] make test passes


## Notes

**2026-04-15T12:32:00Z**

Added 7 new table-driven tests to gematria_test.go: TestParseAllHebrewLetters (all 27 runes via Parse), TestParseLatin_MultiToken (3-token Latin → letter sequence), TestParseLatin_CaseInsensitive (case variants via Parse), TestComputeHechrachi_Emet (אמת=441), TestComputeGadol_AllSofitValues (all 5 sofit extended values 500–900), TestComputeAtbash_Tav (tav→aleph value 1), TestComputeResultTotalInvariant (Total==sum(Letters[i].Value) across 6 input/system pairs). Note: InvalidCharError cannot be returned from Parse() via normal input — the pre-scan labels any non-Hebrew rune as Latin, so parsePureHebrew is never called with an invalid character. InvalidCharError is already tested directly via LookupLetter.
