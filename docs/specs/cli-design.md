# Gematria CLI - CLI Design (DRAFT)

> **Status**: Ready

This document describes the UX design principles and patterns for the gematria
CLI. It covers both human and agent interaction modes. It does not prescribe
implementation details — only the external behavior and design rationale.

The requirements it builds on live in [requirements.md](requirements.md).

---

## 1. Design Philosophy

The gematria CLI is a **single-purpose lookup tool**. It computes gematria
values, displays letter data, and performs reverse lookups against a word list.
It produces no side effects: no files written, no state mutated, no network
calls. This read-only nature is a design advantage — it removes entire
categories of risk (destructive operations, confirmation prompts, dry-run
modes) and lets us focus on **input clarity** and **output precision**.

### Dual-Mode Principle

The CLI must serve two distinct audiences from the same binary:

- **Humans** working in a terminal — they need readable output, visual
  hierarchy, RTL-aware rendering, contextual help, and forgiving input
  handling.
- **Agents** (LLMs, scripts, pipelines) — they need structured output,
  predictable parsing, minimal token waste, and strict input validation with
  actionable error messages.

These audiences have overlapping but different needs. The design must satisfy
both without requiring a separate "agent mode" flag. TTY detection drives
cosmetic behavior (color, usage hints) but not output format — the default
output format is always `line`, regardless of context. Agents request `json`
explicitly via `--output json`.

### Unix Conventions

The CLI follows standard Unix behavior:

- Results go to **stdout**. Errors go to **stderr**.
- **Exit code 0** on success, **non-zero** on error.
- Input comes from a **positional argument** or **stdin** (one entry per
  line).
- Flags use **GNU-style** long (`--mispar`) and short (`-m`) forms.
- Environment variables provide persistent defaults; flags override them.
- The tool is composable: it works with pipes, `xargs`, shell loops, and
  other standard mechanisms.

---

## 2. Agent DX Assessment

Using the Agent DX CLI Scale (0–3 per axis, 0–21 total), this section sets
design targets for each axis. The goal is **Agent-tolerant** (6–10) — the
right range for a read-only, single-purpose tool with a small flag surface.

### 2.1 Machine-Readable Output — Target: 2

The CLI already requires four output formats (`line`, `value`, `card`, `json`).
To reach a score of 2:

- The `json` format must be **consistent and complete** across all operations
  (single letter, word, reverse lookup).
- Errors must also produce **structured JSON on stderr** when `--output json`
  is active, so agents can parse failures without heuristics. When any other
  output format is selected, errors remain plain text on stderr.
- The default output format is always `line`, regardless of TTY context.
  Agents pass `--output json` explicitly. This follows the principle of least
  surprise — `gematria aleph` and `echo aleph | gematria` produce the same
  format.

**Why not 3**: NDJSON streaming is not justified for this tool's output volume.
Reverse lookup results are paginated but finite. Reaching score 3 would add
complexity without meaningful benefit.

### 2.2 Raw Payload Input — Target: 1

This tool is read-only — there are no mutating commands and no underlying API
schema to mirror. Input is either Hebrew text or transliterated names, not
structured payloads.

Score 1 is appropriate: stdin provides a channel for bulk input (one word per
line), which is the natural "raw" input mode for this domain.

**Why not higher**: Raw JSON input would add a translation layer where none is
needed. The input domain is strings, not objects.

### 2.3 Schema Introspection — Target: 1

A `--help` flag with clear, complete usage information is the baseline. To
reach score 1:

- The `--help` output must document all flags, environment variables, valid
  values for enum-like flags (`--mispar`, `--output`), and include examples.
- Valid values for constrained flags (e.g., `--mispar hechrachi|gadol|siduri|atbash`)
  must be enumerated in help text so agents can discover them without
  documentation.

**Why not higher**: Full JSON schema introspection is overengineered for a tool
with a handful of flags. The flag surface area is small enough that help text
suffices.

### 2.4 Context Window Discipline — Target: 2

Agents have finite context windows. The CLI must help them control response
size:

- The `--output value` format returns only the numeric result — the most
  compact representation possible. This is the ideal agent format for simple
  lookups.
- The `--limit` flag on reverse lookups directly controls result volume.
- No `--fields` flag. The tool's output is small enough that field filtering
  adds complexity without meaningful token savings. Agents that need specific
  JSON fields can pipe through `jq`.

**Why not 3**: Streaming pagination and skill-file guidance on field usage would
be premature for a tool of this scope.

### 2.5 Input Hardening — Target: 2

Agents hallucinate differently than humans typo. The CLI must validate inputs
with this in mind:

- Reject characters outside the expected Hebrew Unicode range (U+05D0–U+05EA
  plus sofit forms) and the ASCII range used for transliteration.
