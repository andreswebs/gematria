# Gematria CLI - Requirements

## Introduction

Gematria is a command-line tool for computing Hebrew gematria values. Given a Hebrew letter, word, or phrase, it returns the numeric value along with the letter name, its traditional meaning, and a per-letter breakdown. The tool supports multiple gematria numbering systems and output formats.

The tool follows Unix conventions: it accepts input as a positional argument or from stdin, outputs structured results to stdout, and reports errors to stderr with non-zero exit codes. It is designed to be composable with other command-line tools via pipes and scripting.

A reverse lookup feature allows users to find words matching a given numeric value from a user-supplied word list, supporting the traditional gematria practice of discovering connections between words of equal value.

## Requirements

### 1. Letter Dictionary

**User Story**: As a gematria student, I want a complete reference of all Hebrew letters with their values, names, and meanings, so that I can look up any letter quickly.

**Acceptance Criteria**:

- The system shall store data for all 22 standard Hebrew letters (א through ת).
- The system shall store data for the 5 final (sofit) letter forms: ך, ם, ן, ף, ץ.
- The system shall store, for each letter: the Hebrew character, canonical English name, traditional pictographic meaning, and numeric values for all supported gematria systems.
- The system shall store multiple transliteration aliases for each letter (e.g., "vav", "waw", "vaw" for ו).

### 2. Input Handling

**User Story**: As a user, I want to provide input as either Hebrew characters or English transliterated names, so that I can use the tool regardless of my keyboard setup.

**Acceptance Criteria**:

- WHEN a positional argument is provided, the system shall use it as input.
- WHEN no positional argument is provided, the system shall read input from stdin, one word per line.
- WHEN the input is a Hebrew Unicode character (U+05D0–U+05EA, including sofit forms), the system shall match it directly.
- WHEN the input is a Latin string, the system shall perform case-insensitive matching against all known transliteration aliases.
- WHEN multiple space-separated Latin transliteration names are provided as positional arguments (e.g., `gematria aleph bet`), the system shall resolve each to its Hebrew letter and treat the sequence as a word (e.g., אב).
- The system shall accept single letters and multi-letter words/phrases as input.
- WHEN reading from stdin, the system shall process each line as a separate input and produce one result per line.
- WHEN the `--transliterate` / `-t` flag is provided, Latin input shall be interpreted as Hebrew words (phonetic transliteration) instead of letter aliases. See [transliteration.md](transliteration.md).

### 3. Gematria Systems

**User Story**: As a gematria practitioner, I want to choose between different numbering systems, so that I can study using the method relevant to my tradition.

**Acceptance Criteria**:

- The system shall support the following gematria systems:
  - **Hechrachi** (Standard): Aleph=1 through Tav=400, sofit forms carry same value as their normal form.
  - **Gadol**: Same as standard, but sofit forms carry extended values (ך=500, ם=600, ן=700, ף=800, ץ=900).
  - **Siduri** (Ordinal): Each letter is valued by its position 1–22.
  - **Atbash**: Substitution cipher where Aleph↔Tav, Bet↔Shin, etc. When selected as the mispar system, gematria values are computed on the substituted letters using Hechrachi values.
- The system shall default to Hechrachi.
- WHEN the `--atbash` flag is provided, the system shall output the Atbash-substituted letters alongside the normal output. This flag is independent of `--mispar` and can be combined with any gematria system.
- WHEN the `--mispar` / `-m` flag is provided, the system shall use the specified system.
- WHEN the `GEMATRIA_MISPAR` environment variable is set, the system shall use it as the default system.
- WHEN both the flag and environment variable are present, the flag shall take precedence.
- IF an unrecognized system name is provided, THEN the system shall exit with a non-zero exit code and an error message to stderr.

### 4. Output Formats

**User Story**: As a user, I want to control the output format so that I can use the tool for both human reading and scripting.

**Acceptance Criteria**:

- The system shall support the following output formats:
  - **line** (default): Single structured line, e.g., `Dalet (ד) = 4`.
  - **value**: Numeric value only, e.g., `4`.
  - **card**: Multi-line verbose output including letter, name, value, position, meaning, and gematria system.
  - **json**: Machine-readable JSON object with all fields.
