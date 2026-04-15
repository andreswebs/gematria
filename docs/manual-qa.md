# Manual QA Plan — Gematria CLI v0.1.0

## 1. Overview

### Objective

Manually validate the gematria CLI against the behaviors defined in
[docs/specs/requirements.md](specs/requirements.md),
[docs/specs/cli-design.md](specs/cli-design.md), and
[docs/specs/code-architecture.md](specs/code-architecture.md). This plan
covers the golden-path functionality plus error handling, edge cases,
output formats, environment-variable precedence, reverse lookup backends,
and the `index` subcommand.

### Scope

**In scope:**

- Single-letter and multi-letter gematria computation across all four
  systems (`hechrachi`, `gadol`, `siduri`, `atbash`).
- All four output formats (`line`, `value`, `card`, `json`).
- Input modes: positional argument, stdin (batch), Hebrew Unicode, Latin
  transliteration, mixed input.
- Reverse lookup via `--find` across all backends (memory, index, sqlite,
  remote).
- `gematria index` subcommand for both output formats (`sqlite`, `index`).
- Error paths and exit codes (0/1/2/3/4).
- Environment variables and precedence.
- TTY-sensitive behavior: color, no-args usage hint.
- `--help`, `--version`, `--atbash`, `--no-color`, `--fail-early`.
- Word transliteration via `--transliterate` / `-t` and `--scheme`
  (`academic`, `israeli`); `GEMATRIA_SCHEME` env var with lazy validation;
  `Result.Scheme` rendering in JSON and card formats.

**Out of scope:**

- Automated test suite (covered by `make test`).
- Performance/load testing.
- Cross-platform testing beyond macOS/Linux terminals.
- Word list prompt-injection defenses (documented non-feature).

### Environment

| Item              | Value                                                                |
| ----------------- | -------------------------------------------------------------------- |
| OS                | macOS (Darwin) or Linux                                              |
| Shell             | zsh or bash                                                          |
| Build command     | `make build` (produces `bin/gematria-<os>-<arch>`)                   |
| Install step      | `cp bin/gematria-<os>-<arch> ~/.local/bin/gematria` (on `PATH`)      |
| Binary under test | `gematria` on `PATH` (resolved from `~/.local/bin/gematria`)         |
| Go version        | Per `src/go.mod`                                                     |

**Assumption**: `~/.local/bin` is on `PATH` and the installed `gematria`
binary is the one produced by `make build` in this working tree. Verify with
`command -v gematria` and `gematria --version` before starting the session.
No alias is used — every `gematria` invocation below resolves through
`PATH`.

### Pre-Test Setup

Run once from the project root before starting:

```sh
# Build the binary and install it onto PATH.
make build
cp "bin/gematria-$(go env GOOS)-$(go env GOARCH)" ~/.local/bin/gematria

# Sanity-check: the installed binary is the one we just built.
command -v gematria
gematria --version

# Create a test fixtures directory.
TEST_DIR="$(mktemp -d -t gematria-qa-XXXXXX)"
echo "Test fixtures: ${TEST_DIR}"

# Write a minimal plain-text word list.
cat > "${TEST_DIR}/words.txt" <<'EOF'
# Comment line — should be ignored
שלום
אמת
אור
אב
דעת

אהבה
EOF

# Write a TSV word list with transliteration and meaning.
printf 'שלום\tshalom\tpeace\nאמת\temet\ttruth\nאור\tor\tlight\nאב\tav\tfather\nדעת\tdaat\tknowledge\nאהבה\tahava\tlove\n' > "${TEST_DIR}/words.tsv"

# Write a word list with known duplicates for hasMore testing.
# All these words sum to 13 under hechrachi (אהבה=13, אחד=13, גיא=14... pick carefully)
# We'll synthesize a list with many matches for value 1 (just aleph repeated).
printf 'א\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\nא\n' > "${TEST_DIR}/many-alephs.txt"

# Generate index and sqlite backends from words.tsv.
gematria index --wordlist "${TEST_DIR}/words.tsv" --format sqlite --output "${TEST_DIR}/words.db"
gematria index --wordlist "${TEST_DIR}/words.tsv" --format index  --output "${TEST_DIR}/words.idx"

echo "Setup complete. TEST_DIR=${TEST_DIR}"
```

Keep the `TEST_DIR` variable exported for the session. All test cases
reference it.

### Exit Criteria

- All P0 and P1 test cases pass.
- Any failures have filed bug reports with reproduction steps.
- No regression in previously-passing cases between the start and end of
  the session.

---

## 2. Test Cases

Naming convention: `TC-[MODULE]-[NNN]`. Priority per the skill reference:
P0 = blocks release; P1 = fix before release; P2 = fix next release;
P3 = fix when possible.

---

### Module: COMPUTE (positional-argument compute)

#### TC-COMPUTE-001 — Single Hebrew letter, default system, default format

**Priority**: P0 · **Type**: Functional

**Preconditions**: Binary built and installed onto `PATH` (see Pre-Test Setup).

**Steps**:

1. Run `gematria א` → expect: `Aleph (א) = 1` (with RTL/LTR marks around `א`); exit `0`.
2. Run `gematria ת` → expect: `Tav (ת) = 400`; exit `0`.
3. Run `gematria ך` (final kaf) → expect: `Kaf Sofit (ך) = 20` (Hechrachi treats sofit same as base); exit `0`.

**Expected**: Each invocation produces a single line on stdout, empty stderr, exit 0.

---

#### TC-COMPUTE-002 — Latin transliteration, default system

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria aleph` → expect: `Aleph (א) = 1`; exit `0`.
2. Run `gematria ALEPH` → same result (case-insensitive); exit `0`.
3. Run `gematria waw` → expect: `Vav (ו) = 6` (alias match); exit `0`.
4. Run `gematria vaw` → should fail: unrecognized alias. **Verify against current aliases** — if `vaw` is not in the list (only `vav`/`vau`/`waw`), expect exit `1` with suggestions.

**Note**: Per [letters.go:32](src/letters.go#L32), vav aliases are `vav`, `vau`, `waw`. `vaw` is not a valid alias and should fail with suggestions.

---

#### TC-COMPUTE-003 — Multi-letter Hebrew word with breakdown (line format)

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria אמת` → expect stdout to contain `אמת = 441` with per-letter breakdown `(א=1 + מ=40 + ת=400)`; exit `0`.
2. Run `gematria שלום` → expect `שלום = 376` with breakdown `(ש=300 + ל=30 + ו=6 + ם=40)`; exit `0`.