- Reject control characters, zero-width characters, and homoglyph
  lookalikes that an LLM might produce when "writing Hebrew."
- Error messages must identify the **invalid character and its position** (per
  requirements), giving the agent a precise correction signal rather than a
  vague rejection.
- For enum-like flags (`--mispar`, `--output`), reject invalid values with a
  message that lists all valid options — this serves as a self-correcting
  loop for agents.

**Why not 3**: Path traversal and sandboxing concerns don't apply — the tool
reads one optional file (word list) and writes nothing.

### 2.6 Safety Rails — Target: 0

The tool is entirely read-only. There are no mutating operations to dry-run,
no state to corrupt, no side effects to prevent. A dry-run mode would be
meaningless. Response sanitization against prompt injection is not applicable
since the tool returns its own computed data, not untrusted third-party
content.

The one exception is the **word list file**: its contents are user-supplied and
appear in output. If an agent uses this tool and pipes its output into an LLM,
a malicious word list could contain prompt injection payloads. The CLI passes
word list content through verbatim — it is a data tool, not a security
boundary. This risk should be documented in `AGENTS.md`.

### 2.7 Agent Knowledge Packaging — Target: 1

The project already has an `AGENTS.md` file (currently a placeholder). To
reach score 1:

- `AGENTS.md` should document the recommended agent workflow: which flags to
  use, which output format to prefer, how to handle errors, and what the exit
  codes mean.
- A brief "quick start for agents" section should provide copy-paste examples
  that an agent can use directly.

**Why not higher**: Structured skill files and an agent guardrail library are
overkill for a single-purpose tool with a small flag surface.

### Projected Total: 8 (Agent-tolerant)

| Axis                      | Target | Notes                                               |
| ------------------------- | ------ | --------------------------------------------------- |
| Machine-Readable Output   | 2      | Consistent JSON, structured errors on stderr        |
| Raw Payload Input         | 1      | Stdin bulk input                                    |
| Schema Introspection      | 1      | Complete `--help` with enum values                  |
| Context Window Discipline | 1      | `--output value` and `--limit`                      |
| Input Hardening           | 2      | Unicode validation, positional errors, enum listing |
| Safety Rails              | 0      | Read-only tool, no mutations                        |
| Agent Knowledge Packaging | 1      | `AGENTS.md` with workflow guidance                  |

This is appropriate for a read-only lookup tool. The investment in
machine-readable output and input hardening gives agents a good experience
without over-engineering the tool's small surface area.

---

## 3. Input Design

### 3.1 Positional Arguments

The primary input mode is a positional argument:

```
gematria אלף
gematria aleph
gematria aleph bet gimel
```

When multiple space-separated Latin names are given, they resolve to a Hebrew
word (e.g., `aleph bet` → `אב`). This is a natural, human-friendly input
mode.

For agents, positional arguments are also clean — no flag ceremony for the
most common operation.

### 3.2 Stdin

When no positional argument is provided, the tool reads from stdin, one entry
per line. This enables pipeline composition:

```
echo "אמת" | gematria
cat words.txt | gematria --output value
```

For agents, stdin is the natural batch input channel. Each line produces one
result, making output predictable and parseable.

### 3.3 Transliteration Matching

Case-insensitive matching against aliases (e.g., "vav", "waw", "vaw") is
human-friendly. Agents benefit from the same flexibility since LLMs may
produce varying transliterations. Flag values (`--mispar`, `--output`) require
exact match — no prefix matching, no fuzzy matching. The valid set is small
and stable; prefix matching would create a hidden contract that breaks when
new values are added.

The error path matters most here: when a transliteration doesn't match, the
error message should list near-matches (Levenshtein-based suggestions) so
both humans and agents can self-correct. Suggestions fire only when the edit
distance is at most half the input length (rounded up), capped at a maximum
distance of 2. This avoids noise on garbage input (`xyzzy` gets no
suggestion) while catching real typos and close hallucinations (`shen` →
`shin`).

**Word transliteration mode**: An opt-in `--transliterate` / `-t` flag
switches Latin parsing from letter-alias mode to word transliteration. This
is a separate feature with its own design — see
[transliteration.md](transliteration.md). Without `-t`, the alias behavior
described above applies.

### 3.4 Input Conflict Resolution

Flags always override environment variables. This is stated in the
requirements and is standard Unix practice. The precedence order:

1. Explicit flag (`--mispar gadol`)
2. Environment variable (`GEMATRIA_MISPAR=gadol`)
3. Built-in default (`hechrachi`)

This should be documented in `--help` and in `AGENTS.md`.

---

## 4. Output Design

### 4.1 Format Selection

Four formats serve distinct use cases:

