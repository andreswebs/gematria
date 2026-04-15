---
id: wor-cg08
status: closed
deps: [wor-gdda]
links: []
created: 2026-04-15T04:09:32Z
type: task
priority: 1
parent: wor-jxhz
tags: [core-compute, result-types, task]
---
# Define LetterResult and Result types

Add LetterResult and Result structs to the root package. These are the exported types returned by Compute() and consumed by all CLI output formatters.

## Implementation Details

LetterResult embeds a full Letter (from the letter dictionary — name, meaning, position, all system values) plus the computed Value int for the selected system. This self-contained embedding means formatters can render any output without additional lookups.

Result holds:
- Input string — the original input as provided
- System System — which system was used
- Total int — sum of all letter values
- Letters []LetterResult — per-letter breakdown

Place these types in src/gematria.go (or a new src/result.go if that keeps the file cleaner).

The Letter embedding in LetterResult allows card and json formatters to access name, meaning, and position without any further dictionary queries. This is the "self-contained result" design from the architecture spec.

## Spec References

- docs/specs/code-architecture.md §3.5 (Result Type, LetterResult composition)
- docs/specs/requirements.md §4 (output formats all need per-letter breakdown)

## Acceptance Criteria

- [ ] LetterResult struct is exported with Char rune, Name string, Value int, and embedded or referenced Letter fields for meaning and position
- [ ] Result struct is exported with Input string, System System, Total int, Letters []LetterResult
- [ ] Both types compile with no imports from internal/ or os (root package stays pure)
- [ ] make build passes


## Notes

**2026-04-15T11:20:50Z**

Added src/result.go with LetterResult (embeds Letter + Value int) and Result (Input, System, Total, Letters). Tests in src/result_test.go verify field promotion via embedding and zero-value safety. Both types are pure data — no I/O, no os imports. make build passes.
