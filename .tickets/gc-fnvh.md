---
id: gc-fnvh
status: closed
deps: [gc-h6f3]
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, unit-tests, task]
---
# Unit tests for Transliterate, scheme tables, errors, and refactored Compute

Write thorough unit tests in the root package covering all new transliteration functionality and confirming the Compute refactor is regression-free.

## Test files

- `src/transliteration_test.go` (NEW or extended in Tasks 3-5): tests for Transliterate, transliterateAcademic, transliterateIsraeli
- `src/errors_test.go` (extended): tests for UnknownWordError and InvalidSchemeError
- `src/result_test.go` (extended): tests for Result.Scheme field semantics
- `src/gematria_test.go` (verified — should pass unchanged): regression check for Compute

## Coverage checklist

**Per-scheme tables (Tasks 3 and 4)**: verified to be present, but consolidate any missing cases here.
- Table-driven: input string + scheme → expected []Letter (or expected error type)
- At least 10 cases per scheme covering multi-char sequences, vowels, sofit, ASCII fallbacks, ambiguous-combo resolutions
- Negative cases: garbage input → UnknownWordError with correct fields

**Transliterate dispatch**:
- Valid scheme: dispatches to correct per-scheme function
- Invalid scheme: returns InvalidSchemeError
- Multi-token input: per Task 1 resolution
- Mixed Hebrew/Latin input behavior

**Errors (errors_test.go)**:
- UnknownWordError.Error() format with and without suggestions
- InvalidSchemeError.Error() lists valid schemes
- errors.As round-trip succeeds for both new error types

**Result.Scheme**:
- Compute returns Result with Scheme == \"\"
- ComputeTransliterated returns Result with Scheme == requested scheme
- Atbash + transliterate composes correctly (Result.System=Atbash AND Result.Scheme=academic)

**Compute refactor regression**:
- All cases in existing gematria_test.go pass unchanged
- ComputeFromLetters returns identical Result to Compute when given the same letters

## Test style (per go-testing skill)

- Table-driven tests with `t.Run` subtests using clear names (e.g., `\"shalom_academic\"`)
- Failure messages: `Transliterate(%q, %q) = %v, want %v`
- Use `cmp.Diff(want, got)` for struct comparisons (Result, []Letter)
- Test error semantics with `errors.As`, never with `err.Error() == \"...\"`
- Use `t.Helper()` in test helpers
- No assertion libraries

## Spec References

- docs/specs/transliteration.md §4 (mappings), §5 (errors)
- docs/specs/code-architecture.md §7 (testing strategy)
- See parent epic.

## Acceptance Criteria

- [ ] Table-driven tests for academic scheme (≥ 10 cases) pass
- [ ] Table-driven tests for israeli scheme (≥ 10 cases) pass
- [ ] At least 5 negative-case tests (UnknownWordError) per scheme
- [ ] Tests for Transliterate covering valid dispatch, invalid scheme, multi-token, mixed input
- [ ] errors.As round-trip tests for both new error types
- [ ] Tests for Result.Scheme set/unset across all three compute functions
- [ ] All existing tests in gematria_test.go pass without modification
- [ ] No use of err.Error() string matching for semantic checks
- [ ] cmp.Diff used for struct comparisons
- [ ] `make test` passes; `make lint` clean