---

#### TC-COMPUTE-004 — Value format outputs bare integer

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --output value אמת` → expect: stdout is exactly `441\n`; exit `0`.
2. Run `gematria -o value aleph` → expect: `1\n`; exit `0`.
3. Run `gematria --output value אמת | wc -l` → expect: `1`.
4. Confirm no RTL marks, no ANSI codes, no trailing metadata.

---

#### TC-COMPUTE-005 — JSON format structure and schema

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --output json אמת` and pipe to `jq`:
   ```sh
   gematria --output json אמת | jq .
   ```
2. Verify the JSON has these top-level keys: `input`, `system`, `total`, `letters`.
3. Verify `total = 441`, `system = "hechrachi"`, `input = "אמת"`.
4. Verify `letters` is a 3-element array, each with `char`, `name`, `value`, `meaning`, `position`.
5. Verify no ANSI codes or RTL marks inside the JSON strings.

---

#### TC-COMPUTE-006 — Card format (multi-line verbose)

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria --output card ד` → expect multi-line output including letter, name, value, position, meaning, and gematria system name; exit `0`.
2. Run `gematria -o card דעת` → expect per-letter table plus total sum.

---

#### TC-COMPUTE-007 — All four gematria systems on the same word

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --mispar hechrachi אמת -o value` → expect `441`.
2. Run `gematria --mispar gadol     אמת -o value` → expect `441` (no sofit in אמת, so same as hechrachi).
3. Run `gematria --mispar siduri    אמת -o value` → expect `36` (1+13+22 — Aleph=1, Mem=13, Tav=22 ordinal).
4. Run `gematria --mispar atbash    אמת -o value` → expect: recompute as Atbash: א↔ת (400), מ↔י (10), ת↔א (1) = `411`.
5. Run `gematria -m gadol דךםןףץ -o value` → should exercise sofit extended values. Sum: ד(4)+ך(500)+ם(600)+ן(700)+ף(800)+ץ(900) = `3504`.

**Note**: Verify case 4 and 5 results match what the spec requires. Report any mismatch as a bug.

---

#### TC-COMPUTE-008 — Multiple Latin tokens treated as separate words

**Priority**: P1 · **Type**: Functional

