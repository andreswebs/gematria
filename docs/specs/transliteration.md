# Word Transliteration

> **Status**: Ready

This document specifies the word transliteration feature: typing Hebrew words
in the Latin alphabet and having them recognized as Hebrew text for gematria
computation. It complements [requirements.md](requirements.md),
[cli-design.md](cli-design.md), and [code-architecture.md](code-architecture.md).

---

## 1. Motivation

The CLI's existing Latin-input mode treats each space-separated token as a
**letter alias** (e.g., `aleph` → א). This works for spelling out words
letter-by-letter but does not let users type Hebrew words phonetically:

- `gematria gadol` currently fails — `gadol` is not a letter alias, and the
  Levenshtein suggestion (`gamel`) is misleading.
- Users who don't have a Hebrew keyboard or don't know letter names cannot
  type Hebrew words like שלום, אמת, or גדול.

This feature adds opt-in word transliteration so a user can type `gematria -t shalom`
and have it computed as שלום.

---

## 2. Scope

**In scope:**

- A new opt-in `--transliterate` / `-t` flag that switches Latin input
  parsing from letter-alias mode to word-transliteration mode.
- Multi-scheme support via `--scheme` and `GEMATRIA_SCHEME`.
- Two schemes in v1: `academic` (default) and `israeli`.
- Per-scheme rules for consonant mappings and vowel handling.
- A new typed error `UnknownWordError` for transliteration failures.
- New root-package functions: `Transliterate`, `ComputeTransliterated`,
  `ComputeFromLetters`.

**Out of scope (v1):**

- Hebrew dictionary lookups (the tool transliterates letter-by-letter; it
  does not recognize that `bet` could mean בית).
- Suggestion sources for unknown words (suggestions may be empty in v1).
- Additional schemes beyond `academic` and `israeli`.
- Round-trip support (Hebrew → Latin transliteration). One-way only.
- Niqqud (vowel pointing) handling. Output is unpointed Hebrew.

---

## 3. User-Facing Design

### 3.1 Flags

| Flag                           | Short | Default    | Notes                               |
| ------------------------------ | ----- | ---------- | ----------------------------------- |
| `--transliterate`              | `-t`  | off        | Opt-in. Switches Latin input mode.  |
| `--scheme {academic\|israeli}` | —     | `academic` | Selects the transliteration scheme. |

### 3.2 Environment Variables

| Variable          | Equivalent flag | Notes                                 |
| ----------------- | --------------- | ------------------------------------- |
| `GEMATRIA_SCHEME` | `--scheme`      | Sets default scheme. Lazy validation. |

Precedence follows the existing pattern: explicit flag > env var > built-in
default. `GEMATRIA_SCHEME` is only validated when `--transliterate` is active
(consistent with `GEMATRIA_LIMIT` lazy validation).

### 3.3 Mode Switching

`--transliterate` switches Latin input parsing **globally** for the
invocation:

- **Without `-t`** (current behavior): every Latin token is looked up as a
  letter alias. `gematria aleph bet` → אב.
- **With `-t`**: every Latin token is parsed as a Hebrew word using the
  selected scheme. `gematria -t shalom emet` → שלום (376), אמת (441).

There is no fallback. With `-t` set, `gematria -t aleph` does NOT resolve as
the letter Aleph — it is transliterated letter-by-letter as a-l-e-p-h. If a
user wants letter aliases, they don't pass `-t`. If they want word
transliteration, they do.

Hebrew Unicode input works in both modes — `-t` only changes how Latin
tokens are interpreted.

### 3.4 Multi-Token and Compound Input

Each space-separated positional argument produces a separate computation, same
as existing behavior:

```sh
gematria -t shalom emet     # Two computations: שלום=376, אמת=441
```

A quoted string passed as one positional argument is treated as a single
computation. Internal spaces are word separators within that computation:
each space-delimited part is transliterated independently (so sofit and
word-initial vowel rules apply at the boundaries of each part), and the
resulting letter sequences are concatenated into one Result:

```sh
gematria -t "shalom emet"   # One computation: שלוםאמת (combined)
```

For batch mode (stdin), each line is one transliterated word, same as the
existing batch dispatch.

### 3.5 Compatibility with Other Flags

`--transliterate` composes orthogonally with all other flags:

- `--mispar` — picks the gematria system applied to the resolved Hebrew letters.
- `--output` — same four formats; `Result.Scheme` appears in JSON when set.
- `--atbash` — Atbash mappings are applied to the resolved Hebrew letters
  after transliteration completes. The display shows the Latin input, the
  transliterated Hebrew, and the Atbash substitution.
