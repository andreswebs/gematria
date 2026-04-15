---
id: wor-jq3u
status: closed
deps: [wor-q01l, wor-g979, wor-93fx]
links: []
created: 2026-04-15T03:55:22Z
type: task
priority: 2
parent: wor-24my
tags: [reverse-lookup, wordlist-tests, task]
---
# Unit tests for ParseWordList and FindByValue

Write unit tests for the root package's ParseWordList function, in-memory wordList.FindByValue, and the exported gematria.FindByValue function. All tests live under src/ (no I/O, no filesystem, use strings.NewReader for word list input).

## Spec References
- docs/specs/code-architecture.md §7 (Testing Strategy: root package tests use strings.NewReader)
- docs/specs/requirements.md §6 (reverse lookup behavior), §7 (word list format)

## Implementation Details

**Test file:** src/wordlist_test.go (package gematria or gematria_test)

**ParseWordList tests (table-driven):**

| Input | Expected |
|-------|----------|
| empty reader | empty source, no error |
| single plain-text word "שלום" | Word{Hebrew:"שלום"} |
| TSV two cols "שלום\tshalom" | Word{Hebrew:"שלום", Transliteration:"shalom"} |
| TSV three cols "שלום\tshalom\tpeace" | Word{Hebrew:"שלום", Transliteration:"shalom", Meaning:"peace"} |
| line starting with '#' | skipped, not stored |
| blank line | skipped |
| mixed plain+TSV | all words stored |
| extra tabs beyond 3 cols | ignored gracefully |
| reader error (use custom errReader) | returns non-nil error |

For each case: call ParseWordList(strings.NewReader(input)), then exercise FindByValue with a known matching value to verify words were parsed correctly.

**wordList.FindByValue tests (table-driven) — test via ParseWordList + FindByValue:**

| Scenario | Expected |
|----------|----------|
| value matches one word, limit=20 | returns 1 word, hasMore=false |
| value matches 3 words, limit=2 | returns 2 words, hasMore=true |
| value matches 2 words, limit=2 | returns 2 words, hasMore=false (exactly limit) |
| value matches 0 words | returns nil, hasMore=false, err=nil |
| word with invalid Hebrew stored | silently skipped (no match, no error) |
| limit=0 passed to root-pkg FindByValue | normalised to 20 by root-pkg fn |
| different systems (hechrachi vs gadol) | returns different matches |

Use real Hebrew words with known gematria values (e.g., "א"=1 in Hechrachi, "א"=1 in Gadol) to avoid dependency on System constants not yet implemented. Or use single-letter words for deterministic values.

**exported gematria.FindByValue tests:**
- nil source → non-nil error, no panic
- limit=0 → delegates with limit=20 (verify via result count)
- limit=1 with 2+ matching words → hasMore=true

**Test helpers:**
- Use strings.NewReader for all word list input
- For errReader: a simple struct implementing io.Reader that returns an error on Read
- Follow go-testing patterns: table-driven where applicable, t.Helper() for shared setup, errors.Is for error type assertions

**Failure message format (from go-testing skill):**
```go
t.Errorf("ParseWordList(%q): got %v words, want %v", input, len(got), want)
```

## Acceptance Criteria

- [ ] Tests cover: empty input, plain text, 2-col TSV, 3-col TSV, comment lines, blank lines, mixed
- [ ] Tests cover: reader error propagation from ParseWordList
- [ ] Tests cover: FindByValue returns correct matches and hasMore for below-limit, at-limit, above-limit cases
- [ ] Tests cover: nil source returns error (not panic) from exported FindByValue
- [ ] Tests cover: limit=0 normalization via exported FindByValue
- [ ] Tests cover: invalid Hebrew word is silently skipped in FindByValue
- [ ] All tests use strings.NewReader (no file I/O)
- [ ] Failure messages include function name, inputs, got, want
- [ ] make test passes with race detector (make test-race)


## Notes

**2026-04-15T12:35:23Z**

Existing wordlist_test.go already covered most acceptance criteria. Added three missing tests: TestParseWordList_MixedPlainAndTSV, TestParseWordList_ExtraTabsIgnored, TestFindByValue_DifferentSystems. The ExtraTabsIgnored test exposed a real bug: SplitN(line, '\t', 3) caused the third column to capture everything including further tabs (e.g., 'peace\textra'). Fixed by switching to strings.Split (no limit) so only parts[0..2] are used and extra columns are naturally ignored. All tests pass with race detector.
