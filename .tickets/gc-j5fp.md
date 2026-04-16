---
id: gc-j5fp
status: closed
deps: [gc-auzm]
links: []
created: 2026-04-16T04:30:00Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-sjh7
tags: [index-migration, tests, task]
---
# Migrate and add tests for --index flag

Rewrite existing run_index_test.go tests for the new --index flag syntax. Add tests for flag conflicts, positional arg rejection, and GEMATRIA_WORDLIST resolution.

## Existing tests to migrate

All tests in run_index_test.go that use `Run([]string{"index", ...})` must be rewritten to use `Run([]string{"--index", ...})`. The test logic (verify exit code, stdout content, stderr content) stays the same; only the args change.

## New tests to add

**Conflict rejection tests** (all expect exit 2):
- `--index --find 376 --wordlist w.txt` → mutually exclusive
- `--index -t --wordlist w.txt` → mutually exclusive
- `--index-output foo.db` (without `--index`) → requires --index
- `--index-format index` (without `--index`) → requires --index
- `--index --wordlist w.txt shalom` (positional arg) → does not accept positional args

**Wordlist resolution tests**:
- `--index` with `GEMATRIA_WORDLIST` env var (no --wordlist flag) → succeeds
- `--index` without --wordlist and no GEMATRIA_WORDLIST → exit 2 with message

**Regression test**:
- `Run([]string{"index", "--wordlist", path})` (old subcommand syntax) → exit 1 (treated as unknown Latin input "index"), NOT exit 0. Confirms no migration shim.

**--help test**:
- Verify --help output contains "--index", "--index-output", "--index-format", "Indexing:"

## Test patterns

Follow existing run_test.go patterns: Run() with injected I/O, pipeCapture, makeStdinPipe, envWith.

## Spec References

- docs/specs/gematria-index.md (Flag Conflicts table, Implementation Plan Step 3)

## Acceptance Criteria

- [ ] All existing run_index_test.go tests pass with new --index flag syntax
- [ ] 5 conflict-rejection tests (all 5 combinations from the spec)
- [ ] GEMATRIA_WORDLIST resolution test for --index
- [ ] Missing-wordlist test for --index
- [ ] Old subcommand syntax regression test (exit 1, not 0)
- [ ] --help contains "Indexing:" heading and all three index flags
- [ ] At least 12 total test cases
- [ ] make test passes; make lint clean


## Notes

**2026-04-16T04:44:25Z**

All existing run_index_test.go tests were already using the new --index flag syntax (migration from gc-auzm was complete). Added 7 new test cases: 5 conflict-rejection tests (--index+--find, --index+-t, --index-output without --index, --index-format without --index, --index with positional args), GEMATRIA_WORDLIST env var resolution test, and old subcommand syntax regression test (confirms 'index' as literal arg exits 1 not 0). Enhanced TestRun_index_help_exit0 to check for 'Indexing:' heading and all three index flags (--index, --index-output, --index-format). Final count: 23 test cases. make build passes with 0 lint issues.
