---
id: wor-gdda
status: closed
deps: []
links: []
created: 2026-04-15T04:01:10Z
type: task
priority: 1
parent: wor-g7f2
tags: [letter-dictionary, letter-type, task]
---
# Define Letter struct, System type, and 27-letter dictionary

Add src/letters.go with the Letter struct, System named string type with four constants, and the 27-entry letters map literal. This file contains only static data — no functions, no init(), no computation.

## Spec References
- docs/specs/requirements.md §1 (Letter Dictionary: 27 letters, names, meanings, aliases)
- docs/specs/code-architecture.md §3.1 (map[rune]Letter literal, alias list per letter), §3.2 (System type)

## Implementation Details

### System Type

    type System string

    const (
        Hechrachi System = "hechrachi"
        Gadol     System = "gadol"
        Siduri    System = "siduri"
        Atbash    System = "atbash"
    )

Named string type so constants compare directly against CLI flag values without conversion.

### Letter Struct

    type Letter struct {
        Char     rune
        Name     string   // canonical English name, e.g. "Aleph"
        Meaning  string   // pictographic meaning, e.g. "ox"
        Position int      // 1-22; sofit forms share position of normal form
        Aliases  []string // lowercase transliteration aliases, e.g. ["aleph","alef"]
        IsSofit  bool     // true for final forms
    }

System values are NOT stored here — they live in systems.go lookup tables.

### Dictionary Map

    var letters = map[rune]Letter{
        'א': {Char: 'א', Name: "Aleph", Meaning: "ox",        Position: 1,  Aliases: []string{"aleph", "alef"}, IsSofit: false},
        'ב': {Char: 'ב', Name: "Bet",   Meaning: "house",     Position: 2,  Aliases: []string{"bet", "beth", "vet"}, IsSofit: false},
        // ... all 22 standard letters ...
        'ך': {Char: 'ך', Name: "Kaf Sofit",  Meaning: "palm", Position: 11, Aliases: []string{"kaf sofit", "chaf sofit", "final kaf"}, IsSofit: true},
        // ... all 5 sofit forms ...
    }

All aliases are lowercase in the slice. The alias map (aliases map[string]rune) is built at init() in gematria.go by iterating this map — letters.go itself does not call strings.ToLower or contain init().

Unicode positions for standard letters: U+05D0 (aleph) through U+05EA (tav), noting gaps at U+05DA (kaf sofit), U+05DD (mem sofit), U+05DF (nun sofit), U+05E3 (pe sofit), U+05E5 (tsadi sofit). Sofit Unicode codepoints: U+05DA ך, U+05DD ם, U+05DF ן, U+05E3 ף, U+05E5 ץ.

## Acceptance Criteria

- [ ] `System` is a named string type with constants Hechrachi, Gadol, Siduri, Atbash matching CLI flag values exactly
- [ ] `Letter` struct has Char (rune), Name (string), Meaning (string), Position (int), Aliases ([]string), IsSofit (bool) — all exported, MixedCaps
- [ ] Package-level `letters` is `map[rune]Letter` initialized as a map literal (not make())
- [ ] All 22 standard Hebrew letters are present with correct Unicode rune literals
- [ ] All 5 sofit forms are present: ך ם ן ף ץ with IsSofit: true
- [ ] Sofit entries carry the same Position as their normal form (kaf/kaf-sofit both position 11, etc.)
- [ ] Every entry has at least one alias in the Aliases slice; all alias strings are lowercase
- [ ] File contains no `os` import, no `init()` function, no computation logic
- [ ] `go build ./...` from src/ passes


## Notes

**2026-04-15T11:09:27Z**

Implemented src/letters.go with System named string type (4 constants: hechrachi/gadol/siduri/atbash), Letter struct (Char/Name/Meaning/Position/Aliases/IsSofit), and 27-entry letters map literal (22 standard + 5 sofit forms). Tests in src/letters_test.go verify: map size=27, all runes present, sofit positions match base letters, all aliases lowercase and non-empty. Lint note: range over string directly (not []rune(string)) to satisfy staticcheck SA6003. Also fixed Makefile CMD_DIR from ./cmd/server to ./cmd/gematria.
