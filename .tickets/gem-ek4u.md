---
id: gem-ek4u
status: closed
deps: []
links: []
created: 2026-07-24T22:47:04Z
type: task
priority: 2
assignee: Andre Silva
---
# Remove the remote/HTTP wordlist backend

## Context and rationale

The remote backend (`NewRemoteWordSource`) is overengineered for a CLI
tool that computes Hebrew numerology values: it brings network
dependencies, bearer-token authentication, HTTP error handling, and a
response-size limit, all for a path nobody uses. Users who need shared
word lists can distribute `.db` or index files directly.

Removing it shrinks the public API surface, eliminates `net/http` from
the root package, removes the `GEMATRIA_WORDLIST_TOKEN` credential
from the environment surface entirely (which also retires the
unredacted-token security finding this ticket originally tracked), and
simplifies `--wordlist-format` from four values to three.

Background: `.local/planning/remove-remote-backend.md` is the original
draft plan. It predates the module move from `src/` to the repo root,
so its paths and line numbers are stale; this ticket supersedes it
with details verified on 2026-07-25 against the current tree. The spec
`docs/specs/wordlist-backends.md` was already updated historically and
records the removal at line 23; the code was never changed.

Repository facts: module `github.com/andreswebs/gematria`, rooted at
the repository top level. Quality gate: `make build` from the repo
root (fmt-check, vet, lint, test, compile).

## Code changes (verified file:line, re-verify before editing)

### Delete entirely

1. `backend_remote.go` (119 lines): `remoteWordSource`,
   `RemoteOption`, `WithAuthToken`, `WithHTTPClient`,
   `NewRemoteWordSource`, `remoteResponseBody`, `remoteWordJSON`, and
   the `FindByValue` method.
2. `backend_remote_test.go` (~310 lines, 12 test functions, all
   exclusively about the remote backend).

### `internal/cli/run.go`

1. Remove the `openRemoteWordSource` helper (lines 331-340).
2. In `openWordSource` (line 276): remove the `http://`/`https://`
   prefix dispatch to `openRemoteWordSource` (lines 279-282) and the
   "path begins with http:// or https:// -> remote backend" entry in
   the doc comment's detection-order list (line 267); renumber the
   remaining detection steps in that comment.
3. In `openWordSourceByFormat` (line 297): remove the
   `case "remote":` branch (lines 303-304).
4. Help text: remove the `GEMATRIA_WORDLIST_TOKEN` env var line
   (line 59) and change the `--wordlist-format` description
   (line 43) from `sqlite|index|remote|memory` to
   `sqlite|index|memory`.
5. Parameter cleanup: after the removals, `getenv` becomes unused in
   both `openWordSource` and `openWordSourceByFormat`; drop the
   parameter from both signatures and update their call sites
   (`runFind` at line 372 and the internal call at line 278). Keep
   `getenv` on `runFind` itself; it is still used by
   `discoverDefaultIndex` (line 361).

### `internal/cli/config.go`

1. Line 37: `validWordlistFormats` from
   `{"sqlite", "index", "remote", "memory"}` to
   `{"sqlite", "index", "memory"}`.
2. Line 76: the `--wordlist-format` flag usage string,
   `backend override: sqlite|index|remote|memory`, drops `remote`.
3. Line 27: the `WordlistFormat` field comment
    `(sqlite|index|remote|memory)` drops `remote`.

### `internal/cli/run_backends_test.go`

1. Remove three tests (the old plan listed only two; the third was
    verified present):
    - `TestRun_find_httpURL_usesRemoteBackend` (lines 292-318)
    - `TestRun_find_wordlistToken_sentAsAuthHeader` (lines 322-344)
    - `TestRun_find_wordlistFormatRemote_explicit` (lines 348-373)
2. Remove the `net/http`/`net/http/httptest` imports if they become
    unused in that file; keep shared helpers (for example `envWith`)
    if other tests still use them, delete them if orphaned.

### Not affected (checked; do not touch)

- `internal/cli/exitcodes.go`: the registry and the exit 74 wording
  ("wordlist/index/backend") stay accurate without the remote
  backend. No registry change: the remote path produced no exit code
  of its own.
- `internal/cli/helptext_test.go`: the drift tests key off registered
  flags; `--wordlist-format` remains a flag, so only its usage string
  changes (already covered above).

## Documentation changes

