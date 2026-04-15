---
id: wor-tyjq
status: closed
deps: [wor-22kl, wor-g979, wor-93fx]
links: []
created: 2026-04-15T03:54:39Z
type: task
priority: 1
parent: wor-24my
tags: [reverse-lookup, cli-find-orchestration, task]
---
# Orchestrate reverse lookup in Run()

Add the reverse lookup execution branch to src/internal/cli/run.go. When --find is active, load the word list file, call gematria.FindByValue, render output through the Formatter, and handle all error cases with correct exit codes.

## Spec References
- docs/specs/requirements.md §6 (reverse lookup behavior), §7 (word list error handling)
- docs/specs/cli-design.md §7.2 (word list error handling), §6.2 (exit codes)
- docs/specs/code-architecture.md §4.1 (Run signature), §4.3 (Formatter interface)

## Implementation Details

**Branch in Run() after config resolution and formatter construction:**
```go
if cfg.FindSet {
    // Step 1: resolve word list path (already in cfg.Wordlist from config resolution)
    if cfg.Wordlist == "" {
        fmt.Fprintln(stderr, formatter.FormatError(
            errors.New("--find requires a word list: use --wordlist <path> or set GEMATRIA_WORDLIST")))
        return 2
    }

    // Step 2: open file
    f, err := os.Open(cfg.Wordlist)
    if err != nil {
        fmt.Fprintln(stderr, formatter.FormatError(
            fmt.Errorf("cannot open word list %q: %w", cfg.Wordlist, err)))
        return 3
    }
    defer f.Close()

    // Step 3: parse word list
    source, err := gematria.ParseWordList(f)
    if err != nil {
        fmt.Fprintln(stderr, formatter.FormatError(
            fmt.Errorf("error reading word list %q: %w", cfg.Wordlist, err)))
        return 3
    }

    // Step 4: find by value
    words, hasMore, err := gematria.FindByValue(cfg.FindValue, source, cfg.System, cfg.Limit)
    if err != nil {
        fmt.Fprintln(stderr, formatter.FormatError(err))
        return 1
    }

    // Step 5: format and write output
    fmt.Fprint(stdout, formatter.FormatLookup(words, hasMore))
    return 0
}
```

**Error exit codes (from cli-design.md §6.2):**
- 2: CLI misuse — --find used without a word list path (no --wordlist and GEMATRIA_WORDLIST empty)
- 3: File error — word list file not found or unreadable
- 1: Input error — FindByValue returns an error (e.g., invalid system)

**Error message requirements (cli-design.md §7.2):**
- Missing word list: explain that --find requires a word list, show both --wordlist and GEMATRIA_WORDLIST.
- File not found: name the path tried and where it came from (flag vs env var).
- File unreadable: name the path and the OS error.

**Stdout empty on error:** stdout must have no output when any error occurs. The formatter.FormatError output goes to stderr only. fmt.Fprintln/Fprintf writes to stderr; fmt.Fprint(stdout, ...) only in the success path.

**Formatter usage:** Call formatter.FormatLookup(words, hasMore). The selected Formatter already knows the output format (line/value/card/json) — no switch statement needed here.

## Acceptance Criteria

- [ ] When --find is set and no word list is available: stderr gets error message, stdout is empty, exit code 2
- [ ] When word list file cannot be opened (not found, permission denied): stderr gets error naming the path and OS error, stdout empty, exit code 3
- [ ] When word list parses with an io error: stderr gets error, stdout empty, exit code 3
- [ ] When FindByValue returns an error: stderr gets error, stdout empty, exit code 1
- [ ] On success: formatter.FormatLookup output on stdout, nothing on stderr, exit code 0
- [ ] Error messages for missing word list mention both --wordlist flag and GEMATRIA_WORDLIST env var
- [ ] os.Open is used for file opening (CLI concern, not root package)
- [ ] make vet and make lint pass


## Notes

**2026-04-15T12:21:42Z**

Implemented runFind() function in run.go. Key points: (1) formatter construction branches on cfg.FindSet — uses NewFormatterWithLookup (with cfg.FindValue + cfg.Mispar) when --find is active so JSON/card formatters embed lookup context; (2) error path uses fmt.Fprint + formatter.FormatError consistent with computeArgs/processBatch pattern (no extra newline added by caller); (3) defer f.Close() uses _ = f.Close() to satisfy errcheck linter; (4) 6 new integration tests added to run_test.go covering success, no-wordlist exit 2, file-not-found exit 3, env var wordlist, no-results, and JSON output with value/system fields.
