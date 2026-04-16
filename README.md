# gematria-cli

A command-line tool for **Hebrew gematria** computation. Look up the numeric
value of Hebrew letters and words, choose between four classical gematria
systems, and find words matching a given value from your own word list.

The tool is read-only and dependency-free — no network calls, no state, no
side effects. It composes naturally with pipes, `xargs`, shell scripts, and
`jq`.

---

## Installation

### Homebrew (macOS, Linux)

```sh
brew install andreswebs/tap/gematria
```

### From source

Requires Go 1.26+:

```sh
git clone https://github.com/andreswebs/gematria
cd gematria
make build
./bin/gematria-$(go env GOOS)-$(go env GOARCH) --version
```

### Pre-built binaries

Download the tarball for your platform from the
[releases page](https://github.com/andreswebs/gematria/releases) and place
the `gematria` binary on your `PATH`.

---

## Quick Start

Compute the value of a Hebrew word:

```sh
gematria ‏שלום‎
# ‏שלום‎ = 376 (‏ש‎=300 + ‏ל‎=30 + ‏ו‎=6 + ‏ם‎=40)
```

Get just the number (ideal for piping into scripts):

```sh
gematria --output value ‏שלום‎
# 376
```

Type a Hebrew word phonetically in Latin (academic scheme by default):

```sh
gematria -t shalom
# ‏שלם‎ = 370 (‏ש‎=300 + ‏ל‎=30 + ‏ם‎=40)
```

Find Hebrew words from a word list whose value equals 376:

```sh
gematria --find 376 --wordlist words.tsv
# 1. ‏שלום‎
#    Transliteration: shalom
#    Meaning: peace
```

---

## Input

Three input modes are supported. They can be mixed within a single argument.

### Hebrew Unicode

Pass Hebrew text directly:

```sh
gematria ‏אמת‎              # → 441
gematria ‏דעת‎              # → 474
```

### Latin letter aliases (default Latin mode)

Each space-separated Latin token is one Hebrew letter. Multiple aliases are
accepted (`vav`, `waw`, `vau`); matching is case-insensitive.

```sh
gematria aleph            # → 1   (single letter Aleph)
gematria aleph mem tav    # → three computations: 1, 40, 400
```

To compute a multi-letter word from Latin, pass each letter as a separate
stdin line, or use Hebrew Unicode directly. **For phonetic word
transliteration, use `-t`** (see below).

### Word transliteration (opt-in via `-t`)

When `--transliterate` / `-t` is set, Latin input is interpreted as a phonetic
spelling of a Hebrew word, not as letter aliases. Two schemes are available:

| Scheme     | Style                                                                                                                               | Example: `shalom`               |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------- | ------------------------------- |
| `academic` | strict consonantal (default; vowels dropped)                                                                                        | <bdi dir="rtl">שלם</bdi> → 370  |
| `israeli`  | modern phonetic with matres lectionis (<bdi dir="rtl">ו</bdi>/<bdi dir="rtl">י</bdi>/<bdi dir="rtl">א</bdi>/<bdi dir="rtl">ה</bdi>) | <bdi dir="rtl">שלום</bdi> → 376 |

```sh
gematria -t shalom                          # academic: 370
gematria -t --scheme israeli shalom         # israeli:  376
gematria -t --scheme israeli gadol          # israeli:  43 (‏גדול‎)
```

With `-t`, every Latin token goes through the scheme — there is no fallback
to letter aliases. Without `-t`, behavior is unchanged. Hebrew Unicode passes
through identically in both modes.

Full scheme tables, sofit handling, and ambiguity resolutions are documented
in [docs/specs/transliteration.md](docs/specs/transliteration.md).

---

## Gematria Systems

The `--mispar` / `-m` flag selects the gematria numbering system:

| System      | Description                                                                                     |
| ----------- | ----------------------------------------------------------------------------------------------- |
| `hechrachi` | Standard values (Aleph=1 ... Tav=400). Sofit forms = base. _(default)_                          |
| `gadol`     | Same as Hechrachi for base letters; sofit forms carry extended values 500–900.                  |
| `siduri`    | Ordinal values — each letter is numbered 1–22 by position.                                      |
| `atbash`    | Substitution cipher (Aleph↔Tav, Bet↔Shin...) using Hechrachi values on the substituted letters. |

```sh
gematria --mispar gadol ‏שרה‎        # gadol values
gematria -m siduri ‏אמת‎              # ordinal values
```

The `--atbash` flag is a **display modifier** (separate from `--mispar
atbash`) that shows the Atbash letter mapping alongside the normal output.
Useful for studying the cipher without changing the computed value.

---

## Output Formats

The `--output` / `-o` flag controls how results are rendered:

| Format  | Use case                                                         |
| ------- | ---------------------------------------------------------------- |
| `line`  | Single-line summary with breakdown. _(default; for humans)_      |
| `value` | Bare numeric value, one line. Best for piping into scripts.      |
| `card`  | Multi-line per-letter table with name, value, position, meaning. |
| `json`  | Structured JSON for programmatic consumption.                    |

```sh
gematria --output card ‏דעת‎          # detailed table
gematria --output json ‏שלום‎         # JSON for scripts
gematria -o value ‏שלום‎              # bare 376
```

JSON output is stable across patch versions: existing fields are never
removed or renamed. New fields may be added.

---

## Reverse Lookup

Find Hebrew words from a word list whose gematria value equals a target
number:

```sh
gematria --find 376 --wordlist words.tsv
gematria --find 26  --wordlist words.tsv --output json | jq '.results[].word'
```

The word list comes from `--wordlist` or the `GEMATRIA_WORDLIST` environment
variable. By default at most 20 results are returned; use `--limit` / `-l`
to change this.

### Word list format

Plain text (one Hebrew word per line):

```
‏שלום‎
‏אמת‎
‏אור‎
```

Or TSV with optional transliteration and meaning columns:

```tsv
‏שלום‎	shalom	peace
‏אמת‎	emet	truth
‏אור‎	or	light
```

- Lines starting with `#` are treated as comments.
- Blank lines are ignored.
- The `--mispar` flag affects which gematria system is used to compute values during the search.

---

## Stdin Batch Processing

When no positional argument is given and stdin is not a terminal, lines are
processed one per line:

```sh
printf '‏שלום‎\n‏אמת‎\n‏אור‎\n' | gematria --output value
# 376
# 441
# 207
```

Processing **continues on error** by default — invalid lines produce errors
on stderr (with line numbers); valid lines produce results on stdout. The
exit code is `4` for partial success, `1` if all lines fail.

To stop on the first error:

```sh
cat words.txt | gematria --output value --fail-early
```

Stdin batch works with all input modes, including `-t`:

```sh
printf 'shalom\nemet\ngadol\n' | gematria -t --scheme israeli -o value
```

---

## Composing with Other Tools

```sh
# Pipe JSON through jq for field extraction
gematria --find 376 --wordlist words.tsv --output json | jq -r '.results[].word'

# Use in a shell loop
for w in ‏שלום‎ ‏אמת‎ ‏אור‎; do
  printf '%s = %s\n' "$w" "$(gematria -o value "$w")"
done

# Compare values across systems
echo "‏אמת‎" | gematria -m hechrachi -o value
echo "‏אמת‎" | gematria -m gadol -o value

# Compute the same word in both transliteration schemes
gematria -t --scheme academic -o value shalom
gematria -t --scheme israeli  -o value shalom
```

---

## Environment Variables

Set these in your shell to change the defaults without repeating flags:

| Variable            | Equivalent flag | Notes                                                              |
| ------------------- | --------------- | ------------------------------------------------------------------ |
| `GEMATRIA_MISPAR`   | `--mispar`      | Default gematria system                                            |
| `GEMATRIA_OUTPUT`   | `--output`      | Default output format                                              |
| `GEMATRIA_SCHEME`   | `--scheme`      | Default transliteration scheme; only validated when `-t` is active |
| `GEMATRIA_WORDLIST` | `--wordlist`    | Default word list path for `--find`                                |
| `GEMATRIA_LIMIT`    | `--limit`       | Default result limit for `--find`                                  |
| `NO_COLOR`          | `--no-color`    | Set to any value to disable ANSI color output                      |

**Precedence**: explicit flag > environment variable > built-in default.

Environment variables are validated lazily — only when the feature they
control is invoked. A stale `GEMATRIA_WORDLIST` does not block unrelated
operations.

---

## Exit Codes

| Code | Meaning                                                                                |
| ---- | -------------------------------------------------------------------------------------- |
| 0    | Success                                                                                |
| 1    | Input error (invalid character, unknown name, untranslatable word)                     |
| 2    | CLI misuse (invalid flag value, invalid env var, or missing `--wordlist` for `--find`) |
| 3    | File error (word list not found or unreadable)                                         |
| 4    | Partial success (stdin batch: some lines succeeded, some failed)                       |

Stdout is always empty on error. Errors go to stderr; with `--output json`,
they are emitted as JSON on stderr too.

---

## Help

```sh
gematria --help            # full flag and env var reference
gematria --version         # tool version
```

---

## Authors

**Andre Silva** - [@andreswebs](https://github.com/andreswebs)

## License

This project is licensed under the [GPL-3.0-or-later](LICENSE).
