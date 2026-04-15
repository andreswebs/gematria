---
id: gc-5t6t
status: closed
deps: [gc-a4ob]
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 1
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, integration-tests, task]
---
# CLI integration tests for --transliterate workflow

Add `src/internal/cli/run_transliterate_test.go` with end-to-end CLI tests exercising the full transliteration workflow via the `Run()` entry point with injected I/O.

## Test patterns

Follow the existing patterns in `run_compute_test.go` and `run_find_test.go`:
- Use `Run(args, stdin, stdout, stderr, getenv)` directly
- Inject fake `*os.File` pipes for stdin/stdout/stderr
- Use a `getenv` closure backed by a map for environment isolation
- Verify exit codes, stdout content, stderr content separately

## Test scenarios

| Scenario | Expected behavior |
| -------- | ----------------- |
| `gematria -t shalom` | Default scheme (academic), correct line output, exit 0 |
| `gematria -t --scheme israeli shalom` | Israeli scheme, correct output, exit 0 |
| `GEMATRIA_SCHEME=israeli gematria -t shalom` | Env var honored, exit 0 |
| `GEMATRIA_SCHEME=israeli gematria -t --scheme academic shalom` | Flag overrides env, exit 0 |
| `gematria -t --output json shalom` | JSON includes `scheme` field |
| `gematria shalom` (no -t) | Existing behavior — exit 1 with UnknownNameError suggestion |
| `gematria -t shalom emet` | Two computations, both succeed, both lines on stdout |
| `printf 'shalom\\nemet\\n' \\| gematria -t` | Stdin batch transliterates each line |
| `gematria -t --scheme bogus shalom` | Exit 2, valid-list error message |
| `GEMATRIA_SCHEME=bogus gematria aleph` (no -t) | Exit 0 — lazy validation |
| `GEMATRIA_SCHEME=bogus gematria -t shalom` | Exit 2 — env validated when -t active |
| `gematria -t qzxw` | Exit 1, UnknownWordError on stderr |
| `gematria -t --output json qzxw` | Exit 1, JSON UnknownWordError on stderr with `scheme` field |
| `gematria -t shalom --atbash` | Atbash mappings displayed (orthogonality verified) |
| `gematria -t --mispar gadol shalom` | Gadol applied to transliterated letters |

## Spec References

- docs/specs/transliteration.md §3 (User-Facing Design), §5 (errors)
- docs/specs/cli-design.md §6 (errors and exit codes)
- See parent epic.

## Acceptance Criteria

- [ ] All scenarios above implemented as table-driven or focused tests
- [ ] No use of real `os.Getenv`, `os.Stdout`, `os.Stderr`, `os.Stdin`
- [ ] Both schemes exercised via flag and env var
- [ ] Lazy GEMATRIA_SCHEME validation verified
- [ ] Exit codes verified for: 0 (success), 1 (UnknownWordError), 2 (InvalidSchemeError)
- [ ] stdout/stderr separation verified for each scenario
- [ ] Existing `run_compute_test.go` tests pass without modification (no regressions)
- [ ] `make test` passes
- [ ] At least 12 distinct test cases

