---
id: wor-mg99
status: closed
deps: [wor-gdda, wor-j5l7, wor-usvq, wor-ydfb]
links: []
created: 2026-04-15T04:02:46Z
type: task
priority: 2
parent: wor-g7f2
tags: [letter-dictionary, unit-tests, task]
---
# Write unit tests for letter dictionary and low-level API

Add src/letters_test.go (and optionally src/systems_test.go) with table-driven unit tests covering the letter dictionary, all four gematria system tables, typed error fields, and all three low-level API functions.

## Spec References
- docs/specs/code-architecture.md §7 (testing strategy: pure unit tests, table-driven, errors.As, no I/O)
- docs/specs/requirements.md §1 (all 27 letters), §3 (all 4 systems with specific values)

## Implementation Details

### Dictionary Completeness Test

Table-driven test iterating all 27 expected runes and asserting Letter() returns the correct Name, Meaning, Position, IsSofit, and non-empty Aliases:

    var letterTests = []struct {
        char     rune
        name     string
        meaning  string
        position int
        isSofit  bool
    }{
        {'א', "Aleph", "ox", 1, false},
        {'ך', "Kaf Sofit", "palm", 11, true},
        // ... all 27 entries
    }

    func TestLetterDictionary(t *testing.T) {
        for _, tc := range letterTests {
            t.Run(tc.name, func(t *testing.T) {
                got, err := Letter(tc.char)
                // assert fields
            })
        }
    }

### System Value Tests

Table-driven test covering all four systems for a representative sample of letters including at least: Aleph, Kaf, Kaf Sofit, Lamed, Tav, Tsadi Sofit. Verify Hechrachi, Gadol, Siduri, and Atbash values against the expected constants from the spec:

    func TestSystemValues(t *testing.T) {
        cases := []struct{
            r      rune
            system System
            want   int
        }{
            {'א', Hechrachi, 1},
            {'ת', Hechrachi, 400},
            {'ך', Hechrachi, 20},   // sofit = normal in Hechrachi
            {'ך', Gadol,     500},  // sofit extended in Gadol
            {'א', Siduri,    1},
            {'ת', Siduri,    22},
            {'ך', Siduri,    11},   // sofit shares position
            {'א', Atbash,    400},  // Aleph mirrors Tav
            {'ת', Atbash,    1},    // Tav mirrors Aleph
        }
        // ...
    }

### Alias Resolution Tests

Test LetterByName with multiple aliases per letter and case variants:

    func TestLetterByName(t *testing.T) {
        cases := []struct{ input string; wantChar rune }{
            {"aleph", 'א'}, {"ALEPH", 'א'}, {"Aleph", 'א'}, {"alef", 'א'},
            {"bet", 'ב'}, {"beth", 'ב'}, {"vet", 'ב'},
            {"shin", 'ש'}, {"sin", 'ש'},
            // sofit aliases
            {"kaf sofit", 'ך'}, {"final kaf", 'ך'},
        }
        // ...
    }

### Error Case Tests

Use errors.As to assert typed error fields:

    func TestLetterError(t *testing.T) {
        _, err := Letter('x') // ASCII, not Hebrew
        var charErr *InvalidCharError
        if !errors.As(err, &charErr) {
            t.Fatal("expected InvalidCharError")
        }
        if charErr.Char != 'x' {
            t.Errorf("Char = %q, want 'x'", charErr.Char)
        }
    }

    func TestLetterByNameError(t *testing.T) {
        _, err := LetterByName("shen")
        var nameErr *UnknownNameError
        if !errors.As(err, &nameErr) {
            t.Fatal("expected UnknownNameError")
        }
        // verify Suggestions contains "shin"
    }

    func TestLetterByNameNoSuggestion(t *testing.T) {
        _, err := LetterByName("xyzzy")
        var nameErr *UnknownNameError
        errors.As(err, &nameErr)
        if len(nameErr.Suggestions) != 0 {
            t.Errorf("expected no suggestions for garbage input, got %v", nameErr.Suggestions)
        }
    }

### AtbashSubstitute Roundtrip Tests

    func TestAtbashSubstitute(t *testing.T) {
        pairs := [][2]rune{{'א','ת'}, {'ב','ש'}, {'כ','ל'}, {'ך','ל'}}
        for _, p := range pairs {
            if AtbashSubstitute(p[0]) != p[1] { ... }
            if AtbashSubstitute(p[1]) != p[0] { ... } // roundtrip
        }
    }

## Acceptance Criteria

- [ ] Table-driven test covers all 27 letters via `Letter()`: Name, Meaning, Position, IsSofit, and non-empty Aliases all verified
- [ ] System value test covers all 4 systems for at least 9 representative cases including sofit extended values
- [ ] Alias tests cover at least 2 aliases per letter and case-insensitive variants (upper, mixed, lower)
- [ ] Sofit letter aliases resolve correctly (e.g., "kaf sofit", "final kaf" → 'ך')
- [ ] `errors.As` used to assert `*InvalidCharError` fields (Char) on unknown rune input
- [ ] `errors.As` used to assert `*UnknownNameError` Suggestions on near-match input ("shen" → ["shin"])
- [ ] `errors.As` used to assert empty Suggestions on garbage input ("xyzzy" → [])
- [ ] AtbashSubstitute roundtrip test passes for at least 5 pairs including one sofit
- [ ] `make test` passes with 0 failures (no I/O, no file system access in any test)
- [ ] All test functions use `t.Run()` subtests for organized output


## Notes

**2026-04-15T12:27:14Z**

Added four table-driven t.Run tests to satisfy acceptance criteria: TestLookupLetterAllEntries (all 27 letters, Name/Meaning/Position/IsSofit/Aliases via LookupLetter), TestLetterByNameAliasVariants (2+ aliases per letter, sofit aliases, case variants), TestSystemValuesRepresentative (all 4 systems × 19 cases via systemValues dispatch, including sofit extended values), TestAtbashSubstitutePairs (11 specific pairs including 5 sofit forms). All existing tests were already green; new tests added to gematria_test.go and letters_test.go.