| Format  | Audience       | Use Case                         |
| ------- | -------------- | -------------------------------- |
| `line`  | Humans         | Quick lookup, single-line answer |
| `value` | Agents/scripts | Bare numeric result for piping   |
| `card`  | Humans         | Detailed study, reference card   |
| `json`  | Agents/scripts | Full structured data for parsing |

### 4.2 Stable Default Format

The default output format is always `line`, regardless of whether stdout is a
TTY or a pipe. This follows the principle of least surprise: `gematria aleph`
and `echo aleph | gematria` produce identical output.

TTY detection controls **cosmetic** behavior only:

- **Color**: Enabled when stdout is a TTY, disabled when piped. Respects
  `NO_COLOR` env var and `--no-color` flag.
- **No-args behavior**: When stdin is a TTY and no arguments are given, print
  a short usage hint to stderr and exit (instead of blocking on stdin). When
  stdin is piped, read from it as normal.

Agents request structured output explicitly via `--output json`. The `value`
format (`--output value`) is the most efficient for agents that need only the
numeric result.

### 4.3 Human Output Patterns

For human-facing formats (`line`, `card`):

- **Visual hierarchy**: Use bold/bright for primary information (letter,
  value), normal weight for secondary (name, meaning), and dim for metadata
  (gematria system name, position).
- **Symbols**: Use `=` for value assignment, `+` for breakdowns, `→` for
  Atbash mappings. These are ASCII-safe and carry meaning without color.
- **Spacing**: Use blank lines to separate sections in `card` output. Align
  columns in per-letter breakdown tables.
- **RTL marks**: Prefix Hebrew text with U+200F (RTL mark) and suffix with
  U+200E (LTR mark) per requirements. This is invisible but critical for
  correct terminal rendering and copy-paste behavior.
- **Color**: Use semantic color sparingly — green for values, default for
  text. Never encode meaning through color alone. Color is controlled by
  three mechanisms in priority order:
  1. `--no-color` flag (always disables)
  2. `NO_COLOR` env var (disables if set)
  3. TTY detection (enabled on TTY, disabled when piped)
- **`--atbash` in `value` format**: The `--atbash` flag is a display modifier
  that shows Atbash letter mappings alongside normal output. In `value`
  format, which strips all presentation, `--atbash` has no visible effect.
  To compute a value using the Atbash system, use `--mispar atbash`.

### 4.4 Agent Output Patterns

For agent-facing formats (`json`, `value`):

- **`value` format**: Bare number, no decoration, no trailing newline
  character ambiguity. Just the integer, followed by a newline.
- **`json` format**: A single JSON object per invocation. For single letters:
  all fields. For words: a `letters` array and a `total`. For reverse
  lookups: a `results` array with a `hasMore` boolean.
- **Stable schema**: The JSON schema must not change between minor versions.
  New fields may be added; existing fields must not be removed or renamed.
  This is a contract with agent consumers.
- **No decoration**: No ANSI codes, no RTL marks, no Unicode symbols in JSON
  values. Data only.

### 4.5 Structured Errors

Errors should be structured in both modes:

- **Human mode** (stderr, plain text): A clear sentence identifying what went
  wrong, what's valid, and how to fix it. Example:
  `Error: Unknown letter 'x' at position 3. Did you mean 'shin'?`
- **Agent mode** (stderr, JSON when `--output json`): A JSON object with
  `error`, `position`, `invalid_input`, and `suggestions` fields. This gives
  agents a precise, parseable correction signal.

In both cases, **stdout produces no output on error** and the exit code is
non-zero.

---

## 5. Help and Discoverability

### 5.1 The --help Flag

`--help` is the primary discoverability surface. It must include:

- **Usage line**: `gematria [OPTIONS] [INPUT...]`
- **Description**: One-sentence summary of what the tool does.
- **Arguments**: What `INPUT` accepts (Hebrew characters or transliterated
  names).
- **Flags**: All flags with short forms, descriptions, and valid values for
  enum-type flags.
- **Environment variables**: Each variable, what it controls, and its
  relationship to the corresponding flag.
- **Examples**: At least 3 examples covering: single letter lookup, word
  lookup, reverse lookup. Show both Hebrew and transliterated input.

### 5.2 Flag Validation Messages

When a flag receives an invalid value, the error message must:

1. Name the flag that's invalid.
2. Show the invalid value provided.
3. List all valid values.

Example:

```
Error: Invalid value 'standard' for --mispar.
Valid values: hechrachi, gadol, siduri, atbash
```

This pattern is equally useful for humans and agents — it closes the feedback
loop in one round-trip.

### 5.3 Version Flag

`--version` prints `gematria X.Y.Z` and exits. When combined with
`--output json`, it emits `{"name": "gematria", "version": "X.Y.Z"}`. Agents
use this to verify which version they're working with.

### 5.4 No-Args Behavior

When the CLI is invoked with no arguments and stdin is a TTY (interactive
terminal), it prints a short usage hint to stderr and exits with code 0:

