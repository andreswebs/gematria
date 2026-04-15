---
id: gc-43xt
status: closed
deps: []
links: []
created: 2026-04-15T19:25:18Z
type: epic
priority: 1
assignee: Andre Silva
tags: [transliteration, epic]
---
# Word Transliteration

Add opt-in word transliteration so users can type Hebrew words phonetically in the Latin alphabet. Currently, Latin tokens are interpreted as letter aliases (`aleph` → א). After this epic, an opt-in `--transliterate` / `-t` flag switches Latin parsing to word transliteration: `gematria -t shalom` computes שלום=376.

This is a purely additive feature. Default behavior (no `-t`) is unchanged. No backwards-compatibility concerns.

## Architecture

Three layers of changes:

1. **Domain (root package)** — new Scheme type, two scheme implementations, Transliterate function, refactored Compute API, new typed errors.
2. **CLI layer (internal/cli/)** — new flag, env var, dispatch logic, help text, formatter scheme display.
3. **Documentation** — AGENTS.md additions, --help text updates.

## Schemes (v1)

- `academic` (default): strict consonantal mapping with ASCII fallbacks for diacritics. Vowels handled per the resolved spec rules.
- `israeli`: modern Israeli phonetic with matres lectionis (vowels mapped to ו/י/א/ה).

Each scheme has a deterministic mapping table. Ambiguous English combinations (kh, ch, s, sh, ts) resolve to one canonical letter per scheme. Users needing alternatives use Hebrew Unicode directly.

## Public API additions (root package)

```go
type Scheme string

const (
    SchemeAcademic Scheme = "academic"
    SchemeIsraeli  Scheme = "israeli"
)

func ValidSchemes() []Scheme

func Transliterate(input string, scheme Scheme) ([]Letter, error)
func ComputeFromLetters(input string, letters []Letter, system System) (Result, error)
func ComputeTransliterated(input string, system System, scheme Scheme) (Result, error)

type UnknownWordError struct {
    Input       string
    Scheme      Scheme
    Position    int
    Suggestions []string
}

type InvalidSchemeError struct {
    Name  string
    Valid []Scheme
}
```

`Result` gains a `Scheme` field (empty when not transliterated, omitted from JSON via `omitempty`).

## Files affected

**New files (root package):**
- `src/transliteration.go` — Scheme type, constants, ValidSchemes(), Transliterate, ComputeTransliterated, ComputeFromLetters
- `src/transliteration_academic.go` — academic scheme tables and rules
- `src/transliteration_israeli.go` — israeli scheme tables and rules
- `src/transliteration_test.go` — unit tests for transliteration

**Modified files (root package):**
- `src/errors.go` — add UnknownWordError, InvalidSchemeError
- `src/errors_test.go` — tests for new errors
- `src/gematria.go` — refactor Compute (now thin wrapper over Parse + ComputeFromLetters)
- `src/result.go` — add Scheme field
- `src/result_test.go` — tests for Scheme field

**New files (CLI layer):**
- `src/internal/cli/run_transliterate_test.go` — integration tests

**Modified files (CLI layer):**
- `src/internal/cli/config.go` — Transliterate, Scheme fields; flag parsing; env var
- `src/internal/cli/config_test.go` — tests for new fields and precedence
- `src/internal/cli/run.go` — dispatch to ComputeTransliterated; update help text; map new errors
- `src/internal/cli/json.go` — add Scheme field to jsonResult and jsonError with omitempty
- `src/internal/cli/card.go` — display Scheme line when present

**Documentation:**
- `AGENTS.md` — transliteration documentation
- `docs/specs/transliteration.md` — Section 7 open questions resolved (in Task 1)

## CLI surface

- `--transliterate` / `-t` (bool): opt-in mode switch. Without it, existing behavior.
- `--scheme {academic|israeli}` (string): scheme selection. Default `academic`.
- `GEMATRIA_SCHEME` env var with lazy validation (only checked when `-t` is active).
- `cli.Run()` dispatches to `ComputeTransliterated` when `cfg.Transliterate` is true; otherwise to existing `Compute`. Both positional and stdin paths.
- Exit codes: `*UnknownWordError` → 1, `*InvalidSchemeError` → 2.
- `--help` documents the new flag, scheme values, env var, and includes examples.

## Tasks

1. Spec research: finalize scheme mapping tables (resolves Q1-Q7 in transliteration.md)
2. Define Scheme type, constants, ValidSchemes(), and new typed errors (UnknownWordError, InvalidSchemeError)
3. Implement academic scheme (data + per-scheme logic)
4. Implement israeli scheme (data + per-scheme logic)
5. Implement Transliterate() public function
6. Refactor Compute API (ComputeFromLetters + ComputeTransliterated) and add Result.Scheme field
7. Add CLI flags --transliterate/-t, --scheme; GEMATRIA_SCHEME env var
8. Wire CLI dispatch in run.go; update formatters for Scheme; update --help
9. Unit tests for Transliterate, scheme tables, errors, refactored Compute
10. CLI integration tests for --transliterate workflow
11. Update AGENTS.md with transliteration documentation

## Spec References

- docs/specs/transliteration.md (full spec — primary reference)
- docs/specs/requirements.md §2 (Input Handling — cross-reference added)
- docs/specs/cli-design.md §3.3 (Transliteration Matching — cross-reference added)
- docs/specs/code-architecture.md §3.4 (additions noted)

## Acceptance Criteria

- [ ] All 11 child tasks closed
- [ ] `make build` passes (fmt-check, vet, lint, test, compile)
- [ ] `gematria -t shalom` produces a valid result for the academic scheme
- [ ] `gematria -t --scheme israeli shalom` produces the israeli-scheme result
- [ ] `gematria -t shalom --output json | jq .scheme` returns the scheme name
- [ ] Existing scripts using `gematria` (without `-t`) produce identical output before and after this epic
- [ ] AGENTS.md and --help both document the new flags, env var, and schemes
- [ ] docs/specs/transliteration.md status updated from Draft to Ready


## Notes

**2026-04-15T21:56:50Z**

All 11 tasks closed. Build passes (fmt-check + vet + lint + test + compile). CLI verified via smoke test: academic shalom=370, israeli shalom=376, JSON includes scheme when transliterated and omits otherwise, card displays Scheme line, --scheme validation errors with valid-list message.
