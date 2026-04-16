---
id: gc-h4fl
status: closed
deps: [gc-j1c2, gc-sv06]
links: []
created: 2026-04-16T03:45:47Z
type: task
priority: 2
assignee: Andre Silva
parent: gc-ed0f
tags: [index-location, documentation, task]
---
# Update help text, CLAUDE.md, and env var documentation

Update all user-facing documentation to reflect the new env vars and auto-discovery behavior.

## Spec References
- docs/specs/gematria-index.md §"Implementation Plan" Step 6
- docs/specs/gematria-index.md §"Environment Variables (updated)"

## Implementation Details

### indexHelpText in src/internal/cli/run.go

Update the help string for gematria index to document the new default behavior:

  Options:
      --wordlist string   Path to input word list (required)
      --output string     Output file path (default: ~/.local/share/gematria/gematria.db)
      --format string     Output format: sqlite|index (default: sqlite)
    -h, --help            Show this help message

  Environment Variables:
      GEMATRIA_INDEX_LOCATION   Directory for index files (overrides XDG_DATA_HOME)
      GEMATRIA_INDEX_NAME       Index filename without extension (default: gematria)
      XDG_DATA_HOME             XDG base directory (default: ~/.local/share)

### Main --help env var section

Add GEMATRIA_INDEX_LOCATION and GEMATRIA_INDEX_NAME to the main help text env var table (helpText constant or wherever it's generated).

### CLAUDE.md

Update the Environment Variables table to include:
- GEMATRIA_INDEX_LOCATION — Directory for index files; overrides XDG_DATA_HOME
- GEMATRIA_INDEX_NAME — Index filename without extension; default: gematria
- XDG_DATA_HOME — XDG base; default: ~/.local/share

Update the --find section to mention that --wordlist is optional when a default index exists.

Update the Quick Start section if appropriate to show the simpler workflow:
  gematria index --wordlist words.txt    # index once
  gematria --find 441 --output json      # query without --wordlist

### Error message for missing --find wordlist

Already covered in task gc-sv06, but ensure the message mentions 'gematria index' as the way to create a default index.


## Notes

**2026-04-16T04:06:52Z**

Added GEMATRIA_INDEX_LOCATION, GEMATRIA_INDEX_NAME, XDG_DATA_HOME to main helpText env vars section in run.go. Added index workflow examples to helpText (gematria index --wordlist, gematria --find without --wordlist). Updated CLAUDE.md: Quick Start shows build-once-then-query pattern; Recommended Flags table splits --find into default-index and explicit-list rows; Environment Variables table adds the three new vars with notes; lazy validation note extended to cover index vars.
