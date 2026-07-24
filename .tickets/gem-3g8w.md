---
id: gem-3g8w
status: closed
deps: [gem-kj4m]
links: [gem-kj4m]
created: 2026-07-24T21:58:23Z
type: task
priority: 2
assignee: Andre Silva
---
# Add a signed SBOM to the release, aligned with the fleet cookiecutter

# Add a signed SBOM to the gematria release, aligned with the fleet cookiecutter

## Goal

gematria is the only fleet CLI whose release pipeline emits no SBOM and no
cosign signatures. Every other CLI (and the cookiecutter template they derive
from) publishes, for each tagged release: a cross-platform archive set, a
`SHA256SUMS.txt`, a Syft-generated SPDX SBOM, cosign keyless signatures over
both the checksums and the SBOM, and SLSA build provenance.

This ticket brings gematria's release up to that standard by adding the SBOM
and the two cosign signatures. The provenance attestation already exists and
is left as-is (see "Out of scope").

## Prerequisite

Depends on the module-root migration (the "Migrate Go module from src/ to
repo root" ticket). Do that first. This plan assumes the module already lives
at the repo root, so the SBOM scans `path: .` and a root `.syft.yaml` governs
the scope. If for any reason this is picked up before the migration, stop and
sequence the migration first: doing the SBOM against `src/` would then have to
be redone.

## Reference: how the cookiecutter does it

The template's `.github/workflows/release.yml` (in the `_cookiecutter` repo,
under `{{cookiecutter.project_name}}/.github/workflows/release.yml`) is the
canonical shape. The relevant steps, in order, after `make dist`:

1. Install cosign (`sigstore/cosign-installer`), which provides cosign v3.x
   and writes the standardized Sigstore bundle format (`.sigstore.json`).
2. Generate SBOM (`anchore/sbom-action`) with `SYFT_SOURCE_NAME` and
   `SYFT_SOURCE_VERSION` set so the SBOM's root package is the released
   software's identity rather than the scanned path; `format: spdx-json`;
   `output-file: dist/<name>-<ref>.spdx.json`; `upload-artifact: false`.
3. Sign the checksums file: `cosign sign-blob --yes --bundle
   dist/SHA256SUMS.txt.sigstore.json dist/SHA256SUMS.txt`.
4. Sign the SBOM: `cosign sign-blob --yes --bundle <sbom>.sigstore.json
   <sbom>`.

Signing the checksums file transitively covers every archive listed in it.
Keyless signing requires `id-token: write` on the job.

## gematria's current release path (verified)

gematria does not use a single `release.yml`. It uses reusable workflows:

- `.github/workflows/build.yaml` — on push to `main` and on `v*` tags. Calls
  `validate.dispatch.yaml`, then `build.dispatch.yaml`, then (tags only) a
  `release` job that downloads the build artifact and runs `gh release
  create` with `dist/*.tar.gz dist/*.zip dist/SHA256SUMS.txt`.
- `.github/workflows/build.dispatch.yaml` — the reusable build job. Checks
  out, sets up Go, runs `make dist VERSION=...`, smoke-tests the linux-amd64
  binary, runs `actions/attest@v4` over the archives + checksums (build
  provenance), and uploads all of `dist/` as an artifact. This job already
  has `permissions: contents: write, id-token: write, attestations: write`.
- `validate.dispatch.yaml`, `pull-request.yaml` — validation only, not
  release-related.

Because `build.dispatch.yaml` is where `dist/` is produced and uploaded, the
SBOM and signatures must be generated there so they travel in the artifact to
the `release` job. The `release` job then needs to publish them.

## Changes

### 1. New file: `.syft.yaml` at the repo root

Preserve a Go-source-only SBOM scope now that the scan root is the repo root
(the release job runs `make dist` before the SBOM step, so `bin/` and `dist/`
are already populated and must not be cataloged as if they were sources):

```yaml
exclude:
  - ./bin/**
  - ./dist/**
  - ./docs/**
  - ./.git/**
```

### 2. `.github/workflows/build.dispatch.yaml`

Insert three steps into `jobs.build.steps`, after the existing
"Smoke-test built artifact" step and before the existing
"Attest build provenance" step. Keep the existing 6-space step indentation.

```yaml
      # cosign-installer v4 installs cosign v3.x, which writes the
      # standardized Sigstore bundle format (.sigstore.json). Consumers need
      # cosign v3+ to verify.
      - name: Install cosign
        uses: sigstore/cosign-installer@6f9f17788090df1f26f669e9d70d6ae9567deba6 # v4.1.2

      # SYFT_SOURCE_* names the SBOM's root package explicitly; without them
      # Syft derives it from the scanned directory path instead of the
      # released software's identity. Scope is governed by the root
      # .syft.yaml (Go source only).
      - name: Generate SBOM
        uses: anchore/sbom-action@e22c389904149dbc22b58101806040fa8d37a610 # v0.24.0
        env:
          SYFT_SOURCE_NAME: gematria
          SYFT_SOURCE_VERSION: ${{ github.ref_name }}
        with:
          path: .
          format: spdx-json
          output-file: dist/gematria-${{ github.ref_name }}.spdx.json
          upload-artifact: false

      # Keyless-sign the checksums (transitively covers every archive) and the
      # SBOM. Requires id-token: write, already present on this job.
      - name: Sign checksums and SBOM
        env:
          VERSION: ${{ github.ref_name }}
        run: |
          cosign sign-blob --yes \
            --bundle dist/SHA256SUMS.txt.sigstore.json \
            dist/SHA256SUMS.txt
          cosign sign-blob --yes \
            --bundle "dist/gematria-${VERSION}.spdx.json.sigstore.json" \
            "dist/gematria-${VERSION}.spdx.json"
```

