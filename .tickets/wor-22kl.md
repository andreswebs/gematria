---
id: wor-22kl
status: closed
deps: [wor-q01l, wor-g979, wor-93fx]
links: []
created: 2026-04-15T03:54:21Z
type: task
priority: 1
parent: wor-24my
tags: [reverse-lookup, cli-find-flags, task]
---
# Add --find, --wordlist, --limit flags and env vars to CLI Config

Extend src/internal/cli/config.go (or create it) with three new flags for reverse lookup: --find, --wordlist, --limit. Also wire GEMATRIA_WORDLIST and GEMATRIA_LIMIT environment variables with lazy validation semantics.

## Spec References
- docs/specs/requirements.md §6 (--find, --wordlist, --limit), §9 (GEMATRIA_WORDLIST, GEMATRIA_LIMIT)
- docs/specs/cli-design.md §3.4 (flag > env var > default precedence), §6.6 (lazy env var validation)
- docs/specs/code-architecture.md §4.2 (Config struct, lazy env var validation)

## Implementation Details

**New fields on Config struct:**
```go
type Config struct {
    // ... existing fields (System, OutputFormat, Atbash, NoColor, FailEarly) ...
    FindValue int    // numeric value to look up (only valid when FindSet=true)
    FindSet   bool   // true when --find was explicitly provided
    Wordlist  string // resolved word list path (empty means none provided yet)
    Limit     int    // maximum results (already normalized: always >= 1 after resolution)
}
```

**Flag registration (pflag):**
- `--find <int>`: No short form. Description: "find words whose gematria value equals N". Sets FindSet=true when parsed.
- `--wordlist <path>`: No short form. Description: "path to word list file (or set GEMATRIA_WORDLIST)".
- `--limit` / `-l` `<int>`: Default 0 (meaning "use default"). Description: "maximum results to return (default 20)".

**Config resolution logic:**
After flag parsing, in a resolveConfig() helper (called only when building Config for Run()):

For Wordlist:
```
if --wordlist was not provided: cfg.Wordlist = getenv("GEMATRIA_WORDLIST")
// Do NOT validate the path here — lazy validation in run.go when --find is active
```

For Limit:
```
if --limit was not provided (== 0) and GEMATRIA_LIMIT is set:
    parse GEMATRIA_LIMIT as int
    if parse fails or value < 1: return error "GEMATRIA_LIMIT must be a positive integer, got: <value>"
    cfg.Limit = parsed
else if cfg.Limit < 1:
    cfg.Limit = gematria.DefaultLookupLimit
```
Only perform GEMATRIA_LIMIT validation when FindSet is true (lazy: a stale GEMATRIA_LIMIT does not block non-lookup invocations).

**Key design constraints from specs:**
- GEMATRIA_WORDLIST is validated lazily: only checked/opened when --find is active (cli-design.md §6.6).
- GEMATRIA_LIMIT is validated lazily: only parsed when --find is active.
- Flag always beats env var always beats built-in default (20).
- --limit value of 0 or less is replaced by DefaultLookupLimit (20) after resolution.

**Injected getenv:** Config resolution must use the injected getenv function (not os.Getenv directly), to support test isolation.

## Acceptance Criteria

- [ ] Config struct has FindValue int, FindSet bool, Wordlist string, Limit int fields
- [ ] --find flag registered (no short form); sets FindSet=true when used
- [ ] --wordlist flag registered (no short form)
- [ ] --limit / -l flag registered, default behavior leads to 20 after resolution
- [ ] GEMATRIA_WORDLIST used when --wordlist not provided (no validation at parse time)
- [ ] GEMATRIA_LIMIT used when --limit not provided and FindSet is true; invalid value returns exit-code-2 error
- [ ] Flag takes precedence over env var for both wordlist and limit
- [ ] Limit <= 0 after resolution is replaced by DefaultLookupLimit (20)
- [ ] Config resolution uses injected getenv, not os.Getenv
- [ ] make vet and make lint pass


## Notes

**2026-04-15T12:17:56Z**

Added FindValue int, FindSet bool, Wordlist string, Limit int fields to Config struct. Registered --find (no short form), --wordlist (no short form), --limit/-l flags via pflag. Lazy env var resolution: GEMATRIA_WORDLIST stored without validation; GEMATRIA_LIMIT parsed and validated only when FindSet=true (invalid value returns error). Flag always beats env var; Limit<=0 after resolution is replaced by DefaultLookupLimit (20). All uses injected getenv rather than os.Getenv. 13 new tests added to config_test.go; make build passes clean.
