---
id: gc-j1c2
status: closed
deps: [gc-j0zh]
links: []
created: 2026-04-16T03:45:09Z
type: task
priority: 2
assignee: Andre Silva
parent: gc-ed0f
tags: [index-location, run-index, task]
---
# Wire resolveIndexPath into runIndex with directory auto-creation

Update runIndex in src/internal/cli/run.go to use resolveIndexPath when --output is not set, and auto-create the resolved directory before writing.

## Spec References
- docs/specs/gematria-index.md §"Implementation Plan" Step 2

## Implementation Details

### Changes to runIndex (src/internal/cli/run.go)

Replace the current default output path logic:

  // BEFORE (lines ~119-124):
  if outputPath == "" {
    if format == "sqlite" { outputPath = wordlistPath + ".db" }
    else { outputPath = wordlistPath + ".idx" }
  }

  // AFTER:
  if outputPath == "" {
    outputPath, err = resolveIndexPath(format, getenv)
    if err != nil {
      fmt.Fprintf(stderr, "Error: %s\n", err.Error())
      return 2
    }
  }

### Directory Auto-Creation

Before writing the index (before the switch on format), create the parent directory:

  if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
    fmt.Fprintf(stderr, "Error: cannot create directory %q: %v\n", filepath.Dir(outputPath), err)
    return 3
  }

This applies to both env-var-resolved paths and --output paths. Creating the parent directory is safe and idempotent — MkdirAll is a no-op if the directory already exists.

### runIndex Signature Change

runIndex currently receives (args, stdout, stderr). It needs getenv to call resolveIndexPath. Update the signature to include getenv:

  func runIndex(args []string, stdout *os.File, stderr *os.File, getenv func(string) string) int

Update the dispatch call in Run() to pass getenv through.

### Exit Code for Invalid Env Var

Invalid GEMATRIA_INDEX_NAME (path separator) returns exit code 2 (CLI misuse), consistent with other invalid env var values.


## Notes

**2026-04-16T03:54:20Z**

runIndex signature changed to include getenv (passed from Run()). Default output path now uses resolveIndexPath (env var + XDG) instead of wordlistPath+ext. os.MkdirAll auto-creates parent dir before writing (applies to both --output and default paths). Existing tests TestRun_index_defaultOutputPath_* updated to pass GEMATRIA_INDEX_LOCATION via envWith so they control the output dir. TestRun_index_badOutputPath_exit3 updated: MkdirAll fails first so error mentions the directory, not the full file path. indexHelpText updated with GEMATRIA_INDEX_LOCATION, GEMATRIA_INDEX_NAME, XDG_DATA_HOME entries.
