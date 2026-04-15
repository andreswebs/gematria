---
id: wor-ydfb
status: closed
deps: [wor-gdda, wor-j5l7, wor-usvq]
links: []
created: 2026-04-15T04:02:18Z
type: task
priority: 1
parent: wor-g7f2
tags: [letter-dictionary, low-level-api, task]
---
# Build alias map at init() and implement low-level API

Add or extend src/gematria.go with the package-level aliases map, the init() function that populates it, and three exported low-level API functions: Letter(), LetterByName(), and AtbashSubstitute(). Also add the unexported levenshtein() helper used by LetterByName() to generate suggestions.

## Spec References
- docs/specs/code-architecture.md §3.3 (Letter, LetterByName, AtbashSubstitute signatures and behavior)
- docs/specs/cli-design.md §3.3 (case-insensitive alias matching, Levenshtein suggestion threshold)
- docs/specs/requirements.md §2 (case-insensitive Latin input matching), §8 (error identifies invalid char and position)

## Implementation Details

### Alias Map and init()

    var aliases map[string]rune // populated by init()

    func init() {
        aliases = make(map[string]rune, 60) // ~2-3 aliases per letter * 27 letters
        for r, l := range letters {
            for _, alias := range l.Aliases {
                aliases[alias] = r // aliases already stored lowercase in letters map
            }
        }
    }

init() only populates the alias map. No other side effects. Aliases in the letters map are pre-lowercased so no strings.ToLower call is needed in init().

### Letter()

    func Letter(r rune) (Letter, error) {
        l, ok := letters[r]
        if !ok {
            return Letter{}, &InvalidCharError{Char: r}
        }
        return l, nil
    }

Note: Position and Input fields on InvalidCharError are set by the caller (Parse() in gematria.go) which has the full input context. Letter() sets only Char.

### LetterByName()

    func LetterByName(name string) (Letter, error) {
        key := strings.ToLower(name)
        if r, ok := aliases[key]; ok {
            return letters[r], nil
        }
        suggestions := suggestAliases(key)
        return Letter{}, &UnknownNameError{Name: name, Suggestions: suggestions}
    }

    func suggestAliases(input string) []string {
        maxDist := min(int(math.Ceil(float64(len(input))/2)), 2)
        var out []string
        seen := map[string]bool{}
        for alias := range aliases {
            d := levenshtein(input, alias)
            if d > 0 && d <= maxDist && !seen[alias] {
                out = append(out, alias)
                seen[alias] = true
            }
        }
        sort.Strings(out) // stable order for deterministic test assertions
        return out
    }

### AtbashSubstitute()

    func AtbashSubstitute(r rune) rune {
        if mirror, ok := atbashMirror[r]; ok {
            return mirror
        }
        return r // unknown rune: return unchanged
    }

Uses the atbashMirror map from systems.go.

### levenshtein() Helper

Classic dynamic programming implementation. Operates on runes (not bytes) to handle multi-byte Unicode correctly. O(m*n) time, O(min(m,n)) space using two-row rolling array.

    func levenshtein(a, b string) int {
        ra, rb := []rune(a), []rune(b)
        // standard DP with two-row rolling buffer
    }

Accepts and returns ints. Not exported — only used internally by suggestAliases().

## Acceptance Criteria

- [ ] Package-level `aliases map[string]rune` declared and populated by `init()`
- [ ] `init()` iterates `letters` map once; no other side effects; no I/O
- [ ] `Letter(r rune) (Letter, error)` returns correct Letter or `*InvalidCharError` for unknown rune
- [ ] `LetterByName(name string) (Letter, error)` is case-insensitive: "ALEPH", "Aleph", "aleph" all resolve
- [ ] `LetterByName` returns `*UnknownNameError` with non-empty Suggestions when edit distance <= min(ceil(len/2), 2)
- [ ] `LetterByName("shen")` returns suggestion containing "shin" (edit distance 1)
- [ ] `LetterByName("xyzzy")` returns empty Suggestions (no alias within distance 2)
- [ ] `AtbashSubstitute('א')` returns 'ת'; `AtbashSubstitute('ת')` returns 'א' (roundtrip)
- [ ] `levenshtein()` is unexported and operates on runes (not bytes)
- [ ] File contains no `os` import
- [ ] `go build ./...` from src/ passes


## Notes

**2026-04-15T11:18:22Z**

Implemented in src/gematria.go. Key deviation: the spec names the function 'Letter(r rune) (Letter, error)' but that conflicts with the 'Letter' struct type already declared in letters.go — Go does not allow a function and a type with the same identifier in the same package. Renamed to LookupLetter(r rune) (Letter, error). All other functions (LetterByName, AtbashSubstitute, levenshtein, suggestAliases) implemented as specified. Levenshtein uses two-row rolling DP on runes. Tests in gematria_test.go cover happy/error paths for all three exported functions.
