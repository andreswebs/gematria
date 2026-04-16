---
id: gc-ed0f
status: closed
deps: []
links: []
created: 2026-04-16T03:44:45Z
type: epic
priority: 2
assignee: Andre Silva
tags: [index-location, epic]
---
# Default Index Location & Auto-Discovery

Add well-known default index location (XDG_DATA_HOME), two new env vars (GEMATRIA_INDEX_LOCATION, GEMATRIA_INDEX_NAME), and auto-discovery of the default index for --find lookups. This eliminates the need to pass --wordlist for reverse lookups once an index has been built, and gives users a stable, configurable location for index files.

## Spec References
- docs/specs/gematria-index.md (primary — full design, precedence rules, implementation plan)
- docs/specs/wordlist-backends.md (backend interface, auto-detection logic)
- docs/specs/cli-design.md §7 (reverse lookup flags)

## Design Summary

### Output path precedence for gematria index

  --output flag (literal full path, no auto-extension)
    > GEMATRIA_INDEX_LOCATION / GEMATRIA_INDEX_NAME.<ext>
      > GEMATRIA_INDEX_LOCATION / gematria.<ext>
        > XDG_DATA_HOME/gematria/gematria.<ext>
          > ~/.local/share/gematria/gematria.<ext>

### Wordlist resolution for gematria --find

  --wordlist flag
    > GEMATRIA_WORDLIST env
      > default XDG index (.db first, then .idx)
        > error

### Key Decisions
- XDG_DATA_HOME on all platforms (including macOS); GEMATRIA_INDEX_LOCATION overrides
- Auto-create directories with os.MkdirAll(path, 0o755)
- Lazy validation of env vars (only when gematria index or --find runs)
- --output is a complete bypass — no partial overrides, no auto-extension
- GEMATRIA_INDEX_NAME rejects path separators, allows dots, defaults to "gematria"
- Additive merge into the shared default index is intentional
- Auto-discovery for --find when no wordlist specified — SQLite (.db) preferred over .idx
- This is a breaking change: default output of gematria index moves from <wordlist>.<ext> to XDG location

## Tasks

1. index-path-resolver — resolveIndexPath helper function for env var + XDG resolution
2. update-run-index — Wire resolveIndexPath into runIndex; auto-create directories
3. discover-default-index — discoverDefaultIndex helper for --find auto-discovery
4. update-find-resolution — Wire auto-discovery into --find wordlist resolution
5. validate-index-name — GEMATRIA_INDEX_NAME path separator validation
6. update-docs — Update help text, CLAUDE.md env var table, and indexHelpText
7. tests — Full test coverage for path resolution, auto-discovery, and precedence


## Notes

**2026-04-16T04:10:19Z**

All child tickets completed. Epic verified as fully implemented: resolveIndexPath (XDG_DATA_HOME + GEMATRIA_INDEX_LOCATION + GEMATRIA_INDEX_NAME with path separator validation) wired into runIndex with auto-MkdirAll; discoverDefaultIndex (prefers .db over .idx) wired into runFind auto-discovery with GEMATRIA_WORDLIST precedence; help text updated; full unit+integration test coverage passing.
