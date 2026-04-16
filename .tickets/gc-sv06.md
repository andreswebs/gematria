---
id: gc-sv06
status: closed
deps: [gc-pny0]
links: []
created: 2026-04-16T03:45:34Z
type: task
priority: 2
assignee: Andre Silva
parent: gc-ed0f
tags: [index-location, find-resolution, task]
---
# Wire auto-discovery into --find wordlist resolution

Update the --find wordlist resolution in the CLI so that when no --wordlist flag and no GEMATRIA_WORDLIST env var are set, the tool auto-discovers the default index at the XDG location.

## Spec References
- docs/specs/gematria-index.md §"Implementation Plan" Step 4
- docs/specs/gematria-index.md §"Wordlist resolution for --find"

## Implementation Details

### Precedence Chain

  --wordlist flag
    > GEMATRIA_WORDLIST env
      > discoverDefaultIndex (try .db then .idx)
        > error: "no word list specified and no default index found"

### Changes to src/internal/cli/config.go or run.go

Currently, when --find is active and no wordlist is resolved, the CLI returns:
  "Error: --wordlist is required when using --find"

Replace this with:

  1. If cfg.wordlistPath is empty after flag + env resolution:
     a. Call discoverDefaultIndex(getenv).
     b. If found, set cfg.wordlistPath = path.
     c. If not found, return error:
        "no word list specified and no default index found; run 'gematria index --wordlist <path>' to create one, or pass --wordlist explicitly"
  2. Proceed with normal backend auto-detection (openWordSource).

The error message should guide the user toward the solution — not just say what's wrong.

### Where the Check Lives

The cleanest place is in the --find validation block in Run() (or parseConfig), after resolving wordlistPath from flag + GEMATRIA_WORDLIST. This keeps the auto-discovery attempt close to the existing error path it replaces.

### getenv Propagation

discoverDefaultIndex needs getenv. If the check is in parseConfig, getenv is already available. If in Run(), it's passed through. Either way, no new plumbing needed beyond what task gc-j1c2 already adds.


## Notes

**2026-04-16T04:02:42Z**

Wired discoverDefaultIndex into runFind(). When cfg.Wordlist is empty after flag+env resolution, attempts auto-discovery. If found, sets cfg.Wordlist to the discovered path. If not found, returns exit 2 with a new error message that mentions 'gematria index' to guide users. Updated existing no-wordlist tests (TestRun_find_noWordlist_exit2, TestRun_find_emptyWordlistEnv_exit2, TestRun_find_outputJSON_errorIsJSONOnStderr) to set GEMATRIA_INDEX_LOCATION to an empty temp dir so they're resilient to real indexes on disk. Added 5 new tests covering: .db auto-discovery, .idx fallback, .db-over-.idx preference, new error message, and GEMATRIA_WORDLIST precedence.
