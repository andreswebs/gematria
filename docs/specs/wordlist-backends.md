# Word List Backends

> **Status**: Implemented

## Context

The gematria CLI defines a `WordSource` interface for reverse lookups, with
three implementations:

| Backend            | File                    | Constructor                         |
| ------------------ | ----------------------- | ----------------------------------- |
| In-memory          | `src/wordlist.go`       | `ParseWordList(io.Reader)`          |
| SQLite             | `src/backend_sqlite.go` | `OpenSQLiteWordSource(path)`        |
| Pre-computed index | `src/backend_index.go`  | `NewIndexWordSource(io.ReadSeeker)` |

Backend selection is handled in `internal/cli/run.go` via `openWordSource`,
which detects the backend by file extension, companion `.idx` file, or
explicit `--wordlist-format` override.

The `--index` flag triggers generation of SQLite (`.db`) and index (`.idx`)
files from plain-text or TSV word lists.

A remote/HTTP backend was previously implemented but removed — it was
overengineered for the project's scope. Users who need shared word lists can
distribute `.db` files directly.

---

## Interface

```go
type WordSource interface {
    FindByValue(value int, system System, limit int) ([]Word, bool, error)
}
```

The interface is minimal. It accepts a target value, a gematria system, and a
result limit. It does not support streaming (the `--limit` flag caps results
at a modest number), listing all words, or searching across multiple sources.

`ParseWordListSlice(io.Reader) ([]Word, error)` exposes the raw `[]Word`
slice for callers that need to iterate all words directly (e.g., the
`--index` code path). Listing is an indexing concern, not a lookup concern —
it stays outside `WordSource`.

Multiple word lists are handled by additive indexing: run
`gematria --index --wordlist <path>` for each word list and they merge into
a single `.db`. No `MultiWordSource` wrapper is needed.

---

## Resource Cleanup

Backends that need cleanup (SQLite, index file) implement `io.Closer`
separately. `io.Closer` is NOT embedded in the `WordSource` interface — the
in-memory backend has no cleanup, and forcing a no-op `Close()` on it would
be wrong. The CLI checks via type assertion:

```go
source, closer, err := openWordSource(path, format, getenv)
if closer != nil {
    defer func() { _ = closer.Close() }()
}
```

---

## Backend Selection

The root package defines backends as pure domain types. Backend detection and
construction live in `internal/cli/run.go` (`openWordSource`). The root
package never imports `os`.

Detection order (when `--wordlist-format` is not set):

1. `.db` extension -> SQLite
2. `.idx` extension -> index file
3. companion `.idx` file exists -> index file
4. default -> in-memory

---

## Indexing

Indexing is explicit. `gematria --index --wordlist <path>` must be run before
SQLite or index-file backends can be used. There is no automatic on-first-use
indexing.

Both SQLite and index writes are idempotent and additive:

- **SQLite**: `ON CONFLICT(hebrew) DO NOTHING` with `UNIQUE` constraints on
  `words(hebrew)` and `word_values(word_id, system)`. Running `--index`
  twice with the same word list produces no duplicates. Running it with
  different word lists merges them into one database.
- **Index file**: `WriteIndexFile` deduplicates by Hebrew via a `seen` map.
  `runIndex` reads existing `.idx` entries before writing, merging old and new
  words.

Default index location, env vars (`GEMATRIA_INDEX_LOCATION`,
`GEMATRIA_INDEX_NAME`), and auto-discovery for `--find` are covered in
[gematria-index.md](gematria-index.md).

---

## Format Versioning

Both index formats are derived artifacts — the word list is the source of
truth. Regeneration is the migration path:

- **Index file**: The `# gematria-index v1` magic header is validated on open.
  If the format changes, bump the magic header. `NewIndexWordSource` rejects
  files with an unrecognized header. Users re-run `gematria --index` to
  regenerate.
- **SQLite**: `verifySQLiteSchema` checks that both required tables exist. If
  the schema changes, it rejects old databases with a clear error. Users
  re-index.

No schema version tables, no migration SQL, no forward-compatible format
evolution. The index is cheap to regenerate.

---

## Testing

All backend tests compare results against the in-memory `ParseWordList`
implementation using the same input data (oracle pattern). If the in-memory
backend returns a result, every other backend must return the same result for
the same input.
