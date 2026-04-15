---
id: wor-92gu
status: closed
deps: [wor-alrj, wor-9xcu]
links: []
created: 2026-04-15T04:17:29Z
type: task
priority: 2
parent: wor-6uug
tags: [wordlist-backends, index-subcommand, task]
---
# Add gematria index subcommand for generating word index files

Add a gematria index subcommand to the CLI that reads a plain-text or TSV word list and produces either a SQLite database (.db) or a pre-computed index file (.idx). This subcommand is required before the SQLite and index-file backends can be used.

## Spec References
- docs/specs/wordlist-backends.md §"Pre-Computed Index File" (on-first-use vs explicit indexing) and §"Embedded Database"
- docs/specs/cli-design.md §"Exit codes" (0 success, 2 CLI misuse, 3 file error, 1 domain error)

## Implementation Details

### Dispatch (src/internal/cli/run.go)

At the very top of Run(), before any pflag parsing:
  if len(args) > 0 && args[0] == "index" {
    return runIndex(args[1:], stdin, stdout, stderr, getenv)
  }

This keeps index subcommand flags completely separate from the main flag set.

### runIndex Flags

Parsed with a fresh pflag.FlagSet:
  --wordlist ${WORDLIST_PATH}   required; path to plain text or TSV word list to read
  --output  ${OUTPUT_PATH}      optional; defaults to wordlist+".db" (sqlite) or wordlist+".idx" (index)
  --format  sqlite|index        optional; default "sqlite"

--help on the index subcommand prints its own usage and exits 0.

### Processing Pipeline

1. Validate flags: missing --wordlist → stderr error, return 2.
2. Open --wordlist file → os.Open; failure → stderr error, return 3.
3. gematria.ParseWordList(f) → WordSource (in-memory); close file.
4. Call source.FindByValue for all values and systems? No — iterate the internal word slice directly. Since ParseWordList returns an interface, the index subcommand needs a separate helper:

   func AllWords(source WordSource) ([]gematria.Word, error)

   But WordSource has no list-all method (by spec design). Instead, accept that the index subcommand calls gematria.ParseWordList directly (the in-memory implementation is a package-level function, not just a constructor) and casts or uses a package-internal helper to get the []Word slice for iteration.

   Simpler approach: gematria.ParseWordList already returns a WordSource backed by []Word. Expose a package-internal helper (unexported) or add an exported Words() method to the in-memory type only. Given the root package design, the cleanest path is to add:
     func ParseWordListSlice(r io.Reader) ([]Word, error)
   This returns the raw []Word slice. Used only by the index subcommand; the WordSource interface is still the primary API.

5. For each Word, call Compute(word.Hebrew, system) for each of the 4 systems (Hechrachi, Gadol, Siduri, Atbash). Words that fail Compute for any system are still included for the remaining systems; computation errors are logged to stderr but do not stop processing.

6. Write output:
   - format=sqlite: CREATE TABLE + CREATE INDEX, INSERT all words and their values in a single transaction.
   - format=index: write magic comment line, then all (system, value, hebrew, transliteration, meaning) lines sorted by (system, value). Sort in memory before writing.

7. Print summary to stdout: "Indexed N words → ${OUTPUT_PATH}" (or JSON if --output json is active in the outer Run scope — but since runIndex is dispatched before flag parsing, it always prints plain text).

8. Return 0 on success.

### Exit Codes

- 0: success
- 1: one or more words failed Compute for all systems (partial index written)
- 2: missing required flag or invalid --format value
- 3: cannot open --wordlist or cannot create/write --output file


## Notes

**2026-04-15T13:10:50Z**

Implemented 'gematria index' subcommand dispatched at the top of Run() before any pflag parsing. Three new root-package helpers: ParseWordListSlice (wordlist.go), WriteIndexSQLite (backend_sqlite.go), WriteIndexFile (backend_index.go). runIndex lives in run.go and handles --wordlist (required), --output (optional, defaults to wordlist+.db/.idx), --format sqlite|index (default sqlite). Exit codes: 0 success, 2 CLI misuse, 3 file/write error. Summary line 'Indexed N words → path' on stdout. 11 CLI integration tests cover dispatch, all error paths, default output paths, help, and round-trip correctness via OpenSQLiteWordSource/NewIndexWordSource. Note: --find with a .db or .idx file will not auto-select the backend until wor-22g6 (backend auto-selection) is implemented.
