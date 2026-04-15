---
id: wor-24my
status: closed
deps: []
links: []
created: 2026-04-15T03:53:22Z
type: epic
priority: 1
tags: [reverse-lookup, epic]
---
# Reverse Lookup

Implement the reverse lookup feature: given a target gematria value, find all Hebrew words in a user-supplied word list whose computed value matches it. Invoked via --find <value> with a word list source (--wordlist or GEMATRIA_WORDLIST). Results are paginated via --limit / GEMATRIA_LIMIT (default 20). Output respects --output across all four formats (line/value/card/json).

## Spec References
- docs/specs/requirements.md §6 (Reverse Lookup) and §7 (Word List Format)
- docs/specs/cli-design.md §7 (Reverse Lookup Design)
- docs/specs/code-architecture.md §3.6 (WordSource Interface), §3.4 (High-Level API)
- docs/specs/wordlist-backends.md (interface rationale and future extensibility)

## Tasks

1. word-type-and-interface — Define Word type and WordSource interface in root package (wordlist.go)
2. parse-word-list — Implement ParseWordList(io.Reader) WordSource with plain text and TSV support
3. find-by-value-fn — Implement root-package FindByValue function and DefaultLookupLimit constant
4. cli-find-flags — Add --find, --wordlist, --limit flags and GEMATRIA_WORDLIST, GEMATRIA_LIMIT env vars to Config
5. cli-find-orchestration — Orchestrate reverse lookup in Run(): open word list, call FindByValue, handle all errors
6. lookup-output-formats — Implement FormatLookup in all four formatters (line, value, card, json)
7. wordlist-tests — Unit tests for ParseWordList and in-memory WordSource.FindByValue
8. cli-integration-tests — CLI integration tests for the full --find workflow

## Implementation Plan

### Root Package: src/wordlist.go

**Word type:**
```go
type Word struct {
    Hebrew          string // Hebrew Unicode text (required)
    Transliteration string // Latin transliteration (optional, TSV col 2)
    Meaning         string // English meaning (optional, TSV col 3)
}
```

**WordSource interface** (from code-architecture.md §3.6):
```go
type WordSource interface {
    FindByValue(value int, system System, limit int) ([]Word, bool, error)
}
```
Unexported concrete type wordList implements WordSource. ParseWordList returns the interface (hide implementation, expose interface per Effective Go). Add compile-time check: `var _ WordSource = (*wordList)(nil)`.

**ParseWordList:**
- Accepts io.Reader (root package never imports os)
- Uses bufio.Scanner line by line
- Skip blank lines and lines starting with '#'
- Line with tab character: split on '\t', up to 3 fields → Word{Hebrew, Transliteration, Meaning}
- Line without tab: entire trimmed line → Word{Hebrew: line}
- Words with invalid Hebrew stored as-is; FindByValue silently skips them (Compute returns error → skip)

**wordList.FindByValue(value, system, limit):**
- Iterate over stored words; for each call Compute(w.Hebrew, system)
- Skip words where Compute returns an error
- Collect matching words into a results slice, stop collecting after limit+1 matches
- Return results[:min(len,limit)], len(results) > limit, nil

### Root Package: src/gematria.go

**FindByValue standalone function:**
```go
const DefaultLookupLimit = 20

func FindByValue(value int, source WordSource, system System, limit int) ([]Word, bool, error) {
    if source == nil { return nil, false, errors.New("gematria: nil WordSource") }
    if limit <= 0 { limit = DefaultLookupLimit }
    return source.FindByValue(value, system, limit)
}
```

### CLI Layer

**src/internal/cli/config.go — additions to Config struct:**
```go
FindValue int    // target value for --find; only meaningful when FindSet==true
FindSet   bool   // true when --find was explicitly provided on the command line
Wordlist  string // resolved wordlist path (--wordlist flag or GEMATRIA_WORDLIST)
Limit     int    // max results (--limit / GEMATRIA_LIMIT / DefaultLookupLimit)
```

New flags registered with pflag:
- --find <int>: triggers reverse lookup mode
- --wordlist <path>: word list file path (no short form)
- --limit / -l <int>: max results, default 20

Env var resolution (lazy, only when FindSet is true):
- GEMATRIA_WORDLIST: only used when --wordlist not provided; validated only when --find is active
- GEMATRIA_LIMIT: only used when --limit not provided; must be positive integer

**src/internal/cli/run.go — reverse lookup branch:**
```
if cfg.FindSet:
    wordlistPath = cfg.Wordlist
    if wordlistPath == "":
        print error: --find requires word list, show --wordlist / GEMATRIA_WORDLIST
        return 2
    f = os.Open(wordlistPath)
    if error: print "cannot open word list <path>: <err>", return 3
    defer f.Close()
    source = gematria.ParseWordList(f)
    words, hasMore, err = gematria.FindByValue(cfg.FindValue, source, cfg.System, cfg.Limit)
    if err: print formatter.FormatError(err), return 1
    print formatter.FormatLookup(words, hasMore)
    return 0
```

**src/internal/cli/output.go — Formatter interface:**
```go
type Formatter interface {
    FormatLetter(gematria.Result) string
    FormatWord(gematria.Result) string
    FormatLookup(words []gematria.Word, hasMore bool) string
    FormatError(err error) string
}
```

**FormatLookup per formatter:**

LineFormatter: one line per word: `‏<Hebrew>‎ (<transliteration>) — <meaning>`. If hasMore: trailing line `(more results available — increase --limit to see them)`.

ValueFormatter: one Hebrew word per line, no other content. hasMore: no indicator (value format strips all presentation).

CardFormatter: header block with value+system, numbered entries each showing Hebrew, transliteration, meaning on separate indented lines. hasMore: trailing line `(more results available — increase --limit to see them)`.

JSONFormatter:
```json
{
  "value": 376,
  "system": "hechrachi",
  "results": [{"word": "שלום", "transliteration": "shalom", "meaning": "peace"}],
  "hasMore": false
}
```
No ANSI codes, no RTL marks in JSON values.

