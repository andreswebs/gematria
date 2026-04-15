---
id: wor-jkye
status: closed
deps: [wor-cg08, wor-nvv6, wor-sgb3]
links: []
created: 2026-04-15T04:10:36Z
type: task
priority: 1
parent: wor-jxhz
tags: [core-compute, output-formatters, task]
---
# Implement four compute output formatters

Create src/internal/cli/line.go, value.go, card.go, and json.go. Each file implements the Formatter interface for its format. All four FormatResult methods handle both single-letter and multi-letter (word) input.

## RTL Marks

Hebrew text in human-facing formats (line, card) must be wrapped:
  prefix: U+200F (RIGHT-TO-LEFT MARK)
  suffix: U+200E (LEFT-TO-RIGHT MARK)
JSON format must NOT include RTL marks or ANSI codes — data only.

## LineFormatter (line.go)

Single letter: "Aleph (‏א‎) = 1"
Word: "‏אמת‎ = 441 (‏א‎=1 + ‏מ‎=40 + ‏ת‎=400)"
With --atbash flag: append "→ ‏תשא‎" (Atbash substitution) after the word
Color: apply green to numeric values if useColor is true; bold to Hebrew text

## ValueFormatter (value.go)

Bare integer followed by newline. No labels, no Hebrew, no color. --atbash has no visible effect in value format (it is a display modifier; to compute using Atbash, use --mispar atbash).

## CardFormatter (card.go)

Multi-line output per letter in an aligned table:
  Letter | Name      | Value | Meaning
  ‏א‎     | Aleph     |     1 | Ox
  ‏מ‎     | Mem       |    40 | Water
  ‏ת‎     | Tav       |   400 | Cross
  ─────────────────────────────────
  Total: 441  [hechrachi]
With --atbash: add an Atbash column showing substituted letter

## JSONFormatter (json.go)

Single letter:
  {"input":"aleph","system":"hechrachi","total":1,"letters":[{"char":"א","name":"Aleph","value":1,"meaning":"Ox","position":1}]}

Word:
  {"input":"אמת","system":"hechrachi","total":441,"letters":[...]}

FormatError returns a JSON object for stderr:
  {"error":"Unknown letter 'x' at position 3","invalid_input":"x","position":3,"suggestions":["shin"]}

No ANSI codes, no RTL marks anywhere in JSON output.

## Spec References

- docs/specs/requirements.md §4 (output formats), §5 (RTL Output)
- docs/specs/cli-design.md §4.3 (human patterns), §4.4 (agent patterns), §4.5 (structured errors)

## Acceptance Criteria

- [ ] LineFormatter.FormatResult wraps Hebrew with U+200F prefix and U+200E suffix
- [ ] LineFormatter shows per-letter breakdown for words (e.g., "א=1 + מ=40 + ת=400")
- [ ] ValueFormatter.FormatResult returns only the integer as a string with newline
- [ ] CardFormatter.FormatResult produces aligned table with Letter, Name, Value, Meaning columns
- [ ] JSONFormatter.FormatResult produces valid JSON with no RTL marks and no ANSI codes
- [ ] JSONFormatter.FormatError produces {"error":..., "invalid_input":..., "position":..., "suggestions":...}
- [ ] Plain text FormatError (line/value/card) produces human-readable stderr message
- [ ] All four formatters compile and satisfy the Formatter interface
- [ ] make build passes


## Notes

**2026-04-15T11:49:02Z**

Implemented all four compute output formatters (line, value, card, json) across separate files: line.go, value.go, card.go, json.go. Key decisions: (1) Added showAtbash bool as third param to NewFormatter — wor-by4k must pass cfg.Atbash when it calls NewFormatter. (2) Moved shared helpers (wrapHebrew, formatErrorPlain, ANSI consts, rtlMark/ltrMark) into output.go. (3) JSONFormatter.FormatError handles InvalidCharError, UnknownNameError, InvalidSystemError with typed fields; nil Suggestions serialises as [] not null. (4) CardFormatter uses fmt.Fprintf(&sb, ...) not WriteString(fmt.Sprintf(...)) to satisfy QF1012. (5) FormatLookup is a stub returning empty string on all four formatters — wor-g7cz implements the full lookup formatting.