Notes:

- These action pins (`sbom-action` v0.24.0, `cosign-installer` v4.1.2) are the
  exact versions the cookiecutter vets. gematria's other steps currently use
  floating major tags (`@v6`, `@v4`). Prefer the pinned SHAs here for
  supply-chain consistency with the rest of the fleet; if the repo owner would
  rather keep this file uniform with its floating-tag style, that is
  acceptable, but pinning is the fleet default.
- The `SYFT_SOURCE_VERSION`/`output-file` use `github.ref_name`, which on a
  `v*` tag build is the tag. The SBOM step only needs to run on tag builds;
  since `build.dispatch.yaml` also runs on pushes to `main`, either guard the
  three new steps with `if: ${{ startsWith(github.ref, 'refs/tags/') }}` (matches
  how the `release` job is gated), or accept that a `main`-branch run produces
  an SBOM named `gematria-main.spdx.json` that is simply never published.
  Recommended: add the `if:` guard to all three new steps so `main` builds stay
  unchanged.

Optional alignment (not required): the existing
"Attest build provenance" step uses `actions/attest@v4` with an explicit
`subject-path`. The cookiecutter instead uses
`actions/attest-build-provenance` with `subject-checksums:
dist/SHA256SUMS.txt`. Leaving gematria's provenance step as-is is fine; if you
want the SBOM covered by provenance too, add
`dist/gematria-${{ github.ref_name }}.spdx.json` to its `subject-path` list.

### 3. `.github/workflows/build.yaml` (release job)

The `release` job downloads the `dist/` artifact (which now contains the SBOM
and the two `.sigstore.json` bundles) and must attach them to the GitHub
release. Update the `gh release create` file list:

Before:

```yaml
          gh release create "${VERSION}" \
            --repo "${GITHUB_REPOSITORY}" \
            --title "${VERSION}" \
            --generate-notes \
            dist/*.tar.gz dist/*.zip dist/SHA256SUMS.txt
```

After:

```yaml
          gh release create "${VERSION}" \
            --repo "${GITHUB_REPOSITORY}" \
            --title "${VERSION}" \
            --generate-notes \
            dist/*.tar.gz dist/*.zip dist/SHA256SUMS.txt \
            dist/*.spdx.json dist/*.sigstore.json
```

No permission changes are needed: keyless signing happens in the build job
(which already has `id-token: write`); the release job only creates the
release (`contents: write`).

## Verification

- YAML lint / actionlint both dispatch files if available.
- Dry-run the SBOM and signing locally against a `make dist` output tree:
  - `make dist VERSION=v0.0.0-test`
  - `syft scan dir:. -o spdx-json` honoring `.syft.yaml` (confirm the catalog
    lists Go modules and does NOT list the archives under `dist/` or binaries
    under `bin/`).
  - `COSIGN_EXPERIMENTAL=1 cosign sign-blob --yes --bundle
    dist/SHA256SUMS.txt.sigstore.json dist/SHA256SUMS.txt` (interactive OIDC
    locally; in CI it is non-interactive via the workflow's id-token).
- End to end: push a throwaway pre-release tag (for example `v0.0.0-rc.test`)
  on a branch the owner controls, confirm the release contains the archives,
  `SHA256SUMS.txt`, `gematria-<tag>.spdx.json`, and the two `.sigstore.json`
  bundles, then delete the test release and tag.
- Consumer check: `cosign verify-blob --bundle
  gematria-<tag>.spdx.json.sigstore.json gematria-<tag>.spdx.json` succeeds
  with cosign v3+.

## Out of scope

- The module-root migration itself (separate ticket, prerequisite).
- Container/image SBOMs: gematria ships no container image, so the
  cookiecutter's `docker.yml` image-SBOM path does not apply.
- Rewriting the provenance attestation to the cookiecutter's
  `subject-checksums` form (optional alignment only).
- Migrating gematria from reusable dispatch workflows to a single
  `release.yml`. This ticket adds the SBOM within the existing structure and
  does not restructure the workflows.

Branching, commits, and tags are handled by the repository owner. Do not
create git commits or tags.


## Notes

**2026-07-24T22:34:46Z**

Added signed SBOM to release, aligned with fleet cookiecutter. Changes: (1) new .syft.yaml at repo root excluding ./bin/**, ./dist/**, ./docs/**, ./.git/** so the SBOM catalogs Go sources only now that the module scans path '.'; (2) build.dispatch.yaml gains three steps between smoke-test and attest — Install cosign (sigstore/cosign-installer pinned v4.1.2), Generate SBOM (anchore/sbom-action pinned v0.24.0, SYFT_SOURCE_NAME=gematria/SYFT_SOURCE_VERSION=ref_name, spdx-json to dist/gematria-<ref>.spdx.json, upload-artifact:false), and Sign checksums+SBOM (cosign sign-blob --bundle *.sigstore.json). All three guarded with 'if: startsWith(github.ref, refs/tags/)' so main-branch builds are unchanged; (3) build.yaml release job appends 'dist/*.spdx.json dist/*.sigstore.json' to gh release create. Signing runs in the build job (already has id-token:write); release job only needs contents:write. Left provenance step as-is (optional alignment, out of scope). Verified: actionlint clean on both workflows (installed via go install; the one SC2086 info at line 68 is the pre-existing smoke-test step, not this change); make build passes; make dist produces the expected dist/ tree the new steps consume. syft/cosign not installed locally and cosign keyless signing needs CI OIDC, so the local scan/sign dry-run could not run here — end-to-end verification via a throwaway tag is left to the repo owner per ticket.
