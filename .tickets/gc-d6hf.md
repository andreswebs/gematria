---
id: gc-d6hf
status: closed
deps: [gc-0mhe, gc-6c6c, gc-ahoy]
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, transliterate-fn, task]
---
# Implement Transliterate(input, scheme) public function

Implement the public `Transliterate` function in `src/transliteration.go`. This is the orchestration layer between the public API and the per-scheme transliterators (Tasks 3 and 4).

## Function signature

```go
// Transliterate parses input as a Hebrew word using the given scheme,
// returning the resolved sequence of letters.
//
// Returns *InvalidSchemeError if scheme is not recognized.
// Returns *UnknownWordError (with Input, Scheme, Position) if the input
// cannot be resolved to any Hebrew letters.
//
// Each space-separated token is processed independently (per spec §3.4).
// Hebrew Unicode runes mixed into the input are passed through unchanged
// per the resolved Task 1 rule.
func Transliterate(input string, scheme Scheme) ([]Letter, error)
```

## Behavior

1. Validate `scheme`: if not in `ValidSchemes()`, return `*InvalidSchemeError{Name: string(scheme), Valid: ValidSchemes()}`.
2. Dispatch to the per-scheme transliterator based on `scheme`:
   - `SchemeAcademic` → `transliterateAcademic(input)`
   - `SchemeIsraeli` → `transliterateIsraeli(input)`
3. Convert each returned rune to its `Letter` via the existing `letters` map (from `src/letters.go`).
4. Wrap any per-scheme error in `*UnknownWordError` with `Input` (the original Latin input), `Scheme`, and `Position` (token index where the error occurred).
5. Return `[]Letter` in input order.

## Multi-token handling

Per Task 1 spec resolution: typically split on spaces and process each token independently, but Hebrew runes (if present in the input) pass through unchanged. The exact split-vs-phrase behavior is decided in Task 1 (Q5).

## Spec References

- docs/specs/transliteration.md §3.4 (multi-token), §5 (errors), §6.1 (public API)
- See parent epic for how Transliterate composes with Compute (Task 6).

## Acceptance Criteria

- [ ] `Transliterate(input, scheme) ([]Letter, error)` exported in `src/transliteration.go`
- [ ] Returns `*InvalidSchemeError` for unknown scheme
- [ ] Returns `*UnknownWordError` for unmappable input with all fields populated (Input, Scheme, Position)
- [ ] Dispatches correctly to transliterateAcademic vs. transliterateIsraeli
- [ ] Multi-token input handled per the resolved Task 1 rule
- [ ] godoc comment documents behavior, error returns, and multi-token semantics
- [ ] At least 5 unit tests covering scheme dispatch, error cases, and multi-token input
- [ ] `make build` passes


## Notes

**2026-04-15T19:57:58Z**

Implemented Transliterate(input, scheme) in src/transliteration.go. The function: (1) validates scheme via a switch, returning *InvalidSchemeError for unknowns; (2) splits input with strings.Fields; (3) for each word-part, checks isAllHebrew() — all-Hebrew parts pass through via LookupLetter; Latin parts dispatch to transliterateAcademic or transliterateIsraeli; (4) wraps per-scheme errors in *UnknownWordError with Input, Scheme, and Position (0-based token index) set. Also added transliteratePart and isAllHebrew as package-private helpers in transliteration.go. Added 8 unit tests covering: academic dispatch, israeli dispatch, invalid scheme, unmappable input, multi-token concatenation, Hebrew passthrough, and position field accuracy. All tests follow TDD red-green cycles. make build passes.