- WHEN the input is a word, the **line** format shall include a per-letter breakdown, e.g., `דעת = 474 (ד=4 + ע=70 + ת=400)`.
- WHEN the input is a word, the **value** format shall output the sum only.
- WHEN the input is a word, the **card** format shall display a per-letter table plus the total sum.
- WHEN the input is a word, the **json** format shall include an array of per-letter objects and a total.
- WHEN the `--output` / `-o` flag is provided, the system shall use the specified format.
- WHEN the `GEMATRIA_OUTPUT` environment variable is set, the system shall use it as the default format.
- WHEN both the flag and environment variable are present, the flag shall take precedence.
- IF an unrecognized format name is provided, THEN the system shall exit with a non-zero exit code and an error message to stderr.

### 5. RTL Output

**User Story**: As a user viewing Hebrew in a terminal, I want correct right-to-left rendering, so that Hebrew text displays and copies correctly.

**Acceptance Criteria**:

- The system shall prefix Hebrew text in output with a Unicode RTL mark (U+200F).
- The system shall suffix Hebrew text in output with a Unicode LTR mark (U+200E).

### 6. Reverse Lookup

**User Story**: As a gematria student, I want to find words that share the same numeric value, so that I can discover connections between words.

**Acceptance Criteria**:

- WHEN the `--find <value>` flag is provided with a numeric value, the system shall return all words from the word list whose gematria sum equals that value.
- WHEN `--find` is used, a word list must be available via `--wordlist` flag or `GEMATRIA_WORDLIST` environment variable.
- IF `--find` is used without a word list, THEN the system shall exit with a non-zero exit code and an error message to stderr.
- The system shall compute word values using the currently selected gematria system (respecting `--mispar`).
- The output format of reverse lookup results shall respect the `--output` flag.
- The system shall paginate reverse lookup results, defaulting to 20 results per page.
- WHEN the `--limit` / `-l` flag is provided, the system shall use it as the maximum number of results to return.
- WHEN the `GEMATRIA_LIMIT` environment variable is set, the system shall use it as the default limit.
- WHEN both the flag and environment variable are present, the flag shall take precedence.
- WHEN more results exist beyond the limit, the system shall indicate that additional results are available.

### 7. Word List Format

**User Story**: As a user, I want to supply my own word list in a simple format, so that I can use my own curated vocabulary for reverse lookups.

**Acceptance Criteria**:

- The system shall accept a word list file via `--wordlist` flag or `GEMATRIA_WORDLIST` environment variable.
- WHEN both the flag and environment variable are present, the flag shall take precedence.
- The system shall support plain text format with one Hebrew word per line.
- The system shall support TSV format with columns: word (required), transliteration (optional), meaning (optional).
- WHEN a line contains a tab character, the system shall parse it as TSV.
- WHEN a line does not contain a tab character, the system shall parse it as a plain word.
- The system shall ignore blank lines.
- The system shall ignore lines starting with `#` as comments.
- WHERE the word list includes transliteration and meaning columns, the system shall include them in output for matching words.
- IF the word list file cannot be read, THEN the system shall exit with a non-zero exit code and an error message to stderr.

### 8. Error Handling

**User Story**: As a user, I want clear error messages on invalid input, so that I can correct my mistakes quickly.

**Acceptance Criteria**:

- IF the input contains any character that is not a recognized Hebrew letter, sofit form, or known transliteration alias, THEN the system shall exit with a non-zero exit code and an error message to stderr identifying the invalid character and its position.
- The system shall produce no output to stdout on error.
- The system shall report all errors to stderr.
- IF an invalid flag value is provided, THEN the system shall exit with a non-zero exit code and a usage hint to stderr.

### 9. Environment Variable Configuration

**User Story**: As a power user, I want to set defaults via environment variables, so that I don't have to repeat flags on every invocation.

**Acceptance Criteria**:

- The system shall support `GEMATRIA_MISPAR` to set the default gematria system.
- The system shall support `GEMATRIA_OUTPUT` to set the default output format.
- The system shall support `GEMATRIA_WORDLIST` to set the default word list file path.
- The system shall support `GEMATRIA_LIMIT` to set the default result limit for reverse lookups.
- WHEN a corresponding CLI flag is also provided, the flag shall take precedence over the environment variable.
- IF an environment variable contains an invalid value, THEN the system shall exit with a non-zero exit code and an error message to stderr.
