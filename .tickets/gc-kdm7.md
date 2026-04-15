---
id: gc-kdm7
status: closed
deps: [gc-0mhe]
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, cli-flags, task]
---
# Add CLI flags --transliterate/-t, --scheme; GEMATRIA_SCHEME env var

Add to `src/internal/cli/config.go` the flag and env-var infrastructure for transliteration. **No dispatch wiring yet** — that's Task 8.

## Config struct additions

```go
type Config struct {
    // ... existing fields ...
    Transliterate bool
    Scheme        string
}
```

## Flag definitions

Inside `parseConfig`:

```go
var transliterate bool
var schemeFlag string

fs.BoolVarP(&transliterate, \"transliterate\", \"t\", false,
    \"interpret Latin input as Hebrew words (per --scheme)\")
fs.StringVar(&schemeFlag, \"scheme\", \"\",
    \"transliteration scheme (academic, israeli)\")
```

Verify `-t` does not collide with any other short flag in current `config.go` — current short flags are `-m` (mispar), `-o` (output), `-h` (help), `-l` (limit). `-t` is free.

## Resolution and validation

```go
var validSchemes = []string{\"academic\", \"israeli\"}

// Eager validation when --scheme explicitly provided
scheme := schemeFlag
if fs.Changed(\"scheme\") && !contains(validSchemes, scheme) {
    return Config{}, fmt.Errorf(\"invalid value %q for --scheme\\nvalid values: %s\",
        scheme, strings.Join(validSchemes, \", \"))
}
// Env var resolution when no flag
if scheme == \"\" {
    scheme = getenv(\"GEMATRIA_SCHEME\")
}
// Lazy validation: only validate env var when transliterate is active
if transliterate && scheme != \"\" && !contains(validSchemes, scheme) {
    return Config{}, fmt.Errorf(\"invalid value %q for GEMATRIA_SCHEME\\nvalid values: %s\",
        scheme, strings.Join(validSchemes, \", \"))
}
// Default
if transliterate && scheme == \"\" {
    scheme = \"academic\"
}
```

This mirrors the lazy-validation pattern used for `GEMATRIA_LIMIT` (only validated when `--find` is active).

## Add to validation lists

Add `validSchemes` to the package-level vars near `validSystems` and `validOutputs`.

## Spec References

- docs/specs/transliteration.md §3.1 (flags), §3.2 (env var)
- docs/specs/cli-design.md §3.4 (precedence), §6.6 (lazy env var validation)
- See parent epic.

## Acceptance Criteria

- [ ] `Config` has `Transliterate` (bool) and `Scheme` (string) fields
- [ ] `--transliterate`/`-t` parses as bool via pflag
- [ ] `--scheme` parses as string via pflag
- [ ] Eager validation: invalid `--scheme` returns error with exit-code-2 semantics; error message lists valid values
- [ ] Lazy validation: `GEMATRIA_SCHEME=bogus` (no -t) does NOT error
- [ ] Lazy validation: `GEMATRIA_SCHEME=bogus` with -t errors with exit-code-2 semantics
- [ ] Default: `cfg.Scheme == \"academic\"` when `Transliterate` is true and no flag/env set
- [ ] Precedence: explicit flag > env var > default (verified in tests)
- [ ] No regression: existing config_test.go tests pass
- [ ] At least 6 new test cases in config_test.go (precedence, valid, invalid, lazy validation, default)
- [ ] `make build` passes; `make test` passes


## Notes

**2026-04-15T19:39:49Z**

Added Transliterate (bool) and Scheme (string) fields to Config. Added --transliterate/-t bool flag and --scheme string flag. Added GEMATRIA_SCHEME env var with lazy validation (only validated when -t active, mirroring GEMATRIA_LIMIT pattern). Eager validation when --scheme flag is explicitly set. Default scheme is 'academic' only when Transliterate=true. Added validSchemes package-level var alongside existing validSystems/validOutputs. 8 new test cases covering short/long flag, default, scheme values, eager/lazy validation, env var, and flag-over-env precedence.
