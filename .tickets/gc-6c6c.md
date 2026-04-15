---
id: gc-6c6c
status: closed
deps: [gc-tw8e]
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, academic-scheme, task]
---
# Implement academic transliteration scheme

Implement the academic scheme — strict consonantal mapping per the tables finalized in Task 1. Place data and per-scheme logic in `src/transliteration_academic.go`.

## Components

- **Mapping tables** (Go map literals, package-level vars):
  - `academicMultiChar map[string][]rune` — multi-character sequences (e.g., `\"sh\" → ש`)
  - `academicSingleChar map[byte]rune` — single Latin character mappings
  - ASCII fallbacks for diacritic characters per spec table
- **Vowel handling**: per Task 1 spec — drop or selective mater per the academic rule.
- **Sofit handling**: per Task 1 spec (e.g., always-final or word-boundary).
- **Internal function**: `transliterateAcademic(input string) ([]rune, error)` — returns the resolved Hebrew runes (no Letter conversion yet; that happens in Transliterate, Task 5).

## Algorithm

1. Lowercase (or apply per-scheme capitalization rule from Task 1).
2. Greedy left-to-right scan with longest-match-first multi-char lookup (try 2-char, then 1-char).
3. For each match: append the mapped rune(s) to output. If unmappable: return error (Transliterate wraps it in UnknownWordError).
4. Apply sofit transformation per the rule (likely a post-pass that swaps the last consonant of each word with its sofit form when applicable).

## Testing (in transliteration_test.go)

Add table-driven tests using inputs and expected outputs from the Task 1 spec table. At least 10 cases exercising:
- Multi-char sequences (sh, ts, etc.)
- Vowel handling (drop or selective)
- Sofit transformation at word end
- ASCII fallbacks for diacritics
- Negative cases (returns error)

## Spec References

- docs/specs/transliteration.md §4.1 (academic), §4.4 (strict mapping)
- Resolved tables from Task 1 (depends on)
- See parent epic for the orchestration into Transliterate.

## Acceptance Criteria

- [ ] `academicMultiChar` and `academicSingleChar` Go map literals defined per spec
- [ ] All consonants and multi-char sequences from the spec table are present
- [ ] ASCII fallbacks documented in code comments next to table entries
- [ ] `transliterateAcademic` greedy longest-match algorithm implemented
- [ ] Vowel handling matches spec
- [ ] Sofit transformation applied per spec rule
- [ ] At least 10 table-driven test cases pass (covering multi-char, vowels, sofit, fallbacks)
- [ ] Internal function returns error for unmappable input
- [ ] `make build` passes; `make test` passes


## Notes

**2026-04-15T19:44:21Z**

Implemented src/transliteration_academic.go with academicMultiChar (sh/kh/ch/ts/tz/ph), academicSingleChar (19 consonants + apostrophe→Aleph), academicVowels (drop a/e/i/o/u), and sofitMap (כ→ך מ→ם נ→ן פ→ף צ→ץ). The transliterateAcademic() function does greedy 2-then-1 char scan, silently drops vowels, errors on unmappable chars, errors on empty result, and applies sofit to the last letter. sofitMap is defined here (not in transliteration_israeli.go) since it is shared by both schemes. Added 23 table-driven tests covering all behaviors in spec §4.1. Tests live in transliteration_test.go alongside the existing Scheme constant tests.
