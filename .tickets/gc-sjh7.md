---
id: gc-sjh7
status: closed
deps: []
links: []
created: 2026-04-16T04:30:00Z
type: epic
priority: 1
assignee: Andre Silva
tags: [index-migration, epic]
---
# Migrate index from subcommand to --index flag

Migrate `gematria index` from a subcommand to `--index` flag on the main command. This is a breaking change (v0.x).

## Summary

The `gematria index` subcommand is replaced by `--index`, a boolean flag on the main command. Index-specific options use the `--index-` prefix: `--index-output` (explicit path) and `--index-format` (sqlite|index). The `--wordlist` flag and `GEMATRIA_WORDLIST` env var serve both `--find` and `--index`.

## Key decisions

- `--index` is a boolean flag that triggers index-building mode.
- `--wordlist` / `GEMATRIA_WORDLIST` provides the source word list (same resolution as --find).
- `--index-output` replaces the old `--output` (path) on the subcommand. The main `--output` stays format-only.
- `--index-format sqlite|index` replaces the old `--format` on the subcommand.
- Flag conflicts rejected at parse time (exit 2): `--index`+`--find`, `--index`+`-t`, `--index-output` without `--index`, `--index-format` without `--index`, positional args with `--index`.
- Irrelevant env vars silently ignored in index mode.
- Stdin NOT accepted for `--index` — `--wordlist` required.
- `--help` groups indexing flags under an "Indexing:" heading.
- No migration shim for the old `gematria index` syntax — clean break.
- XDG default path, auto-discovery for `--find`, directory auto-creation all preserved unchanged.

## Tasks

1. Migrate index flags to Config (add --index, --index-output, --index-format; conflict validation; remove old subcommand flag set)
2. Wire index mode in Run() (remove subcommand dispatch; add index branch; update help text)
3. Migrate and add tests (rewrite run_index_test.go; add conflict tests; regression tests)
4. Update docs (AGENTS.md, README.md, CLAUDE.md)

## Spec References

- docs/specs/gematria-index.md (updated spec — primary reference)
- docs/specs/cli-design.md

## Acceptance Criteria

- [ ] All 4 child tasks closed
- [ ] `make build` passes (fmt-check, vet, lint, test, compile)
- [ ] `gematria --index --wordlist words.txt` builds index at XDG default path
- [ ] `gematria --index --wordlist words.txt --index-output custom.db` builds at custom path
- [ ] `gematria --index --find 376` exits 2 with conflict error
- [ ] `gematria index --wordlist words.txt` (old syntax) fails — no migration shim
- [ ] `gematria --find 376` auto-discovers default index (unchanged behavior)
- [ ] AGENTS.md, README.md, CLAUDE.md updated with new syntax
- [ ] --help shows Indexing: section with --index, --index-output, --index-format