- `--find` — reverse lookup is unaffected; the value is numeric.
- `--no-color`, `--fail-early`, etc. — unchanged.

---

## 4. Schemes

### 4.1 `academic` Scheme

The academic scheme is **strictly consonantal**, inspired by SBL (Society of
Biblical Literature) conventions with ASCII fallbacks replacing diacritics.
Vowels are silently dropped; the output is biblical-style consonantal Hebrew.

**Default justification**: more conservative — makes fewer assumptions about
implicit vowels. A user typing `shalom` expecting biblical שלם gets that.

**Capitalization**: case-insensitive. All input is lowercased before lookup.
No uppercase distinctions.

**Rationale for ambiguous choices:**
- `ch` and `kh` both → ח (Het): in academic/scholarly contexts, `ch` and
  `kh` both denote the pharyngeal/velar fricative (IPA /χ/ or /ħ/), which
  corresponds to ח (Het), the historically distinct guttural. This aligns with
  SBL and ISO 259 usage.
- `s` → ס (Samekh): the plain sibilant maps to Samekh; שׁ (Shin) requires
  the explicit `sh` digraph.
- `t` → ת (Tav): the plain dental stop maps to Tav; ט (Tet) is not accessible
  via ASCII in this scheme and must be entered as Hebrew Unicode directly.
- Ayin (ע) is not accessible via ASCII (apostrophe `'` is reserved for Aleph);
  use Hebrew Unicode directly.

#### 4.1.1 Multi-Character Sequences (checked left-to-right, greedy)

| Input | Hebrew | Name  | Hechrachi value |
| ----- | ------ | ----- | --------------- |
| `sh`  | ש      | Shin  | 300             |
| `kh`  | ח      | Het   | 8               |
| `ch`  | ח      | Het   | 8               |
| `ts`  | צ      | Tsade | 90              |
| `tz`  | צ      | Tsade | 90              |
| `ph`  | פ      | Pe    | 80              |

Multi-character sequences are matched before single characters. Matching is
greedy and left-to-right: at each position the longest matching sequence wins.

#### 4.1.2 Single Characters

| Input | Hebrew | Name   | Hechrachi value | Notes                          |
| ----- | ------ | ------ | --------------- | ------------------------------ |
| `'`   | א      | Aleph  | 1               | apostrophe = explicit glottal  |
| `b`   | ב      | Bet    | 2               |                                |
| `g`   | ג      | Gimel  | 3               |                                |
| `d`   | ד      | Dalet  | 4               |                                |
| `h`   | ה      | He     | 5               |                                |
| `v`   | ו      | Vav    | 6               |                                |
| `w`   | ו      | Vav    | 6               | alternate ASCII spelling       |
| `z`   | ז      | Zayin  | 7               |                                |
| `x`   | ח      | Het    | 8               | ASCII fallback (x for guttural)|
| `y`   | י      | Yod    | 10              |                                |
| `k`   | כ      | Kaf    | 20              |                                |
| `l`   | ל      | Lamed  | 30              |                                |
| `m`   | מ      | Mem    | 40              |                                |
| `n`   | נ      | Nun    | 50              |                                |
| `s`   | ס      | Samekh | 60              |                                |
| `p`   | פ      | Pe     | 80              |                                |
| `f`   | פ      | Pe     | 80              | fricative (bet without dagesh) |
| `q`   | ק      | Qof    | 100             |                                |
| `r`   | ר      | Resh   | 200             |                                |
| `t`   | ת      | Tav    | 400             |                                |

#### 4.1.3 Vowels (silently dropped)

| Chars           | Action | Rationale                               |
| --------------- | ------ | --------------------------------------- |
| `a`, `e`, `i`, `o`, `u` | drop   | Vowels have no consonantal representation in academic Hebrew |

A token consisting entirely of vowels produces an empty letter sequence and
returns `*UnknownWordError`.

#### 4.1.4 Examples

| Input      | Resolved Hebrew | Notes                           |
| ---------- | --------------- | ------------------------------- |
| `shalom`   | שלם             | sh→ש, a→drop, l→ל, o→drop, m→מ |
| `gadol`    | גדל             | g→ג, a→drop, d→ד, o→drop, l→ל  |
| `emet`     | מת              | e→drop, m→מ, e→drop, t→ת       |
| `bereshit` | ברשׁת            | b→ב, e→drop, r→ר, sh→ש, i→drop, t→ת |
| `'emet`    | אמת             | '→א, m→מ, e→drop, t→ת          |

