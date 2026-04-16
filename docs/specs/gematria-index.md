# Gematria Index — Default Location & Auto-Discovery

> **Status**: Accepted

## Context

The `--index` flag triggers generation of pre-computed index files (SQLite or
TSV) from word lists. The default output path follows XDG conventions, and
`--find` auto-discovers the default index when no explicit word list is
provided.

Previously, indexing was a `gematria index` subcommand with its own flag set.
It has been redesigned as a flag (`--index`) on the main command, with
`--index-output` and `--index-format` for index-specific options. This
eliminates the subcommand dispatch and keeps the CLI as a single flat command.

---

## User-Facing Design

### Invocation

```sh
# Build default index (SQLite at XDG default path)
gematria --index --wordlist words.txt

# Build with explicit output path
gematria --index --wordlist words.txt --index-output custom.db

# Build as .idx format
gematria --index --wordlist words.txt --index-format index

# Use default index for reverse lookup (no --wordlist needed after indexing)
gematria --find 376

# Reverse lookup with explicit word list (bypasses default index)
gematria --find 376 --wordlist words.tsv
```

### Flags

| Flag                         | Short | Default   | Notes                                          |
| ---------------------------- | ----- | --------- | ---------------------------------------------- |
| `--index`                    | —     | off       | Triggers index-building mode                   |
| `--index-output`             | —     | (XDG)     | Explicit output file path (bypasses env vars)  |
| `--index-format`             | —     | `sqlite`  | Index format: `sqlite` or `index`              |

The `--wordlist` flag (or `GEMATRIA_WORDLIST` env var) provides the source
word list. Same resolution chain as `--find`: flag > env var > error.

### Flag Conflicts

These combinations are rejected at parse time with exit code 2:

| Combination                       | Error message                                         |
| --------------------------------- | ----------------------------------------------------- |
| `--index` + `--find`              | `--index and --find are mutually exclusive`            |
| `--index` + `-t`                  | `--index and --transliterate are mutually exclusive`   |
| `--index-output` without `--index`| `--index-output requires --index`                     |
| `--index-format` without `--index`| `--index-format requires --index`                     |
| `--index` + positional arguments  | `--index does not accept positional arguments`         |

Irrelevant flags (`--mispar`, `--atbash`, `--output`, `--fail-early`,
`--limit`, `--scheme`) and their env vars are silently ignored when `--index`
is set. They are parsed but unused — consistent with the lazy validation
pattern.

### `--help` Grouping

Indexing flags are grouped under a separate "Indexing:" heading in `--help`:

```
Options:
  -m, --mispar string           ...
  -o, --output string           ...
  ...

Indexing:
      --index                   Build a pre-computed index from --wordlist
      --index-output string     Output file path (bypasses env var resolution)
      --index-format string     Index format: sqlite|index (default: sqlite)
```

### Breaking Change

The `gematria index` subcommand is removed. There is no migration shim.
Old invocations like `gematria index --wordlist words.txt` will fail with
an unknown-name error. This is a v0.x breaking change.

---

## Design Decisions

### 1. Default index location follows XDG

The default index directory is resolved as:

```
GEMATRIA_INDEX_LOCATION
  > XDG_DATA_HOME/gematria
    > ~/.local/share/gematria
```

`XDG_DATA_HOME` is used on all platforms (including macOS). The index is
persistent user data — not cache, not config — so `XDG_DATA_HOME` is the
correct XDG base directory.

`GEMATRIA_INDEX_LOCATION` takes precedence over `XDG_DATA_HOME` when set.

### 2. Default index filename

The default filename (without extension) is controlled by:

```
GEMATRIA_INDEX_NAME  >  "gematria"
```

The format extension (`.db` or `.idx`) is appended automatically based on
`--index-format`. The resulting default paths are:

```
~/.local/share/gematria/gematria.db     # --index-format sqlite (default)
~/.local/share/gematria/gematria.idx    # --index-format index
```

