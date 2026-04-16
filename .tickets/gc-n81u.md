---
id: gc-n81u
status: closed
deps: [gc-auzm]
links: []
created: 2026-04-16T04:30:00Z
type: task
priority: 2
assignee: Andre Silva
parent: gc-sjh7
tags: [index-migration, docs, task]
---
# Update docs for --index flag migration

Update all documentation to replace the old subcommand syntax with the new flag syntax.

## Files to update

### AGENTS.md
- Quick Start: replace \`gematria index --wordlist\` with \`gematria --index --wordlist\`
- Recommended Flags table: update index-related rows
- Environment Variables table: ensure GEMATRIA_INDEX_LOCATION, GEMATRIA_INDEX_NAME, XDG_DATA_HOME are present

### README.md
- Quick Start: update index example if present
- Reverse Lookup section: update indexing mention
- Any references to \`gematria index\` subcommand

### CLAUDE.md
- Key Design Decisions: mention --index flag instead of subcommand
- Env var table: verify GEMATRIA_INDEX_* entries present

## Spec References

- docs/specs/gematria-index.md (already updated)

## Acceptance Criteria

- [ ] AGENTS.md: all "gematria index" → "gematria --index" in examples and text
- [ ] AGENTS.md: --index, --index-output, --index-format in Recommended Flags
- [ ] README.md: updated examples (no "gematria index" subcommand syntax)
- [ ] CLAUDE.md: updated to reflect --index flag, not subcommand
- [ ] No stale references to "gematria index" subcommand in any doc


## Notes

**2026-04-16T04:47:34Z**

Updated AGENTS.md, CLAUDE.md, and README.md. AGENTS.md and CLAUDE.md: replaced 'gematria index --wordlist' with 'gematria --index --wordlist' in Quick Start; added --index, --index-output, --index-format rows to Recommended Flags table; updated GEMATRIA_WORDLIST notes to mention --index; fixed env var lazy validation note to say '--index is active' instead of 'gematria index runs'. README.md: added Indexing subsection under Reverse Lookup with examples and flag table; added GEMATRIA_INDEX_LOCATION, GEMATRIA_INDEX_NAME, XDG_DATA_HOME to Environment Variables table; updated GEMATRIA_WORDLIST notes to mention --index. CLAUDE.md was already up-to-date from prior implementation work (it mirrors AGENTS.md).
