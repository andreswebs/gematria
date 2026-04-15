---
id: wor-nvv6
status: closed
deps: [wor-gdda]
links: []
created: 2026-04-15T04:10:02Z
type: task
priority: 1
parent: wor-jxhz
tags: [core-compute, cli-config, task]
---
# Implement CLI Config struct and flag parsing

Create src/internal/cli/config.go with the Config struct and a parseConfig function that resolves all flags and environment variables for the compute workflow.

## Config Struct Fields

- Mispar gematria.System — selected gematria system (default: hechrachi)
- Output string — selected output format (default: "line")
- NoColor bool — disable ANSI color output
- Atbash bool — display Atbash substitution alongside normal output
- FailEarly bool — stop stdin batch on first error
- Version bool — print version and exit
- Args []string — remaining positional arguments (the input words/letters)

## Flag Definitions (using pflag)

- --mispar / -m <system> — gematria system; valid: hechrachi, gadol, siduri, atbash
- --output / -o <format> — output format; valid: line, value, card, json
- --no-color — disable color
- --atbash — show Atbash letter substitutions in output
- --fail-early — stop on first stdin error
- --version — print version and exit
- --help / -h — print help and exit (handled by pflag)

## Env Var Resolution (flag > env > default)

- GEMATRIA_MISPAR: used as default if --mispar not provided; validated lazily (only when compute is invoked)
- GEMATRIA_OUTPUT: used as default if --output not provided; validated on startup since it affects how errors are rendered

## Enum Validation

Flag values for --mispar and --output require exact match (no prefix matching). On invalid value, return an error message that names the flag, shows the invalid value, and lists all valid options. Example:
  "Error: Invalid value 'standard' for --mispar.\nValid values: hechrachi, gadol, siduri, atbash"

This is exit code 2 (CLI misuse).

## Spec References

- docs/specs/code-architecture.md §4.2 (Config struct, strict enum matching, lazy env var validation)
- docs/specs/cli-design.md §3.4 (flag > env > default), §5.2 (flag validation messages), §6.6 (lazy env var validation)
- docs/specs/requirements.md §3 (--mispar, GEMATRIA_MISPAR), §4 (--output, GEMATRIA_OUTPUT), §9 (env vars)

## Acceptance Criteria

- [ ] Config struct has all required fields
- [ ] parseConfig resolves --mispar with flag > GEMATRIA_MISPAR > "hechrachi" precedence
- [ ] parseConfig resolves --output with flag > GEMATRIA_OUTPUT > "line" precedence
- [ ] Invalid --mispar value produces exit-code-2 error listing all valid systems
- [ ] Invalid --output value produces exit-code-2 error listing all valid formats
- [ ] --no-color, --atbash, --fail-early, --version flags are parsed correctly
- [ ] Positional args are captured in Config.Args
- [ ] Uses pflag (not stdlib flag) for GNU-style --long and -s short forms
- [ ] make build passes


## Notes

**2026-04-15T11:32:59Z**

Implemented src/internal/cli/config.go with Config struct and parseConfig(). Added pflag v1.0.10 as a dependency. Key design: GEMATRIA_OUTPUT is validated eagerly (needed for error rendering format), GEMATRIA_MISPAR is validated lazily (only when compute is invoked) — invalid env var stored as-is in Config.Mispar, Compute() handles validation. Error strings are lowercase to satisfy staticcheck ST1005. Full TDD with 8 tests covering defaults, flag/env precedence, invalid enum values, lazy vs eager env var validation, bool flags, positional args, and short flag forms.
