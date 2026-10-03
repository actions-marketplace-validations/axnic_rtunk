# Contributing to rtunk

The path from a fresh clone to an accepted change: setting up a development environment, the
tests and lint every change must pass locally, this project's commit conventions, and how to
submit a pull request. Audience: external contributors.

All participation in this project is governed by [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## Development environment

Use [docs/Installation.md](docs/Installation.md)'s "clone and build" path rather than
`go install` — it's the setup this file assumes:

```bash
git clone https://github.com/axnic/rtunk.git
cd rtunk
mise trust    # if mise prompts about this repo's .mise.toml
mise install  # installs the Go toolchain and dev tools .mise.toml declares
go build -o rtunk ./cmd/rtunk
```

`mise install` resolves the toolchain declared in [`.mise.toml`](.mise.toml), pinned by
`mise.lock`: the Go compiler matching `go.mod`'s floor, `govulncheck`,
`goreleaser`, `cosign`, `syft` and Node with `commitlint`. The metalinter used below is rtunk
itself, built from your clone, so there is nothing else to install. See
[docs/Installation.md](docs/Installation.md) for prerequisites and platform support (macOS and
Linux only) — this file doesn't repeat that.

### Dev container

[`.devcontainer/devcontainer.json`](.devcontainer/devcontainer.json) gives VS Code, Codespaces
or the `devcontainer` CLI the same toolchain without installing anything on the host: a
non-root Ubuntu image (amd64 and arm64) with `git`, `gh`, `mise` and the Go extension. On first
creation it runs `mise install`, which resolves [`.mise.toml`](.mise.toml) as pinned by
`mise.lock`; the config is pre-trusted, so no `mise trust` is needed. Downloaded tools live in
the `rtunk-mise-data` volume and survive rebuilds. `mise run ci` is the local gate there as well.

### Mise tasks

CI never calls a tool directly: the central workflows run tasks of `.mise.toml`, so a green local run
matches what a pull request goes through (`lint` and `security:audit` sit outside the `ci:` namespace). `mise tasks` lists them.

| Task                      | What it does                                                                                                         |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `mise run lint`           | `rtunk check .` with the released rtunk mise installs (pinned by `mise.lock`), every file                            |
| `mise run lint:fix`       | `rtunk check --fix .`                                                                                                |
| `mise run ci:lint`        | Checks `action.yml` can be published to the Marketplace (single-line name, description of at most 125 characters)    |
| `mise run ci:build`       | Builds `./rtunk`                                                                                                     |
| `mise run ci:test`        | `go test -race` with a coverage profile                                                                              |
| `mise run ci:coverage`    | Fails under the 80% statement-coverage floor, from the `coverage.txt` of `ci:test` (runs the tests itself if absent) |
| `mise run ci:commitlint`  | Validates commit messages (`-- --from <sha> --to <sha>`)                                                             |
| `mise run security:audit` | `govulncheck ./...`                                                                                                  |
| `mise run ci`             | `ci:lint`, `ci:build`, `ci:test`, `ci:coverage`: the release gate, without `lint`, commit messages and the audit     |

`lint` does not build: it runs the released rtunk, not `./rtunk`, so use `./rtunk` (after `mise run ci:build`)
to exercise your own changes.

## Tests and lint

Run all of the following before opening a pull request:

```bash
go test ./...
go vet ./...
gofmt -l .
./rtunk check
```

`go test ./...` and `go vet ./...` exit `0` with no output on success. `gofmt -l .` exits `0` and
prints nothing when the tree is already formatted; any path it lists needs `gofmt -w`. `./rtunk
check` (build it first with `mise run ci:build`; `mise run lint` runs the released rtunk instead) runs the full
lint stack declared in [`.rtunk/rtunk.yaml`](.rtunk/rtunk.yaml) — `gofmt`, `golangci-lint2`,
`markdownlint`, `prettier`, `yamllint`, `taplo`, plus the security scanners (`grype`,
`osv-scanner`, `checkov`, `trufflehog`); it's read-only and exits non-zero on any finding. It
accepts path arguments to scope a run to what you changed; with none, the default is changed files
(see `--from`), not the whole repository.

Testing policy: every new feature and every bug fix adds or updates automated tests of that
behavior in the same pull request; a fix starts with a test that fails without it. The pull
request states how the change was tested in the "How this was tested" section of its template,
and a change that cannot be tested automatically says why there.

Two Go native fuzz tests, `FuzzPathMatches` (`pkg/ignore/path_fuzz_test.go`) and
`FuzzInstallDownloadTarGz` (`pkg/cache/download/extract_fuzz_test.go`), run their seed corpus with
every `go test ./...`. To explore beyond it:

```bash
go test -run '^$' -fuzz=FuzzPathMatches -fuzztime=30s ./pkg/ignore
go test -run '^$' -fuzz=FuzzInstallDownloadTarGz -fuzztime=30s ./pkg/cache/download
```

A crash writes a reproducer under `testdata/fuzz/` next to the test; commit it with the fix.

`mise run ci` runs the `action.yml` check, build and tests with the coverage floor locally: the same
tasks the central `test` workflow runs. It does not lint: run `mise run lint` for that.

For documentation-only changes, this repository's own convention is `./rtunk fmt <path>` then
`./rtunk check <path>` scoped to the files touched, in place of the full `./rtunk check` above.

## Continuous integration and releases

CI is defined centrally in the `axnic/.github` repository as reusable workflows. The callers in
[`.github/workflows/`](.github/workflows) are generated by Terraform, named `<triggers>.<action>.yaml`:
do not edit them here.

| Action      | Runs                                                                                      |
| ----------- | ----------------------------------------------------------------------------------------- |
| `qa`        | Quality Assurance: lint, rtunk, commit messages (`merge_group,pull_request,push.qa.yaml`) |
| `test`      | Quality Assurance: build and tests (`ci:build`, `ci:test`, `ci:coverage`)                 |
| `review`    | AI Review (PR Agent)                                                                      |
| `scan`      | Code Scanning                                                                             |
| `deps`      | Dependency Updates (auto-merges Dependabot PRs; merge commit subject `[deps]: Bump ...`)  |
| `audit`     | Dependency Audit (`security:audit`)                                                       |
| `scorecard` | OpenSSF Scorecard                                                                         |
| `wiki`      | Publishes `docs/` to the GitHub Wiki                                                      |
| `release`   | Release (run manually, see below; `workflow_dispatch.release.yaml`)                       |

The hand-written `merge_group,pull_request,push.action.yaml` (Action Check) dogfoods the
[GitHub Action](docs/GitHub-Action.md) on rtunk's latest release.

Look in [`.github/workflows/`](.github/workflows) for the exact trigger-prefixed file names and in
`axnic/.github` for what each reusable workflow does. Some of the tasks they call are defined in
[`.mise.toml`](.mise.toml): `ci:commitlint`, `ci:build`, `ci:test`, `ci:coverage`, `security:audit`,
and `ci:lint` when present.

[PR Agent](https://github.com/The-PR-Agent/pr-agent) is the AI Review workflow. It is configured
centrally; its allowed users, model and secrets are not described here, check `axnic/.github`.
Its answers are advisory and never block a merge; they are not a substitute for human review.

A release is cut from `main` through Actions, Release, Run workflow (defined by the Terraform-generated
`workflow_dispatch.release.yaml`). Pass exactly one input: `bump` (`auto`, `patch`, `minor` or `major`)
or `version` (an exact version such as `0.13.0`, or `0.13.0-rc.1` for a release candidate). The workflow
runs `mise run ci`, creates the tag and a draft release with AI-written notes, then publishes the
`darwin` and `linux` archives (`amd64`, `arm64`) with GoReleaser, signed keyless with cosign, with an SBOM
per archive and SLSA build provenance. [SECURITY.md](SECURITY.md#verifying-a-release) shows how to verify
the result. Releases from `v0.14.0` on are signed; earlier ones are not.

Listing the action on the GitHub Marketplace is a manual step, as GitHub documents no API or CLI for it: when
publishing the draft, tick "Publish this Release to the GitHub Marketplace" in the release form (it needs the
valid `action.yml` at the repository root and the Marketplace Developer Agreement accepted once by the owner).
`mise run ci:lint` (part of `mise run ci`, so it also runs before a release is tagged) fails on an `action.yml`
without a single-line name or with a description over 125 characters; it cannot check that the name (`rtunk check`)
is unique, which only GitHub knows.

### `.trunk` and `.rtunk`

This repository is linted by rtunk itself, through [`.rtunk/rtunk.yaml`](.rtunk/rtunk.yaml): trunk
is not needed to work on rtunk, and the former `.trunk/` directory no longer exists. rtunk reads
`.rtunk/rtunk.yaml` first and falls back to `.trunk/trunk.yaml`; `rtunk init` migrates a `.trunk/`
directory into `.rtunk/`.

## Commit conventions

Every commit follows `type[scope]: Subject` — a one-character type symbol (`+` Add, `-` Remove,
`~` Improve, `!` Fix, `=` Refactor, `^` Bump, `>` Move, `<` Revert, `@` Docs, `$` Security, `?`
Experiment, `*` Wildcard; `+!`/`~!`/`-!` for breaking changes), a mandatory bracketed scope, and
an imperative, sentence-case subject with no trailing period. commitlint also requires a header of
at most 100 characters that is the very first line of the message, a sentence-case body (first
letter a capital, so never start it with a lowercase word such as a command name) and body and
footer lines of at most 80 characters, with a blank line before the footer. An AI-assisted commit
must carry an `Assisted-by: <provider>:<model-id>` trailer (dots in version numbers, not hyphens),
and must not carry `Co-authored-by:`/`Co-Authored-By:` for the AI tool: a tool a human directs isn't
a co-author. commitlint has no rule on trailers, so it does not reject a `Co-Authored-By:` line that
an assistant adds on its own; the rule is a project convention, enforced in review. If your tooling
adds that trailer, strip it (amend the message or disable the tool's attribution setting) and keep
`Assisted-by:`. Every commit is GPG-signed (`git commit -S`); never add `-s`/`--signoff`, since
DCO sign-off is the human committer's own attestation and an AI assistant must stay out of it.
Full type/scope tables, the commitlint rules commits are checked against, and the drafting
workflow: [`.agents/skills/git-commit/SKILL.md`](.agents/skills/git-commit/SKILL.md) — the
canonical reference; this section is a summary, not a substitute.

## Documentation

`docs/` is the source of the project's GitHub Wiki: flat `Title-Case.md` pages, `docs/README.md`
as Home, `_Sidebar.md` for navigation. A change to a command, flag or config key updates its page
in the same commit. Page inventory, templates and writing rules:
[`.agents/skills/wiki-docs/SKILL.md`](.agents/skills/wiki-docs/SKILL.md).

## Submitting a change

1. Fork the repository and branch off `main` — the project's only active integration branch;
   releases are tagged from it, so there is no separate branch to keep in sync.
2. Make the change, with its tests (see the testing policy under "Tests and lint"), running the
   commands there as you go, not only at the end.
3. Commit following "Commit conventions" above: GPG-signed, `Assisted-by:` if AI-assisted (no
   `Co-Authored-By:` for the tool), never `--signoff`.
4. Open a pull request against `main` using the repository's pull request template
   (`.github/PULL_REQUEST_TEMPLATE.md`). `main` is protected: changes land through a pull request
   only, never by a direct push. The repository only allows merge commits (no squash, no rebase),
   and the merge commit takes the pull request title as its subject and the description as its
   body: write the title as a valid `type[scope]: Subject` header, since CI validates that merge
   commit on the push to `main`.
5. CI re-runs lint, commit-message validation, build and tests (with a coverage floor) on every
   pull request; the build and tests run locally with `mise run ci`, the lint with `mise run lint`. Passing "Tests and lint" locally
   before opening the PR is what keeps review fast.
