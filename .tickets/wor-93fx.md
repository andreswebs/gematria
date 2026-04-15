---
id: wor-93fx
status: closed
deps: [wor-q01l]
links: []
created: 2026-04-15T03:54:02Z
type: task
priority: 1
parent: wor-24my
tags: [reverse-lookup, find-by-value-fn, task]
---
# Implement root-package FindByValue function

Add the exported FindByValue function and DefaultLookupLimit constant to the root package (src/gematria.go or a new src/findbyvalue.go). This is the high-level public API for reverse lookup that delegates to a WordSource.

## Spec References
- docs/specs/code-architecture.md §3.4 (FindByValue signature in high-level API)
- docs/specs/requirements.md §6 (limit defaults to 20)

## Implementation Details

**Location:** Add to src/gematria.go (preferred) or src/wordlist.go — whichever keeps package organization cleaner.

**Constant:**
```go
// DefaultLookupLimit is the default maximum number of results returned
// by FindByValue when limit is zero or negative.
const DefaultLookupLimit = 20
```

**Function:**
```go
// FindByValue finds words in source whose gematria value under system
// equals value, returning at most limit results.
//
// If limit is zero or negative, DefaultLookupLimit (20) is used.
// Returns (nil, false, errors.New(...)) if source is nil.
// Otherwise returns the result of source.FindByValue directly.
func FindByValue(value int, source WordSource, system System, limit int) ([]Word, bool, error) {
    if source == nil {
        return nil, false, errors.New("gematria: FindByValue called with nil WordSource")
    }
    if limit <= 0 {
        limit = DefaultLookupLimit
    }
    return source.FindByValue(value, system, limit)
}
```

**Design notes:**
- The function validates the nil-source invariant at the library boundary so internal code never needs to guard against it.
- Limit normalization happens here, not in wordList.FindByValue, because the default is a domain-level constant that all current and future WordSource implementations should benefit from.
- The function does not open files or touch the filesystem — that is a CLI concern.

## Acceptance Criteria

- [ ] DefaultLookupLimit constant equals 20 and is exported with doc comment
- [ ] FindByValue(value int, source WordSource, system System, limit int) ([]Word, bool, error) exported
- [ ] nil source returns a non-nil error (not a panic)
- [ ] limit <= 0 is normalized to DefaultLookupLimit before delegating
- [ ] Positive limit is passed through unchanged
- [ ] Function delegates to source.FindByValue (does not duplicate matching logic)
- [ ] make vet and make lint pass


## Notes

**2026-04-15T12:14:49Z**

Added DefaultLookupLimit=20 constant and package-level FindByValue(value, source, system, limit) to src/gematria.go. Nil-source guard returns error (not panic). Limit<=0 normalizes to DefaultLookupLimit before delegating to source.FindByValue. Tests added to src/wordlist_test.go covering: nil source, constant value, zero limit, negative limit, positive limit passthrough, and delegation correctness. make build passes clean.
