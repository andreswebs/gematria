---
id: gc-pny0
status: closed
deps: [gc-j0zh]
links: []
created: 2026-04-16T03:45:20Z
type: task
priority: 2
assignee: Andre Silva
parent: gc-ed0f
tags: [index-location, auto-discovery, task]
---
# Implement discoverDefaultIndex for --find auto-discovery

Add a discoverDefaultIndex function to src/internal/cli/index_path.go that locates an existing index file at the default XDG location for use by --find.

## Spec References
- docs/specs/gematria-index.md §"Implementation Plan" Step 3
- docs/specs/gematria-index.md §"Auto-discovery for --find"

## Implementation Details

### Function Signature

  func discoverDefaultIndex(getenv func(string) string) (path string, found bool)

### Logic

1. Resolve location directory (same logic as resolveIndexPath, without name/extension):
   - GEMATRIA_INDEX_LOCATION > XDG_DATA_HOME/gematria > ~/.local/share/gematria
2. Resolve name: GEMATRIA_INDEX_NAME or "gematria".
3. Check for <location>/<name>.db — if os.Stat succeeds, return (path, true).
4. Check for <location>/<name>.idx — if os.Stat succeeds, return (path, true).
5. Return ("", false).

### Key Constraints

- SQLite (.db) is preferred over index file (.idx) when both exist. This matches the default --format for gematria index (sqlite).
- Does NOT create directories. This function only reads existing state.
- Does NOT validate file contents — backend auto-detection (by extension) handles that downstream.
- Does NOT error on invalid GEMATRIA_INDEX_NAME — if the name is invalid, the files simply won't exist, and found=false is the correct result.
- Consider extracting the shared "resolve location directory" logic into a private helper (resolveIndexDir) used by both resolveIndexPath and discoverDefaultIndex to avoid duplication.


## Notes

**2026-04-16T03:57:36Z**

Implemented discoverDefaultIndex in index_path.go. Extracted resolveIndexDir private helper to avoid duplicating location-resolution logic between resolveIndexPath and discoverDefaultIndex. The function checks .db first (preferred), then .idx; returns ('', false) when neither exists. Invalid GEMATRIA_INDEX_NAME silently returns not-found (no error) as spec requires. Added 7 tests covering all branches.