---

### 4.2 `israeli` Scheme

The israeli scheme uses **modern Israeli phonetic conventions** with matres
lectionis. Vowels are mapped to ו, י, א, or ה per the rules below. Output
matches how a modern Israeli would spell the word in everyday Hebrew writing.

**Capitalization**: case-insensitive. All input is lowercased before lookup.

**Rationale for ambiguous choices:**
- `ch` → ח (Het): same as academic. Both ח and כ are pronounced identically
  in modern Israeli ("kh"), but `ch` as a digraph is conventional for this
  sound. Since the tool uses `kh` for the same letter, `ch` is a synonym.
- `s` → ס (Samekh): consistent with academic to avoid surprise for users
  switching between schemes.
- Vowel treatment: modern Israeli Hebrew frequently uses matres lectionis
  (ו for o/u, י for i) to represent vowels. Word-initial vowels receive an
  aleph vowel carrier. Word-final short vowels (a, e) use ה. Medial short
  vowels (a, e) are silently dropped, consistent with standard Israeli spelling
  of unaccented syllables.

#### 4.2.1 Multi-Character Sequences (checked left-to-right, greedy)

Same as academic:

| Input | Hebrew | Name  | Hechrachi value |
| ----- | ------ | ----- | --------------- |
| `sh`  | ש      | Shin  | 300             |
| `kh`  | ח      | Het   | 8               |
| `ch`  | ח      | Het   | 8               |
| `ts`  | צ      | Tsade | 90              |
| `tz`  | צ      | Tsade | 90              |
| `ph`  | פ      | Pe    | 80              |

#### 4.2.2 Single Consonants

Identical to the academic consonant table (Section 4.1.2). All 19 consonant
mappings are shared between both schemes.

#### 4.2.3 Vowel-to-Mater Rules

Vowels are processed with awareness of their position within the word-part
being transliterated. "Word-part" here means one space-delimited segment of
the input (applies when a quoted compound input is split on internal spaces).

| Vowel | Position         | Hebrew | Name  | Rationale                          |
| ----- | ---------------- | ------ | ----- | ---------------------------------- |
| `a`   | word-initial     | א      | Aleph | vowel carrier for word-initial 'a' |
| `a`   | medial           | (drop) | —     | short medial 'a' unwritten in Israeli spelling |
| `a`   | word-final       | ה      | He    | mater lectionis for final 'a' sound (e.g., "shira" → שירה) |
| `e`   | word-initial     | א      | Aleph | same carrier as 'a' (both short front vowels) |
| `e`   | medial           | (drop) | —     | short medial 'e' unwritten in Israeli spelling |
| `e`   | word-final       | ה      | He    | mater for final 'e' (e.g., "yafe" → יפה) |
| `i`   | any              | י      | Yod   | mater lectionis for /i/ throughout |
| `o`   | word-initial     | או     | Aleph+Vav | aleph carrier + vav mater (e.g., "or" → אור) |
| `o`   | non-initial      | ו      | Vav   | mater lectionis for /o/            |
| `u`   | word-initial     | או     | Aleph+Vav | same carrier rule as 'o'       |
| `u`   | non-initial      | ו      | Vav   | mater lectionis for /u/            |

**Position definitions:**
- *word-initial*: the vowel character occupies the first position in the
  word-part, before any consonant has been emitted for this part.
- *word-final*: the vowel is the last character in the word-part, or all
  remaining characters after it are also vowels that will be dropped or
  handled as final vowels.
- *medial*: any other position (at least one consonant has been emitted for
  this part and at least one consonant follows).

**Implementation note:** "word-final" detection requires a lookahead pass.
The simplest implementation: after processing all characters left-to-right,
check whether the last emitted letter was a vowel-derived mater; if not, and
if the last character(s) of the input were `a` or `e`, append ה at the end.

#### 4.2.4 Examples

| Input      | Resolved Hebrew | Notes                                                   |
| ---------- | --------------- | ------------------------------------------------------- |
| `shalom`   | שלום            | sh→ש, a(medial)→drop, l→ל, o(non-initial)→ו, m→מ       |
| `gadol`    | גדול            | g→ג, a(medial)→drop, d→ד, o(non-initial)→ו, l→ל        |
| `emet`     | אמת             | e(initial)→א, m→מ, e(medial)→drop, t→ת                 |
| `or`       | אור             | o(initial)→או, r→ר → אור                               |
| `shira`    | שירה            | sh→ש, i→י, r→ר, a(final)→ה → שירה                      |
| `yafe`     | יפה             | y→י, a(medial)→drop, f→פ, e(final)→ה → יפה             |
| `david`    | דויד            | d→ד, a(medial)→drop, v→ו, i→י, d→ד → דויד              |

