---
id: gem-kj4m
status: open
deps: []
links: [gem-yosr]
created: 2026-07-24T20:37:36Z
type: task
priority: 1
assignee: Andre Silva
---
# Migrate Go module from src/ to repo root

# Migrate the Go module from `src/` to the repo root

## Summary

The gematria repository keeps its Go module under `src/` but the module
declares its path as `github.com/andreswebs/gematria`. The Go toolchain and
the module proxy map that path to a `go.mod` at the **repository root**, not
under `src/`. As published today the module is therefore unreachable through
the proxy:

- `go install github.com/andreswebs/gematria/cmd/gematria@latest` fails.
- The library packages at the module root (the `gematria` package and its
  helpers) cannot be imported by any third party via `go get`.
- pkg.go.dev, govulncheck, and dependabot cannot resolve the module.

This plan moves the entire module up one level so that `go.mod` lives at the
repository root. Import paths do **not** change, because the declared module
path already omits `src`. No `.go` file needs editing.

gematria is the library case of the fleet: the module root holds roughly
twenty loose library `.go` files (the `package gematria` sources) alongside
`cmd/gematria` and `internal/cli`. After the move those land directly at the
repository root, which is the idiomatic layout for an importable Go library.

This is a working-tree change only. Do not create commits, branches, or tags;
leave all git history operations to the repository owner. The verification
steps that need a published tag are called out explicitly and depend on the
owner tagging a release after merge.

## Decision: keep library files at the repo root

The draft raised an open question: leave the library `.go` files at the
repository root (idiomatic, no import-path break) or move them into a named
subpackage such as `gematria/` (cleaner root, but a breaking import-path
change).

**Recommendation: keep the library files at the repository root.**

Rationale:

- The package is already named `gematria` and its public identifiers are
  addressed as `github.com/andreswebs/gematria`. Keeping the files at the
  root preserves that import path exactly, so this migration stays a pure
  relocation with zero source edits and zero consumer impact.
- Root-level library packages are standard and idiomatic in Go. Many widely
  used libraries keep their primary package at the module root. A root that
  holds around twenty `.go` files plus `cmd/`, `internal/`, `README.md`, and
  `Makefile` is unremarkable.
- Introducing a subpackage (for example `github.com/andreswebs/gematria/gematria`)
  would be a breaking change to the public import path. Since no external
  consumer exists yet, it would not break anyone today, but it permanently
  couples the import path to an extra path segment for no functional gain and
  would still have to be decided before the first resolvable tag. There is no
  reason to pay that cost.
- The CLI main package stays at `github.com/andreswebs/gematria/cmd/gematria`
  either way, so the subpackage option only affects the library ergonomics,
  and it makes them worse (a redundant `gematria.gematria` feel at the call
  site).

If root clutter ever becomes a real problem, the internal-only helpers could
later move under `internal/`, which does not affect the public API. That is
out of scope here.

## Current layout (before)

```text
gematria/
  Makefile
  README.md
  agents.md            # CLAUDE.md is a symlink to this
  LICENSE
  .github/
    dependabot.yml
    workflows/
      build.yaml
      pull-request.yaml
      build.dispatch.yaml
      validate.dispatch.yaml
  docs/
  bin/
  dist/
  src/
    go.mod             # module github.com/andreswebs/gematria
    go.sum
    backend_index.go        backend_index_test.go
    backend_remote.go       backend_remote_test.go
    backend_sqlite.go       backend_sqlite_test.go
    errors.go               errors_test.go
    gematria.go             gematria_test.go
    letters.go              letters_test.go
    result.go               result_test.go
    systems.go
    transliteration.go
    transliteration_academic.go
    transliteration_israeli.go
    transliteration_test.go
    wordlist.go             wordlist_test.go
    cmd/
      gematria/
        main.go        # package main, imports .../internal/cli
    internal/
      cli/             # run.go, config.go, batch.go, output.go, ... (+ tests)
```

There is no `.golangci.yml` in the repository; `golangci-lint` runs with its
built-in defaults, so there is no lint config file to move.

## Target layout (after)

```text
gematria/
  go.mod             # module github.com/andreswebs/gematria (now at root)
  go.sum
  backend_index.go        backend_index_test.go
  backend_remote.go       backend_remote_test.go
  backend_sqlite.go       backend_sqlite_test.go
  errors.go               errors_test.go
  gematria.go             gematria_test.go
  letters.go              letters_test.go
  result.go               result_test.go
  systems.go
  transliteration.go
  transliteration_academic.go
  transliteration_israeli.go
  transliteration_test.go
  wordlist.go             wordlist_test.go
  cmd/
    gematria/
      main.go
  internal/
    cli/
  Makefile, README.md, agents.md, LICENSE, .github/, docs/, bin/, dist/  (unchanged locations)
```

