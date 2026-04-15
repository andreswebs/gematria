---
id: gc-tw8e
status: closed
deps: []
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, spec-research, task]
---
# Spec research: finalize transliteration scheme mappings and rules

Resolve the open questions in transliteration.md (Q1-Q7) by producing concrete, documented mapping tables for both `academic` and `israeli` schemes. **This is a research and spec task — no Go code is written.**

The current spec defines the structure and behavior but leaves exact mapping tables and several edge-case rules as open questions. Implementation tasks (#3, #4) cannot proceed without these decisions.

## Open questions to resolve

- **Q1 — Academic scheme mapping table**: choose alignment (ISO 259, SBL, ALA-LC, or custom). Document every consonant mapping, multi-character sequence (sh, ts, kh, ch, etc.), and ASCII fallback for diacritics.
- **Q2 — Israeli scheme mapping table**: pick canonical resolutions for ambiguous combos (`ch` → ח or כ?, `kh` → ?, `s` → ס or ש?, `tz` vs `ts`). Document vowel-to-mater rules (when `o`/`u` becomes ו, `i`/`e` becomes י, `a` becomes א or ה).
- **Q4 — Sofit handling**: rule for when to use sofit forms (always at end-of-word, or based on internal token boundaries).
- **Q5 — Compound input** (`gematria -t \"shalom emet\"`): treat as one phrase, split on internal spaces and treat each as a separate word, or reject internal spaces.
- **Q6 — Capitalization**: case-insensitive globally, or per-scheme rules (e.g., `H` vs `h` distinction in academic for ḥ vs ה).
- **Q7 — Numerics and punctuation**: drop, error, or treat as separators.

## Output

Update `docs/specs/transliteration.md`:
- Move resolved questions from Section 7 into Section 4 (Schemes) as concrete tables and rules.
- Section 4 must contain complete mapping tables: every consonant, every multi-char sequence, every ambiguous combo resolution, every vowel rule.
- Section 7 reduced to genuinely deferred items only (e.g., suggestion sources for v2).
- Add citations or rationale for non-obvious choices.
- Spec status remains Draft until all implementation is complete (final flip to Ready by the epic).

## Spec References

- docs/specs/transliteration.md (sections 4 and 7)
- See parent epic for full architecture context.

## Acceptance Criteria

- [ ] transliteration.md Section 4.1 contains the complete academic consonant table (single chars + multi-char sequences + ASCII fallbacks)
- [ ] transliteration.md Section 4.2 contains the complete israeli consonant table + vowel-to-mater rules
- [ ] Each ambiguous English combo (sh, ts, kh, ch, s, h) has a documented per-scheme resolution
- [ ] Sofit handling rule documented (Section 4 or new subsection)
- [ ] Capitalization rule documented per scheme
- [ ] Compound input behavior documented (Section 3.4 update)
- [ ] Numeric/punctuation handling documented
- [ ] Section 7 reduced to truly deferred items only (e.g., suggestion sources)
- [ ] Citations/rationale present for non-obvious choices


## Notes

**2026-04-15T19:35:22Z**

Resolved all open questions Q1-Q7 by updating docs/specs/transliteration.md Section 4 with complete mapping tables. Key decisions: (1) Both schemes share the same 19-consonant table; ch/kh→ח in both. (2) Academic: vowels silently dropped, consonantal output; apostrophe for explicit aleph. (3) Israeli: vowel-to-mater rules — i→י, o/u→ו, word-initial a/e→א, word-initial o/u→או, word-final a/e→ה, medial a/e→drop. (4) Sofit: last letter of each word-part uses sofit form if available. (5) Capitalization: case-insensitive in both schemes. (6) Compound quoted input split on spaces, each part processed independently. (7) Digits/punctuation → UnknownWordError. Section 7 now only contains Q3 (suggestion source, deferred v2).
