---
id: gc-fp0o
status: closed
deps: [gc-j0zh, gc-j1c2, gc-pny0, gc-sv06]
links: []
created: 2026-04-16T03:46:03Z
type: task
priority: 2
assignee: Andre Silva
parent: gc-ed0f
tags: [index-location, tests, task]
---
# Tests for index path resolution, auto-discovery, and precedence

Comprehensive test coverage for all new path resolution and auto-discovery logic.

## Spec References
- docs/specs/gematria-index.md §"Implementation Plan" Step 7

## Implementation Details

### Unit Tests: src/internal/cli/index_path_test.go (new file)

resolveIndexPath tests:
- No env vars set → returns ~/.local/share/gematria/gematria.db (for sqlite) or .idx (for index)
- GEMATRIA_INDEX_LOCATION set → uses that directory instead of XDG default
- GEMATRIA_INDEX_NAME set → uses that name instead of "gematria"
- Both GEMATRIA_INDEX_LOCATION and GEMATRIA_INDEX_NAME set → combines them
- XDG_DATA_HOME set (no GEMATRIA_INDEX_LOCATION) → uses XDG_DATA_HOME/gematria/
- GEMATRIA_INDEX_NAME with "/" → returns error
- GEMATRIA_INDEX_NAME with "\" → returns error
- GEMATRIA_INDEX_NAME with dots (e.g., "my.index") → succeeds, produces my.index.db

discoverDefaultIndex tests:
- .db exists at default location → returns .db path, found=true
- .idx exists (no .db) at default location → returns .idx path, found=true
- Both .db and .idx exist → returns .db path (SQLite preferred)
- Neither exists → returns "", found=false
- GEMATRIA_INDEX_LOCATION overrides discovery directory
- GEMATRIA_INDEX_NAME overrides discovery filename

### CLI Integration Tests: src/internal/cli/run_index_test.go (extend)

- runIndex with no --output writes to XDG default location, directory auto-created
- runIndex with --output bypasses env vars entirely
- runIndex with GEMATRIA_INDEX_LOCATION writes to specified directory
- runIndex with invalid GEMATRIA_INDEX_NAME returns exit code 2

### CLI Integration Tests: src/internal/cli/run_find_test.go (extend)

- --find auto-discovers .db at default location → success
- --find auto-discovers .idx when no .db exists → success
- --find prefers .db over .idx when both exist
- --find errors when no wordlist and no default index exists
- --find with GEMATRIA_WORDLIST takes precedence over default index
- --find with --wordlist flag takes precedence over everything

### Test Pattern

All tests use the getenv function pattern (envWith helper) to inject env vars without modifying the real environment. Tests that need files on disk use t.TempDir() and set GEMATRIA_INDEX_LOCATION (or XDG_DATA_HOME) to the temp dir, keeping tests isolated and hermetic.

For home directory resolution in unit tests, consider injecting a "home" override or mocking os.UserHomeDir via an internal function variable. Alternatively, always set XDG_DATA_HOME or GEMATRIA_INDEX_LOCATION in tests so the ~ fallback path is never exercised in CI (test it once explicitly).


## Notes

**2026-04-16T04:07:26Z**

All specified tests were already present and passing when this ticket was picked up (unit tests for resolveIndexPath and discoverDefaultIndex in index_path_test.go; CLI integration tests in run_index_test.go and run_find_test.go). Added one missing test: TestRun_index_outputFlagBypassesEnvVars — verifies that passing --output writes to the explicit path and ignores GEMATRIA_INDEX_LOCATION entirely.