## Step 1: move the module contents up one level

Run every command from the repository root. Using `git mv` preserves file
history. Move the two directories, both module files, and each loose library
`.go` file explicitly.

```sh
# Directories
git mv src/cmd cmd
git mv src/internal internal

# Module files
git mv src/go.mod go.mod
git mv src/go.sum go.sum

# Library sources and their tests (package gematria, at the module root)
git mv src/backend_index.go            backend_index.go
git mv src/backend_index_test.go       backend_index_test.go
git mv src/backend_remote.go           backend_remote.go
git mv src/backend_remote_test.go      backend_remote_test.go
git mv src/backend_sqlite.go           backend_sqlite.go
git mv src/backend_sqlite_test.go      backend_sqlite_test.go
git mv src/errors.go                   errors.go
git mv src/errors_test.go              errors_test.go
git mv src/gematria.go                 gematria.go
git mv src/gematria_test.go            gematria_test.go
git mv src/letters.go                  letters.go
git mv src/letters_test.go             letters_test.go
git mv src/result.go                   result.go
git mv src/result_test.go              result_test.go
git mv src/systems.go                  systems.go
git mv src/transliteration.go          transliteration.go
git mv src/transliteration_academic.go transliteration_academic.go
git mv src/transliteration_israeli.go  transliteration_israeli.go
git mv src/transliteration_test.go     transliteration_test.go
git mv src/wordlist.go                 wordlist.go
git mv src/wordlist_test.go            wordlist_test.go
```

After the moves, `src/` should be empty. Remove it:

```sh
rmdir src
```

Confirm nothing remains and that the tree is clean:

```sh
ls -A src 2>/dev/null && echo "src still has contents" || echo "src removed"
git status
```

You should see 25 renamed entries: 2 directories (`cmd`, `internal`),
`go.mod`, `go.sum`, and 21 loose `.go` files.

Note: if `git mv` refuses because the destination directory is not tracked or
similar, a plain `mv` followed by `git add -A` produces the same result; git
detects the renames on commit.

## Step 2: edit the Makefile

File: `Makefile`. Only line 2 changes. Every recipe uses `cd $(SRC_DIR) &&`
and `CMD_DIR := ./cmd/gematria` (line 5) is already relative to the module
root, so pointing `SRC_DIR` at `$(CURDIR)` makes all recipes correct without
further edits.

Before (line 2):

```make
SRC_DIR     := $(CURDIR)/src
```

After (line 2):

```make
SRC_DIR     := $(CURDIR)
```

Leave `VERSION_PKG := github.com/andreswebs/gematria/internal/cli` (line 7)
and `LDFLAGS` (line 8) unchanged; the module path does not change, so the
`-X` linker target is still valid.

The `cd $(SRC_DIR) && ...` prefixes on `build-local`, the `build-target`
template, `run`, `test`, `test-race`, `vet`, `fmt`, `fmt-check`, and `lint`
now `cd` into `$(CURDIR)`, which is a harmless no-op. This is the minimal
diff. Optionally, a follow-up could drop `SRC_DIR` and the `cd` prefixes
entirely for a cleaner Makefile; that is a larger diff and is not required
for the migration to work. Recommend the minimal one-line change here.

## Step 3: edit the GitHub workflows

The `src/` references live only in the two dispatch workflows. `build.yaml`
and `pull-request.yaml` merely call the dispatch workflows and contain no
`src/` references, so they need no changes. There is no SBOM or Syft step in
this repo (the release job in `build.yaml` uses `actions/attest` and
`gh release create` only), so there is no `path: src` to update.

File: `.github/workflows/build.dispatch.yaml`, the "Set up Go" step
(lines 50 and 51).

Before:

```yaml
          go-version-file: src/go.mod
          cache-dependency-path: src/go.sum
```

After:

```yaml
          go-version-file: go.mod
          cache-dependency-path: go.sum
```

File: `.github/workflows/validate.dispatch.yaml`, the "Set up Go" step
(lines 27 and 28). Identical change:

Before:

```yaml
          go-version-file: src/go.mod
          cache-dependency-path: src/go.sum
```

After:

```yaml
          go-version-file: go.mod
          cache-dependency-path: go.sum
```