---

### 4.3 Scheme Differences

| Latin input | `academic` (consonantal) | `israeli` (with maters) | Notes                     |
| ----------- | ------------------------ | ----------------------- | ------------------------- |
| `shalom`    | שלם                      | שלום                    | 'o' kept as ו in Israeli  |
| `gadol`     | גדל                      | גדול                    | 'o' kept as ו in Israeli  |
| `emet`      | מת                       | אמת                     | 'e' drops in academic; becomes aleph carrier in Israeli |
| `'emet`     | אמת                      | —                       | explicit glottal in academic gives aleph |
| `or`        | ר                        | אור                     | 'o' dropped in academic; aleph+vav in Israeli |
| `shira`     | שר                       | שירה                    | 'i' and final 'a' kept in Israeli |

---

### 4.4 Strict Per-Scheme Mapping

Each scheme has documented, deterministic mappings. Ambiguous English
combinations resolve to one canonical letter per scheme:

| Sequence | `academic` | `israeli` | Rationale                                   |
| -------- | ---------- | --------- | ------------------------------------------- |
| `kh`     | ח (Het)    | ח (Het)   | Standard scholarly and colloquial "kh" = Het |
| `ch`     | ח (Het)    | ח (Het)   | Synonym for `kh` in both schemes            |
| `s`      | ס (Samekh) | ס (Samekh)| Plain sibilant = Samekh; Shin needs `sh`    |
| `sh`     | ש (Shin)   | ש (Shin)  | Digraph unambiguously = Shin                |
| `ts`     | צ (Tsade)  | צ (Tsade) | Standard affricate = Tsade                  |
| `tz`     | צ (Tsade)  | צ (Tsade) | Alternate spelling of same                  |
| `h`      | ה (He)     | ה (He)    | Single h = He (not Het; Het uses `kh`/`ch`) |

Users needing a specific Hebrew letter that has no unambiguous ASCII
representation should use Hebrew Unicode directly or per-letter aliases
(without `-t`).

---

### 4.5 Sofit (Final Letter) Forms

Hebrew has five letters with distinct final forms used at the end of a word:

| Base | Sofit | Name       |
| ---- | ----- | ---------- |
| כ    | ך     | Kaf sofit  |
| מ    | ם     | Mem sofit  |
| נ    | ן     | Nun sofit  |
| פ    | ף     | Pe sofit   |
| צ    | ץ     | Tsade sofit|

**Rule:** when one of these five letters is the **last letter** in the
translated sequence for a word-part, its sofit form is used instead.

A "word-part" is one space-delimited segment of the input. For a single-token
input like `shalom`, the final letter מ becomes ם. For a quoted compound input
`"shalom emet"`, each part is processed separately: שלום ends with מ→ם, and
אמת ends with ת (no sofit for tav).

This rule applies to both `academic` and `israeli` schemes.

**Sofit substitution happens after all other letter-resolution rules.** The
algorithm:
1. Transliterate the full word-part into a letter sequence.
2. Inspect the last letter in the sequence.
3. If it is one of {כ, מ, נ, פ, צ}, replace it with the corresponding sofit.

---

### 4.6 Capitalization

**Both schemes:** input is converted to lowercase before any lookup. All
multi-character sequences and single-character mappings are defined on
lowercase inputs only.

**Rationale:** consistent with the existing letter-alias mode (which is
case-insensitive). Academic transliteration sometimes uses uppercase to
distinguish emphatic consonants (e.g., ṭ vs t), but those distinctions are
handled by diacritics, not case, and this tool uses ASCII fallbacks (multi-char
sequences) instead. Introducing case sensitivity would create surprising
behaviour for users typing `Shalom` vs `shalom`.

---

### 4.7 Numeric and Punctuation Handling

Input characters that are not in the mapping tables (Section 4.1 and 4.2) and
are not vowels subject to the drop rule are treated as errors.

| Character class        | Action           | Notes                                      |
| ---------------------- | ---------------- | ------------------------------------------ |
| Digits (`0`–`9`)       | `UnknownWordError` | Numeric characters have no Hebrew mapping  |
| Hyphen, underscore     | `UnknownWordError` | Not treated as word separators within a token |
| Punctuation (except `'`) | `UnknownWordError` | Apostrophe `'` is reserved for Aleph       |
| Whitespace             | word separator   | Splits a single input into word-parts (see §3.4); not an error within a quoted compound arg |

