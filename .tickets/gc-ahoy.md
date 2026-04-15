---
id: gc-ahoy
status: closed
deps: [gc-tw8e]
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, israeli-scheme, task]
---
# Implement israeli transliteration scheme

Same shape as Task 3 but for the israeli scheme. Israeli uses matres lectionis (vowels mapped to ו/י/א/ה) and israeli-specific resolutions for ambiguous combos.

## Components

- `israeliMultiChar map[string][]rune`
- `israeliSingleChar map[byte]rune`
- `israeliVowels map[byte]rune` — vowel-to-mater mapping per spec
- `transliterateIsraeli(input string) ([]rune, error)`

Place in `src/transliteration_israeli.go`.

## Differences from academic

- Vowels become ו (for o/u), י (for i/e), א (silent), ה (silent terminal) per the spec rules from Task 1.
- Multi-char resolutions for ambiguous combos may differ from academic (e.g., `ch` may map to ח while academic uses `H` for ḥ).
- Sofit handling per Task 1 spec (likely identical to academic but verified per the spec).

## Testing (in transliteration_test.go)

At least 10 table-driven cases including the canonical examples:
- `shalom` → שלום (vav inserted as o-mater)
- `gadol` → גדול (vav inserted as o-mater)
- `emet` → אמת (no maters needed mid-word)
- ambiguous combos exercising the israeli-specific resolutions per Task 1
- terminal-vowel handling (e.g., `ahava` → אהבה if Task 1 says so)

## Spec References

- docs/specs/transliteration.md §4.2 (israeli), §4.3 (illustrative differences)
- Resolved tables from Task 1 (depends on)
- See parent epic.

## Acceptance Criteria

- [ ] Israeli scheme tables defined as Go map literals
- [ ] Vowel-to-mater mapping implemented per spec
- [ ] Multi-char resolutions match spec (israeli-specific)
- [ ] Sofit transformation applied per spec rule
- [ ] At least 10 table-driven test cases pass
- [ ] Test coverage includes shalom, gadol, emet, and at least 2 ambiguous-combo cases
- [ ] Returns error for unmappable input
- [ ] `make build` passes; `make test` passes


## Notes

**2026-04-15T19:50:46Z**

Implemented transliteration_israeli.go with transliterateIsraeli(). Consonant tables reuse academicMultiChar and academicSingleChar (identical per spec §4.2). Vowels are position-aware: a/e word-initial→א, medial→drop, word-final→ה; i→י always; o/u word-initial→או, non-initial→ו. israeliHasConsonantAfter() does lookahead to detect word-final position. israeliVowels map is used inside the lookahead to skip vowels explicitly (satisfies unused linter). Sofit substitution uses shared sofitMap from transliteration_academic.go. 30 test cases added covering all 7 spec examples, ambiguous combos, position rules, sofit, case folding, and error cases.
