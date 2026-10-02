# Changelog

All notable changes to `rtunk` are documented in this file. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

rtunk's first tagged release is `v0.13.0` (2026-10-01); `v1.0` is planned and not yet tagged.
`v0.14.0` changes the Go module path (breaking for importers and `go install` users).
`[0.13.0]` below collects every user-visible change shipped by the milestones
[ROADMAP.md](./ROADMAP.md) records as done (`v0.10` through `v0.13`), grouped by milestone. Later
releases list only what changed since the previous tag. A build with no version metadata (a local
`go build`) still reports `dev`.

## [Unreleased]

Nothing yet.

## [0.14.0] - 2026-10-02

### Added (v0.14.0)

- A reusable GitHub Action, `axnic/rtunk`, installs and verifies rtunk, caches its tools and
  plugins, runs `rtunk check` and reports the findings as annotations and a job summary. See
  [docs/GitHub-Action.md](./docs/GitHub-Action.md).
- Release artifacts are signed and attested: `checksums.txt` is signed keyless with cosign
  (`checksums.txt.sigstore.json`), every archive ships an SBOM (`*.sbom.json`) and a SLSA build
  provenance attestation. This is the first signed release; see
  [SECURITY.md](./SECURITY.md#verifying-a-release).
- A dev container (Ubuntu, `gh`, `mise`, Go tooling) for contributors.

### Changed (v0.14.0)

- **Breaking:** the Go module path is now `github.com/axnic/rtunk` (was `github.com/xunleii/rtunk`).
  Update imports and use `go install github.com/axnic/rtunk/cmd/rtunk@v0.14.0`. The `v0.13.x` tags
  still declare the old path, so `v0.14.0` is the first version installable under the new one.
  Release archives are unaffected.
- CI runs the tests on Linux and macOS.
- This repository is linted with rtunk itself (`.rtunk/rtunk.yaml`, through the GitHub Action)
  instead of trunk, and every finding rtunk reported across CI, docs and config is fixed.
- Documentation statements that contradicted the released state are fixed.

## [0.13.2] - 2026-10-02

### Fixed (v0.13.2)

- The release workflow no longer silently falls back to an incomplete draft of the release notes
  when the AI-generated text is rejected: the model now only writes the summary paragraph, and a
  rejected answer is logged for diagnosis.

## [0.13.1] - 2026-10-01

### Added (v0.13.1)

- The release workflow generates the release notes with an OpenRouter model.

### Changed (v0.13.1)

- The CI workflows pin the mise toolchain and share one cache.
- Dependency bumps: `github.com/ulikunitz/xz` 0.5.17, `github.com/stretchr/testify` 1.12.1, and the
  GitHub Actions the workflows use.

## [0.13.0] - 2026-10-01

First tagged release: darwin and linux archives for amd64 and arm64, with `checksums.txt`.

### v0.13 — Documentation

#### Added (v0.13)

- `rtunk download` is a top-level command that downloads what the configuration needs, with a
  live install view.
- `rtunk toolbox link` symlinks the `.rtunk` cache paths into the repository.
- `rtunk linters enable` and `rtunk actions enable` offer an interactive picker; unknown ids are
  rejected.
- Local override files layer over the loaded configuration; `rtunk linters list` and the
  enable/disable warnings say when an override decides a linter's state.
- Inline suppression with `rtunk-ignore` / `trunk-ignore` directives, which require a comment
  leader.
- `--security-only` runs only the commands tagged `is_security`.
- A `shellcheck` output-format parser.

#### Changed (v0.13)

- `rtunk init` migrates an existing `.trunk` directory to `.rtunk` and chains the pickers and the
  download.
- `rtunk cache clean` reports per-subtree progress and the size freed.
- `rtunk linters list` output is colorized.
- `rtunk renovate` is promoted out of the hidden `toolbox` command group.

#### Removed (v0.13)

- The `upgrade` command and the self-update capability.
- The desktop notification on action failure.

#### Fixed (v0.13)

- `fmt` and `check` no longer error on untracked symlinks at the repository root.
- `check` streams its own pass live instead of buffering it whole.
- Hook scripts of actions point at the `sync` / `unsync` commands.
- A recipe-level `executable` is honored for raw-binary downloads.
- An unsubstituted `${major_version}` in tool package paths is now resolved.
- `extractVersion` is kept in Renovate comments when linters are rewritten.

### v0.12 — Renovate integration, promoted to a public command

#### Changed (v0.12)

- The `renovate` command group (annotate a configuration for Renovate and print its regex-manager
  snippet) is now listed in `rtunk help` by default, without needing `--all`.

### v0.11 — Cache, provisioning, and catalog fidelity

#### Added (v0.11)

- A linter whose catalog entry declares a "suggest when" condition is now offered as a suggestion
  once that condition holds for the current repository.
- A command with a declared run timeout is now stopped and reported as failed if it exceeds that
  timeout, instead of being able to run indefinitely.
- Findings from a command flagged security-related in the catalog are now tagged, so they can be
  filtered or displayed separately from ordinary findings.
- A tool that declares companion packages now gets them installed alongside it automatically.
- Concurrent installs of the same item now fail fast, by name: a second process that finds an
  install already in progress either proceeds immediately (if the recorded process is no longer
  running) or stops immediately with an error naming the repository and item, instead of waiting.

#### Changed (v0.11)

- `rtunk cache clean` is now a full, unconditional wipe of the entire cache root — downloads,
  plugin sources, and logs — replacing the previous separate "destroy everything" and age-based
  "prune" commands.
- `rtunk cache prune` now keeps exactly what a currently-existing, currently-configured repository
  still needs, instead of pruning by age; a repository that no longer exists is dropped from the
  record.
- The archive used to install a tool or runtime is no longer kept on disk once the install
  completes.

#### Fixed (v0.11)

- A command with a declared concurrency limit is now honored and never runs more instances in
  parallel than that limit, regardless of the run's overall worker count.
- A tool with declared health checks is now verified right after install, instead of a broken
  install only surfacing later when a linter using it fails.
- When both a specific command and the more generic linter it supersedes are enabled, the
  superseded one's duplicate findings on the same issue are no longer reported twice.
- A command that declares a one-time setup step now runs that step before its first invocation.

### v0.10 — Unify and correct check, fmt, and fix

#### Added (v0.10)

- `--format-before-check` flag runs every formatter first, then checks the reformatted files — the
  "format, then check" behavior that used to be `check --fix`'s only behavior is now explicit and
  opt-in instead of implicit and default.
- Enabling a linter or command flagged deprecated in the catalog now produces a warning naming its
  replacement.

#### Changed (v0.10)

- `check --fix` now applies linter fix commands and finding-level autofixes only, then reports
  whatever remains unfixed; it no longer runs formatters as a side effect.
- Inside a git repository with no upstream branch, a path-less `check`/`fmt` now picks up every
  staged, unstaged, and new untracked file, instead of staged changes only.
- Enabling a linter that still uses the old, single-command declaration shape is now rejected at
  configuration time, naming its replacement, instead of running and silently finding nothing.
- Outside a git repository, a path-less `check`/`fmt` now fails with an explicit error requiring
  paths, instead of exiting successfully having checked nothing.

#### Fixed (v0.10)

- A fix-only linter (one whose autofix is an in-place fix without also being a formatter) is now
  selectable and actually runs under `check --fix`, where it previously never ran under any
  command.
- A finding that already carries a computer-generated replacement from its own linter is now
  applied by `check --fix`, instead of only being reported as text.
- A formatter that rewrites content through its own standard input/output, rather than editing the
  file in place, now runs correctly under `fmt`, instead of being permanently inert.
- A linter command declared for a specific tool-version range now only runs when the resolved tool
  version falls in that range, instead of every declared variant running unconditionally.
- A command variant declared for a platform rtunk does not support (Windows) is no longer attempted
  on a supported platform.
