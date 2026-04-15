---
id: wor-viox
status: closed
deps: [wor-ekci, wor-jkye]
links: []
created: 2026-04-15T04:10:47Z
type: task
priority: 1
parent: wor-jxhz
tags: [core-compute, stdin-batch, task]
---
# Implement stdin batching

Create src/internal/cli/batch.go implementing multi-line stdin processing with default continue-on-error and optional --fail-early mode.

## processBatch Signature

  func processBatch(
      scanner *bufio.Scanner,
      compute func(string) (gematria.Result, error),
      formatter Formatter,
      stdout, stderr *os.File,
      failEarly bool,
  ) int

The compute parameter is a closure capturing the chosen system, allowing the function to stay decoupled from the Config struct.

## Default Behavior (failEarly = false)

Process every line regardless of errors:
- Successful lines: write formatter.FormatResult(result) to stdout
- Failed lines: write formatter.FormatError(err) to stderr, include the 1-based line number in the error
- After all lines: if any errors occurred AND some lines succeeded → return 4 (partial success)
- If all lines succeeded → return 0
- If all lines failed → return 1 (input error)

## --fail-early Behavior (failEarly = true)

Stop on the first error:
- Write the error to stderr (with line number)
- Return the appropriate exit code for the error type:
  - InvalidCharError or UnknownNameError → 1
  - InvalidSystemError → 2
  - File errors → 3

## Error Message with Line Number

Per-line error on stderr should include the line number:
  "line 3: unknown letter 'x' at position 0"
or in JSON format:
  {"error":"...","line":3,"position":0,...}

## Spec References

- docs/specs/code-architecture.md §4.5 (Stdin Batching)
- docs/specs/cli-design.md §6.5 (Stdin Batch Error Behavior), §6.2 (exit codes 0/1/4)
- docs/specs/requirements.md §2 (stdin: one result per line)

## Acceptance Criteria

- [ ] processBatch processes all lines when failEarly=false
- [ ] Successful lines write to stdout; failed lines write to stderr
- [ ] Returns exit 4 when some lines succeed and some fail
- [ ] Returns exit 0 when all lines succeed
- [ ] Returns exit 1 when all lines fail
- [ ] Per-line errors include 1-based line number
- [ ] With failEarly=true, stops on first error and returns correct exit code
- [ ] make build passes


## Notes

**2026-04-15T11:55:06Z**

Implemented processBatch in batch.go with batchLineError wrapper type. processBatch takes a bufio.Scanner, compute closure, Formatter, stdout/stderr *os.File, and failEarly bool. Returns 0 (all success), 1 (all fail or failEarly input error), 2 (failEarly InvalidSystemError), or 4 (partial success). Updated json.go: added Line field (omitempty) to jsonError struct and unwraps batchLineError to use inner error message and expose line number in JSON. formatErrorPlain in output.go handles batchLineError naturally via err.Error() which returns 'line N: <msg>'. errcheck linter requires '_, _ = fmt.Fprint(...)' for writes to os.File.
