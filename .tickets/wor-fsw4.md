---
id: wor-fsw4
status: closed
deps: [wor-22kl, wor-tyjq, wor-g7cz, wor-jq3u]
links: []
created: 2026-04-15T03:55:44Z
type: task
priority: 2
parent: wor-24my
tags: [reverse-lookup, cli-integration-tests, task]
---
# CLI integration tests for --find workflow

Write integration tests for the full --find reverse lookup workflow in src/internal/cli/. Tests call Run() directly with injected args, fake stdin/stdout/stderr (via os.Pipe()), fake getenv closures, and temporary word list files.

## Spec References
- docs/specs/code-architecture.md §7 (Testing Strategy: internal/cli uses Run() with injected primitives)
- docs/specs/requirements.md §6 (all acceptance criteria for reverse lookup)
- docs/specs/cli-design.md §6.2 (exit codes), §7.2 (word list error handling), §7.3 (enriched results)

## Implementation Details

**Test file:** src/internal/cli/run_find_test.go (package cli or cli_test)

**Test infrastructure:**
Create a helper runWithArgs that:
1. Creates os.Pipe() pairs for stdout and stderr.
2. Writes a temporary word list file via os.CreateTemp (or t.TempDir()).
3. Calls Run(args, os.Stdin, stdoutW, stderrW, getenv).
4. Closes write ends, reads stdout/stderr to strings.
5. Returns (exitCode int, stdoutStr string, stderrStr string).

**Test cases:**

Exit-code and error routing:
- --find 376 with no word list configured → exit 2, stderr non-empty, stdout empty
- --find 376 --wordlist /nonexistent/path → exit 3, stderr names the path, stdout empty
- --find 376 --wordlist <valid file> → exit 0, stdout non-empty, stderr empty
- GEMATRIA_WORDLIST=<path> --find 376 (no --wordlist flag) → exit 0 (env var used)
- GEMATRIA_WORDLIST="" --find 376 → exit 2

Flag precedence:
- --wordlist flag overrides GEMATRIA_WORDLIST when both set
- --limit 1 with a word list containing 2+ matches → exactly 1 result in output, hasMore indicator present (for line/card formats)
- GEMATRIA_LIMIT=1 (no --limit flag) with 2+ matches → 1 result
- --limit 5 overrides GEMATRIA_LIMIT=1 → 5 results returned

Output format coverage:
- --find 376 --wordlist <tsv file> --output line → output contains Hebrew word with RTL marks and transliteration
- --find 376 --wordlist <tsv file> --output value → output is one word per line, no transliteration
- --find 376 --wordlist <tsv file> --output json → output is valid JSON, hasMore field present
- --find 376 --wordlist <tsv file> --output card → multi-line output with numbered entries

Env var validation (lazy):
- GEMATRIA_LIMIT=notanumber --find 376 --wordlist <file> → exit 2, stderr mentions GEMATRIA_LIMIT
- GEMATRIA_LIMIT=notanumber without --find → exit 0 (lazy: invalid env var not validated)

JSON structured error (--output json):
- --find 376 (no word list) --output json → stderr contains valid JSON error object, stdout empty, exit 2

**Word list fixture helper:**
```go
func writeTempWordList(t *testing.T, content string) string {
    t.Helper()
    f, err := os.CreateTemp(t.TempDir(), "wordlist*.txt")
    if err != nil { t.Fatalf("Setup: %v", err) }
    if _, err := f.WriteString(content); err != nil { t.Fatalf("Setup: %v", err) }
    f.Close()
    return f.Name()
}
```

Use known Hebrew words/values for deterministic testing. Single-letter words are easiest: "א" (aleph) has value 1 in Hechrachi, so a word list with one entry "א" and --find 1 --mispar hechrachi should return it.

## Acceptance Criteria

- [ ] --find without word list: exit 2, stderr non-empty, stdout empty
- [ ] --find with nonexistent word list file: exit 3, stderr mentions path
- [ ] --find with valid word list: exit 0, stdout non-empty
- [ ] GEMATRIA_WORDLIST env var used when --wordlist not provided
- [ ] --wordlist flag overrides GEMATRIA_WORDLIST
- [ ] --limit limits result count; hasMore indicator present in line/card output when applicable
- [ ] GEMATRIA_LIMIT env var used when --limit not provided (only when --find active)
- [ ] GEMATRIA_LIMIT=invalid with --find → exit 2
- [ ] GEMATRIA_LIMIT=invalid without --find → exit 0 (lazy validation)
- [ ] --output json: stdout is valid JSON with results array and hasMore field
- [ ] --output value: stdout is one Hebrew word per line, no other content
- [ ] --output json with error: stderr is JSON error object, stdout empty
- [ ] All tests use t.TempDir() for temp files (auto-cleaned)
- [ ] Tests use injected getenv closure (not os.Getenv)
- [ ] make test and make test-race pass


## Notes

**2026-04-15T12:45:34Z**

Added run_find_test.go with 11 integration tests covering: --wordlist flag overrides GEMATRIA_WORDLIST env, --limit restricts results with hasMore indicator in line output, GEMATRIA_LIMIT env var applies when --find is active, GEMATRIA_LIMIT=invalid with --find exits 2, GEMATRIA_LIMIT=invalid without --find exits 0 (lazy validation), --output value yields bare Hebrew per line, --output line yields RTL-wrapped Hebrew with transliteration, --output card yields numbered entries, --output json error yields JSON on stderr, hasMore:true in JSON output, empty GEMATRIA_WORDLIST treated as absent. writeTempWordList helper defined in the file. All tests use t.TempDir() and injected getenv closure. make test and make test-race both pass.
