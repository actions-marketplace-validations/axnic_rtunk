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
`mise.lock`: the Go compiler matching `go.mod`'s floor, `golangci-lint`, `govulncheck`,
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
the `rtunk-mise-data` volume and survive rebuilds. `mise run ci` is the gate there as well.

### Mise tasks

CI never calls a tool directly: every job runs a task of `.mise.toml`, so a green local run is the
same gate a pull request goes through. `mise tasks` lists them.

| Task                     | What it does                                                                                         |
| ------------------------ | ---------------------------------------------------------------------------------------------------- |
| `mise run ci:lint`       | `golangci-lint run ./...`                                                                            |
| `mise run ci:build`      | Builds `./rtunk`                                                                                     |
| `mise run ci:rtunk`      | `rtunk check` with the released rtunk mise installs (pinned by `mise.lock`), changed files           |
| `mise run ci:test`       | `go test -race` with a coverage profile                                                              |
| `mise run ci:coverage`   | `ci:test`, then fails under the 80% statement-coverage floor                                         |
| `mise run ci:scripts`    | Tests of the release tooling in `scripts/` (Node's built-in runner)                                  |
| `mise run ci:commitlint` | Validates commit messages (`-- --from <sha> --to <sha>`)                                             |
| `mise run ci:vulncheck`  | `govulncheck ./...`                                                                                  |
| `mise run ci`            | `ci:lint`, `ci:build`, `ci:coverage`, `ci:scripts`: the CI gate minus `ci:rtunk` and commit messages |

Extra arguments reach rtunk through the task: `mise run ci:rtunk -- docs/Installation.md`. `ci:rtunk`
does not build: it runs the released rtunk, not `./rtunk`, so use `./rtunk` (after `mise run ci:build`)
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
check` (build it first with `mise run ci:build`; `mise run ci:rtunk` runs the released rtunk instead) runs the full
lint stack declared in [`.rtunk/rtunk.yaml`](.rtunk/rtunk.yaml) — `gofmt`, `golangci-lint2`,
`markdownlint`, `prettier`, `yamllint`, `taplo`, plus the security scanners (`grype`,
`osv-scanner`, `checkov`, `trufflehog`); it's read-only and exits non-zero on any finding. It
accepts path arguments to scope a run to what you changed; with none, the default is changed files
(see `--from`), not the whole repository.

`mise run ci` runs lint, build, tests with the coverage floor and the release-tooling tests
locally: the same gate the `lint`, `build` and `test` jobs below run, without the `rtunk` and
`commitlint` jobs.

For documentation-only changes, this repository's own convention is `./rtunk fmt <path>` then
`./rtunk check <path>` scoped to the files touched, in place of the full `./rtunk check` above.

## Continuous integration and releases

Workflows live in [`.github/workflows/`](.github/workflows), named `<triggers>.<name>.yaml`.

| Workflow                                  | Runs                                                                                                      |
| ----------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `merge_group,pull_request,push.lint.yaml` | `Lint`, on pull requests, merge-queue entries and pushes to `main`, always: `lint`, `rtunk`, `commitlint` |
| `merge_group,pull_request,push.go.yaml`   | `Go`, same triggers but path-filtered: `build`, `test`                                                    |
| `schedule.security.yaml`                  | Daily `govulncheck`, CodeQL and OpenSSF Scorecard                                                         |
| `push,workflow_dispatch.wiki.yaml`        | Publishes `docs/` to the GitHub Wiki on pushes to `main` that touch it                                    |
| `pull_request.dependabot-auto-merge.yaml` | Approves and auto-merges Dependabot patch and security updates                                            |
| `issue_comment,pull_request.pr-agent.yaml` | Posts a comment listing the commands when a pull request is opened; runs `/describe`, `/review`, `/improve`, `/ask`, `/help` comments from allowed users (PR Agent) |
| `workflow_dispatch.release.yaml`          | Cuts a release (run manually, see below)                                                                  |

The `lint` workflow always runs, on every file, with no path filter. `lint` runs golangci-lint; the
`rtunk` job dogfoods the tool and the [GitHub Action](docs/GitHub-Action.md) on its latest release:
it checks every file, and findings appear as annotations; `commitlint` validates the commit messages.

The `go` workflow runs `build` and `test`. On pull requests and pushes to `main` it only runs when
`**.go`, `go.mod`, `go.sum` or the workflow file itself changed. GitHub does not evaluate path
filters for `merge_group`, so merge-queue entries always run it. The `test` job runs on Linux and
macOS (`fail-fast: false`, so one platform's failure does not hide the other's); the release-tooling
tests run on Linux only.

[PR Agent](https://github.com/The-PR-Agent/pr-agent) (pinned by SHA, v0.47.0) never runs on its own
on pull requests or pushes. When a pull request is opened, a `welcome` job posts one comment listing
the commands, without calling the model; it is skipped for bots (Dependabot included), draft pull
requests and pull requests from forks, which cannot read secrets, and when `OPENROUTER_API_KEY` is
unset. A comment starting with `/` (`/describe`, `/review`, `/improve`, `/ask <question>`, `/help`)
runs that command in the `pr-agent` job, only for the GitHub logins listed in the repository
variable `PR_AGENT_ALLOWED_USERS` (Settings, Secrets and variables, Actions, Variables): a JSON list
such as `["xunleii", "someone"]`, compared case-insensitively and defaulting to `["xunleii"]` when
unset. Everyone else's comments are ignored, so nobody else can spend the key; edit the variable,
not the workflow, to add or remove someone. It never edits the pull request's title or body:
`/describe` publishes one persistent comment. The model (Pareto 26.10 preview, with Claude Sonnet
5.5 as the fallback) is reached through OpenRouter with the same `OPENROUTER_API_KEY` secret as the
release notes. The `welcome` job has `pull-requests: write` permission, the `pr-agent` job
`contents: read`, `issues: write` and `pull-requests: write`. The answers are advisory and never
block a merge; they are not a substitute for human review.

A release is cut from `main` through Actions, Release, Run workflow: the workflow computes the next
version from the last tag (or takes an explicit `version`), runs `mise run ci`, tags, builds the
`darwin` and `linux` archives (`amd64`, `arm64`) with GoReleaser, signs `checksums.txt` keyless
with cosign, attaches an SBOM per archive and records SLSA build provenance, then publishes a draft
release with generated notes for the maintainer to review. [SECURITY.md](SECURITY.md#verifying-a-release)
shows how to verify the result. Releases from `v0.14.0` on are signed; earlier ones are not.

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
2. Make the change, running the commands under "Tests and lint" as you go, not only at the end.
3. Commit following "Commit conventions" above: GPG-signed, `Assisted-by:` if AI-assisted (no
   `Co-Authored-By:` for the tool), never `--signoff`.
4. Open a pull request against `main` using the repository's pull request template
   (`.github/PULL_REQUEST_TEMPLATE.md`). `main` is protected: changes land through a pull request
   only, never by a direct push. The repository only allows merge commits (no squash, no rebase),
   and the merge commit takes the pull request title as its subject and the description as its
   body: write the title as a valid `type[scope]: Subject` header, since CI validates that merge
   commit on the push to `main`.
5. CI re-runs lint, commit-message validation, build and tests (with a coverage floor) on every
   pull request; the same gate runs locally with `mise run ci`. Passing "Tests and lint" locally
   before opening the PR is what keeps review fast.
