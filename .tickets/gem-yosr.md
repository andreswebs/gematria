---
id: gem-yosr
status: open
deps: [gem-kj4m]
links: [gem-kj4m]
created: 2026-07-22T17:31:30Z
type: task
priority: 2
assignee: Andre Silva
---
# Migrate exit codes to the fleet taxonomy (ADR 0001)

## Context

`docs/adr/0001-exit-code-taxonomy.md` (already in this repo) adopts a
family-wide exit-code taxonomy: 0 success; 1 recoverable result; 2-63
optional result sub-codes; 64 EX_USAGE; 65 EX_DATAERR; 70 EX_SOFTWARE;
74 EX_IOERR; 78 EX_CONFIG; 130/143 signals. This ticket migrates the
CLI from the current 0/1/2/3/4 scheme to that taxonomy. This is a
deliberate breaking change to the CLI contract (the project is v0.x);
include release notes.

The tool does not catch signals, so per the ADR no signal handling is
needed (default signal death already yields 130/143).

## Old to new mapping (decided; implement exactly this)

| Old | Condition | New |
| --- | --------- | --- |
| 0 | success | 0 (unchanged) |
| 1 | input data errors: `InvalidCharError`, `UnknownNameError`, `UnknownWordError` | 65 |
| 2 | flag misuse: bad flag value, mutually exclusive flags, `--find` without wordlist, other `parseConfig` flag errors | 64 |
| 2 | invalid environment variable value (`GEMATRIA_MISPAR`, `GEMATRIA_SCHEME`, `GEMATRIA_LIMIT`, `GEMATRIA_OUTPUT` when the value came from the env) | 78 |
| 3 | filesystem/backend I/O: cannot open/create/write wordlist or index, backend query failures | 74 |
| 3 | malformed wordlist/index content (parse failures of file content) | 65 |
| 4 | partial batch success (some stdin lines ok, some failed) | 1 |
| 1 | all stdin lines failed (batch completed, every line rejected) | 65 |

No 2-63 result sub-codes are claimed for this tool. There is currently
no internal-error (70) path; do not add one artificially, but the
registry (below) documents only codes the tool can actually produce.

## Production sites to change (verified 2026-07-22)

All in `src/internal/cli/`:

- `run.go:157` (`parseConfig` error): currently flat 2. Split by
  provenance: flag-sourced errors return 64, env-sourced return 78.
  `parseConfig` in `config.go` knows the source at each site:
  - flag errors: `config.go:110` (invalid `--output`; env-sourced when
    the flag was empty and the value came from `GEMATRIA_OUTPUT`,
    provenance is visible at `config.go:101-104`), `config.go:119`
    (invalid `--mispar` flag, eager), `config.go:136`
    (`--wordlist-format`), `config.go:141` (`--index-format`),
    `config.go:146-161` (mutual-exclusion and `--index` arg rules),
    `config.go:184` (invalid `--scheme` flag, eager).
  - env errors: `config.go:171` (`GEMATRIA_LIMIT` not a positive
    integer) and any other site that reads `getenv`.
  - Suggested mechanism: have `parseConfig` return a typed error (for
    example `usageError` / `configError` wrapper structs local to the
    package) and map them in `Run`; do not string-match.
- `run.go:237-244` `exitCodeForComputeError` and `batch.go:77-84`
  `exitCodeForBatchError`: the misuse branch (`InvalidSystemError` or
  `InvalidSchemeError`, currently 2) becomes 78, because flag values
  for `--mispar`/`--scheme` are validated eagerly at parse time
  (`config.go:119`, `config.go:184`), so a compute-time invalid
  system/scheme can only come from a lazily validated env var. Add a
  comment stating that reasoning. The fallthrough branch (currently 1:
  invalid char, unknown name/word) becomes 65.
- `runIndex` (`run.go:91` resolveIndexPath error, currently 2): this is
  env/XDG resolution, return 78. I/O sites `run.go:98,105,121,130,138,144`
  (mkdir, open, create, write, read existing index) become 74. The
  wordlist parse failure at `run.go:111` (`ParseWordListSlice`) becomes
  65 (malformed content), distinct from the open failure above it.
