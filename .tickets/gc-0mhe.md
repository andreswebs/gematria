---
id: gc-0mhe
status: closed
deps: []
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, types, task]
---
# Define Scheme type, constants, and new typed errors

Add to the root package the foundational types needed by all subsequent transliteration tasks.

## Scheme type

Add to a new file `src/transliteration.go` (top of file):

```go
type Scheme string

const (
    SchemeAcademic Scheme = \"academic\"
    SchemeIsraeli  Scheme = \"israeli\"
)

func ValidSchemes() []Scheme { return []Scheme{SchemeAcademic, SchemeIsraeli} }
```

Mirrors the existing `System` type and `ValidSystems()` pattern in `src/systems.go`.

## New typed errors (errors.go)

Add two new error types alongside InvalidCharError, UnknownNameError, InvalidSystemError:

```go
// UnknownWordError is returned by Transliterate when input cannot be resolved
// to Hebrew letters under the given scheme. Suggestions is empty in v1.
type UnknownWordError struct {
    Input       string
    Scheme      Scheme
    Position    int      // token index for multi-token input
    Suggestions []string
}

// InvalidSchemeError parallels InvalidSystemError. Returned when a Scheme
// string does not match any known constant.
type InvalidSchemeError struct {
    Name  string
    Valid []Scheme
}
```

Compile-time interface checks at package level:

```go
var _ error = (*UnknownWordError)(nil)
var _ error = (*InvalidSchemeError)(nil)
```

Error message formats (matching existing patterns; lowercase, no trailing punctuation):
- `UnknownWordError`: `input %q cannot be transliterated in scheme %q` (with suggestions appended when present, same join style as UnknownNameError)
- `InvalidSchemeError`: `unknown scheme %q; valid values: %s` (parallel to InvalidSystemError)

## Spec References

- docs/specs/transliteration.md §3.1 (Scheme constants), §5.1 (UnknownWordError), §5.3 (InvalidSchemeError)
- See parent epic for the full public API surface.

## Acceptance Criteria

- [ ] `Scheme` type defined as string-based, parallel to `System`
- [ ] Constants `SchemeAcademic`, `SchemeIsraeli` defined
- [ ] `ValidSchemes()` returns both in stable order
- [ ] `UnknownWordError` struct defined with Input, Scheme, Position, Suggestions fields and godoc
- [ ] `UnknownWordError.Error()` follows the spec format; lowercase, no trailing punctuation; includes suggestions when non-empty
- [ ] `InvalidSchemeError` defined parallel to InvalidSystemError, with godoc
- [ ] `InvalidSchemeError.Error()` lists valid schemes
- [ ] Compile-time interface assertions present for both new errors
- [ ] `make build` passes


## Notes

**2026-04-15T19:28:33Z**

Implemented Scheme type, SchemeAcademic/SchemeIsraeli constants, and ValidSchemes() in new src/transliteration.go. Added UnknownWordError and InvalidSchemeError to src/errors.go with compile-time interface assertions. Error message formats: UnknownWordError uses 'input %q cannot be transliterated in scheme %q' with optional '; did you mean: ...' suffix; InvalidSchemeError parallels InvalidSystemError. All tests pass, make build clean.
