---
id: wor-j5l7
status: closed
deps: [wor-gdda]
links: []
created: 2026-04-15T04:01:35Z
type: task
priority: 1
parent: wor-g7f2
tags: [letter-dictionary, system-tables, task]
---
# Implement four gematria system value lookup tables

Add src/systems.go with four package-level map[rune]int lookup tables (one per gematria system), a systemValues aggregator map, and ValidSystems(). This is pure static data — no computation, no I/O.

## Spec References
- docs/specs/requirements.md §3 (four systems: Hechrachi, Gadol, Siduri, Atbash with exact values)
- docs/specs/code-architecture.md §3.2 (map[rune]int per system, pure data)

## Implementation Details

### Per-System Maps

    var hechrachi = map[rune]int{
        'א': 1,   'ב': 2,   'ג': 3,   'ד': 4,   'ה': 5,
        'ו': 6,   'ז': 7,   'ח': 8,   'ט': 9,   'י': 10,
        'כ': 20,  'ל': 30,  'מ': 40,  'נ': 50,  'ס': 60,
        'ע': 70,  'פ': 80,  'צ': 90,  'ק': 100, 'ר': 200,
        'ש': 300, 'ת': 400,
        // sofit same as normal
        'ך': 20, 'ם': 40, 'ן': 50, 'ף': 80, 'ץ': 90,
    }

    var gadol = map[rune]int{
        // standard letters identical to hechrachi
        // sofit extended:
        'ך': 500, 'ם': 600, 'ן': 700, 'ף': 800, 'ץ': 900,
    }

    var siduri = map[rune]int{
        // position 1-22; sofit shares position of normal form
        'א': 1,  'ב': 2,  ..., 'ת': 22,
        'ך': 11, 'ם': 13, 'ן': 14, 'ף': 17, 'ץ': 18,
    }

    var atbash = map[rune]int{
        // hechrachi value of mirrored letter (pos 1<->22, 2<->21, ... 11<->12)
        'א': 400, 'ב': 300, 'ג': 200, 'ד': 100, 'ה': 90,
        'ו': 80,  'ז': 70,  'ח': 60,  'ט': 50,  'י': 40,
        'כ': 30,  'ל': 20,  'מ': 10,  'נ': 9,   'ס': 8,
        'ע': 7,   'פ': 6,   'צ': 5,   'ק': 4,   'ר': 3,
        'ש': 2,   'ת': 1,
        // sofit mirrors through their normal form's pair
        'ך': 30, 'ם': 10, 'ן': 9, 'ף': 6, 'ץ': 5,
    }

### Aggregator and atbashMirror

    var systemValues = map[System]map[rune]int{
        Hechrachi: hechrachi,
        Gadol:     gadol,
        Siduri:    siduri,
        Atbash:    atbash,
    }

    // atbashMirror maps each rune to its Atbash pair rune (for AtbashSubstitute in gematria.go)
    var atbashMirror = map[rune]rune{
        'א': 'ת', 'ב': 'ש', 'ג': 'ר', 'ד': 'ק', 'ה': 'צ',
        'ו': 'פ', 'ז': 'ע', 'ח': 'ס', 'ט': 'נ', 'י': 'מ',
        'כ': 'ל', 'ל': 'כ', 'מ': 'י', 'נ': 'ט', 'ס': 'ח',
        'ע': 'ז', 'פ': 'ו', 'צ': 'ה', 'ק': 'ד', 'ר': 'ג',
        'ש': 'ב', 'ת': 'א',
        // sofit forms substitute to the normal form's pair
        'ך': 'ל', 'ם': 'י', 'ן': 'ט', 'ף': 'ו', 'ץ': 'ה',
    }

### ValidSystems

    func ValidSystems() []System {
        return []System{Hechrachi, Gadol, Siduri, Atbash}
    }

Returns a stable ordered slice used in InvalidSystemError.Valid and CLI flag validation messages.

## Acceptance Criteria

- [ ] Four unexported package-level maps defined: hechrachi, gadol, siduri, atbash — each map[rune]int covering all 27 runes
- [ ] Hechrachi: standard letters aleph=1 through tav=400; sofit same value as normal form
- [ ] Gadol: standard letters same as Hechrachi; sofit ך=500, ם=600, ן=700, ף=800, ץ=900
- [ ] Siduri: aleph=1 through tav=22 by position; sofit ך=11, ם=13, ן=14, ף=17, ץ=18
- [ ] Atbash: aleph=400, bet=300, ..., tav=1; sofit uses normal form's mirror pair value
- [ ] `systemValues map[System]map[rune]int` aggregates all four tables for O(1) dispatch
- [ ] `atbashMirror map[rune]rune` covers all 27 runes and maps bidirectionally (A->T and T->A)
- [ ] `ValidSystems() []System` returns all four systems in stable order
- [ ] File contains no `os` import, no `init()`, no computation
- [ ] `go build ./...` from src/ passes


## Notes

**2026-04-15T11:12:10Z**

Implemented src/systems.go with four unexported map[rune]int lookup tables (hechrachi, gadol, siduri, atbash) covering all 27 runes, systemValues aggregator for O(1) dispatch, atbashMirror map, and ValidSystems(). Tests in letters_test.go verify exact values for all tables, systemValues dispatch, atbashMirror bidirectionality, and full 27-rune coverage per table. All tests pass, make build clean.
