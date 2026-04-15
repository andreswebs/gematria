---
id: gc-h6f3
status: closed
deps: [gc-0mhe, gc-d6hf]
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, compute-refactor, task]
---
# Refactor Compute API: ComputeFromLetters, ComputeTransliterated, Result.Scheme

Refactor the existing `Compute` function to expose a primitive `ComputeFromLetters` that other compute functions can share, then add `ComputeTransliterated` as a convenience wrapper. Add a `Scheme` field to `Result`.

## Files affected

- `src/result.go` — add Scheme field
- `src/gematria.go` — refactor Compute, add ComputeFromLetters and ComputeTransliterated

## Refactor Compute

Current (in `src/gematria.go`):
```go
func Compute(input string, system System) (Result, error) {
    // validate system, parse, build LetterResult, sum, return Result
}
```

Refactored:
```go
// ComputeFromLetters computes a Result from a pre-resolved letter sequence.
// Provided as a primitive for consumers that want to inspect or modify
// letters before computation.
func ComputeFromLetters(input string, letters []Letter, system System) (Result, error) {
    if err := validateSystem(system); err != nil { return Result{}, err }
    table := systemValues[system]
    letterResults := make([]LetterResult, len(letters))
    total := 0
    for i, l := range letters {
        val := table[l.Char]
        letterResults[i] = LetterResult{Letter: l, Value: val}
        total += val
    }
    return Result{Input: input, System: system, Total: total, Letters: letterResults}, nil
}

// Compute (refactored to thin wrapper)
func Compute(input string, system System) (Result, error) {
    ltrs, err := Parse(input)
    if err != nil { return Result{}, err }
    return ComputeFromLetters(input, ltrs, system)
}

// ComputeTransliterated transliterates input using scheme, then computes
// gematria values using system. Sets Result.Scheme.
func ComputeTransliterated(input string, system System, scheme Scheme) (Result, error) {
    ltrs, err := Transliterate(input, scheme)
    if err != nil { return Result{}, err }
    r, err := ComputeFromLetters(input, ltrs, system)
    if err != nil { return Result{}, err }
    r.Scheme = scheme
    return r, nil
}
```

`ComputeFromLetters` does NOT set `Scheme` (leaves zero value). `ComputeTransliterated` sets it explicitly after computation. This keeps the primitive scheme-agnostic.

## Result.Scheme field

```go
type Result struct {
    Input   string
    System  System
    Scheme  Scheme         // NEW: empty when not transliterated
    Total   int
    Letters []LetterResult
}
```

## Atbash special handling

Preserve the existing approach: `systemValues[Atbash]` already maps each rune to its mirror's hechrachi value, so no AtbashSubstitute call is needed at compute time. Both `Compute` and `ComputeTransliterated` work identically with Atbash.

## Backward compatibility

`Compute(input, system)` MUST return identical results before and after the refactor. Existing tests in `src/gematria_test.go` should pass unchanged. Result.Scheme has zero value (empty string) for non-transliterated results, which JSON output omits via `omitempty` (Task 8).

## Spec References

- docs/specs/transliteration.md §6 (Code Architecture), §6.2 (Result field)
- docs/specs/code-architecture.md §3.4 (additions)
- See parent epic.

## Acceptance Criteria

- [ ] `Result` struct has new `Scheme Scheme` field
- [ ] `ComputeFromLetters(input, letters, system) (Result, error)` exported in `src/gematria.go`
- [ ] `Compute(input, system)` refactored to call Parse + ComputeFromLetters; signature and behavior unchanged
- [ ] `ComputeTransliterated(input, system, scheme) (Result, error)` exported; sets Result.Scheme
- [ ] Atbash preserved (no behavioral change for any of the four systems via Compute)
- [ ] All existing tests in `src/gematria_test.go` pass without modification
- [ ] godoc on each of the three functions explains its role and how it relates to the others
- [ ] `make build` passes; `make test` passes