**Per [AGENTS.md:58-69](AGENTS.md#L58-L69)**, each space-separated Latin token is a separate word.

**Steps**:

1. Run `gematria --output value aleph mem tav` → expect three separate outputs on stdout, one per argument: `1\n40\n400\n`; exit `0`.
2. Run `gematria --output line aleph bet` → expect two lines, `Aleph (א) = 1` then `Bet (ב) = 2`.

**Note**: The CLI iterates positional arguments and computes each independently ([run.go:212-222](src/internal/cli/run.go#L212-L222)).

---

#### TC-COMPUTE-009 — Mixed Hebrew + Latin in single argument

**Priority**: P2 · **Type**: Functional

**Steps**:

1. Run `gematria --output value "אmem"` (quoted to pass as one argument) → expect `41` (א=1 + מ=40).
2. Run `gematria --output json "אtav"` → expect `total = 401` with two letters.

---

#### TC-COMPUTE-010 — Empty input (no args, stdin TTY)

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria` with no args in an interactive terminal (no piped stdin).
2. Expect: stderr receives `Usage: gematria [OPTIONS] [INPUT...] — try 'gematria --help'` (or similar); stdout is empty; exit `0`.
3. The terminal should NOT hang waiting for input.

---

### Module: INPUT (input validation and error paths)

#### TC-INPUT-001 — Invalid Hebrew character (non-Hebrew Unicode)

**Priority**: P0 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria abcאxyz` → expect exit `1`; stdout empty; stderr identifies the invalid character and its position.
2. Run `gematria --output json "foo"` where `foo` is not an alias → expect exit `1`; stdout empty; stderr contains JSON with `error`, `invalid_input`, `position`, `suggestions` fields.

---

#### TC-INPUT-002 — Unknown transliteration with near-match suggestion

**Priority**: P0 · **Type**: Functional (Error Path)

**Per [cli-design.md#3.3](docs/specs/cli-design.md)**: edit distance ≤ half input length, capped at 2.

**Steps**:

1. Run `gematria shen` → expect exit `1`; stderr includes suggestion `shin` (distance 1 from `shen`, length 4, threshold 2).
2. Run `gematria alef` → should resolve (it's a valid alias for Aleph). Exit `0`.
3. Run `gematria xyzzy` → expect exit `1`; stderr contains NO suggestions (distance > 2 from any alias).
4. Run `gematria --output json shen 2>&1 1>/dev/null | jq .` → verify `suggestions` array contains `"shin"`.

---

#### TC-INPUT-003 — Invalid `--mispar` value

**Priority**: P0 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria --mispar standard אמת` → expect exit `2`; stderr includes `Error:`, the invalid value `standard`, and the valid list `hechrachi, gadol, siduri, atbash`.
2. Confirm stdout is empty even with `--output json` (flag-level errors are always plain text per [AGENTS.md:167-169](AGENTS.md#L167-L169)).

---

#### TC-INPUT-004 — Invalid `--output` value

**Priority**: P0 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria --output xml אמת` → expect exit `2`; stderr lists valid formats.

---

#### TC-INPUT-005 — Invalid `--wordlist-format` value

**Priority**: P1 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria --find 1 --wordlist /tmp/x --wordlist-format xml` → expect exit `2`; stderr lists `sqlite, index, remote, memory`.

---

#### TC-INPUT-006 — Prefix matching is NOT supported

**Priority**: P1 · **Type**: Functional (Error Path)

**Per [cli-design.md#3.3](docs/specs/cli-design.md)**: exact match only.

**Steps**:

1. Run `gematria --mispar hech אמת` → expect exit `2` (NOT interpreted as `hechrachi`).
2. Run `gematria --output li אמת` → expect exit `2` (NOT interpreted as `line`).
3. Run `gematria -t --scheme acad shalom` → expect exit `2` (NOT interpreted as `academic`).

---

### Module: STDIN (batch processing)

#### TC-STDIN-001 — Basic stdin batch, all valid

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `printf 'שלום\nאמת\nאור\n' | gematria --output value` → expect three lines on stdout (`376`, `441`, `207`); empty stderr; exit `0`.
2. Run `printf 'שלום\nאמת\nאור\n' | gematria --output json | jq -c .` → expect three JSON objects, one per line; exit `0`.

---

#### TC-STDIN-002 — Stdin batch with mixed valid/invalid lines (partial success)

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run:
   ```sh
   printf 'שלום\nxyzzy\nאור\n' | gematria --output value
   echo "exit=$?"
   ```
2. Expect: stdout has `376\n207\n` (only valid lines); stderr has error for line 2 with `line 2:` prefix; exit `4` (partial success).

---

#### TC-STDIN-003 — Stdin batch with `--fail-early`

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run:
   ```sh
   printf 'שלום\nxyzzy\nאור\n' | gematria --output value --fail-early
   echo "exit=$?"
   ```
2. Expect: stdout has `376\n` only (processing stopped at line 2); stderr has error for line 2; exit `1`.

---

#### TC-STDIN-004 — Stdin batch where all lines fail

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `printf 'xyzzy\nfoobar\n' | gematria --output value` → expect exit `1` (not 4 — all failed, not partial).

---

#### TC-STDIN-005 — Stdin batch with JSON output

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `printf 'שלום\nxyzzy\n' | gematria --output json 2>/tmp/err 1>/tmp/out; echo "exit=$?"`.
2. Verify `/tmp/out` contains one JSON object (success for `שלום`).
3. Verify `/tmp/err` contains one JSON error object with `line: 2`.
4. Verify exit code is `4`.

---

### Module: ENV (environment variable precedence)

#### TC-ENV-001 — `GEMATRIA_MISPAR` changes default

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `GEMATRIA_MISPAR=gadol gematria --output value ך` → expect `500` (gadol sofit).
2. Run `gematria --output value ך` (no env var) → expect `20` (hechrachi default).

---

#### TC-ENV-002 — Flag overrides env var

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `GEMATRIA_MISPAR=gadol gematria --mispar hechrachi --output value ך` → expect `20` (flag wins).

---

#### TC-ENV-003 — `GEMATRIA_OUTPUT` changes default format

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `GEMATRIA_OUTPUT=value gematria אמת` → expect `441\n` (bare number).
2. Run `GEMATRIA_OUTPUT=json gematria --output line אמת` → expect line format (flag wins).

---

#### TC-ENV-004 — `GEMATRIA_WORDLIST` resolves for `--find`

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `GEMATRIA_WORDLIST="${TEST_DIR}/words.txt" gematria --find 26 --output value` → expect words matching value 26 on stdout (if any) or empty result; exit `0` regardless.
2. Run `GEMATRIA_WORDLIST="${TEST_DIR}/words.tsv" gematria --find 441 --output json | jq '.results[].word'` → expect `"אמת"` in results.

---

#### TC-ENV-005 — `GEMATRIA_LIMIT` honored for `--find`

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `GEMATRIA_LIMIT=2 gematria --find 1 --wordlist "${TEST_DIR}/many-alephs.txt" --output json | jq '.results | length'` → expect `2`.
2. Run `GEMATRIA_LIMIT=2 gematria --find 1 --wordlist "${TEST_DIR}/many-alephs.txt" --output json | jq '.hasMore'` → expect `true`.
3. Run `GEMATRIA_LIMIT=2 gematria --limit 5 --find 1 --wordlist "${TEST_DIR}/many-alephs.txt" --output json | jq '.results | length'` → expect `5` (flag wins).

---

#### TC-ENV-006 — Invalid `GEMATRIA_LIMIT` value

**Priority**: P2 · **Type**: Functional (Error Path)

**Steps**:

1. Run `GEMATRIA_LIMIT=notanumber gematria --find 1 --wordlist "${TEST_DIR}/words.txt"` → expect exit `2`; stderr mentions `GEMATRIA_LIMIT` must be positive integer.
2. Run `GEMATRIA_LIMIT=0 gematria --find 1 --wordlist "${TEST_DIR}/words.txt"` → expect exit `2`.
3. Run `GEMATRIA_LIMIT=notanumber gematria אמת` → expect exit `0` (lazy validation — LIMIT only checked when --find is used).

---

#### TC-ENV-007 — Stale `GEMATRIA_WORDLIST` does not block unrelated ops

**Priority**: P1 · **Type**: Functional (Lazy Validation)

**Steps**:

1. Run `GEMATRIA_WORDLIST=/does/not/exist.tsv gematria אמת -o value` → expect `441`; exit `0` (env var not validated because --find not used).
2. Run `GEMATRIA_WORDLIST=/does/not/exist.tsv gematria --find 26` → expect exit `3` (file not found); stderr mentions the path.

---

### Module: FIND (reverse lookup)

#### TC-FIND-001 — Basic reverse lookup (plain text wordlist, memory backend)

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --find 441 --wordlist "${TEST_DIR}/words.txt" --output json | jq .` → verify `value: 441`, `system: "hechrachi"`, `results` contains an object with `word: "אמת"`.
2. Verify `hasMore: false`.
3. Verify plain-text wordlist results omit `transliteration` and `meaning` (not present in input).

---

#### TC-FIND-002 — TSV wordlist produces enriched results

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --find 441 --wordlist "${TEST_DIR}/words.tsv" --output json | jq '.results[0]'` → verify result has `word: "אמת"`, `transliteration: "emet"`, `meaning: "truth"`.

---

#### TC-FIND-003 — `--find` without word list

**Priority**: P0 · **Type**: Functional (Error Path)

**Steps**:

1. Unset `GEMATRIA_WORDLIST` (`unset GEMATRIA_WORDLIST`), then run `gematria --find 26`.
2. Expect exit `2`; stderr explains `--find` needs `--wordlist` or env var.

---

#### TC-FIND-004 — Wordlist file not found

**Priority**: P0 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria --find 26 --wordlist /does/not/exist.tsv` → expect exit `3`; stderr names the path and the OS error.

---

#### TC-FIND-005 — `--limit` and `hasMore` indicator

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --find 1 --wordlist "${TEST_DIR}/many-alephs.txt" --limit 3 --output json | jq '.results | length, .hasMore'` → expect `3` and `true`.
2. Run `gematria --find 1 --wordlist "${TEST_DIR}/many-alephs.txt" --limit 100 --output json | jq '.results | length, .hasMore'` → expect `25` and `false`.
3. Run the same with `--output line` → expect human-readable "more results available" indicator when `hasMore` is true.

---

#### TC-FIND-006 — Reverse lookup respects `--mispar`

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria --mispar siduri --find 36 --wordlist "${TEST_DIR}/words.tsv" --output json | jq '.system, .results[0].word'` → verify `system: "siduri"` and result matches a word whose siduri sum is 36 (e.g., אמת: 1+13+22=36).

---

#### TC-FIND-007 — Comment and blank lines are ignored

**Priority**: P2 · **Type**: Functional

**Steps**:

1. Inspect `${TEST_DIR}/words.txt` — note the `# Comment line` and blank line.
2. Run `gematria --find 441 --wordlist "${TEST_DIR}/words.txt" --output value` → expect only `אמת` (not the comment or blank).

---

### Module: BACKEND (word-source backends)

#### TC-BACKEND-001 — SQLite backend auto-detection via `.db` extension

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --find 441 --wordlist "${TEST_DIR}/words.db" --output json | jq '.results[0].word'` → expect `"אמת"`.
2. Verify the result quality matches the memory backend (TC-FIND-002).

---

#### TC-BACKEND-002 — Index-file backend auto-detection via `.idx` extension

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --find 441 --wordlist "${TEST_DIR}/words.idx" --output json | jq '.results[0].word'` → expect `"אמת"`.

---

#### TC-BACKEND-003 — Companion `.idx` file auto-selected

**Priority**: P1 · **Type**: Functional

**Per [run.go:258-261](src/internal/cli/run.go#L258-L261)**: if `path+".idx"` exists, it is used.

**Steps**:

1. Copy `${TEST_DIR}/words.idx` alongside `${TEST_DIR}/words.tsv` as `${TEST_DIR}/words.tsv.idx`:
   ```sh
   cp "${TEST_DIR}/words.idx" "${TEST_DIR}/words.tsv.idx"
   ```
2. Run `gematria --find 441 --wordlist "${TEST_DIR}/words.tsv" --output json` — this should use the companion .idx (verify via timing or by corrupting the .tsv to prove the .idx is what's being read).
3. To prove .idx was used: run `echo "corrupted" > "${TEST_DIR}/words.tsv"` and re-run the find. Result should still work if .idx is in use.
4. Clean up: `rm "${TEST_DIR}/words.tsv.idx"` and restore `words.tsv` from the setup script.

---

#### TC-BACKEND-004 — Explicit `--wordlist-format memory` override

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria --find 441 --wordlist "${TEST_DIR}/words.db" --wordlist-format memory` → expect exit `3` or an error (treating a SQLite file as a text file will fail parsing or return no results).

**Note**: Record exact behavior — if the tool silently treats the binary file as text, that may be acceptable but worth documenting.

---

#### TC-BACKEND-005 — Remote backend (optional, requires HTTP server)

**Priority**: P2 · **Type**: Integration

**Preconditions**: A test HTTP server running that serves word-list JSON (skip if not available).

**Steps**:

1. With a local server on `localhost:8080`, run `gematria --find 1 --wordlist http://localhost:8080/words --output json`.
2. Verify results are returned.
3. Test auth token: `GEMATRIA_WORDLIST_TOKEN=secret gematria --find 1 --wordlist https://example.com/words` — verify the token is sent in the Authorization header (requires server-side logs or a proxy to inspect).

**Note**: This test is most valuable with a fixture server. Skip if no fixture exists and document why.

---

#### TC-BACKEND-006 — Consistency across backends (same results)

**Priority**: P0 · **Type**: Integration

**Steps**:

1. For value `441` and system `hechrachi`:
   ```sh
   MEM=$(gematria --find 441 --wordlist "${TEST_DIR}/words.tsv" --output json | jq -c '.results')
   DB=$(gematria --find 441 --wordlist "${TEST_DIR}/words.db"  --output json | jq -c '.results')
   IDX=$(gematria --find 441 --wordlist "${TEST_DIR}/words.idx" --output json | jq -c '.results')
   echo "MEM=${MEM}"; echo "DB=${DB}"; echo "IDX=${IDX}"
   ```
2. Verify all three produce equivalent results (word, transliteration, meaning).

---

### Module: INDEX (`gematria index` subcommand)

#### TC-INDEX-001 — Index subcommand with default (sqlite) format

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria index --wordlist "${TEST_DIR}/words.tsv"` → expect a new file `${TEST_DIR}/words.tsv.db`; stdout reports `Indexed N words → <path>`; exit `0`.
2. Verify the `.db` file exists and is non-empty.
3. Verify it's usable: `gematria --find 441 --wordlist "${TEST_DIR}/words.tsv.db"` returns the expected word.

---

#### TC-INDEX-002 — Index subcommand with explicit format and output

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria index --wordlist "${TEST_DIR}/words.tsv" --format index --output "${TEST_DIR}/custom.idx"` → expect `custom.idx` file created; exit `0`.
2. Verify: `gematria --find 441 --wordlist "${TEST_DIR}/custom.idx" --output value` returns the correct value.

---

#### TC-INDEX-003 — Index subcommand missing `--wordlist`

**Priority**: P1 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria index` → expect exit `2`; stderr says `--wordlist is required`.

---

#### TC-INDEX-004 — Index subcommand invalid `--format`

**Priority**: P1 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria index --wordlist "${TEST_DIR}/words.tsv" --format xml` → expect exit `2`; stderr lists valid formats.

---

#### TC-INDEX-005 — Index subcommand missing wordlist file

**Priority**: P1 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria index --wordlist /does/not/exist.tsv` → expect exit `3`; stderr names the path.

---

#### TC-INDEX-006 — Index subcommand `--help`

**Priority**: P2 · **Type**: Functional

**Steps**:

1. Run `gematria index --help` → expect exit `0`; stdout shows index-specific usage.

---

### Module: FORMAT (output format details)

#### TC-FORMAT-001 — RTL marks present in human formats

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria ד | xxd | grep -E '(e2 80 8e|e2 80 8f)'` → expect matches for U+200E (LTR) and U+200F (RTL) byte sequences.
2. Run `gematria --output json ד | xxd | grep -E '(e2 80 8e|e2 80 8f)'` → expect NO matches (JSON should be decoration-free per cli-design).

---

#### TC-FORMAT-002 — `--no-color` flag disables ANSI codes

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria --output card --no-color ד | od -c | grep -E 'esc|033'` → expect no matches.
2. Run `gematria --output card ד` (in a TTY) — may show color. In a pipe (e.g., `gematria -o card ד | cat`), should be auto-plain.

---

#### TC-FORMAT-003 — `NO_COLOR` env var disables color

**Priority**: P2 · **Type**: Functional

**Steps**:

1. Run `NO_COLOR=1 gematria --output card ד` in a TTY — expect no ANSI codes (verify by `NO_COLOR=1 gematria --output card ד | od -c | grep -c 033` → `0`).

---

#### TC-FORMAT-004 — Color auto-off when piped

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria --output card ד | od -c | grep -c 033` → expect `0` (stdout is a pipe, not a TTY; color should be off).

---

#### TC-FORMAT-005 — JSON schema stability (per AGENTS.md)

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --output json ד | jq 'keys'` → expect `["input", "letters", "system", "total"]` (sorted).
2. Run `gematria --output json ד | jq '.letters[0] | keys'` → expect `["char", "meaning", "name", "position", "value"]` (sorted).
3. Verify the output matches the schema documented in [AGENTS.md:89-102](AGENTS.md#L89-L102).

---

#### TC-FORMAT-006 — `--atbash` display modifier

**Priority**: P2 · **Type**: Functional

**Per [cli-design.md#4.3](docs/specs/cli-design.md)**: `--atbash` shows substitution mappings alongside normal output in non-value formats; no effect in value format.

**Steps**:

1. Run `gematria --atbash --output card ד` → expect Atbash mapping visible (e.g., `ד → ק` or similar).
2. Run `gematria --atbash --output json ד | jq .` → inspect for an atbash-related field.
3. Run `gematria --atbash --output value ד` → expect bare number `4` (Atbash flag has no visible effect in value format).

---

### Module: TRANSLITERATE (word transliteration via `--transliterate` / `-t`)

These cases verify the word transliteration feature: typing Hebrew words
phonetically in the Latin alphabet. Spec:
[transliteration.md](specs/transliteration.md). The feature is **opt-in**
via `-t`; without `-t`, Latin tokens remain letter aliases (existing behavior).

#### TC-TRANSLITERATE-001 — Default scheme (academic) basic compute

**Priority**: P0 · **Type**: Functional

**Per [transliteration.md §4.1.4](specs/transliteration.md)**: academic drops
vowels and applies sofit at end-of-word.

**Steps**:

1. Run `gematria -t shalom --output value` → expect `370` (academic = שלם with mem-sofit; ש=300 + ל=30 + ם=40).
2. Run `gematria -t gadol --output value` → expect `37` (academic = גדל; 3+4+30; final ל not a sofit candidate).
3. Run `gematria -t emet --output value` → expect `440` (academic = מת; 40+400; final ת not a sofit candidate, mem in middle stays as מ).
4. Run `gematria -t bereshit --output value` → expect `902` (academic = ברשת; 2+200+300+400).

**Expected**: Exit 0 each time; stdout one line per invocation.

---

#### TC-TRANSLITERATE-002 — Israeli scheme via `--scheme`

**Priority**: P0 · **Type**: Functional

**Per [transliteration.md §4.2.4](specs/transliteration.md)**: israeli uses
matres lectionis (vowels mapped to ו/י/א/ה).

**Steps**:

1. Run `gematria -t --scheme israeli shalom --output value` → expect `376` (שלום; 300+30+6+40).
2. Run `gematria -t --scheme israeli gadol --output value` → expect `43` (גדול; 3+4+6+30).
3. Run `gematria -t --scheme israeli emet --output value` → expect `441` (אמת; 1+40+400).
4. Run `gematria -t --scheme israeli or --output value` → expect `207` (אור; 1+6+200; word-initial 'o' = aleph+vav).
5. Run `gematria -t --scheme israeli shira --output value` → expect `515` (שירה; 300+10+200+5; final 'a' becomes ה).
6. Run `gematria -t --scheme israeli yafe --output value` → expect `95` (יפה; 10+80+5; final 'e' becomes ה).
7. Run `gematria -t --scheme israeli david --output value` → expect `24` (דויד; 4+6+10+4).

---

#### TC-TRANSLITERATE-003 — Scheme produces different values for same input

**Priority**: P0 · **Type**: Functional

This is the canonical test that schemes are actually distinguished.

**Steps**:

1. Run `gematria -t --scheme academic shalom --output value` → expect `370`.
2. Run `gematria -t --scheme israeli  shalom --output value` → expect `376`.
3. Run `gematria -t --scheme academic emet --output value` → expect `440`.
4. Run `gematria -t --scheme israeli  emet --output value` → expect `441`.

**Expected**: Different values per scheme verify scheme dispatch is wired.

---

#### TC-TRANSLITERATE-004 — Sofit applied at word-end

**Priority**: P1 · **Type**: Functional

**Per [transliteration.md §4.5](specs/transliteration.md)**: when the last
letter is one of {כ,מ,נ,פ,צ}, its sofit form is used.

**Steps**:

1. Run `gematria -t shalom --output json | jq -r '.letters[-1].char'` → expect `ם` (mem-sofit), not `מ`.
2. Run `gematria -t shalom --mispar gadol --output value` → expect `930` (sofit mem under gadol = 600; ש=300 + ל=30 + ם=600 = 930).
3. Run `gematria -t --scheme israeli shalom --mispar gadol --output value` → expect `936` (300+30+6+600).

**Expected**: Sofit substitution visible in JSON char field; gadol values reflect extended sofit values.

---

#### TC-TRANSLITERATE-005 — Apostrophe = explicit Aleph in academic

**Priority**: P1 · **Type**: Functional

**Per [transliteration.md §4.1.2](specs/transliteration.md)**: `'` maps to א
in academic.

**Steps**:

1. Run `gematria -t --scheme academic "'emet" --output value` → expect `441` (אמת; 1+40+400).
2. Run `gematria -t --scheme academic emet --output value` → expect `440` (no aleph; 40+400).

**Expected**: Apostrophe adds Aleph; without it, the leading vowel is dropped per academic rules.

---

#### TC-TRANSLITERATE-006 — Multi-token positional input (separate computations)

**Priority**: P1 · **Type**: Functional

**Per [transliteration.md §3.4](specs/transliteration.md)**: each space-separated
positional argument is a separate computation.

**Steps**:

1. Run `gematria -t shalom emet --output value` → expect two lines: `370` then `440` (academic).
2. Run `gematria -t --scheme israeli shalom emet --output value` → expect `376` then `441`.

---

#### TC-TRANSLITERATE-007 — Quoted compound input (combined computation)

**Priority**: P2 · **Type**: Functional

**Per [transliteration.md §3.4](specs/transliteration.md)**: a quoted string is
one computation; internal spaces split into word-parts that are transliterated
independently then concatenated.

**Steps**:

1. Run `gematria -t --scheme israeli "shalom emet" --output value` → expect `817` (376 + 441; one combined result).
2. Run `gematria -t --scheme israeli "shalom emet" --output json | jq '.letters | length'` → expect `7` (4 letters from שלום + 3 from אמת).

**Expected**: One Result with all letters concatenated; total = sum of parts.

---

#### TC-TRANSLITERATE-008 — Hebrew Unicode passes through with `-t`

**Priority**: P1 · **Type**: Functional

**Per [transliteration.md §3.3](specs/transliteration.md)**: Hebrew Unicode
input is unaffected by `-t`.

**Steps**:

1. Run `gematria -t שלום --output value` → expect `376` (Hebrew passes through; same as without `-t`).
2. Run `gematria -t --scheme academic שלום --output value` → expect `376` (scheme doesn't change Hebrew Unicode behavior).

**Expected**: Hebrew input gives identical result regardless of `-t` and scheme.

---

#### TC-TRANSLITERATE-009 — Mode exclusivity: `-t aleph` is NOT the letter Aleph

**Priority**: P0 · **Type**: Functional

**Per [transliteration.md §3.3](specs/transliteration.md)**: with `-t`, `aleph`
is transliterated letter-by-letter (a-l-e-p-h), NOT recognized as the letter
alias.

**Steps**:

1. Run `gematria aleph --output value` (no `-t`) → expect `1` (letter alias for א).
2. Run `gematria -t --scheme academic aleph --output value` → expect `190` (transliterated: a→drop, l→ל, e→drop, p→פ, h→ה; 30+80+5; ph is NOT a digraph mid-word in `aleph` because the 'p' precedes 'h' which is its own consonant — verify the actual academic algorithm: greedy left-to-right parsing on lowercase 'aleph' gives [a,l,e,p,h] → [drop, ל, drop, פ, ה]. Wait — `ph` is in the multi-char table mapping to פ. Greedy match at position 3 sees 'ph' and takes it = פ, leaving nothing else. So [drop, ל, drop, פ] = ל + פ = 30 + 80 = 110. With sofit: last letter is פ → ף = 80 (hechrachi same). Final = 110.).

**Expected**: `gematria aleph` → 1 (letter alias). `gematria -t aleph` → some number that is NOT 1, demonstrating mode exclusivity. The exact value depends on the greedy parse; record what the implementation actually returns and confirm it is documented in the spec or expected per the algorithm.

**Note**: This test verifies the documented "no fallback" rule. The exact
expected value is implementation-driven by the greedy parse. If the value is
unexpected, file a bug against either the spec or the implementation.

---

#### TC-TRANSLITERATE-010 — Case-insensitive input

**Priority**: P1 · **Type**: Functional

**Per [transliteration.md §4.6](specs/transliteration.md)**: input is
lowercased before lookup.

**Steps**:

1. Run `gematria -t --scheme israeli SHALOM --output value` → expect `376`.
2. Run `gematria -t --scheme israeli Shalom --output value` → expect `376`.
3. Run `gematria -t --scheme israeli sHaLoM --output value` → expect `376`.

---

#### TC-TRANSLITERATE-011 — Unmappable input → `UnknownWordError`

**Priority**: P0 · **Type**: Functional (Error Path)

**Per [transliteration.md §4.7](specs/transliteration.md)**: digits and
unmapped punctuation cause errors. Empty letter sequence (e.g., all vowels in
academic) also errors.

**Steps**:

1. Run `gematria -t qzxw --output value 2>&1; echo "exit=$?"` → expect exit `1`; stderr identifies the unmappable input. (Wait — `q` and `z` ARE mapped per §4.1.2: q→ק, z→ז. `x` is mapped to ח. `w` is mapped to ו. So `qzxw` actually parses fully → קזחו = 100+7+8+6 = 121. Pick a truly unmappable input.)
2. Run `gematria -t 'h3llo' --output value 2>&1; echo "exit=$?"` → expect exit `1` (digit `3` is unmappable per §4.7).
3. Run `gematria -t aeiou --output value 2>&1; echo "exit=$?"` → expect exit `1` for academic (all vowels drop → empty sequence).
4. Run `gematria -t --output json 'h3llo' 2>/dev/null; gematria -t --output json 'h3llo' 2>&1 1>/dev/null | jq .` → expect JSON error with `error`, `invalid_input`, `scheme`, `position`, `suggestions: []`.

---

#### TC-TRANSLITERATE-012 — Invalid `--scheme` value

**Priority**: P0 · **Type**: Functional (Error Path)

**Per [transliteration.md §5.3](specs/transliteration.md)**: invalid scheme
flag is exit `2` with valid-list message.

**Steps**:

1. Run `gematria -t --scheme bogus shalom 2>&1; echo "exit=$?"` → expect exit `2`; stderr lists `academic, israeli`.
2. Run `gematria -t --scheme "" shalom 2>&1; echo "exit=$?"` → behavior depends on whether empty-string is treated as "not provided" (default would apply). Record the actual behavior.

---

#### TC-TRANSLITERATE-013 — `GEMATRIA_SCHEME` env var honored

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `GEMATRIA_SCHEME=israeli gematria -t shalom --output value` → expect `376`.
2. Run `GEMATRIA_SCHEME=academic gematria -t shalom --output value` → expect `370`.

---

#### TC-TRANSLITERATE-014 — Flag overrides env var

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `GEMATRIA_SCHEME=israeli gematria -t --scheme academic shalom --output value` → expect `370` (flag wins).
2. Run `GEMATRIA_SCHEME=academic gematria -t --scheme israeli shalom --output value` → expect `376`.

---

#### TC-TRANSLITERATE-015 — Lazy `GEMATRIA_SCHEME` validation

**Priority**: P1 · **Type**: Functional (Lazy Validation)

**Per [transliteration.md §5.3](specs/transliteration.md)**: env var only
validated when `-t` is active.

**Steps**:

1. Run `GEMATRIA_SCHEME=bogus gematria aleph --output value` → expect `1` (without `-t`, env not validated); exit `0`.
2. Run `GEMATRIA_SCHEME=bogus gematria -t shalom 2>&1; echo "exit=$?"` → expect exit `2`; stderr lists valid schemes (with `-t`, env IS validated).

---

#### TC-TRANSLITERATE-016 — Default scheme is `academic`

**Priority**: P1 · **Type**: Functional

**Per [transliteration.md §3.1](specs/transliteration.md)**: default scheme
when `-t` is set without `--scheme` is `academic`.

**Steps**:

1. `unset GEMATRIA_SCHEME`, then run `gematria -t shalom --output json | jq -r '.scheme'` → expect `"academic"`.
2. `unset GEMATRIA_SCHEME`, then run `gematria -t shalom --output value` → expect `370` (academic value).

---

#### TC-TRANSLITERATE-017 — JSON output includes `scheme` field

**Priority**: P0 · **Type**: Functional

**Per [transliteration.md §6.2](specs/transliteration.md)**: JSON uses
`omitempty` — present when transliterated, absent otherwise.

**Steps**:

1. Run `gematria -t --scheme israeli shalom --output json | jq 'keys'` → expect `scheme` to be in the keys list.
2. Run `gematria -t --scheme israeli shalom --output json | jq -r '.scheme'` → expect `"israeli"`.
3. Run `gematria שלום --output json | jq 'has("scheme")'` → expect `false` (no transliteration; scheme omitted).
4. Run `gematria שלום --output json | jq 'keys'` → expect the same keys as before this feature (no `scheme`); confirms backward compatibility.

---

#### TC-TRANSLITERATE-018 — JSON error includes `scheme` field

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria -t --output json 'h3llo' 2>&1 1>/dev/null | jq -r '.scheme'` → expect `"academic"` (scheme attempted).
2. Run `gematria -t --scheme israeli --output json 'h3llo' 2>&1 1>/dev/null | jq -r '.scheme'` → expect `"israeli"`.

---

#### TC-TRANSLITERATE-019 — Card format displays scheme

**Priority**: P2 · **Type**: Functional

**Steps**:

1. Run `gematria -t --scheme israeli shalom --output card` → expect a `Scheme: israeli` line (or similar) in the multi-line output.
2. Run `gematria שלום --output card` → expect NO scheme line (transliteration not used).

---

#### TC-TRANSLITERATE-020 — Composes with `--mispar`

**Priority**: P0 · **Type**: Functional

**Per [transliteration.md §3.5](specs/transliteration.md)**: `-t` composes
orthogonally with all other flags.

**Steps**:

1. Run `gematria -t --scheme israeli --mispar gadol shalom --output value` → expect `936` (300+30+6+600 with sofit-mem-gadol=600).
2. Run `gematria -t --scheme academic --mispar siduri shalom --output value` → expect `64` (siduri ordinals: ש=21, ל=12, ם=13 → wait, siduri sofit = base position; ם=13. Total: 21+12+13 = 46). Re-derive: per [systems.go siduri table](src/systems.go), ם=13. So siduri shalom-academic = 21+12+13 = 46. Correct expected: `46`.

**Note**: Verify the siduri value against [systems.go](src/systems.go) before
declaring a failure.

---

#### TC-TRANSLITERATE-021 — Composes with `--atbash` display modifier

**Priority**: P2 · **Type**: Functional

**Per [transliteration.md §3.5](specs/transliteration.md)**: `--atbash`
display applies to the resolved Hebrew letters.

**Steps**:

1. Run `gematria -t --scheme israeli shalom --atbash --output card` → expect Atbash mappings shown alongside the resolved letters.
2. Run `gematria -t --scheme israeli shalom --atbash --output value` → expect `376` (atbash flag has no visible effect in value format, per existing behavior).

---

#### TC-TRANSLITERATE-022 — Composes with `--mispar atbash`

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria -t --scheme israeli --mispar atbash shalom --output value` → expect the Atbash-substituted value (per [systems.go atbash table](src/systems.go): ש→2, ל→20, ו→80, ם→? — sofit forms map through normal form, so ם → ל's pair → wait, actually the atbash table in systems.go has `'ם': 10`. So 2+20+80+10 = 112. Verify: ש=2, ל=20, ו=80, ם=10 → 112).

---

#### TC-TRANSLITERATE-023 — Stdin batch with `-t`

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `printf 'shalom\nemet\ngadol\n' | gematria -t --scheme israeli --output value` → expect three lines: `376`, `441`, `43`; exit `0`.
2. Run `printf 'shalom\nh3llo\nemet\n' | gematria -t --scheme israeli --output value 2>/tmp/err 1>/tmp/out; echo "exit=$?"` → expect stdout = `376\n441\n`, stderr contains line-2 error, exit `4` (partial success).
3. Run `printf 'shalom\nh3llo\n' | gematria -t --scheme israeli --output value --fail-early; echo "exit=$?"` → expect stdout = `376\n` only, exit `1`.

---

#### TC-TRANSLITERATE-024 — `-t` does NOT affect `--find` (numeric value)

**Priority**: P2 · **Type**: Functional

**Per [transliteration.md §3.5](specs/transliteration.md)**: reverse lookup is
unaffected (the value is numeric).

**Steps**:

1. Run `gematria --find 376 --wordlist "${TEST_DIR}/words.tsv" --output json | jq '.results[0].word'` → expect `"שלום"`.
2. Run `gematria -t --find 376 --wordlist "${TEST_DIR}/words.tsv" --output json | jq '.results[0].word'` → expect `"שלום"` (same result; -t has no effect on --find).

---

#### TC-TRANSLITERATE-025 — Without `-t`, transliteration vocabulary stays as letter aliases

**Priority**: P1 · **Type**: Functional (Regression)

This is the regression test for the original user complaint: `gematria gadol`
without `-t` should still produce the existing error.

**Steps**:

1. Run `gematria gadol 2>&1; echo "exit=$?"` → expect exit `1`; stderr contains `unknown letter name "gadol"` with Levenshtein suggestion (`gamel` per existing alias for ג).
2. Run `gematria gadol --output json 2>&1 1>/dev/null | jq 'has("scheme")'` → expect `false` (no scheme attempted).

---

### Module: META (help, version, misc)

#### TC-META-001 — `--help` output

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --help` → expect exit `0`; stdout contains: `Usage:`, `Options:`, `Environment Variables:`, `Examples:`.
2. Verify `--mispar` documents all four systems.
3. Verify `--output` documents all four formats.
4. Verify `--transliterate` / `-t` is documented.
5. Verify `--scheme` is documented and lists `academic|israeli`.
6. Verify `GEMATRIA_SCHEME` is documented in the Environment Variables section.
7. Verify Examples section includes at least one transliteration example.
8. Run `gematria -h` → same output as `--help`.

---

#### TC-META-002 — `--version` plain text

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --version` → expect `gematria 0.1.0\n` on stdout; exit `0`.

---

#### TC-META-003 — `--version --output json`

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run `gematria --version --output json` → expect a single JSON object with key `version` (and possibly `name`); exit `0`.
2. Run `gematria --version --output json | jq .version` → expect `"0.1.0"`.

---

#### TC-META-004 — Invalid flag produces usage hint

**Priority**: P1 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria --bogus-flag` → expect exit `2`; stderr contains error message.

---

#### TC-META-005 — No output on stdout when error occurs

**Priority**: P0 · **Type**: Functional (Error Path)

**Steps**:

1. Run `gematria xyzzy 1>/tmp/out 2>/tmp/err; echo "exit=$?"`.
2. Verify `/tmp/out` is empty.
3. Verify `/tmp/err` has error content.
4. Verify exit code is non-zero.

---

### Module: UNIX (composability)

#### TC-UNIX-001 — Piping into `jq`

**Priority**: P0 · **Type**: Functional

**Steps**:

1. Run `gematria --find 376 --wordlist "${TEST_DIR}/words.tsv" --output json | jq '.results[].word'`.
2. Expect: `"שלום"` on stdout; exit `0`.

---

#### TC-UNIX-002 — Using in a shell loop

**Priority**: P1 · **Type**: Functional

**Steps**:

1. Run:
   ```sh
   for word in שלום אמת אור; do
     echo -n "${word} = "; gematria --output value "${word}"
   done
   ```
2. Expect three lines, each `<word> = <value>`; exit `0`.

---

#### TC-UNIX-003 — xargs-style processing

**Priority**: P2 · **Type**: Functional

**Steps**:

1. Run `printf 'שלום\nאמת\n' | xargs -I{} gematria --output value {}`.
2. Expect two lines on stdout; exit `0`.

---

## 3. Risks & Mitigations

| Risk                                                 | Mitigation                                                                                                        |
| ---------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Terminal does not render Hebrew RTL correctly        | Verify bytes with `xxd` / `od` rather than relying on visual rendering                                            |
| Shell variable expansion strips Hebrew characters    | Use `printf` with explicit UTF-8 and double-quote Hebrew positional args                                          |
| `GEMATRIA_*` env vars leak between test cases        | `unset GEMATRIA_MISPAR GEMATRIA_OUTPUT GEMATRIA_WORDLIST GEMATRIA_LIMIT GEMATRIA_WORDLIST_TOKEN GEMATRIA_SCHEME` between sections |
| SQLite backend requires CGO or a specific build      | Skip TC-BACKEND-001 if `gematria index --format sqlite` fails at setup                                            |
| Remote backend tests require external infrastructure | Mark TC-BACKEND-005 as skipped with note if no test server is available                                           |
| `TEST_DIR` collisions if tests run in parallel       | Use per-run `mktemp` directory (already in setup)                                                                 |

---

## 4. Bug Reporting

When a test case fails, file a bug using the `qa-test-plan` skill's
`scripts/create-bug-report.sh` or manually capture:

- **Title**: `[Feature] <short description>` (e.g., `[--find] --limit not respected with .idx backend`).
- **Test case ID**: the failing TC.
- **Reproduction**: exact shell commands, environment, expected vs actual.
- **Severity**: per the P0–P3 scale in the skill reference.

Save reports under `docs/bugs/` (create if needed) or attach to the ticket
system.

---

## 5. Cleanup

After the session:

```sh
rm -rf "${TEST_DIR}"
unset GEMATRIA_MISPAR GEMATRIA_OUTPUT GEMATRIA_WORDLIST GEMATRIA_LIMIT GEMATRIA_WORDLIST_TOKEN GEMATRIA_SCHEME

# The installed binary at ~/.local/bin/gematria is left in place so
# subsequent sessions can pick up where this one left off. Remove it
# manually with `rm ~/.local/bin/gematria` if a clean uninstall is desired.
```
