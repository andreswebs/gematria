---
id: wor-q01l
status: closed
deps: []
links: []
created: 2026-04-15T03:53:35Z
type: task
priority: 1
parent: wor-24my
tags: [reverse-lookup, word-type-and-interface, task]
---
# Define Word type and WordSource interface

Create src/wordlist.go in the root package (github.com/andreswebs/gematria). Define the Word struct and WordSource interface that all other reverse-lookup tasks build on.

## Spec References
- docs/specs/code-architecture.md §3.6 (WordSource interface), §3.5 (Result type)
- docs/specs/requirements.md §7 (Word List Format)
- docs/specs/wordlist-backends.md (interface rationale)

## Implementation Details

**File:** src/wordlist.go

**Word struct:**
```go
// Word is a single entry from a word list used for reverse lookups.
// Hebrew is required. Transliteration and Meaning are optional and come
// from the second and third tab-separated columns of a TSV word list.
type Word struct {
    Hebrew          string
    Transliteration string
    Meaning         string
}
```

**WordSource interface:**
```go
// WordSource is the interface for reverse-lookup word-list backends.
// ParseWordList returns the only current in-memory implementation.
// Future backends (embedded database, pre-computed index) must satisfy
// this interface without changing the root package API.
//
// FindByValue returns at most limit Words whose gematria value in system
// equals value. hasMore is true when additional matching words exist
// beyond the returned slice. Returns (nil, false, nil) when no words match.
type WordSource interface {
    FindByValue(value int, system System, limit int) ([]Word, bool, error)
}
```

**Design notes:**
- The interface exists for future extensibility (wordlist-backends.md), not over-engineering.
- The unexported concrete type (wordList) added in the parse-word-list task must satisfy this interface; add compile-time check `var _ WordSource = (*wordList)(nil)` in wordlist.go.
- WordSource carries no io.Closer — backends that need cleanup implement io.Closer separately; CLI checks via type assertion (see wordlist-backends.md open question 2).
- Root package never imports "os" — Word and WordSource are pure data/behavior types.

## Acceptance Criteria

- [ ] src/wordlist.go exists in package gematria
- [ ] Word struct has Hebrew, Transliteration, Meaning string fields
- [ ] WordSource interface defines FindByValue(value int, system System, limit int) ([]Word, bool, error)
- [ ] Both types are exported with doc comments
- [ ] File compiles cleanly (make vet passes)
- [ ] Root package does not import "os" or any I/O package


## Notes

**2026-04-15T11:34:53Z**

Created src/wordlist.go with Word struct and WordSource interface. Compile-time check (var _ WordSource = (*wordList)(nil)) was omitted since wordList is defined in the parse-word-list task (wor-g979) — add it there. No os imports, pure data types. make build passes.
