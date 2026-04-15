# AGENTS.md — gematria CLI agent guide

This document is for automated agents (LLMs, scripts, pipelines) using the
`gematria` CLI. It documents recommended flags, output schemas, error handling,
exit codes, and input formats.

---

## Quick Start

```sh
# Compute a value — structured JSON output
gematria --output json אמת

# Compute from Latin transliteration
gematria --output json aleph

# Get just the number (most compact — ideal for pipelines)
gematria --output value אמת

# Reverse lookup: find Hebrew words whose value equals N
gematria --find 441 --wordlist "${WORDLIST_PATH}" --output json

# Transliterate a Hebrew word from Latin (academic scheme by default)
gematria -t --output json shalom

# Transliterate using the Israeli phonetic scheme (with matres lectionis)
gematria -t --scheme israeli --output json gadol

# Verify which version you are working with
gematria --version --output json
```

---

## Recommended Flags

| Goal                              | Flags                          |
| --------------------------------- | ------------------------------ |
| Structured output for parsing     | `--output json`                |
| Bare number for pipelines         | `--output value`               |
| Select gematria system            | `--mispar hechrachi` (default) |
| All valid `--mispar` values       | `hechrachi`, `gadol`, `siduri`, `atbash` |
| Reverse lookup                    | `--find ${VALUE} --wordlist ${WORDLIST_PATH}` |
| Control result volume             | `--limit ${N}` (default: 20)   |
| Stop batch on first error         | `--fail-early`                 |
| Transliterate Latin → Hebrew word | `--transliterate` / `-t`       |
| Select transliteration scheme     | `--scheme academic` (default) or `--scheme israeli` |
| All valid `--scheme` values       | `academic`, `israeli`          |

Flag values for `--mispar`, `--output`, and `--scheme` require exact match —
no prefix matching.

---

## Input Formats

### Hebrew Unicode

Pass Hebrew text directly as a positional argument:

```sh
gematria --output json שלום
gematria --output json אמת
```

### Latin Transliteration

Each space-separated token resolves to a single Hebrew letter. Each token is
treated as a separate word:

```sh
# Computes aleph as a single-letter word (value 1)
gematria --output json aleph

# Computes three separate words: aleph (1), mem (40), tav (400)
gematria --output json aleph mem tav
```

To compute a multi-letter word by transliteration, pass each letter as a
separate stdin line, or use the Hebrew Unicode directly.

Transliteration matching is case-insensitive. Multiple aliases are accepted
(e.g., `vav`, `waw`, `vaw`). On an unknown name, the error includes
`suggestions` based on Levenshtein distance.

### Mixed Input

Hebrew Unicode and transliteration may be combined within a single argument.

### Word Transliteration (opt-in via `-t`)

When `--transliterate` / `-t` is set, Latin input is interpreted as a phonetic
spelling of a Hebrew word — not as a sequence of letter aliases. The scheme
controls how Latin characters map to Hebrew letters:

- `academic` (default): strict consonantal mapping. Vowels are dropped. Output
  is biblical-style consonantal Hebrew. Apostrophe `'` is the explicit
  Aleph carrier.
- `israeli`: modern Israeli phonetic with matres lectionis (vowels mapped to
  ו/י/א/ה). Output matches modern everyday Hebrew spelling.

```sh
gematria -t shalom                          # → שלם (academic, total 370)
gematria -t --scheme israeli shalom         # → שלום (israeli, total 376)
gematria -t --scheme israeli gadol          # → גדול (israeli, total 43)
```

Mode exclusivity: with `-t`, every Latin token goes through transliteration
— there is no fallback to the letter-alias table. Without `-t`, behavior is
unchanged. Hebrew Unicode input passes through identically in both modes.

The full scheme tables, sofit handling rules, and ambiguity resolutions are
documented in [docs/specs/transliteration.md](docs/specs/transliteration.md).

---

## Output Schemas (JSON)

All JSON output is written to **stdout**. All errors are written to **stderr**.
One JSON object is produced per input (one per line in batch mode).

### Compute result

```json
{
  "input": "אמת",
  "system": "hechrachi",
  "total": 441,
  "letters": [
    {"char": "א", "name": "Aleph", "value": 1,   "meaning": "ox",    "position": 1},
    {"char": "מ", "name": "Mem",   "value": 40,  "meaning": "water", "position": 13},
    {"char": "ת", "name": "Tav",   "value": 400, "meaning": "cross", "position": 22}
  ]
}
```

When `-t` was used, an additional `scheme` field is present (omitted otherwise):

```json
{
  "input": "shalom",
  "system": "hechrachi",
  "scheme": "israeli",
  "total": 376,
  "letters": [ ... ]
}
```