## Step 4: edit dependabot config

File: `.github/dependabot.yml`. The `gomod` ecosystem points at `/src`
(line 9). After the move the module is at the repository root.

Before (lines 8 to 9):

```yaml
  - package-ecosystem: gomod
    directory: /src
```

After (lines 8 to 9):

```yaml
  - package-ecosystem: gomod
    directory: /
```

The `github-actions` ecosystem entry above it already uses `directory: /`
and is unaffected.

## Step 5: update the fleet workspace glue

These two files live one level above the gematria repository, in the
containing Go workspace. They are operational glue for local multi-module
development and are not part of the published gematria module. Update them so
the workspace keeps resolving after the move. If repositories migrate one at
a time, only the gematria entry changes now.

- Workspace `go.work` (repository's parent directory): change the gematria
  `use` entry from `./gematria/src` to `./gematria`. After editing, run
  `go work use` (or `go work sync`) from the workspace root so the file is
  normalized.

  Before:

  ```text
  use (
      ...
      ./gematria/src
      ...
  )
  ```

  After:

  ```text
  use (
      ...
      ./gematria
      ...
  )
  ```

- Workspace `AGENTS.md` (repository's parent directory) states that each
  module lives at `<project>/src/`. If migrating incrementally, add a note
  that gematria now lives at its repository root; rewrite fully once all
  repositories have migrated. This edit is optional for gematria to build and
  can be deferred to the fleet-wide cleanup.

## Step 6: documentation sweep

Import paths do not change, so no prose about the module path needs editing.
What does change is that source files formerly under `src/` are now at the
repository root, so in-repo docs that cite file paths like `src/systems.go`
or embed a `src/` directory tree must drop the `src/` prefix. These are
non-blocking for the build but should be corrected in the same change so the
docs stay accurate.

Files and lines to update (drop the leading `src/` from the cited paths):

- `docs/specs/code-architecture.md`: the package-structure tree starting at
  line 35 is rooted at `src/`. Re-root it at the repository root (remove the
  `src/` top level; keep the same file listing).
- `docs/specs/gematria-index.md`: lines 213, 238, 257 reference
  `src/internal/cli/...`.
- `docs/specs/transliteration.md`: lines 490 to 492 reference
  `src/transliteration*.go`.
- `docs/specs/wordlist-backends.md`: lines 12 to 14 reference `src/wordlist.go`,
  `src/backend_sqlite.go`, `src/backend_index.go`.
- `docs/manual-qa.md`: lines 51 (`Per src/go.mod`), 152, 235, 586, 1146,
  1148, 1173 reference `src/...` file paths and anchors.
- `docs/learnings.md`: line 11 references `src/gematria.go`.
- `.agent/PROMPT.plan.md`: line 1 references source files under `src/`. This
  is a historical prompt artifact; update only if the repository owner wants
  it kept accurate, otherwise leave as a historical record.

`agents.md` (and its `CLAUDE.md` symlink) and `README.md` contain no `src/`
references and need no changes for this migration.

The `.local/planning/*.md` files (for example `homebrew-tap-distribution.md`,
`remove-remote-backend.md`, `slsa-build-job-split.md`) also cite `src/`
paths, but the `.local/` directory is locally ignored and not part of the
published repository, so it is out of scope. Update it only if the owner
maintains those planning notes locally.

After editing, verify no functional `src/` references remain:

```sh
grep -rn 'src/' --exclude-dir=.git --exclude-dir=.tickets --exclude-dir=.local .
```

The only acceptable remaining hits are in `.agent/` or `.local/` if the owner
chose to leave those historical notes untouched. There must be no `src/`
references left in the `Makefile`, `.github/`, or `docs/` trees.

## Step 7: local verification

Run from the repository root.

```sh
# Module resolves at the root, dependencies are consistent
go mod verify
go build ./...

# Project checks (fmt-check, vet, lint, test) via the Makefile
make validate

# Build the local binary and smoke-test it
make build
./bin/gematria-$(go env GOOS)-$(go env GOARCH) --version
./bin/gematria-$(go env GOOS)-$(go env GOARCH) shalom
./bin/gematria-$(go env GOOS)-$(go env GOARCH) --help

# Convenience run target
make run
```

Expected: `make validate` passes (formatting clean, vet clean, lint clean,
all tests green) and the binary prints `gematria <version>` for `--version`
and a computed result for a word argument. The main package path is
`github.com/andreswebs/gematria/cmd/gematria` (confirmed in
`cmd/gematria/main.go`, which calls `cli.Run`).

## Step 8: external verification of both consumption modes

These steps require a resolvable tag on the default branch. They can only
pass after the change is merged and the repository owner pushes a new
version tag (for example `v0.1.1`), because the proxy resolves modules from
tagged commits. Old tags were never resolvable through the proxy (the module
was under `src/`), so nothing needs retracting.

Perform these from a scratch directory outside the workspace so the local
`go.work` does not shadow the proxy. Setting `GOWORK=off` guarantees that.

Consumption mode 1, install the CLI:

```sh
tmp="$(mktemp -d)"
cd "${tmp}"
GOWORK=off GOPROXY=proxy.golang.org GOBIN="${tmp}/bin" \
  go install github.com/andreswebs/gematria/cmd/gematria@latest
"${tmp}/bin/gematria" --version
"${tmp}/bin/gematria" shalom
```

Consumption mode 2, import the library from a fresh module:

```sh
tmp="$(mktemp -d)"
cd "${tmp}"
GOWORK=off go mod init example.com/gematria-consumer-check
cat > main.go <<'EOF'
package main

import (
 "fmt"

 "github.com/andreswebs/gematria"
)

func main() {
 r, err := gematria.Compute("shalom", gematria.MisparHechrachi)
 if err != nil {
  panic(err)
 }
 fmt.Println(r.Total)
}
EOF
GOWORK=off GOPROXY=proxy.golang.org go get github.com/andreswebs/gematria@latest
GOWORK=off go run .
```

Before relying on the exact identifiers in the import example above, confirm
the current public API names in `gematria.go` (the `Compute` signature and
the system constant name), since the example must compile against whatever
the tagged version exports. The goal of mode 2 is only to prove that
`go get github.com/andreswebs/gematria` resolves and that the root package is
importable; adjust the sample call to match the exported API if needed.

Expected: both commands succeed. Mode 1 installs and runs the CLI from the
proxy; mode 2 downloads the module and compiles a program that imports the
root `gematria` package. Together they prove the module is now distributable
in both forms.

## Rollback

The change is a pure relocation plus small config edits, all in the working
tree. To abandon before committing:

```sh
git checkout -- .
git clean -fd    # only if untracked copies were left by a manual mv
```

Because history is preserved via `git mv`, reverting the eventual commit (or
resetting the branch) restores the `src/` layout exactly. The workspace
`go.work` edit is independent; revert it separately if the repository move is
rolled back, so the `use` entry points back at `./gematria/src`.

## Change checklist

- [ ] `git mv` the 2 directories, `go.mod`, `go.sum`, and 21 loose `.go`
      files from `src/` to the repository root; `rmdir src`.
- [ ] `Makefile` line 2: `SRC_DIR := $(CURDIR)`.
- [ ] `.github/workflows/build.dispatch.yaml` lines 50 to 51: drop `src/`.
- [ ] `.github/workflows/validate.dispatch.yaml` lines 27 to 28: drop `src/`.
- [ ] `.github/dependabot.yml` line 9: `directory: /`.
- [ ] Workspace `go.work`: `./gematria/src` becomes `./gematria`; run
      `go work use`.
- [ ] Workspace `AGENTS.md`: note or rewrite the `src/` layout claim
      (optional, fleet cleanup).
- [ ] Docs sweep: `docs/specs/*`, `docs/manual-qa.md`, `docs/learnings.md`
      (and optionally `.agent/PROMPT.plan.md`).
- [ ] `make validate` and `make build` pass; binary runs.
- [ ] After merge and a new tag: `go install .../cmd/gematria@latest` and
      `go get github.com/andreswebs/gematria` from a scratch module both
      succeed.


## Notes

**2026-07-24T21:00:48Z**

FLEET DECISIONS finalized (see .local/planning/go-module-root-migration.md
"Decisions"):

- Library `.go` files stay at the repo root (as this plan recommends).
  Confirmed — no named subpackage, no import-path break.
- Makefile: minimal diff — `SRC_DIR := $(CURDIR)`, keep the cd prefixes.
- `.github/dependabot.yml`: gomod `directory: /src` -> `/` (already flagged).

SBOM follow-up (separate scope): unlike the other CLIs, this repo has no
signed-SBOM step (its release path is the build.dispatch/validate.dispatch
workflow set, not the template's release.yml). Adding one to match the fleet
is a workflow-modernization change beyond the module-root move. Do NOT block
this migration on it; it is tracked as its own concern.