If after applying all drop and mapping rules the resulting letter sequence for
a word-part is **empty** (e.g., the entire input was vowels in the academic
scheme), return `*UnknownWordError` for that word-part.

---

## 5. Errors

### 5.1 New Typed Error: `UnknownWordError`

When transliteration cannot resolve a token to any Hebrew letters (e.g.,
input contains characters not in the scheme, or the input is empty after
stripping unrecognizable characters), the root package returns
`*UnknownWordError`.

Fields:

- `Input` — the original Latin input that failed.
- `Scheme` — which scheme was attempted.
- `Position` — token index (for multi-token input).
- `Suggestions` — near-match suggestions from a future suggestion source.
  Empty in v1.

### 5.2 Error Output

The CLI renders `UnknownWordError` in plain text or JSON depending on
`--output`, consistent with existing error rendering:

```json
{
  "error": "input \"qzxw\" cannot be transliterated in scheme \"academic\"",
  "invalid_input": "qzxw",
  "scheme": "academic",
  "position": 0,
  "suggestions": []
}
```

Exit code: `1` (input error), same as `UnknownNameError`.

### 5.3 Invalid Scheme Value

`--scheme bogus` is rejected at flag-parse time with exit code `2` and a
plain-text message listing valid schemes — same pattern as invalid `--mispar`
or `--output`.

`GEMATRIA_SCHEME=bogus` is validated lazily — only when `--transliterate` is
active. Same pattern as `GEMATRIA_LIMIT` (validated only when `--find` is
active).

---

## 6. Code Architecture

### 6.1 New Public API in Root Package

```go
// Scheme is the transliteration scheme to use.
type Scheme string

const (
    SchemeAcademic Scheme = "academic"
    SchemeIsraeli  Scheme = "israeli"
)

// Transliterate parses input as a Hebrew word using the given scheme,
// returning the resolved sequence of letters.
// Returns *UnknownWordError if the input cannot be resolved.
func Transliterate(input string, scheme Scheme) ([]Letter, error)

// ComputeFromLetters computes a Result from a pre-resolved letter sequence.
// Provided as a primitive for consumers that want to inspect or modify
// letters before computation.
func ComputeFromLetters(input string, letters []Letter, system System) (Result, error)

// ComputeTransliterated transliterates input using scheme, then computes
// gematria values using system. Convenience wrapper around Transliterate
// + ComputeFromLetters.
func ComputeTransliterated(input string, system System, scheme Scheme) (Result, error)
```

The existing `Compute` is refactored to call `ComputeFromLetters` internally.
Three exported functions, single source of truth for the computation step.

### 6.2 New Result Field

```go
type Result struct {
    Input   string
    System  System
    Scheme  Scheme  // NEW: empty when transliteration was not used
    Total   int
    Letters []LetterResult
}
```

JSON output uses `omitempty` so the field is absent for non-transliterated
results.

### 6.3 Scheme Data

Scheme mapping tables live as Go map literals in separate files:
- `src/transliteration_academic.go` — academic scheme tables
- `src/transliteration_israeli.go` — israeli scheme tables
- `src/transliteration.go` — shared types, Transliterate, ComputeTransliterated, ComputeFromLetters

### 6.4 CLI Integration

`internal/cli/config.go` gains:

- `Transliterate bool` — set by `-t` / `--transliterate`.
- `Scheme string` — resolved from flag → env var → default `academic`.

`internal/cli/run.go` dispatches:

- If `cfg.Transliterate`, call `gematria.ComputeTransliterated(input, cfg.Mispar, cfg.Scheme)`.
- Otherwise, call `gematria.Compute(input, cfg.Mispar)` (existing behavior).

---

## 7. Open Questions (Deferred to v2)

### Q3: Suggestion source for `UnknownWordError`

In v1, `UnknownWordError.Suggestions` will be empty. Future work could draw
from:

- The user's word list (when one is loaded for `--find`).
- A static common-gematria-vocabulary list bundled with the binary.
- Levenshtein matches against successfully-transliterated words from a
  session history (out of scope for a stateless CLI).

This is deferred to v2. The field is present in the error struct and JSON
output but will always be `[]` until a suggestion source is implemented.

---

*Previously open questions Q1, Q2, Q4, Q5, Q6, Q7 were resolved and
incorporated into Section 4 as part of spec task gc-tw8e. Q8 (migration) is
not applicable — this feature is purely additive. Q9 (--help documentation)
is addressed in epic task gc-evoy.*