- `runFind` (`run.go:350` no wordlist and no default index, currently
  2): this is missing required input for the flag, return 64.
  `run.go:358` (`openWordSource` failure) becomes 74. `run.go:367`
  (`FindByValue` error, currently flat 1): classify with `errors.As`:
  `InvalidSystemError`/`InvalidSchemeError` return 78, anything else is
  a backend read failure, return 74.
- `processBatch` (`batch.go:64-70`): partial (some ok, some failed,
  currently 4) becomes 1; all-failed (currently 1) becomes 65; all-ok
  stays 0. Update the doc comment table at `batch.go:31-35`.
- `run.go:194` (interactive terminal, no args: usage hint to stderr,
  exit 0): unchanged, deliberate.

## Registry as data (new requirement from the ADR)

Add a declared exit-code table, for example
`src/internal/cli/exitcodes.go`:

```go
// ExitCodes declares every exit code this tool can produce,
// per docs/adr/0001-exit-code-taxonomy.md.
var ExitCodes = map[int]string{
    0:  "success",
    1:  "partial batch success",
    64: "usage error",
    65: "data error (invalid input, unknown word, malformed wordlist, all batch lines failed)",
    74: "I/O error (wordlist/index/backend)",
    78: "configuration error (invalid environment variable value)",
}
```

Add a conformance test (for example `exitcodes_test.go`) that drives
`cli.Run` through at least one scenario per declared code and asserts
(a) the observed code equals the expected code and (b) every observed
code is a key in `ExitCodes`. The existing test helpers (`pipeCapture`,
`makeStdinPipe`, `noenv`/`envWith`) cover all of these scenarios
in-process already.

## Tests to update (inline exit-code assertions)

All in `src/internal/cli/`; update expected integers per the mapping
table: `run_test.go` (asserts 0-4 broadly), `batch_test.go` (0/1/2/4 at
lines 63, 88, 124, 149, 180, 199, 263), `run_index_test.go`,
`run_transliterate_test.go`, `run_find_test.go`, `run_compute_test.go`,
`run_backends_test.go`, `config_test.go`. There are no golden files;
all assertions are inline.

## Docs to update (every stated code changes)

- `AGENTS.md` "Exit Codes" table (around line 246) and the agent
  branching shell example that follows it; also the batch note around
  line 282.
- `README.md` "Exit Codes" table (around line 315) and the batch
  sentence at line 253.
- `docs/specs/cli-design.md` section 6.2 (table around line 408,
  including the "exit code 2 aligns with GNU/Bash convention" sentence,
  which is superseded by the ADR) and section 6.5 batch codes around
  line 440.
- `docs/specs/transliteration.md` lines 426 and 430 (codes 1 and 2).
- `docs/specs/gematria-index.md` line 53 (parse-time rejections, code 2).
- `docs/manual-qa.md` (codes 0/1/2/3/4 mentioned around lines 27, 397).
- `docs/specs/requirements.md` uses "non-zero" language and mostly
  needs no change; scan for literal codes anyway.

## Acceptance

- Every production site above returns the new code; no site returns
  the literals 2, 3, or 4 with the old meanings.
- The declared `ExitCodes` table exists and its conformance test
  exercises every declared code.
- All existing tests updated and passing; `make build` (fmt-check,
  vet, lint, test, compile) passes from the repo root.
- All listed docs updated; no in-repo document still describes the old
  0/1/2/3/4 scheme (`grep -rn "exit" README.md AGENTS.md docs/ | grep
  -i code` shows only the new scheme and the ADR).
- Release notes drafted (breaking change: exit codes renumbered per
  ADR 0001) for the next tag.
- Leave committing to the repo owner; do not create git commits.

## Notes

**2026-07-24T20:37:45Z**

Depends on and relates to gem-kj4m (Migrate Go module from src/ to repo root). After that migration lands, the Go module and all library packages live at the REPOSITORY ROOT, not under src/. Every 'src/...' path in this ticket must be re-checked and rewritten with the src/ prefix dropped: the production sites become internal/cli/run.go, internal/cli/config.go, internal/cli/batch.go; the new files become internal/cli/exitcodes.go and internal/cli/exitcodes_test.go; the test files become internal/cli/*_test.go; and the ADR path docs/adr/0001-exit-code-taxonomy.md is unchanged (docs/ does not move). The cited line numbers may also shift. Sequence the exit-code work after gem-kj4m to avoid rebasing path changes.
