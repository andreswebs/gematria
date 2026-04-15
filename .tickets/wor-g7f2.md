---
id: wor-g7f2
status: closed
deps: []
links: []
created: 2026-04-15T04:00:50Z
type: epic
priority: 1
tags: [letter-dictionary, epic]
---
# Letter Dictionary, Gematria Systems, and Low-Level API

Implement the root package data foundation: the 27-letter Hebrew letter dictionary, four gematria system lookup tables, typed error types, and the low-level API functions (Letter, LetterByName, AtbashSubstitute). This is the prerequisite for all other features — nothing can be computed or looked up without this layer.

## Purpose

The root package (github.com/andreswebs/gematria) is a pure domain library with no I/O, no os imports, and no external dependencies beyond the standard library. This epic lays the entire data and API foundation on which Parse, Compute, FindByValue, and all CLI formatting depend.

## Spec References

- docs/specs/requirements.md §1 (Letter Dictionary), §3 (Gematria Systems), §8 (Error Handling)
- docs/specs/code-architecture.md §3.1 (Letter Dictionary), §3.2 (Gematria Systems), §3.3 (Low-Level API), §3.7 (Typed Errors)

## Files Affected

- src/letters.go — Letter struct, System type, 27-letter map literal
- src/systems.go — Four system lookup tables, ValidSystems(), systemValues aggregator
- src/errors.go — InvalidCharError, UnknownNameError, InvalidSystemError
- src/gematria.go — aliases map, init(), Letter(), LetterByName(), AtbashSubstitute(), levenshtein() helper
- src/letters_test.go — Unit tests for dictionary, systems, errors, and low-level API

## Task List

1. [letter-type] Define Letter struct, System type, and 27-letter dictionary (letters.go)
2. [system-tables] Implement four gematria system value lookup tables (systems.go)
3. [typed-errors] Define typed error types (errors.go)
4. [low-level-api] Build alias map at init() and implement Letter(), LetterByName(), AtbashSubstitute() (gematria.go)
5. [unit-tests] Write unit tests for the letter dictionary and low-level API

## Key Design Decisions

### System Type

System is a named string type (not iota int), enabling readable constants and direct string comparison with flag values:

    type System string

    const (
        Hechrachi System = "hechrachi"
        Gadol     System = "gadol"
        Siduri    System = "siduri"
        Atbash    System = "atbash"
    )

This matches the enum values used in CLI flags (--mispar hechrachi) with no conversion layer and makes ValidSystems() usable directly in flag validation error messages.

### Letter Struct

Fields derived from spec requirements (code-architecture.md §3.1):
- Char rune — the Hebrew Unicode character (primary key in the dictionary map)
- Name string — canonical English name (e.g., "Aleph")
- Meaning string — traditional pictographic meaning (e.g., "ox")
- Position int — 1-22; sofit forms share the position of their normal form
- Aliases []string — all transliteration aliases in lowercase (e.g., ["aleph", "alef"])
- IsSofit bool — true for final forms (ך ם ן ף ץ)

System values per letter are NOT stored on the Letter struct; they live in systems.go lookup tables. This keeps letter data static and separates it from computation concerns.

### Dictionary Map

A package-level map[rune]Letter literal with 27 entries (22 standard + 5 sofit). Map literals are compile-time type-checked. Standard letter Unicode range: U+05D0 (aleph) through U+05EA (tav). Sofit forms: U+05DA (kaf sofit ך), U+05DD (mem sofit ם), U+05DF (nun sofit ן), U+05E3 (pe sofit ף), U+05E5 (tsadi sofit ץ).

### System Lookup Tables

Four package-level map[rune]int variables, one per system, aggregated in systemValues map[System]map[rune]int for O(1) dispatch. All tables cover all 27 runes.

Hechrachi (standard): classical values 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 200, 300, 400; sofit same as normal form.

Gadol (great): same as Hechrachi for standard letters; sofit extended: ך=500, ם=600, ן=700, ף=800, ץ=900.

Siduri (ordinal): letters valued 1-22 by position; sofit shares position of normal form (kaf sofit=11, mem sofit=13, nun sofit=14, pe sofit=17, tsadi sofit=18).

Atbash (cipher): Hechrachi value of the mirrored letter. Mirror: position 1 <-> 22, 2 <-> 21, 3 <-> 20, 4 <-> 19, 5 <-> 18, 6 <-> 17, 7 <-> 16, 8 <-> 15, 9 <-> 14, 10 <-> 13, 11 <-> 12. So Aleph Atbash value = 400, Bet = 300, ..., Tav = 1. Sofit forms use their normal form's mirror pair.

### Alias Map and Levenshtein Suggestions

A package-level aliases map[string]rune is built at init() time by iterating the letters map and lowercasing each alias. LetterByName() lower-cases input before lookup. On miss, it computes Levenshtein distance against all alias keys; suggestions are included when distance <= min(ceil(len(input)/2), 2). Candidates with distance 0 are skipped (exact match would have succeeded). The Suggestions slice is populated in UnknownNameError and formatted by the CLI layer — the root package never prints.

### No os Imports

The root package must never import os. All errors are returned as typed values; no log.Fatal, no os.Exit, no file I/O.

