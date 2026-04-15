---
id: wor-e4m1
status: closed
deps: []
links: []
created: 2026-04-15T04:11:18Z
type: task
priority: 2
parent: wor-jxhz
tags: [core-compute, agents-md, task]
---
# Write AGENTS.md documentation

Write the AGENTS.md file at the project root. The current file is a placeholder. This document is the primary "agent knowledge packaging" artifact described in the cli-design spec.

## Required Sections

### Quick Start for Agents

Copy-paste examples an agent can use immediately:
  # Compute a value (JSON output)
  gematria --output json אמת

  # Compute from transliteration
  gematria --output json aleph mem tav

  # Get just the number
  gematria --output value אמת

  # Reverse lookup
  gematria --find 441 --wordlist "${WORDLIST}" --output json

### Recommended Flags

- --output json for structured parsing
- --output value for bare numeric results in pipelines
- --mispar to select the gematria system (hechrachi|gadol|siduri|atbash)

### Error Handling

Explain the exit code table (0/1/2/3/4) and what each means for agent branching. Show how to parse JSON errors from stderr when --output json is active.

### Input Formats

Document both Hebrew Unicode and Latin transliteration input, with examples of each. Note that --mispar and --output require exact enum values.

### Batch Processing

Show stdin batch usage and when to use --fail-early.

### Security Note

Word list contents appear verbatim in output. If piping gematria output to an LLM, a malicious word list could contain prompt injection payloads. The tool is a data pipe — it does not sanitize word list content.

## Spec References

- docs/specs/cli-design.md §2.7 (Agent Knowledge Packaging), §6.2 (exit codes), §4.4 (agent output patterns)

## Acceptance Criteria

- [ ] AGENTS.md exists at the project root
- [ ] Includes a Quick Start section with copy-paste examples using real flag syntax
- [ ] Documents all exit codes with meanings
- [ ] Documents --output json for structured errors on stderr
- [ ] Documents input format options (Hebrew Unicode and Latin transliteration)
- [ ] Includes batch processing example with --fail-early
- [ ] Includes security note about word list contents and prompt injection
- [ ] Flag placeholders use environment variable style (e.g., "${WORDLIST}") not angle-bracket style


## Notes

**2026-04-15T12:49:03Z**

Replaced placeholder AGENTS.md (which was a copy of CLAUDE.md developer guide) with agent-focused documentation. Covers: quick-start examples with ${ENV_VAR} style placeholders, recommended flags, both input formats (Hebrew Unicode and Latin transliteration), full JSON output schemas for compute/lookup/version/error, exit code table with shell branching example, batch processing with --fail-early, env var precedence, and security note on word list prompt injection. Key gotcha: flag-level errors (e.g. invalid --mispar) are always plain text on stderr even with --output json, because they occur before the formatter is initialized — documented this explicitly.
