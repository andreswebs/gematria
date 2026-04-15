---
id: gc-a4ob
status: closed
deps: [gc-h6f3, gc-kdm7]
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, cli-dispatch, task]
---
# Wire CLI dispatch in run.go; update formatters and --help for Scheme

Wire the transliteration dispatch in `src/internal/cli/run.go`, update formatters to render the new `Result.Scheme` field, and update the `--help` text.

## Dispatch wiring

In `computeArgs` (positional path):

```go
for _, arg := range cfg.Args {
    var (
        result gematria.Result
        err    error
    )
    if cfg.Transliterate {
        result, err = gematria.ComputeTransliterated(arg, cfg.Mispar, gematria.Scheme(cfg.Scheme))
    } else {
        result, err = gematria.Compute(arg, cfg.Mispar)
    }
    if err != nil {
        _, _ = fmt.Fprint(stderr, formatter.FormatError(err))
        return exitCodeForComputeError(err)
    }
    _, _ = fmt.Fprint(stdout, formatter.FormatResult(result))
}
return 0
```

In the stdin batch path (`processBatch` or its wiring), the `compute` closure switches:

```go
compute := func(input string) (gematria.Result, error) {
    if cfg.Transliterate {
        return gematria.ComputeTransliterated(input, cfg.Mispar, gematria.Scheme(cfg.Scheme))
    }
    return gematria.Compute(input, cfg.Mispar)
}
```

## Error → exit code mapping

Update `exitCodeForComputeError` and `exitCodeForBatchError`:

```go
func exitCodeForComputeError(err error) int {
    var ise *gematria.InvalidSystemError
    var iscse *gematria.InvalidSchemeError
    if errors.As(err, &ise) || errors.As(err, &iscse) {
        return 2
    }
    return 1  // covers UnknownWordError and existing InvalidCharError/UnknownNameError
}
```

## Formatter updates

- `src/internal/cli/json.go`:
  - Extend `jsonResult` struct with `Scheme string \`json:\"scheme,omitempty\"\`` and populate from `r.Scheme`. Existing JSON output for non-transliterated computations is byte-identical.
  - Extend `jsonError` struct with `Scheme string \`json:\"scheme,omitempty\"\`` and populate from `*UnknownWordError.Scheme` via `errors.As` type assertion.
- `src/internal/cli/card.go`: when `r.Scheme != \"\"`, include a `Scheme: <name>` line in the card output.
- `src/internal/cli/line.go`: no change (line format does not show scheme).
- `src/internal/cli/value.go`: no change.

## --help text additions

In `helpText` constant in run.go, add:
- Under Options:
  - `-t, --transliterate           Interpret Latin input as Hebrew words (per --scheme)`
  - `    --scheme string           Transliteration scheme: academic|israeli (default: academic)`
- Under Environment Variables:
  - `GEMATRIA_SCHEME          Default transliteration scheme (academic|israeli)`
- Under Examples (at least 2 new lines):
  - `gematria -t shalom                      Compute שלום (academic scheme)`
  - `gematria -t --scheme israeli gadol      Compute גדול (israeli scheme)`

## Spec References

- docs/specs/transliteration.md §3.3 (mode switching), §3.5 (composition), §5.2 (error output)
- docs/specs/cli-design.md §6.2 (exit codes)
- See parent epic.

## Acceptance Criteria

- [ ] `cfg.Transliterate=true` triggers dispatch to `ComputeTransliterated` for both positional args and stdin batch
- [ ] `cfg.Transliterate=false` preserves existing dispatch to `Compute` (byte-identical output)
- [ ] `*UnknownWordError` exits 1; `*InvalidSchemeError` exits 2 (verified for both positional and batch paths)
- [ ] `jsonResult` has `Scheme` field with `omitempty`; non-transliterated JSON output unchanged
- [ ] `cardFormatter` displays scheme line when `r.Scheme != \"\"`
- [ ] `jsonError` has `Scheme` field populated from `*UnknownWordError`
- [ ] `line` and `value` formatters unchanged
- [ ] `--help` documents `-t`, `--transliterate`, `--scheme`, and `GEMATRIA_SCHEME`
- [ ] `--help` includes at least 2 transliteration examples (one per scheme)
- [ ] `make build` passes; `make test` passes (existing tests still pass)

