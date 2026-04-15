---
id: wor-g979
status: closed
deps: [wor-q01l]
links: []
created: 2026-04-15T03:53:50Z
type: task
priority: 1
parent: wor-24my
tags: [reverse-lookup, parse-word-list, task]
---
# Implement ParseWordList with plain text and TSV support

Implement ParseWordList(io.Reader) (WordSource, error) in src/wordlist.go. This is the only current in-memory WordSource implementation. It parses plain text and TSV Hebrew word lists.

## Spec References
- docs/specs/requirements.md §7 (Word List Format)
- docs/specs/code-architecture.md §3.6 (ParseWordList accepts io.Reader, not file path)

## Implementation Details

**ParseWordList signature:**
```go
// ParseWordList parses a word list from r and returns an in-memory WordSource.
// It accepts two formats:
//   - Plain text: one Hebrew word per line.
//   - TSV: word[TAB]transliteration[TAB]meaning (transliteration and meaning optional).
// Blank lines and lines beginning with '#' are silently ignored.
// Returns an error only if reading from r fails.
func ParseWordList(r io.Reader) (WordSource, error)
```

**Unexported wordList type (add to wordlist.go):**
```go
type wordList struct {
    words []Word
}

// Compile-time check added after wordList is defined:
var _ WordSource = (*wordList)(nil)
```

**Parsing algorithm (use bufio.Scanner):**
1. Create scanner over r with default split (ScanLines).
2. For each scanned line:
   a. Trim leading/trailing whitespace.
   b. Skip if blank or starts with '#'.
   c. If line contains '\t': split on '\t' (strings.SplitN with n=3). Assign fields to Hebrew, Transliteration, Meaning in order. Extra tabs beyond 3 columns are ignored.
   d. Otherwise: the entire (trimmed) line is the Hebrew field.
3. If scanner.Err() != nil, return nil, scanner.Err().
4. Return &wordList{words: collected}, nil.

**wordList.FindByValue(value, system, limit) implementation:**
```go
func (wl *wordList) FindByValue(value int, system System, limit int) ([]Word, bool, error) {
    var matches []Word
    for _, w := range wl.words {
        result, err := Compute(w.Hebrew, system)
        if err != nil {
            continue // silently skip words with invalid/empty Hebrew
        }
        if result.Value == value {
            matches = append(matches, w)
            if len(matches) == limit+1 {
                // Collected one more than limit — stop early, signal hasMore
                return matches[:limit], true, nil
            }
        }
    }
    return matches, false, nil
}
```
Stopping at limit+1 avoids scanning the entire list when unnecessary.

**Imports needed in wordlist.go:** "bufio", "io", "strings" — no "os".

## Acceptance Criteria

- [ ] ParseWordList(io.Reader) (WordSource, error) exists in package gematria
- [ ] Returns error only on io.Reader read failure; invalid Hebrew words are silently stored
- [ ] Plain text lines (no tab) stored as Word{Hebrew: line}
- [ ] TSV lines split on tab into up to 3 fields: Hebrew, Transliteration, Meaning
- [ ] Lines starting with '#' are skipped
- [ ] Blank lines are skipped
- [ ] Compile-time check var _ WordSource = (*wordList)(nil) present
- [ ] wordList.FindByValue returns up to limit matches; hasMore=true when more exist
- [ ] Words where Compute() fails are silently skipped in FindByValue
- [ ] Root package imports no "os" package
- [ ] make vet and make lint pass


## Notes

**2026-04-15T12:04:21Z**

Implemented ParseWordList(io.Reader) and unexported wordList type in wordlist.go. Added compile-time interface check (var _ WordSource = (*wordList)(nil)). Uses bufio.Scanner for line-by-line parsing; blank lines and '#' comments are skipped; tab-separated lines split into up to 3 fields (Hebrew, Transliteration, Meaning); plain-text lines use the entire trimmed line as Hebrew. wordList.FindByValue iterates words, calls Compute, skips errors, and stops early (limit+1 sentinel) to set hasMore. Tests in wordlist_test.go cover plain text, TSV 2/3 columns, blank/comment skipping, no-match, invalid-Hebrew skip, limit enforcement, hasMore flag, empty reader, and reader error propagation. All pass; make build clean.