`GEMATRIA_INDEX_NAME` must be a bare filename — path separators (`/`, `\`)
are rejected. Dots are allowed (e.g., `my.index` → `my.index.db`).

### 3. Output path precedence for `--index`

```
--index-output flag (literal full path, no auto-extension)
  > GEMATRIA_INDEX_LOCATION / GEMATRIA_INDEX_NAME.<ext>
    > GEMATRIA_INDEX_LOCATION / gematria.<ext>
      > XDG_DATA_HOME/gematria/gematria.<ext>
        > ~/.local/share/gematria/gematria.<ext>
```

When `--index-output` is set, it is used as-is — no extension appended, no
directory resolution. It completely bypasses both env vars.

### 4. Auto-discovery for `--find`

When `--find` is used without `--wordlist` and without `GEMATRIA_WORDLIST`,
the tool checks the default XDG index location for an existing index file:

1. Try `<location>/gematria.db` (SQLite — preferred for performance)
2. Try `<location>/gematria.idx` (index file — fallback)
3. Error if neither exists

The full wordlist resolution precedence for `--find`:

```
--wordlist flag
  > GEMATRIA_WORDLIST env
    > default XDG index (.db first, then .idx)
      > error
```

### 5. Directory auto-creation

The resolved index directory is created automatically with
`os.MkdirAll(path, 0o755)` when it does not exist. This applies during
`--index` only — `--find` auto-discovery does not create directories.

### 6. Lazy validation

`GEMATRIA_INDEX_LOCATION` and `GEMATRIA_INDEX_NAME` are validated only when
`--index` is active. A stale or invalid value does not cause errors for
unrelated commands (e.g., `gematria --output json אמת`).

This is consistent with the existing env var validation strategy
(`GEMATRIA_SCHEME`, `GEMATRIA_WORDLIST`).

### 7. Wordlist source for `--index`

`--index` uses the same `--wordlist` / `GEMATRIA_WORDLIST` resolution chain
as `--find`. This enables the zero-config workflow:

```sh
export GEMATRIA_WORDLIST=~/words.tsv
gematria --index                    # indexes ~/words.tsv
gematria --find 376                 # searches the default index
```

Stdin is NOT accepted as a word list source for `--index` — stdin is reserved
for compute batch mode. `--wordlist` (or env var) is always required.

---

## Environment Variables (updated)

| Variable                 | Equivalent flag | Notes                                                    |
| ------------------------ | --------------- | -------------------------------------------------------- |
| `GEMATRIA_MISPAR`        | `--mispar`      | Default gematria system                                  |
| `GEMATRIA_OUTPUT`        | `--output`      | Default output format                                    |
| `GEMATRIA_SCHEME`        | `--scheme`      | Default transliteration scheme; lazy validation          |
| `GEMATRIA_WORDLIST`      | `--wordlist`    | Default word list path for `--find` and `--index`        |
| `GEMATRIA_LIMIT`         | `--limit`       | Default result limit for `--find`                        |
| `GEMATRIA_INDEX_LOCATION`| —               | Directory for index files; overrides XDG_DATA_HOME       |
| `GEMATRIA_INDEX_NAME`    | —               | Index filename without extension; default: `gematria`    |
| `XDG_DATA_HOME`          | —               | XDG base; default: `~/.local/share`                      |
| `NO_COLOR`               | `--no-color`    | Disable ANSI color                                       |

Precedence: explicit flag > env var > built-in default.

---

## Implementation Plan

### Step 1: Migrate index flags to Config

**Files**: `src/internal/cli/config.go`, `src/internal/cli/config_test.go`

Add to `Config`:
```
Index       bool
IndexOutput string
IndexFormat string
```

Add flag definitions:
```
fs.BoolVar(&index, "index", false, "build a pre-computed index from --wordlist")
fs.StringVar(&indexOutput, "index-output", "", "output file path for index")
fs.StringVar(&indexFormat, "index-format", "sqlite", "index format: sqlite|index")
```

Add validation:
- `--index-format` must be `sqlite` or `index` (exit 2 on invalid).
- Flag conflict checks (5 rejections from the table above).
- `--index` without `--wordlist` (and no `GEMATRIA_WORDLIST`): reject.

Remove the old `runIndex` pflag set entirely.

### Step 2: Wire index mode in Run()

**File**: `src/internal/cli/run.go`

- Remove `args[0] == "index"` subcommand dispatch at the top of `Run()`.
- Remove the old `runIndex` function and `indexHelpText`.
- After config parsing, add a new branch:
  ```
  if cfg.Index {
      return runIndex(cfg, stdout, stderr, getenv)
  }
  ```
- Write a new `runIndex(cfg Config, ...)` that uses `cfg.Wordlist`,
  `cfg.IndexOutput`, `cfg.IndexFormat` from the unified config.
- When `cfg.IndexOutput` is empty, call `resolveIndexPath(cfg.IndexFormat, getenv)`.
- Auto-create directory, parse word list, write index — same logic as today.

Update `helpText` to add the "Indexing:" section.

### Step 3: Migrate and add tests

**Files**: `src/internal/cli/run_index_test.go`, `src/internal/cli/config_test.go`

- Rewrite all `run_index_test.go` tests to use `--index` flag syntax.
- Add conflict-rejection tests for all 5 combinations.
- Add positional-args-with-`--index` rejection test.
- Add `GEMATRIA_WORDLIST` resolution test for `--index`.
- Verify the old `gematria index ...` invocation fails (no migration shim).

### Step 4: Update docs

**Files**: `AGENTS.md`, `README.md`, `CLAUDE.md`

- Replace `gematria index --wordlist ...` with `gematria --index --wordlist ...`.
- Replace `--format` with `--index-format` where it appears.
- Replace `--output` (index path) with `--index-output`.
- Add `--index`, `--index-output`, `--index-format` to flag tables.
- Update examples in Quick Start sections.

---

## Behavior Change Summary

| Scenario | Before | After |
| --- | --- | --- |
| `gematria index --wordlist w.txt` | Builds index | Error (unknown name "index") |
| `gematria --index --wordlist w.txt` | Error (unknown flag) | Builds index at default XDG path |
| `gematria --index --wordlist w.txt --index-output x.db` | N/A | Builds index at `x.db` |
| `gematria --index --find 376` | N/A | Error: mutually exclusive |
| `gematria --find 376` (no wordlist) | Auto-discovers default index | Auto-discovers default index (unchanged) |
| `gematria --find 376 --wordlist w.db` | Uses `w.db` | Uses `w.db` (unchanged) |
