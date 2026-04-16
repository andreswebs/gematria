---
id: gc-auzm
status: closed
deps: [gc-f4u5]
links: []
created: 2026-04-16T04:30:00Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-sjh7
tags: [index-migration, run-dispatch, task]
---
# Wire index mode in Run() and update --help

Remove the old subcommand dispatch. Wire --index mode into the main Run() flow. Update --help text.

## Remove old subcommand dispatch

Delete from Run():
```go
if len(args) > 0 && args[0] == "index" {
    return runIndex(args[1:], stdout, stderr, getenv)
}
```

Delete the old `runIndex` function (the one with its own pflag.NewFlagSet).
Delete the `indexHelpText` constant.

## Add index branch in Run()

After config parsing and before --find/compute dispatch:

```go
if cfg.Index {
    return runIndex(cfg, stdout, stderr, getenv)
}
```

Write a new `runIndex(cfg Config, stdout, stderr *os.File, getenv func(string) string) int` that:
1. Resolves output path: if `cfg.IndexOutput != ""`, use it; otherwise call `resolveIndexPath(cfg.IndexFormat, getenv)`.
2. Auto-create parent directory with `os.MkdirAll`.
3. Open and parse word list from `cfg.Wordlist`.
4. Write index in the requested format (same logic as old function).
5. Print "Indexed N words → path" to stdout; return 0.

The `resolveIndexPath` and `discoverDefaultIndex` functions in `index_path.go` are unchanged.

## Update helpText

Add an "Indexing:" section in the Options area:

```
Indexing:
      --index                   Build a pre-computed index from --wordlist
      --index-output string     Output file path (bypasses env var resolution)
      --index-format string     Index format: sqlite|index (default: sqlite)
```

Add indexing env vars (GEMATRIA_INDEX_LOCATION, GEMATRIA_INDEX_NAME, XDG_DATA_HOME) to the Environment Variables section.

Update examples to show `gematria --index --wordlist words.txt` instead of `gematria index --wordlist words.txt`.

## Spec References

- docs/specs/gematria-index.md (Implementation Plan Steps 2, --help Grouping)

## Acceptance Criteria

- [ ] Old subcommand dispatch (`args[0] == "index"`) removed from Run()
- [ ] Old runIndex function and indexHelpText deleted
- [ ] New runIndex(cfg, ...) uses cfg.Wordlist, cfg.IndexOutput, cfg.IndexFormat
- [ ] resolveIndexPath called when cfg.IndexOutput is empty
- [ ] os.MkdirAll on parent directory before writing
- [ ] "Indexed N words → path" printed to stdout on success
- [ ] --help has "Indexing:" section with --index, --index-output, --index-format
- [ ] --help examples updated from subcommand to flag syntax
- [ ] --help env vars include GEMATRIA_INDEX_LOCATION, GEMATRIA_INDEX_NAME, XDG_DATA_HOME
- [ ] make build passes


## Notes

**2026-04-16T04:41:36Z**

Removed old 'index' subcommand dispatch and runIndex(args []string, ...) from Run(). Deleted indexHelpText constant. Added new runIndex(cfg Config, ...) that uses cfg.Wordlist, cfg.IndexOutput, cfg.IndexFormat from the already-parsed Config. Added cfg.Index branch in Run() between Version check and formatter init. Updated helpText with Indexing: section (--index, --index-output, --index-format) and updated example from 'gematria index --wordlist' to 'gematria --index --wordlist'. Updated runFind error message and all run_index_test.go tests to use new --index flag syntax. Also updated run_find_test.go and run_test.go where they checked for 'gematria index' in the error message. make build passes cleanly.