```
Usage: gematria [OPTIONS] [INPUT...] — try 'gematria --help'
```

This prevents the confusing "blank hanging terminal" experience for first-time
users. When stdin is piped (non-interactive), the tool reads from stdin as
normal per the requirements.

---

## 6. Error Design

### 6.1 Error Message Template

Every error message should follow this structure:

1. **What went wrong** — specific, not generic.
2. **What's valid** — the correct input space.
3. **How to fix** — a concrete suggestion or example.

### 6.2 Exit Codes

| Code | Meaning                                                          |
| ---- | ---------------------------------------------------------------- |
| 0    | Success                                                          |
| 1    | Input error (invalid character, unknown transliteration)         |
| 2    | CLI misuse (invalid flag, invalid env var value)                 |
| 3    | File error (word list not found, unreadable)                     |
| 4    | Partial success (stdin batch: some lines succeeded, some failed) |

Exit code 2 aligns with the GNU/Bash convention for incorrect command usage.
Exit code 4 only applies to stdin batch processing (see section 6.5).

Using distinct exit codes for different error classes lets agents branch on
the failure type without parsing the error message.

### 6.3 No Stdout on Error

This is stated in the requirements but bears repeating as a design principle:
on any error, stdout must be empty. This prevents agents from accidentally
parsing an error message as a valid result.

### 6.4 Stderr-Only Diagnostics

Warnings, suggestions, and "did you mean?" hints go to stderr. This keeps
stdout clean for piping regardless of whether the input was slightly wrong.

### 6.5 Stdin Batch Error Behavior

When processing multiple lines from stdin, the CLI continues processing all
lines by default. Valid lines produce results on stdout; invalid lines produce
per-line errors on stderr (including the line number). If any line fails, the
exit code is 4 (partial success). If all lines fail, the exit code is 1
(input error).

A `--fail-early` flag stops processing on the first error and exits
immediately with the appropriate error code (1, 2, or 3). This is useful for
agents that want strict all-or-nothing behavior.

### 6.6 Environment Variable Validation

Environment variables are validated **lazily** — only when the feature they
control is actually used. A stale `GEMATRIA_WORDLIST` pointing to a deleted
file does not block `gematria aleph`. It only fails when `--find` is used
without an overriding `--wordlist` flag. This matches the behavior of tools
like `$EDITOR` — validated on use, not on startup.

---

## 7. Reverse Lookup Design

### 7.1 Result Volume Control

Reverse lookup (`--find <value>`) can return many results. The CLI must give
both humans and agents control over volume:

- `--limit` / `-l` controls maximum results (default 20).
- When more results exist, the output must indicate this. For humans: a
  trailing message like `... and 42 more results (use --limit to show more)`.
  For JSON: a `hasMore: true` field and a `totalCount` field.

### 7.2 Word List Error Handling

The word list is the one external file dependency. Errors here need special
care:

- **File not found**: Name the path that was tried and where it came from
  (flag vs. env var).
- **File unreadable**: Name the path and the OS error (permission denied,
  etc.).
- **No word list provided**: Explain that `--find` requires a word list and
  show how to provide one (both flag and env var).

### 7.3 Enriched Results

When the word list includes transliteration and meaning columns (TSV format),
these appear in all output formats. For `json`, they become fields on each
result object. For `line` and `card`, they appear alongside the Hebrew word.
For `value`, they are omitted — `value` format only ever shows numbers.

---

## 8. Composability Patterns

The CLI should work naturally in these Unix compositions:

**Batch processing via stdin:**

```
cat hebrew-words.txt | gematria --output value
```

**Reverse lookup piped to further processing:**

```
gematria --find 26 --wordlist "${WORDLIST}" --output json | jq '.results[].word'
```

**Comparing values across systems:**

```
echo "אמת" | gematria -m hechrachi -o value
echo "אמת" | gematria -m gadol -o value
```

**Agent-style structured lookup:**

```
echo "aleph" | gematria --output json
```

**Batch processing with fail-early:**

```
cat hebrew-words.txt | gematria --output value --fail-early
```

These patterns work without special modes — Unix conventions and explicit
flags make them natural.

---

## 9. Accessibility

### 9.1 Color

- Support the `NO_COLOR` environment variable (see https://no-color.org/).
- Provide a `--no-color` flag as an explicit override.
- Never encode information through color alone — always pair with symbols
  or text.
- When color is disabled, output must remain fully legible and meaningful.

### 9.2 Unicode

- The tool inherently requires Unicode support (Hebrew characters). This is a
  prerequisite, not a design choice.
- Symbols used in output (`=`, `+`, `→`) have ASCII fallbacks where needed.
- RTL marks are Unicode control characters; they're invisible but essential
  for correct display in terminals that support bidi text.