1. `docs/manual-qa.md`:
    - Line 25: drop `remote` from the backend list in the overview.
    - Line 319: the expected stderr for an invalid format currently
      reads `sqlite, index, remote, memory`; update to
      `sqlite, index, memory` (this message is generated from
      `validWordlistFormats`, so the doc must match the new output).
    - Remove test case `TC-BACKEND-005 - Remote backend` entirely
      (around lines 619-633). Leave the other TC-BACKEND numbering as
      is; a gap is acceptable and renumbering would churn
      cross-references.
    - Lines 1338 and 1366: remove `GEMATRIA_WORDLIST_TOKEN` from both
      `unset` lists.
2. `docs/learnings.md`:
    - Line 28 (threading getenv for `GEMATRIA_WORDLIST_TOKEN`):
      remove the entry; it documents machinery this ticket deletes.
    - Line 29 (httptest for remote backend CLI tests): remove, or
      generalize to a statement about httptest for CLI tests that
      does not reference the remote backend.
    - Line 23 (S1016 struct-tag conversion, mentions
      `remoteWordJSON` as its example): keep; it is a general Go
      learning whose example happens to be historical.
3. `docs/specs/wordlist-backends.md`: already records the removal;
    verify it needs no further edit.
4. `docs/specs/gematria-index.md` and
    `docs/specs/code-architecture.md`: verified free of remote
    references on 2026-07-25; re-verify with the grep sweep below and
    fix if anything drifted.

## Historical ticket annotations

1. Add a short dated note (do not rewrite the historical content) to
    each of these closed tickets, preserving their frontmatter:
    - `.tickets/wor-6uug.md` (backends epic): remote backend removed.
    - `.tickets/wor-0153.md` (remote backend implementation):
      implemented, later removed.
    - `.tickets/wor-z816.md` (backend tests): remote tests removed.
    - `.tickets/wor-22g6.md` (auto-selection): remote detection path
      and `GEMATRIA_WORDLIST_TOKEN` removed.
    Run markdownlint on each edited ticket file (use the global
    config `~/.markdownlint.yaml` unless a repo-local config exists).

## Acceptance criteria

1. This grep returns zero matches across all Go source (the docs
   exceptions are `wordlist-backends.md`'s historical note,
   `learnings.md` line 23, and ticket files):

   `grep -rn "remoteWordSource\|RemoteOption\|WithAuthToken\|WithHTTPClient\|NewRemoteWordSource\|openRemoteWordSource\|GEMATRIA_WORDLIST_TOKEN" --include='*.go' .`

2. `gematria --find 1 --wordlist /tmp/x --wordlist-format remote`
   exits 64 with the invalid-value message listing
   `sqlite, index, memory`.
3. `go mod tidy` produces no `go.mod`/`go.sum` changes (the backend
   used only stdlib).
4. `make build` from the repo root is green (fmt-check, vet, lint,
   test, compile). Do not silence lint errors with `_ =`.
5. All documentation edits above are applied and the edited markdown
   files pass markdownlint.

## Notes

- Do not create git commits or branches; leave all git workflows to
  the owner.
- Line numbers were verified on 2026-07-25; re-verify against the
  current code before editing, since they drift.
- This ticket replaces its own earlier scope (wrapping the token in a
  redaction type): removing the backend removes the credential, so no
  redaction work is needed. If a future feature reintroduces a
  credential, vendor the standard `secret.Value` redaction type per
  the logging ADR (`docs/adr/0004-logging.md`).

**2026-07-25T03:21:44Z**

Removed the remote/HTTP wordlist backend. Deleted backend_remote.go and backend_remote_test.go. In internal/cli: dropped openRemoteWordSource, the http:// dispatch and 'remote' format case, and the now-unused getenv param from openWordSource/openWordSourceByFormat (kept getenv on runFind for discoverDefaultIndex); removed GEMATRIA_WORDLIST_TOKEN help line; --wordlist-format now sqlite|index|memory in help, config validWordlistFormats, flag usage, and Config field comment. Removed 3 remote tests + net/http imports from run_backends_test.go (envWith kept, still used widely). Docs: manual-qa.md (overview backend list, invalid-format expected stderr, removed TC-BACKEND-005, dropped token from unset lists + troubleshooting row), learnings.md (removed token-threading and generalized httptest bullet; kept S1016). Added dated annotations to wor-6uug/wor-0153/wor-z816/wor-22g6. All ACs green: Go grep zero matches, exit 64 lists 'sqlite, index, memory', go mod tidy clean, make build green. NOTE: no markdownlint config exists in repo or at ~/.markdownlint.yaml here; default rules flag pre-existing project-wide MD013/MD060 style, but edits introduced no new violations (error counts dropped on every edited file).
