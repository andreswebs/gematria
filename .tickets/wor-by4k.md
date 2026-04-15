---
id: wor-by4k
status: closed
deps: [wor-ekci, wor-nvv6, wor-sgb3, wor-jkye, wor-viox]
links: []
created: 2026-04-15T04:11:04Z
type: task
priority: 1
parent: wor-jxhz
tags: [core-compute, run-compute, task]
---
# Orchestrate compute path in Run()

Implement the compute orchestration branch of src/internal/cli/run.go, replacing the current stub that returns 0. Also implements --help, --version, and no-args TTY behavior.

## Execution Flow

1. Parse config: call parseConfig(args, getenv). On error, write to stderr and return 2.
2. Handle --help: write complete help text to stdout, return 0.
3. Handle --version: write "gematria X.Y.Z" (or JSON) to stdout, return 0.
4. Resolve color: call UseColor(cfg.NoColor, getenv("NO_COLOR"), stdout).
5. Create formatter: call NewFormatter(cfg.Output, useColor).
6. No-args + stdin TTY: if len(cfg.Args)==0 and IsTerminal(stdin), write usage hint to stderr, return 0.
7. Positional args: if len(cfg.Args) > 0, compute each arg, write result to stdout. On error, write formatter.FormatError to stderr, return appropriate exit code.
8. Stdin (no args, stdin is not a TTY): create a bufio.Scanner over stdin, call processBatch.

## --help Text Requirements

Must include:
- Usage line: "gematria [OPTIONS] [INPUT...]"
- One-sentence description
- Arguments section: what INPUT accepts
- All flags with short forms and valid enum values (hechrachi|gadol|siduri|atbash, line|value|card|json)
- Environment variables section (GEMATRIA_MISPAR, GEMATRIA_OUTPUT, GEMATRIA_WORDLIST, GEMATRIA_LIMIT)
- At least 3 examples covering: single letter, word, reverse lookup

## Error-to-Exit-Code Mapping

- errors.As InvalidCharError or UnknownNameError → return 1
- errors.As InvalidSystemError → return 2
- file open error → return 3

## No-args + TTY

Print to stderr (not stdout):
  "Usage: gematria [OPTIONS] [INPUT...] — try 'gematria --help'"
Return 0 (not an error, just guidance).

## Spec References

- docs/specs/code-architecture.md §4.1 (Run signature, OS injection), §4.6 (No-Args Behavior)
- docs/specs/cli-design.md §5.1 (--help), §5.3 (--version), §5.4 (No-Args), §6.2 (exit codes), §6.3 (no stdout on error)
- docs/specs/requirements.md §8 (errors to stderr, no stdout on error)

## Acceptance Criteria

- [ ] Run() no longer returns hardcoded 0
- [ ] --help prints usage with all flags, env vars, enum values, and examples to stdout; returns 0
- [ ] --version prints "gematria X.Y.Z" (or JSON when --output json); returns 0
- [ ] No-args + stdin TTY prints usage hint to stderr and returns 0
- [ ] No-args + stdin pipe (non-TTY) reads from stdin via processBatch
- [ ] Positional args compute and output results; exit 0 on success
- [ ] On InvalidCharError or UnknownNameError: no stdout output, error to stderr, return 1
- [ ] On InvalidSystemError: return 2
- [ ] On file error: return 3
- [ ] make build passes


## Notes

**2026-04-15T12:01:19Z**

Implemented Run() orchestration in run.go. Added Help bool to Config and --help/-h flag to parseConfig (with fs.SetOutput(io.Discard) to suppress pflag's own stderr output). Run() flow: parseConfig error→exit 2; --help→stdout help+exit 0; --version→stdout version (JSON when --output json)+exit 0; positional args→computeArgs (stops on first error, no partial stdout); no-args+TTY→usage hint to stderr+exit 0; no-args+non-TTY→processBatch. Wrote 13 tests in run_test.go covering all branches. Binary tested end-to-end with Hebrew, Latin, JSON, version, help, and stdin batch.