The `scheme` field is `omitempty` — agents that branch on its presence can
distinguish transliterated from non-transliterated results.

### Reverse lookup result

```json
{
  "value": 376,
  "system": "hechrachi",
  "results": [
    {"word": "שלום", "transliteration": "shalom", "meaning": "peace"}
  ],
  "hasMore": false
}
```

`transliteration` and `meaning` are omitted when absent from the word list.
When `hasMore` is `true`, rerun with a larger `--limit` to retrieve more
results.

### Version

```json
{"version": "0.1.0"}
```

---

## Error Handling

Errors always go to **stderr**. Stdout is always empty on error. When
`--output json` is active, errors are also JSON-encoded on stderr.

### JSON error schema (stderr)

```json
{
  "error": "unknown letter name \"xyzzy\"",
  "invalid_input": "xyzzy",
  "position": 0,
  "suggestions": []
}
```

In batch mode, a `"line"` field is added:

```json
{
  "error": "unknown letter name \"xyzzy\"",
  "line": 2,
  "invalid_input": "xyzzy",
  "position": 0,
  "suggestions": []
}
```

Fields:

| Field           | Description                                              |
| --------------- | -------------------------------------------------------- |
| `error`         | Human-readable error message                             |
| `line`          | 1-based source line number (batch mode only)             |
| `invalid_input` | The specific input that triggered the error              |
| `scheme`        | Transliteration scheme attempted (only for `UnknownWordError`) |
| `position`      | 0-based position within the input string (where known)   |
| `suggestions`   | Near-match candidates for unknown names (may be empty)   |

When `-t` was active and the input could not be transliterated, the JSON
error includes `scheme` so agents can branch on which scheme was attempted:

```json
{
  "error": "input \"h3llo\" cannot be transliterated in scheme \"academic\"",
  "invalid_input": "h3llo",
  "scheme": "academic",
  "position": 0,
  "suggestions": []
}
```

**Note:** Flag-level errors (e.g., invalid `--mispar` value, invalid
`--scheme` value) are always plain text on stderr, even with `--output json`,
because they occur before the formatter is initialized.

---

## Exit Codes

| Code | Meaning                                                                      |
| ---- | ---------------------------------------------------------------------------- |
| 0    | Success                                                                      |
| 1    | Input error — invalid character or unknown transliteration name              |
| 2    | CLI misuse — invalid flag value, invalid env var, or missing `--wordlist` when using `--find` |
| 3    | File error — word list not found or unreadable                               |
| 4    | Partial success — stdin batch: some lines succeeded, some failed             |

Agent branching example (shell):

```sh
gematria --output json "${INPUT}"
CODE=$?
case $CODE in
  0) echo "success" ;;
  1) echo "bad input — check suggestions field in stderr JSON" ;;
  2) echo "misuse — check flag/env values" ;;
  3) echo "file error — check word list path" ;;
  4) echo "partial success — some lines failed" ;;
esac
```

---

## Batch Processing

When no positional argument is given and stdin is not a terminal, the tool
reads one entry per line:

```sh
printf 'שלום\nאמת\nאור\n' | gematria --output json
```

Each line produces one JSON object on stdout (or one JSON error on stderr).
Processing continues on error by default; exit code 4 indicates partial
success.

Use `--fail-early` to stop on the first error:

```sh
cat "${WORDS_FILE}" | gematria --output value --fail-early
```

---

## Environment Variables

| Variable            | Equivalent flag    | Notes                                                |
| ------------------- | ------------------ | ---------------------------------------------------- |
| `GEMATRIA_MISPAR`   | `--mispar`         | Default gematria system                              |
| `GEMATRIA_OUTPUT`   | `--output`         | Default output format                                |
| `GEMATRIA_SCHEME`   | `--scheme`         | Default transliteration scheme; lazy validation (only checked when `-t` is active) |
| `GEMATRIA_WORDLIST` | `--wordlist`       | Default word list path for `--find`                  |
| `GEMATRIA_LIMIT`    | `--limit`          | Default result limit for `--find`                    |
| `NO_COLOR`          | `--no-color`       | Set to any value to disable ANSI color               |

Precedence: explicit flag > environment variable > built-in default.

Environment variables are validated lazily — only when the feature they
control is invoked. A stale `GEMATRIA_WORDLIST` does not cause errors unless
`--find` is used without an overriding `--wordlist` flag. Likewise,
`GEMATRIA_SCHEME` is only validated when `-t` is set.

---

## Security Note

The `--wordlist` file contents appear verbatim in output. If you pipe
`gematria` output into an LLM, a malicious word list could contain prompt
injection payloads. The tool is a data pipe — it does not sanitize word list
content. Only use word lists from trusted sources.
