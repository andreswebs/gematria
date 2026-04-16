---
id: gc-j0zh
status: closed
deps: []
links: []
created: 2026-04-16T03:44:57Z
type: task
priority: 2
assignee: Andre Silva
parent: gc-ed0f
tags: [index-location, index-path-resolver, task]
---
# Implement resolveIndexPath helper for env var and XDG resolution

Create src/internal/cli/index_path.go with a resolveIndexPath function that resolves the default output path for gematria index from env vars and XDG defaults.

## Spec References
- docs/specs/gematria-index.md §"Implementation Plan" Step 1

## Implementation Details

### New File: src/internal/cli/index_path.go

  func resolveIndexPath(format string, getenv func(string) string) (string, error)

Logic:
1. Read GEMATRIA_INDEX_LOCATION from getenv. If empty:
   a. Read XDG_DATA_HOME from getenv. If empty, default to ~/.local/share.
   b. Append /gematria to get the location directory.
2. Read GEMATRIA_INDEX_NAME from getenv. If empty, default to "gematria".
3. Validate name: reject if it contains "/" or "\". Return clear error:
   "invalid GEMATRIA_INDEX_NAME %q: must not contain path separators"
4. Determine extension from format arg: ".db" for "sqlite", ".idx" for "index".
5. Return filepath.Join(location, name+ext).

### Key Constraints
- Uses the getenv function parameter (not os.Getenv directly) for testability, consistent with the rest of the CLI layer.
- os.UserHomeDir() for ~ resolution (cross-platform).
- Does NOT create the directory — that's the caller's responsibility (Step 2 task).
- Does NOT validate that the directory exists — lazy validation means we only check at write time.


## Notes

**2026-04-16T03:49:56Z**

Implemented resolveIndexPath in src/internal/cli/index_path.go. Uses getenv func for testability (consistent with rest of CLI). userHomeDir is a package-level var wrapping os.UserHomeDir to allow overriding in future tests if needed. Validates GEMATRIA_INDEX_NAME rejects path separators ('/' and '\'). Extension is '.db' for sqlite (default) and '.idx' for index format. Does NOT create the directory — that's the caller's responsibility. Full test coverage in index_path_test.go: default XDG path, GEMATRIA_INDEX_LOCATION override, XDG_DATA_HOME override, custom name, dots-in-name allowed, path separator rejection, precedence.
