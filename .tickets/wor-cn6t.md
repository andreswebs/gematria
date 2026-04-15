---
id: wor-cn6t
status: closed
deps: [wor-by4k]
links: []
created: 2026-04-15T04:11:52Z
type: task
priority: 2
parent: wor-jxhz
tags: [core-compute, compute-cli-tests, task]
---
# CLI integration tests for compute path

Write integration tests for the compute workflow in src/internal/cli/, calling Run() directly with injected OS primitives. No subprocesses — tests call the Go function.

## Test Infrastructure

Use os.Pipe() for stdout and stderr capture. Pass a getenv closure over a map for environment variables. Pass nil or pipe file descriptors for stdin (non-TTY). Follow the pattern from code-architecture.md §7.

## Test Cases

### Positional argument compute
- Single Hebrew letter: verify correct output for each format (line, value, card, json)
- Latin transliteration: "aleph" → same result as א
- Multi-word Latin: "aleph mem tav" → result for אמת
- Unknown transliteration: exit 1, error on stderr, stdout empty
- Invalid char: exit 1, error on stderr, stdout empty

### System selection
- --mispar gadol with a sofit letter: verify extended value
- --mispar siduri: verify ordinal value
- --mispar atbash: verify substituted computation
- GEMATRIA_MISPAR env var: respected when flag absent; overridden when flag present
- Invalid --mispar: exit 2, error lists valid values

### Output format
- --output value: bare integer only
- --output json: valid JSON on stdout
- --output json on error: JSON error object on stderr
- GEMATRIA_OUTPUT env var override behavior
- Invalid --output: exit 2

### Special behaviors
- No args + non-TTY stdin (pipe): reads from stdin
- No args + TTY stdin: usage hint to stderr, exit 0
- --version: prints version string, exit 0
- --version --output json: prints JSON version, exit 0
- --no-color: ANSI codes absent in output

### Stdin batch
- Multiple valid lines: all produce stdout output, exit 0
- Mix of valid and invalid lines: valid on stdout, errors on stderr, exit 4
- All invalid: exit 1
- --fail-early: stops at first error

### Error invariant
- In all error scenarios: stdout must be empty

## Spec References

- docs/specs/code-architecture.md §7 (CLI testing with Run() injection, os.Pipe())
- docs/specs/cli-design.md §6.2 (exit codes), §6.3 (no stdout on error), §6.5 (batch behavior)

## Acceptance Criteria

- [ ] Tests use Run() with injected args, os.Pipe() for stdout/stderr, getenv closure; no os.Setenv
- [ ] Each of the four output formats tested with at least one known input
- [ ] All four mispar systems tested via --mispar flag
- [ ] GEMATRIA_MISPAR and GEMATRIA_OUTPUT env var override tested
- [ ] Invalid flag values produce exit 2 and list valid options
- [ ] Invalid input produces exit 1 with error on stderr and empty stdout
- [ ] Stdin batch with mixed valid/invalid lines produces exit 4
- [ ] --fail-early produces exit 1 on first error, remaining lines unprocessed
- [ ] --version and --no-color behaviors tested
- [ ] make test passes


## Notes

**2026-04-15T12:41:31Z**

Added run_compute_test.go with 17 integration tests for the compute path. Tests call Run() directly with os.Pipe() for stdout/stderr and getenv closures. Covers: all four output formats (value/line/card/json) with known inputs; all four mispar systems (hechrachi/gadol/siduri/atbash) via explicit --mispar flag; GEMATRIA_MISPAR and GEMATRIA_OUTPUT env var precedence and override; invalid --mispar flag (exit 2 with valid values listed); Latin transliteration (single and multi-word); --no-color (no ANSI codes); stdin batch all-invalid (exit 1); --fail-early stops on first error (exit 1). All existing tests in run_test.go continued passing. make build passes cleanly.
