---
id: gc-f4u5
status: closed
deps: []
links: []
created: 2026-04-16T04:30:00Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-sjh7
tags: [index-migration, config, task]
---
# Migrate index flags to Config

Add --index, --index-output, --index-format to the unified Config struct. Add flag conflict validation. Remove the old subcommand pflag set.

## Config struct additions

```go
Index       bool   // triggers index-building mode
IndexOutput string // explicit output path (bypasses env var resolution)
IndexFormat string // "sqlite" or "index"
```

## Flag definitions (in parseConfig)

```go
fs.BoolVar(&index, "index", false, "build a pre-computed index from --wordlist")
fs.StringVar(&indexOutput, "index-output", "", "output file path for index")
fs.StringVar(&indexFormat, "index-format", "sqlite", "index format: sqlite|index")
```

## Validation (after parsing, in parseConfig)

1. `--index-format` when explicitly set: must be "sqlite" or "index" (exit 2 with valid-list).
2. `--index` + `--find`: reject with "mutually exclusive" error.
3. `--index` + `-t` (transliterate): reject with "mutually exclusive" error.
4. `--index-output` without `--index`: reject with "--index-output requires --index".
5. `--index-format` explicitly set without `--index`: reject with "--index-format requires --index".
6. `--index` + positional args (`fs.Args()` non-empty): reject with "--index does not accept positional arguments".
7. `--index` without wordlist (no `--wordlist` flag AND no `GEMATRIA_WORDLIST`): reject with "--index requires --wordlist or GEMATRIA_WORDLIST".

## What to remove

The old `runIndex` function in run.go has its own `pflag.NewFlagSet("gematria index", ...)`. That flag set and the `indexHelpText` constant are removed in the Run() wiring task (T2), but the config side should be ready first.

## Spec References

- docs/specs/gematria-index.md (Flag Conflicts table, Implementation Plan Step 1)

## Acceptance Criteria

- [ ] Config has Index (bool), IndexOutput (string), IndexFormat (string) fields
- [ ] --index parses as bool
- [ ] --index-output parses as string
- [ ] --index-format parses as string with "sqlite"/"index" validation
- [ ] Conflict: --index + --find → exit 2 with clear message
- [ ] Conflict: --index + -t → exit 2
- [ ] Conflict: --index-output without --index → exit 2
- [ ] Conflict: --index-format without --index → exit 2
- [ ] Conflict: --index + positional args → exit 2
- [ ] --index without wordlist → exit 2 with message mentioning --wordlist and GEMATRIA_WORDLIST
- [ ] GEMATRIA_WORDLIST resolves for --index (same as --find)
- [ ] Existing config_test.go tests pass (no regressions)
- [ ] At least 8 new test cases in config_test.go covering all conflicts and valid combinations
- [ ] make build passes


## Notes

**2026-04-16T04:34:45Z**

Added Index (bool), IndexOutput (string), IndexFormat (string) to Config struct. Added --index, --index-output, --index-format flags in parseConfig. IndexFormat defaults to 'sqlite' and is validated eagerly when explicitly set. Added 7 conflict checks in order: invalid index-format, index+find, index+transliterate, index-output without index, index-format without index, index with positional args, index without wordlist. GEMATRIA_WORDLIST env var resolves correctly for --index via the existing wordlist resolution. Added 12 new tests in config_test.go. run.go/run_index_test.go untouched (that's gc-auzm's scope).
