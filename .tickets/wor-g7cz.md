---
id: wor-g7cz
status: closed
deps: [wor-q01l]
links: []
created: 2026-04-15T03:55:02Z
type: task
priority: 1
parent: wor-24my
tags: [reverse-lookup, lookup-output-formats, task]
---
# Implement FormatLookup in all four output formatters

Add FormatLookup(words []gematria.Word, hasMore bool) string to each of the four output formatters in src/internal/cli/. The Formatter interface in output.go must include this method.

## Spec References
- docs/specs/cli-design.md §4.3 (human output patterns), §4.4 (agent output patterns), §7.1 (result volume control), §7.3 (enriched results)
- docs/specs/requirements.md §6 (--find output format respects --output), §7 (TSV transliteration/meaning in output)

## Implementation Details

**Formatter interface (src/internal/cli/output.go):**
```go
type Formatter interface {
    FormatLetter(r gematria.Result) string
    FormatWord(r gematria.Result) string
    FormatLookup(words []gematria.Word, hasMore bool) string
    FormatError(err error) string
}
```

---

**LineFormatter.FormatLookup** (src/internal/cli/line.go):
Each word on its own line. Include optional transliteration and meaning when present.
Format per word: `\u200F<Hebrew>\u200E` (RTL+LTR marks around Hebrew), then ` (<transliteration>)` if non-empty, then ` — <meaning>` if non-empty.
If no words: single line `(no results)`.
If hasMore: trailing line `(more results available — increase --limit to see them)`.
Example:
```
‏שלום‎ (shalom) — peace
‏אמת‎ (emet) — truth
(more results available — increase --limit to see them)
```
Color: apply green to the Hebrew word when color is enabled (same rule as other value output).

---

**ValueFormatter.FormatLookup** (src/internal/cli/value.go):
One Hebrew word per line, bare — no transliteration, no meaning, no RTL marks, no hasMore indicator.
The value format strips all presentation (cli-design.md §4.3: "--atbash has no visible effect in value format").
Example:
```
שלום
אמת
```
If no words: empty string (no output).

---

**CardFormatter.FormatLookup** (src/internal/cli/card.go):
Header block showing the target value and system. Numbered entries, each on multiple indented lines.
Example:
```
Reverse Lookup: 376
System: hechrachi

1. ‏שלום‎
   Transliteration: shalom
   Meaning: peace

2. ‏אמת‎
   Transliteration: emet
   Meaning: truth

(more results available — increase --limit to see them)
```
Omit Transliteration/Meaning lines when those fields are empty.
If no words: `(no results)`.

---

**JSONFormatter.FormatLookup** (src/internal/cli/json.go):
Single JSON object. No ANSI codes, no RTL marks in JSON string values.
Schema (stable per cli-design.md §4.4):
```json
{
  "value": 376,
  "system": "hechrachi",
  "results": [
    {
      "word": "שלום",
      "transliteration": "shalom",
      "meaning": "peace"
    }
  ],
  "hasMore": false
}
```
Omit "transliteration" and "meaning" keys when empty (or include as empty strings — pick one and be consistent). End with a trailing newline.
On empty results: results is [] and hasMore is false.

**Note on target value availability:** The formatter receives only (words, hasMore). To include the target value and system in card/json output, the Formatter implementations need access to cfg.FindValue and cfg.System. These may be passed to the formatter constructor or held as fields on the formatter struct. Exact approach determined during implementation — the acceptance criteria test the output shape.

## Acceptance Criteria

- [ ] Formatter interface in output.go includes FormatLookup(words []gematria.Word, hasMore bool) string
- [ ] LineFormatter: one line per word with RTL/LTR marks, optional transliteration and meaning, hasMore indicator
- [ ] LineFormatter: empty results prints "(no results)"
- [ ] ValueFormatter: one Hebrew word per line, no other content, no hasMore indicator
- [ ] ValueFormatter: empty results produces empty output
- [ ] CardFormatter: header with value+system, numbered entries, optional transliteration/meaning lines, hasMore indicator
- [ ] CardFormatter: empty results prints "(no results)"
- [ ] JSONFormatter: emits valid JSON object with value, system, results array, hasMore boolean
- [ ] JSONFormatter: no ANSI codes or RTL marks in JSON string values
- [ ] JSONFormatter: trailing newline present
- [ ] All formatters: no stdout output when words is nil/empty except LineFormatter and CardFormatter "(no results)"
- [ ] make vet and make lint pass


## Notes

**2026-04-15T12:12:49Z**

Implemented FormatLookup in all four output formatters. Added NewFormatterWithLookup(output, useColor, showAtbash, findValue, findSystem) alongside the existing NewFormatter (which now delegates to NewFormatterWithLookup with 0/empty defaults). cardFormatter and jsonFormatter gained findValue/findSystem fields used by their FormatLookup implementations. Line: RTL-wrapped Hebrew + optional transliteration + optional meaning + hasMore indicator + (no results) on empty. Value: bare Hebrew words only, no metadata. Card: header with value+system, numbered entries with optional transliteration/meaning lines. JSON: object with value, system, results[], hasMore — no RTL marks or ANSI. All tests pass via TDD red-green cycles.
